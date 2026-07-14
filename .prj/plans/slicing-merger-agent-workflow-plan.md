# Merger Agent Workflow Slicing Plan

Status: planned

Active Slice: none

Spec: `.prj/specs/specs/merger-agent-workflow.md`
Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
Mirror: `.chr/default/plans/merger-agent-workflow-plan.md`

## Purpose

Implement the merger role so reviewed ticket pull requests can move from
`ready_for_merge` to `done` through a formal, auditable, objective integration
preflight.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is local to this slicing and starts at `Slice 1`.
- Each slice must update its implementation report under
  `.prj/reports/slices/merger-agent-workflow/`.
- Each task gets one conventional commit unless the tracker explicitly records
  a bundled planning/docs task.
- After each merged slice, return to updated `dev` before continuing.
- Do not run development, reviewer, and merger agents concurrently in the same
  repository workspace.

## Scope

This slicing includes:

- merger role metadata and handoff contract;
- objective merger preflight;
- merge execution against configured pull request providers;
- blocked merge handling and ticket state updates;
- user-facing and reference documentation for the merger workflow.

## Out of Scope

This slicing does not include:

- a second code review phase;
- replacing ticket review;
- replacing the separate `dev -> main` release alignment gate;
- parallel multi-agent execution;
- broad provider abstractions beyond merge readiness and execution needs.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | merger handoff contract | `feat/slice-1-merger-handoff-contract` | `feat(slice-1): merger handoff contract` | `.prj/reports/slices/merger-agent-workflow/slice-1-merger-handoff-contract.md` |
| Slice 2 | merger preflight | `feat/slice-2-merger-preflight` | `feat(slice-2): merger preflight` | `.prj/reports/slices/merger-agent-workflow/slice-2-merger-preflight.md` |
| Slice 3 | merge execution command | `feat/slice-3-merge-execution-command` | `feat(slice-3): merge execution command` | `.prj/reports/slices/merger-agent-workflow/slice-3-merge-execution-command.md` |
| Slice 4 | merger workflow docs | `feat/slice-4-merger-workflow-docs` | `feat(slice-4): merger workflow docs` | `.prj/reports/slices/merger-agent-workflow/slice-4-merger-workflow-docs.md` |

## Slice 1: Merger Handoff Contract

Goal: make the reviewer-to-merger handoff explicit in ticket metadata and
workflow state.

Behavior:

- Tickets can represent `ready_for_merge` as a merger pickup signal.
- Passed review metadata is preserved as merge authorization.
- The ticket model can record merge status, merge actor, merge timestamp, and
  merge blocker reason when needed.
- The merger role cannot reinterpret a passed review as a new review request.

Tasks:

- T1.1 Add or normalize ticket metadata for merge handoff and merge outcome.
- T1.2 Update ticket parsing, rendering, and indexing preservation behavior.
- T1.3 Add service-level handoff checks for `ready_for_merge` tickets.
- T1.4 Add tests for passed-review handoff and legacy ticket preservation.
- T1.5 Add slice report and tracker updates.

## Slice 2: Merger Preflight

Goal: verify objective merge readiness before any provider merge call.

Behavior:

- Preflight validates ticket status, review outcome, branch metadata, pull
  request metadata, expected base, PR head, provider readability, mergeability,
  serial repair blockers, and merge strategy.
- Preflight returns machine-readable blocker reasons.
- Preflight never performs code review or quality judgment.
- Operational blockers stay in merge attention state; repair blockers can route
  back to `ready_for_dev`.

Tasks:

- T2.1 Add merger preflight service boundary and result model.
- T2.2 Reuse provider diagnostics for PR readability and branch/base checks.
- T2.3 Add serial repair blocker checks.
- T2.4 Add blocker classification for operational versus repair blockers.
- T2.5 Add tests for pass, missing metadata, wrong base, stale branch, conflict,
  credentials, and serial blockers.
- T2.6 Add slice report and tracker updates.

## Slice 3: Merge Execution Command

Goal: merge reviewed pull requests through an explicit merger command or loop.

Behavior:

- A command can merge one ticket in `ready_for_merge`.
- A dry-run mode reports preflight and intended merge action without changing
  remote or ticket state.
- Successful merge updates ticket state to `done` and records merge metadata.
- Blocked merge attempts record a compact local blocker only when attention is
  needed.
- The command uses configured merge strategy and provider capabilities.

Tasks:

- T3.1 Add CLI/service command for merging a reviewed ticket.
- T3.2 Implement provider merge execution for Forgejo/Codeberg first.
- T3.3 Add GitHub merge execution support when provider config is available.
- T3.4 Apply merge result to ticket metadata and status.
- T3.5 Add tests with provider fakes for dry-run, success, provider failure,
  conflict, and blocked merge.
- T3.6 Add slice report and tracker updates.

## Slice 4: Merger Workflow Docs

Goal: document the merger workflow in the correct Diataxis locations.

Behavior:

- User guide explains how reviewed tickets are merged.
- Reference docs describe metadata, command behavior, preflight reasons, and
  state transitions.
- Explanation docs clarify role boundaries and the difference from the
  `dev -> main` gate.
- Ticket review docs link to merger handoff behavior.

Tasks:

- T4.1 Update how-to documentation for reviewed ticket merge.
- T4.2 Add or update reference documentation for merger metadata and commands.
- T4.3 Update explanation docs for `@dev`, `@reviewer`, and `@merger` roles.
- T4.4 Link ticket review workflow docs to merger handoff docs.
- T4.5 Run docs-only validation.
- T4.6 Add slice report and tracker closure updates.

## Validation Strategy

- Run focused Go tests for changed packages in implementation slices.
- Run `make verify` when shared contracts, service boundaries, or provider
  behavior change.
- For docs-only work, run whitespace/diff validation.
