package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/observability/gamemetrics"
	"github.com/studio/platform/internal/pkg/apperr"
	cryptopkg "github.com/studio/platform/internal/pkg/crypto"
	"github.com/studio/platform/internal/pkg/response"
	wsbridge "github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

const (
	gameWriteWait      = 10 * time.Second
	gamePongWait       = 60 * time.Second
	gamePingPeriod     = 50 * time.Second
	gameMaxMessageSize = 4096
)

type GameHandler struct {
	service    *usecase.HexBlitzRoomService
	jwtConfig  configs.JWTConfig
	tokenStore *redis.TokenStore
	logger     *zap.Logger

	mu    sync.RWMutex
	rooms map[uuid.UUID]map[*gameWSClient]struct{}
}

type gameWSClient struct {
	handler   *GameHandler
	conn      *websocket.Conn
	send      chan []byte
	roomID    uuid.UUID
	sessionID string
}

type gameWSIncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func NewGameHandler(
	service *usecase.HexBlitzRoomService,
	jwtConfig configs.JWTConfig,
	tokenStore *redis.TokenStore,
	logger *zap.Logger,
) *GameHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	h := &GameHandler{
		service:    service,
		jwtConfig:  jwtConfig,
		tokenStore: tokenStore,
		logger:     logger,
		rooms:      make(map[uuid.UUID]map[*gameWSClient]struct{}),
	}
	if service != nil {
		service.SetNotifier(h.broadcastRoomState)
	}
	return h
}

func (h *GameHandler) ListHexBlitzRooms(c *gin.Context) {
	response.Success(c, gin.H{
		"rooms": h.service.ListRooms(),
	})
}

func (h *GameHandler) GetHexBlitzRoom(c *gin.Context) {
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

func (h *GameHandler) CreateHexBlitzRoom(c *gin.Context) {
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

	room, err := h.service.CreateRoom(usecase.CreateHexBlitzRoomInput{
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

func (h *GameHandler) ListHexBlitzLeaderboard(c *gin.Context) {
	limit := 10
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.ListLeaderboard(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"entries": items})
}

func (h *GameHandler) ListHexBlitzRecentMatches(c *gin.Context) {
	limit := 8
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.ListRecentMatches(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"matches": items})
}

func (h *GameHandler) ListMyHexBlitzRecentMatches(c *gin.Context) {
	userID := middlewareUserID(c)
	if userID == nil {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	limit := 8
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.ListUserRecentMatches(c.Request.Context(), *userID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"matches": items})
}

func (h *GameHandler) ServeHexBlitzWS(c *gin.Context) {
	roomID, err := uuid.Parse(c.Query("room_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id"})
		return
	}

	sessionID := strings.TrimSpace(c.Query("session_id"))
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	playerName := strings.TrimSpace(c.Query("player_name"))
	var userID *uuid.UUID
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		resolvedUserID, resolveErr := h.resolveUserIDFromToken(c, token)
		if resolveErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		userID = resolvedUserID
	}

	conn, err := wsbridge.Upgrade(c.Writer, c.Request)
	if err != nil {
		h.logger.Error("failed to upgrade game websocket", zap.Error(err))
		return
	}

	room, err := h.service.JoinRoom(usecase.JoinHexBlitzRoomInput{
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

	client := &gameWSClient{
		handler:   h,
		conn:      conn,
		send:      make(chan []byte, 64),
		roomID:    roomID,
		sessionID: sessionID,
	}

	h.registerClient(client)
	h.sendToClient(client, "joined", gin.H{
		"session_id": sessionID,
		"room":       room,
	})
	h.broadcastRoomState(roomID)

	go client.writePump()
	go client.readPump()
}

func (h *GameHandler) resolveUserIDFromToken(c *gin.Context, token string) (*uuid.UUID, error) {
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

func (h *GameHandler) registerClient(client *gameWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*gameWSClient]struct{})
	}
	h.rooms[client.roomID][client] = struct{}{}
	gamemetrics.SetActiveConnections(h.connectionCountLocked())
}

func (h *GameHandler) unregisterClient(client *gameWSClient) {
	h.mu.Lock()
	if roomClients, ok := h.rooms[client.roomID]; ok {
		delete(roomClients, client)
		if len(roomClients) == 0 {
			delete(h.rooms, client.roomID)
		}
	}
	connectionCount := h.connectionCountLocked()
	h.mu.Unlock()
	close(client.send)
	gamemetrics.SetActiveConnections(connectionCount)
}

func (h *GameHandler) connectionCountLocked() int {
	total := 0
	for _, roomClients := range h.rooms {
		total += len(roomClients)
	}
	return total
}

func (h *GameHandler) broadcastRoomState(roomID uuid.UUID) {
	room, ok := h.service.GetRoom(roomID)
	if !ok {
		h.closeRoomClients(roomID)
		return
	}
	clients := h.roomClients(roomID)
	h.broadcastToClients(clients, "room_state", room)
	for _, client := range clients {
		if board, exists := h.service.GetPlayerBoardState(roomID, client.sessionID); exists {
			h.sendToClient(client, "board_state", board)
		}
	}
}

func (h *GameHandler) closeRoomClients(roomID uuid.UUID) {
	clients := h.roomClients(roomID)

	for _, client := range clients {
		h.sendToClient(client, "room_closed", gin.H{"room_id": roomID.String()})
		_ = client.conn.Close()
	}
}

func (h *GameHandler) broadcastToRoom(roomID uuid.UUID, messageType string, payload any) {
	h.broadcastToClients(h.roomClients(roomID), messageType, payload)
}

func (h *GameHandler) broadcastToClients(clients []*gameWSClient, messageType string, payload any) {
	data, err := json.Marshal(gin.H{
		"type":    messageType,
		"payload": payload,
	})
	if err != nil {
		h.logger.Error("failed to marshal game ws message", zap.Error(err))
		return
	}
	for _, client := range clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

func (h *GameHandler) roomClients(roomID uuid.UUID) []*gameWSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	roomClients := h.rooms[roomID]
	clients := make([]*gameWSClient, 0, len(roomClients))
	for client := range roomClients {
		clients = append(clients, client)
	}
	return clients
}

func (h *GameHandler) sendToClient(client *gameWSClient, messageType string, payload any) {
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

func (c *gameWSClient) readPump() {
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

		var message gameWSIncomingMessage
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
			if _, err := c.handler.service.SetReady(usecase.SetHexBlitzReadyInput{
				RoomID:    c.roomID,
				SessionID: c.sessionID,
				Ready:     payload.Ready,
			}); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "start_match":
			if _, err := c.handler.service.StartMatch(c.roomID, c.sessionID); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "make_move":
			var payload struct {
				TileID string `json:"tile_id"`
			}
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": "操作消息错误"})
				continue
			}
			if _, err := c.handler.service.ApplyMove(c.roomID, c.sessionID, payload.TileID); err != nil {
				c.handler.sendToClient(c, "error", gin.H{"message": err.Error()})
			}
		case "score_update":
			c.handler.sendToClient(c, "error", gin.H{"message": "score_update 已弃用，请改为发送 make_move"})
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

func (c *gameWSClient) writePump() {
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

func middlewareUserID(c *gin.Context) *uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}
	parsed, err := uuid.Parse(userID.(string))
	if err != nil {
		return nil
	}
	return &parsed
}
