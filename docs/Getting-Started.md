---
id: getting-started
title: Getting Started
description: Install Lunie and run your first workflow in minutes.
slug: /getting-started
---

# Getting Started

This guide is for operators and users who want to install Lunie and run workflows.

If you want to contribute to Lunie itself, use `docs/Development.md`.

## 10-Minute Path

If `lunie` is already installed, this is the fastest path from zero to first run.

```bash
# 1) Initialize and start stack (external datastores)
lunie init \
  --database-url "postgresql://user:pass@db.example.com:5432/lunie" \
  --redis-url "redis://redis.example.com:6379"
lunie status

# 2) Verify auth works
lunie auth whoami

# 3) Create a minimal workflow definition
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

# 4) Create workflow and run it
lunie workflow create --name "First Workflow" --definition /tmp/lunie-definition.json
lunie workflow list
# copy workflow ID from list output
lunie workflow run <workflow-id>

# 5) Inspect execution
lunie run list <workflow-id>
lunie step list <workflow-id> <run-id>
```

## What You Need

- Docker + Docker Compose
- A Lunie CLI binary

## Install the CLI

Choose one method.

### Homebrew

```bash
brew tap gentij/lunie
brew install lunie
```

### AUR (Arch Linux)

```bash
yay -S lunie-cli-bin
```

### GitHub Release Binary

Download the matching release artifact from:

- `https://github.com/gentij/lunie/releases`

Extract `lunie` and put it on your `PATH`.

## Initialize Stack

```bash
lunie init \
  --database-url "postgresql://user:pass@db.example.com:5432/lunie" \
  --redis-url "redis://redis.example.com:6379"
```

This creates local config and starts the Lunie stack (server, worker).

If you want Lunie to run bundled local Postgres and Redis instead, use:

```bash
lunie init --with-local-datastores
```

Check status:

```bash
lunie status
```

Verify auth:

```bash
lunie auth whoami
```

## Run Your First Workflow

Create a minimal definition:

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
```

Create and run:

```bash
lunie workflow create --name "First Workflow" --definition /tmp/lunie-definition.json
lunie workflow list
# use workflow id from list output
lunie workflow run <workflow-id>
```

Inspect run details:

```bash
lunie run list <workflow-id>
lunie step list <workflow-id> <run-id>
```

## Optional: Configure Public Webhook Ingress

Create a webhook trigger and rotate a path key:

```bash
lunie trigger create <workflow-id> --type WEBHOOK --name "Inbound"
lunie trigger webhook rotate-key <workflow-id> <trigger-id>
```

This prints a URL in the format:

- `http://<host>:3000/v1/api/hooks/:workflowId/:triggerId/:webhookKey`

Use that URL in your webhook provider. For reverse proxy starters, see:

- `deploy/ingress/nginx.lunie.conf.example`
- `deploy/ingress/Caddyfile.example`

## Optional: Open the TUI

```bash
lunie tui
```

## Next Steps

- [CLI Usage](./CLI-Usage.md)
- [Workflow Definitions](./Lunie%20-%20Workflow%20Definitions.md)
- [TUI Guide](./Lunie%20-%20TUI.md)
- Full CLI reference: `apps/cli/README.md`
