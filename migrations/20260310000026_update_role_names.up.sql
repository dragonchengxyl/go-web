-- Migration 026: Add supporter/member role values
-- Roles are stored as strings so no enum changes needed.
-- Update any existing 'premium' roles to 'supporter' and 'player' roles to 'member'.

UPDATE users SET role = 'supporter' WHERE role = 'premium';
UPDATE users SET role = 'member' WHERE role = 'player';
