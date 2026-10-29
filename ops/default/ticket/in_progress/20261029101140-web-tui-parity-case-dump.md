---
id: TKT-20261029101140
title: "Case and Dump forms"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: tui, parity, web-tui-parity
source: review
reported_at: 2026-10-29T10:11:40Z
commits:
  - 8a5691065226ff3b8499290bc7ec39ce36145073
  - d6f933abf9b328a3befc4837bd5744733108c4cc
  - cf8c5e4b3d68ffb2d61349075f1cd6d7507a9b55
  - fbd0b6d5a6dc7be995713fbbd8cbbb059ae9e6f2
branch: feat/slice-5-tui-case-dump
started_at: 2026-10-29T12:44:00Z
ready_at: 2026-10-29T10:11:40Z
---

# Case and Dump forms

## Context

Slice 5 of 14 in the maintainer-requested Web and TUI parity delivery. Source findings: G-015, G-017, G-018, G-019.

## Scope

- **T5.1:** Create cases through Name, Description and Context before one Create; retain existing metadata editing and unresolved references.
- **T5.2:** Create and edit dumps through scope, title and body using existing canonical operations.
- **T5.3:** Expose AppendDump with free text and make external-editor shortcuts and displayed hints agree in textarea mode.
- **T5.4:** Exercise stale-save, confirmed reload/discard and unsent draft retention with owned Web and TUI fixtures.
- **T5.5:** Complete focused validation, `make verify`, the slice report, PR review and controller-owned dev integration; close this ticket only after the merge is verified.

## Acceptance criteria

- Cancel creates no blank case; entered fields survive validation failure.
- Project-wide and case-bound dumps preserve actual user content without fixed placeholder text.
- The reachable append action uses the selected dump; editor round trips preserve edits.
- Runtime disposition and evidence for G-018 are recorded; no lost draft or silent overwrite.
- Required checks pass on the exact candidate; a report links actual commands/results and the integration commit.

## Dependencies and boundary

Starts after Slice 4 is integrated; the tracker enforces the recorded sequence.
Use existing application/domain owners and terminal styling. No new persistence model, dependency, UI-specific domain engine, main promotion or unrelated remediation. A discovered need outside this scope is a blocker/proposal, not implicit authority.

## Delivery

Plan: `ops/default/plan/web-tui-parity.md`
Tracker: `ops/default/tracker/web-tui-parity.md`
Specification: `ops/default/spec/web-tui-parity.md`
Branch: `feat/slice-5-tui-case-dump`
Expected PR title: `feat(slice-5): add complete TUI case and dump forms`
Report: `ops/default/report/slices/web-tui-parity/slice-5-case-dump.md`
One branch and one PR to `dev`; tasks T5.1-T5.5 are the canonical work units. Looper owns the bounded sequence and conditional Rebase + Fast-forward merges.

## Evidence

T5.1-T5.4 are implemented and focused checks pass. Owned production terminals prove complete creation, cancellation without canonical writes, append and unsaved external-editor round trips. A real concurrent writer reproduces a stale Save; cancellation retains the draft, while confirmed reload replaces it before a fresh Save. Web HTTP reference fixtures pass. The Slice 5 report records commands and G-018 runtime disposition; final exact-HEAD controller validation and dev integration remain pending.
