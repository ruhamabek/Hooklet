package store_test

import (
	"bytes"
	"context"
	"hooklet/internal/model"
	"hooklet/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) store.Store {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	s, err := store.NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize test store: %v", err)
	}

	t.Cleanup(func() {
		_ = s.Close()
	})

	return s
}

func TestSaveAndGetRequest(t *testing.T) {
	ctx := context.Background()
	 s := newTestStore(t)

	
	rawBody := []byte(`{"id":"evt_123","type":"payment_intent.succeeded"}`)

	original := &model.WebhookRequest{
		ID:          "req-abc-123",
		Method:      "POST",
		Path:        "/webhooks/stripe",
		QueryString: "account=acct_xyz",
		Headers: model.HeaderMap{
			"Content-Type":     {"application/json"},
			"Stripe-Signature": {"t=1600000000,v1=secret_signature"},
		},
		Body:          rawBody,
		ContentType:   "application/json",
		ContentLength: int64(len(rawBody)),
		RemoteAddr:    "127.0.0.1:45678",
		CreatedAt:     time.Now().UTC().Truncate(time.Millisecond),
	}

	err := s.SaveRequest(ctx, original)

	if err != nil {
		t.Fatalf("failed to save request: %v", err)
	}

	retrieved, err := s.GetRequest(ctx, original.ID)

	if err != nil {
		t.Fatalf("Failed to get request: %v", err)
	}

	if retrieved.ID != original.ID {
		t.Errorf("ID mismatch: got %q, want %q", retrieved.ID, original.ID)
	}
   
	if !bytes.Equal(retrieved.Body, original.Body) {
		t.Errorf("Body mismatch: got %s, want %s", string(retrieved.Body), string(original.Body))
	}
	if retrieved.Headers["Stripe-Signature"][0] != "t=1600000000,v1=secret_signature" {
		t.Errorf("Signature header mismatch: got %v", retrieved.Headers)
	}
}

func TestGetRequest_NotFound(t *testing.T) {
	ctx := context.Background()
	 s := newTestStore(t)
	 _,err := s.GetRequest(ctx, "non-existent-id")
	 if err != store.ErrNotFound {
		t.Fatalf("expected store.ErrNotFound, got %v", err)
	 }
}

func TestListRequests(t *testing.T){
	ctx := context.Background()
	s := newTestStore(t)

	now := time.Now().UTC()

	for i := 1; i <= 3; i++ {
		req := &model.WebhookRequest{
			ID:        string(rune('A' + i - 1)), 
			Method:    "POST",
			Path:      "/wh/test",
			Headers:   model.HeaderMap{"Content-Type": {"application/json"}},
			Body:      []byte(`{}`),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		}
		if err := s.SaveRequest(ctx, req); err != nil {
			t.Fatalf("failed to save req %d: %v", i, err)
		}
	}
    
	list, err := s.ListRequests(ctx, 2, 0)
	if err != nil {
		t.Fatalf("failed to list requests: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("Expected 2 items, got %v", len(list))
	}

	if list[0].ID != "C" || list[1].ID != "B" {
		t.Errorf("ordering mismatch: expected C, B; got %s, %s", list[0].ID, list[1].ID)
	}

   	page2, err := s.ListRequests(ctx, 2, 2)
	if err != nil {
		t.Fatalf("failed to list page 2: %v", err)
	}
	if len(page2) != 1 || page2[0].ID != "A" {
		t.Errorf("pagination mismatch: expected [A], got %v", page2)
	}

}

func TestSaveAndGetReplayAttempts(t *testing.T){
	ctx := context.Background()
	 s := newTestStore(t)
     
	 req := &model.WebhookRequest{
		ID:        "req-to-replay",
		Method:    "POST",
		Path:      "/webhooks/test",
		Headers:   model.HeaderMap{"Content-Type": {"application/json"}},
		Body:      []byte(`{"test":true}`),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.SaveRequest(ctx, req); err != nil {
		t.Fatalf("failed to save req: %v", err)
	}
    
	attempt := &model.ReplayAttempt{
		ID:           "rep-001",
		RequestID:    req.ID,
		TargetURL:    "http://localhost:8000/webhooks/test",
		StatusCode:   500,
		ResponseBody: []byte(`{"error":"database connection failed"}`),
		LatencyMs:    42,
		CreatedAt:    time.Now().UTC().Truncate(time.Millisecond),
	}

	if err := s.SaveReplayAttempt(ctx, attempt); err != nil {
		t.Fatalf("failed to save replay attempt: %v", err)
	}


	attempts, err := s.GetReplayAttempts(ctx, req.ID)
	if err != nil {
		t.Fatalf("failed to get replay attempts: %v", err)
	}
	if len(attempts) != 1 {
		t.Fatalf("expected 1 replay attempt, got %d", len(attempts))
	}
	if attempts[0].StatusCode != 500 {
		t.Errorf("expected status 500, got %d", attempts[0].StatusCode)
	}
	if attempts[0].LatencyMs != 42 {
		t.Errorf("expected latency 42ms, got %d", attempts[0].LatencyMs)
	}
	if !bytes.Equal(attempts[0].ResponseBody, attempt.ResponseBody) {
		t.Errorf("expected response body %s, got %s", attempt.ResponseBody, attempts[0].ResponseBody)
	}

}