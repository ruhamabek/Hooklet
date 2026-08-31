package replay_test

import (
	"bytes"
	"context"
	"hooklet/internal/model"
	"io"
	"net/http"
	"net/http/httptest"
	"hooklet/internal/replay"
	"testing"
	"time"
)

type SpyReplayStore struct {
    savedAttempts []*model.ReplayAttempt
}

func (s *SpyReplayStore) Close() error{ return nil }
func (s *SpyReplayStore) SaveRequest(ctx context.Context, req *model.WebhookRequest) error { return nil }
func (s *SpyReplayStore) GetRequest(ctx context.Context, id string) (*model.WebhookRequest, error) {
	return nil, nil
}
func (s *SpyReplayStore) ListRequests(ctx context.Context, limit int, offset int) ([]*model.WebhookRequest, error) {
	return nil, nil
}
func (s *SpyReplayStore) SaveReplayAttempt(ctx context.Context, attempt *model.ReplayAttempt) error {
	s.savedAttempts = append(s.savedAttempts, attempt)
	return nil
}
func (s *SpyReplayStore) GetReplayAttempts(ctx context.Context, requestID string) ([]model.ReplayAttempt, error) {
	return nil, nil
}

func TestDispatcher_RelaySuccess(t *testing.T){
	var receivedHeaders http.Header
	var receivedBody []byte
	fakeFastAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received": true}`))
	}))
	defer fakeFastAPI.Close()

	spyStore := &SpyReplayStore{}
	dispatcher := replay.NewDispatcher(spyStore, http.DefaultClient)
	original := &model.WebhookRequest{
		ID:     "req-stripe-123",
		Method: "POST",
		Path:   "/webhooks/stripe",
		Headers: model.HeaderMap{
			"Content-Type":     {"application/json"},
			"Stripe-Signature": {"t=123,v1=sig"},
		},
		Body:      []byte(`{"event":"charge.succeeded"}`),
		CreatedAt: time.Now().UTC(),
	}

	attempt, err := dispatcher.Replay(context.Background(), original, fakeFastAPI.URL)
	if err != nil {
		t.Fatalf("unexpected error replaying: %v", err)
	}

	if !bytes.Equal(receivedBody, original.Body) {
		t.Errorf("body mismatch: got %s, want %s", string(receivedBody), string(original.Body))
	}
	if receivedHeaders.Get("Stripe-Signature") != "t=123,v1=sig" {
		t.Errorf("Stripe-Signature header not forwarded: %v", receivedHeaders)
	}
	if receivedHeaders.Get("X-Hooklet-Replayed") != "true" {
		t.Errorf("expected X-Hooklet-Replayed header to be true, got %q", receivedHeaders.Get("X-Hooklet-Replayed"))
	}

	if attempt.StatusCode != http.StatusOK {
		t.Errorf("expected attempt status 200, got %d", attempt.StatusCode)
	}
	if string(attempt.ResponseBody) != `{"received": true}` {
		t.Errorf("unexpected response body: %s", string(attempt.ResponseBody))
	}
	if len(spyStore.savedAttempts) != 1 {
		t.Fatalf("expected 1 attempt saved to store, got %d", len(spyStore.savedAttempts))
	}
}

func TestDispatcher_ConnectionRefused(t *testing.T) {
	spyStore := &SpyReplayStore{}
	dispatcher := replay.NewDispatcher(spyStore, http.DefaultClient)

	req := &model.WebhookRequest{
		ID:        "req-fail",
		Method:    "POST",
		Path:      "/webhook",
		Body:      []byte(`{}`),
		CreatedAt: time.Now().UTC(),
	}

 	attempt, err := dispatcher.Replay(context.Background(), req, "http://127.0.0.1:59999/wh")
	if err == nil {
		t.Fatalf("expected error for offline server, got nil")
	}

	if attempt.StatusCode != 0 {
		t.Errorf("expected status code 0, got %d", attempt.StatusCode)
	}
	if attempt.Error == "" {
		t.Errorf("expected error message to be recorded, got empty string")
	}
	if len(spyStore.savedAttempts) != 1 {
		t.Errorf("expected failed attempt to still be saved to store, got %d", len(spyStore.savedAttempts))
	}
}