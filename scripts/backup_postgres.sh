#!/bin/bash

# PostgreSQL Backup Script
# This script creates full and incremental backups of PostgreSQL database

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-studio_platform}"
DB_USER="${DB_USER:-postgres}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/backup_${DB_NAME}_${TIMESTAMP}.sql.gz"

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

echo "Starting PostgreSQL backup at $(date)"
echo "Database: ${DB_NAME}"
echo "Backup file: ${BACKUP_FILE}"

# Perform backup
pg_dump -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        --format=custom \
        --compress=9 \
        --file="${BACKUP_FILE}"

# Check if backup was successful
if [ $? -eq 0 ]; then
    echo "Backup completed successfully"

    # Get backup file size
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo "Backup size: ${BACKUP_SIZE}"

    # Create checksum
    sha256sum "${BACKUP_FILE}" > "${BACKUP_FILE}.sha256"
    echo "Checksum created: ${BACKUP_FILE}.sha256"

    # Remove old backups
    echo "Removing backups older than ${RETENTION_DAYS} days..."
    find "${BACKUP_DIR}" -name "backup_${DB_NAME}_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
    find "${BACKUP_DIR}" -name "backup_${DB_NAME}_*.sql.gz.sha256" -mtime +${RETENTION_DAYS} -delete

    echo "Backup completed at $(date)"
    exit 0
else
    echo "Backup failed!"
    exit 1
fi
