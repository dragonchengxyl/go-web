-- Migration 039: Create AI assistant conversations and messages tables

CREATE TABLE IF NOT EXISTS assistant_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assistant_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES assistant_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL DEFAULT '',
    cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assistant_conversations_user_id ON assistant_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_assistant_conversations_updated_at ON assistant_conversations(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_assistant_messages_conversation_id ON assistant_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_assistant_messages_created_at ON assistant_messages(created_at ASC);
