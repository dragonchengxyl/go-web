-- Add full-text search support for games
ALTER TABLE games ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;

-- Create GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_games_search ON games USING GIN(search_vector);

-- Create function to update search vector
CREATE OR REPLACE FUNCTION games_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(array_to_string(NEW.tags, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS update_games_search_vector ON games;
CREATE TRIGGER update_games_search_vector
    BEFORE INSERT OR UPDATE ON games
    FOR EACH ROW
    EXECUTE FUNCTION games_search_vector_update();

-- Update existing records
UPDATE games SET search_vector =
    setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(description, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(array_to_string(tags, ' '), '')), 'C');

-- Add full-text search for albums
ALTER TABLE albums ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;
CREATE INDEX IF NOT EXISTS idx_albums_search ON albums USING GIN(search_vector);

CREATE OR REPLACE FUNCTION albums_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.artist, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_albums_search_vector ON albums;
CREATE TRIGGER update_albums_search_vector
    BEFORE INSERT OR UPDATE ON albums
    FOR EACH ROW
    EXECUTE FUNCTION albums_search_vector_update();

UPDATE albums SET search_vector =
    setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(artist, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(description, '')), 'C');

-- Add full-text search for tracks
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;
CREATE INDEX IF NOT EXISTS idx_tracks_search ON tracks USING GIN(search_vector);

CREATE OR REPLACE FUNCTION tracks_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.artist, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_tracks_search_vector ON tracks;
CREATE TRIGGER update_tracks_search_vector
    BEFORE INSERT OR UPDATE ON tracks
    FOR EACH ROW
    EXECUTE FUNCTION tracks_search_vector_update();

UPDATE tracks SET search_vector =
    setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(artist, '')), 'B');

-- Create search history table
CREATE TABLE IF NOT EXISTS search_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    result_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_history_user_id ON search_history(user_id);
CREATE INDEX idx_search_history_created_at ON search_history(created_at DESC);

-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram indexes for fuzzy matching
CREATE INDEX IF NOT EXISTS idx_games_title_trgm ON games USING GIN(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_albums_title_trgm ON albums USING GIN(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tracks_title_trgm ON tracks USING GIN(title gin_trgm_ops);
