-- 游戏基本信息表
CREATE TABLE games (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(100) NOT NULL UNIQUE,
    title           VARCHAR(255) NOT NULL,
    subtitle        VARCHAR(255),
    description     TEXT,
    cover_key       VARCHAR(512),
    banner_key      VARCHAR(512),
    trailer_url     VARCHAR(512),
    genre           VARCHAR(50)[],
    tags            VARCHAR(50)[],
    engine          VARCHAR(50) NOT NULL DEFAULT 'gvn',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    release_date    DATE,
    developer_id    UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_games_slug ON games(slug);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_games_developer_id ON games(developer_id);
CREATE INDEX idx_games_release_date ON games(release_date);

-- 游戏截图表
CREATE TABLE game_screenshots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id     UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    oss_key     VARCHAR(512) NOT NULL,
    caption     VARCHAR(255),
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_screenshots_game_id ON game_screenshots(game_id);

-- 游戏发行分支表
CREATE TABLE game_branches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id     UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    name        VARCHAR(50) NOT NULL,
    description TEXT,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(game_id, name)
);

CREATE INDEX idx_game_branches_game_id ON game_branches(game_id);
