ALTER TABLE posts ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) NOT NULL DEFAULT 'pending';
CREATE INDEX IF NOT EXISTS idx_posts_moderation_status ON posts(moderation_status);
