ALTER TABLE audio_works
    ADD COLUMN IF NOT EXISTS like_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS comment_count INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS audio_work_likes (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_id UUID NOT NULL REFERENCES audio_works(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, work_id)
);

CREATE INDEX IF NOT EXISTS idx_audio_work_likes_work_id
    ON audio_work_likes (work_id, created_at DESC);

ALTER TABLE user_bookmarks
    DROP CONSTRAINT IF EXISTS user_bookmarks_target_type_check;

ALTER TABLE user_bookmarks
    ADD CONSTRAINT user_bookmarks_target_type_check CHECK (
        target_type IN ('post', 'group', 'event', 'audio_work')
    );
