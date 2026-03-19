ALTER TABLE user_bookmarks
    DROP CONSTRAINT IF EXISTS user_bookmarks_target_type_check;

ALTER TABLE user_bookmarks
    ADD CONSTRAINT user_bookmarks_target_type_check CHECK (
        target_type IN ('post', 'group', 'event')
    );

DROP INDEX IF EXISTS idx_audio_work_likes_work_id;
DROP TABLE IF EXISTS audio_work_likes;

ALTER TABLE audio_works
    DROP COLUMN IF EXISTS comment_count,
    DROP COLUMN IF EXISTS like_count;
