---
id: TKT-20261029101144
title: "Ticket editing, comments and references"
status: ready
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:44Z
commits: []
ready_at: 2026-10-29T10:11:44Z
---

# Ticket editing, comments and references

## Context

Slice 9 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-027, G-029, G-030, I-004.

## Scope

- **T9.1:** Share neutral canonical ticket metadata options and read-only rules with the existing application owner; map Web and TUI separately.
- **T9.2:** Create and edit tickets with user title, description, case and metadata; show phase/slice and read-only explanations.
- **T9.3:** Expose free-text comment input and retained history instead of the fixed action comment.
- **T9.4:** Expose AttachReviewRef with user input and preserve external editing, checklist display and canonical review guards.
- **T9.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- No TUI dependency on completeTicketEditor Web types; options and field restrictions agree.
- No fixed TUI-created body/area; revision and managed-ticket restrictions remain effective.
- Round trip multiple distinct comments and reject empty/stale submissions without losing input.
- Review reference is validated and attached to the selected ticket; append wiring from Slice 5 remains reachable.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 8 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-9-tui-tickets`
Expected PR title: `feat(slice-9): add complete TUI ticket interactions`
Report: `ops/default/report/slices/web-tui-parity/slice-9-tickets.md`
One branch and one PR to `dev`; tasks T9.1-T9.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
