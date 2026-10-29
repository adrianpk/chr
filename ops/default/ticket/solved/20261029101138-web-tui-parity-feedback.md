---
id: TKT-20261029101138
title: "Shared operation feedback and draft retention"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:38Z
commits: [eb484dd34676fa4fa8f518642d797e83b7ef21a1, 8046584d910969a0be349bb85f9a7307e3f86d8a, bc8e124dadb420312d6d19bff536e588741941f3, eb2a245aa88f1291ec339ef42fbfa4394de2e139]
branch: feat/slice-3-tui-feedback
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/457
resolution: fixed
closed_at: 2026-10-29T12:08:30Z
reviewed_at: 2026-10-29T11:58:43Z
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

PR [#457](https://forge.adrianpk.com/adrianpk/chr/pulls/457) implements neutral operation outcomes, Action feedback, confirmed draft replacement and captured reply ownership. Focused component, service, Web and production PTY checks passed during implementation. The Slice 3 report and tracker record the evidence; controller validation 7IDEK3RU26UI7CUQSTNJ7T3NAK passed all four ordered checks on c91ba4feac5b4296548a854e26b8d9cb59170ed7, and canonical Forgejo/Acta confirm integration at that exact commit.

Related existing ticket: `TKT-20261023210422` (`ops/default/ticket/open/20261023210422-standardize-web-feedback-messages.md`) owns broader Web feedback consolidation; this set reuses neutral outcomes for the TUI and does not claim that separate Web work is resolved.
