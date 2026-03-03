-- 游戏版本发布表
CREATE TABLE game_releases (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id       UUID NOT NULL REFERENCES game_branches(id) ON DELETE CASCADE,
    version         VARCHAR(20) NOT NULL,
    title           VARCHAR(255),
    changelog       TEXT,
    oss_key         VARCHAR(512),
    manifest_key    VARCHAR(512),
    file_size       BIGINT,
    checksum_sha256 VARCHAR(64),
    min_os_version  VARCHAR(20),
    is_published    BOOLEAN NOT NULL DEFAULT FALSE,
    published_at    TIMESTAMPTZ,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(branch_id, version)
);

CREATE INDEX idx_game_releases_branch_id ON game_releases(branch_id);
CREATE INDEX idx_game_releases_is_published ON game_releases(is_published);
CREATE INDEX idx_game_releases_created_at ON game_releases(created_at);

-- DLC 表
CREATE TABLE dlc (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id         UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    slug            VARCHAR(100) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    cover_key       VARCHAR(512),
    dlc_type        VARCHAR(50),
    price_cents     INT,
    currency        CHAR(3) NOT NULL DEFAULT 'CNY',
    is_free         BOOLEAN NOT NULL DEFAULT FALSE,
    release_date    DATE,
    oss_key         VARCHAR(512),
    file_size       BIGINT,
    checksum_sha256 VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(game_id, slug)
);

CREATE INDEX idx_dlc_game_id ON dlc(game_id);
CREATE INDEX idx_dlc_slug ON dlc(slug);
