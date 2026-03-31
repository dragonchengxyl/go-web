package eventbus

import (
	"context"

	"github.com/studio/platform/internal/infra/streams"
)

type HandlerFunc = streams.HandlerFunc

type Publisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
	Close() error
}

type Consumer interface {
	Start(ctx context.Context, group string, handler HandlerFunc) error
	Close() error
}
