CREATE TABLE doudizhu_action_events (
    id                    UUID PRIMARY KEY,
    match_id              UUID NOT NULL REFERENCES doudizhu_matches(id) ON DELETE CASCADE,
    turn_no               INTEGER NOT NULL,
    action_index          INTEGER NOT NULL,
    session_id            TEXT NOT NULL,
    user_id               UUID REFERENCES users(id) ON DELETE SET NULL,
    player_name           TEXT NOT NULL,
    seat                  SMALLINT NOT NULL,
    action_type           TEXT NOT NULL,
    cards_json            JSONB,
    combo_type            TEXT,
    combo_main_rank       INTEGER,
    combo_sequence_length INTEGER,
    combo_total_cards     INTEGER,
    multiplier_after      INTEGER NOT NULL DEFAULT 1,
    occurred_at           TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_doudizhu_action_events_match_id
    ON doudizhu_action_events(match_id, action_index);

CREATE INDEX idx_doudizhu_action_events_user_id
    ON doudizhu_action_events(user_id);
