---
id: TKT-20261029101136
title: "Safe queue selection and canonical moves"
status: reviewing
kind: bug
severity: high
priority: high
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:36Z
commits: [61a300444ff4814d9d2fbd61b4378577900060e4, 450d25ab66885018c476092716f9af6cc6823ed4, 3b8baf497125dd908d20011c4f0b5d66dea56d97, b97e23635b67fbb6ffa908380db394bb9c7ab899]
branch: fix/slice-1-tui-safe-actions
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/455
reviewed_at: 2026-10-29T10:31:30Z
started_at: 2026-10-29T10:17:58Z
ready_at: 2026-10-29T10:11:36Z
---

# Safe queue selection and canonical moves

## Context

Slice 1 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-026, G-034, I-001.

## Scope

- **T1.1:** Reproduce Dash/Board target divergence, legacy review movement and busy-key swallowing through production adapters.
- **T1.2:** Give Dash an independent stable selected key and bind preview/start to that key and the observed revision.
- **T1.3:** Expose the six canonical Board targets through the existing guarded ticket move operation; preserve managed-ticket and review-evidence restrictions.
- **T1.4:** Route Ctrl+C through graceful UI shutdown during busy operations, preserving retained-work ownership and cancellation semantics.
- **T1.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Distinct Board/queue keys; alias review target without evidence; active busy operation.
- Arrow movement, empty queue, reorder/removal and stale confirmation cannot run another ticket.
- Web and TUI reject identical forbidden/stale moves; successful writes persist canonical status.
- Owned terminal process exits; no implicit retained-work cancellation or goroutine leak.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Planning artifacts are committed before this first slice starts.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `fix/slice-1-tui-safe-actions`
Expected PR title: `fix(slice-1): bind TUI actions to visible targets`
Report: `ops/default/report/slices/web-tui-parity/slice-1-safe-actions.md`
One branch and one PR to `dev`; tasks T1.1-T1.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Implemented T1.1-T1.4 with reproduced defects, independent Dash target/revision binding, guarded canonical moves and graceful busy shutdown. Focused model, production adapter and PTY checks pass. PR #455 targets dev; exact final candidate checks are recorded by Looper after this PR-reference follow-up. T1.5 and ticket resolution await verified controller integration. See the canonical Slice 1 report and tracker for commands and task commits.
