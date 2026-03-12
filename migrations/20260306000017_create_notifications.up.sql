CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id),
    type       VARCHAR(50) NOT NULL,
    title      VARCHAR(255) NOT NULL,
    body       TEXT,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
