package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const wsChannelPrefix = "ws:user:"

// DistributedHub wraps a local Hub with Redis Pub/Sub for cross-node message routing.
// When SendToUser is called, the message is published to a Redis channel. All nodes
// (including the publisher) subscribe and deliver messages to their local clients.
type DistributedHub struct {
	local  *Hub
	rdb    *redis.Client
	logger *zap.Logger
}

// NewDistributedHub creates a hub backed by Redis Pub/Sub for horizontal scaling.
func NewDistributedHub(rdb *redis.Client, logger *zap.Logger) *DistributedHub {
	return &DistributedHub{
		local:  NewHub(logger),
		rdb:    rdb,
		logger: logger,
	}
}

// Run starts the local hub event loop and a Redis subscriber goroutine.
func (d *DistributedHub) Run(ctx context.Context) {
	go d.local.Run(ctx)
	go d.subscribe(ctx)
	<-ctx.Done()
}

// subscribe listens on Redis pattern ws:user:* and delivers messages to local clients.
func (d *DistributedHub) subscribe(ctx context.Context) {
	pubsub := d.rdb.PSubscribe(ctx, wsChannelPrefix+"*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case redisMsg, ok := <-ch:
			if !ok {
				return
			}
			var envelope struct {
				UserID  string    `json:"user_id"`
				Message WSMessage `json:"message"`
			}
			if err := json.Unmarshal([]byte(redisMsg.Payload), &envelope); err != nil {
				d.logger.Error("distributed_hub: failed to unmarshal redis message", zap.Error(err))
				continue
			}
			uid, err := uuid.Parse(envelope.UserID)
			if err != nil {
				d.logger.Error("distributed_hub: failed to parse user_id", zap.Error(err))
				continue
			}
			d.local.deliverToUser(uid, envelope.Message)
		}
	}
}

// SendToUser publishes the message to the Redis channel for the target user.
// All nodes subscribing to that channel will deliver it to their local clients.
func (d *DistributedHub) SendToUser(userID uuid.UUID, msg WSMessage) {
	envelope := struct {
		UserID  string    `json:"user_id"`
		Message WSMessage `json:"message"`
	}{
		UserID:  userID.String(),
		Message: msg,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		d.logger.Error("distributed_hub: failed to marshal message", zap.Error(err))
		return
	}
	channel := fmt.Sprintf("%s%s", wsChannelPrefix, userID.String())
	if err := d.rdb.Publish(context.Background(), channel, data).Err(); err != nil {
		d.logger.Error("distributed_hub: redis publish failed", zap.Error(err))
	}
}

// Register delegates to the local hub.
func (d *DistributedHub) Register(c *Client) {
	d.local.Register(c)
}

// Unregister delegates to the local hub.
func (d *DistributedHub) Unregister(c *Client) {
	d.local.Unregister(c)
}

// ConnCount returns the number of local connections for a user.
func (d *DistributedHub) ConnCount(userID uuid.UUID) int {
	return d.local.ConnCount(userID)
}
