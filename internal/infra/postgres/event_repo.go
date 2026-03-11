package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/event"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, e *event.Event) error {
	tags, _ := json.Marshal(e.Tags)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO events (id, organizer_id, title, description, location, is_online,
			start_time, end_time, max_capacity, tags, status, attendee_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, e.ID, e.OrganizerID, e.Title, e.Description, e.Location, e.IsOnline,
		e.StartTime, e.EndTime, e.MaxCapacity, tags, e.Status, e.AttendeeCount, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	var e event.Event
	var tags []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, organizer_id, title, description, location, is_online,
			start_time, end_time, max_capacity, tags, status, attendee_count, created_at, updated_at
		FROM events WHERE id = $1
	`, id).Scan(
		&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Location, &e.IsOnline,
		&e.StartTime, &e.EndTime, &e.MaxCapacity, &tags, &e.Status, &e.AttendeeCount,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, event.ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(tags, &e.Tags)
	return &e, nil
}

func (r *EventRepository) List(ctx context.Context, filter event.ListFilter) ([]*event.Event, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT id, organizer_id, title, description, location, is_online,
			start_time, end_time, max_capacity, tags, status, attendee_count, created_at, updated_at
		FROM events
		WHERE ($1::uuid IS NULL OR organizer_id = $1)
		  AND ($2::text IS NULL OR status = $2)
		ORDER BY start_time ASC
		LIMIT $3 OFFSET $4
	`, filter.OrganizerID, filter.Status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		var e event.Event
		var tags []byte
		if err := rows.Scan(
			&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Location, &e.IsOnline,
			&e.StartTime, &e.EndTime, &e.MaxCapacity, &tags, &e.Status, &e.AttendeeCount,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			continue
		}
		_ = json.Unmarshal(tags, &e.Tags)
		events = append(events, &e)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM events
		WHERE ($1::uuid IS NULL OR organizer_id = $1)
		  AND ($2::text IS NULL OR status = $2)
	`, filter.OrganizerID, filter.Status).Scan(&total)

	return events, total, nil
}

func (r *EventRepository) Update(ctx context.Context, e *event.Event) error {
	tags, _ := json.Marshal(e.Tags)
	_, err := r.pool.Exec(ctx, `
		UPDATE events SET title=$1, description=$2, location=$3, is_online=$4,
			start_time=$5, end_time=$6, max_capacity=$7, tags=$8, status=$9, updated_at=$10
		WHERE id=$11
	`, e.Title, e.Description, e.Location, e.IsOnline, e.StartTime, e.EndTime,
		e.MaxCapacity, tags, e.Status, e.UpdatedAt, e.ID)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE events SET status='cancelled', updated_at=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (r *EventRepository) Attend(ctx context.Context, a *event.EventAttendee) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO event_attendees (event_id, user_id, status, joined_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (event_id, user_id) DO UPDATE SET status=EXCLUDED.status
	`, a.EventID, a.UserID, a.Status, a.JoinedAt)
	return err
}

func (r *EventRepository) UpdateAttendee(ctx context.Context, a *event.EventAttendee) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE event_attendees SET status=$1 WHERE event_id=$2 AND user_id=$3
	`, a.Status, a.EventID, a.UserID)
	return err
}

func (r *EventRepository) GetAttendee(ctx context.Context, eventID, userID uuid.UUID) (*event.EventAttendee, error) {
	var a event.EventAttendee
	err := r.pool.QueryRow(ctx, `
		SELECT event_id, user_id, status, joined_at FROM event_attendees
		WHERE event_id=$1 AND user_id=$2
	`, eventID, userID).Scan(&a.EventID, &a.UserID, &a.Status, &a.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *EventRepository) ListAttendees(ctx context.Context, eventID uuid.UUID, page, pageSize int) ([]*event.EventAttendee, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT event_id, user_id, status, joined_at FROM event_attendees
		WHERE event_id=$1 ORDER BY joined_at ASC LIMIT $2 OFFSET $3
	`, eventID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var attendees []*event.EventAttendee
	for rows.Next() {
		var a event.EventAttendee
		if err := rows.Scan(&a.EventID, &a.UserID, &a.Status, &a.JoinedAt); err != nil {
			continue
		}
		attendees = append(attendees, &a)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM event_attendees WHERE event_id=$1`, eventID).Scan(&total)
	return attendees, total, nil
}

func (r *EventRepository) IncrementAttendeeCount(ctx context.Context, eventID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE events SET attendee_count=attendee_count+1 WHERE id=$1`, eventID)
	return err
}

func (r *EventRepository) DecrementAttendeeCount(ctx context.Context, eventID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE events SET attendee_count=GREATEST(0, attendee_count-1) WHERE id=$1`, eventID)
	return err
}

func (r *EventRepository) ListByOrganizer(ctx context.Context, organizerID uuid.UUID, page, pageSize int) ([]*event.Event, int64, error) {
	filter := event.ListFilter{OrganizerID: &organizerID, Page: page, PageSize: pageSize}
	return r.List(ctx, filter)
}

func (r *EventRepository) ListAttending(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*event.Event, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.organizer_id, e.title, e.description, e.location, e.is_online,
			e.start_time, e.end_time, e.max_capacity, e.tags, e.status, e.attendee_count, e.created_at, e.updated_at
		FROM events e
		JOIN event_attendees ea ON ea.event_id = e.id
		WHERE ea.user_id=$1 AND ea.status='attending'
		ORDER BY e.start_time ASC LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		var e event.Event
		var tags []byte
		if err := rows.Scan(
			&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Location, &e.IsOnline,
			&e.StartTime, &e.EndTime, &e.MaxCapacity, &tags, &e.Status, &e.AttendeeCount,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			continue
		}
		_ = json.Unmarshal(tags, &e.Tags)
		events = append(events, &e)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM events e
		JOIN event_attendees ea ON ea.event_id = e.id
		WHERE ea.user_id=$1 AND ea.status='attending'
	`, userID).Scan(&total)
	return events, total, nil
}
