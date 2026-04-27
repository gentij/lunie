---
id: cli-usage
title: CLI Usage
description: Common Lunie CLI commands for day-to-day operations.
slug: /cli
---

# CLI Usage

This page is a practical command reference for running Lunie.

For full command coverage, see `apps/cli/README.md`.

## Global Flags

- `--server`: API base URL (default `http://localhost:3000/v1/api`)
- `--output`: `table` or `json`
- `--quiet`: print IDs only
- `--no-color`: disable colored output
- `--config`: config file path

## Pagination and Sorting

List commands support pagination and server-side sorting.

Common list flags:

- `--page`
- `--page-size`
- `--sort-by`
- `--sort-order` (`asc|desc`)

Supported `--sort-by` values:

- `workflow list`: `createdAt|updatedAt`
- `workflow version list`: `version|createdAt`
- `trigger list`: `createdAt|updatedAt`
- `run list`: `createdAt|updatedAt`
- `step list`: `createdAt|updatedAt`
- `secret list`: `createdAt|updatedAt`

Examples:

```bash
lunie workflow list --sort-by updatedAt --sort-order desc
lunie run list <workflow-id> --sort-by createdAt --sort-order asc
lunie workflow version list <workflow-id> --sort-by version --sort-order desc
```

## Stack Commands

Run once (external datastores):

```bash
lunie init \
  --database-url "postgresql://user:pass@db.example.com:5432/lunie" \
  --redis-url "redis://redis.example.com:6379"
```

Local bundled Postgres/Redis (optional):

```bash
lunie init --with-local-datastores
```

Day-to-day:

```bash
lunie start
lunie status
lunie logs --follow
lunie stop
```

## Auth Commands

```bash
lunie auth login --token "<LUNIE_ADMIN_TOKEN>"
lunie auth whoami
lunie auth status
lunie auth logout
```

## Workflow Lifecycle

```bash
lunie workflow create --name "My Workflow" --definition definition.json
lunie workflow list
lunie workflow get <workflow-id>
lunie workflow update <workflow-id> --name "New Name"
lunie workflow run <workflow-id> --input input.json
lunie workflow delete <workflow-id>
```

Notes:

- `--input` values override colliding keys from workflow definition `input`.
- `--overrides` is for per-step HTTP request overrides (`query`/`body`) keyed by step key.

## Run Notifications

Workflow definitions support optional `notifications` entries for run completion events.

- Providers: `discord`, `slack`
- Events: `SUCCEEDED`, `FAILED`
- `webhook` can be absolute `http(s)` or `{{secret.NAME}}`

Use `lunie workflow create` or `lunie workflow version create` with a definition that includes `notifications`.

## Triggers

```bash
lunie trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
lunie trigger create <workflow-id> --type WEBHOOK --name "Inbound"
lunie trigger webhook rotate-key <workflow-id> <trigger-id>
lunie trigger list <workflow-id>
lunie trigger get <workflow-id> <trigger-id>
lunie trigger update <workflow-id> <trigger-id> --is-active=false
lunie trigger delete <workflow-id> <trigger-id>
```

Public webhook ingress path format:

- `POST /v1/api/hooks/:workflowId/:triggerId/:webhookKey`

After rotating a webhook key, call the generated URL directly from your webhook provider.

## Runs and Steps

```bash
lunie run list <workflow-id>
lunie run get <workflow-id> <run-id>
lunie step list <workflow-id> <run-id>
lunie step get <workflow-id> <run-id> <step-run-id>
```

## Secrets

```bash
lunie secret create --name API_KEY --value "secret"
lunie secret list
lunie secret get <secret-id>
lunie secret update <secret-id> --description "Rotated"
lunie secret delete <secret-id>
```

## TUI

```bash
lunie tui
```
