---
id: TKT-20261029101146
title: "Wizard entry and Dump in shared tabs"
status: ready
kind: task
severity: medium
priority: high
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:46Z
commits: []
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

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
