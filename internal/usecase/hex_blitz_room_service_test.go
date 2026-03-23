package usecase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/gameplay"
	"go.uber.org/zap"
)

func TestHexBlitzRoomServiceCreateJoinAndStart(t *testing.T) {
	svc := NewHexBlitzRoomService(
		zap.NewNop(),
		WithHexBlitzRoomTiming(20*time.Millisecond, 40*time.Millisecond),
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

	updatedScoreRoom, err := svc.UpdateScore(UpdateHexBlitzScoreInput{
		RoomID:    hostRoom.ID,
		SessionID: "guest-session",
		Score:     1880,
	})
	require.NoError(t, err)
	var guestScore int
	for _, player := range updatedScoreRoom.Players {
		if player.SessionID == "guest-session" {
			guestScore = player.Score
		}
	}
	require.Equal(t, 1880, guestScore)

	time.Sleep(50 * time.Millisecond)

	finishedRoom, ok := svc.GetRoom(hostRoom.ID)
	require.True(t, ok)
	require.Equal(t, gameplay.RoomStatusFinished, finishedRoom.Status)
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
