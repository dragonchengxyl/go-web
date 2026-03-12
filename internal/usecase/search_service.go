package usecase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SearchService only retains popular-search analytics for the current community product.
type SearchService struct {
	db *pgxpool.Pool
}

func NewSearchService(db *pgxpool.Pool) *SearchService {
	return &SearchService{db: db}
}

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
