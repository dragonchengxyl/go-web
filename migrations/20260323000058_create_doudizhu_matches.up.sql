CREATE TABLE doudizhu_matches (
    id             UUID PRIMARY KEY,
    room_id        UUID NOT NULL,
    room_code      TEXT NOT NULL,
    room_title     TEXT NOT NULL,
    match_mode     TEXT NOT NULL,
    started_at     TIMESTAMPTZ NOT NULL,
    finished_at    TIMESTAMPTZ NOT NULL,
    landlord_seat  SMALLINT NOT NULL,
    winner_side    TEXT NOT NULL,
    multiplier     INTEGER NOT NULL DEFAULT 1,
    bomb_count     INTEGER NOT NULL DEFAULT 0,
    spring         BOOLEAN NOT NULL DEFAULT FALSE,
    anti_spring    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_doudizhu_matches_finished_at
    ON doudizhu_matches(finished_at DESC);

CREATE INDEX idx_doudizhu_matches_room_id
    ON doudizhu_matches(room_id);

CREATE INDEX idx_doudizhu_matches_mode
    ON doudizhu_matches(match_mode);

CREATE TABLE doudizhu_match_players (
    id           UUID PRIMARY KEY,
    match_id     UUID NOT NULL REFERENCES doudizhu_matches(id) ON DELETE CASCADE,
    session_id   TEXT NOT NULL,
    user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    is_bot       BOOLEAN NOT NULL DEFAULT FALSE,
    bot_level    TEXT,
    seat         SMALLINT NOT NULL,
    player_name  TEXT NOT NULL,
    role         TEXT NOT NULL,
    bid_score    INTEGER NOT NULL DEFAULT 0,
    cards_left   INTEGER NOT NULL DEFAULT 0,
    is_winner    BOOLEAN NOT NULL DEFAULT FALSE,
    score_delta  INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_doudizhu_match_players_match_id
    ON doudizhu_match_players(match_id);

CREATE INDEX idx_doudizhu_match_players_user_id
    ON doudizhu_match_players(user_id);

CREATE INDEX idx_doudizhu_match_players_score_delta
    ON doudizhu_match_players(score_delta DESC, created_at DESC);
