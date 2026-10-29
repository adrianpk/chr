---
id: TKT-20261029101146
title: "Wizard entry and Dump in shared tabs"
status: reviewing
kind: task
severity: medium
priority: high
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:46Z
commits:
  - d0516dc70ac86ac9491a1ff5076650d40d7c9297
  - a1ed8105573e40b269362b2559e902cf895e465e
  - 7524663d82fafeb04658281cc8e14319e977f977
  - 06d019e5d3117982539da0c14977ddd0c10d3e1c
branch: feat/slice-11-tui-wizard-entry
reviewed_at: 2026-10-29T16:58:28Z
started_at: 2026-10-29T16:27:19Z
ready_at: 2026-10-29T10:11:46Z
---

# Wizard entry and Dump in shared tabs

## Context

Slice 11 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-005, G-006, G-007, G-008.

## Scope

- **T11.1:** Expose Wizard and Open workflow in Cases and start with one Name/Create through StartNamedGuidedUseCase.
- **T11.2:** Replace the detached wizard presentation with case/run binding in the existing shell, stage indicator and shared forms.
- **T11.3:** Use the existing Dump form with a single Continue to quiz submission of the current content.
- **T11.4:** Keep non-current wizard tabs read-only with Continue to the current stage; isolate manual and wizard controls.
- **T11.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- No random-ID input ceremony or second name step; existing case/run re-entry is deterministic.
- Stages map to existing tabs; no parallel menu, duplicated editor or raw namespace/UUID prompt.
- Runtime G-008 evidence proves one exact retained input and retry reconciliation without duplicates.
- Manual save cannot accidentally break wizard progression; viewing another tab never submits work.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 10 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-11-tui-wizard-entry`
Expected PR title: `feat(slice-11): start TUI wizards inside shared tabs`
Report: `ops/default/report/slices/web-tui-parity/slice-11-wizard-entry.md`
One branch and one PR to `dev`; tasks T11.1-T11.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

T11.1–T11.4 implemented with named entry, case-bound re-entry, normal tabs,
shared forms and Dump body editor, exact-content Continue/retry and read-only
mode guards. Focused tests cover real persistence, committed original bytes,
keyboard/palette boundaries and a production managed terminal journey. The report
records successful commands. Final clean-HEAD validation and integration remain
pending; this ticket remains reviewing.
