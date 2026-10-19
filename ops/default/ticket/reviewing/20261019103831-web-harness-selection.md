---
id: TKT-20261019103831
title: "Web harness selection"
status: reviewing
kind: task
severity: medium
priority: normal
scope: ui
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:31Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T16:50:50Z
reviewed_at: 2026-10-19T17:15:15Z
branch: feat/slice-5-web-harness-selection
commits: [57aca35d38f424236a93b34602f215bed4b5198e, 8df1d8911dca3858896dbe9862052aecbdf3c147]
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/416
---

# Web harness selection

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 5 of 8 — Web harness selection.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-5-web-harness-selection`.

Expected PR title: `feat(slice-5): change project harness selections from the Web UI`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-5-web-harness-selection.md`.

Activated after canonical PR #415 was verified merged into dev.
T5.1 introduced the canonical report in its implementation commit.

Prerequisite: Slice 4 integrated into dev; implementation authorization for this set.

## Required Outcome

An authenticated user can inspect and change both role defaults in a compact project-context control, with clear effects on new work.

Owners: Existing Web projections/actions/templates and the shared application selection operation.

Acceptance: HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T5.1 | Add the compact project-context profile chooser using shared readable selection/readiness projections; show role defaults separately from active invocation bindings, support clearing a role and state that changes apply to new operations. |
| T5.2 | Submit exact revision-bound changes through existing authenticated mutation/origin guards; preserve input on rejection and verify login, stale updates, project scope and the absence of a service restart or credential/path editor. |

## Validation

Passed focused HTTP/application checks:
`go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`.

Passed the owned browser journey and the same focused regressions:
`CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`.

The journey covers login, both role changes while a claim remains active,
original versus new invocation bindings, independent-process persistence,
lost acknowledgement/replay, stale input, cancel and 1280/390 px layouts.
HTTP tests also cover project/confirmation/origin guards, closed bounded fields,
unavailable profiles, clear, private-setup failure and full-page fallback.
The final ordered development gate and focused checks are recorded by the
controller against the clean PR head after PR-reference follow-up commits.

## Implementation Progress

T5.1 and T5.2 are implemented. The compact project-context chooser observes shared
selection/readiness and retained worker bindings. Exact revision-bound saves use
the existing authenticated action path and shared configuration owner. Rejected
input remains selected, receipt retries are idempotent, and new operations see
persisted defaults without restarting the service.

Report: [Slice 5](../../report/slices/codex-pi-harnesses/slice-5-web-harness-selection.md).

Canonical PR: [#416](https://forge.adrianpk.com/adrianpk/chr/pulls/416). Integration remains controller-owned.
