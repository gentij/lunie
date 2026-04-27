# Lunie CLI

The Lunie CLI lets you manage workflows, triggers, runs, steps, and secrets from the terminal.

For self-hosted setup, run `lunie init` first to provision `~/.lunie/` and stack config.

## Quickstart

```bash
lunie auth login --token "<LUNIE_ADMIN_TOKEN>"
lunie workflow list
```

If your API is not on the default port, pass `--server`:

```bash
lunie --server http://localhost:3100/v1/api auth whoami
```

Create a workflow from a definition file:

```bash
lunie workflow create --name "My Workflow" --definition definition.json
```

Create a CRON trigger:

```bash
lunie trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
```

Run a workflow and check runs/steps:

```bash
lunie workflow run <workflow-id> --input input.json
lunie run list <workflow-id>
lunie step list <workflow-id> <run-id>
```

## Global Flags

- `--output table|json` (default: `table`)
- `--quiet` (print IDs only)
- `--no-color` (disable colored status output)
- `--server` (API base URL, default: `http://localhost:3000/v1/api`)
- `--config` (config file path)

`NO_COLOR=1` disables color output as well.

## Pagination and Sorting

List commands support pagination and server-side sorting.

Common flags:

- `--page` (default: `1`)
- `--page-size` (default: `25`)
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
lunie workflow list --page 1 --page-size 25 --sort-by updatedAt --sort-order desc
lunie run list <workflow-id> --sort-by createdAt --sort-order asc
lunie workflow version list <workflow-id> --sort-by version --sort-order desc
```

## Auth

```bash
lunie auth login
lunie auth status
lunie auth whoami
lunie auth logout
```

## Stack

`lunie init` must be run once before using stack commands.

By default, init expects external Postgres and Redis URLs:

```bash
lunie init \
  --database-url "postgresql://user:pass@db.example.com:5432/lunie" \
  --redis-url "redis://redis.example.com:6379"
```

To run bundled local Postgres/Redis containers instead:

```bash
lunie init --with-local-datastores
```

```bash
lunie start
lunie start server worker
lunie start --foreground
lunie status
lunie logs --follow
lunie stop
```

## Workflows

```bash
lunie workflow list
lunie workflow get <workflow-id>
lunie workflow create --name "My Workflow" --definition definition.json
lunie workflow update <workflow-id> --name "New Name"
lunie workflow update <workflow-id> --is-active=false
lunie workflow delete <workflow-id>
lunie workflow run <workflow-id> --input input.json --overrides overrides.json
lunie workflow validate <workflow-id> --definition definition.json
```

Notes:

- `--input` values override colliding keys from workflow definition `input`.
- `--overrides` applies only to `http` steps and supports request `query`/`body` overrides keyed by step key.

Sample `definition.json`:

```json
{
  "input": {
    "apiBase": "https://jsonplaceholder.typicode.com",
    "postId": 1
  },
  "steps": [
    {
      "key": "fetch_post",
      "type": "http",
      "request": {
        "method": "GET",
        "url": "{{input.apiBase}}/posts/{{input.postId}}"
      }
    }
  ]
}
```

Sample with run notifications:

```json
{
  "input": {
    "shouldPass": true
  },
  "notifications": [
    {
      "provider": "discord",
      "webhook": "{{secret.DISCORD_WEBHOOK_URL}}",
      "on": ["FAILED"]
    },
    {
      "provider": "slack",
      "webhook": "{{secret.SLACK_WEBHOOK_URL}}",
      "on": ["SUCCEEDED", "FAILED"]
    }
  ],
  "steps": [
    {
      "key": "gate",
      "type": "condition",
      "request": {
        "expr": "input.shouldPass",
        "assert": true
      }
    }
  ]
}
```

Notes:

- `webhook` accepts absolute `http(s)` URLs or `{{secret.NAME}}` references.
- Notifications are evaluated on final run status (`SUCCEEDED`/`FAILED`).
- Delivery is best-effort and does not change workflow run status.

## Workflow Versions

```bash
lunie workflow version list <workflow-id>
lunie workflow version get <workflow-id> <version>
lunie workflow version create <workflow-id> --definition definition.json
```

## Triggers

```bash
lunie trigger list <workflow-id>
lunie trigger get <workflow-id> <trigger-id>
lunie trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
lunie trigger create <workflow-id> --type WEBHOOK --name "Inbound"
lunie trigger webhook rotate-key <workflow-id> <trigger-id>
lunie trigger update <workflow-id> <trigger-id> --is-active=false
lunie trigger delete <workflow-id> <trigger-id>
```

Sample `cron.json`:

```json
{
  "cron": "0 2 * * *",
  "timezone": "UTC"
}
```

## Runs

```bash
lunie run list <workflow-id>
lunie run get <workflow-id> <run-id>
```

## Steps

```bash
lunie step list <workflow-id> <run-id>
lunie step get <workflow-id> <run-id> <step-run-id>
```

## Secrets

```bash
lunie secret list
lunie secret get <secret-id>
lunie secret create --name API_KEY --value "my-secret"
lunie secret update <secret-id> --description "Rotated"
lunie secret delete <secret-id>
```

Secrets are not printed in table output. Use `--output json` if you need raw JSON.

## Output Modes

- `--output table` shows human-readable tables (default)
- `--output json` prints raw JSON for scripting
- `--quiet` prints only IDs

## Troubleshooting

- **Token not set**: run `lunie auth login`
- **Validation errors**: verify JSON files match the server schema (e.g., CRON uses `cron`, not `expression`)
