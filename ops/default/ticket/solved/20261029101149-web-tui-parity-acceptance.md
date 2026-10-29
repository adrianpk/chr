---
id: TKT-20261029101149
title: "Parity acceptance and capability inventory"
status: solved
resolution: fixed
closed_at: 2026-10-29T18:49:48Z
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:49Z
commits:
  - 6c7920ddce29df8a6ede341c84cc81e7294d3d74
  - 80b1dcdcf47c65a58e53747a1561f6d30ac04816
  - e941712f0c977722a961ff4f68991927a371dff5
  - c2a07937d9c2a9ac23b2de155989df450a072b8c
  - 45ef4cb4f8156d67c59f67c3f1b948488450705d
  - d94d5af3dedad35b430f6fc3952b5efce88c75c4
reviewed_at: 2026-10-29T18:37:55Z
started_at: 2026-10-29T18:12:24Z
branch: test/slice-14-tui-acceptance
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/468
ready_at: 2026-10-29T10:11:49Z
---

# Parity acceptance and capability inventory

## Context

Slice 14 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: I-003.

## Scope

- **T14.1:** Reconcile all 38 gap rows and four incidental findings with exact tests and dispositions; close runtime uncertainties without changing source claims into assumed results.
- **T14.2:** Reconcile the palette and legacy capability inventory; gate unsupported safe checklist writers before confirmation without inventing a new domain or filesystem writer.
- **T14.3:** Extend owned production PTY and Web reference fixtures to exercise manual forms and the complete guided journey, including focus, resize, cancel and restart.
- **T14.4:** Update the User Guide and consolidate the delivered user-visible changelog outcomes; run final focused acceptance and publish the cycle report.
- **T14.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Every coverage row links a delivered slice/report and passing evidence or an explicit maintainer-approved exception.
- No advertised unconditional failure remains; useful baseline extras retain tested access.
- Real Bubble Tea process plus real HTTP/Chromium, filesystem, Git and SQLite; deterministic model/provider responses only.
- No claims based on screenshots or reducer-only tests; no broad suite, new dependency or unapproved live project mutation.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 13 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `test/slice-14-tui-acceptance`
Expected PR title: `test(slice-14): verify complete Web and TUI workflow parity`
Report: `ops/default/report/slices/web-tui-parity/slice-14-acceptance.md`
One branch and one PR to `dev`; tasks T14.1-T14.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

All 42 source IDs are reconciled without changing baseline assertions. Checklist
writers stop before confirmation. The named acceptance family reuses owned
production boundaries; focused fresh full wizard PTY, manual resize/input and
Chromium polling/resize/cancellation checks pass. Exact commands, runtime
uncertainties and retained capabilities are recorded in the Slice 14 report.
PR #468 merged into dev at `d94d5af3dedad35b430f6fc3952b5efce88c75c4`,
verified through canonical Forgejo state and Acta merge event
`01M64FPHBRCSXTN5Q8YG2K1ZS6`. Controller validation
`4OVYDKWV4PUPDT57DT7IE3UMT2` passed all five ordered checks on that exact
clean PR-reference head, including nine acceptance roots and 14 named parity
scenarios without skips. The delivered Slice 14 report records exact commands,
results and fixture boundaries. The controller's separate final delivery-set
`make verify` passed on `c5ae18e761d7c65e77784c04143197c8bba77462` in validation
`AX3QL4I5ZYL2KD6M76ZBPYNHNN`; the cycle report records the exact evidence and
documentation-only closure boundary.
