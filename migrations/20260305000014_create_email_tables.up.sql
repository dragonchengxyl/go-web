-- Create emails table for email queue
CREATE TABLE IF NOT EXISTS emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_address VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    email_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    error TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_emails_status ON emails(status);
CREATE INDEX idx_emails_created_at ON emails(created_at DESC);
CREATE INDEX idx_emails_type ON emails(email_type);

-- Create email_subscriptions table
CREATE TABLE IF NOT EXISTS email_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token VARCHAR(100) NOT NULL,
    game_releases BOOLEAN NOT NULL DEFAULT TRUE,
    updates BOOLEAN NOT NULL DEFAULT TRUE,
    promotions BOOLEAN NOT NULL DEFAULT FALSE,
    unsubscribe_token VARCHAR(100) NOT NULL,
    subscribed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    unsubscribed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_email_subscriptions_email ON email_subscriptions(email);
CREATE INDEX idx_email_subscriptions_verified ON email_subscriptions(is_verified);
CREATE INDEX idx_email_subscriptions_unsubscribed ON email_subscriptions(unsubscribed_at);

COMMENT ON TABLE emails IS '邮件队列表';
COMMENT ON TABLE email_subscriptions IS '邮件订阅表';
