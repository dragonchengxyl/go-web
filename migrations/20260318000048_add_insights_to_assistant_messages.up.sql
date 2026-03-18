ALTER TABLE assistant_messages
ADD COLUMN IF NOT EXISTS insights JSONB NOT NULL DEFAULT '[]'::jsonb;
