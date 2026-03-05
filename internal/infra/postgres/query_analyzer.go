package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// QueryAnalyzer analyzes database queries for performance issues
type QueryAnalyzer struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewQueryAnalyzer creates a new query analyzer
func NewQueryAnalyzer(pool *pgxpool.Pool, logger *zap.Logger) *QueryAnalyzer {
	return &QueryAnalyzer{
		pool:   pool,
		logger: logger,
	}
}

// SlowQuery represents a slow query
type SlowQuery struct {
	Query         string
	Calls         int64
	TotalTime     float64
	MeanTime      float64
	MinTime       float64
	MaxTime       float64
	StddevTime    float64
	Rows          int64
}

// GetSlowQueries retrieves slow queries from pg_stat_statements
func (qa *QueryAnalyzer) GetSlowQueries(ctx context.Context, limit int) ([]SlowQuery, error) {
	query := `
		SELECT
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			min_exec_time,
			max_exec_time,
			stddev_exec_time,
			rows
		FROM pg_stat_statements
		WHERE query NOT LIKE '%pg_stat_statements%'
		ORDER BY mean_exec_time DESC
		LIMIT $1
	`

	rows, err := qa.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get slow queries: %w", err)
	}
	defer rows.Close()

	var slowQueries []SlowQuery
	for rows.Next() {
		var sq SlowQuery
		err := rows.Scan(
			&sq.Query,
			&sq.Calls,
			&sq.TotalTime,
			&sq.MeanTime,
			&sq.MinTime,
			&sq.MaxTime,
			&sq.StddevTime,
			&sq.Rows,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slow query: %w", err)
		}
		slowQueries = append(slowQueries, sq)
	}

	return slowQueries, nil
}

// ExplainQuery analyzes a query using EXPLAIN ANALYZE
func (qa *QueryAnalyzer) ExplainQuery(ctx context.Context, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)

	var result string
	err := qa.pool.QueryRow(ctx, explainQuery).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	return result, nil
}

// IndexUsage represents index usage statistics
type IndexUsage struct {
	SchemaName string
	TableName  string
	IndexName  string
	IndexScans int64
	TupleReads int64
	TupleFetch int64
	IndexSize  int64
}

// GetUnusedIndexes retrieves indexes that are not being used
func (qa *QueryAnalyzer) GetUnusedIndexes(ctx context.Context) ([]IndexUsage, error) {
	query := `
		SELECT
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch,
			pg_relation_size(indexrelid) as index_size
		FROM pg_stat_user_indexes
		WHERE idx_scan = 0
		AND indexrelname NOT LIKE '%_pkey'
		ORDER BY pg_relation_size(indexrelid) DESC
	`

	rows, err := qa.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unused indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexUsage
	for rows.Next() {
		var idx IndexUsage
		err := rows.Scan(
			&idx.SchemaName,
			&idx.TableName,
			&idx.IndexName,
			&idx.IndexScans,
			&idx.TupleReads,
			&idx.TupleFetch,
			&idx.IndexSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index usage: %w", err)
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// TableStats represents table statistics
type TableStats struct {
	SchemaName    string
	TableName     string
	RowCount      int64
	TotalSize     int64
	IndexSize     int64
	SeqScans      int64
	SeqTupleReads int64
	IdxScans      int64
	IdxTupleReads int64
	LastVacuum    *time.Time
	LastAnalyze   *time.Time
}

// GetTableStats retrieves table statistics
func (qa *QueryAnalyzer) GetTableStats(ctx context.Context) ([]TableStats, error) {
	query := `
		SELECT
			schemaname,
			tablename,
			n_live_tup as row_count,
			pg_total_relation_size(schemaname||'.'||tablename) as total_size,
			pg_indexes_size(schemaname||'.'||tablename) as index_size,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			last_vacuum,
			last_analyze
		FROM pg_stat_user_tables
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`

	rows, err := qa.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table stats: %w", err)
	}
	defer rows.Close()

	var stats []TableStats
	for rows.Next() {
		var ts TableStats
		err := rows.Scan(
			&ts.SchemaName,
			&ts.TableName,
			&ts.RowCount,
			&ts.TotalSize,
			&ts.IndexSize,
			&ts.SeqScans,
			&ts.SeqTupleReads,
			&ts.IdxScans,
			&ts.IdxTupleReads,
			&ts.LastVacuum,
			&ts.LastAnalyze,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table stats: %w", err)
		}
		stats = append(stats, ts)
	}

	return stats, nil
}

// MissingIndex represents a potential missing index
type MissingIndex struct {
	TableName      string
	SeqScans       int64
	SeqTupleReads  int64
	IdxScans       int64
	Recommendation string
}

// GetMissingIndexes identifies tables that might benefit from indexes
func (qa *QueryAnalyzer) GetMissingIndexes(ctx context.Context) ([]MissingIndex, error) {
	query := `
		SELECT
			tablename,
			seq_scan,
			seq_tup_read,
			idx_scan,
			CASE
				WHEN seq_scan > 0 AND idx_scan = 0 THEN 'Consider adding an index - only sequential scans'
				WHEN seq_scan > idx_scan * 10 THEN 'High sequential scan ratio - review query patterns'
				WHEN seq_tup_read > 100000 AND idx_scan = 0 THEN 'Large table with no index usage - critical'
				ELSE 'Monitor query patterns'
			END as recommendation
		FROM pg_stat_user_tables
		WHERE seq_scan > 100
		AND (idx_scan = 0 OR seq_scan > idx_scan * 5)
		ORDER BY seq_tup_read DESC
		LIMIT 20
	`

	rows, err := qa.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing indexes: %w", err)
	}
	defer rows.Close()

	var missing []MissingIndex
	for rows.Next() {
		var mi MissingIndex
		err := rows.Scan(
			&mi.TableName,
			&mi.SeqScans,
			&mi.SeqTupleReads,
			&mi.IdxScans,
			&mi.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan missing index: %w", err)
		}
		missing = append(missing, mi)
	}

	return missing, nil
}

// VacuumAnalyze runs VACUUM ANALYZE on all tables
func (qa *QueryAnalyzer) VacuumAnalyze(ctx context.Context) error {
	query := `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`

	rows, err := qa.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}

	// Run VACUUM ANALYZE on each table
	for _, table := range tables {
		qa.logger.Info("Running VACUUM ANALYZE", zap.String("table", table))
		_, err := qa.pool.Exec(ctx, fmt.Sprintf("VACUUM ANALYZE %s", table))
		if err != nil {
			qa.logger.Error("Failed to VACUUM ANALYZE table",
				zap.String("table", table),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// ResetStats resets pg_stat_statements statistics
func (qa *QueryAnalyzer) ResetStats(ctx context.Context) error {
	_, err := qa.pool.Exec(ctx, "SELECT pg_stat_statements_reset()")
	if err != nil {
		return fmt.Errorf("failed to reset stats: %w", err)
	}
	return nil
}

// GetConnectionStats retrieves connection statistics
func (qa *QueryAnalyzer) GetConnectionStats(ctx context.Context) (map[string]any, error) {
	query := `
		SELECT
			count(*) as total,
			count(*) FILTER (WHERE state = 'active') as active,
			count(*) FILTER (WHERE state = 'idle') as idle,
			count(*) FILTER (WHERE state = 'idle in transaction') as idle_in_transaction,
			count(*) FILTER (WHERE wait_event_type IS NOT NULL) as waiting
		FROM pg_stat_activity
		WHERE pid != pg_backend_pid()
	`

	var total, active, idle, idleInTransaction, waiting int64
	err := qa.pool.QueryRow(ctx, query).Scan(&total, &active, &idle, &idleInTransaction, &waiting)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection stats: %w", err)
	}

	return map[string]any{
		"total":               total,
		"active":              active,
		"idle":                idle,
		"idle_in_transaction": idleInTransaction,
		"waiting":             waiting,
	}, nil
}

// FormatBytes formats bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// NormalizeQuery normalizes a SQL query by removing literals
func NormalizeQuery(query string) string {
	// Simple normalization - replace numbers and strings with placeholders
	normalized := strings.ReplaceAll(query, "'", "")
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}
