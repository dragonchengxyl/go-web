CREATE TABLE IF NOT EXISTS refunds (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID        NOT NULL REFERENCES orders(id),
    payment_id   UUID        NOT NULL REFERENCES payments(id),
    amount_cents BIGINT      NOT NULL,
    reason       TEXT,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | succeeded | failed
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refunds_order_id   ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_payment_id ON refunds(payment_id);
