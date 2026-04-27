---
id: queues-and-workers-plan
title: Queues and Workers Plan (Archive)
description: Planning document for Lunie queue and worker architecture.
slug: /archive/queues-and-workers-plan
---

# Lunie - Queues and Workers Plan

> Planning document.
>
> It explains the execution model rationale. For the current shipped behavior,
> use `docs/Architecture.md` and runtime docs in `README.md`.

This document outlines how Lunie should execute workflows using NestJS
BullMQ queues and a separate worker app. It is based on the NestJS queues
documentation and is aligned with the current data model (Workflow, Trigger,
Event, WorkflowRun, StepRun).

## Goals

- Durable, scalable execution using Redis-backed queues.
- Deterministic runs that reference a fixed WorkflowVersion.
- Separate worker process for isolation and operational safety.
- Minimal v1 surface area, with room to grow.

## Why BullMQ + Redis

NestJS provides `@nestjs/bullmq` which integrates BullMQ with Nest providers.
BullMQ is actively maintained and TypeScript-first. Redis is required as the
backing store for jobs and queue state. This enables:

- Decoupled producers and consumers
- Retries, backoff, and job persistence
- Horizontal scaling via additional workers

## Queue Strategy (Single Queue)

We will start with a single queue named `step-runs` and use job names to
differentiate step types.

- Queue name: `step-runs`
- Job name: step type (e.g. `http`, `transform`, `condition`)
- Job payload: references to persisted entities, never the whole definition

Example job payload (conceptual):

```
{
  "workflowRunId": "wfr_123",
  "stepRunId": "sr_456",
  "stepKey": "step_1",
  "workflowVersionId": "wfv_789",
  "input": { ... }
}
```

Notes:
- BullMQ does not support `@Process('jobName')`; job name routing must happen
  inside the processor `switch`.
- Use `jobId = stepRunId` to avoid duplicate scheduling for the same step.

## App Layout

Two Nest apps in the same repo:

1) Server app (API + orchestration)
   - Receives triggers
   - Creates Event, WorkflowRun, StepRuns
   - Enqueues step jobs to BullMQ

2) Worker app (consumers)
   - Hosts BullMQ processors
   - Executes steps and updates StepRun/WorkflowRun

These can share modules (executors, parsers, DTOs) but have separate
entrypoints and environment configs.

## Execution Flow (v1)

1) Trigger endpoint receives input
2) Event is created
3) WorkflowRun is created (status QUEUED)
4) StepRuns are created (status QUEUED)
5) Each StepRun is enqueued into `step-runs`
6) Worker processes job and updates StepRun
7) WorkflowRun status is updated when all steps complete

## Worker Processing

- Processor class: `@Processor('step-runs')`
- Extends `WorkerHost` and implements `process(job)`
- Switch on `job.name` to dispatch to executor

Pseudo flow in worker:

1) Load StepRun + WorkflowRun from DB
2) Resolve step definition from WorkflowVersion
3) Execute step handler
4) Update StepRun:
   - status
   - output or error
   - logs
   - durationMs
5) If last step succeeded, update WorkflowRun status to SUCCEEDED

## Retries and Backoff

Use BullMQ job options on enqueue:

- attempts: 3 (default)
- backoff: exponential, 5s base (example)
- removeOnComplete: keep recent N for debugging
- removeOnFail: keep recent N for debugging

StepRun should reflect actual retry counts via `attempt` and `lastErrorAt`.

## Observability

StepRun already includes:

- `error` (Json?)
- `logs` (Json?)
- `durationMs` (Int?)
- `startedAt` / `finishedAt`

Optional later:

- Queue event listeners for active/completed/failed transitions
- Artifacts table for large logs or files

## Security and Secrets

- Secrets are stored in DB and referenced by name in workflow definitions.
- At runtime, resolver loads secrets and injects them into step input.
- Avoid storing secret values in StepRun.output or logs.

## Configuration

Shared BullMQ config via `BullModule.forRootAsync()`:

- `QUEUE_HOST`
- `QUEUE_PORT`
- `QUEUE_PREFIX` (optional)

Queue registration:

- `BullModule.registerQueue({ name: 'step-runs' })`

## Rollout Plan

Phase 1
- Add BullMQ dependency
- Add QueueModule
- Create Worker app entrypoint
- Implement step executor registry

Phase 2
- Implement orchestrator
- Enqueue StepRuns
- Worker updates StepRun + WorkflowRun

Phase 3
- Add queue events and metrics
- Add artifacts for large outputs

## Open Decisions

- Which step types are in v1 (HTTP, transform, condition?)
- Default retry/backoff values
- Whether to support per-workflow concurrency limits
