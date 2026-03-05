#!/bin/bash

# PostgreSQL Restore Script
# This script restores PostgreSQL database from backup

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-studio_platform}"
DB_USER="${DB_USER:-postgres}"

# Check if backup file is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <backup_file>"
    echo ""
    echo "Available backups:"
    ls -lh "${BACKUP_DIR}"/backup_${DB_NAME}_*.sql.gz 2>/dev/null || echo "No backups found"
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "${BACKUP_FILE}" ]; then
    echo "Error: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

# Verify checksum if exists
if [ -f "${BACKUP_FILE}.sha256" ]; then
    echo "Verifying backup checksum..."
    if sha256sum -c "${BACKUP_FILE}.sha256"; then
        echo "Checksum verified successfully"
    else
        echo "Error: Checksum verification failed!"
        exit 1
    fi
fi

echo "WARNING: This will drop and recreate the database: ${DB_NAME}"
echo "Backup file: ${BACKUP_FILE}"
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "${CONFIRM}" != "yes" ]; then
    echo "Restore cancelled"
    exit 0
fi

echo "Starting database restore at $(date)"

# Drop existing database (if exists)
echo "Dropping existing database..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -c "DROP DATABASE IF EXISTS ${DB_NAME};"

# Create new database
echo "Creating new database..."
psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -c "CREATE DATABASE ${DB_NAME};"

# Restore from backup
echo "Restoring from backup..."
pg_restore -h "${DB_HOST}" \
           -p "${DB_PORT}" \
           -U "${DB_USER}" \
           -d "${DB_NAME}" \
           --verbose \
           "${BACKUP_FILE}"

if [ $? -eq 0 ]; then
    echo "Database restored successfully at $(date)"
    exit 0
else
    echo "Database restore failed!"
    exit 1
fi
