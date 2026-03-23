package usecase

import (
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/gameplay"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

const (
	hexBlitzGameSlug      = "hex-blitz"
	hexBlitzMaxPlayers    = 4
	hexBlitzDefaultTitle  = "Hex Blitz 房间"
	hexBlitzDefaultPlayer = "玩家"
)

type HexBlitzRoomService struct {
	mu            sync.RWMutex
	rooms         map[uuid.UUID]*hexBlitzRoomState
	sessionToRoom map[string]uuid.UUID
	logger        *zap.Logger
	countdown     time.Duration
	roundDuration time.Duration
	notifier      func(uuid.UUID)
}

type HexBlitzRoomServiceOption func(*HexBlitzRoomService)

type CreateHexBlitzRoomInput struct {
	SessionID  string
	PlayerName string
	Title      string
	UserID     *uuid.UUID
}

type JoinHexBlitzRoomInput struct {
	RoomID     uuid.UUID
	SessionID  string
	PlayerName string
	UserID     *uuid.UUID
}

type SetHexBlitzReadyInput struct {
	RoomID    uuid.UUID
	SessionID string
	Ready     *bool
}

type UpdateHexBlitzScoreInput struct {
	RoomID    uuid.UUID
	SessionID string
	Score     int
}

type hexBlitzRoomState struct {
	id               uuid.UUID
	code             string
	title            string
	status           gameplay.RoomStatus
	hostSessionID    string
	countdownSec     int
	roundDurationSec int
	countdownStarted *time.Time
	startedAt        *time.Time
	endsAt           *time.Time
	createdAt        time.Time
	updatedAt        time.Time
	players          map[string]*gameplay.RoomPlayer
	sequence         int64
}

func NewHexBlitzRoomService(logger *zap.Logger, opts ...HexBlitzRoomServiceOption) *HexBlitzRoomService {
	svc := &HexBlitzRoomService{
		rooms:         make(map[uuid.UUID]*hexBlitzRoomState),
		sessionToRoom: make(map[string]uuid.UUID),
		logger:        logger,
		countdown:     3 * time.Second,
		roundDuration: 75 * time.Second,
	}
	if svc.logger == nil {
		svc.logger = zap.NewNop()
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithHexBlitzRoomTiming(countdown, round time.Duration) HexBlitzRoomServiceOption {
	return func(s *HexBlitzRoomService) {
		if countdown > 0 {
			s.countdown = countdown
		}
		if round > 0 {
			s.roundDuration = round
		}
	}
}

func WithHexBlitzRoomNotifier(notifier func(uuid.UUID)) HexBlitzRoomServiceOption {
	return func(s *HexBlitzRoomService) {
		s.notifier = notifier
	}
}

func (s *HexBlitzRoomService) SetNotifier(notifier func(uuid.UUID)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifier = notifier
}

func (s *HexBlitzRoomService) ListRooms() []*gameplay.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*gameplay.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		items = append(items, snapshotHexBlitzRoom(room))
	}

	slices.SortFunc(items, func(a, b *gameplay.Room) int {
		switch {
		case a.UpdatedAt.After(b.UpdatedAt):
			return -1
		case a.UpdatedAt.Before(b.UpdatedAt):
			return 1
		default:
			return strings.Compare(a.Title, b.Title)
		}
	})

	return items
}

func (s *HexBlitzRoomService) GetRoom(roomID uuid.UUID) (*gameplay.Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return nil, false
	}
	return snapshotHexBlitzRoom(room), true
}

func (s *HexBlitzRoomService) CreateRoom(input CreateHexBlitzRoomInput) (*gameplay.Room, error) {
	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		return nil, apperr.BadRequest("缺少 session_id")
	}

	now := time.Now()
	roomID := uuid.New()
	room := &hexBlitzRoomState{
		id:               roomID,
		code:             strings.ToUpper(roomID.String()[:6]),
		title:            sanitizeHexBlitzTitle(input.Title),
		status:           gameplay.RoomStatusWaiting,
		hostSessionID:    sessionID,
		countdownSec:     int(s.countdown / time.Second),
		roundDurationSec: int(s.roundDuration / time.Second),
		createdAt:        now,
		updatedAt:        now,
		players:          make(map[string]*gameplay.RoomPlayer),
	}
	room.players[sessionID] = &gameplay.RoomPlayer{
		SessionID: sessionID,
		UserID:    cloneUUIDPtr(input.UserID),
		Name:      sanitizeHexBlitzPlayerName(input.PlayerName),
		Ready:     false,
		Connected: false,
		IsHost:    true,
		Score:     0,
		JoinedAt:  now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.rooms[roomID] = room
	s.sessionToRoom[sessionID] = roomID
	s.mu.Unlock()

	s.notify(roomID)
	return snapshotHexBlitzRoom(room), nil
}

func (s *HexBlitzRoomService) JoinRoom(input JoinHexBlitzRoomInput) (*gameplay.Room, error) {
	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		return nil, apperr.BadRequest("缺少 session_id")
	}

	var notifyRoomIDs []uuid.UUID

	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		for _, roomID := range notifyRoomIDs {
			s.notify(roomID)
		}
	}()

	if previousRoomID, ok := s.sessionToRoom[sessionID]; ok && previousRoomID != input.RoomID {
		if previousRoom, exists := s.rooms[previousRoomID]; exists {
			removeHexBlitzPlayer(previousRoom, sessionID)
			if len(previousRoom.players) == 0 {
				delete(s.rooms, previousRoomID)
			} else {
				reassignHexBlitzHost(previousRoom)
				previousRoom.updatedAt = time.Now()
			}
			notifyRoomIDs = append(notifyRoomIDs, previousRoomID)
		}
		delete(s.sessionToRoom, sessionID)
	}

	room, ok := s.rooms[input.RoomID]
	if !ok {
		return nil, apperr.ErrNotFound
	}

	now := time.Now()
	if player, exists := room.players[sessionID]; exists {
		player.Name = sanitizeHexBlitzPlayerName(input.PlayerName)
		player.UserID = cloneUUIDPtr(input.UserID)
		player.Connected = true
		player.UpdatedAt = now
		room.updatedAt = now
		s.sessionToRoom[sessionID] = room.id
		notifyRoomIDs = append(notifyRoomIDs, room.id)
		return snapshotHexBlitzRoom(room), nil
	}

	if room.status != gameplay.RoomStatusWaiting {
		return nil, apperr.BadRequest("当前房间已开局，暂不支持中途加入")
	}
	if len(room.players) >= hexBlitzMaxPlayers {
		return nil, apperr.BadRequest("房间人数已满")
	}

	room.players[sessionID] = &gameplay.RoomPlayer{
		SessionID: sessionID,
		UserID:    cloneUUIDPtr(input.UserID),
		Name:      sanitizeHexBlitzPlayerName(input.PlayerName),
		Ready:     false,
		Connected: true,
		IsHost:    room.hostSessionID == sessionID,
		Score:     0,
		JoinedAt:  now,
		UpdatedAt: now,
	}
	room.updatedAt = now
	s.sessionToRoom[sessionID] = room.id
	notifyRoomIDs = append(notifyRoomIDs, room.id)
	return snapshotHexBlitzRoom(room), nil
}

func (s *HexBlitzRoomService) LeaveRoom(roomID uuid.UUID, sessionID string) (*gameplay.Room, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, apperr.BadRequest("缺少 session_id")
	}

	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if _, exists := room.players[sessionID]; !exists {
		s.mu.Unlock()
		return nil, apperr.New(apperr.CodeForbidden, "您不在该房间中")
	}

	removeHexBlitzPlayer(room, sessionID)
	delete(s.sessionToRoom, sessionID)
	var snapshot *gameplay.Room
	if len(room.players) == 0 {
		delete(s.rooms, roomID)
		s.mu.Unlock()
		s.notify(roomID)
		return nil, nil
	}
	reassignHexBlitzHost(room)
	room.updatedAt = time.Now()
	snapshot = snapshotHexBlitzRoom(room)
	s.mu.Unlock()

	s.notify(roomID)
	return snapshot, nil
}

func (s *HexBlitzRoomService) DisconnectSession(sessionID string) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}

	var notifyRoomID uuid.UUID
	shouldNotify := false

	s.mu.Lock()
	roomID, ok := s.sessionToRoom[sessionID]
	if !ok {
		s.mu.Unlock()
		return
	}
	room, exists := s.rooms[roomID]
	if !exists {
		delete(s.sessionToRoom, sessionID)
		s.mu.Unlock()
		return
	}
	player, exists := room.players[sessionID]
	if !exists {
		delete(s.sessionToRoom, sessionID)
		s.mu.Unlock()
		return
	}

	player.Connected = false
	player.Ready = false
	player.UpdatedAt = time.Now()
	room.updatedAt = player.UpdatedAt
	reassignHexBlitzHost(room)
	notifyRoomID = roomID
	shouldNotify = true

	if room.status == gameplay.RoomStatusWaiting && !hasHexBlitzConnectedPlayers(room) {
		for playerSessionID := range room.players {
			delete(s.sessionToRoom, playerSessionID)
		}
		delete(s.rooms, roomID)
	}
	s.mu.Unlock()

	if shouldNotify {
		s.notify(notifyRoomID)
	}
}

func (s *HexBlitzRoomService) SetReady(input SetHexBlitzReadyInput) (*gameplay.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.status != gameplay.RoomStatusWaiting {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前房间不在准备阶段")
	}
	player, exists := room.players[input.SessionID]
	if !exists {
		s.mu.Unlock()
		return nil, apperr.New(apperr.CodeForbidden, "您不在该房间中")
	}

	if input.Ready != nil {
		player.Ready = *input.Ready
	} else {
		player.Ready = !player.Ready
	}
	player.UpdatedAt = time.Now()
	room.updatedAt = player.UpdatedAt
	snapshot := snapshotHexBlitzRoom(room)
	s.mu.Unlock()

	s.notify(input.RoomID)
	return snapshot, nil
}

func (s *HexBlitzRoomService) StartMatch(roomID uuid.UUID, sessionID string) (*gameplay.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.hostSessionID != sessionID {
		s.mu.Unlock()
		return nil, apperr.New(apperr.CodeForbidden, "只有房主可以开始对局")
	}
	if room.status != gameplay.RoomStatusWaiting {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前房间不可开始新对局")
	}
	if !hasHexBlitzConnectedPlayers(room) {
		s.mu.Unlock()
		return nil, apperr.BadRequest("房间内没有在线玩家")
	}
	if !allHexBlitzConnectedPlayersReady(room) {
		s.mu.Unlock()
		return nil, apperr.BadRequest("还有在线玩家未准备")
	}

	now := time.Now()
	room.status = gameplay.RoomStatusCountdown
	room.countdownStarted = &now
	room.startedAt = nil
	room.endsAt = nil
	room.updatedAt = now
	room.sequence++
	sequence := room.sequence
	for _, player := range room.players {
		player.Score = 0
		player.UpdatedAt = now
	}
	snapshot := snapshotHexBlitzRoom(room)
	s.mu.Unlock()

	s.notify(roomID)
	go s.runHexBlitzMatch(roomID, sequence)
	return snapshot, nil
}

func (s *HexBlitzRoomService) UpdateScore(input UpdateHexBlitzScoreInput) (*gameplay.Room, error) {
	if input.Score < 0 {
		input.Score = 0
	}
	if input.Score > 999999 {
		input.Score = 999999
	}

	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.status != gameplay.RoomStatusRunning {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前房间不在对局中")
	}
	player, exists := room.players[input.SessionID]
	if !exists {
		s.mu.Unlock()
		return nil, apperr.New(apperr.CodeForbidden, "您不在该房间中")
	}
	player.Score = input.Score
	player.UpdatedAt = time.Now()
	room.updatedAt = player.UpdatedAt
	snapshot := snapshotHexBlitzRoom(room)
	s.mu.Unlock()

	s.notify(input.RoomID)
	return snapshot, nil
}

func (s *HexBlitzRoomService) runHexBlitzMatch(roomID uuid.UUID, sequence int64) {
	time.Sleep(s.countdown)

	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok || room.sequence != sequence || room.status != gameplay.RoomStatusCountdown {
		s.mu.Unlock()
		return
	}
	startedAt := time.Now()
	endsAt := startedAt.Add(s.roundDuration)
	room.status = gameplay.RoomStatusRunning
	room.startedAt = &startedAt
	room.endsAt = &endsAt
	room.updatedAt = startedAt
	for _, player := range room.players {
		player.Ready = false
		player.Score = 0
		player.UpdatedAt = startedAt
	}
	s.mu.Unlock()
	s.notify(roomID)

	time.Sleep(s.roundDuration)

	s.mu.Lock()
	room, ok = s.rooms[roomID]
	if !ok || room.sequence != sequence || room.status != gameplay.RoomStatusRunning {
		s.mu.Unlock()
		return
	}
	finishedAt := time.Now()
	room.status = gameplay.RoomStatusFinished
	room.updatedAt = finishedAt
	for _, player := range room.players {
		player.Ready = false
		player.UpdatedAt = finishedAt
	}
	s.mu.Unlock()
	s.notify(roomID)
}

func (s *HexBlitzRoomService) notify(roomID uuid.UUID) {
	s.mu.RLock()
	notifier := s.notifier
	s.mu.RUnlock()
	if notifier != nil {
		notifier(roomID)
	}
}

func snapshotHexBlitzRoom(room *hexBlitzRoomState) *gameplay.Room {
	players := make([]gameplay.RoomPlayer, 0, len(room.players))
	readyCount := 0
	for _, player := range room.players {
		if player.Ready {
			readyCount++
		}
		players = append(players, gameplay.RoomPlayer{
			SessionID: player.SessionID,
			UserID:    cloneUUIDPtr(player.UserID),
			Name:      player.Name,
			Ready:     player.Ready,
			Connected: player.Connected,
			IsHost:    player.IsHost,
			Score:     player.Score,
			JoinedAt:  player.JoinedAt,
			UpdatedAt: player.UpdatedAt,
		})
	}

	slices.SortFunc(players, func(a, b gameplay.RoomPlayer) int {
		switch {
		case a.IsHost && !b.IsHost:
			return -1
		case !a.IsHost && b.IsHost:
			return 1
		case a.Connected && !b.Connected:
			return -1
		case !a.Connected && b.Connected:
			return 1
		case a.Score > b.Score:
			return -1
		case a.Score < b.Score:
			return 1
		case a.JoinedAt.Before(b.JoinedAt):
			return -1
		case a.JoinedAt.After(b.JoinedAt):
			return 1
		default:
			return strings.Compare(a.Name, b.Name)
		}
	})

	return &gameplay.Room{
		ID:               room.id,
		Code:             room.code,
		GameSlug:         hexBlitzGameSlug,
		Title:            room.title,
		Status:           room.status,
		HostSessionID:    room.hostSessionID,
		CountdownSec:     room.countdownSec,
		RoundDurationSec: room.roundDurationSec,
		PlayerCount:      len(players),
		ReadyCount:       readyCount,
		CountdownStarted: cloneTimePtr(room.countdownStarted),
		StartedAt:        cloneTimePtr(room.startedAt),
		EndsAt:           cloneTimePtr(room.endsAt),
		CreatedAt:        room.createdAt,
		UpdatedAt:        room.updatedAt,
		Players:          players,
	}
}

func sanitizeHexBlitzTitle(value string) string {
	title := strings.TrimSpace(value)
	if title == "" {
		return hexBlitzDefaultTitle
	}
	if len([]rune(title)) > 32 {
		return string([]rune(title)[:32])
	}
	return title
}

func sanitizeHexBlitzPlayerName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return hexBlitzDefaultPlayer
	}
	if len([]rune(name)) > 24 {
		return string([]rune(name)[:24])
	}
	return name
}

func removeHexBlitzPlayer(room *hexBlitzRoomState, sessionID string) {
	delete(room.players, sessionID)
}

func reassignHexBlitzHost(room *hexBlitzRoomState) {
	if len(room.players) == 0 {
		room.hostSessionID = ""
		return
	}

	var nextHost *gameplay.RoomPlayer
	for _, player := range room.players {
		if player.Connected {
			if nextHost == nil || player.JoinedAt.Before(nextHost.JoinedAt) {
				nextHost = player
			}
		}
	}
	if nextHost == nil {
		for _, player := range room.players {
			if nextHost == nil || player.JoinedAt.Before(nextHost.JoinedAt) {
				nextHost = player
			}
		}
	}
	if nextHost == nil {
		room.hostSessionID = ""
		return
	}
	room.hostSessionID = nextHost.SessionID
	for _, player := range room.players {
		player.IsHost = player.SessionID == room.hostSessionID
	}
}

func hasHexBlitzConnectedPlayers(room *hexBlitzRoomState) bool {
	for _, player := range room.players {
		if player.Connected {
			return true
		}
	}
	return false
}

func allHexBlitzConnectedPlayersReady(room *hexBlitzRoomState) bool {
	connectedCount := 0
	for _, player := range room.players {
		if !player.Connected {
			continue
		}
		connectedCount++
		if !player.Ready {
			return false
		}
	}
	return connectedCount > 0
}

func cloneUUIDPtr(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	dup := *value
	return &dup
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	dup := *value
	return &dup
}
