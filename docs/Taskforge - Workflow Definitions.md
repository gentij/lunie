---
id: workflow-definitions
title: Workflow Definitions
description: JSON schema and rules for Taskforge workflow definitions.
slug: /workflow-definitions
---

# Taskforge - Workflow Definitions

This document describes the JSON format used for Taskforge workflow definitions.

Taskforge stores a workflow definition inside a **WorkflowVersion** (immutable snapshot).

## Top-level Shape

```json
{
  "input": {},
  "notifications": [],
  "steps": []
}
```

- `input`: static defaults for workflow input.
- `notifications` (optional): run-completion notifications.
- `steps`: ordered list of step definitions.

At runtime, a run’s input is a merge of:
- `input` (workflow defaults)
- trigger input (manual/webhook payload)
- `step.input` (per-step constants)

Precedence (highest to lowest):
- `step.input`
- trigger input
- `definition.input`

## Run Notifications (MVP)

Taskforge can send workflow run completion notifications directly from a workflow definition.

Supported providers:

- `discord`
- `slack`

Supported events:

- `SUCCEEDED`
- `FAILED`

Notification shape:

```json
{
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
  ]
}
```

Rules:

- `webhook` must be an absolute `http(s)` URL or `{{secret.NAME}}`.
- Notifications are evaluated per workflow version.
- Delivery is best-effort: notification failures do not change run status.
- Discord notifications are sent as embeds; Slack notifications use text payloads.

## Step Keys (v1 Rule)

Step keys must use underscores only:

- Allowed: `^[A-Za-z0-9_]+$`
- Examples: `fetch_posts`, `build_embed`, `assert_has_posts`

This is required so JMESPath expressions can reference step outputs using `steps.fetch_posts`.

## Dependencies and Ordering

Each step can declare dependencies:

```json
{
  "key": "send",
  "dependsOn": ["build_embed"],
  "type": "http",
  "request": { "method": "POST", "url": "..." }
}
```

Taskforge also infers dependencies from `{{steps.*}}` references found in step `request` payloads.

Rules:
- Unknown step references are rejected at version creation time.
- Cycles are rejected at version creation time.

## Templates

Templates are string interpolations in the form `{{ ... }}`.

Supported namespaces:
- `{{input.KEY}}`
- `{{steps.STEP_KEY.output...}}`
- `{{secret.NAME}}`

Validation rules (strict):
- `{{input.KEY}}` must reference a key declared in `definition.input` or in the step’s `input`.
- `{{steps.STEP_KEY...}}` must reference an existing step key.
- `{{secret.NAME}}` must reference an existing secret.

### Common Examples

```json
{ "url": "{{input.apiUrl}}/posts" }
```

```json
{ "content": "First title: {{steps.fetch_posts.output.0.title}}" }
```

```json
{ "url": "{{secret.discordWebhook}}" }
```

## Step Types (v1)

### http

Performs an HTTP request.

```json
{
  "key": "fetch_posts",
  "type": "http",
  "request": {
    "method": "GET",
    "url": "{{input.apiUrl}}/posts",
    "headers": { "Accept": "application/json" },
    "timeoutMs": 30000
  }
}
```

Notes:
- Default timeout is 30s if not provided.
- Response bodies over the soft limit (256KB) fail the step.

When referencing `{{steps.fetch_posts.output...}}`, Taskforge resolves `.output` against the HTTP response body data.

### transform

Produces derived JSON using JMESPath expressions.

```json
{
  "key": "build_embed",
  "type": "transform",
  "dependsOn": ["fetch_posts"],
  "request": {
    "source": {
      "posts": "{{steps.fetch_posts.output}}"
    },
    "output": {
      "title": "Posts Summary",
      "count": { "$jmes": "length(source.posts)" },
      "firstTitle": { "$jmes": "source.posts[0].title" }
    }
  }
}
```

`$jmes` nodes are explicit objects of the form:

```json
{ "$jmes": "<expression>" }
```

JMESPath root context:
- `input`: workflow input
- `source`: `request.source`
- `steps`: step output bodies (common case)
- `stepResponses`: full step outputs (status/headers/etc when needed)

### condition

Evaluates a JMESPath boolean-ish expression and either fails or records the result.

```json
{
  "key": "assert_has_posts",
  "type": "condition",
  "dependsOn": ["fetch_posts"],
  "request": {
    "expr": "steps.fetch_posts[0] != null",
    "message": "Expected at least one post"
  }
}
```

Fields:
- `expr` (required): JMESPath expression
- `assert` (optional, default `true`): if true and expression is falsy, the step fails
- `message` (optional): appended to the failure message

Malformed/missing expressions are treated as falsy; `assert` determines whether that should fail.

## Full Examples

### Example A: Fetch + Condition + Discord

```json
{
  "input": {
    "apiUrl": "https://jsonplaceholder.typicode.com"
  },
  "notifications": [
    {
      "provider": "discord",
      "webhook": "{{secret.DISCORD_WEBHOOK_URL}}",
      "on": ["SUCCEEDED", "FAILED"]
    }
  ],
  "steps": [
    {
      "key": "fetch_posts",
      "type": "http",
      "request": {
        "method": "GET",
        "url": "{{input.apiUrl}}/posts"
      }
    },
    {
      "key": "assert_has_first_post",
      "type": "condition",
      "dependsOn": ["fetch_posts"],
      "request": {
        "expr": "steps.fetch_posts[0] != null",
        "message": "Expected at least one post"
      }
    },
    {
      "key": "send",
      "type": "http",
      "dependsOn": ["assert_has_first_post"],
      "request": {
        "method": "POST",
        "url": "{{secret.discordWebhook}}",
        "headers": {
          "Content-Type": "application/json"
        },
        "body": {
          "content": "Posts ok. First title: {{steps.fetch_posts.output.0.title}}"
        }
      }
    }
  ]
}
```

### Example B: Build a Discord Embed with transform

```json
{
  "input": {
    "apiUrl": "https://jsonplaceholder.typicode.com"
  },
  "steps": [
    {
      "key": "fetch_posts",
      "type": "http",
      "request": {
        "method": "GET",
        "url": "{{input.apiUrl}}/posts"
      }
    },
    {
      "key": "build_embed",
      "type": "transform",
      "dependsOn": ["fetch_posts"],
      "request": {
        "source": {
          "posts": "{{steps.fetch_posts.output}}"
        },
        "output": {
          "title": "Fetch Complete",
          "description": "Fetched posts",
          "fields": [
            {
              "name": "Posts",
              "value": { "$jmes": "to_string(length(source.posts))" },
              "inline": true
            },
            {
              "name": "First Title",
              "value": { "$jmes": "source.posts[0].title" },
              "inline": false
            }
          ],
          "color": 5814783
        }
      }
    },
    {
      "key": "send",
      "type": "http",
      "dependsOn": ["build_embed"],
      "request": {
        "method": "POST",
        "url": "{{secret.discordWebhook}}",
        "headers": { "Content-Type": "application/json" },
        "body": {
          "embeds": ["{{steps.build_embed.output}}"]
        }
      }
    }
  ]
}
```

## Notes on Secrets

- Prefer using `{{secret.discordWebhook}}` over putting webhooks/tokens into workflow input.
- Secrets are encrypted at rest on the server and decrypted in the worker.
- Secrets can be updated; future runs use the latest value.
