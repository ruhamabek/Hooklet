package event

import (
	"hooklet/internal/model"
	"sync"
)


type Broker struct {
	mu          sync.RWMutex
	subscribers map[chan *model.WebhookRequest]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan *model.WebhookRequest]struct{}),
	}
}

func (b *Broker) Subscribe() chan *model.WebhookRequest {
	  b.mu.Lock()
	  defer b.mu.Unlock()

	  ch := make(chan *model.WebhookRequest, 16)
	  b.subscribers[ch] = struct{}{}
	  return ch
}

func (b *Broker) Unsubscribe(ch chan *model.WebhookRequest){
	b.mu.Lock()
	defer b.mu.Unlock()


	if _, exists := b.subscribers[ch]; exists {
		delete(b.subscribers, ch)
		close(ch)
	}
}

func (b *Broker) Publish(req *model.WebhookRequest) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- req:
		default:
			// Slow consumer: buffer is full, drop to prevent blocking the ingest server
		}
	}
}