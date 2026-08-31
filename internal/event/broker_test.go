package event_test

import (
	"hooklet/internal/event"
	"hooklet/internal/model"
	"testing"
	"time"
)

func TestBroker_SubscribeAndPublish(t *testing.T){
	 broker := event.NewBroker()

	 ch1 := broker.Subscribe()
	 ch2 := broker.Subscribe()

	 req1 := &model.WebhookRequest{
		ID: "req-sse-1",
		Method: "POST",
		Path: "/wh/stripe",
	 }

	 broker.Publish(req1)

	 select{
	 case received := <-ch1:
		if received.ID != req1.ID {
			t.Errorf("ch1 expected ID: %v, but got %v", req1.ID, received.ID)
		 }
		case <- time.After(100 * time.Millisecond):
			t.Fatal("ch1 timed out waiting for event")
	 }

	 select {
		case received := <-ch2:
			if received.ID != req1.ID {
				t.Errorf("ch2 expected ID %s, got %s", req1.ID, received.ID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("ch2 timed out waiting for event")
	 }

	 broker.Unsubscribe(ch1)
    
	 req2 := &model.WebhookRequest{
		ID:     "req-sse-2",
		Method: "POST",
		Path:   "/wh/github",
	}
	broker.Publish(req2)

	select {
	case received := <-ch2:
		if received.ID != req2.ID {
			t.Errorf("ch2 expected ID %s, got %s", req2.ID, received.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch2 timed out waiting for event")
	}

	select {
	case msg, ok := <-ch1:
		if ok {
			t.Fatalf("ch1 received event after unsubscribing: %v", msg)
		}
	default:
		// Expected: nothing received on ch1
	}


}