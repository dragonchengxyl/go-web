CREATE TABLE IF NOT EXISTS assistant_media_analysis_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_url TEXT NOT NULL UNIQUE,
    alt_text TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    image_summary TEXT NOT NULL DEFAULT '',
    moderation_summary TEXT NOT NULL DEFAULT '',
    risk_level VARCHAR(32) NOT NULL DEFAULT '',
    safety_notes JSONB NOT NULL DEFAULT '[]'::jsonb,
    provider VARCHAR(64) NOT NULL DEFAULT '',
    model VARCHAR(128) NOT NULL DEFAULT '',
    fallback BOOLEAN NOT NULL DEFAULT false,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assistant_media_analysis_expires_at
    ON assistant_media_analysis_cache(expires_at DESC);
