package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/pkg/apperr"
)

// EventService handles event business logic.
type EventService struct {
	repo event.Repository
}

// NewEventService creates a new EventService.
func NewEventService(repo event.Repository) *EventService {
	return &EventService{repo: repo}
}

// CreateEventInput is the input for creating an event.
type CreateEventInput struct {
	OrganizerID uuid.UUID
	Title       string
	Description string
	Location    string
	IsOnline    bool
	StartTime   time.Time
	EndTime     time.Time
	MaxCapacity int
	Tags        []string
}

// CreateEvent creates a new event.
func (s *EventService) CreateEvent(ctx context.Context, input CreateEventInput) (*event.Event, error) {
	if input.Title == "" {
		return nil, apperr.BadRequest("活动标题不能为空")
	}
	if input.StartTime.IsZero() || input.EndTime.IsZero() {
		return nil, apperr.BadRequest("活动时间不能为空")
	}
	if input.EndTime.Before(input.StartTime) {
		return nil, apperr.BadRequest("结束时间不能早于开始时间")
	}
	now := time.Now()
	e := &event.Event{
		ID:            uuid.New(),
		OrganizerID:   input.OrganizerID,
		Title:         input.Title,
		Description:   input.Description,
		Location:      input.Location,
		IsOnline:      input.IsOnline,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		MaxCapacity:   input.MaxCapacity,
		Tags:          input.Tags,
		Status:        event.EventStatusPublished,
		AttendeeCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建活动失败", err)
	}
	return e, nil
}

// GetEvent returns an event by ID.
func (s *EventService) GetEvent(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

// ListEventsInput is the input for listing events.
type ListEventsInput struct {
	Status   *event.EventStatus
	Tags     []string
	Page     int
	PageSize int
}

// ListEvents returns a paginated list of events.
func (s *EventService) ListEvents(ctx context.Context, input ListEventsInput) ([]*event.Event, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	return s.repo.List(ctx, event.ListFilter{
		Status:   input.Status,
		Tags:     input.Tags,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
}

// UpdateEventInput is the input for updating an event.
type UpdateEventInput struct {
	Title       string
	Description string
	Location    string
	IsOnline    bool
	StartTime   time.Time
	EndTime     time.Time
	MaxCapacity int
	Tags        []string
	Status      event.EventStatus
}

// UpdateEvent updates an event (organizer only).
func (s *EventService) UpdateEvent(ctx context.Context, userID, eventID uuid.UUID, input UpdateEventInput) (*event.Event, error) {
	e, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	if e.OrganizerID != userID {
		return nil, apperr.New(apperr.CodeForbidden, "无权修改此活动")
	}

	e.Title = input.Title
	e.Description = input.Description
	e.Location = input.Location
	e.IsOnline = input.IsOnline
	e.StartTime = input.StartTime
	e.EndTime = input.EndTime
	e.MaxCapacity = input.MaxCapacity
	e.Tags = input.Tags
	if input.Status != "" {
		e.Status = input.Status
	}
	e.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新活动失败", err)
	}
	return e, nil
}

// CancelEvent cancels an event (organizer only).
func (s *EventService) CancelEvent(ctx context.Context, userID, eventID uuid.UUID) error {
	e, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}
	if e.OrganizerID != userID {
		return apperr.New(apperr.CodeForbidden, "无权取消此活动")
	}
	return s.repo.Delete(ctx, eventID)
}

// AttendEvent records or updates attendance.
func (s *EventService) AttendEvent(ctx context.Context, userID, eventID uuid.UUID, status event.AttendeeStatus) error {
	e, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}
	if e.Status != event.EventStatusPublished {
		return apperr.BadRequest("活动不可报名")
	}
	if e.MaxCapacity > 0 && e.AttendeeCount >= e.MaxCapacity && status == event.AttendeeStatusAttending {
		return apperr.BadRequest("活动已满员")
	}

	existing, err := s.repo.GetAttendee(ctx, eventID, userID)
	if err != nil {
		return err
	}

	a := &event.EventAttendee{
		EventID:  eventID,
		UserID:   userID,
		Status:   status,
		JoinedAt: time.Now(),
	}

	if existing == nil {
		if err := s.repo.Attend(ctx, a); err != nil {
			return apperr.Wrap(apperr.CodeInternalError, "报名失败", err)
		}
		if status == event.AttendeeStatusAttending {
			_ = s.repo.IncrementAttendeeCount(ctx, eventID)
		}
	} else {
		if err := s.repo.UpdateAttendee(ctx, a); err != nil {
			return apperr.Wrap(apperr.CodeInternalError, "更新报名状态失败", err)
		}
		// Adjust count
		if existing.Status == event.AttendeeStatusAttending && status != event.AttendeeStatusAttending {
			_ = s.repo.DecrementAttendeeCount(ctx, eventID)
		} else if existing.Status != event.AttendeeStatusAttending && status == event.AttendeeStatusAttending {
			_ = s.repo.IncrementAttendeeCount(ctx, eventID)
		}
	}
	return nil
}

// ListAttendees returns attendees for an event.
func (s *EventService) ListAttendees(ctx context.Context, eventID uuid.UUID, page, pageSize int) ([]*event.EventAttendee, int64, error) {
	return s.repo.ListAttendees(ctx, eventID, page, pageSize)
}

// MyEvents returns events organized by a user.
func (s *EventService) MyEvents(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*event.Event, int64, error) {
	return s.repo.ListByOrganizer(ctx, userID, page, pageSize)
}

// MyAttending returns events a user is attending.
func (s *EventService) MyAttending(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*event.Event, int64, error) {
	return s.repo.ListAttending(ctx, userID, page, pageSize)
}
