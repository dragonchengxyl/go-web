#!/bin/bash

# Health Check Script
# This script monitors the health of all services

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
ALERT_EMAIL="${ALERT_EMAIL:-admin@example.com}"
LOG_FILE="${LOG_FILE:-/var/log/health_check.log}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "${LOG_FILE}"
}

# Function to check HTTP endpoint
check_http() {
    local url=$1
    local name=$2
    local timeout=${3:-5}

    log "Checking ${name}..."

    response=$(curl -s -o /dev/null -w "%{http_code}" --max-time ${timeout} "${url}" 2>&1)

    if [ "${response}" == "200" ]; then
        echo -e "${GREEN}✓${NC} ${name}: OK"
        return 0
    else
        echo -e "${RED}✗${NC} ${name}: FAILED (HTTP ${response})"
        return 1
    fi
}

# Function to check database
check_database() {
    log "Checking PostgreSQL..."

    if psql -h "${DB_HOST:-localhost}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-studio_platform}" -c "SELECT 1" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} PostgreSQL: OK"
        return 0
    else
        echo -e "${RED}✗${NC} PostgreSQL: FAILED"
        return 1
    fi
}

# Function to check Redis
check_redis() {
    log "Checking Redis..."

    if redis-cli -h "${REDIS_HOST:-localhost}" -p "${REDIS_PORT:-6379}" ping > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Redis: OK"
        return 0
    else
        echo -e "${RED}✗${NC} Redis: FAILED"
        return 1
    fi
}

# Function to check disk space
check_disk_space() {
    log "Checking disk space..."

    local threshold=80
    local usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

    if [ "${usage}" -lt "${threshold}" ]; then
        echo -e "${GREEN}✓${NC} Disk space: ${usage}% used"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} Disk space: ${usage}% used (threshold: ${threshold}%)"
        return 1
    fi
}

# Function to check memory
check_memory() {
    log "Checking memory..."

    local threshold=80
    local usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')

    if [ "${usage}" -lt "${threshold}" ]; then
        echo -e "${GREEN}✓${NC} Memory: ${usage}% used"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} Memory: ${usage}% used (threshold: ${threshold}%)"
        return 1
    fi
}

# Function to send alert
send_alert() {
    local message=$1

    log "ALERT: ${message}"

    # Send email (requires mailutils or similar)
    if command -v mail &> /dev/null; then
        echo "${message}" | mail -s "Health Check Alert" "${ALERT_EMAIL}"
    fi

    # You can add other alerting mechanisms here (Slack, PagerDuty, etc.)
}

# Main health check
main() {
    log "=== Starting Health Check ==="

    local failed=0

    # Check API health endpoint
    check_http "${API_URL}/health" "API Health" || ((failed++))

    # Check API readiness endpoint
    check_http "${API_URL}/ready" "API Readiness" || ((failed++))

    # Check database (if credentials are available)
    if [ -n "${DB_HOST}" ]; then
        check_database || ((failed++))
    fi

    # Check Redis (if host is available)
    if [ -n "${REDIS_HOST}" ]; then
        check_redis || ((failed++))
    fi

    # Check system resources
    check_disk_space || ((failed++))
    check_memory || ((failed++))

    # Summary
    log "=== Health Check Complete ==="

    if [ ${failed} -eq 0 ]; then
        echo -e "${GREEN}All checks passed!${NC}"
        exit 0
    else
        echo -e "${RED}${failed} check(s) failed!${NC}"
        send_alert "Health check failed: ${failed} service(s) are unhealthy"
        exit 1
    fi
}

# Run main function
main
