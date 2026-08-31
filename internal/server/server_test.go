package server_test

import (
	"context"
	"encoding/json"
	"hooklet/internal/event"
	"hooklet/internal/model"
	"hooklet/internal/replay"
	"hooklet/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type FakeServerStore struct {
	requests map[string]*model.WebhookRequest
	replays  map[string][]model.ReplayAttempt
}


func newFakeStore() *FakeServerStore {
	return &FakeServerStore{
		requests: make(map[string]*model.WebhookRequest),
		replays:  make(map[string][]model.ReplayAttempt),
	}
}

func (f *FakeServerStore) Close() error { return nil }

func (f *FakeServerStore) SaveRequest(ctx context.Context, req *model.WebhookRequest) error {
	f.requests[req.ID] = req
	return nil
}

func (f *FakeServerStore) GetRequest(ctx context.Context, id string)(*model.WebhookRequest, error){
	req, ok := f.requests[id]

	if !ok{
		return nil, nil
	}

	clone := *req
	clone.ReplayAttempts = f.replays[id]
	return &clone, nil
}

func (f *FakeServerStore) ListRequests(ctx context.Context, limit int, offset int)([]*model.WebhookRequest, error){
	var list []*model.WebhookRequest
	for _, req := range f.requests{
		list = append(list, req)
	}
	return list, nil
}

func(f *FakeServerStore) SaveReplayAttempt(ctx context.Context, attempt *model.ReplayAttempt)error {
	f.replays[attempt.RequestID] = append(f.replays[attempt.RequestID], *attempt)
	return nil
}

func(f *FakeServerStore) GetReplayAttempts(ctx context.Context, requestID string)([]model.ReplayAttempt, error){
	return f.replays[requestID], nil
}

func TestServer_ListRequestsAPI(t *testing.T) {
	store := newFakeStore()
	broker := event.NewBroker()
	dispatcher := replay.NewDispatcher(store, http.DefaultClient)
	_ = store.SaveRequest(context.Background(), &model.WebhookRequest{
		ID:        "req-api-1",
		Method:    "POST",
		Path:      "/wh/stripe",
		CreatedAt: time.Now().UTC(),
	})
	srv := server.New(store, broker, dispatcher, "http://localhost:8000")
	req := httptest.NewRequest(http.MethodGet, "/api/requests", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var returned []*model.WebhookRequest
	if err := json.Unmarshal(rec.Body.Bytes(), &returned); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(returned) != 1 || returned[0].ID != "req-api-1" {
		t.Errorf("unexpected requests returned: %v", returned)
	}
}
func TestServer_ReplayAPI(t *testing.T) {
	fakeTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer fakeTarget.Close()
 	store := newFakeStore()
	broker := event.NewBroker()
	dispatcher := replay.NewDispatcher(store, http.DefaultClient)
	orig := &model.WebhookRequest{
		ID:        "req-to-replay",
		Method:    "POST",
		Path:      "/wh/stripe",
		Body:      []byte(`{"test":123}`),
		CreatedAt: time.Now().UTC(),
	}
	_ = store.SaveRequest(context.Background(), orig)
	srv := server.New(store, broker, dispatcher, fakeTarget.URL)
 	req := httptest.NewRequest(http.MethodPost, "/api/requests/req-to-replay/replay", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
 	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var attempt model.ReplayAttempt
	if err := json.Unmarshal(rec.Body.Bytes(), &attempt); err != nil {
		t.Fatalf("failed to unmarshal replay attempt: %v", err)
	}
	if attempt.StatusCode != http.StatusOK {
		t.Errorf("expected attempt status 200, got %d", attempt.StatusCode)
	}
}