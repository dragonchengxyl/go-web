CREATE TABLE IF NOT EXISTS audio_works (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_job_id UUID NOT NULL UNIQUE REFERENCES audio_jobs(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    cover_image_url TEXT,
    audio_url TEXT NOT NULL,
    duration_sec DOUBLE PRECISION NOT NULL DEFAULT 0,
    visibility TEXT NOT NULL DEFAULT 'public',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    waveform_preview JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT audio_works_visibility_check CHECK (
        visibility IN ('public', 'private')
    )
);

CREATE INDEX IF NOT EXISTS idx_audio_works_author_published_at
    ON audio_works (author_id, published_at DESC);

CREATE INDEX IF NOT EXISTS idx_audio_works_visibility_published_at
    ON audio_works (visibility, published_at DESC);
