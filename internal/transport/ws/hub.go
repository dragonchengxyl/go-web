package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HubInterface abstracts the WebSocket hub for single-node and distributed modes.
type HubInterface interface {
	SendToUser(userID uuid.UUID, msg WSMessage)
	Register(c *Client)
	Unregister(c *Client)
	Run(ctx context.Context)
	// ConnCount returns the number of active connections for a user (used for max-conn enforcement).
	ConnCount(userID uuid.UUID) int
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeChat         MessageType = "chat"
	MessageTypeNotification MessageType = "notification"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// WSMessage is a WebSocket message envelope
type WSMessage struct {
	Type           MessageType `json:"type"`
	ConversationID *uuid.UUID  `json:"conversation_id,omitempty"`
	Payload        any         `json:"payload"`
}

// Hub manages WebSocket client connections and message routing
type Hub struct {
	// clients maps userID -> set of clients
	clients map[uuid.UUID]map[*Client]struct{}
	mu      sync.RWMutex

	register   chan *Client
	unregister chan *Client
	broadcast  chan *broadcastMsg

	logger *zap.Logger
}

type broadcastMsg struct {
	toUserID uuid.UUID
	msg      WSMessage
}

// NewHub creates a new WebSocket Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan *broadcastMsg, 256),
		logger:     logger,
	}
}

// Run starts the hub event loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case msg := <-h.broadcast:
			h.deliverToUser(msg.toUserID, msg.msg)
		}
	}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*Client]struct{})
	}
	h.clients[c.userID][c] = struct{}{}
	h.logger.Debug("ws client registered", zap.String("user_id", c.userID.String()))
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.userID)
		}
	}
	close(c.send)
	h.logger.Debug("ws client unregistered", zap.String("user_id", c.userID.String()))
}

func (h *Hub) deliverToUser(userID uuid.UUID, msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userID]
	if !ok {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal ws message", zap.Error(err))
		return
	}
	for c := range conns {
		select {
		case c.send <- data:
		default:
			// client buffer full — will be disconnected by write pump
		}
	}
}

// SendToUser sends a message to all connections of a user
func (h *Hub) SendToUser(userID uuid.UUID, msg WSMessage) {
	h.broadcast <- &broadcastMsg{toUserID: userID, msg: msg}
}

// Register registers a client with the hub
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// ConnCount returns the number of active connections for a user.
func (h *Hub) ConnCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID])
}
