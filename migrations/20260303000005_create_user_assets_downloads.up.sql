-- 用户游戏资产库表
CREATE TABLE user_game_assets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id     UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    asset_type  VARCHAR(50) NOT NULL,
    asset_id    UUID NOT NULL,
    obtained_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source      VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, asset_type, asset_id)
);

CREATE INDEX idx_user_game_assets_user_id ON user_game_assets(user_id);
CREATE INDEX idx_user_game_assets_game_id ON user_game_assets(game_id);
CREATE INDEX idx_user_game_assets_asset_type ON user_game_assets(asset_type);

-- 下载日志表
CREATE TABLE download_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    release_id  UUID NOT NULL REFERENCES game_releases(id),
    client_ip   INET NOT NULL,
    user_agent  TEXT,
    downloaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_download_logs_user_id ON download_logs(user_id);
CREATE INDEX idx_download_logs_release_id ON download_logs(release_id);
CREATE INDEX idx_download_logs_downloaded_at ON download_logs(downloaded_at);
