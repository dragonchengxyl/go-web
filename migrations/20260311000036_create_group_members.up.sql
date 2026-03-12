CREATE TABLE IF NOT EXISTS group_members (
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT 'member',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id);
