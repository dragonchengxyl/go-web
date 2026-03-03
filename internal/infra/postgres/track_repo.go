package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/music"
)

type TrackRepository struct {
	pool *pgxpool.Pool
}

func NewTrackRepository(pool *pgxpool.Pool) *TrackRepository {
	return &TrackRepository{pool: pool}
}

const createTrackSQL = `
	INSERT INTO tracks (id, album_id, track_number, disc_number, title, artist, duration_sec,
	                    stream_key, stream_size, hifi_key, hifi_format, hifi_bitdepth,
	                    hifi_samplerate, hifi_size, lrc_key, play_count, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
`

func (r *TrackRepository) Create(ctx context.Context, t *music.Track) error {
	_, err := r.pool.Exec(ctx, createTrackSQL,
		t.ID, t.AlbumID, t.TrackNumber, t.DiscNumber, t.Title, t.Artist, t.DurationSec,
		t.StreamKey, t.StreamSize, t.HiFiKey, t.HiFiFormat, t.HiFiBitDepth,
		t.HiFiSampleRate, t.HiFiSize, t.LRCKey, t.PlayCount, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create track: %w", err)
	}
	return nil
}

const getTrackByIDSQL = `
	SELECT id, album_id, track_number, disc_number, title, artist, duration_sec,
	       stream_key, stream_size, hifi_key, hifi_format, hifi_bitdepth,
	       hifi_samplerate, hifi_size, lrc_key, play_count, created_at, updated_at
	FROM tracks WHERE id = $1
`

func (r *TrackRepository) GetByID(ctx context.Context, id uuid.UUID) (*music.Track, error) {
	var t music.Track
	err := r.pool.QueryRow(ctx, getTrackByIDSQL, id).Scan(
		&t.ID, &t.AlbumID, &t.TrackNumber, &t.DiscNumber, &t.Title, &t.Artist, &t.DurationSec,
		&t.StreamKey, &t.StreamSize, &t.HiFiKey, &t.HiFiFormat, &t.HiFiBitDepth,
		&t.HiFiSampleRate, &t.HiFiSize, &t.LRCKey, &t.PlayCount, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, music.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get track: %w", err)
	}
	return &t, nil
}

const updateTrackSQL = `
	UPDATE tracks
	SET album_id = $2, track_number = $3, disc_number = $4, title = $5, artist = $6,
	    duration_sec = $7, stream_key = $8, stream_size = $9, hifi_key = $10,
	    hifi_format = $11, hifi_bitdepth = $12, hifi_samplerate = $13, hifi_size = $14,
	    lrc_key = $15, updated_at = $16
	WHERE id = $1
`

func (r *TrackRepository) Update(ctx context.Context, t *music.Track) error {
	_, err := r.pool.Exec(ctx, updateTrackSQL,
		t.ID, t.AlbumID, t.TrackNumber, t.DiscNumber, t.Title, t.Artist,
		t.DurationSec, t.StreamKey, t.StreamSize, t.HiFiKey,
		t.HiFiFormat, t.HiFiBitDepth, t.HiFiSampleRate, t.HiFiSize,
		t.LRCKey, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}
	return nil
}

const deleteTrackSQL = `DELETE FROM tracks WHERE id = $1`

func (r *TrackRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteTrackSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}

	if result.RowsAffected() == 0 {
		return music.ErrNotFound
	}

	return nil
}

const listTracksByAlbumIDSQL = `
	SELECT id, album_id, track_number, disc_number, title, artist, duration_sec,
	       stream_key, stream_size, hifi_key, hifi_format, hifi_bitdepth,
	       hifi_samplerate, hifi_size, lrc_key, play_count, created_at, updated_at
	FROM tracks
	WHERE album_id = $1
	ORDER BY disc_number ASC, track_number ASC
`

func (r *TrackRepository) ListByAlbumID(ctx context.Context, albumID uuid.UUID) ([]*music.Track, error) {
	rows, err := r.pool.Query(ctx, listTracksByAlbumIDSQL, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tracks: %w", err)
	}
	defer rows.Close()

	tracks := make([]*music.Track, 0)
	for rows.Next() {
		var t music.Track
		err := rows.Scan(
			&t.ID, &t.AlbumID, &t.TrackNumber, &t.DiscNumber, &t.Title, &t.Artist, &t.DurationSec,
			&t.StreamKey, &t.StreamSize, &t.HiFiKey, &t.HiFiFormat, &t.HiFiBitDepth,
			&t.HiFiSampleRate, &t.HiFiSize, &t.LRCKey, &t.PlayCount, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}
		tracks = append(tracks, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tracks, nil
}

const incrementPlayCountSQL = `
	UPDATE tracks SET play_count = play_count + 1 WHERE id = $1
`

func (r *TrackRepository) IncrementPlayCount(ctx context.Context, trackID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, incrementPlayCountSQL, trackID)
	if err != nil {
		return fmt.Errorf("failed to increment play count: %w", err)
	}
	return nil
}
