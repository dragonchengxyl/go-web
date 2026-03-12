-- Create analytics_events table for event tracking
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(100) NOT NULL,
    properties JSONB,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    referrer TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_analytics_events_event_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_events_session_id ON analytics_events(session_id);
CREATE INDEX idx_analytics_events_created_at ON analytics_events(created_at DESC);
CREATE INDEX idx_analytics_events_properties ON analytics_events USING GIN(properties);

-- Create composite indexes for common queries
CREATE INDEX idx_analytics_events_type_date ON analytics_events(event_type, created_at DESC);
CREATE INDEX idx_analytics_events_user_date ON analytics_events(user_id, created_at DESC);

-- Create daily metrics materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_metrics AS
SELECT
    DATE(created_at) as date,
    COUNT(DISTINCT user_id) FILTER (WHERE event_type = 'user_login') as dau,
    COUNT(DISTINCT user_id) FILTER (WHERE event_type = 'user_register') as new_users,
    COUNT(*) FILTER (WHERE event_type = 'game_view') as game_views,
    COUNT(*) FILTER (WHERE event_type = 'game_download') as game_downloads,
    COUNT(*) FILTER (WHERE event_type = 'purchase_complete') as purchases
FROM analytics_events
GROUP BY DATE(created_at)
ORDER BY date DESC;

CREATE UNIQUE INDEX idx_daily_metrics_date ON daily_metrics(date);

COMMENT ON TABLE analytics_events IS '用户行为事件追踪表';
COMMENT ON COLUMN analytics_events.event_type IS '事件类型';
COMMENT ON COLUMN analytics_events.properties IS '事件属性（JSON格式）';
COMMENT ON COLUMN analytics_events.session_id IS '会话ID';
