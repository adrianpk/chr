---
id: TKT-20261029101137
title: "Case-oriented shell and contextual help"
status: ready
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:37Z
commits: []
ready_at: 2026-10-29T10:11:37Z
---

# Case-oriented shell and contextual help

## Context

Slice 2 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-001, G-002, G-003, G-033, I-002.

## Scope

- **T2.1:** Align tabs and initial view by View ID: Cases, Dumps, Quizzes, Specs, Plans, Board, Dash; F1-F7 and F10 Exit.
- **T2.2:** Keep Project in header context and an alt+p action; expose real ? Help outside text modes and retain palette access.
- **T2.3:** Propagate functional case/run binding between tabs and show active-view actions in the real shell footer.
- **T2.4:** Inventory useful legacy views and retain workspace-index health, runtime information and existing commands through Project, Dash or palette.
- **T2.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Navigation, click targets and restored selection survive the reordered view list.
- No text-field key interception; project editing and Harness selection remain accessible.
- Case context survives navigation without submitting or mutating documents.
- Every baseline extra has a reachable destination; no duplicate legacy tab bar.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 1 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-2-tui-navigation`
Expected PR title: `feat(slice-2): align TUI navigation with case context`
Report: `ops/default/report/slices/web-tui-parity/slice-2-navigation.md`
One branch and one PR to `dev`; tasks T2.1-T2.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
