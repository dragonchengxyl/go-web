package eventspec

import (
	"strings"

	"github.com/studio/platform/configs"
)

const (
	TopicContent = "content"
	TopicSocial  = "social"
	TopicAudio   = "audio"
	TopicDLQ     = "dlq"
)

func TopicForEvent(eventType string) string {
	switch eventType {
	case "post.created", "post.moderated":
		return TopicContent
	case "audio.job.created":
		return TopicAudio
	case "post.liked", "user.followed", "tip.sent", "comment.created":
		return TopicSocial
	default:
		return TopicSocial
	}
}

func KafkaTopicName(cfg configs.KafkaConfig, logicalTopic string) string {
	base := strings.TrimSpace(cfg.Topic)
	if base == "" {
		base = "furry-events"
	}
	if logicalTopic == "" {
		return base
	}
	return base + "." + logicalTopic
}

func TopicsForConsumerGroup(cfg configs.KafkaConfig, group string) []string {
	switch group {
	case "moderation-group":
		return []string{KafkaTopicName(cfg, TopicContent)}
	case "notification-group":
		return []string{
			KafkaTopicName(cfg, TopicSocial),
			KafkaTopicName(cfg, TopicContent),
		}
	case "audio-job-group":
		return []string{KafkaTopicName(cfg, TopicAudio)}
	default:
		return []string{KafkaTopicName(cfg, TopicSocial)}
	}
}
