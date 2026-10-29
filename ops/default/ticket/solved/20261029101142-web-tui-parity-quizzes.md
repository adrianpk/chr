---
id: TKT-20261029101142
title: "Complete Quiz interaction"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:42Z
commits:
  - 6d12928d88911fa896fba7ddb1c2dc025fe536c8
  - 552751e3b451550946d842cc36027aad39b85c50
  - 4fa89c658afee30b98e5499cd352633988ab5aec
  - 15fbe9b97c82dec1f13599edf1b4142080be992b
  - 8466be11507a83418ca18b1d5b8bbefc7c9cbfc0
  - 8958f0200940db15a1e17b19d3a20591897fc788
branch: feat/slice-7-tui-quizzes
pr: 461
resolution: fixed
closed_at: 2026-10-29T14:24:15Z
reviewing_at: 2026-10-29T14:19:02Z
started_at: 2026-10-29T13:50:00Z
ready_at: 2026-10-29T10:11:42Z
---

# Complete Quiz interaction

## Context

Slice 7 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-020, G-021, G-022.

## Scope

- **T7.1:** Make Deep/Fast selection visible before start and mark targets with pending answers.
- **T7.2:** Show answer history alongside the current question and retain answer, skip, done, cancel and resume.
- **T7.3:** Bind each submission to its exact session, current question, revision and input digest; retain draft after errors.
- **T7.4:** Replace the always-failing context-aware quiz entry with the existing supported operation and real scope input, or explicit pre-action unavailability if unsupported.
- **T7.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Mode reaches the shared service; targets are scoped to the functional case.
- Multiple rounds preserve history and do not lose the active question.
- Stale/double submission cannot answer another question; cancel/discard has confirmation.
- No confirmation invites an operation known to fail; no invented generator or new domain workflow.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 6 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-7-tui-quizzes`
Expected PR title: `feat(slice-7): complete TUI quiz controls and history`
Report: `ops/default/report/slices/web-tui-parity/slice-7-quizzes.md`
One branch and one PR to `dev`; tasks T7.1-T7.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

T7.1-T7.4 are implemented. Production adapter, owned canonical IO/HTTP references and real PTY checks prove visible modes, pending targets, retained answer history, exact captured question/revision, stale and duplicate submissions, confirmed reload, cancellation and resume. The separate contextual-authoring entry takes the approved pre-action unavailable path with no writes or failing confirmation. Task commands and correction evidence are in the Slice 7 report. PR #461 integrated at abc766b3e15ed14cd16d215ac823d1a16d25e582 after all four ordered checks passed in JRP44LHIFOTCAKMSHR3ELP542E. Canonical merge and Acta evidence are recorded in the Slice 7 report.
