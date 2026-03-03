-- Remove website field from users table
ALTER TABLE users DROP COLUMN IF EXISTS website;
