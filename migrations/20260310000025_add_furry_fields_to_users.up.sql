-- Migration 025: Add Furry community fields to users table

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS furry_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS species VARCHAR(100);
