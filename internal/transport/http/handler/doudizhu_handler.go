package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/doudizhu"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/pkg/apperr"
	cryptopkg "github.com/studio/platform/internal/pkg/crypto"
	"github.com/studio/platform/internal/pkg/response"
	wsbridge "github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

type DoudizhuHandler struct {
	service    *usecase.DoudizhuRoomService
	jwtConfig  configs.JWTConfig
	tokenStore *redis.TokenStore
	logger     *zap.Logger

	mu    sync.RWMutex
	rooms map[uuid.UUID]map[*doudizhuWSClient]struct{}
}

type doudizhuWSClient struct {
	handler   *DoudizhuHandler
	conn      *websocket.Conn
	send      chan []byte
	roomID    uuid.UUID
	sessionID string
}

type doudizhuIncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func NewDoudizhuHandler(
	service *usecase.DoudizhuRoomService,
	jwtConfig configs.JWTConfig,
	tokenStore *redis.TokenStore,
	logger *zap.Logger,
) *DoudizhuHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	h := &DoudizhuHandler{
		service:    service,
		jwtConfig:  jwtConfig,
		tokenStore: tokenStore,
		logger:     logger,
		rooms:      make(map[uuid.UUID]map[*doudizhuWSClient]struct{}),
	}
	if service != nil {
		service.SetNotifier(h.broadcastRoomState)
	}
	return h
}

func (h *DoudizhuHandler) ListRooms(c *gin.Context) {
	response.Success(c, gin.H{"rooms": h.service.ListRooms()})
}

func (h *DoudizhuHandler) GetRoom(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的房间 ID"))
		return
	}
	room, ok := h.service.GetRoom(roomID)
	if !ok {
		response.Error(c, apperr.ErrNotFound)
		return
	}
	response.Success(c, room)
}

func (h *DoudizhuHandler) CreateRoom(c *gin.Context) {
	var req struct {
		Title      string `json:"title"`
		PlayerName string `json:"player_name"`
		SessionID  string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("请求参数错误"))
		return
	}

	var userID *uuid.UUID
	if parsed := middlewareUserID(c); parsed != nil {
		userID = parsed
	}
	room, err := h.service.CreateRoom(usecase.CreateDoudizhuRoomInput{
		SessionID:  req.SessionID,
		PlayerName: req.PlayerName,
		Title:      req.Title,
		UserID:     userID,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, room)
}

func (h *DoudizhuHandler) CreateDemoRoom(c *gin.Context) {
	var req struct {
		Title      string `json:"title"`
		PlayerName string `json:"player_name"`
		SessionID  string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("请求参数错误"))
		return
	}

	var userID *uuid.UUID
	if parsed := middlewareUserID(c); parsed != nil {
		userID = parsed
	}
	room, err := h.service.CreateDemoRoom(usecase.CreateDoudizhuRoomInput{
		SessionID:  req.SessionID,
		PlayerName: req.PlayerName,
		Title:      req.Title,
		UserID:     userID,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, room)
}

func (h *DoudizhuHandler) ServeWS(c *gin.Context) {
	roomID, err := uuid.Parse(c.Query("room_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id"})
		return
	}
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	playerName := c.Query("player_name")

	var userID *uuid.UUID
	if token := c.Query("token"); token != "" {
		resolvedUserID, resolveErr := h.resolveUserIDFromToken(c, token)
		if resolveErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		userID = resolvedUserID
	}

	conn, err := wsbridge.Upgrade(c.Writer, c.Request)
	if err != nil {
		h.logger.Error("failed to upgrade doudizhu websocket", zap.Error(err))
		return
	}

	room, err := h.service.JoinRoom(usecase.JoinDoudizhuRoomInput{
		RoomID:     roomID,
		SessionID:  sessionID,
		PlayerName: playerName,
		UserID:     userID,
	})
	if err != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "payload": gin.H{"message": err.Error()}})
		_ = conn.Close()
		return
	}

	client := &doudizhuWSClient{
		handler:   h,
		conn:      conn,
		send:      make(chan []byte, 64),
		roomID:    roomID,
		sessionID: sessionID,
	}
	h.registerClient(client)

	privateState, _ := h.service.GetPrivateState(roomID, sessionID)
	h.sendToClient(client, "joined", gin.H{
		"session_id":    sessionID,
		"room":          room,
		"private_state": privateState,
	})
	h.broadcastRoomState(roomID)

	go client.writePump()
	go client.readPump()
}

func (h *DoudizhuHandler) resolveUserIDFromToken(c *gin.Context, token string) (*uuid.UUID, error) {
	claims, err := cryptopkg.VerifyToken(token, h.jwtConfig.Secret)
	if err != nil {
		return nil, err
	}
	if h.tokenStore != nil {
		blacklisted, blacklistErr := h.tokenStore.IsTokenBlacklisted(c.Request.Context(), claims.JTI)
		if blacklistErr != nil {
			return nil, blacklistErr
		}
		if blacklisted {
			return nil, apperr.New(apperr.CodeUnauthorized, "令牌已被撤销")
		}
	}
	parsed, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (h *DoudizhuHandler) registerClient(client *doudizhuWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*doudizhuWSClient]struct{})
	}
	h.rooms[client.roomID][client] = struct{}{}
}

func (h *DoudizhuHandler) unregisterClient(client *doudizhuWSClient) {
	h.mu.Lock()
	if roomClients, ok := h.rooms[client.roomID]; ok {
		delete(roomClients, client)
		if len(roomClients) == 0 {
			delete(h.rooms, client.roomID)
		}
	}
	h.mu.Unlock()
	close(client.send)
}

func (h *DoudizhuHandler) roomClients(roomID uuid.UUID) []*doudizhuWSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	roomClients := h.rooms[roomID]
	clients := make([]*doudizhuWSClient, 0, len(roomClients))
	for client := range roomClients {
		clients = append(clients, client)
	}
	return clients
}

func (h *DoudizhuHandler) broadcastRoomState(roomID uuid.UUID) {
	room, ok := h.service.GetRoom(roomID)
	if !ok {
		h.closeRoomClients(roomID)
		return
	}
	clients := h.roomClients(roomID)
	h.broadcastToClients(clients, "room_state", room)
	for _, client := range clients {
		if privateState, exists := h.service.GetPrivateState(roomID, client.sessionID); exists {
			h.sendToClient(client, "private_state", privateState)
		}
	}
}

func (h *DoudizhuHandler) closeRoomClients(roomID uuid.UUID) {
	clients := h.roomClients(roomID)
	for _, client := range clients {
		h.sendToClient(client, "room_closed", gin.H{"room_id": roomID.String()})
		_ = client.conn.Close()
	}
}

func (h *DoudizhuHandler) broadcastToClients(clients []*doudizhuWSClient, messageType string, payload any) {
	data, err := json.Marshal(gin.H{
		"type":    messageType,
		"payload": payload,
	})
	if err != nil {
		h.logger.Error("failed to marshal doudizhu ws message", zap.Error(err))
		return
	}
	for _, client := range clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

func (h *DoudizhuHandler) sendToClient(client *doudizhuWSClient, messageType string, payload any) {
	data, err := json.Marshal(gin.H{
		"type":    messageType,
		"payload": payload,
	})
	if err != nil {
		return
	}
	select {
	case client.send <- data:
	default:
	}
}

func (c *doudizhuWSClient) readPump() {
	leftRoom := false
	defer func() {
		c.handler.unregisterClient(c)
		if !leftRoom {
			c.handler.service.DisconnectSession(c.sessionID)
		}
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(gameMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(gamePongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(gamePongWait))
	})

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var message doudizhuIncomingMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			c.handler.sendToClient(c, "error", gin.H{"message": "消息格式错误"})
			continue
		}

		switch message.Type {
		case "ping":
			c.handler.sendToClient(c, "pong", gin.H{"ts": time.Now().UnixMilli()})
		case "set_ready":
			var payload struct {
				Ready *bool `json:"ready"`
			}
			if len(message.Payload) > 0 {
				_ = json.Unmarshal(message.Payload, &payload)
			}
			if _, err := c.handler.service.SetReady(usecase.SetDoudizhuReadyInput{
				RoomID:    c.roomID,
				SessionID: c.sessionID,
				Ready:     payload.Ready,
			}); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "start_round":
			if _, err := c.handler.service.StartRound(c.roomID, c.sessionID); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "bid":
			var payload struct {
				Score int `json:"score"`
			}
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": "叫分消息错误"})
				continue
			}
			result, err := c.handler.service.Bid(usecase.DoudizhuBidInput{
				RoomID:    c.roomID,
				SessionID: c.sessionID,
				Score:     payload.Score,
			})
			if err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
				continue
			}
			c.handler.sendToClient(c, "action_result", result)
		case "pass_bid":
			result, err := c.handler.service.PassBid(c.roomID, c.sessionID)
			if err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
				continue
			}
			c.handler.sendToClient(c, "action_result", result)
		case "play_cards":
			var payload struct {
				Cards []doudizhu.Card `json:"cards"`
			}
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": "出牌消息错误"})
				continue
			}
			result, err := c.handler.service.PlayCards(usecase.DoudizhuPlayInput{
				RoomID:    c.roomID,
				SessionID: c.sessionID,
				Cards:     payload.Cards,
			})
			if err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
				continue
			}
			c.handler.sendToClient(c, "action_result", result)
		case "pass_turn":
			result, err := c.handler.service.PassTurn(c.roomID, c.sessionID)
			if err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
				continue
			}
			c.handler.sendToClient(c, "action_result", result)
		case "toggle_auto_play":
			var payload struct {
				Enabled *bool `json:"enabled"`
			}
			if len(message.Payload) > 0 {
				_ = json.Unmarshal(message.Payload, &payload)
			}
			if _, err := c.handler.service.ToggleAutoPlay(usecase.ToggleDoudizhuAutoPlayInput{
				RoomID:    c.roomID,
				SessionID: c.sessionID,
				Enabled:   payload.Enabled,
			}); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "leave_room":
			if _, err := c.handler.service.LeaveRoom(c.roomID, c.sessionID); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
				continue
			}
			leftRoom = true
			_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "left room"))
			_ = c.conn.Close()
			return
		default:
			c.handler.sendToClient(c, "error", gin.H{"message": "不支持的消息类型"})
		}
	}
}

func (c *doudizhuWSClient) writePump() {
	ticker := time.NewTicker(gamePingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(gameWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(gameWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
