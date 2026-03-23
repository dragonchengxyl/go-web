CREATE TABLE hex_blitz_matches (
    id            UUID PRIMARY KEY,
    room_id       UUID NOT NULL,
    room_code     VARCHAR(16) NOT NULL,
    room_title    VARCHAR(100) NOT NULL,
    game_slug     VARCHAR(50) NOT NULL DEFAULT 'hex-blitz',
    started_at    TIMESTAMPTZ NOT NULL,
    finished_at   TIMESTAMPTZ NOT NULL,
    duration_sec  INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hex_blitz_matches_finished_at
    ON hex_blitz_matches(finished_at DESC);

CREATE INDEX idx_hex_blitz_matches_room_id
    ON hex_blitz_matches(room_id);

CREATE TABLE hex_blitz_match_results (
    id           UUID PRIMARY KEY,
    match_id     UUID NOT NULL REFERENCES hex_blitz_matches(id) ON DELETE CASCADE,
    room_id      UUID NOT NULL,
    room_code    VARCHAR(16) NOT NULL,
    room_title   VARCHAR(100) NOT NULL,
    user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    player_name  VARCHAR(64) NOT NULL,
    score        INT NOT NULL DEFAULT 0,
    rank         INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hex_blitz_match_results_match_id
    ON hex_blitz_match_results(match_id);

CREATE INDEX idx_hex_blitz_match_results_user_id
    ON hex_blitz_match_results(user_id);

CREATE INDEX idx_hex_blitz_match_results_score
    ON hex_blitz_match_results(score DESC, created_at DESC);
