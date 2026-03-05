#!/bin/bash

# Redis Backup Script
# This script creates backups of Redis RDB snapshots

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/redis}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
REDIS_DATA_DIR="${REDIS_DATA_DIR:-/var/lib/redis}"
REDIS_RDB_FILE="${REDIS_RDB_FILE:-dump.rdb}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/redis_backup_${TIMESTAMP}.rdb"

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

echo "Starting Redis backup at $(date)"
echo "Backup file: ${BACKUP_FILE}"

# Trigger Redis BGSAVE
redis-cli BGSAVE

# Wait for BGSAVE to complete
echo "Waiting for BGSAVE to complete..."
while [ "$(redis-cli LASTSAVE)" == "$(redis-cli LASTSAVE)" ]; do
    sleep 1
done

# Copy RDB file
if [ -f "${REDIS_DATA_DIR}/${REDIS_RDB_FILE}" ]; then
    cp "${REDIS_DATA_DIR}/${REDIS_RDB_FILE}" "${BACKUP_FILE}"

    # Compress backup
    gzip "${BACKUP_FILE}"
    BACKUP_FILE="${BACKUP_FILE}.gz"

    echo "Backup completed successfully"

    # Get backup file size
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo "Backup size: ${BACKUP_SIZE}"

    # Create checksum
    sha256sum "${BACKUP_FILE}" > "${BACKUP_FILE}.sha256"
    echo "Checksum created: ${BACKUP_FILE}.sha256"

    # Remove old backups
    echo "Removing backups older than ${RETENTION_DAYS} days..."
    find "${BACKUP_DIR}" -name "redis_backup_*.rdb.gz" -mtime +${RETENTION_DAYS} -delete
    find "${BACKUP_DIR}" -name "redis_backup_*.rdb.gz.sha256" -mtime +${RETENTION_DAYS} -delete

    echo "Backup completed at $(date)"
    exit 0
else
    echo "Redis RDB file not found: ${REDIS_DATA_DIR}/${REDIS_RDB_FILE}"
    exit 1
fi
