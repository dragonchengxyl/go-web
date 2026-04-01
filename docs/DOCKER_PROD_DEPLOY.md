# Docker Production Deployment

This repository now supports a single-server production topology that keeps
`nginx` on the host and runs `frontend`, `backend`, `postgres`, and `redis`
inside Docker.

## Topology

- Host `nginx` proxies:
  - `/` -> `127.0.0.1:13000`
  - `/api/` -> `127.0.0.1:18080`
  - `/ws` -> `127.0.0.1:18080`
- Docker network contains:
  - `frontend`
  - `backend`
  - `postgres`
  - `redis`

## Files

- `Dockerfile`: production backend image
- `apps/web/Dockerfile`: production frontend image
- `docker-compose.prod.yml`: single-server production stack
- `.env.prod.example`: production environment template

## First-time setup

1. Copy `.env.prod.example` to `.env.prod`
2. Fill in real credentials and image names
3. Start the stack:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
```

## Migration flow

Run schema migrations as a one-shot task before the backend comes up:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml up migrate
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d backend frontend
```

The `backend` container runs with `-run-migrations=false` so multiple replicas
do not compete on startup.

## Virtual data seeding

The production compose stack also exposes a one-shot `seed` task that uses the
same backend image and config as the running service.

Default usage:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

Recommended first run on a fresh environment:

```bash
SEED_PROFILE=small docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

Then move to a denser dataset:

```bash
SEED_PROFILE=medium SEED_NAMESPACE=bulk docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

Notes:

- `SEED_MODE` defaults to `bulk`
- `SEED_PROFILE` supports `small`, `medium`, `large`
- `SEED_NAMESPACE` keeps generated UUID-based records deterministic across reruns
- The command prints database size before and after seeding

## Kafka and async workers

The production stack uses Kafka + Outbox for business events. Start:

- `kafka` via `--profile kafka`
- `outbox-relay`, `notification-svc`, `moderation-svc`, `audio-worker` via `--profile async`

The bundled Kafka service is mirrored to GHCR and referenced via `KAFKA_IMAGE`,
so production hosts do not need to pull Kafka directly from Docker Hub.

Detailed enablement and verification steps:

- [`docs/KAFKA_RUNBOOK.md`](KAFKA_RUNBOOK.md)

## Data migration recommendation

When migrating from host-installed PostgreSQL and Redis:

1. Stop old app traffic
2. Dump PostgreSQL with `pg_dumpall`
3. Run `BGSAVE` for Redis and copy `dump.rdb`
4. Start containerized `postgres` and `redis`
5. Restore data into the containers
6. Start `migrate`, `backend`, and `frontend`
7. Switch host `nginx` upstreams to the new local ports

## CI/CD shape

Recommended pipeline:

1. Push to `main`
2. CI builds backend and frontend images and pushes them to GHCR
3. After the image workflow succeeds, deploy runs automatically and updates the server

Manual fallback:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
```
