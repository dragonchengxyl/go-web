-- Add website field to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS website VARCHAR(255);

-- Add comment for website field
COMMENT ON COLUMN users.website IS 'User website URL';
