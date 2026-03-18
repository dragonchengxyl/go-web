CREATE TABLE IF NOT EXISTS audio_jobs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    task_type TEXT NOT NULL,
    status TEXT NOT NULL,
    source_audio_url TEXT,
    reference_audio_url TEXT,
    prompt TEXT,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    CONSTRAINT audio_jobs_task_type_check CHECK (
        task_type IN ('ai_music', 'voice_convert', 'voice_enhance', 'audio_master')
    ),
    CONSTRAINT audio_jobs_status_check CHECK (
        status IN ('queued', 'running', 'succeeded', 'failed')
    )
);

CREATE INDEX IF NOT EXISTS idx_audio_jobs_user_created_at
    ON audio_jobs (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audio_jobs_status_created_at
    ON audio_jobs (status, created_at DESC);
