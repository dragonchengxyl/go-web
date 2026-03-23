ALTER TABLE hex_blitz_matches
    ADD COLUMN IF NOT EXISTS seed BIGINT NOT NULL DEFAULT 0;

CREATE TABLE hex_blitz_move_events (
    id            UUID PRIMARY KEY,
    match_id      UUID NOT NULL REFERENCES hex_blitz_matches(id) ON DELETE CASCADE,
    session_id    VARCHAR(64) NOT NULL,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    player_name   VARCHAR(64) NOT NULL,
    tile_id       VARCHAR(32) NOT NULL,
    move_index    INT NOT NULL,
    cleared_count INT NOT NULL DEFAULT 0,
    gained_score  INT NOT NULL DEFAULT 0,
    score_after   INT NOT NULL DEFAULT 0,
    combo_after   INT NOT NULL DEFAULT 0,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hex_blitz_move_events_match_id
    ON hex_blitz_move_events(match_id, move_index);

CREATE INDEX idx_hex_blitz_move_events_user_id
    ON hex_blitz_move_events(user_id);
