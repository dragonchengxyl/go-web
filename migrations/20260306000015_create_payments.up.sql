CREATE TABLE IF NOT EXISTS payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID        NOT NULL REFERENCES orders(id),
    gateway     VARCHAR(20) NOT NULL,          -- 'alipay' | 'wechat' | 'mock'
    trade_no    VARCHAR(128),                  -- gateway-assigned trade number
    amount_cents BIGINT     NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | paid | closed | refunded
    raw_response TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_trade_no  ON payments(trade_no);
