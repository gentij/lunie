# Taskforge TUI Guide

This document covers the current Taskforge TUI user workflow and developer architecture.

It is intended for:

- Operators who run `taskforge tui` day to day
- Contributors extending screens, actions, and themes

## What the TUI Is

The TUI is a keyboard-first interface over the Taskforge API, built with Bubble Tea.

It provides:

- Live snapshot browsing for workflows, runs, triggers, events, secrets, and tokens
- In-place actions (run/toggle/archive/update/create) through keybindings and modals
- Context tabs for Overview, JSON, Steps, and Logs
- A command palette for navigation, actions, themes, and network simulation
- A run inspector for deep log and step exploration

## Launching the TUI

From the CLI:

```bash
taskforge tui
```

The command lives in `apps/cli/internal/cli/tui.go` and starts the Bubble Tea program in alt-screen mode.

### Config and Auth

The TUI uses the same CLI config as other commands.

Config shape (`apps/cli/internal/config/config.go`):

- `serverUrl`
- `token`
- `profile` (optional)
- `theme` (optional)

Default config path is OS-specific (via `os.UserConfigDir`), typically under `taskforge/config.json`.

## Layout and Focus Model

Primary layout:

- Left sidebar (navigation)
- Main panel (table + dashboard/list content)
- Context drawer (bottom panel, collapsible)
- Footer (hints + paging)

Focus panes:

- `sidebar`
- `main`
- `context`

Focus controls:

- `tab` next pane
- `shift+tab` previous pane
- `left` / `right` directional pane jump

Context drawer:

- `ctrl+j` toggles collapsed/expanded

## Global Keybindings

Defined in `apps/cli/internal/tui/app/keymap.go`.

Core:

- `q`, `ctrl+c` quit
- `?` toggle help view
- `ctrl+k` open command palette
- `/` table search mode
- `ctrl+f` context search mode
- `ctrl+r` retry refresh (when stale/error)
- `g` jump top
- `G` jump bottom
- `s` cycle sort column
- `S` toggle sort direction

Pane controls:

- `tab` next pane
- `shift+tab` prev pane
- `ctrl+j` toggle context drawer

Context tab controls (when context pane focused):

- `[` / `h` previous tab
- `]` / `l` next tab
- `1` Overview
- `2` JSON
- `3` Steps
- `4` Logs

Main/context scrolling:

- Main panel: `alt+up`, `alt+down`, `pgup`, `pgdown`, `home`, `end`
- Context panel: `j`, `k`, `pgup`, `pgdown`, `ctrl+u`, `ctrl+d`, `home`, `end`

## Screen Reference

Screen data and column definitions are in `apps/cli/internal/tui/screens/screens.go`.

### Dashboard

Shows summary cards and recent runs table.

Columns:

- Run ID
- Workflow
- Status
- Started

Actions:

- Mostly read-only; use palette for navigation and global actions

### Workflows

Columns:

- Name
- Active
- Latest Version
- Triggers
- Last Run
- Updated

Actions:

- `r` run selected workflow
- `e` toggle active/archive state
- `n` rename selected workflow
- `c` create trigger (for selected workflow)
- `d` archive workflow (confirmation modal)
- `f` cycle status scope (`all -> active -> inactive`)

### Runs

Columns:

- Run ID
- Workflow
- Status
- Trigger
- Started
- Duration

Actions:

- `enter` open run inspector

### Triggers

Columns:

- Name
- Type
- Workflow
- Active
- Created

Actions:

- `e` toggle active/archive state
- `n` update selected trigger
- `c` create trigger
- `d` archive trigger (confirmation modal)
- `f` cycle status scope (`all -> active -> inactive`)

### Events

Columns:

- Event ID
- Trigger
- Type
- Received
- Linked Run

Actions:

- Read-only view in current implementation

### Secrets

Columns:

- Name
- Description
- Created

Actions:

- `c` create secret
- `n` update selected secret
- `d` hard-delete selected secret (typed confirmation phrase)

Security behavior:

- Secret values are never shown in context
- Value input is masked in modals

### API Tokens

Columns:

- Name
- Scopes
- Created
- Last Used
- Status

Actions:

- `d` currently shows an unavailable warning (revoke action not implemented in TUI yet)

## Context Panel and Tabs

Context content generation is in `apps/cli/internal/tui/screens/context.go`.

Tabs:

- Overview: entity summary/details
- JSON: structured payloads and config
- Steps: step timeline/details for runs
- Logs: step logs for runs

Notes:

- Steps/Logs tabs are meaningful for `Runs` and `Dashboard`-selected runs
- Other views show an unsupported message for those tabs

## Modals and Mutation Flows

Modal openers and behavior:

- Openers: `apps/cli/internal/tui/app/model_modals.go`
- Validation/helpers: `apps/cli/internal/tui/app/model_modal_helpers.go`
- Modal update/submit: `apps/cli/internal/tui/app/model_modal_update.go`

Mutation commands:

- `apps/cli/internal/tui/app/model_mutations.go`

Destructive confirmation semantics:

- Workflow archive: phrase `ARCHIVE <workflow-id>`
- Trigger archive: phrase `ARCHIVE <trigger-id>`
- Secret delete: phrase `DELETE <secret-id>`

Archive behavior:

- Workflow/trigger "delete" in TUI is archive semantics (inactive), aligned with current UX copy

## Command Palette

Open with `ctrl+k`.

Palette features:

- Navigate screens
- Execute actions for selected row
- Toggle auto-refresh
- Clear filters
- Theme switching
- Network profile switching
- CLI handoff shortcuts for workflow authoring
- Recent command history (top entries)

Palette disables commands that do not apply in current context and shows reason text.

Implementation: `apps/cli/internal/tui/app/model_palette.go`.

## Run Inspector

Open from `Runs` screen with `enter`.

Implementation: `apps/cli/internal/tui/app/inspector.go`.

Behavior:

- Left column: steps list
- Right column: logs viewport
- `tab` switches focus between columns
- `w` toggles wrap mode
- `/` starts log search
- `esc` exits inspector

## Sorting, Filtering, and Status Scope

Filtering and table state logic:

- `apps/cli/internal/tui/app/model_view_state.go`
- `apps/cli/internal/tui/app/model_table.go`
- `apps/cli/internal/tui/screens/screens.go`

Capabilities:

- Full-table text filtering via `/`
- Per-view sort column and sort direction memory
- Status scope cycling for Workflows and Triggers only

## Refresh and Network Profiles

Runtime behavior is in `apps/cli/internal/tui/app/model_runtime.go`.

Profiles:

- Fast
- Normal
- Slow
- Flaky

Effects:

- API fetch delay simulation
- Auto-refresh interval
- Failure simulation for flaky profile

Surface states:

- loading
- refreshing
- success
- empty
- error
- stale

## Theme System

Theme registry and style mapping:

- `apps/cli/internal/tui/styles/theme.go`

Available themes:

- taskforge
- dracula
- one-dark-pro
- rose-pine-moon
- solarized-dark
- solarized-light
- nord
- gruvbox-dark
- tokyo-night
- catppuccin
- fallout
- retro-amber

Theme can be changed:

- From command palette (`ctrl+k` -> Themes)
- Persisted to config as `theme`

## Known Limitations

- API token revoke flow is not wired in TUI actions yet.
- Worker count/workspace indicators are currently static placeholders in the UI model.
- TUI is keyboard-first; mouse-specific interactions are not a primary path.

## Troubleshooting

### API appears offline

- Verify server URL and token in config.
- Use `ctrl+r` to retry from stale/error states.

### Context text readability issues (theme-specific)

- Switch theme from palette to compare.
- If needed, set a safer theme in config (`taskforge`, `nord`, `one-dark-pro`).

### Weird ANSI artifacts in table rows

- Ensure you are running the latest built CLI binary.
- Rebuild:

```bash
cd apps/cli
go build -o ../../taskforge ./cmd/taskforge
```

## Developer Architecture

TUI packages:

- `apps/cli/internal/tui/app`: state machine, input routing, modals, mutations, rendering orchestration
- `apps/cli/internal/tui/screens`: row builders, sorting semantics, context builders, lookup helpers
- `apps/cli/internal/tui/components`: reusable table/modal/cards rendering primitives
- `apps/cli/internal/tui/styles`: theme tokens and style mapping
- `apps/cli/internal/tui/layout`: responsive layout computation
- `apps/cli/internal/tui/data`: normalized in-memory data model
- `apps/cli/internal/tui/utils`: text helpers (wrap, filter, JSON formatting)

Data loading pipeline:

- `apps/cli/internal/tui/app/data_source.go` fetches paginated API data, normalizes to `data.Store`, and emits snapshot messages

## Contributing to the TUI

### Add a new theme

1. Add token set in `styles.ThemeRegistry` (`apps/cli/internal/tui/styles/theme.go`)
2. Add palette labels and entries in `apps/cli/internal/tui/app/model_palette.go`
3. Optionally add a font hint in `apps/cli/internal/tui/app/view_chrome.go`
4. Verify table/context readability in both selected and non-selected states

### Add a new screen

1. Add `ViewID` and registration in `apps/cli/internal/tui/app/model.go`
2. Implement rows/sort/context in `apps/cli/internal/tui/screens`
3. Wire any screen-specific actions in `apps/cli/internal/tui/app/model_input.go`
4. Add tests and update this doc

### Add a new mutation action

1. Add opener/modal wiring in `model_modals.go` and modal update/validation helpers
2. Add API mutation command in `model_mutations.go`
3. Add palette action entry if relevant
4. Ensure success/error toast + refresh behavior is correct

## Build and Test (CLI)

```bash
cd apps/cli
go test ./...
go build -o ../../taskforge ./cmd/taskforge
```
