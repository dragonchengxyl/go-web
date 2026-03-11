package streams

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Publisher publishes events to the Redis stream.
type Publisher struct {
	client *redis.Client
}

// NewPublisher creates a new Publisher.
func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

// Publish publishes an event to the furry:events stream.
// payload must be JSON-marshalable.
func (p *Publisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("streams: marshal payload: %w", err)
	}

	ev := StreamEvent{
		Type:    eventType,
		Payload: json.RawMessage(b),
	}

	evBytes, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("streams: marshal event: %w", err)
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamKey,
		MaxLen: 100_000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(evBytes),
		},
	}).Err()
}
