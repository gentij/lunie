# Taskforge CLI

The Taskforge CLI lets you manage workflows, triggers, runs, steps, and secrets from the terminal.

For local self-hosted setup, run `taskforge init` first to provision `~/.taskforge/` and local stack config.

## Quickstart

```bash
taskforge auth login --token "<TASKFORGE_ADMIN_TOKEN>"
taskforge workflow list
```

If your API is not on the default port, pass `--server`:

```bash
taskforge --server http://localhost:3100/v1/api auth whoami
```

Create a workflow from a definition file:

```bash
taskforge workflow create --name "My Workflow" --definition definition.json
```

Create a CRON trigger:

```bash
taskforge trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
```

Run a workflow and check runs/steps:

```bash
taskforge workflow run <workflow-id> --input input.json
taskforge run list <workflow-id>
taskforge step list <workflow-id> <run-id>
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
taskforge workflow list --page 1 --page-size 25 --sort-by updatedAt --sort-order desc
taskforge run list <workflow-id> --sort-by createdAt --sort-order asc
taskforge workflow version list <workflow-id> --sort-by version --sort-order desc
```

## Auth

```bash
taskforge auth login
taskforge auth status
taskforge auth whoami
taskforge auth logout
```

## Stack

`taskforge init` must be run once before using stack commands.

```bash
taskforge init
taskforge start
taskforge start server worker
taskforge start --foreground
taskforge status
taskforge logs --follow
taskforge stop
```

## Workflows

```bash
taskforge workflow list
taskforge workflow get <workflow-id>
taskforge workflow create --name "My Workflow" --definition definition.json
taskforge workflow update <workflow-id> --name "New Name"
taskforge workflow update <workflow-id> --is-active=false
taskforge workflow delete <workflow-id>
taskforge workflow run <workflow-id> --input input.json --overrides overrides.json
taskforge workflow validate <workflow-id> --definition definition.json
```

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

## Workflow Versions

```bash
taskforge workflow version list <workflow-id>
taskforge workflow version get <workflow-id> <version>
taskforge workflow version create <workflow-id> --definition definition.json
```

## Triggers

```bash
taskforge trigger list <workflow-id>
taskforge trigger get <workflow-id> <trigger-id>
taskforge trigger create <workflow-id> --type CRON --name "Nightly" --config cron.json
taskforge trigger update <workflow-id> <trigger-id> --is-active=false
taskforge trigger delete <workflow-id> <trigger-id>
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
taskforge run list <workflow-id>
taskforge run get <workflow-id> <run-id>
```

## Steps

```bash
taskforge step list <workflow-id> <run-id>
taskforge step get <workflow-id> <run-id> <step-run-id>
```

## Secrets

```bash
taskforge secret list
taskforge secret get <secret-id>
taskforge secret create --name API_KEY --value "my-secret"
taskforge secret update <secret-id> --description "Rotated"
taskforge secret delete <secret-id>
```

Secrets are not printed in table output. Use `--output json` if you need raw JSON.

## Output Modes

- `--output table` shows human-readable tables (default)
- `--output json` prints raw JSON for scripting
- `--quiet` prints only IDs

## Troubleshooting

- **Token not set**: run `taskforge auth login`
- **Validation errors**: verify JSON files match the server schema (e.g., CRON uses `cron`, not `expression`)
