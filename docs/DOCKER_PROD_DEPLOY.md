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
