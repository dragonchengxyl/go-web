package block

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserBlock struct {
	BlockerID uuid.UUID `json:"blocker_id"`
	BlockedID uuid.UUID `json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	Block(ctx context.Context, blockerID, blockedID uuid.UUID) error
	Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
	ListBlockedIDs(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error)
}

var ErrNotBlocked = errors.New("not blocked")
