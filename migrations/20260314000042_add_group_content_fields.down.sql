ALTER TABLE groups
    DROP COLUMN IF EXISTS featured_post_id,
    DROP COLUMN IF EXISTS rules,
    DROP COLUMN IF EXISTS announcement;
