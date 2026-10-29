---
id: TKT-20261029101139
title: "Shared queue observation and stable refresh"
status: ready
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:39Z
commits: []
ready_at: 2026-10-29T10:11:39Z
---

# Shared queue observation and stable refresh

## Context

Slice 4 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-032, G-035, G-036, G-037.

## Scope

- **T4.1:** Project ObserveWork into a terminal-neutral queue, counts, readiness reasons, exclusions and pagination.
- **T4.2:** Keep Dash queue selection and Board selection separately by key across refresh, sorting and removal.
- **T4.3:** Render current work, recent activity and results in Dash; preserve filter, order, Run all ready and running-work details.
- **T4.4:** Expose Preview Work and Workflow details with exact scope/policy confirmation and cancellation.
- **T4.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Compare the same application observation across adapters, including ineligible and empty pages.
- Selection follows the same key or a deterministic visible fallback without starting work.
- Existing queue extras remain usable; large/empty event collections are bounded.
- Changed snapshot or scope invalidates confirmation; no execution from passive refresh.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 3 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-4-tui-dash`
Expected PR title: `feat(slice-4): share TUI queue readiness and activity`
Report: `ops/default/report/slices/web-tui-parity/slice-4-dash.md`
One branch and one PR to `dev`; tasks T4.1-T4.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
