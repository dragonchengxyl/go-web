CREATE TABLE IF NOT EXISTS groups (
    id              UUID PRIMARY KEY,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    avatar_key      TEXT,
    tags            JSONB NOT NULL DEFAULT '[]',
    privacy         TEXT NOT NULL DEFAULT 'public',
    member_count    INT NOT NULL DEFAULT 0,
    post_count      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_privacy ON groups(privacy);
CREATE INDEX IF NOT EXISTS idx_groups_member_count ON groups(member_count DESC);
