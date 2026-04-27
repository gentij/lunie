---
id: development-guide
title: Development Guide
description: Local development setup for server, worker, CLI, and TUI.
slug: /development
---

# Development Guide

This guide covers local development for Lunie (server, worker, CLI, and TUI).

Use this page when contributing to Lunie itself. If you only want to install and run Lunie, start with `docs/Getting-Started.md`.

## Prerequisites

- Node.js 20+
- pnpm 9+
- Docker + Docker Compose
- Go 1.25+ (for CLI/TUI development)
- Git

## Clone and Install

```bash
git clone https://github.com/gentij/lunie.git
cd lunie
pnpm install
```

## Infrastructure Setup

Start PostgreSQL and Redis:

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

If your Docker installation does not support `docker compose`, use:

```bash
docker-compose -f deploy/compose/docker-compose.yml up -d
```

Verify services:

```bash
docker compose -f deploy/compose/docker-compose.yml ps
```

## Server Environment Setup

Create a local env file:

```bash
cp apps/server/.env.example apps/server/.env
```

Set required values in `apps/server/.env`:

- `DATABASE_URL`
- `REDIS_URL`
- `LUNIE_ADMIN_TOKEN` (minimum 32 characters)
- `LUNIE_SECRET_KEY` (64-char hex or base64 for 32 bytes)

Optional helper commands for secure values:

```bash
# 64-char hex (good for LUNIE_SECRET_KEY)
openssl rand -hex 32

# 64-char token (good for LUNIE_ADMIN_TOKEN)
openssl rand -hex 32
```

## Database and Build Setup

```bash
# Generate Prisma client
pnpm -C apps/server prisma:generate

# Apply migrations
pnpm -C apps/server prisma:migrate deploy

# Build shared packages
pnpm -C packages/db-access build
pnpm -C packages/queue build
pnpm -C packages/contracts build
```

## Run the Applications

Start server:

```bash
pnpm -C apps/server start:dev
```

Start worker (separate terminal):

```bash
pnpm -C apps/worker start:dev
```

API URLs:

- Base API: `http://localhost:3000/v1/api`
- Swagger: `http://localhost:3000/api`
- Health: `http://localhost:3000/v1/api/health`

If `3000` is busy, start server on another port:

```bash
PORT=3100 pnpm -C apps/server start:dev
```

Then use `--server http://localhost:3100/v1/api` in CLI commands.

## API Quickstart

Swagger UI is served at `/api` and uses bearer token auth.

Use your `LUNIE_ADMIN_TOKEN` to call the API directly:

```bash
TOKEN="<LUNIE_ADMIN_TOKEN>"

# Health (public)
curl -sS "http://localhost:3000/v1/api/health"

# Auth check (protected)
curl -sS "http://localhost:3000/v1/api/auth/whoami" \
  -H "Authorization: Bearer $TOKEN"
```

## CLI and TUI Development

Build CLI binary from source:

```bash
cd apps/cli
go build -o ../../lunie ./cmd/lunie
cd ../..
```

Common commands:

```bash
./lunie auth login --token "<LUNIE_ADMIN_TOKEN>"
./lunie auth whoami
./lunie workflow list
./lunie tui
```

## Testing

Run server tests:

```bash
pnpm -C apps/server test
pnpm -C apps/server test:e2e
```

Run worker tests:

```bash
pnpm -C apps/worker test
```

Run CLI/TUI tests:

```bash
cd apps/cli
go test ./...
cd ../..
```

## Linting and Formatting

Server:

```bash
pnpm -C apps/server lint
pnpm -C apps/server lint:fix
pnpm -C apps/server format:check
pnpm -C apps/server format
```

Worker:

```bash
pnpm -C apps/worker lint
pnpm -C apps/worker lint:fix
pnpm -C apps/worker format:check
pnpm -C apps/worker format
```

## Documentation Site

Lunie docs are served by Docusaurus from `apps/docs`, using markdown source files in `docs/`.

```bash
pnpm docs:dev
pnpm docs:build
```

## Environment Variable Reference

Lunie server variables (from runtime validation):

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string (`postgres://` or `postgresql://`) |
| `REDIS_URL` | Yes | Redis connection string (`redis://`) |
| `LUNIE_ADMIN_TOKEN` | Yes | Admin bearer token used by CLI/TUI and API clients |
| `LUNIE_SECRET_KEY` | Yes | Encryption key for secrets (64-char hex or base64 for 32 bytes) |
| `PORT` | No | API server port (default `3000`) |
| `CACHE_TTL_SECONDS` | No | Cache TTL in seconds (default `60`) |
| `CACHE_REDIS_PREFIX` | No | Cache key namespace prefix |
| `VERSION` | No | API version prefix used in routes (default `1`) |

## Common Development Tasks

### Add a New Workflow Step Type

1. Add schemas in `packages/contracts`
2. Implement executor in `apps/worker/src/executors/<step-type>/`
3. Register executor in `apps/worker/src/executors/executor-registry.ts`
4. Add/adjust validation in server if needed
5. Add tests (worker + server as applicable)

### Add a New API Endpoint

1. Add DTOs in the feature module
2. Add controller/service behavior
3. Add unit + e2e tests
4. Verify Swagger output

## Troubleshooting

### API returns unauthorized

- Verify `LUNIE_ADMIN_TOKEN` in `apps/server/.env`
- Re-run `./lunie auth login --token "<token>"`

### Worker is not processing runs

- Confirm worker is running (`pnpm -C apps/worker start:dev`)
- Confirm Redis is healthy
- Check worker logs for queue/executor errors

### Prisma/client type errors after schema changes

```bash
pnpm -C apps/server prisma:generate
pnpm -C packages/db-access build
pnpm -C apps/server build
pnpm -C apps/worker build
```

## Related Docs

- [Getting Started](./Getting-Started.md)
- [CLI Usage](./CLI-Usage.md)
- [Architecture Overview](./Architecture.md)
- [Lunie TUI Guide](./Lunie%20-%20TUI.md)
- [Workflow Engine Concepts](./Lunie%20-%20Workflow%20Engine.md)
