package store

import (
	"context"
	"errors"

	"hooklet/internal/model"
)

var ErrNotFound = errors.New("record not found")

type Store interface {
	Close() error
	SaveRequest(ctx context.Context, req *model.WebhookRequest) error
	GetRequest(ctx context.Context, id string) (*model.WebhookRequest, error)
	ListRequests(ctx context.Context, limit int, offset int) ([]*model.WebhookRequest, error)
	SaveReplayAttempt(ctx context.Context, attempt *model.ReplayAttempt) error
	GetReplayAttempts(ctx context.Context, requestID string) ([]model.ReplayAttempt, error)
}

 