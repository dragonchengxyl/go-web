ALTER TABLE audio_works
    ADD COLUMN IF NOT EXISTS moderation_status TEXT NOT NULL DEFAULT 'approved',
    ADD COLUMN IF NOT EXISTS moderation_note TEXT;

ALTER TABLE audio_works
    ADD CONSTRAINT audio_works_moderation_status_check
        CHECK (moderation_status IN ('pending', 'approved', 'blocked'));

UPDATE audio_works
SET moderation_status = 'approved'
WHERE moderation_status IS NULL OR moderation_status = '';

CREATE INDEX IF NOT EXISTS idx_audio_works_moderation_status_published_at
    ON audio_works (moderation_status, published_at DESC);
