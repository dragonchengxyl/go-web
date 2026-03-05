package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/game"
)

// SearchService handles search operations
type SearchService struct {
	db *pgxpool.Pool
}

// NewSearchService creates a new search service
func NewSearchService(db *pgxpool.Pool) *SearchService {
	return &SearchService{db: db}
}

// SearchResult represents a unified search result
type SearchResult struct {
	Games  []*game.Game `json:"games"`
	Albums []Album      `json:"albums"`
	Tracks []Track      `json:"tracks"`
	Total  int          `json:"total"`
}

// Album represents an album search result
type Album struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	CoverURL    string    `json:"cover_url"`
	TrackCount  int       `json:"track_count"`
}

// Track represents a track search result
type Track struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Artist   string    `json:"artist"`
	AlbumID  uuid.UUID `json:"album_id"`
	Duration int       `json:"duration"`
}

// Search performs a full-text search across all content
func (s *SearchService) Search(ctx context.Context, query string, limit int) (*SearchResult, error) {
	result := &SearchResult{
		Games:  make([]*game.Game, 0),
		Albums: make([]Album, 0),
		Tracks: make([]Track, 0),
	}

	// Search games
	gamesQuery := `
		SELECT id, title, description, tags, created_at, updated_at,
		       ts_rank(search_vector, plainto_tsquery('simple', $1)) as rank
		FROM games
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		   OR title ILIKE '%' || $1 || '%'
		ORDER BY rank DESC, created_at DESC
		LIMIT $2
	`
	rows, err := s.db.Query(ctx, gamesQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var g game.Game
		var rank float64
		err := rows.Scan(
			&g.ID, &g.Title, &g.Description, &g.Tags,
			&g.CreatedAt, &g.UpdatedAt, &rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		result.Games = append(result.Games, &g)
	}

	// Search albums
	albumsQuery := `
		SELECT id, title, artist, cover_url,
		       (SELECT COUNT(*) FROM tracks WHERE album_id = albums.id) as track_count,
		       ts_rank(search_vector, plainto_tsquery('simple', $1)) as rank
		FROM albums
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		   OR title ILIKE '%' || $1 || '%'
		   OR artist ILIKE '%' || $1 || '%'
		ORDER BY rank DESC, created_at DESC
		LIMIT $2
	`
	rows, err = s.db.Query(ctx, albumsQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search albums: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var album Album
		var rank float64
		err := rows.Scan(&album.ID, &album.Title, &album.Artist, &album.CoverURL, &album.TrackCount, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}
		result.Albums = append(result.Albums, album)
	}

	// Search tracks
	tracksQuery := `
		SELECT id, title, artist, album_id, duration,
		       ts_rank(search_vector, plainto_tsquery('simple', $1)) as rank
		FROM tracks
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		   OR title ILIKE '%' || $1 || '%'
		   OR artist ILIKE '%' || $1 || '%'
		ORDER BY rank DESC, created_at DESC
		LIMIT $2
	`
	rows, err = s.db.Query(ctx, tracksQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var track Track
		var rank float64
		err := rows.Scan(&track.ID, &track.Title, &track.Artist, &track.AlbumID, &track.Duration, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}
		result.Tracks = append(result.Tracks, track)
	}

	result.Total = len(result.Games) + len(result.Albums) + len(result.Tracks)

	// Record search history
	go s.recordSearchHistory(context.Background(), query, result.Total)

	return result, nil
}

// SearchGames searches only games
func (s *SearchService) SearchGames(ctx context.Context, query string, limit, offset int) ([]*game.Game, int64, error) {
	gamesQuery := `
		SELECT id, title, description, tags, created_at, updated_at,
		       ts_rank(search_vector, plainto_tsquery('simple', $1)) as rank
		FROM games
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		   OR title ILIKE '%' || $1 || '%'
		ORDER BY rank DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.Query(ctx, gamesQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search games: %w", err)
	}
	defer rows.Close()

	games := make([]*game.Game, 0)
	for rows.Next() {
		var g game.Game
		var rank float64
		err := rows.Scan(
			&g.ID, &g.Title, &g.Description, &g.Tags,
			&g.CreatedAt, &g.UpdatedAt, &rank,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &g)
	}

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM games
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		   OR title ILIKE '%' || $1 || '%'
	`
	err = s.db.QueryRow(ctx, countQuery, query).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count games: %w", err)
	}

	return games, total, nil
}

// GetSearchSuggestions returns search suggestions based on query
func (s *SearchService) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	suggestQuery := `
		SELECT DISTINCT title
		FROM (
			SELECT title FROM games WHERE title ILIKE $1 || '%'
			UNION
			SELECT title FROM albums WHERE title ILIKE $1 || '%'
			UNION
			SELECT title FROM tracks WHERE title ILIKE $1 || '%'
		) AS suggestions
		ORDER BY title
		LIMIT $2
	`
	rows, err := s.db.Query(ctx, suggestQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}
	defer rows.Close()

	suggestions := make([]string, 0)
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, fmt.Errorf("failed to scan suggestion: %w", err)
		}
		suggestions = append(suggestions, title)
	}

	return suggestions, nil
}

// GetPopularSearches returns popular search queries
func (s *SearchService) GetPopularSearches(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT query, COUNT(*) as count
		FROM search_history
		WHERE created_at >= NOW() - INTERVAL '7 days'
		GROUP BY query
		ORDER BY count DESC
		LIMIT $1
	`
	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular searches: %w", err)
	}
	defer rows.Close()

	searches := make([]string, 0)
	for rows.Next() {
		var search string
		var count int
		if err := rows.Scan(&search, &count); err != nil {
			return nil, fmt.Errorf("failed to scan popular search: %w", err)
		}
		searches = append(searches, search)
	}

	return searches, nil
}

// recordSearchHistory records a search query
func (s *SearchService) recordSearchHistory(ctx context.Context, query string, resultCount int) {
	insertQuery := `
		INSERT INTO search_history (query, result_count)
		VALUES ($1, $2)
	`
	_, _ = s.db.Exec(ctx, insertQuery, query, resultCount)
}

// GetRecommendedGames returns recommended games based on user behavior
func (s *SearchService) GetRecommendedGames(ctx context.Context, userID *uuid.UUID, limit int) ([]*game.Game, error) {
	var query string
	var args []any

	if userID != nil {
		// Personalized recommendations based on user's game views
		query = `
			SELECT DISTINCT g.id, g.title, g.description, g.tags,
			       g.created_at, g.updated_at
			FROM games g
			WHERE g.tags && (
				SELECT array_agg(DISTINCT tag)
				FROM games g2
				JOIN analytics_events ae ON ae.properties->>'game_id' = g2.id::text
				WHERE ae.user_id = $1
				AND ae.event_type = 'game_view'
				CROSS JOIN LATERAL unnest(g2.tags) AS tag
			)
			AND g.id NOT IN (
				SELECT (properties->>'game_id')::uuid
				FROM analytics_events
				WHERE user_id = $1
				AND event_type = 'game_view'
			)
			ORDER BY g.created_at DESC
			LIMIT $2
		`
		args = []any{userID, limit}
	} else {
		// Popular games for anonymous users
		query = `
			SELECT g.id, g.title, g.description, g.tags,
			       g.created_at, g.updated_at
			FROM games g
			LEFT JOIN analytics_events ae ON ae.properties->>'game_id' = g.id::text
			WHERE ae.event_type = 'game_view'
			AND ae.created_at >= NOW() - INTERVAL '7 days'
			GROUP BY g.id
			ORDER BY COUNT(*) DESC
			LIMIT $1
		`
		args = []any{limit}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommended games: %w", err)
	}
	defer rows.Close()

	games := make([]*game.Game, 0)
	for rows.Next() {
		var g game.Game
		err := rows.Scan(
			&g.ID, &g.Title, &g.Description, &g.Tags,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &g)
	}

	return games, nil
}
