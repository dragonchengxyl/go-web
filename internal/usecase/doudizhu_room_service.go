package usecase

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/doudizhu"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

const (
	doudizhuMaxPlayers          = 3
	doudizhuDefaultRoomTitle    = "斗地主房间"
	doudizhuDefaultDemoTitle    = "斗地主人机演示房"
	doudizhuDefaultPlayerName   = "牌桌玩家"
	doudizhuDefaultBotLevel     = "basic"
	doudizhuRecentActionsLimit  = 12
	doudizhuDefaultTurnDuration = 25 * time.Second
)

type DoudizhuRoomService struct {
	mu            sync.RWMutex
	rooms         map[uuid.UUID]*doudizhuRoomState
	sessionToRoom map[string]uuid.UUID
	logger        *zap.Logger
	repo          doudizhu.Repository
	turnDuration  time.Duration
	notifier      func(uuid.UUID)
	seedSource    func() int64
}

type DoudizhuRoomServiceOption func(*DoudizhuRoomService)

type CreateDoudizhuRoomInput struct {
	SessionID  string
	PlayerName string
	Title      string
	UserID     *uuid.UUID
}

type JoinDoudizhuRoomInput struct {
	RoomID     uuid.UUID
	SessionID  string
	PlayerName string
	UserID     *uuid.UUID
}

type SetDoudizhuReadyInput struct {
	RoomID    uuid.UUID
	SessionID string
	Ready     *bool
}

type DoudizhuBidInput struct {
	RoomID    uuid.UUID
	SessionID string
	Score     int
}

type DoudizhuPlayInput struct {
	RoomID    uuid.UUID
	SessionID string
	Cards     []doudizhu.Card
}

type ToggleDoudizhuAutoPlayInput struct {
	RoomID    uuid.UUID
	SessionID string
	Enabled   *bool
}

type doudizhuPersistPayload struct {
	match   *doudizhu.Match
	results []doudizhu.MatchPlayerResult
	events  []doudizhu.ActionEvent
}

type doudizhuRoomState struct {
	id             uuid.UUID
	code           string
	title          string
	matchMode      doudizhu.MatchMode
	status         doudizhu.RoundPhase
	hostSessionID  string
	players        map[doudizhu.Seat]*doudizhuPlayerState
	sessionSeat    map[string]doudizhu.Seat
	round          *DoudizhuRoundState
	lastPlay       *doudizhu.Combo
	lastPlayCards  []doudizhu.Card
	lastPlaySeat   *doudizhu.Seat
	passCount      int
	winningSide    *doudizhu.PlayerRole
	actions        []doudizhu.ActionRecord
	matchPersisted bool
	createdAt      time.Time
	updatedAt      time.Time
	turnExpiresAt  *time.Time
}

type doudizhuPlayerState struct {
	sessionID string
	userID    *uuid.UUID
	seat      doudizhu.Seat
	name      string
	ready     bool
	connected bool
	isHost    bool
	isBot     bool
	botLevel  string
	autoPlay  bool
	joinedAt  time.Time
	updatedAt time.Time
}

func NewDoudizhuRoomService(logger *zap.Logger, opts ...DoudizhuRoomServiceOption) *DoudizhuRoomService {
	svc := &DoudizhuRoomService{
		rooms:         make(map[uuid.UUID]*doudizhuRoomState),
		sessionToRoom: make(map[string]uuid.UUID),
		logger:        logger,
		turnDuration:  doudizhuDefaultTurnDuration,
		seedSource: func() int64 {
			return time.Now().UnixNano()
		},
	}
	if svc.logger == nil {
		svc.logger = zap.NewNop()
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithDoudizhuNotifier(notifier func(uuid.UUID)) DoudizhuRoomServiceOption {
	return func(s *DoudizhuRoomService) {
		s.notifier = notifier
	}
}

func WithDoudizhuTurnDuration(duration time.Duration) DoudizhuRoomServiceOption {
	return func(s *DoudizhuRoomService) {
		if duration > 0 {
			s.turnDuration = duration
		}
	}
}

func WithDoudizhuSeedSource(seedSource func() int64) DoudizhuRoomServiceOption {
	return func(s *DoudizhuRoomService) {
		if seedSource != nil {
			s.seedSource = seedSource
		}
	}
}

func WithDoudizhuRepository(repo doudizhu.Repository) DoudizhuRoomServiceOption {
	return func(s *DoudizhuRoomService) {
		s.repo = repo
	}
}

func (s *DoudizhuRoomService) SetNotifier(notifier func(uuid.UUID)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifier = notifier
}

func (s *DoudizhuRoomService) ListLeaderboard(ctx context.Context, limit int) ([]*doudizhu.LeaderboardEntry, error) {
	if s.repo == nil {
		return []*doudizhu.LeaderboardEntry{}, nil
	}
	return s.repo.ListLeaderboard(ctx, limit)
}

func (s *DoudizhuRoomService) ListRecentMatches(ctx context.Context, limit int) ([]*doudizhu.MatchSummary, error) {
	if s.repo == nil {
		return []*doudizhu.MatchSummary{}, nil
	}
	return s.repo.ListRecentMatches(ctx, limit)
}

func (s *DoudizhuRoomService) ListUserRecentMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*doudizhu.MatchSummary, error) {
	if s.repo == nil {
		return []*doudizhu.MatchSummary{}, nil
	}
	return s.repo.ListUserRecentMatches(ctx, userID, limit)
}

func (s *DoudizhuRoomService) GetReplay(ctx context.Context, matchID uuid.UUID) (*doudizhu.MatchReplay, error) {
	if s.repo == nil {
		return nil, apperr.ErrNotFound
	}
	replay, err := s.repo.GetReplay(ctx, matchID)
	if err != nil {
		return nil, err
	}
	return replay, nil
}

func (s *DoudizhuRoomService) ListRooms() []*doudizhu.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*doudizhu.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		items = append(items, snapshotDoudizhuRoom(room))
	}

	slices.SortFunc(items, func(a, b *doudizhu.Room) int {
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

func (s *DoudizhuRoomService) GetRoom(roomID uuid.UUID) (*doudizhu.Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[roomID]
	if !ok {
		return nil, false
	}
	return snapshotDoudizhuRoom(room), true
}

func (s *DoudizhuRoomService) GetPrivateState(roomID uuid.UUID, sessionID string) (*doudizhu.PrivateState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[roomID]
	if !ok {
		return nil, false
	}
	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		return nil, false
	}
	return snapshotDoudizhuPrivateState(room, seat), true
}

func (s *DoudizhuRoomService) CreateRoom(input CreateDoudizhuRoomInput) (*doudizhu.Room, error) {
	return s.createRoom(input, doudizhu.MatchModePVP)
}

func (s *DoudizhuRoomService) CreateDemoRoom(input CreateDoudizhuRoomInput) (*doudizhu.Room, error) {
	return s.createRoom(input, doudizhu.MatchModeDemoAI)
}

func (s *DoudizhuRoomService) createRoom(input CreateDoudizhuRoomInput, mode doudizhu.MatchMode) (*doudizhu.Room, error) {
	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		return nil, apperr.BadRequest("缺少 session_id")
	}

	now := time.Now()
	roomID := uuid.New()
	room := &doudizhuRoomState{
		id:            roomID,
		code:          strings.ToUpper(roomID.String()[:6]),
		title:         sanitizeDoudizhuTitle(mode, input.Title),
		matchMode:     mode,
		status:        doudizhu.RoundPhaseWaiting,
		hostSessionID: sessionID,
		players:       make(map[doudizhu.Seat]*doudizhuPlayerState, doudizhuMaxPlayers),
		sessionSeat:   make(map[string]doudizhu.Seat, doudizhuMaxPlayers),
		createdAt:     now,
		updatedAt:     now,
	}
	room.players[doudizhu.Seat0] = &doudizhuPlayerState{
		sessionID: sessionID,
		userID:    cloneUUIDPtr(input.UserID),
		seat:      doudizhu.Seat0,
		name:      sanitizeDoudizhuPlayerName(input.PlayerName),
		isHost:    true,
		joinedAt:  now,
		updatedAt: now,
	}
	room.sessionSeat[sessionID] = doudizhu.Seat0

	if mode == doudizhu.MatchModeDemoAI {
		for _, seat := range []doudizhu.Seat{doudizhu.Seat1, doudizhu.Seat2} {
			botSessionID := uuid.NewString()
			room.players[seat] = &doudizhuPlayerState{
				sessionID: botSessionID,
				seat:      seat,
				name:      demoBotName(seat),
				ready:     true,
				connected: true,
				isBot:     true,
				botLevel:  doudizhuDefaultBotLevel,
				autoPlay:  true,
				joinedAt:  now,
				updatedAt: now,
			}
			room.sessionSeat[botSessionID] = seat
		}
	}

	s.mu.Lock()
	s.rooms[roomID] = room
	s.sessionToRoom[sessionID] = roomID
	if mode == doudizhu.MatchModeDemoAI {
		for _, seat := range []doudizhu.Seat{doudizhu.Seat1, doudizhu.Seat2} {
			s.sessionToRoom[room.players[seat].sessionID] = roomID
		}
	}
	s.mu.Unlock()

	s.notify(roomID)
	return snapshotDoudizhuRoom(room), nil
}

func (s *DoudizhuRoomService) JoinRoom(input JoinDoudizhuRoomInput) (*doudizhu.Room, error) {
	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		return nil, apperr.BadRequest("缺少 session_id")
	}

	now := time.Now()
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}

	if seat, exists := room.sessionSeat[sessionID]; exists {
		player := room.players[seat]
		player.connected = true
		player.updatedAt = now
		if !player.isBot {
			player.name = sanitizeDoudizhuPlayerName(input.PlayerName)
			player.userID = cloneUUIDPtr(input.UserID)
		}
		room.updatedAt = now
		s.mu.Unlock()
		s.notify(room.id)
		return snapshotDoudizhuRoom(room), nil
	}

	if room.matchMode == doudizhu.MatchModeDemoAI {
		s.mu.Unlock()
		return nil, apperr.BadRequest("人机演示房不支持新的真人玩家加入")
	}
	if len(room.players) >= doudizhuMaxPlayers {
		s.mu.Unlock()
		return nil, apperr.BadRequest("房间已满")
	}
	if room.status != doudizhu.RoundPhaseWaiting && room.status != doudizhu.RoundPhaseSettlement {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前房间不允许加入")
	}

	seat, found := firstAvailableDoudizhuSeat(room)
	if !found {
		s.mu.Unlock()
		return nil, apperr.BadRequest("没有可用座位")
	}

	room.players[seat] = &doudizhuPlayerState{
		sessionID: sessionID,
		userID:    cloneUUIDPtr(input.UserID),
		seat:      seat,
		name:      sanitizeDoudizhuPlayerName(input.PlayerName),
		connected: true,
		joinedAt:  now,
		updatedAt: now,
	}
	room.sessionSeat[sessionID] = seat
	room.updatedAt = now
	s.sessionToRoom[sessionID] = room.id
	s.mu.Unlock()

	s.notify(room.id)
	return snapshotDoudizhuRoom(room), nil
}

func (s *DoudizhuRoomService) LeaveRoom(roomID uuid.UUID, sessionID string) (*doudizhu.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.status != doudizhu.RoundPhaseWaiting && room.status != doudizhu.RoundPhaseSettlement {
		s.mu.Unlock()
		return nil, apperr.BadRequest("对局进行中不能退出房间")
	}

	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		s.mu.Unlock()
		return nil, apperr.BadRequest("你不在该房间内")
	}
	player := room.players[seat]
	if player.isBot {
		s.mu.Unlock()
		return nil, apperr.BadRequest("机器人无法主动退出房间")
	}

	delete(room.players, seat)
	delete(room.sessionSeat, sessionID)
	delete(s.sessionToRoom, sessionID)
	room.updatedAt = time.Now()

	if len(room.players) == 0 || !roomHasHumanPlayer(room) {
		delete(s.rooms, room.id)
		s.mu.Unlock()
		s.notify(roomID)
		return nil, nil
	}

	if room.hostSessionID == sessionID {
		room.hostSessionID = pickNextHostSessionID(room)
		if room.hostSessionID != "" {
			if hostSeat, ok := room.sessionSeat[room.hostSessionID]; ok {
				for _, item := range room.players {
					item.isHost = false
				}
				room.players[hostSeat].isHost = true
			}
		}
	}

	snapshot := snapshotDoudizhuRoom(room)
	s.mu.Unlock()
	s.notify(roomID)
	return snapshot, nil
}

func (s *DoudizhuRoomService) DisconnectSession(sessionID string) {
	s.mu.Lock()
	roomID, ok := s.sessionToRoom[sessionID]
	if !ok {
		s.mu.Unlock()
		return
	}
	room, exists := s.rooms[roomID]
	if !exists {
		s.mu.Unlock()
		return
	}
	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		s.mu.Unlock()
		return
	}
	player := room.players[seat]
	if player == nil || player.isBot {
		s.mu.Unlock()
		return
	}

	player.connected = false
	player.autoPlay = true
	player.updatedAt = time.Now()
	room.updatedAt = player.updatedAt

	s.drainAutoActorsLocked(room)
	persist := s.maybeBuildPersistPayloadLocked(room)
	s.mu.Unlock()
	s.persistMatch(persist)
	s.notify(roomID)
}

func (s *DoudizhuRoomService) SetReady(input SetDoudizhuReadyInput) (*doudizhu.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.status != doudizhu.RoundPhaseWaiting && room.status != doudizhu.RoundPhaseSettlement {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前阶段不能修改准备状态")
	}

	seat, exists := room.sessionSeat[input.SessionID]
	if !exists {
		s.mu.Unlock()
		return nil, apperr.BadRequest("你不在该房间内")
	}
	player := room.players[seat]
	if player.isBot {
		s.mu.Unlock()
		return nil, apperr.BadRequest("机器人始终处于准备状态")
	}
	ready := true
	if input.Ready != nil {
		ready = *input.Ready
	}
	player.ready = ready
	player.updatedAt = time.Now()
	room.updatedAt = player.updatedAt
	snapshot := snapshotDoudizhuRoom(room)
	s.mu.Unlock()
	s.notify(room.id)
	return snapshot, nil
}

func (s *DoudizhuRoomService) StartRound(roomID uuid.UUID, sessionID string) (*doudizhu.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	if room.hostSessionID != sessionID {
		s.mu.Unlock()
		return nil, apperr.ErrForbidden
	}
	if room.status != doudizhu.RoundPhaseWaiting && room.status != doudizhu.RoundPhaseSettlement && room.status != doudizhu.RoundPhaseRedeal {
		s.mu.Unlock()
		return nil, apperr.BadRequest("当前房间无法开始新一局")
	}
	if len(room.players) != doudizhuMaxPlayers {
		s.mu.Unlock()
		return nil, apperr.BadRequest("房间人数不足 3 人")
	}
	if !allDoudizhuPlayersReady(room) {
		s.mu.Unlock()
		return nil, apperr.BadRequest("还有玩家未准备")
	}

	startingSeat := room.sessionSeat[room.hostSessionID]
	round, err := NewDoudizhuRound(s.seedSource(), room.matchMode, startingSeat)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	room.round = round
	room.status = round.Phase
	room.lastPlay = nil
	room.lastPlayCards = nil
	room.lastPlaySeat = nil
	room.passCount = 0
	room.actions = nil
	room.winningSide = nil
	room.matchPersisted = false
	room.updatedAt = time.Now()
	for seat, hand := range round.Hands {
		player := room.players[doudizhu.Seat(seat)]
		if player != nil {
			player.ready = player.isBot
			player.autoPlay = player.isBot || !player.connected
			player.updatedAt = room.updatedAt
			_ = hand
		}
	}
	s.setTurnExpiryLocked(room)
	s.drainAutoActorsLocked(room)
	snapshot := snapshotDoudizhuRoom(room)
	s.mu.Unlock()
	s.notify(room.id)
	return snapshot, nil
}

func (s *DoudizhuRoomService) Bid(input DoudizhuBidInput) (*doudizhu.ActionResult, error) {
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	result, err := s.applyBidLocked(room, input.SessionID, input.Score)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.drainAutoActorsLocked(room)
	persist := s.maybeBuildPersistPayloadLocked(room)
	s.mu.Unlock()
	s.persistMatch(persist)
	s.notify(room.id)
	return result, nil
}

func (s *DoudizhuRoomService) PassBid(roomID uuid.UUID, sessionID string) (*doudizhu.ActionResult, error) {
	return s.Bid(DoudizhuBidInput{
		RoomID:    roomID,
		SessionID: sessionID,
		Score:     0,
	})
}

func (s *DoudizhuRoomService) PlayCards(input DoudizhuPlayInput) (*doudizhu.ActionResult, error) {
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	result, err := s.playCardsLocked(room, input.SessionID, input.Cards, false)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.drainAutoActorsLocked(room)
	persist := s.maybeBuildPersistPayloadLocked(room)
	s.mu.Unlock()
	s.persistMatch(persist)
	s.notify(room.id)
	return result, nil
}

func (s *DoudizhuRoomService) PassTurn(roomID uuid.UUID, sessionID string) (*doudizhu.ActionResult, error) {
	s.mu.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	result, err := s.passTurnLocked(room, sessionID, false)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.drainAutoActorsLocked(room)
	persist := s.maybeBuildPersistPayloadLocked(room)
	s.mu.Unlock()
	s.persistMatch(persist)
	s.notify(room.id)
	return result, nil
}

func (s *DoudizhuRoomService) ToggleAutoPlay(input ToggleDoudizhuAutoPlayInput) (*doudizhu.Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[input.RoomID]
	if !ok {
		s.mu.Unlock()
		return nil, apperr.ErrNotFound
	}
	seat, exists := room.sessionSeat[input.SessionID]
	if !exists {
		s.mu.Unlock()
		return nil, apperr.BadRequest("你不在该房间内")
	}
	player := room.players[seat]
	if player.isBot {
		s.mu.Unlock()
		return nil, apperr.BadRequest("机器人托管状态不可修改")
	}
	enabled := !player.autoPlay
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	player.autoPlay = enabled
	player.updatedAt = time.Now()
	room.updatedAt = player.updatedAt
	s.drainAutoActorsLocked(room)
	persist := s.maybeBuildPersistPayloadLocked(room)
	snapshot := snapshotDoudizhuRoom(room)
	s.mu.Unlock()
	s.persistMatch(persist)
	s.notify(room.id)
	return snapshot, nil
}

func (s *DoudizhuRoomService) applyBidLocked(room *doudizhuRoomState, sessionID string, score int) (*doudizhu.ActionResult, error) {
	if room.round == nil || room.round.Phase != doudizhu.RoundPhaseBidding {
		return nil, apperr.BadRequest("当前不在叫分阶段")
	}
	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		return nil, apperr.BadRequest("你不在该房间内")
	}
	if room.round.CurrentBidder != seat {
		return nil, apperr.ErrForbidden
	}

	if err := room.round.ApplyBid(seat, score); err != nil {
		return nil, mapDoudizhuRoundError(err)
	}

	now := time.Now()
	room.status = room.round.Phase
	room.updatedAt = now
	record := doudizhu.BidRecord{Seat: seat, Score: score, At: now}
	room.actions = appendActionRecord(room.actions, doudizhu.ActionRecord{
		Seat:       seat,
		ActionType: "bid",
		At:         now,
		Message:    bidActionMessage(room.players[seat].name, score),
		ActorName:  room.players[seat].name,
	})

	if room.status == doudizhu.RoundPhasePlaying && room.round.Landlord != nil {
		landlord := *room.round.Landlord
		room.actions = appendActionRecord(room.actions, doudizhu.ActionRecord{
			Seat:       landlord,
			ActionType: "landlord_assigned",
			At:         now,
			Message:    room.players[landlord].name + " 成为地主",
			ActorName:  room.players[landlord].name,
		})
	}

	room.turnExpiresAt = nil
	if room.status == doudizhu.RoundPhaseBidding || room.status == doudizhu.RoundPhasePlaying {
		s.setTurnExpiryLocked(room)
	}
	_ = record

	return &doudizhu.ActionResult{
		ActionType: "bid",
		Seat:       seat,
		ActorName:  room.players[seat].name,
		HandCount:  len(room.round.Hands[seat]),
		NextTurn:   currentDoudizhuTurn(room),
		Status:     room.status,
		Message:    bidActionMessage(room.players[seat].name, score),
	}, nil
}

func (s *DoudizhuRoomService) playCardsLocked(room *doudizhuRoomState, sessionID string, cards []doudizhu.Card, auto bool) (*doudizhu.ActionResult, error) {
	if room.round == nil || room.round.Phase != doudizhu.RoundPhasePlaying {
		return nil, apperr.BadRequest("当前不在出牌阶段")
	}
	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		return nil, apperr.BadRequest("你不在该房间内")
	}
	if room.round.CurrentTurn == nil || *room.round.CurrentTurn != seat {
		return nil, apperr.ErrForbidden
	}

	combo, err := DoudizhuEvaluateCombo(cards)
	if err != nil {
		return nil, apperr.BadRequest("非法牌型")
	}
	if !doudizhuPlayerHasCards(room.round.Hands[seat], cards) {
		return nil, apperr.BadRequest("手牌与出牌内容不匹配")
	}
	if room.lastPlay != nil && room.lastPlaySeat != nil && *room.lastPlaySeat != seat {
		ok, compareErr := DoudizhuCompareCombos(combo, room.lastPlay)
		if compareErr != nil {
			return nil, apperr.BadRequest("当前出牌无法压过上一手")
		}
		if !ok {
			return nil, apperr.BadRequest("当前出牌无法压过上一手")
		}
	}

	room.round.Hands[seat] = doudizhuRemoveCards(room.round.Hands[seat], cards)
	room.lastPlay = combo
	room.lastPlayCards = doudizhuSortedCards(cards)
	room.lastPlaySeat = seatPtr(seat)
	room.passCount = 0
	now := time.Now()
	room.updatedAt = now

	result := &doudizhu.ActionResult{
		ActionType: "play_cards",
		Seat:       seat,
		ActorName:  room.players[seat].name,
		Combo:      combo,
		Cards:      doudizhuSortedCards(cards),
		HandCount:  len(room.round.Hands[seat]),
		Status:     room.round.Phase,
		Message:    room.players[seat].name + " 出牌",
	}

	if len(room.round.Hands[seat]) == 0 {
		winningSide := room.round.Roles[seat]
		room.winningSide = rolePtr(winningSide)
		room.round.Phase = doudizhu.RoundPhaseSettlement
		room.status = room.round.Phase
		room.turnExpiresAt = nil
		room.actions = appendActionRecord(room.actions, doudizhu.ActionRecord{
			Seat:        seat,
			ActionType:  "settlement",
			Cards:       doudizhuSortedCards(cards),
			Combo:       combo,
			At:          now,
			Message:     room.players[seat].name + " 已出完手牌，本局结束",
			ActorName:   room.players[seat].name,
			WinningSide: room.winningSide,
		})
		result.Status = room.status
		result.WinningSide = room.winningSide
		return result, nil
	}

	next := seat.Next()
	room.round.CurrentTurn = &next
	room.status = room.round.Phase
	s.setTurnExpiryLocked(room)
	room.actions = appendActionRecord(room.actions, doudizhu.ActionRecord{
		Seat:       seat,
		ActionType: actionType(auto, "play_cards"),
		Cards:      doudizhuSortedCards(cards),
		Combo:      combo,
		At:         now,
		Message:    room.players[seat].name + " 出牌",
		ActorName:  room.players[seat].name,
	})
	result.NextTurn = room.round.CurrentTurn
	return result, nil
}

func (s *DoudizhuRoomService) passTurnLocked(room *doudizhuRoomState, sessionID string, auto bool) (*doudizhu.ActionResult, error) {
	if room.round == nil || room.round.Phase != doudizhu.RoundPhasePlaying {
		return nil, apperr.BadRequest("当前不在出牌阶段")
	}
	seat, exists := room.sessionSeat[sessionID]
	if !exists {
		return nil, apperr.BadRequest("你不在该房间内")
	}
	if room.round.CurrentTurn == nil || *room.round.CurrentTurn != seat {
		return nil, apperr.ErrForbidden
	}
	if room.lastPlay == nil || room.lastPlaySeat == nil || *room.lastPlaySeat == seat {
		return nil, apperr.BadRequest("当前不能过牌")
	}

	now := time.Now()
	room.passCount++
	result := &doudizhu.ActionResult{
		ActionType: actionType(auto, "pass_turn"),
		Seat:       seat,
		ActorName:  room.players[seat].name,
		HandCount:  len(room.round.Hands[seat]),
		Status:     room.round.Phase,
		Message:    room.players[seat].name + " 选择不出",
	}

	if room.passCount >= doudizhuMaxPlayers-1 {
		next := *room.lastPlaySeat
		room.round.CurrentTurn = &next
		room.lastPlay = nil
		room.lastPlayCards = nil
		room.lastPlaySeat = nil
		room.passCount = 0
		result.NextTurn = &next
		result.Message = "其余玩家都已过牌，重新轮到领先玩家出牌"
	} else {
		next := seat.Next()
		room.round.CurrentTurn = &next
		result.NextTurn = &next
	}

	room.updatedAt = now
	s.setTurnExpiryLocked(room)
	room.actions = appendActionRecord(room.actions, doudizhu.ActionRecord{
		Seat:       seat,
		ActionType: actionType(auto, "pass_turn"),
		At:         now,
		Message:    result.Message,
		ActorName:  room.players[seat].name,
	})
	return result, nil
}

func (s *DoudizhuRoomService) drainAutoActorsLocked(room *doudizhuRoomState) {
	for room.round != nil && (room.round.Phase == doudizhu.RoundPhaseBidding || room.round.Phase == doudizhu.RoundPhasePlaying) {
		seat, player, ok := currentAutoActor(room)
		if !ok {
			return
		}

		switch room.round.Phase {
		case doudizhu.RoundPhaseBidding:
			score := chooseDoudizhuBid(room.round.Hands[seat], room.round.HighestBid)
			if _, err := s.applyBidLocked(room, player.sessionID, score); err != nil {
				s.logger.Warn("auto bid failed", zap.Error(err), zap.String("session_id", player.sessionID))
				return
			}
		case doudizhu.RoundPhasePlaying:
			cards := chooseDoudizhuPlay(room.round.Hands[seat], room.lastPlay, room.lastPlaySeat, seat)
			if len(cards) == 0 {
				if _, err := s.passTurnLocked(room, player.sessionID, true); err != nil {
					s.logger.Warn("auto pass failed", zap.Error(err), zap.String("session_id", player.sessionID))
					return
				}
				continue
			}
			if _, err := s.playCardsLocked(room, player.sessionID, cards, true); err != nil {
				s.logger.Warn("auto play failed", zap.Error(err), zap.String("session_id", player.sessionID))
				if _, passErr := s.passTurnLocked(room, player.sessionID, true); passErr != nil {
					s.logger.Warn("auto fallback pass failed", zap.Error(passErr), zap.String("session_id", player.sessionID))
					return
				}
			}
		}
	}
}

func buildDoudizhuMatch(room *doudizhuRoomState) (*doudizhu.Match, []doudizhu.MatchPlayerResult, []doudizhu.ActionEvent) {
	if room == nil || room.round == nil || room.round.Landlord == nil || room.winningSide == nil {
		return nil, nil, nil
	}

	finishedAt := room.updatedAt
	room.round.FinishedAt = &finishedAt
	bombCount := countDoudizhuBombActions(room.actions)
	baseScore := doudizhuMaxInt(1, room.round.HighestBid)
	multiplier := baseScore * (1 << bombCount)
	matchID := uuid.New()

	match := &doudizhu.Match{
		ID:           matchID,
		RoomID:       room.id,
		RoomCode:     room.code,
		RoomTitle:    room.title,
		MatchMode:    room.matchMode,
		StartedAt:    room.round.StartedAt,
		FinishedAt:   finishedAt,
		LandlordSeat: *room.round.Landlord,
		WinnerSide:   *room.winningSide,
		Multiplier:   multiplier,
		BombCount:    bombCount,
		Spring:       false,
		AntiSpring:   false,
		CreatedAt:    finishedAt,
	}

	results := make([]doudizhu.MatchPlayerResult, 0, len(room.players))
	for _, seat := range []doudizhu.Seat{doudizhu.Seat0, doudizhu.Seat1, doudizhu.Seat2} {
		player := room.players[seat]
		if player == nil {
			continue
		}
		role := room.round.Roles[seat]
		isWinner := role == *room.winningSide
		scoreDelta := doudizhuScoreDelta(role, *room.winningSide, multiplier)
		results = append(results, doudizhu.MatchPlayerResult{
			ID:          uuid.New(),
			MatchID:     matchID,
			SessionID:   player.sessionID,
			UserID:      cloneUUIDPtr(player.userID),
			IsBot:       player.isBot,
			BotLevel:    player.botLevel,
			Seat:        seat,
			PlayerName:  player.name,
			DisplayName: player.name,
			Role:        role,
			BidScore:    bidScoreForSeat(room.round.BidHistory, seat),
			CardsLeft:   len(room.round.Hands[seat]),
			IsWinner:    isWinner,
			ScoreDelta:  scoreDelta,
			CreatedAt:   finishedAt,
		})
	}
	slices.SortFunc(results, func(a, b doudizhu.MatchPlayerResult) int {
		switch {
		case a.ScoreDelta > b.ScoreDelta:
			return -1
		case a.ScoreDelta < b.ScoreDelta:
			return 1
		case a.CardsLeft < b.CardsLeft:
			return -1
		case a.CardsLeft > b.CardsLeft:
			return 1
		default:
			return strings.Compare(a.PlayerName, b.PlayerName)
		}
	})
	match.Results = append([]doudizhu.MatchPlayerResult(nil), results...)

	events := make([]doudizhu.ActionEvent, 0, len(room.actions))
	turnNo := 0
	for index, action := range room.actions {
		if action.ActionType == "play_cards" || action.ActionType == "auto_play_cards" {
			turnNo++
		}
		player := room.players[action.Seat]
		displayName := ""
		var userID *uuid.UUID
		if player != nil {
			displayName = player.name
			userID = cloneUUIDPtr(player.userID)
		}
		events = append(events, doudizhu.ActionEvent{
			ID:              uuid.New(),
			MatchID:         matchID,
			TurnNo:          turnNo,
			ActionIndex:     index + 1,
			SessionID:       sessionIDForAction(room, action.Seat),
			UserID:          userID,
			PlayerName:      action.ActorName,
			DisplayName:     displayName,
			Seat:            action.Seat,
			ActionType:      action.ActionType,
			Cards:           append([]doudizhu.Card(nil), action.Cards...),
			Combo:           comboPtr(action.Combo),
			MultiplierAfter: multiplierAfterAction(room.actions[:index+1], baseScore),
			OccurredAt:      action.At,
		})
	}

	return match, results, events
}

func currentAutoActor(room *doudizhuRoomState) (doudizhu.Seat, *doudizhuPlayerState, bool) {
	if room.round == nil {
		return 0, nil, false
	}
	switch room.round.Phase {
	case doudizhu.RoundPhaseBidding:
		seat := room.round.CurrentBidder
		player := room.players[seat]
		if player == nil {
			return 0, nil, false
		}
		if player.isBot || player.autoPlay || !player.connected {
			return seat, player, true
		}
	case doudizhu.RoundPhasePlaying:
		if room.round.CurrentTurn == nil {
			return 0, nil, false
		}
		seat := *room.round.CurrentTurn
		player := room.players[seat]
		if player == nil {
			return 0, nil, false
		}
		if player.isBot || player.autoPlay || !player.connected {
			return seat, player, true
		}
	}
	return 0, nil, false
}

func snapshotDoudizhuRoom(room *doudizhuRoomState) *doudizhu.Room {
	players := make([]doudizhu.RoomPlayer, 0, len(room.players))
	for _, seat := range []doudizhu.Seat{doudizhu.Seat0, doudizhu.Seat1, doudizhu.Seat2} {
		player, ok := room.players[seat]
		if !ok {
			continue
		}
		cardCount := 0
		role := doudizhu.PlayerRoleFarmer
		if room.round != nil {
			cardCount = len(room.round.Hands[seat])
			role = room.round.Roles[seat]
		}
		players = append(players, doudizhu.RoomPlayer{
			SessionID: player.sessionID,
			UserID:    cloneUUIDPtr(player.userID),
			Seat:      player.seat,
			Name:      player.name,
			Ready:     player.ready,
			Connected: player.connected,
			IsHost:    player.isHost,
			IsBot:     player.isBot,
			BotLevel:  player.botLevel,
			AutoPlay:  player.autoPlay,
			CardCount: cardCount,
			Role:      role,
			JoinedAt:  player.joinedAt,
			UpdatedAt: player.updatedAt,
		})
	}

	var currentBidder *doudizhu.Seat
	var highestBidder *doudizhu.Seat
	var landlord *doudizhu.Seat
	var currentTurn *doudizhu.Seat
	var bidHistory []doudizhu.BidRecord
	if room.round != nil {
		currentBidder = seatPtr(room.round.CurrentBidder)
		highestBidder = seatPtrFromPtr(room.round.HighestBidder)
		landlord = seatPtrFromPtr(room.round.Landlord)
		currentTurn = seatPtrFromPtr(room.round.CurrentTurn)
		bidHistory = make([]doudizhu.BidRecord, len(room.actions)) // overwritten below if records exist
		_ = bidHistory
	}
	bids := extractDoudizhuBidHistory(room.actions)
	actions := append([]doudizhu.ActionRecord(nil), room.actions...)

	var lastPlay *doudizhu.Combo
	if room.lastPlay != nil {
		cloned := *room.lastPlay
		lastPlay = &cloned
	}

	var bottomCards []doudizhu.Card
	if room.round != nil && room.round.Landlord != nil {
		bottomCards = append([]doudizhu.Card(nil), room.round.BottomCards...)
	}

	return &doudizhu.Room{
		ID:            room.id,
		Code:          room.code,
		Title:         room.title,
		MatchMode:     room.matchMode,
		Status:        room.status,
		HostSessionID: room.hostSessionID,
		PlayerCount:   len(players),
		ReadyCount:    countDoudizhuReadyPlayers(players),
		CurrentBidder: currentBidder,
		HighestBid:    currentHighestBid(room),
		HighestBidder: highestBidder,
		Landlord:      landlord,
		CurrentTurn:   currentTurn,
		LastPlay:      lastPlay,
		LastPlaySeat:  seatPtrFromPtr(room.lastPlaySeat),
		WinningSide:   rolePtrFromPtr(room.winningSide),
		TurnExpiresAt: timePtr(room.turnExpiresAt),
		BottomCards:   bottomCards,
		Players:       players,
		BidHistory:    bids,
		RecentActions: actions,
		CreatedAt:     room.createdAt,
		UpdatedAt:     room.updatedAt,
	}
}

func snapshotDoudizhuPrivateState(room *doudizhuRoomState, seat doudizhu.Seat) *doudizhu.PrivateState {
	if room.round == nil {
		return &doudizhu.PrivateState{
			SessionID: room.players[seat].sessionID,
			RoomID:    room.id,
			Status:    room.status,
			Hand:      []doudizhu.Card{},
			CanPass:   false,
			Role:      doudizhu.PlayerRoleFarmer,
		}
	}
	role := room.round.Roles[seat]
	var bottomCards []doudizhu.Card
	if room.round.Landlord != nil && (*room.round.Landlord == seat || room.round.Phase == doudizhu.RoundPhaseSettlement) {
		bottomCards = append([]doudizhu.Card(nil), room.round.BottomCards...)
	}
	return &doudizhu.PrivateState{
		SessionID:     room.players[seat].sessionID,
		RoomID:        room.id,
		Status:        room.round.Phase,
		Hand:          append([]doudizhu.Card{}, room.round.Hands[seat]...),
		CanPass:       doudizhuCanPass(room, seat),
		Role:          role,
		BottomCards:   append([]doudizhu.Card{}, bottomCards...),
		LastPlay:      comboPtr(room.lastPlay),
		LastPlaySeat:  seatPtrFromPtr(room.lastPlaySeat),
		TurnExpiresAt: timePtr(room.turnExpiresAt),
	}
}

func doudizhuCanPass(room *doudizhuRoomState, seat doudizhu.Seat) bool {
	return room.round != nil &&
		room.round.Phase == doudizhu.RoundPhasePlaying &&
		room.round.CurrentTurn != nil &&
		*room.round.CurrentTurn == seat &&
		room.lastPlay != nil &&
		room.lastPlaySeat != nil &&
		*room.lastPlaySeat != seat
}

func (s *DoudizhuRoomService) setTurnExpiryLocked(room *doudizhuRoomState) {
	if s.turnDuration <= 0 || room.round == nil {
		room.turnExpiresAt = nil
		return
	}
	expiresAt := time.Now().Add(s.turnDuration)
	room.turnExpiresAt = &expiresAt
}

func allDoudizhuPlayersReady(room *doudizhuRoomState) bool {
	if len(room.players) != doudizhuMaxPlayers {
		return false
	}
	for _, player := range room.players {
		if player.isBot {
			continue
		}
		if !player.ready || !player.connected {
			return false
		}
	}
	return true
}

func roomHasHumanPlayer(room *doudizhuRoomState) bool {
	for _, player := range room.players {
		if !player.isBot {
			return true
		}
	}
	return false
}

func pickNextHostSessionID(room *doudizhuRoomState) string {
	for _, seat := range []doudizhu.Seat{doudizhu.Seat0, doudizhu.Seat1, doudizhu.Seat2} {
		player := room.players[seat]
		if player == nil || player.isBot {
			continue
		}
		return player.sessionID
	}
	return ""
}

func firstAvailableDoudizhuSeat(room *doudizhuRoomState) (doudizhu.Seat, bool) {
	for _, seat := range []doudizhu.Seat{doudizhu.Seat0, doudizhu.Seat1, doudizhu.Seat2} {
		if _, exists := room.players[seat]; !exists {
			return seat, true
		}
	}
	return 0, false
}

func appendActionRecord(items []doudizhu.ActionRecord, item doudizhu.ActionRecord) []doudizhu.ActionRecord {
	next := append(items, item)
	if len(next) > doudizhuRecentActionsLimit {
		next = append([]doudizhu.ActionRecord(nil), next[len(next)-doudizhuRecentActionsLimit:]...)
	}
	return next
}

func currentHighestBid(room *doudizhuRoomState) int {
	if room.round == nil {
		return 0
	}
	return room.round.HighestBid
}

func currentDoudizhuTurn(room *doudizhuRoomState) *doudizhu.Seat {
	if room.round == nil {
		return nil
	}
	if room.round.Phase == doudizhu.RoundPhaseBidding {
		return seatPtr(room.round.CurrentBidder)
	}
	return seatPtrFromPtr(room.round.CurrentTurn)
}

func countDoudizhuReadyPlayers(players []doudizhu.RoomPlayer) int {
	total := 0
	for _, player := range players {
		if player.Ready {
			total++
		}
	}
	return total
}

func extractDoudizhuBidHistory(actions []doudizhu.ActionRecord) []doudizhu.BidRecord {
	items := make([]doudizhu.BidRecord, 0)
	for _, action := range actions {
		if action.ActionType != "bid" && action.ActionType != "auto_bid" {
			continue
		}
		score := 0
		if strings.Contains(action.Message, "叫 3 分") {
			score = 3
		} else if strings.Contains(action.Message, "叫 2 分") {
			score = 2
		} else if strings.Contains(action.Message, "叫 1 分") {
			score = 1
		}
		items = append(items, doudizhu.BidRecord{
			Seat:  action.Seat,
			Score: score,
			At:    action.At,
		})
	}
	return items
}

func mapDoudizhuRoundError(err error) error {
	switch err {
	case doudizhu.ErrRoundNotBidding:
		return apperr.BadRequest("当前不在叫分阶段")
	case doudizhu.ErrNotCurrentBidder:
		return apperr.ErrForbidden
	case doudizhu.ErrInvalidBidScore:
		return apperr.BadRequest("当前叫分无效")
	case doudizhu.ErrInvalidSeat:
		return apperr.BadRequest("无效座位")
	default:
		return err
	}
}

func bidActionMessage(name string, score int) string {
	if score <= 0 {
		return name + " 选择不叫"
	}
	return name + " 叫 " + string(rune('0'+score)) + " 分"
}

func sanitizeDoudizhuTitle(mode doudizhu.MatchMode, value string) string {
	title := strings.TrimSpace(value)
	switch {
	case title != "":
		return title
	case mode == doudizhu.MatchModeDemoAI:
		return doudizhuDefaultDemoTitle
	default:
		return doudizhuDefaultRoomTitle
	}
}

func sanitizeDoudizhuPlayerName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return doudizhuDefaultPlayerName
	}
	return name
}

func demoBotName(seat doudizhu.Seat) string {
	switch seat {
	case doudizhu.Seat1:
		return "机器人阿橘"
	case doudizhu.Seat2:
		return "机器人小夜"
	default:
		return "机器人"
	}
}

func doudizhuPlayerHasCards(hand []doudizhu.Card, cards []doudizhu.Card) bool {
	counts := make(map[doudizhu.Card]int, len(hand))
	for _, card := range hand {
		counts[card]++
	}
	for _, card := range cards {
		counts[card]--
		if counts[card] < 0 {
			return false
		}
	}
	return true
}

func doudizhuRemoveCards(hand []doudizhu.Card, cards []doudizhu.Card) []doudizhu.Card {
	toRemove := make(map[doudizhu.Card]int, len(cards))
	for _, card := range cards {
		toRemove[card]++
	}
	next := make([]doudizhu.Card, 0, len(hand)-len(cards))
	for _, card := range hand {
		if toRemove[card] > 0 {
			toRemove[card]--
			continue
		}
		next = append(next, card)
	}
	return doudizhuSortedCards(next)
}

func seatPtr(value doudizhu.Seat) *doudizhu.Seat {
	return &value
}

func seatPtrFromPtr(value *doudizhu.Seat) *doudizhu.Seat {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func rolePtr(value doudizhu.PlayerRole) *doudizhu.PlayerRole {
	return &value
}

func rolePtrFromPtr(value *doudizhu.PlayerRole) *doudizhu.PlayerRole {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func comboPtr(value *doudizhu.Combo) *doudizhu.Combo {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func timePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func actionType(auto bool, base string) string {
	if auto {
		return "auto_" + base
	}
	return base
}

func doudizhuScoreDelta(role, winningSide doudizhu.PlayerRole, multiplier int) int {
	switch {
	case role == doudizhu.PlayerRoleLandlord && winningSide == doudizhu.PlayerRoleLandlord:
		return multiplier * 2
	case role == doudizhu.PlayerRoleLandlord:
		return -multiplier * 2
	case winningSide == doudizhu.PlayerRoleFarmer:
		return multiplier
	default:
		return -multiplier
	}
}

func bidScoreForSeat(records []DoudizhuBidRecord, seat doudizhu.Seat) int {
	score := 0
	for _, record := range records {
		if record.Seat == seat && record.Score > score {
			score = record.Score
		}
	}
	return score
}

func countDoudizhuBombActions(actions []doudizhu.ActionRecord) int {
	total := 0
	for _, action := range actions {
		if action.Combo == nil {
			continue
		}
		if action.Combo.Type == doudizhu.ComboBomb || action.Combo.Type == doudizhu.ComboRocket {
			total++
		}
	}
	return total
}

func sessionIDForAction(room *doudizhuRoomState, seat doudizhu.Seat) string {
	player := room.players[seat]
	if player == nil {
		return ""
	}
	return player.sessionID
}

func multiplierAfterAction(actions []doudizhu.ActionRecord, baseScore int) int {
	totalBombs := countDoudizhuBombActions(actions)
	return doudizhuMaxInt(1, baseScore) * (1 << totalBombs)
}

func doudizhuMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *DoudizhuRoomService) maybeBuildPersistPayloadLocked(room *doudizhuRoomState) *doudizhuPersistPayload {
	if room == nil || room.round == nil || room.status != doudizhu.RoundPhaseSettlement || room.matchPersisted {
		return nil
	}
	match, results, events := buildDoudizhuMatch(room)
	if match == nil {
		return nil
	}
	room.matchPersisted = true
	return &doudizhuPersistPayload{
		match:   match,
		results: results,
		events:  events,
	}
}

func (s *DoudizhuRoomService) persistMatch(payload *doudizhuPersistPayload) {
	if s.repo == nil || payload == nil || payload.match == nil {
		return
	}
	if err := s.repo.SaveMatch(context.Background(), payload.match, payload.results, payload.events); err != nil {
		s.logger.Warn("failed to persist doudizhu match", zap.Error(err), zap.String("match_id", payload.match.ID.String()))
	}
}

func (s *DoudizhuRoomService) notify(roomID uuid.UUID) {
	if s.notifier != nil {
		s.notifier(roomID)
	}
}
