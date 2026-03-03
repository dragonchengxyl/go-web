-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    avatar_key VARCHAR(255),
    bio TEXT,
    location VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'player',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMP,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Add comments
COMMENT ON TABLE users IS 'User accounts';
COMMENT ON COLUMN users.id IS 'User UUID';
COMMENT ON COLUMN users.username IS 'Unique username';
COMMENT ON COLUMN users.email IS 'Unique email address';
COMMENT ON COLUMN users.password_hash IS 'Argon2id password hash';
COMMENT ON COLUMN users.avatar_key IS 'OSS key for avatar image';
COMMENT ON COLUMN users.role IS 'User role: super_admin, admin, moderator, creator, premium, player, guest';
COMMENT ON COLUMN users.status IS 'Account status: active, inactive, suspended, banned';
