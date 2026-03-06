---
id: getting-started
title: Getting Started
description: Install Taskforge and run your first workflow in minutes.
slug: /getting-started
---

# Getting Started

This guide is for operators and users who want to install Taskforge and run workflows.

If you want to contribute to Taskforge itself, use `docs/Development.md`.

## 10-Minute Path

If `taskforge` is already installed, this is the fastest path from zero to first run.

```bash
# 1) Initialize and start local stack
taskforge init
taskforge status

# 2) Verify auth works
taskforge auth whoami

# 3) Create a minimal workflow definition
cat > /tmp/tf-definition.json <<'JSON'
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
taskforge workflow create --name "First Workflow" --definition /tmp/tf-definition.json
taskforge workflow list
# copy workflow ID from list output
taskforge workflow run <workflow-id>

# 5) Inspect execution
taskforge run list <workflow-id>
taskforge step list <workflow-id> <run-id>
```

## What You Need

- Docker + Docker Compose
- A Taskforge CLI binary

## Install the CLI

Choose one method.

### Homebrew

```bash
brew tap gentij/taskforge
brew install taskforge
```

### AUR (Arch Linux)

```bash
yay -S taskforge-bin
```

### GitHub Release Binary

Download the matching release artifact from:

- `https://github.com/gentij/taskforge/releases`

Extract `taskforge` and put it on your `PATH`.

## Initialize Local Stack

```bash
taskforge init
```

This creates local config and starts the Taskforge stack (server, worker, Postgres, Redis).

Check status:

```bash
taskforge status
```

Verify auth:

```bash
taskforge auth whoami
```

## Run Your First Workflow

Create a minimal definition:

```bash
cat > /tmp/tf-definition.json <<'JSON'
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
taskforge workflow create --name "First Workflow" --definition /tmp/tf-definition.json
taskforge workflow list
# use workflow id from list output
taskforge workflow run <workflow-id>
```

Inspect run details:

```bash
taskforge run list <workflow-id>
taskforge step list <workflow-id> <run-id>
```

## Optional: Open the TUI

```bash
taskforge tui
```

## Next Steps

- [CLI Usage](./CLI-Usage.md)
- [Workflow Definitions](./Taskforge%20-%20Workflow%20Definitions.md)
- [TUI Guide](./Taskforge%20-%20TUI.md)
- Full CLI reference: `apps/cli/README.md`
