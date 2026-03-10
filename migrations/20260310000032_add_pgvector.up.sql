CREATE EXTENSION IF NOT EXISTS vector;
ALTER TABLE posts ADD COLUMN IF NOT EXISTS embedding vector(1536);
CREATE INDEX IF NOT EXISTS idx_posts_embedding ON posts USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
