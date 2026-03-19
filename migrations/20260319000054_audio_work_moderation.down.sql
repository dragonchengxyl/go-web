DROP INDEX IF EXISTS idx_audio_works_moderation_status_published_at;

ALTER TABLE audio_works
    DROP CONSTRAINT IF EXISTS audio_works_moderation_status_check;

ALTER TABLE audio_works
    DROP COLUMN IF EXISTS moderation_note,
    DROP COLUMN IF EXISTS moderation_status;
