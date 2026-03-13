-- Migration 040: Create assistant settings table

CREATE TABLE IF NOT EXISTS assistant_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT true,
    persona_name VARCHAR(64) NOT NULL DEFAULT '霜牙',
    system_prompt TEXT NOT NULL DEFAULT '',
    max_context_items INT NOT NULL DEFAULT 6,
    include_pages BOOLEAN NOT NULL DEFAULT true,
    include_posts BOOLEAN NOT NULL DEFAULT true,
    include_users BOOLEAN NOT NULL DEFAULT true,
    include_tags BOOLEAN NOT NULL DEFAULT true,
    include_groups BOOLEAN NOT NULL DEFAULT true,
    include_events BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO assistant_settings (
    id,
    enabled,
    persona_name,
    system_prompt,
    max_context_items,
    include_pages,
    include_posts,
    include_users,
    include_tags,
    include_groups,
    include_events
)
VALUES (1, true, '霜牙', '', 6, true, true, true, true, true, true)
ON CONFLICT (id) DO NOTHING;
