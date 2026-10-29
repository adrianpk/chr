---
id: TKT-20261029101148
title: "Wizard execution, results and recovery"
status: reviewing
kind: task
severity: medium
priority: high
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:48Z
commits:
  - ed20cd540f319ecc74809656233f01b53fe771e3
  - 8972f55df1ff14316f8c2282953e3abbebfe772a
  - 66d738e145954fec713ace69f28beb17907b3b74
  - 24112fcb7cc1f4688bc1846ce2a60925214e6730
reviewed_at: 2026-10-29T18:03:11Z
started_at: 2026-10-29T17:38:33Z
branch: feat/slice-13-tui-wizard-results
ready_at: 2026-10-29T10:11:48Z
---

# Wizard execution, results and recovery

## Context

Slice 13 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-011, G-012, G-013, G-014.

## Scope

- **T13.1:** Use the case-filtered Board for wizard execution, progress, review links and eligible lifecycle controls.
- **T13.2:** Render structured terminal results with summary, branch/PR, delivery evidence and validation commands/exit codes.
- **T13.3:** Confirm cancellation with the same saved-document/completed-change warning and show Pause only where domain policy permits it.
- **T13.4:** Keep polling read-only, preserve unsent drafts and navigation, and retain later ideas/originals as existing secondary capabilities.
- **T13.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Counts distinguish validated, confirmed and remaining work using retained receipts.
- Open PR or successful worker cannot be shown as confirmed delivery; JSON export remains available.
- Owned runtime evidence for G-013 proves confirmation and retained artifacts; no implicit rollback.
- Owned runtime evidence for G-014 proves no redirect, draft replacement or hidden progression while typing.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 12 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-13-tui-wizard-results`
Expected PR title: `feat(slice-13): complete TUI wizard delivery and recovery`
Report: `ops/default/report/slices/web-tui-parity/slice-13-wizard-results.md`
One branch and one PR to `dev`; tasks T13.1-T13.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

T13.1–T13.4 implemented with the shared case Board, verified delivery/validation
results, guarded cancellation and stable passive reads. Real Git/filesystem/SQLite
and production terminal checks pass. Cancellation retains exact committed
originals, checkpoints and completed worker changes; polling preserves a complete
unsent answer and earlier document navigation without submitting or advancing.
Task commands and evidence are recorded in the Slice 13 report. Final controller
validation, canonical PR reference and controller-owned dev integration pending.
