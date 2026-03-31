package kafkaevent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/eventspec"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/observability/eventbusmetrics"
	"go.uber.org/zap"
)

type Publisher struct {
	writer *kafka.Writer
	logger *zap.Logger
	cfg    configs.KafkaConfig
}

type deadLetterMessage struct {
	OriginalTopic string          `json:"original_topic"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Error         string          `json:"error"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func NewPublisher(cfg configs.KafkaConfig, logger *zap.Logger) (*Publisher, error) {
	brokers := parseBrokers(cfg.Brokers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka publisher: no brokers configured")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchTimeout: batchTimeout(cfg.BatchTimeoutMS),
	}

	return &Publisher{
		writer: writer,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		eventbusmetrics.RecordPublish("kafka", eventspec.TopicForEvent(eventType), eventType, "error")
		return fmt.Errorf("kafka publisher: marshal payload: %w", err)
	}

	envelope := streams.StreamEvent{
		Type:    eventType,
		Payload: payloadBytes,
	}
	value, err := json.Marshal(envelope)
	if err != nil {
		eventbusmetrics.RecordPublish("kafka", eventspec.TopicForEvent(eventType), eventType, "error")
		return fmt.Errorf("kafka publisher: marshal envelope: %w", err)
	}

	topic := eventspec.KafkaTopicName(p.cfg, eventspec.TopicForEvent(eventType))
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(deriveMessageKey(payload, eventType)),
		Value: value,
		Time:  time.Now().UTC(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		eventbusmetrics.RecordPublish("kafka", topic, eventType, "error")
		return fmt.Errorf("kafka publisher: write message: %w", err)
	}

	eventbusmetrics.RecordPublish("kafka", topic, eventType, "ok")
	return nil
}

func (p *Publisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (p *Publisher) PublishDeadLetter(ctx context.Context, originalTopic, eventType string, payload json.RawMessage, errMsg string) error {
	value, err := json.Marshal(deadLetterMessage{
		OriginalTopic: originalTopic,
		EventType:     eventType,
		Payload:       payload,
		Error:         errMsg,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("kafka publisher: marshal dead letter: %w", err)
	}

	topic := eventspec.KafkaTopicName(p.cfg, eventspec.TopicDLQ)
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(eventType),
		Value: value,
		Time:  time.Now().UTC(),
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publisher: write dead letter: %w", err)
	}
	eventbusmetrics.RecordPublish("kafka", topic, eventType, "dlq")
	return nil
}

func deriveMessageKey(payload interface{}, fallback string) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fallback
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fallback
	}

	for _, key := range []string{
		"job_id",
		"post_id",
		"comment_id",
		"tip_id",
		"followee_id",
		"author_id",
		"receiver_id",
	} {
		if value, ok := obj[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}

	return fallback
}

func batchTimeout(value int) time.Duration {
	if value <= 0 {
		return time.Second
	}
	return time.Duration(value) * time.Millisecond
}

func parseBrokers(value string) []string {
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
