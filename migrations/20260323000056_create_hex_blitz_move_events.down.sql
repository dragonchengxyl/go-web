DROP TABLE IF EXISTS hex_blitz_move_events;

ALTER TABLE hex_blitz_matches
    DROP COLUMN IF EXISTS seed;
