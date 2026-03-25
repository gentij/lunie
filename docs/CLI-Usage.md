---
id: cli-usage
title: CLI Usage
description: Common Taskforge CLI commands for day-to-day operations.
slug: /cli
---

# CLI Usage

This page is a practical command reference for running Taskforge.

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
taskforge workflow list --sort-by updatedAt --sort-order desc
taskforge run list <workflow-id> --sort-by createdAt --sort-order asc
taskforge workflow version list <workflow-id> --sort-by version --sort-order desc
```

## Stack Commands

Run once:

```bash
taskforge init
```

Day-to-day:

```bash
taskforge start
taskforge status
taskforge logs --follow
taskforge stop
```

## Auth Commands

```bash
taskforge auth login --token "<TASKFORGE_ADMIN_TOKEN>"
taskforge auth whoami
taskforge auth status
taskforge auth logout
```

## Workflow Lifecycle

```bash
taskforge workflow create --name "My Workflow" --definition definition.json
taskforge workflow list
taskforge workflow get <workflow-id>
taskforge workflow update <workflow-id> --name "New Name"
taskforge workflow run <workflow-id> --input input.json
taskforge workflow delete <workflow-id>
```

## Run Notifications

Workflow definitions support optional `notifications` entries for run completion events.

- Providers: `discord`, `slack`
- Events: `SUCCEEDED`, `FAILED`
- `webhook` can be absolute `http(s)` or `{{secret.NAME}}`

Use `taskforge workflow create` or `taskforge workflow version create` with a definition that includes `notifications`.

## Triggers

```bash
taskforge trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
taskforge trigger list <workflow-id>
taskforge trigger get <workflow-id> <trigger-id>
taskforge trigger update <workflow-id> <trigger-id> --is-active=false
taskforge trigger delete <workflow-id> <trigger-id>
```

## Runs and Steps

```bash
taskforge run list <workflow-id>
taskforge run get <workflow-id> <run-id>
taskforge step list <workflow-id> <run-id>
taskforge step get <workflow-id> <run-id> <step-run-id>
```

## Secrets

```bash
taskforge secret create --name API_KEY --value "secret"
taskforge secret list
taskforge secret get <secret-id>
taskforge secret update <secret-id> --description "Rotated"
taskforge secret delete <secret-id>
```

## TUI

```bash
taskforge tui
```
