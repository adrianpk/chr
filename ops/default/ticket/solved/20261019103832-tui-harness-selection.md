---
id: TKT-20261019103832
title: "TUI harness selection"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:32Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T17:19:15Z
closed_at: 2026-10-19T17:37:13Z
resolution: fixed
branch: feat/slice-6-tui-harness-selection
reviewed_at: 2026-10-19T17:37:13Z
commits: [03d0027706c678d6a6587f895f1765e590911d56, c44527f3920af192e2dd19c9eb30b95ca14253a9, 6d6489c539c7f38cab30f5565fa399c6480a807a]
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/417
---

# TUI harness selection

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 6 of 8 — TUI harness selection.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-6-tui-harness-selection`.

Expected PR title: `feat(slice-6): change project harness selections from the TUI`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-6-tui-harness-selection.md`.

Activated after canonical PR #416 was verified merged into dev. T6.1 is
the first recorded task; the canonical report accompanies its implementation.

Prerequisite: Slice 5 integrated into dev; implementation authorization for this set.

## Required Outcome

Local users can inspect and change the same per-project role defaults from the TUI without restarting the TUI or service.

Owners: Existing TUI project controls and thin application adapters.

Acceptance: HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T6.1 | Add the project-context role/profile picker with readable current defaults, readiness and separately retained active bindings; reuse the central operation service and local actor mapping. |
| T6.2 | Handle keyboard selection, clear, cancel and stale/conflicting saves without losing the visible choice; verify that Web/CLI changes become visible on the next refresh/admission and no TUI-owned configuration store exists. |

## Validation

Passed focused application/model events:
`go test ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`.

Passed focused asynchronous interaction checks with race instrumentation:
`go test -race ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`.

The real model/application journey covers both role saves and clear, local
actor authority, active versus new worker bindings, receipt replay after a lost
acknowledgement, cancellation and independent-process persistence. An owned
second process saves through real Web/application adapters; stale keyboard
input is retained and explicit refresh sees the external persisted revision.
Model fixtures cover late results, in-flight exclusion, setup failure, mouse
ownership and compact/wide layouts with scrolling.

The controller records the ordered development gate and focused checks against
the clean PR head after PR-reference follow-up commits.

## Implementation Progress

T6.1 and T6.2 are implemented. Project `h` opens the compact shared selection and
readiness projection, with retained worker bindings shown separately. Exact
revision-bound saves use the central configuration owner and local actor
mapping. Rejected input remains visible, unchanged retries recover receipts and
the running TUI sees persisted changes on reload without another settings store.

Report: [Slice 6](../../report/slices/codex-pi-harnesses/slice-6-tui-harness-selection.md).

Canonical PR: [#417](https://forge.adrianpk.com/adrianpk/chr/pulls/417). Integration remains controller-owned.

## Integration

PR #417 merged into dev at `6d6489c539c7f38cab30f5565fa399c6480a807a`.
Controller evidence `WSFV4HWI36FGR3GTGSC63VY2P3` passed all four ordered
commands recorded in the report. Acta confirms the Rebase + Fast-forward merge
in event `01M5AKRR4RQF5084R8KBTXADMC`. Slice 7 is the next recorded unit.
