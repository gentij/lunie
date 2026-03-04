# Development Guide

This guide covers everything you need to know to develop and run Taskforge.

## Prerequisites

- **Node.js** 20 or later
- **pnpm** 9 or later
- **Docker** and **Docker Compose** (for PostgreSQL and Redis)
- **Git**

## Installation

```bash
# Clone the repository
git clone https://github.com/taskforge/taskforge.git
cd taskforge

# Install dependencies
pnpm install
```

## Infrastructure Setup

Start the required services (PostgreSQL and Redis) using Docker Compose:

```bash
cd deploy/compose
docker compose up -d
```

Verify services are running:

```bash
# Check PostgreSQL
docker compose ps | grep postgres

# Check Redis  
docker compose ps | grep redis
```

## Database Setup

Navigate to the server app and set up the database:

```bash
cd apps/server

# Generate Prisma client from schema
pnpm prisma:generate

# Run migrations (creates tables)
pnpm prisma:migrate dev

# Optionally seed with sample data
pnpm prisma:seed
```

**Note:** The `DATABASE_URL` environment variable must be set. See [Environment Variables](#environment-variables).

## Building Packages

Build all packages in the correct order:

```bash
# Build db-access (depends on generated Prisma types)
cd packages/db-access
pnpm build

# Build queue-config
cd ../queue-config
pnpm build

# Build contracts
cd ../contracts
pnpm build
```

Or build everything at once from the root:

```bash
# Build packages
cd packages/db-access && pnpm build
cd ../queue-config && pnpm build
cd ../contracts && pnpm build

# Build apps
cd ../apps/server && pnpm build
cd ../worker && pnpm build
```

## Running the Application

### Development Mode

Start the server in watch mode:

```bash
cd apps/server
pnpm run start:dev
```

The API will be available at `http://localhost:3000`.

### Production Mode

```bash
# Build first
pnpm build

# Start the server
pnpm run start:prod
```

## Testing

Run all tests:

```bash
cd apps/server
pnpm test
```

Run tests in watch mode:

```bash
pnpm test:watch
```

Run tests with coverage:

```bash
pnpm test:cov
```

## Linting and Formatting

### Check Formatting

```bash
pnpm format:check
```

### Auto-format Code

```bash
pnpm format
```

### Check Linting

```bash
pnpm lint
```

### Auto-fix Linting Issues

```bash
pnpm lint:fix
```

## Database Commands

### Generate Prisma Client

```bash
cd apps/server
pnpm prisma:generate
```

**Required** whenever the `prisma/schema.prisma` file changes.

### Run Migrations

```bash
# Development (creates migration file + runs it)
pnpm prisma:migrate dev

# Production (runs existing migrations)
pnpm prisma:migrate deploy

# Reset database (WARNING: deletes all data)
pnpm prisma:reset
```

### Create a New Migration

```bash
pnpm prisma:migrate dev --name migration_name
```

### Open Prisma Studio

Interactive database UI:

```bash
pnpm prisma:studio
```

## Environment Variables

Create a `.env` file in `apps/server/`:

```env
# Required
DATABASE_URL=postgresql://user:password@localhost:5432/taskforge
REDIS_URL=redis://localhost:6379

# Optional
PORT=3000
VERSION=0.1.0
LOG_LEVEL=info
```

### Variable Reference

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `DATABASE_URL` | Yes | PostgreSQL connection string | `postgresql://user:pass@localhost:5432/taskforge` |
| `REDIS_URL` | Yes | Redis connection string | `redis://localhost:6379` |
| `PORT` | No | HTTP server port | `3000` |
| `VERSION` | No | API version string | `0.1.0` |
| `LOG_LEVEL` | No | Logging level | `debug`, `info`, `warn`, `error` |

## Common Tasks

### Adding a New Repository

1. Create the repository file in `packages/db-access/src/<entity>/repository.ts`
2. Export it from `packages/db-access/src/index.ts`
3. Import and use in server/worker modules

### Adding a New Workflow Step Type

1. Add schema in `packages/contracts/src/workflow-definition.ts`
2. Create executor in `apps/worker/src/executors/<step-type>/`
3. Register executor in `apps/worker/src/executors/executor-registry.ts`

### Adding a New API Endpoint

1. Create DTOs in `apps/server/src/<module>/dto/`
2. Create controller in `apps/server/src/<module>/<module>.controller.ts`
3. Register route in module

## Troubleshooting

### Prisma Client Not Generated

```bash
# Regenerate Prisma client
cd apps/server
pnpm prisma:generate
```

### Redis Connection Failed

- Verify Redis is running: `docker compose ps | grep redis`
- Check `REDIS_URL` is correct in `.env`

### Database Connection Failed

- Verify PostgreSQL is running: `docker compose ps | grep postgres`
- Check `DATABASE_URL` is correct in `.env`
- Ensure database exists: `pnpm prisma: migrate dev`

### Type Errors After Schema Change

```bash
# Full rebuild
pnpm prisma:generate
cd ../../packages/db-access && pnpm build
cd ../apps/server && pnpm build
```

## Making Changes

### Code Style

- Use TypeScript strict mode
- Follow existing patterns in the codebase
- Run `pnpm format` before committing
- Run `pnpm lint:fix` to auto-fix issues

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(core): add new workflow trigger type
fix(db): resolve connection timeout issue
docs(readme): update installation instructions
chore: update dependencies
```

## Learning Resources

- [Taskforge TUI Guide](./Taskforge%20-%20TUI.md)
- [Workflow Engine Concepts](./Taskforge%20-%20Workflow%20Engine.md)
- [Queues and Workers Plan](./Taskforge%20-%20Queues%20and%20Workers%20Plan.md)
- [NestJS Documentation](https://docs.nestjs.com)
- [Prisma Documentation](https://www.prisma.io/docs)
- [BullMQ Documentation](https://docs.bullmq.io)
