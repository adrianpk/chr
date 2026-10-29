---
id: TKT-20261029101141
title: "Spec catalog and manual editing"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:41Z
commits: []
branch: feat/slice-6-tui-specs
started_at: 2026-10-29T13:25:17Z
ready_at: 2026-10-29T10:11:41Z
---

# Spec catalog and manual editing

## Context

Slice 6 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-023, G-024.

## Scope

- **T6.1:** Read the same canonical Spec catalog as Web, including manual, generated and project-wide records.
- **T6.2:** Expose shared terminal forms for title, create-time case binding, body and draft state.
- **T6.3:** Apply canonical revision checks, field validation and confirmed reload while keeping unsent text.
- **T6.4:** Retain generation, acceptance-feature generation and document inspection and prepare the same form for wizard mode.
- **T6.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Owned fixture proves all three record kinds appear, with stable functional selection.
- Create, edit and cancel use production service operations; no generated-only assumptions.
- Stale save cannot overwrite a concurrent change; case identity is not silently rebound.
- Manual/wizard presentations share the editor owner rather than copying markup or update handlers.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 5 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-6-tui-specs`
Expected PR title: `feat(slice-6): support manual TUI specifications`
Report: `ops/default/report/slices/web-tui-parity/slice-6-specs.md`
One branch and one PR to `dev`; tasks T6.1-T6.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.
