ALTER TABLE hex_blitz_match_results
    ADD COLUMN IF NOT EXISTS session_id VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_hex_blitz_match_results_session_id
    ON hex_blitz_match_results(session_id);
