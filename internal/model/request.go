package model


import (
	"time"
)

type HeaderMap map[string][]string

type WebhookRequest struct {
	ID            string    `json:"id"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	QueryString   string    `json:"query_string"`
	Headers       HeaderMap `json:"headers"`
	Body          []byte    `json:"body"`
	ContentType   string    `json:"content_type"`
	ContentLength int64     `json:"content_length"`
	RemoteAddr    string    `json:"remote_addr"`
	CreatedAt     time.Time `json:"created_at"`
	ReplayAttempts []ReplayAttempt `json:"replay_attempts,omitempty"` 
}

type ReplayAttempt struct {
	ID           string    `json:"id"`
	RequestID    string    `json:"request_id"`
	TargetURL    string    `json:"target_url"`
	StatusCode   int       `json:"status_code"`
	ResponseBody []byte    `json:"response_body"`
	LatencyMs    int64     `json:"latency_ms"`
	Error        string    `json:"error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

