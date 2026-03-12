-- Migration 019: Create posts table

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500),
    content TEXT NOT NULL,
    media_urls JSONB DEFAULT '[]'::jsonb,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    visibility VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'followers_only', 'private')),
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at) WHERE deleted_at IS NULL;
