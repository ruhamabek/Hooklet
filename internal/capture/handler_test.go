package capture_test

import (
	"bytes"
	"context"
	"errors"
	"hooklet/internal/capture"
	"hooklet/internal/event"
	"hooklet/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type SpyStore struct {
	savedRequests []*model.WebhookRequest
	saveErr       error
}


func (s *SpyStore) Close() error {return nil}

func(s *SpyStore) SaveRequest(ctx context.Context, req *model.WebhookRequest) error {
	if s.saveErr != nil{
		return s.saveErr
	}
	s.savedRequests = append(s.savedRequests, req)
	return nil
}

func (s *SpyStore) GetRequest(ctx context.Context, id string)(*model.WebhookRequest,error) {
    return nil,nil
}


func (s *SpyStore) ListRequests(ctx context.Context, limit int, offset int) ([]*model.WebhookRequest, error) {
	return nil, nil
}
func (s *SpyStore) SaveReplayAttempt(ctx context.Context, attempt *model.ReplayAttempt) error {
	return nil
}

func (s *SpyStore) GetReplayAttempts(ctx context.Context, requestID string)([]model.ReplayAttempt, error) {
	return nil, nil
}


func TestCaptureHandler_Success(t *testing.T){
	spy := &SpyStore{}
	broker := event.NewBroker()
	ch := broker.Subscribe()
	defer broker.Unsusbcribe(ch)

	handler := capture.NewHandler(spy, broker)

	payload := []byte(`{"event":"payment_intent.succeeded","amount":2000}`)
	req := httptest.NewRequest(http.MethodPost, "/wh/stripe?source=checkout", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=12345,v1=sig_abc")

	rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", rec.Code)
	}
    
	if len(spy.savedRequests) != 1 {
		t.Fatalf("expected 1 saved request, got %d", len(spy.savedRequests))
	}
	saved := spy.savedRequests[0]
	if saved.Method != "POST" {
		t.Errorf("expected method POST, got %s", saved.Method)
	}
	if saved.Path != "/wh/stripe" {
		t.Errorf("expected path /wh/stripe, got %s", saved.Path)
	}
	if saved.QueryString != "source=checkout" {
		t.Errorf("expected query source=checkout, got %s", saved.QueryString)
	}
	if !bytes.Equal(saved.Body, payload) {
		t.Errorf("body mismatch: got %s, want %s", string(saved.Body), string(payload))
	}
	if saved.Headers["Stripe-Signature"][0] != "t=12345,v1=sig_abc" {
		t.Errorf("header mismatch: got %v", saved.Headers)
	}
	
	select {
		case broadcasted := <-ch:
			if broadcasted.ID != saved.ID {
				t.Errorf("broker broadcasted wrong ID: got %s, want %s", broadcasted.ID, saved.ID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out waiting for broker to broadcast webhook")
		}
}

func TestCaptureHandler_StoreFailure(t *testing.T){
	spy := &SpyStore{
		saveErr: errors.New("disk I/O error"),
	}

	handler := capture.NewHandler(spy, nil) 
    req := httptest.NewRequest(http.MethodPost, "/wh/stripe", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusInternalServerError{
		t.Errorf("expected status 500, got %v", rec.Code)
	}
}

func TestCaptureHandler_PayloadTooLarge(t *testing.T){
	hugePayload := make([]byte,  11*1024*1024)
	spy := &SpyStore{}
	handler := capture.NewHandler(spy, nil)

	req := httptest.NewRequest(http.MethodPost, "/wh/test", bytes.NewReader(hugePayload))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
 	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, got %d", rec.Code)
	}
 
	if len(spy.savedRequests) != 0 {
		t.Errorf("expected 0 requests saved, got %d", len(spy.savedRequests))
	}
}

 