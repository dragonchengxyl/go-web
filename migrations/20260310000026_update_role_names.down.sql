-- Rollback role rename
UPDATE users SET role = 'premium' WHERE role = 'supporter';
UPDATE users SET role = 'player' WHERE role = 'member';
