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
- `--quiet`: print command-ready refs only
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
lunie run list my-workflow --sort-by createdAt --sort-order asc
lunie workflow version list my-workflow --sort-by version --sort-order desc
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
lunie workflow get my-workflow
lunie workflow update my-workflow --name "New Name"
lunie workflow run my-workflow --input input.json
lunie workflow delete my-workflow
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
lunie trigger create my-workflow --type CRON --name "Nightly" --config cron.json
lunie trigger create my-workflow --type WEBHOOK --name "Inbound"
lunie trigger webhook rotate-key my-workflow inbound
lunie trigger list my-workflow
lunie trigger get my-workflow nightly
lunie trigger update my-workflow nightly --is-active=false
lunie trigger delete my-workflow nightly
```

Public webhook ingress path format:

- `POST /v1/api/hooks/:workflowKey/:triggerKey/:webhookKey`

After rotating a webhook key, call the generated URL directly from your webhook provider.

## Runs and Steps

```bash
lunie run list my-workflow
lunie run get my-workflow 42
lunie step list my-workflow 42
lunie step get my-workflow 42 fetch_post
```

## Secrets

```bash
lunie secret create --name API_KEY --value "secret"
lunie secret list
lunie secret get API_KEY
lunie secret update API_KEY --description "Rotated"
lunie secret delete API_KEY
```

## TUI

```bash
lunie tui
```
