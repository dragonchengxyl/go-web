package kafkaevent

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/studio/platform/configs"
)

func CheckConnectivity(ctx context.Context, cfg configs.KafkaConfig) error {
	if !cfg.Enabled {
		return nil
	}

	brokers := parseBrokers(cfg.Brokers)
	if len(brokers) == 0 {
		return fmt.Errorf("kafka enabled but no brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", brokers[0], err)
	}
	defer conn.Close()

	if _, err := conn.Brokers(); err != nil {
		return fmt.Errorf("fetch kafka brokers: %w", err)
	}

	return nil
}

func TopicLabel(cfg configs.KafkaConfig) string {
	topic := strings.TrimSpace(cfg.Topic)
	if topic == "" {
		return "furry-events"
	}
	return topic
}
