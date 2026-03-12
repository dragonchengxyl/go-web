CREATE TABLE IF NOT EXISTS event_attendees (
    event_id    UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'attending',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_event_attendees_user_id ON event_attendees(user_id);
