package eventbus

import "context"

type NoopPublisher struct{}

func (p *NoopPublisher) Publish(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func (p *NoopPublisher) Close() error {
	return nil
}
