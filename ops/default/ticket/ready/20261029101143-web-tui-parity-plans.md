---
id: TKT-20261029101143
title: "Plan and ticket preparation"
status: ready
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:43Z
commits: []
ready_at: 2026-10-29T10:11:43Z
---

# Plan and ticket preparation

## Context

Slice 8 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-004, G-025.

## Scope

- **T8.1:** Route manual Generate Plan through GeneratePlanDelivery, matching the Web plan surface.
- **T8.2:** Reconcile a retained partial preparation by its original identity rather than regenerating plans or duplicate tickets.
- **T8.3:** Preserve explicit Regenerate confirmation and any independent tracking CLI capability without exposing a misleading second required step.
- **T8.4:** Offer Open document and View tickets after successful Spec/Plan generation with context-preserving navigation.
- **T8.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- One action retains the plan and creates the expected native tickets; execution has not started.
- Injected partial failure followed by retry produces exactly one intended ticket set.
- Existing-plan handling and regeneration preserve source/revision and scope checks.
- Actions appear only when their document or ticket result exists.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 7 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-8-tui-plans`
Expected PR title: `feat(slice-8): prepare TUI plans and tickets together`
Report: `ops/default/report/slices/web-tui-parity/slice-8-plans.md`
One branch and one PR to `dev`; tasks T8.1-T8.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
