# Slice 4: Merger Workflow Docs

Status: delivered
Slicing: Merger Agent Workflow
Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
Branch: `feat/slice-4-merger-workflow-docs`
PR: `#286`

## Purpose

Document the reviewed ticket merge workflow so the `@reviewer -> @merger`
handoff, command behavior, metadata, and role boundary are durable product
documentation.

## Delivered Behavior

- Updated the Forgejo review how-to to include reviewed ticket merge execution.
- Added reference documentation for `chr ticket merge`, preflight, provider
  execution, successful merge metadata, blocked merge metadata, and operation
  serialization.
- Linked ticket review workflow and metadata docs to the merge workflow
  reference.
- Clarified that review provider settings also support merge execution.
- Updated the command reference and documentation index to include ticket merge.
- Extended the agent review explanation with the `@dev`, `@reviewer`, and
  `@merger` role boundary.
- Clarified `ready_for_merge` status behavior.

## Implementation Notes

The docs keep the merger role objective. They describe merge as integration
execution after passed review, not as a second review or approval ceremony.

Successful merges are documented as quiet local state transitions. Blocked
merges are documented as compact metadata on the ticket without adding Board
columns.

## Contracts Added or Changed

- Added `docs/reference/ticket-merge-workflow.md`.
- Updated review workflow references to include merge handoff.
- Updated review metadata references to include merge command validation
  expectations.
- Updated review provider configuration docs to mention merge execution.

## Files of Interest

- `docs/how-to/use-forgejo-ticket-review.md`
- `docs/index.md`
- `docs/reference/commands.md`
- `docs/reference/ticket-merge-workflow.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/ticket-review-metadata.md`
- `docs/reference/review-workflow-config.md`
- `docs/reference/ticket-status.md`
- `docs/explanation/agent-review-model.md`

## Validation

- `git diff --check`
- `rg -n "ticket-merge-workflow|ready_for_merge|@merger" docs`

## Risks and Follow-ups

- None recorded.
