package replay

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"hooklet/internal/model"
	"hooklet/internal/store"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Dispatcher struct {
	store  store.Store
	client HTTPClient
}

func NewDispatcher(s store.Store, client HTTPClient)*Dispatcher{
    if client == nil {
		client = http.DefaultClient
	}
	return &Dispatcher{
		store: s,
		client: client,
	}
}

func (d *Dispatcher) Replay(ctx context.Context, req *model.WebhookRequest, targetURL string)(*model.ReplayAttempt, error){
      start := time.Now()

	  httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bytes.NewReader(req.Body))

	  if err != nil {
		return nil, fmt.Errorf("create replay request: %v", err)
	  }

	  for key, values := range req.Headers {
		lowerKey := strings.ToLower(key)
		if lowerKey == "host" || lowerKey == "content-length" || lowerKey == "connection" {
			continue
		}
		for _, v := range values {
			httpReq.Header.Add(key, v)
		}
	}
	httpReq.Header.Set("X-Hooklet-Replayed", "true")

	resp, err := d.client.Do(httpReq)
	
	attempt := &model.ReplayAttempt{
		ID:        generateID(),
		RequestID: req.ID,
		TargetURL: targetURL,
		CreatedAt: time.Now().UTC(),
	}

 	attempt.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		attempt.Error = err.Error()
		attempt.StatusCode = 0
		_ = d.store.SaveReplayAttempt(ctx, attempt)
		return attempt, err
	}
	defer resp.Body.Close()
	attempt.StatusCode = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	attempt.ResponseBody = body
 	if err := d.store.SaveReplayAttempt(ctx, attempt); err != nil {
		return attempt, fmt.Errorf("save replay attempt: %w", err)
	}
	return attempt, nil

}

func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("rep-%x-%x-%x", b[0:4], b[4:6], b[6:10])
}
