#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env.prod}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
KAFKA_PARTITIONS="${KAFKA_PARTITIONS:-3}"
KAFKA_REPLICATION_FACTOR="${KAFKA_REPLICATION_FACTOR:-1}"

if ! docker compose version >/dev/null 2>&1; then
  echo "docker compose is required" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

if [[ "${STUDIO_KAFKA_ENABLED:-false}" != "true" ]]; then
  echo "STUDIO_KAFKA_ENABLED is not true in $ENV_FILE" >&2
  exit 1
fi

DC=(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE")

BASE_TOPIC="${STUDIO_KAFKA_TOPIC:-furry-events}"
KAFKA_BROKERS="${STUDIO_KAFKA_BROKERS:-}"

echo "==> Running database migrations"
"${DC[@]}" run --rm migrate

echo "==> Starting core services"
"${DC[@]}" up -d --force-recreate backend frontend

if [[ "$KAFKA_BROKERS" == *"kafka:"* ]]; then
  echo "==> Starting internal Kafka broker"
  "${DC[@]}" --profile kafka up -d kafka

  echo "==> Waiting for Kafka to become ready"
  for _ in $(seq 1 30); do
    if "${DC[@]}" exec -T kafka /opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done

  echo "==> Ensuring Kafka topics exist"
  for topic in "${BASE_TOPIC}.content" "${BASE_TOPIC}.social" "${BASE_TOPIC}.audio" "${BASE_TOPIC}.dlq"; do
    "${DC[@]}" exec -T kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
      --bootstrap-server localhost:9092 \
      --create \
      --if-not-exists \
      --topic "$topic" \
      --partitions "$KAFKA_PARTITIONS" \
      --replication-factor "$KAFKA_REPLICATION_FACTOR"
  done
else
  echo "==> External Kafka configured: $KAFKA_BROKERS"
  echo "    Ensure topics exist:"
  echo "    - ${BASE_TOPIC}.content"
  echo "    - ${BASE_TOPIC}.social"
  echo "    - ${BASE_TOPIC}.audio"
  echo "    - ${BASE_TOPIC}.dlq"
fi

echo "==> Starting async Kafka workers"
"${DC[@]}" --profile async up -d outbox-relay notification-svc audio-worker

if [[ -n "${STUDIO_MODERATION_ACCESS_KEY_ID:-}" ]]; then
  echo "==> Starting moderation service"
  "${DC[@]}" --profile async up -d moderation-svc
else
  echo "==> Skipping moderation-svc: STUDIO_MODERATION_ACCESS_KEY_ID is empty"
fi

echo "==> Checking health endpoints"
curl -sf "http://127.0.0.1:${BACKEND_HOST_PORT:-18080}/health" >/dev/null && echo "backend ok"
curl -sf "http://127.0.0.1:${FRONTEND_HOST_PORT:-13000}/" >/dev/null && echo "frontend ok"
curl -sf "http://127.0.0.1:${STUDIO_OBSERVABILITY_OUTBOX_HTTP_PORT:-18055}/health" >/dev/null && echo "outbox-relay ok"
curl -sf "http://127.0.0.1:${STUDIO_OBSERVABILITY_NOTIFICATION_HTTP_PORT:-18052}/health" >/dev/null && echo "notification-svc ok"
curl -sf "http://127.0.0.1:${STUDIO_OBSERVABILITY_AUDIO_WORKER_HTTP_PORT:-18054}/health" >/dev/null && echo "audio-worker ok"

if [[ -n "${STUDIO_MODERATION_ACCESS_KEY_ID:-}" ]]; then
  curl -sf "http://127.0.0.1:${STUDIO_OBSERVABILITY_MODERATION_HTTP_PORT:-18053}/health" >/dev/null && echo "moderation-svc ok"
fi

echo "==> Kafka enablement complete"
echo "Metrics:"
echo "  http://127.0.0.1:${STUDIO_OBSERVABILITY_OUTBOX_HTTP_PORT:-18055}/metrics"
echo "  http://127.0.0.1:${STUDIO_OBSERVABILITY_NOTIFICATION_HTTP_PORT:-18052}/metrics"
echo "  http://127.0.0.1:${STUDIO_OBSERVABILITY_AUDIO_WORKER_HTTP_PORT:-18054}/metrics"
if [[ -n "${STUDIO_MODERATION_ACCESS_KEY_ID:-}" ]]; then
  echo "  http://127.0.0.1:${STUDIO_OBSERVABILITY_MODERATION_HTTP_PORT:-18053}/metrics"
fi
