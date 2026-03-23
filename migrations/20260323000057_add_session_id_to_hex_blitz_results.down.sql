DROP INDEX IF EXISTS idx_hex_blitz_match_results_session_id;

ALTER TABLE hex_blitz_match_results
    DROP COLUMN IF EXISTS session_id;
