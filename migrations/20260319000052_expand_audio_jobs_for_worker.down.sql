DROP INDEX IF EXISTS idx_audio_jobs_status_next_retry_at;

ALTER TABLE audio_jobs
    DROP CONSTRAINT IF EXISTS audio_jobs_status_check;

ALTER TABLE audio_jobs
    ADD CONSTRAINT audio_jobs_status_check CHECK (
        status IN ('queued', 'running', 'succeeded', 'failed')
    );

ALTER TABLE audio_jobs
    DROP COLUMN IF EXISTS dead_lettered_at,
    DROP COLUMN IF EXISTS last_error_at,
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS max_attempts,
    DROP COLUMN IF EXISTS attempt_count;
