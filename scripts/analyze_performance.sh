#!/bin/bash

# Database Performance Analysis Script
# This script analyzes PostgreSQL performance and provides optimization recommendations

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-studio_platform}"
DB_USER="${DB_USER:-postgres}"
REPORT_FILE="${REPORT_FILE:-performance_report_$(date +%Y%m%d_%H%M%S).txt}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== PostgreSQL Performance Analysis ===${NC}"
echo "Database: ${DB_NAME}"
echo "Report: ${REPORT_FILE}"
echo ""

# Function to run SQL query
run_query() {
    psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -A -c "$1"
}

# 1. Check database size
echo -e "${BLUE}1. Database Size${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        pg_size_pretty(pg_database_size('${DB_NAME}')) as database_size;
" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 2. Top 10 largest tables
echo -e "${BLUE}2. Top 10 Largest Tables${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        schemaname || '.' || tablename as table_name,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
        pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
        pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as indexes_size
    FROM pg_tables
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
    LIMIT 10;
" | column -t -s '|' | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 3. Unused indexes
echo -e "${BLUE}3. Unused Indexes (candidates for removal)${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        schemaname || '.' || tablename as table_name,
        indexname,
        pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
        idx_scan as scans
    FROM pg_stat_user_indexes
    WHERE idx_scan = 0
    AND indexrelname NOT LIKE '%_pkey'
    ORDER BY pg_relation_size(indexrelid) DESC;
" | column -t -s '|' | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 4. Tables needing indexes (high sequential scans)
echo -e "${BLUE}4. Tables with High Sequential Scans${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        schemaname || '.' || tablename as table_name,
        seq_scan as sequential_scans,
        seq_tup_read as rows_read,
        idx_scan as index_scans,
        CASE
            WHEN idx_scan = 0 THEN 'No index usage'
            ELSE ROUND((seq_scan::numeric / (seq_scan + idx_scan)) * 100, 2)::text || '%'
        END as seq_scan_ratio
    FROM pg_stat_user_tables
    WHERE seq_scan > 100
    ORDER BY seq_scan DESC
    LIMIT 10;
" | column -t -s '|' | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 5. Tables needing VACUUM
echo -e "${BLUE}5. Tables Needing VACUUM${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        schemaname || '.' || tablename as table_name,
        n_dead_tup as dead_tuples,
        n_live_tup as live_tuples,
        ROUND((n_dead_tup::numeric / NULLIF(n_live_tup, 0)) * 100, 2) as dead_ratio,
        last_vacuum,
        last_autovacuum
    FROM pg_stat_user_tables
    WHERE n_dead_tup > 1000
    ORDER BY n_dead_tup DESC
    LIMIT 10;
" | column -t -s '|' | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 6. Slow queries (requires pg_stat_statements)
echo -e "${BLUE}6. Top 10 Slowest Queries${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        ROUND(mean_exec_time::numeric, 2) as avg_time_ms,
        calls,
        ROUND(total_exec_time::numeric, 2) as total_time_ms,
        LEFT(query, 100) as query_preview
    FROM pg_stat_statements
    WHERE query NOT LIKE '%pg_stat_statements%'
    ORDER BY mean_exec_time DESC
    LIMIT 10;
" 2>/dev/null | column -t -s '|' | tee -a "${REPORT_FILE}" || echo "pg_stat_statements not enabled" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 7. Connection statistics
echo -e "${BLUE}7. Connection Statistics${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        state,
        COUNT(*) as count
    FROM pg_stat_activity
    WHERE pid != pg_backend_pid()
    GROUP BY state
    ORDER BY count DESC;
" | column -t -s '|' | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 8. Cache hit ratio
echo -e "${BLUE}8. Cache Hit Ratio${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        ROUND(
            (sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0)) * 100,
            2
        ) as cache_hit_ratio
    FROM pg_statio_user_tables;
" | tee -a "${REPORT_FILE}"
echo "Target: > 99%" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 9. Index hit ratio
echo -e "${BLUE}9. Index Hit Ratio${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        ROUND(
            (sum(idx_blks_hit) / NULLIF(sum(idx_blks_hit) + sum(idx_blks_read), 0)) * 100,
            2
        ) as index_hit_ratio
    FROM pg_statio_user_indexes;
" | tee -a "${REPORT_FILE}"
echo "Target: > 99%" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# 10. Recommendations
echo -e "${BLUE}10. Recommendations${NC}" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# Check for tables without primary keys
echo -e "${YELLOW}Tables without primary keys:${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        schemaname || '.' || tablename as table_name
    FROM pg_tables t
    WHERE schemaname = 'public'
    AND NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        WHERE c.conrelid = (t.schemaname||'.'||t.tablename)::regclass
        AND c.contype = 'p'
    );
" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# Check for missing foreign key indexes
echo -e "${YELLOW}Foreign keys without indexes:${NC}" | tee -a "${REPORT_FILE}"
run_query "
    SELECT
        c.conrelid::regclass as table_name,
        string_agg(a.attname, ', ') as columns
    FROM pg_constraint c
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
    WHERE c.contype = 'f'
    AND NOT EXISTS (
        SELECT 1 FROM pg_index i
        WHERE i.indrelid = c.conrelid
        AND c.conkey::int[] <@ i.indkey::int[]
    )
    GROUP BY c.conrelid, c.conname;
" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

echo -e "${GREEN}Performance analysis complete!${NC}"
echo "Report saved to: ${REPORT_FILE}"
