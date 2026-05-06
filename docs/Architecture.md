---
id: architecture-overview
title: Architecture Overview
description: System architecture for Lunie server, worker, CLI, and data stores.
slug: /architecture
---

# Architecture Overview

Lunie is a self-hosted workflow automation system built as a monorepo.

## System Components

```text
Lunie
|- CLI/TUI (Go)
|  \- Calls Lunie HTTP API
|- Server (NestJS + Fastify)
|  |- Auth, workflows, triggers, runs, steps, secrets
|  \- Enqueues step-run jobs
|- Worker (NestJS + BullMQ)
|  \- Consumes jobs and executes workflow steps
|- PostgreSQL
|  \- Durable state for all entities and run history
\- Redis
   \- Queue backend for step-run execution
```

## Monorepo Structure

### Apps

- `apps/server`: API server
- `apps/worker`: queue consumer and step execution engine
- `apps/cli`: CLI + TUI interface

### Packages

- `packages/contracts`: shared schemas/contracts
- `packages/db-access`: Prisma service and repositories
- `packages/queue`: queue configuration package (`@lunie/queue-config`)

## API Shape

- Global API prefix: `/v{VERSION}/api`
- Default local base URL: `http://localhost:3000/v1/api`
- Swagger UI: `http://localhost:3000/api`

## Execution Flow

### Workflow Creation

1. Create workflow
2. Create workflow version (definition snapshot)
3. Create trigger(s)

### Workflow Run

1. Trigger is invoked (manual/webhook/cron)
2. Event is stored
3. Workflow run is created (`QUEUED`)
4. Step runs are created (`QUEUED`)
5. Step runs are enqueued to Redis (`step-runs` queue)
6. Worker executes each step
7. Step runs transition to terminal state
8. Workflow run resolves to final status

Webhook ingress supports a public path-key route for `WEBHOOK` triggers:

- `POST /v1/api/hooks/:workflowKey/:triggerKey/:webhookKey`

Webhook keys are trigger-scoped and rotated by operators via CLI/API.

## Current MVP Execution Types

Worker executors currently include:

- `http`
- `transform`
- `condition`

## Build and Dependency Order

When contracts or Prisma types change, build in this order:

1. `pnpm -C apps/server prisma:generate`
2. `pnpm -C packages/db-access build`
3. `pnpm -C packages/queue build`
4. `pnpm -C packages/contracts build`
5. `pnpm -C apps/server build`
6. `pnpm -C apps/worker build`

## Runtime Requirements

Lunie server/worker can run with either:

- external managed Postgres + Redis services, or
- bundled local Postgres + Redis containers (CLI opt-in).

Required server env vars:

- `DATABASE_URL`
- `REDIS_URL`
- `LUNIE_ADMIN_TOKEN`
- `LUNIE_SECRET_KEY`

Optional:

- `PORT` (default `3000`)
- `VERSION` (default `1`)
- `CACHE_TTL_SECONDS`
- `CACHE_REDIS_PREFIX`

## Related Documentation

- Getting started: `docs/Getting-Started.md`
- CLI usage: `docs/CLI-Usage.md`
- Development setup: `docs/Development.md`
- CLI commands: `apps/cli/README.md`
- TUI usage and internals: `docs/Lunie - TUI.md`
