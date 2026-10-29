---
id: TKT-20261029101138
title: "Shared operation feedback and draft retention"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:38Z
commits: []
branch: feat/slice-3-tui-feedback
started_at: 2026-10-29T11:17:56Z
ready_at: 2026-10-29T10:11:38Z
---

# Shared operation feedback and draft retention

## Context

Slice 3 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-038.

## Scope

- **T3.1:** Reuse neutral application outcomes and map ok, validation_failed, stale and blocked into the existing Action overlay.
- **T3.2:** Show summary, field findings, changed files and eligible reload/retry action with existing TUI styling.
- **T3.3:** Keep user input on validation/stale/error and require confirmation before destructive draft reload or discard.
- **T3.4:** Bind pending results to captured operation, selection and revision; ignore late results for a newer interaction.
- **T3.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- No TUI import of Web DTOs and no domain import of terminal types.
- Success, validation rejection, stale revision and retained pending operation remain distinguishable.
- Sentinel unsent text survives errors, focus changes and delayed responses.
- Out-of-order completion cannot overwrite another form or apply a stale confirmation.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 2 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-3-tui-feedback`
Expected PR title: `feat(slice-3): show actionable TUI operation results`
Report: `ops/default/report/slices/web-tui-parity/slice-3-feedback.md`
One branch and one PR to `dev`; tasks T3.1-T3.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

Planning only. No runtime implementation or passing acceptance is claimed. Baseline evidence and runtime uncertainties are preserved in `ops/default/report/web-tui-parity-discovery.md`. Record actual validation, commits and PR in the tracker/report during implementation.

Related existing ticket: `TKT-20261023210422` (`ops/default/ticket/open/20261023210422-standardize-web-feedback-messages.md`) owns broader Web feedback consolidation; this set reuses neutral outcomes for the TUI and does not claim that separate Web work is resolved.
