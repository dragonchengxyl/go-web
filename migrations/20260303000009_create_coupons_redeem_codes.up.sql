-- 优惠券表
CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    coupon_type VARCHAR(20) NOT NULL CHECK (coupon_type IN ('percentage', 'fixed', 'amount')),
    discount_value NUMERIC(10,2) NOT NULL,
    min_purchase_cents INT CHECK (min_purchase_cents >= 0),
    max_discount_cents INT CHECK (max_discount_cents >= 0),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    max_uses INT,
    used_count INT DEFAULT 0,
    max_uses_per_user INT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_coupon_period CHECK (end_at > start_at)
);

CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_coupons_active ON coupons(is_active);
CREATE INDEX idx_coupons_period ON coupons(start_at, end_at);

-- 兑换码表
CREATE TABLE redeem_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    description TEXT,
    used_by UUID REFERENCES users(id) ON DELETE SET NULL,
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_redeem_codes_code ON redeem_codes(code);
CREATE INDEX idx_redeem_codes_product ON redeem_codes(product_id);
CREATE INDEX idx_redeem_codes_used_by ON redeem_codes(used_by);
CREATE INDEX idx_redeem_codes_active ON redeem_codes(is_active);
