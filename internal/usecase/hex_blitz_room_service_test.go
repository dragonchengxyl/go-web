package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/gameplay"
	"go.uber.org/zap"
)

func TestHexBlitzRoomServiceCreateJoinAndStart(t *testing.T) {
	repo := &stubHexBlitzRepo{}
	svc := NewHexBlitzRoomService(
		zap.NewNop(),
		WithHexBlitzRoomTiming(20*time.Millisecond, 40*time.Millisecond),
		WithHexBlitzRepository(repo),
	)

	hostRoom, err := svc.CreateRoom(CreateHexBlitzRoomInput{
		SessionID:  "host-session",
		PlayerName: "Host",
	})
	require.NoError(t, err)
	require.Equal(t, gameplay.RoomStatusWaiting, hostRoom.Status)
	require.Len(t, hostRoom.Players, 1)
	require.Equal(t, "host-session", hostRoom.HostSessionID)

	roomAfterJoin, err := svc.JoinRoom(JoinHexBlitzRoomInput{
		RoomID:     hostRoom.ID,
		SessionID:  "guest-session",
		PlayerName: "Guest",
	})
	require.NoError(t, err)
	require.Len(t, roomAfterJoin.Players, 2)

	ready := true
	_, err = svc.SetReady(SetHexBlitzReadyInput{
		RoomID:    hostRoom.ID,
		SessionID: "host-session",
		Ready:     &ready,
	})
	require.NoError(t, err)
	_, err = svc.SetReady(SetHexBlitzReadyInput{
		RoomID:    hostRoom.ID,
		SessionID: "guest-session",
		Ready:     &ready,
	})
	require.NoError(t, err)

	startedRoom, err := svc.StartMatch(hostRoom.ID, "host-session")
	require.NoError(t, err)
	require.Equal(t, gameplay.RoomStatusCountdown, startedRoom.Status)

	time.Sleep(30 * time.Millisecond)

	runningRoom, ok := svc.GetRoom(hostRoom.ID)
	require.True(t, ok)
	require.Equal(t, gameplay.RoomStatusRunning, runningRoom.Status)
	require.NotNil(t, runningRoom.StartedAt)
	require.NotNil(t, runningRoom.EndsAt)

	boardState, ok := svc.GetPlayerBoardState(hostRoom.ID, "guest-session")
	require.True(t, ok)
	require.NotEmpty(t, boardState.Tiles)

	var playableTileID string
	for _, tile := range boardState.Tiles {
		if len(hexBlitzCollectGroup(boardState.Tiles, tile.ID)) >= 2 {
			playableTileID = tile.ID
			break
		}
	}
	require.NotEmpty(t, playableTileID)

	updatedBoardState, err := svc.ApplyMove(hostRoom.ID, "guest-session", playableTileID)
	require.NoError(t, err)
	require.Greater(t, updatedBoardState.Score, 0)
	require.Equal(t, 1, updatedBoardState.Moves)

	time.Sleep(50 * time.Millisecond)

	finishedRoom, ok := svc.GetRoom(hostRoom.ID)
	require.True(t, ok)
	require.Equal(t, gameplay.RoomStatusFinished, finishedRoom.Status)
	require.Len(t, repo.savedMatches, 1)
	require.Len(t, repo.savedMatches[0].Results, 2)

	_, err = svc.SetReady(SetHexBlitzReadyInput{
		RoomID:    hostRoom.ID,
		SessionID: "host-session",
		Ready:     &ready,
	})
	require.NoError(t, err)
	_, err = svc.SetReady(SetHexBlitzReadyInput{
		RoomID:    hostRoom.ID,
		SessionID: "guest-session",
		Ready:     &ready,
	})
	require.NoError(t, err)

	restartedRoom, err := svc.StartMatch(hostRoom.ID, "host-session")
	require.NoError(t, err)
	require.Equal(t, gameplay.RoomStatusCountdown, restartedRoom.Status)
}

func TestHexBlitzRoomServiceReassignsHost(t *testing.T) {
	svc := NewHexBlitzRoomService(zap.NewNop())

	room, err := svc.CreateRoom(CreateHexBlitzRoomInput{
		SessionID:  "alpha",
		PlayerName: "Alpha",
		UserID:     uuidPtr(uuid.New()),
	})
	require.NoError(t, err)

	_, err = svc.JoinRoom(JoinHexBlitzRoomInput{
		RoomID:     room.ID,
		SessionID:  "beta",
		PlayerName: "Beta",
	})
	require.NoError(t, err)

	_, err = svc.LeaveRoom(room.ID, "alpha")
	require.NoError(t, err)

	nextRoom, ok := svc.GetRoom(room.ID)
	require.True(t, ok)
	require.Equal(t, "beta", nextRoom.HostSessionID)
	require.Len(t, nextRoom.Players, 1)
	require.True(t, nextRoom.Players[0].IsHost)
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

type stubHexBlitzRepo struct {
	mu           sync.Mutex
	savedMatches []*gameplay.Match
}

func (s *stubHexBlitzRepo) SaveMatch(_ context.Context, match *gameplay.Match, _ []gameplay.MatchResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.savedMatches = append(s.savedMatches, match)
	return nil
}

func (s *stubHexBlitzRepo) ListLeaderboard(_ context.Context, _ int) ([]*gameplay.LeaderboardEntry, error) {
	return nil, nil
}

func (s *stubHexBlitzRepo) ListRecentMatches(_ context.Context, _ int) ([]*gameplay.MatchSummary, error) {
	return nil, nil
}

func (s *stubHexBlitzRepo) ListUserRecentMatches(_ context.Context, _ uuid.UUID, _ int) ([]*gameplay.MatchSummary, error) {
	return nil, nil
}
