---
id: TKT-20261019103828
title: "Live role selection and retained bindings"
status: ready
kind: task
severity: medium
priority: normal
scope: persistence
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:28Z
ready_at: 2026-10-19T10:38:33Z
commits: []
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

These are planned identities; no implementation branch, PR or report exists
merely because this ticket is ready.

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
