#!/bin/bash

# Database Restore Script
# Usage: ./scripts/restore-db.sh <backup_file>

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <backup_file>"
  echo "Example: $0 ./backups/studio_db_20260304_120000.sql.gz"
  exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: Backup file not found: $BACKUP_FILE"
  exit 1
fi

# Configuration
DB_NAME="${STUDIO_DATABASE_NAME:-studio_db}"
DB_USER="${STUDIO_DATABASE_USER:-studio}"
DB_HOST="${STUDIO_DATABASE_HOST:-localhost}"
DB_PORT="${STUDIO_DATABASE_PORT:-5432}"

echo "WARNING: This will restore the database from backup."
echo "Database: $DB_NAME"
echo "Backup file: $BACKUP_FILE"
echo ""
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
  echo "Restore cancelled."
  exit 0
fi

echo "Starting database restore..."

# Drop existing connections
echo "Terminating existing connections..."
PGPASSWORD="${STUDIO_DATABASE_PASSWORD}" psql \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d postgres \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" \
  2>/dev/null || true

# Restore from backup
echo "Restoring database..."
gunzip -c "$BACKUP_FILE" | PGPASSWORD="${STUDIO_DATABASE_PASSWORD}" psql \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  --single-transaction \
  2>&1

if [ $? -eq 0 ]; then
  echo "✓ Database restored successfully"
else
  echo "✗ Restore failed"
  exit 1
fi

echo ""
echo "Restore process completed successfully!"
