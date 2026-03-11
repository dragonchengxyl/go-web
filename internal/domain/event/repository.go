package event

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence operations for events.
type Repository interface {
	Create(ctx context.Context, e *Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	List(ctx context.Context, filter ListFilter) ([]*Event, int64, error)
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Attendance
	Attend(ctx context.Context, a *EventAttendee) error
	UpdateAttendee(ctx context.Context, a *EventAttendee) error
	GetAttendee(ctx context.Context, eventID, userID uuid.UUID) (*EventAttendee, error)
	ListAttendees(ctx context.Context, eventID uuid.UUID, page, pageSize int) ([]*EventAttendee, int64, error)

	// Counts
	IncrementAttendeeCount(ctx context.Context, eventID uuid.UUID) error
	DecrementAttendeeCount(ctx context.Context, eventID uuid.UUID) error

	// User-specific
	ListByOrganizer(ctx context.Context, organizerID uuid.UUID, page, pageSize int) ([]*Event, int64, error)
	ListAttending(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Event, int64, error)
}
