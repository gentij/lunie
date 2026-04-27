# Lunie Server

Lunie Server is the HTTP API for workflows, triggers, runs, step runs, events, and secrets.

It is built with NestJS + Fastify and backed by PostgreSQL + Redis.

## Prerequisites

- Node.js 20+
- pnpm 9+
- PostgreSQL
- Redis

## Setup

From repo root:

```bash
pnpm install
cp apps/server/.env.example apps/server/.env
```

Set required env vars in `apps/server/.env`:

- `DATABASE_URL`
- `REDIS_URL`
- `LUNIE_ADMIN_TOKEN`
- `LUNIE_SECRET_KEY`

Generate Prisma client and apply migrations:

```bash
pnpm -C apps/server prisma:generate
pnpm -C apps/server prisma:migrate deploy
```

Build required workspace packages:

```bash
pnpm -C packages/db-access build
pnpm -C packages/queue build
pnpm -C packages/contracts build
```

## Run

Development mode:

```bash
pnpm -C apps/server start:dev
```

Production build + run:

```bash
pnpm -C apps/server build
pnpm -C apps/server start:prod
```

## API Endpoints

- Base API: `http://localhost:3000/v1/api`
- Swagger docs: `http://localhost:3000/api`
- Health: `http://localhost:3000/v1/api/health`

## Tests

```bash
pnpm -C apps/server test
pnpm -C apps/server test:e2e
pnpm -C apps/server test:cov
```

## Lint and Format

```bash
pnpm -C apps/server lint
pnpm -C apps/server lint:fix
pnpm -C apps/server format:check
pnpm -C apps/server format
```

## Notes

- Server run execution depends on the worker process for step processing.
- If workflow runs remain queued, verify Redis and worker status.
