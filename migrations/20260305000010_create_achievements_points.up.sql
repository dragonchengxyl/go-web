-- Achievements definition table
CREATE TABLE achievements (
    id              SERIAL PRIMARY KEY,
    slug            VARCHAR(100)  NOT NULL UNIQUE,
    name            VARCHAR(100)  NOT NULL,
    description     TEXT,
    icon_key        VARCHAR(512),
    rarity          VARCHAR(20)   NOT NULL DEFAULT 'common', -- common, rare, epic, legendary
    points          INT           NOT NULL DEFAULT 0,
    condition_type  VARCHAR(50)   NOT NULL DEFAULT 'manual', -- manual, game_download, comment_count, login_streak, etc.
    condition_value JSONB,
    is_secret       BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- User achievements (unlocked)
CREATE TABLE user_achievements (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id  INT         NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    obtained_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, achievement_id)
);
CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);

-- Point transactions log
CREATE TABLE point_transactions (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      INT         NOT NULL,               -- positive = earn, negative = spend
    source      VARCHAR(50) NOT NULL,               -- 'register', 'daily_checkin', 'comment', 'purchase', 'achievement', etc.
    ref_id      VARCHAR(100),                       -- optional reference (e.g. order_id, comment_id)
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_point_transactions_user_id ON point_transactions(user_id);

-- User points balance (cached sum)
CREATE TABLE user_points (
    user_id     UUID    PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance     INT     NOT NULL DEFAULT 0,
    total_earned INT    NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed preset achievements
INSERT INTO achievements (slug, name, description, rarity, points, condition_type, condition_value, is_secret) VALUES
    ('first_download',     '初次探索',   '首次下载任意游戏',                   'common',    10,  'game_download',   '{"count":1}',    FALSE),
    ('early_bird',         '早鸟玩家',   '在游戏正式发布前下载 Demo',           'rare',      50,  'manual',          NULL,             FALSE),
    ('chatterbox',         '话痨',       '发布 100 条评论',                    'rare',      30,  'comment_count',   '{"count":100}',  FALSE),
    ('community_star',     '社区明星',   '评论累计获得 1000 个赞',             'epic',      100, 'like_count',      '{"count":1000}', FALSE),
    ('loyal_fan_30',       '常客',       '连续登录 30 天',                     'epic',      80,  'login_streak',    '{"days":30}',    FALSE),
    ('veteran',            '老玩家',     '注册满 1 周年',                      'legendary', 200, 'account_age',     '{"days":365}',   FALSE),
    ('music_lover',        '音乐鉴赏家', '累计试听 OST 超过 100 首',           'common',    20,  'track_stream',    '{"count":100}',  FALSE),
    ('first_purchase',     '首次支持',   '完成第一笔购买',                     'common',    30,  'purchase_count',  '{"count":1}',    FALSE),
    ('founder',            '创始支持者', '注册于工作室成立后 30 天内',          'legendary', 500, 'manual',          NULL,             TRUE);
