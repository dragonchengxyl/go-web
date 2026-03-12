-- Migration 021: Create user_follows table

CREATE TABLE IF NOT EXISTS user_follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CONSTRAINT no_self_follow CHECK (follower_id != followee_id)
);

CREATE INDEX IF NOT EXISTS idx_user_follows_follower_id ON user_follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_followee_id ON user_follows(followee_id);
