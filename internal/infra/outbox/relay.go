package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/observability/eventbusmetrics"
	"go.uber.org/zap"
)

type deadLetterPublisher interface {
	PublishDeadLetter(ctx context.Context, originalTopic, eventType string, payload json.RawMessage, errMsg string) error
}

type Event struct {
	ID            uuid.UUID
	SourceTable   string
	AggregateType string
	AggregateID   *uuid.UUID
	EventType     string
	Topic         string
	PartitionKey  string
	Payload       json.RawMessage
	AttemptCount  int
}

type Relay struct {
	pool         *pgxpool.Pool
	publisher    eventbus.Publisher
	logger       *zap.Logger
	batchSize    int
	pollInterval time.Duration
	maxAttempts  int
}

func NewRelay(pool *pgxpool.Pool, publisher eventbus.Publisher, logger *zap.Logger, batchSize int, pollInterval time.Duration, maxAttempts int) *Relay {
	if batchSize <= 0 {
		batchSize = 100
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Relay{
		pool:         pool,
		publisher:    publisher,
		logger:       logger,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		maxAttempts:  maxAttempts,
	}
}

func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil && ctx.Err() == nil {
				r.logger.Error("outbox relay batch failed", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin outbox batch transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx, `
		WITH claimed AS (
			SELECT id
			FROM outbox_events
			WHERE status = 'pending'
			  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_events o
		SET status = 'processing',
		    updated_at = NOW()
		FROM claimed
		WHERE o.id = claimed.id
		RETURNING o.id, o.source_table, o.aggregate_type, o.aggregate_id, o.event_type, o.topic,
		          o.partition_key, o.payload, o.attempt_count
	`, r.batchSize)
	if err != nil {
		return fmt.Errorf("claim outbox events: %w", err)
	}
	defer rows.Close()

	events := make([]Event, 0, r.batchSize)
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID,
			&event.SourceTable,
			&event.AggregateType,
			&event.AggregateID,
			&event.EventType,
			&event.Topic,
			&event.PartitionKey,
			&event.Payload,
			&event.AttemptCount,
		); err != nil {
			return fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate outbox events: %w", err)
	}

	if len(events) == 0 {
		return tx.Commit(ctx)
	}

	for _, event := range events {
		if err := r.publisher.PublishEvent(ctx, eventbus.Event{
			EventID: event.ID.String(),
			Type:    event.EventType,
			Payload: event.Payload,
		}); err != nil {
			eventbusmetrics.RecordOutboxDispatch("relay", event.Topic, event.EventType, "error")
			if markErr := r.markFailed(ctx, tx, event, err); markErr != nil {
				return fmt.Errorf("mark outbox failure: %w", markErr)
			}
			r.logger.Error("outbox publish failed",
				zap.Error(err),
				zap.String("event_id", event.ID.String()),
				zap.String("event_type", event.EventType),
				zap.String("topic", event.Topic),
			)
			continue
		}

		if _, err := tx.Exec(ctx, `
			UPDATE outbox_events
			SET status = 'published',
			    published_at = NOW(),
			    updated_at = NOW(),
			    last_error = NULL
			WHERE id = $1
		`, event.ID); err != nil {
			return fmt.Errorf("mark outbox published: %w", err)
		}
		eventbusmetrics.RecordOutboxDispatch("relay", event.Topic, event.EventType, "ok")
	}

	return tx.Commit(ctx)
}

func (r *Relay) markFailed(ctx context.Context, tx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, event Event, publishErr error) error {
	nextAttempt := event.AttemptCount + 1
	status := "pending"
	var retryAt *time.Time
	if nextAttempt >= r.maxAttempts {
		status = "dead_lettered"
		if dlqPublisher, ok := r.publisher.(deadLetterPublisher); ok {
			if dlqErr := dlqPublisher.PublishDeadLetter(ctx, event.Topic, event.EventType, event.Payload, truncateErr(publishErr)); dlqErr != nil {
				r.logger.Error("outbox dead-letter publish failed",
					zap.Error(dlqErr),
					zap.String("event_id", event.ID.String()),
					zap.String("event_type", event.EventType),
				)
			}
		}
	} else {
		backoff := time.Duration(nextAttempt*nextAttempt) * time.Second
		at := time.Now().Add(backoff)
		retryAt = &at
	}

	_, err := tx.Exec(ctx, `
		UPDATE outbox_events
		SET status = $2,
		    attempt_count = attempt_count + 1,
		    last_error = $3,
		    next_retry_at = $4,
		    updated_at = NOW()
		WHERE id = $1
	`, event.ID, status, truncateErr(publishErr), retryAt)
	return err
}

func truncateErr(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if len(text) > 1000 {
		return text[:1000]
	}
	return text
}
