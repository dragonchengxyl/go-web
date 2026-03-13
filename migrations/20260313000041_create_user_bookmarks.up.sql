CREATE TABLE IF NOT EXISTS user_bookmarks (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type  TEXT NOT NULL CHECK (target_type IN ('post', 'group', 'event')),
    target_id    UUID NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_type, target_id)
);

CREATE INDEX IF NOT EXISTS idx_user_bookmarks_user_id ON user_bookmarks(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_bookmarks_target ON user_bookmarks(target_type, target_id);
