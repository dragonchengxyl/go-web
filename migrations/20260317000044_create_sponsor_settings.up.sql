CREATE TABLE IF NOT EXISTS sponsor_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    monthly_goal DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_raised DOUBLE PRECISION NOT NULL DEFAULT 0,
    alipay_qr_url TEXT NOT NULL DEFAULT '',
    wechat_qr_url TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);
