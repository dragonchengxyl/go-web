-- 专辑表
CREATE TABLE albums (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id         UUID REFERENCES games(id) ON DELETE SET NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    title           VARCHAR(255) NOT NULL,
    subtitle        VARCHAR(255),
    description     TEXT,
    cover_key       VARCHAR(512),
    artist          VARCHAR(255),
    composer        VARCHAR(255),
    arranger        VARCHAR(255),
    lyricist        VARCHAR(255),
    total_tracks    INT,
    duration_sec    INT,
    release_date    DATE,
    album_type      VARCHAR(50) DEFAULT 'ost',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_albums_game_id ON albums(game_id);
CREATE INDEX idx_albums_slug ON albums(slug);
CREATE INDEX idx_albums_album_type ON albums(album_type);

-- 音轨表
CREATE TABLE tracks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    album_id        UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    track_number    INT NOT NULL,
    disc_number     INT NOT NULL DEFAULT 1,
    title           VARCHAR(255) NOT NULL,
    artist          VARCHAR(255),
    duration_sec    INT,
    stream_key      VARCHAR(512),
    stream_size     BIGINT,
    hifi_key        VARCHAR(512),
    hifi_format     VARCHAR(10),
    hifi_bitdepth   SMALLINT,
    hifi_samplerate INT,
    hifi_size       BIGINT,
    lrc_key         VARCHAR(512),
    play_count      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(album_id, disc_number, track_number)
);

CREATE INDEX idx_tracks_album_id ON tracks(album_id);
CREATE INDEX idx_tracks_play_count ON tracks(play_count DESC);
