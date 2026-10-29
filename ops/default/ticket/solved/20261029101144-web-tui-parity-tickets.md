---
id: TKT-20261029101144
title: "Ticket editing, comments and references"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:44Z
commits: [fc8cd8e06e07568045c79e8d08b49370d5493cc9, f2ddfdfd17490910683a2a0473a9cda53b09395d, 29095befa1a96578d8eb9a1f81435f524097ce7d, ad327372411f98e7fc5854ed3de1c56ac949dbd7, 52a38b57ae4d238bad17202289bdce1d971dfd4f]
resolution: fixed
closed_at: 2026-10-29T15:43:31Z
branch: feat/slice-9-tui-tickets
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/463
reviewed_at: 2026-10-29T15:32:07Z
started_at: 2026-10-29T15:00:15Z
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

Tasks T9.1-T9.4 are implemented. Canonical metadata choices, complete forms,
separate comments/history, guarded review references and canonical source editing
have focused local legacy/native and production-terminal evidence. Checklist
content and the existing dump-append terminal journey remain covered.

The tracker and Slice 9 report retain exact commands and task commits. PR #463 is open with the immutable report introduction. Controller validation `RNRFX3HTEIWFXQJHIKQSD4HET4` passed all five ordered checks on
`be5cfd9ae8a1fa2899517feddcf1b87dc32ecea6`. Canonical Forgejo and Acta confirm PR #463 merged
into dev at that exact commit. The report and tracker retain the evidence.
