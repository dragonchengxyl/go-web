-- Migration 018: Drop game-related tables
-- Drop dependent tables first

DROP TABLE IF EXISTS download_logs CASCADE;
DROP TABLE IF EXISTS user_game_assets CASCADE;
DROP TABLE IF EXISTS game_releases CASCADE;
DROP TABLE IF EXISTS game_dlcs CASCADE;
DROP TABLE IF EXISTS game_branches CASCADE;
DROP TABLE IF EXISTS game_screenshots CASCADE;
DROP TABLE IF EXISTS games CASCADE;
