---
id: initial-project-overview
title: Initial Project Overview (Archive)
description: Early planning snapshot from initial Lunie design phase.
slug: /archive/initial-project-overview
---

# TaskForge — Self-Hosted Workflow & Automation Engine

> Historical planning document.
>
> Some implementation details below are outdated.
> For current behavior and setup, use:
>
> - `README.md`
> - `docs/Development.md`
> - `docs/Architecture.md`
> - `apps/cli/README.md`

## 1. Project Overview

**TaskForge** is a **self-hosted workflow and automation engine** designed for engineers, developers, and technical users who prefer **local-first, infrastructure-controlled tools**.

It allows users to define workflows made of **triggers** and **actions**, execute them reliably, inspect their runs, and automate real-world tasks without relying on third-party SaaS platforms.

TaskForge is:
- API-first
- CLI-driven
- Self-hosted
- Automation-focused
- No GUI required (CLI/TUI friendly)

---

## 2. Motivation & Philosophy

### Why TaskForge exists
- Most automation platforms:
  - are SaaS-first
  - require accounts, teams, billing, UI-heavy workflows
  - are overkill or undesirable for personal or infra-level automation
- Engineers often want:
  - automation they fully control
  - something that runs locally or on their own server
  - predictable behavior
  - inspectable logs and state
  - no dependency on external services

### Core principles
- **Local-first**: works on a single machine or private server
- **Self-hosted by default**
- **One user is enough** (no network effects)
- **Automation > UI**
- **Reliability over convenience**
- **Explicit behavior** (no magic)

---

## 3. Target Audience

- Software engineers
- Backend developers
- DevOps / infra-minded users
- Power users who automate workflows
- People running personal servers / homelabs

Not targeted at:
- Non-technical users
- Growth-driven SaaS audiences
- Content-based platforms

---

## 4. High-Level Architecture

TaskForge consists of **two main components**:

### 1) TaskForge Server
- Written in **TypeScript**
- Built with **NestJS + Fastify**
- Exposes an HTTP API
- Responsible for:
  - workflow definitions
  - triggers
  - execution orchestration
  - state persistence
  - retries and reliability
  - logging and inspection

### 2) TaskForge CLI
- Written in **Go**
- Installed as a **single binary**
- Acts as:
  - installer
  - initializer
  - control surface
- Communicates with the server via HTTP API

The CLI owns:
- initialization
- configuration
- starting/stopping the stack
- issuing commands
- TUI integration

---

## 5. Distribution & Installation Model

### Installation (example on Arch Linux)
```bash
yay -S lunie-cli-bin
```

This installs:
- the `lunie` CLI binary
- deployment templates (Docker Compose)

### Initialization
```bash
lunie init
```

`lunie init`:
- creates `~/.lunie/`
- generates a config file
- initializes secrets / admin token
- starts the server stack via Docker Compose
- is **idempotent** (won’t overwrite existing config)

### Runtime model
- Server runs inside Docker containers:
  - TaskForge server
  - Postgres
  - Redis
- CLI communicates with server via localhost API
- No auto-start-on-boot in v1 (can be added later)

---

## 6. Repository Structure (Monorepo)

```
lunie/
  docs/
  contracts/
  deploy/
  apps/
    server/    (NestJS + Fastify)
    cli/       (Go)
  packages/
  scripts/
```

### Key shared assets
- **OpenAPI spec** (API contract)
- **JSON schemas** (workflow definitions, config format)
- **Docker Compose templates**
- **Documentation**

Runtime code is **not shared**, only contracts.

---

## 7. Authentication & Security Model (v1)

### Self-hosted security philosophy
- No Google OAuth
- No user signup
- No email verification

### v1 authentication
- Single **admin API token**
- Generated during `lunie init`
- Stored in `~/.lunie/config`
- All API calls require `Authorization: Bearer <token>`

---

## 8. Core Concepts

### Workflow
A workflow defines **what happens**, composed of:
- one trigger
- one or more steps (actions)

### Trigger types (v1)
- Manual (CLI / API)
- Webhook
- Cron (scheduled)

### Step / Action types (v1)
- HTTP request
- JSON transformation
- Conditional branching

### Execution model
- Each trigger creates a **workflow run**
- Each step creates a **step run**
- Workers pick up step runs via **leases**
- State is persisted at every transition
- System is crash-safe and resumable

---

## 9. Reliability & Execution Guarantees

TaskForge aims for:
- **At-least-once execution**
- Durable state persistence
- Retry with backoff
- Dead-letter handling for failed steps
- Idempotency support (later)

Key mechanisms:
- database-backed job queues
- step leasing with expiration
- explicit step states:
  - QUEUED
  - RUNNING
  - SUCCEEDED
  - FAILED

---

## 10. Goals for Version 1 (V1)

### Functional goals
- Define workflows declaratively
- Trigger workflows via:
  - CLI
  - HTTP webhook
  - cron schedule
- Execute HTTP-based automation
- Inspect runs and logs
- Manage everything via CLI

### Non-goals for v1
- GUI
- SaaS hosting
- Team collaboration
- OAuth integrations
- Shell execution
- Multi-user permissions

---

## 11. Post-V1 / Future Ideas

### Execution
- Shell actions (unsafe mode / allowlist)
- Script containers
- Fan-out / fan-in steps
- Parallel execution

### Auth & Access
- Multiple API tokens
- Scoped tokens
- Multi-user support (optional)
- OIDC / Google login (optional)

### UX
- Interactive TUI
- Log tailing
- Run visualization
- Workflow templates

### Integrations
- GitHub
- Slack / Discord
- Generic OAuth connectors
- Webhook signature verification

### Deployment
- systemd user service
- remote server mode
- external DB support

---

## 12. Current Status

- This section reflects an early planning snapshot and is not the live status.
- For current status and release readiness, use:
  - `README.md`
  - `docs/MVP-Release-Readiness.md`
