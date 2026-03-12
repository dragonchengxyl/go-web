CREATE TABLE IF NOT EXISTS events (
    id              UUID PRIMARY KEY,
    organizer_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    location        TEXT NOT NULL DEFAULT '',
    is_online       BOOLEAN NOT NULL DEFAULT FALSE,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    max_capacity    INT NOT NULL DEFAULT 0,
    tags            JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'published',
    attendee_count  INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_organizer_id ON events(organizer_id);
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
