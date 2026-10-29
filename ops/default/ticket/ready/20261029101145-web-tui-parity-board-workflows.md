---
id: TKT-20261029101145
title: "Responsive Board and retained workflows"
status: ready
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:45Z
commits: []
ready_at: 2026-10-29T10:11:45Z
---

# Responsive Board and retained workflows

## Context

Slice 10 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-016, G-028, G-031.

## Scope

- **T10.1:** Show all Board columns and Attention by default when width permits; retain a usable narrow view and maximize.
- **T10.2:** Add a retained workflow list/detail surface within Dash with pagination, members, evidence and current eligible actions.
- **T10.3:** Expose bounded work preview for ticket, slice, set, case and loose scope with canonical filters and ordering.
- **T10.4:** Expose allowed lifecycle controls, allowance extension and integration-decision answers, plus visible case/ticket links.
- **T10.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Wide/narrow resize and focus preserve key selection and canonical column meaning.
- The list handles empty/large sets and does not infer delivery from worker exit or an open PR.
- Selection, later-ticket inclusion, policy and confirmation match the observed application snapshot.
- Stale/revision/budget/approval guards remain host-owned; unavailable controls explain why.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 9 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-10-tui-board-workflows`
Expected PR title: `feat(slice-10): expose TUI board and retained workflows`
Report: `ops/default/report/slices/web-tui-parity/slice-10-board-workflows.md`
One branch and one PR to `dev`; tasks T10.1-T10.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
