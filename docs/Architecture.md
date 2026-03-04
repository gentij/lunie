# Architecture Overview

Taskforge is a self-hosted workflow automation engine built with a monorepo architecture.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Taskforge System                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐ │
│  │    CLI      │      │   Server    │      │   Worker    │ │
│  │  (future)   │─────▶│  (NestJS)   │      │  (BullMQ)   │ │
│  └─────────────┘      │  Fastify    │      │  Consumer   │ │
│                       └──────┬──────┘      └──────┬──────┘ │
│                              │                     │        │
│                              ▼                     ▼        │
│                       ┌─────────────┐       ┌─────────────┐ │
│                       │  PostgreSQL │       │    Redis    │ │
│                       └─────────────┘       └─────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Monorepo Structure

### Apps

| App | Purpose | Commands |
|-----|---------|----------|
| `server` | REST API, trigger handlers, job enqueueing | `pnpm start:dev` |
| `worker` | Consumes jobs from Redis, executes steps | `pnpm start` |

### Packages

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `contracts` | Zod schemas for API, workflow definitions | zod |
| `db-access` | PrismaService + all repositories | @prisma/client, @nestjs/common |
| `queue-config` | BullMQ configuration (Redis connection) | @nestjs/bullmq, @nestjs/config |

## Build Order

```
packages/db-access  ──┐
packages/contracts  ──┼─▶ apps/server ──┐
packages/queue-config ─┘                  │──▶ apps/worker
                                          │
(depends on generated @prisma/client types)
```

**Important:** When the Prisma schema changes:
1. Run `pnpm prisma:generate` in `apps/server`
2. Rebuild `packages/db-access`
3. Rebuild `apps/server` and `apps/worker`

## Data Flow

### Creating a Workflow

1. **Define workflow** → POST `/api/workflows`
2. **Create version** → POST `/api/workflows/:id/versions`
3. **Add trigger** → POST `/api/workflows/:id/triggers`

### Running a Workflow

```
1. Trigger received (webhook, cron, manual)
         ↓
2. Event created in PostgreSQL
         ↓
3. WorkflowRun created (QUEUED)
         ↓
4. StepRuns created (QUEUED) for each step
         ↓
5. Each StepRun enqueued to Redis 'step-runs' queue
         ↓
6. Worker picks up job, executes step
         ↓
7. StepRun updated (RUNNING → SUCCEEDED/FAILED)
         ↓
8. WorkflowRun updated when all steps complete
```

## Key Technologies

- **NestJS** - Application framework
- **Fastify** - HTTP server (faster than Express)
- **Prisma** - ORM with PostgreSQL
- **BullMQ** - Redis-based job queue
- **Zod** - Runtime type validation
- **pnpm** - Monorepo package manager

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | Yes | Redis connection string |
| `PORT` | No | Server port (default: 3000) |
| `VERSION` | No | API version (default: 0.1.0) |
| `LOG_LEVEL` | No | Pino log level (default: info) |

## Related Documentation

- TUI usage and internals: `docs/Taskforge - TUI.md`
- Development setup and workflows: `docs/Development.md`
