ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS announcement TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS rules TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS featured_post_id UUID REFERENCES posts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_groups_featured_post_id ON groups(featured_post_id) WHERE featured_post_id IS NOT NULL;
