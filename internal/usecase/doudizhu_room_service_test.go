package usecase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/doudizhu"
	"go.uber.org/zap"
)

func TestDoudizhuRoomServiceCreateDemoRoomAndBidFlow(t *testing.T) {
	svc := NewDoudizhuRoomService(
		zap.NewNop(),
		WithDoudizhuSeedSource(func() int64 { return 42 }),
		WithDoudizhuTurnDuration(5*time.Second),
	)

	room, err := svc.CreateDemoRoom(CreateDoudizhuRoomInput{
		SessionID:  "host-session",
		PlayerName: "Host",
	})
	require.NoError(t, err)
	require.Equal(t, doudizhu.MatchModeDemoAI, room.MatchMode)
	require.Len(t, room.Players, 3)

	joinedRoom, err := svc.JoinRoom(JoinDoudizhuRoomInput{
		RoomID:     room.ID,
		SessionID:  "host-session",
		PlayerName: "Host",
	})
	require.NoError(t, err)
	require.Equal(t, 3, joinedRoom.PlayerCount)

	ready := true
	_, err = svc.SetReady(SetDoudizhuReadyInput{
		RoomID:    room.ID,
		SessionID: "host-session",
		Ready:     &ready,
	})
	require.NoError(t, err)

	startedRoom, err := svc.StartRound(room.ID, "host-session")
	require.NoError(t, err)
	require.Equal(t, doudizhu.RoundPhaseBidding, startedRoom.Status)

	result, err := svc.Bid(DoudizhuBidInput{
		RoomID:    room.ID,
		SessionID: "host-session",
		Score:     1,
	})
	require.NoError(t, err)
	require.Equal(t, "bid", result.ActionType)

	afterBid, ok := svc.GetRoom(room.ID)
	require.True(t, ok)
	require.Contains(t, []doudizhu.RoundPhase{doudizhu.RoundPhasePlaying, doudizhu.RoundPhaseSettlement}, afterBid.Status)
	require.NotNil(t, afterBid.Landlord)

	privateState, exists := svc.GetPrivateState(room.ID, "host-session")
	require.True(t, exists)
	require.NotEmpty(t, privateState.Hand)
}

func TestDoudizhuRoomServicePVPStartAndPlay(t *testing.T) {
	svc := NewDoudizhuRoomService(
		zap.NewNop(),
		WithDoudizhuSeedSource(func() int64 { return 7 }),
	)

	room, err := svc.CreateRoom(CreateDoudizhuRoomInput{
		SessionID:  "alpha",
		PlayerName: "Alpha",
		UserID:     uuidPtr(uuid.New()),
	})
	require.NoError(t, err)

	_, err = svc.JoinRoom(JoinDoudizhuRoomInput{
		RoomID:     room.ID,
		SessionID:  "alpha",
		PlayerName: "Alpha",
	})
	require.NoError(t, err)
	_, err = svc.JoinRoom(JoinDoudizhuRoomInput{
		RoomID:     room.ID,
		SessionID:  "beta",
		PlayerName: "Beta",
		UserID:     uuidPtr(uuid.New()),
	})
	require.NoError(t, err)
	_, err = svc.JoinRoom(JoinDoudizhuRoomInput{
		RoomID:     room.ID,
		SessionID:  "gamma",
		PlayerName: "Gamma",
		UserID:     uuidPtr(uuid.New()),
	})
	require.NoError(t, err)

	ready := true
	for _, sessionID := range []string{"alpha", "beta", "gamma"} {
		_, err = svc.SetReady(SetDoudizhuReadyInput{
			RoomID:    room.ID,
			SessionID: sessionID,
			Ready:     &ready,
		})
		require.NoError(t, err)
	}

	startedRoom, err := svc.StartRound(room.ID, "alpha")
	require.NoError(t, err)
	require.Equal(t, doudizhu.RoundPhaseBidding, startedRoom.Status)

	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "alpha", Score: 1})
	require.NoError(t, err)
	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "beta", Score: 2})
	require.NoError(t, err)
	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "gamma", Score: 0})
	require.NoError(t, err)

	playingRoom, ok := svc.GetRoom(room.ID)
	require.True(t, ok)
	require.Equal(t, doudizhu.RoundPhasePlaying, playingRoom.Status)
	require.NotNil(t, playingRoom.CurrentTurn)
	require.NotNil(t, playingRoom.Landlord)

	landlordSession := sessionIDBySeat(playingRoom, *playingRoom.Landlord)
	privateState, exists := svc.GetPrivateState(room.ID, landlordSession)
	require.True(t, exists)
	require.NotEmpty(t, privateState.Hand)

	playResult, err := svc.PlayCards(DoudizhuPlayInput{
		RoomID:    room.ID,
		SessionID: landlordSession,
		Cards:     []doudizhu.Card{privateState.Hand[0]},
	})
	require.NoError(t, err)
	require.Equal(t, "play_cards", playResult.ActionType)
	require.Equal(t, len(privateState.Hand)-1, playResult.HandCount)

	nextState, exists := svc.GetPrivateState(room.ID, landlordSession)
	require.True(t, exists)
	require.Equal(t, len(privateState.Hand)-1, len(nextState.Hand))
}

func TestDoudizhuRoomServiceRequestHint(t *testing.T) {
	svc := NewDoudizhuRoomService(
		zap.NewNop(),
		WithDoudizhuSeedSource(func() int64 { return 7 }),
	)

	room, err := svc.CreateRoom(CreateDoudizhuRoomInput{
		SessionID:  "alpha",
		PlayerName: "Alpha",
		UserID:     uuidPtr(uuid.New()),
	})
	require.NoError(t, err)

	for _, player := range []struct {
		sessionID string
		name      string
	}{
		{sessionID: "alpha", name: "Alpha"},
		{sessionID: "beta", name: "Beta"},
		{sessionID: "gamma", name: "Gamma"},
	} {
		_, err = svc.JoinRoom(JoinDoudizhuRoomInput{
			RoomID:     room.ID,
			SessionID:  player.sessionID,
			PlayerName: player.name,
			UserID:     uuidPtr(uuid.New()),
		})
		require.NoError(t, err)
	}

	ready := true
	for _, sessionID := range []string{"alpha", "beta", "gamma"} {
		_, err = svc.SetReady(SetDoudizhuReadyInput{
			RoomID:    room.ID,
			SessionID: sessionID,
			Ready:     &ready,
		})
		require.NoError(t, err)
	}

	_, err = svc.StartRound(room.ID, "alpha")
	require.NoError(t, err)

	bidHint, err := svc.RequestHint(room.ID, "alpha")
	require.NoError(t, err)
	require.Equal(t, "bid", bidHint.ActionType)
	require.NotNil(t, bidHint.BidScore)

	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "alpha", Score: 1})
	require.NoError(t, err)
	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "beta", Score: 2})
	require.NoError(t, err)
	_, err = svc.Bid(DoudizhuBidInput{RoomID: room.ID, SessionID: "gamma", Score: 0})
	require.NoError(t, err)

	playingRoom, ok := svc.GetRoom(room.ID)
	require.True(t, ok)
	require.NotNil(t, playingRoom.Landlord)

	landlordSession := sessionIDBySeat(playingRoom, *playingRoom.Landlord)
	playHint, err := svc.RequestHint(room.ID, landlordSession)
	require.NoError(t, err)
	require.Equal(t, "play_cards", playHint.ActionType)
	require.NotEmpty(t, playHint.Cards)
	require.NotNil(t, playHint.Combo)
}

func TestDoudizhuRoomServiceTimeoutTurnsOnAutoPlay(t *testing.T) {
	svc := NewDoudizhuRoomService(
		zap.NewNop(),
		WithDoudizhuSeedSource(func() int64 { return 42 }),
		WithDoudizhuTurnDuration(40*time.Millisecond),
	)

	room, err := svc.CreateDemoRoom(CreateDoudizhuRoomInput{
		SessionID:  "host-session",
		PlayerName: "Host",
	})
	require.NoError(t, err)

	_, err = svc.JoinRoom(JoinDoudizhuRoomInput{
		RoomID:     room.ID,
		SessionID:  "host-session",
		PlayerName: "Host",
	})
	require.NoError(t, err)

	ready := true
	_, err = svc.SetReady(SetDoudizhuReadyInput{
		RoomID:    room.ID,
		SessionID: "host-session",
		Ready:     &ready,
	})
	require.NoError(t, err)

	_, err = svc.StartRound(room.ID, "host-session")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		state, ok := svc.GetRoom(room.ID)
		if !ok {
			return false
		}
		host := sessionIDBySeat(state, doudizhu.Seat0)
		privateState, exists := svc.GetPrivateState(room.ID, host)
		if !exists {
			return false
		}
		player := roomPlayerBySeat(state, doudizhu.Seat0)
		return player != nil && player.AutoPlay && privateState.Status != doudizhu.RoundPhaseBidding
	}, time.Second, 30*time.Millisecond)
}

func TestDoudizhuSpringFlags(t *testing.T) {
	room := &doudizhuRoomState{
		round: &DoudizhuRoundState{
			Landlord: seatPtr(doudizhu.Seat0),
		},
		winningSide: rolePtr(doudizhu.PlayerRoleLandlord),
		actionLog: []doudizhu.ActionRecord{
			{Seat: doudizhu.Seat0, ActionType: "play_cards"},
			{Seat: doudizhu.Seat0, ActionType: "play_cards"},
		},
	}
	spring, antiSpring := doudizhuSpringFlags(room)
	require.True(t, spring)
	require.False(t, antiSpring)

	room.winningSide = rolePtr(doudizhu.PlayerRoleFarmer)
	room.actionLog = []doudizhu.ActionRecord{
		{Seat: doudizhu.Seat0, ActionType: "play_cards"},
		{Seat: doudizhu.Seat1, ActionType: "play_cards"},
	}
	spring, antiSpring = doudizhuSpringFlags(room)
	require.False(t, spring)
	require.True(t, antiSpring)
}

func sessionIDBySeat(room *doudizhu.Room, seat doudizhu.Seat) string {
	for _, player := range room.Players {
		if player.Seat == seat {
			return player.SessionID
		}
	}
	return ""
}

func roomPlayerBySeat(room *doudizhu.Room, seat doudizhu.Seat) *doudizhu.RoomPlayer {
	for _, player := range room.Players {
		if player.Seat == seat {
			item := player
			return &item
		}
	}
	return nil
}
