# Lunie

Lunie is a self-hosted workflow automation engine for technical users who want local control, API-first automation, and inspectable execution history.

Documentation website: https://gentij.github.io/lunie

## MVP Status

Lunie currently ships:

- API server
- Worker-based step execution
- CLI for lifecycle management
- TUI for operator workflows (with known limitations)

Current MVP step types:

- `http`
- `transform`
- `condition`

For install-and-use instructions (non-contributor path), start with `docs/Getting-Started.md`.

## Requirements

- Node.js 20+
- pnpm 9+
- Docker + Docker Compose
- Go 1.25+ (required for CLI development from source)

## Quick Start (From Source)

```bash
# 1) Clone and install dependencies
git clone https://github.com/gentij/lunie.git
cd lunie
pnpm install

# 2) Start infrastructure
docker compose -f deploy/compose/docker-compose.yml up -d

# 3) Create server env file
cp apps/server/.env.example apps/server/.env

# 4) Set required secrets in apps/server/.env
#    LUNIE_ADMIN_TOKEN must be at least 32 chars
#    LUNIE_SECRET_KEY must be 64-char hex or base64(32 bytes)

# 5) Generate Prisma client + run migrations
pnpm -C apps/server prisma:generate
pnpm -C apps/server prisma:migrate deploy

# 6) Build shared packages
pnpm -C packages/db-access build
pnpm -C packages/queue build
pnpm -C packages/contracts build

# 7) Start API server and worker (in separate terminals)
pnpm -C apps/server start:dev
pnpm -C apps/worker start:dev
```

Default local endpoints:

- API base URL: `http://localhost:3000/v1/api`
- Swagger UI: `http://localhost:3000/api`

If port `3000` is already used on your machine, run the server with a different port, for example:

```bash
PORT=3100 pnpm -C apps/server start:dev
```

And point CLI commands to that base URL:

```bash
./lunie --server "http://localhost:3100/v1/api" auth whoami
```

## First CLI Run

Build the CLI:

```bash
cd apps/cli
go build -o ../../lunie ./cmd/lunie
cd ../..
```

Authenticate and verify:

```bash
./lunie auth login --token "<LUNIE_ADMIN_TOKEN>"
./lunie auth whoami
```

Create and execute a minimal workflow:

```bash
cat > /tmp/lunie-definition.json <<'JSON'
{
  "input": {
    "apiBase": "https://jsonplaceholder.typicode.com"
  },
  "steps": [
    {
      "key": "fetch_post",
      "type": "http",
      "request": {
        "method": "GET",
        "url": "{{input.apiBase}}/posts/1"
      }
    }
  ]
}
JSON

./lunie workflow create --name "MVP Test" --definition /tmp/lunie-definition.json
./lunie workflow list
# run with the workflow id from list output
./lunie workflow run <workflow-id>
```

Run the terminal UI:

```bash
./lunie tui
```

## Documentation

- [Getting Started](./docs/Getting-Started.md)
- [CLI Usage](./docs/CLI-Usage.md)
- [TUI Guide](./docs/Lunie%20-%20TUI.md)

Contributor and architecture docs:

- [Development Guide](./docs/Development.md)
- [Architecture Overview](./docs/Architecture.md)
- [MVP Release Readiness](./docs/MVP-Release-Readiness.md)

Docs website (Docusaurus):

```bash
pnpm docs:dev
```

## Project Structure

```text
lunie/
|- apps/
|  |- server/      # NestJS + Fastify API
|  |- worker/      # BullMQ worker
|  \- cli/         # Go CLI + TUI
|- packages/
|  |- contracts/   # Zod schemas and contracts
|  |- db-access/   # Prisma service + repositories
|  \- queue/       # Queue configuration package
|- deploy/         # Docker and compose files
|- docs/           # Product, architecture, and development docs
\- scripts/        # Utility scripts
```
