DROP INDEX IF EXISTS idx_posts_embedding;
ALTER TABLE posts DROP COLUMN IF EXISTS embedding;
