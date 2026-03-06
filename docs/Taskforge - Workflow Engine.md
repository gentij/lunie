---
id: workflow-engine-mental-model
title: Workflow Engine Mental Model
description: Core entities and execution model of the Taskforge workflow engine.
slug: /workflow-engine
---

# Taskforge – Workflow Engine Mental Model

This document describes the mental model, core entities, and relationships
behind Taskforge’s workflow automation engine.

Taskforge is a **developer-first, self-hosted workflow automation system**.
It is designed for technical users who want deterministic, inspectable,
and reproducible automations.

---

## Core Principles

- **Self-hosted**: the operator controls infrastructure and credentials
- **Token-based auth**: no users/emails/logins
- **Immutable execution history**: past runs never change meaning
- **Observable**: every trigger, run, and step is inspectable
- **Deterministic**: executions always reference a fixed definition

---

## High-level Mental Model

A workflow in Taskforge is a *program*.

- A **Workflow** is the container (the “project”)
- A **WorkflowVersion** is a frozen snapshot of logic
- A **Trigger** starts executions
- An **Event** represents something that happened in the outside world
- A **WorkflowRun** is one execution instance
- A **StepRun** is one executed step inside a run

```
Event → Trigger → Workflow → WorkflowVersion → WorkflowRun → StepRuns
```

---

## Entity Overview

### Workflow

**What it is**  
A named automation that groups versions, triggers, and runs.

**Real-world example**  
“Deploy API on GitHub Release”

**Responsibilities**
- Holds metadata (name, active flag)
- Acts as a stable identity across versions
- Owns triggers and execution history

---

### WorkflowVersion

**What it is**  
An immutable snapshot of a workflow’s definition at a point in time.

**Real-world example**
- v1: build + deploy  
- v2: build + migrate + deploy  

**Responsibilities**
- Stores the workflow definition (steps, ordering, config)
- Guarantees reproducibility of past executions

**Key rule**
> A run ALWAYS points to a specific version.

---

### Trigger

**What it is**  
A configured entrypoint that starts workflow executions.

**Trigger types**
- MANUAL
- WEBHOOK
- CRON

**Real-world examples**
- “On GitHub release published”
- “Every day at 03:00”
- “Run now”

---

### Event

**What it is**  
A persisted record of something that happened externally.

**Why it exists**
- Debuggability
- Reliability
- Deduplication

---

### WorkflowRun

**What it is**  
One execution instance of a workflow.

**Responsibilities**
- Track lifecycle
- Store normalized input/output
- Own step execution history

---

### StepRun

**What it is**  
Execution record of a single step inside a workflow run.

**Responsibilities**
- Track status and retries
- Store step-level output
- Enable deep observability

---

## Execution Flow Example

1. Webhook received
2. Event stored
3. Trigger matched
4. WorkflowRun created
5. WorkflowVersion locked
6. StepRuns executed
7. Final status recorded

---

## Design Decisions

- Linear execution first
- Immutable versions
- No users, only tokens

---

## Future Extensions

- Secrets management
- DAG execution
- Artifacts & logs
- Concurrency limits
- Retry policies

---

## Summary

Taskforge treats automation as **code with history**.
Everything is explicit, inspectable, and reproducible.
