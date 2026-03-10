package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 4096

	// MaxConnsPerUser is the maximum number of concurrent WebSocket connections per user.
	MaxConnsPerUser = 5

	// rateBurst is the max burst of incoming messages allowed.
	rateBurst = 5
	// rateInterval is the refill interval for the token bucket (1 token per rateInterval).
	rateInterval = 200 * time.Millisecond
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to allowed origins in production
		return true
	},
}

// tokenBucket is a simple in-memory token bucket for rate limiting.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   int
	maxTokens int
	lastRefill time.Time
	interval  time.Duration
}

func newTokenBucket(burst int, interval time.Duration) *tokenBucket {
	return &tokenBucket{
		tokens:    burst,
		maxTokens: burst,
		lastRefill: time.Now(),
		interval:  interval,
	}
}

// Allow returns true if a token is available and consumes it.
func (tb *tokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	refill := int(elapsed / tb.interval)
	if refill > 0 {
		tb.tokens += refill
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefill = now
	}

	if tb.tokens <= 0 {
		return false
	}
	tb.tokens--
	return true
}

// Client represents a single WebSocket connection
type Client struct {
	hub    HubInterface
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	logger *zap.Logger
}

// NewClient creates a new WebSocket client and starts its pumps.
// Returns an error if the user has reached MaxConnsPerUser concurrent connections.
func NewClient(hub HubInterface, w http.ResponseWriter, r *http.Request, userID uuid.UUID, logger *zap.Logger) error {
	if hub.ConnCount(userID) >= MaxConnsPerUser {
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return nil
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	c := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: logger,
	}

	hub.Register(c)
	go c.writePump()
	go c.readPump()
	return nil
}

// readPump reads incoming messages from the WebSocket connection.
// A token bucket enforces a max of rateBurst messages per (rateBurst * rateInterval) period.
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
		return nil
	})

	limiter := newTokenBucket(rateBurst, rateInterval)

	for {
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Debug("ws read error", zap.Error(err))
			}
			break
		}

		// Rate limit: drop message and close if over budget
		if !limiter.Allow() {
			c.logger.Warn("ws rate limit exceeded, closing connection", zap.String("user_id", c.userID.String()))
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "rate limit exceeded")) //nolint:errcheck
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			continue
		}

		// Handle ping
		if msg.Type == MessageTypePing {
			pong := WSMessage{Type: MessageTypePong}
			data, _ := json.Marshal(pong)
			select {
			case c.send <- data:
			default:
			}
		}
		// Other message types handled by server-side logic (chat send is via REST)
	}
}

// writePump sends messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
