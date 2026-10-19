---
id: TKT-20261019103828
title: "Live role selection and retained bindings"
status: in_progress
kind: task
severity: medium
priority: normal
scope: persistence
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:28Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T11:23:17Z
branch: feat/slice-2-harness-selection
commits: [d9b19ac1b2ce4e1b577d9497443b2d207c4c1ce7]
---

# Live role selection and retained bindings

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 2 of 8 — Live role selection and retained bindings.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-2-harness-selection`.

Expected PR title: `feat(slice-2): retain live harness selections and invocation bindings`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-2-harness-selection.md`.

The assigned implementation branch is active. No Slice 2 PR is ready yet;
the report will be introduced when the complete slice is ready for integration.

Prerequisite: Slice 1 integrated into dev; implementation authorization for this set.

## Required Outcome

A shared service and CLI select prepared profiles per role without restart, while admitted work retains its binding across processes, retries and recovery.

Owners: Existing workspace configuration and invocation owners in core; the central application operation service and CLI.

Acceptance: HP-02, HP-04, HP-12, HP-14, HP-15, HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T2.1 | Resolve private named profiles and workspace role choices with the documented legacy precedence; return readable readiness and effective selections, preserve omitted/direct-API behavior and reject invalid or unsupported explicit choices. |
| T2.2 | Implement shared inspect/change operations and CLI status/select/clear adapters with authorized actors, whole-document revision comparison, cross-process exclusion, atomic two-role updates and bounded request-id replay; preserve unrelated configuration and unknown fields. |
| T2.3 | Add chr-invocation/v2 binding and validation before worker admission, preserving byte-identical v1 checkpoints and their historical command semantics; retain selected profile, model, runtime/executable and revision evidence for all pending tickets and retries. |
| T2.4 | Freeze generation selection once per authoring/conversation operation and thread it through nested calls, later stages and retries; fresh operations read persisted settings, active operations never rebind, and existing capacity/budget authorities stay shared. |

## Validation

Exercise two independent processes, simultaneous admissions and switches, stale/unauthorized/invalid updates, clear-to-legacy, crash/replay, unchanged unrelated fields, active pending tickets and multi-call generation. An old v1-only reader must reject v2 before spawning.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/core ./internal/app -run '^(TestHarnessSelection|TestHarnessBinding|TestHarnessGenerationBinding)' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.

## Implementation Evidence

T2.1 is implemented: the configuration owner parses strict worker/generation
profile references and exposes its complete persisted revision. Private named
profiles have bounded parsing and readable, path-free projections. Resolution
preserves each legacy action default and direct OpenAI selection, observes fresh
workspace choices without restart, and rejects unknown or unsupported explicit
choices without fallback. Inspection does not launch the model executable.

Local readiness distinguishes executable availability, managed or explicit
credentials, runtime support and remote authentication, which is not tested by
inspection. Pi dispatch and profile resources without an implemented dispatch
path are reported unsupported rather than accepted and ignored.

Successful T2.1 checks:

- `go test ./internal/core ./internal/app -run '^TestHarnessSelection' -count=1 -timeout=3m`
- `go test ./internal/core -run '^(TestLoadWorkspaceConfig|TestWorkflowConfig)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `git diff --check`

T2.2 is implemented: shared inspection reads both roles and their full-document
revision together, lists private named choices through path-free readiness
projections, and keeps invalid explicit choices visible. `chr harness status`,
`chr harness select` and `chr harness clear` delegate to the shared operation
service. Local mutations derive the configured operator rather than accepting
an arbitrary actor flag. Other trusted adapters must supply their current
actor/project authorization check; authority is checked again inside exclusion,
including receipt replay.

The existing configuration owner serializes role updates and workflow upgrades
through one file lock. Each accepted change compares the entire persisted
configuration revision inside that lock, validates changed profiles and local
readiness, preserves unrelated fields/comments, and atomically publishes both
roles with project, actor, UTC timestamp and request-digest metadata. Clearing
one role retains the other and restores the cleared role's legacy dispatch.
The named profile `null` remains distinct from an unquoted null selection.

A bounded configuration journal retains choices and digests rather than copying
unrelated configuration. Recovery recognizes only the exact prior or accepted
image, completes the original result across each publication boundary, and
rejects conflicting edits or changed arguments for an interrupted request ID.
Receipts are limited to 16 KiB each and 4,096 accepted identities; exhaustion
rejects new changes while preserving prior replay. A replay returns the original
snapshot even after later changes or profile unavailability. No workflow record
store, interface settings store, scheduler or capacity pool was added.

Successful T2.2 checks:

- `go test ./internal/core ./internal/app -run '^(TestHarnessSelection|TestWorkflowUpgrade|TestLoadWorkspaceConfig|TestWorkflowConfig)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestHarnessSelection' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `git diff --check`

These focused tests include independent writer processes, atomic two-role
updates, stale/unauthorized/unprepared changes, unknown-field preservation,
receipt capacity, interrupted publication/replay, an upgrade fence, fresh service
inspection and the actual root CLI. The owned subprocess fixtures have stopped.

Next recorded task: T2.3, invocation v2 bindings and historical v1 compatibility.
T2.4 operation-level generation bindings also remains pending. Production
admission and generation caller wiring are not complete. No Slice 2 PR is ready;
the configured final exact-HEAD checks have not run for this partial slice.
