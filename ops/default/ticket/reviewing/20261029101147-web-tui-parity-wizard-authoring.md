---
id: TKT-20261029101147
title: "Wizard Quiz and preparation"
status: reviewing
kind: task
severity: medium
priority: high
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:47Z
commits:
  - 6d43124624539535e1c5a1cf3736e21e25488b01
  - fc3e0563baf91c4fcb98bd6253d3473cff7e16ff
  - d21a31eb2820f2ae2030039cef8679fcdda13c1e
  - bfc05ca37ce59466d85010e7b274fde92c0a89d0
  - 517571b3e4556afa37c502a260435cdb268d95cc
reviewed_at: 2026-10-29T17:32:11Z
started_at: 2026-10-29T17:10:35Z
branch: feat/slice-12-tui-wizard-authoring
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/466
ready_at: 2026-10-29T10:11:47Z
---

# Wizard Quiz and preparation

## Context

Slice 12 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-009, G-010.

## Scope

- **T12.1:** Use the same Quiz component and controls in manual and wizard mode, bound to the retained run and current question.
- **T12.2:** Render specification and preparation stages in their existing tabs from the shared guided operation.
- **T12.3:** Show stage-specific processing, suspended, pending-operation and blocker feedback with eligible actions.
- **T12.4:** Retry the retained operation and handle stale/newer revisions without UI-owned orchestration or duplicate preparation.
- **T12.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Deep/Fast, answer history, skip, answer, done and cancel remain functional without duplicate form logic.
- Specification auto-approval, plan/ticket preparation and committed original inputs remain domain-owned.
- Partial result and timeout keep the original operation identity and saved input visible.
- Restart/re-entry and out-of-order responses preserve exact run state and finite attempts.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 11 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-12-tui-wizard-authoring`
Expected PR title: `feat(slice-12): reuse TUI quiz and preparation in wizards`
Report: `ops/default/report/slices/web-tui-parity/slice-12-wizard-authoring.md`
One branch and one PR to `dev`; tasks T12.1-T12.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

T12.1–T12.4 implemented with shared Quiz editing/controls/history, committed
Specs/Plans rendering, stage-specific feedback and exact retained retry/reload.
Real Git/filesystem/SQLite and production terminal checks pass, including
partial preparation, deadline/lost-ack restart, duplicate prevention and stale
answer rejection. PR #466 is open. Final controller checks on the PR-reference
HEAD and integration remain pending; this ticket remains reviewing.
