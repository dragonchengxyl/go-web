ALTER TABLE audio_jobs
    ADD COLUMN IF NOT EXISTS attempt_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_attempts INT NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_error_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS dead_lettered_at TIMESTAMPTZ;

ALTER TABLE audio_jobs
    DROP CONSTRAINT IF EXISTS audio_jobs_status_check;

ALTER TABLE audio_jobs
    ADD CONSTRAINT audio_jobs_status_check CHECK (
        status IN ('queued', 'running', 'succeeded', 'failed', 'dead_lettered')
    );

CREATE INDEX IF NOT EXISTS idx_audio_jobs_status_next_retry_at
    ON audio_jobs (status, next_retry_at);
