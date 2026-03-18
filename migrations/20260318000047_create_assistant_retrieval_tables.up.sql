CREATE TABLE IF NOT EXISTS assistant_knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type VARCHAR(32) NOT NULL,
    source_key VARCHAR(128) NOT NULL,
    chunk_index INT NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    href VARCHAR(255) NOT NULL DEFAULT '',
    meta VARCHAR(255) NOT NULL DEFAULT '',
    source_label VARCHAR(64) NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    search_text TEXT NOT NULL DEFAULT '',
    search_vector tsvector NOT NULL,
    embedding JSONB NOT NULL DEFAULT '[]'::jsonb,
    indexed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT chk_assistant_knowledge_source_type
        CHECK (source_type IN ('page', 'post', 'group', 'event')),
    CONSTRAINT uq_assistant_knowledge_source_chunk
        UNIQUE (source_type, source_key, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_assistant_knowledge_source
    ON assistant_knowledge_documents(source_type, source_key);

CREATE INDEX IF NOT EXISTS idx_assistant_knowledge_indexed_at
    ON assistant_knowledge_documents(indexed_at DESC);

CREATE INDEX IF NOT EXISTS idx_assistant_knowledge_source_updated_at
    ON assistant_knowledge_documents(source_updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_assistant_knowledge_search_vector
    ON assistant_knowledge_documents USING GIN(search_vector);

CREATE TABLE IF NOT EXISTS assistant_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    response_id UUID NOT NULL UNIQUE,
    conversation_id UUID REFERENCES assistant_conversations(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    value VARCHAR(16) NOT NULL CHECK (value IN ('helpful', 'unhelpful')),
    query_text TEXT NOT NULL DEFAULT '',
    reply_excerpt TEXT NOT NULL DEFAULT '',
    provider VARCHAR(64) NOT NULL DEFAULT '',
    intent VARCHAR(64) NOT NULL DEFAULT '',
    fallback BOOLEAN NOT NULL DEFAULT false,
    page_path VARCHAR(255) NOT NULL DEFAULT '',
    source_counts JSONB NOT NULL DEFAULT '{}'::jsonb,
    cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assistant_feedback_created_at
    ON assistant_feedback(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_assistant_feedback_value
    ON assistant_feedback(value);
