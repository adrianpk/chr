# Ticket Review Integration Slicing Plan

Status: planned

Active Slice: none

Spec Draft: `.prj/specs/drafts/ticket-review-integration-draft.md`
Quiz: `.prj/quizzes/curated/ticket-review-integration-quiz.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Mirror: `.chr/default/plans/ticket-review-integration-plan.md`

## Purpose

Implement ticket-level review integration for chr so agent-produced
ticket work can move through branch, pull request, review, repair, and merge
states without relying on informal chat memory or ad hoc remote inspection.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is local to this slicing and starts at `Slice 1`.
- Each slice must update its implementation report under
  `.prj/reports/slices/ticket-review-integration/`.
- Each task gets one conventional commit unless the tracker explicitly records
  a bundled planning/docs task.
- After each merged slice, return to updated `dev` before continuing.
- Do not run development and reviewer agents concurrently in the same
  repository workspace.

## Scope

This slicing includes:

- ticket metadata for branches, pull request review references, review state,
  and review rounds;
- workflow commands that create or attach pull request references and preserve
  ticket metadata;
- provider configuration and diagnostics for Forgejo/Codeberg first and GitHub
  second;
- remote pull request review inspection and comment writing;
- agentic review command behavior and anti-noise policy;
- review rejection state transitions and repair handoff;
- serialized work selection preflight for slice/ticket series;
- documentation and QA for the complete review workflow.

## Out of Scope

This slicing does not include:

- replacing the separate dev-to-main promotion gate;
- parallel multi-agent execution in one repository workspace;
- a full formal user identity model for `@dev`, `@reviewer`, or `@user`;
- local-only substitute review when remote pull request access is unavailable
  in `agentic_review` mode;
- automatic trusted auto-merge beyond the policy/config plumbing needed to
  represent the mode safely.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | ticket review metadata contract | `feat/slice-1-ticket-review-metadata-contract` | `feat(slice-1): ticket review metadata contract` | `.prj/reports/slices/ticket-review-integration/slice-1-ticket-review-metadata-contract.md` |
| Slice 2 | pr reference workflow | `feat/slice-2-pr-reference-workflow` | `feat(slice-2): pr reference workflow` | `.prj/reports/slices/ticket-review-integration/slice-2-pr-reference-workflow.md` |
| Slice 3 | review mode configuration | `feat/slice-3-review-mode-configuration` | `feat(slice-3): review mode configuration` | `.prj/reports/slices/ticket-review-integration/slice-3-review-mode-configuration.md` |
| Slice 4 | forge provider diagnostics | `feat/slice-4-forge-provider-diagnostics` | `feat(slice-4): forge provider diagnostics` | `.prj/reports/slices/ticket-review-integration/slice-4-forge-provider-diagnostics.md` |
| Slice 5 | remote review adapter | `feat/slice-5-remote-review-adapter` | `feat(slice-5): remote review adapter` | `.prj/reports/slices/ticket-review-integration/slice-5-remote-review-adapter.md` |
| Slice 6 | ticket review command | `feat/slice-6-ticket-review-command` | `feat(slice-6): ticket review command` | `.prj/reports/slices/ticket-review-integration/slice-6-ticket-review-command.md` |
| Slice 7 | serialized work preflight | `feat/slice-7-serialized-work-preflight` | `feat(slice-7): serialized work preflight` | `.prj/reports/slices/ticket-review-integration/slice-7-serialized-work-preflight.md` |
| Slice 8 | repair handoff and docs | `feat/slice-8-repair-handoff-docs` | `feat(slice-8): repair handoff docs` | `.prj/reports/slices/ticket-review-integration/slice-8-repair-handoff-docs.md` |

## Slice 1: Ticket Review Metadata Contract

Goal: make branch, pull request reference, review state, and review round data
first-class ticket metadata.

Behavior:

- Tickets can carry `branch`, `review_refs`, `review_state`, `review_round`,
  `review_result`, `review_comments_url`, and blocking finding references.
- Ticket parsing, rendering, indexing, and TUI/CLI display preserve the new
  fields.
- Board columns remain stable; review state appears as metadata, badge,
  attention reason, or detail text instead of a new column.
- Legacy tickets without these fields continue to parse.

Tasks:

- T1.1 Add ticket metadata fields and validation rules.
- T1.2 Update ticket parser and renderer preservation behavior.
- T1.3 Update ticket index/read models and board attention display.
- T1.4 Add tests for metadata parsing, rendering, indexing, and legacy tickets.
- T1.5 Add reference documentation for the ticket review metadata contract.
- T1.6 Add slice report and tracker updates.

## Slice 2: PR Reference Workflow

Goal: ensure branch and pull request references are written to the ticket by
workflow commands instead of relying on manual frontmatter edits.

Behavior:

- A command can attach or update a ticket pull request reference.
- A command can mark that a ticket branch exists and belongs to the ticket.
- `ready_for_review` validation fails when PR-backed work lacks branch or pull
  request metadata.
- The workflow supports a transitional single `pr_url` only if structured
  `review_refs` are not yet available in all paths.

Tasks:

- T2.1 Add command/service support for attaching branch and pull request refs.
- T2.2 Validate `ready_for_review` requirements for branch and PR metadata.
- T2.3 Preserve existing PR refs during execution result application.
- T2.4 Add CLI/TUI operation paths for branch and PR metadata updates.
- T2.5 Add tests for valid refs, missing refs, and repeated updates.
- T2.6 Add slice report and tracker updates.

## Slice 3: Review Mode Configuration

Goal: make review behavior explicit and configurable without changing the
dev-to-main promotion gate.

Behavior:

- Project config supports `agentic_review`, `human_review`, and
  `trusted_auto_merge` review modes.
- The configured mode applies to ticket or slice integration into `dev`.
- Missing remote access in `agentic_review` mode is an evidence failure.
- Clean agentic review passes do not write ceremonial remote comments.
- Non-blocking remote comments have low configured limits and never block merge.

Tasks:

- T3.1 Add review workflow config schema and defaults.
- T3.2 Add config loader and validation for review modes.
- T3.3 Add dry-run/status output that explains the active review mode.
- T3.4 Add tests for default, human, agentic, and trusted modes.
- T3.5 Document review mode configuration.
- T3.6 Add slice report and tracker updates.

## Slice 4: Forge Provider Diagnostics

Goal: verify provider tooling and credentials before agentic review tries to
read or write pull request state.

Behavior:

- Forgejo/Codeberg are first-class provider targets.
- GitHub is supported for ubiquity without shaping the Forgejo model.
- Diagnostics verify provider mapping from PR URL, required CLI availability
  when configured, token/auth state, and basic PR readability.
- SSH transport is treated as Git access evidence, not API review capability.
- Missing provider capability reports machine-readable reasons.

Tasks:

- T4.1 Add provider configuration model for Forgejo/Codeberg and GitHub.
- T4.2 Add provider detection from pull request URLs.
- T4.3 Add diagnostics for `tea`, Forgejo-compatible tokens, and PR readability.
- T4.4 Add diagnostics for `gh`, GitHub tokens/auth, and PR readability.
- T4.5 Add CLI output and tests for pass, missing tooling, missing token, and
  unsupported provider cases.
- T4.6 Add slice report and tracker updates.

## Slice 5: Remote Review Adapter

Goal: provide a provider-neutral remote review interface backed first by
Forgejo/Codeberg and then GitHub behavior.

Behavior:

- The adapter can read pull request metadata, diff, comments, review state, and
  checks where the provider exposes them.
- The adapter can write inline comments when line anchoring is reliable.
- The adapter can fall back to a general numbered findings comment when inline
  placement is unavailable.
- Clean passes do not write remote comments.
- Blocking and non-blocking comments are distinguishable.

Tasks:

- T5.1 Define remote review adapter contracts and result types.
- T5.2 Implement Forgejo/Codeberg read paths for PR metadata, diff, comments,
  and checks available through the chosen API path.
- T5.3 Implement Forgejo/Codeberg comment writing with inline-first behavior.
- T5.4 Implement GitHub read and comment paths.
- T5.5 Add tests with provider fixtures or fakes for read, inline comment,
  general comment, and unavailable capability cases.
- T5.6 Add slice report and tracker updates.

## Slice 6: Ticket Review Command

Goal: run an inspection-only ticket review over `ready_for_review` work.

Behavior:

- The command selects or accepts a ticket in `ready_for_review`.
- The reviewer inspects ticket metadata, execution evidence, validation logs,
  pull request metadata, diff, comments, and checks.
- The command does not start Docker, launch a Codex worker, reimplement the
  ticket, or rerun expensive validation in normal mode.
- Findings follow high-signal policy and anti-nit limits.
- Blocking findings update local ticket review metadata and write remote
  comments.
- Clean passes update local metadata without remote comment ceremony.

Tasks:

- T6.1 Add ticket review command/service boundary.
- T6.2 Add evidence assembler for ticket, execution, validation, and PR data.
- T6.3 Add reviewer prompt/policy for inspection-only, high-signal review.
- T6.4 Add result application for pass, needs changes, insufficient evidence,
  and human decision cases.
- T6.5 Add tests for clean pass, blocking findings, non-blocking findings,
  missing remote access, and dry-run.
- T6.6 Add slice report and tracker updates.

## Slice 7: Serialized Work Preflight

Goal: ensure development work selection cannot skip unresolved reviews or start
later slices from stale `dev`.

Behavior:

- All agent work selection passes through scheduler preflight.
- One active agent operation can hold the repository workspace execution lock.
- The preflight detects `changes_requested` tickets and prioritizes repair.
- The preflight understands series/sequence/dependency metadata.
- The next slice is ineligible until the prior slice is merged to `dev`, the
  merge commit is contained in local updated `dev`, and the worktree is clean.
- Block reasons are machine-readable.

Tasks:

- T7.1 Add workspace-level agent operation lock for review/development paths.
- T7.2 Add series/dependency metadata to work selection inputs.
- T7.3 Add preflight checks for unresolved review repairs.
- T7.4 Add preflight checks for prerequisite PR merged and contained in updated
  `dev`.
- T7.5 Add scheduler behavior to select repair work before later series work.
- T7.6 Add tests for blocked next slice, repair priority, stale dev, and lock
  conflicts.
- T7.7 Add slice report and tracker updates.

## Slice 8: Repair Handoff and Docs

Goal: connect review rejection to repair execution and document the final
workflow.

Behavior:

- `changes_requested` moves a ticket back to the development queue without
  notifying the human in `agentic_review` mode.
- Repair uses the existing branch and pull request.
- The worker context points to the remote PR comments as the source of detailed
  findings instead of duplicating them in ticket comments.
- If the repair agent cannot read the remote comments, the workflow stops with
  a clear evidence/access failure.
- Reference docs and user guide describe review modes, metadata, provider
  setup, serialized execution, and repair behavior.

Tasks:

- T8.1 Add repair handoff context for existing branch and pull request.
- T8.2 Add behavior for remote comment access failure during repair.
- T8.3 Add tests for repair context, existing branch reuse, and no next-slice
  selection while repair is pending.
- T8.4 Add reference docs for ticket review integration.
- T8.5 Add how-to docs for Forgejo-first setup and review operation.
- T8.6 Run `make verify` and a ticket review QA scenario.
- T8.7 Add final cycle report and tracker closure.

## Open Implementation Notes

- The first Forgejo implementation must choose between direct HTTP and `tea`
  based on reliable support for PR metadata, comments, and inline review.
- The first structured ticket representation should prefer `review_refs`; use
  `pr_url` only as a transitional compatibility path if necessary.
- A formal user identity model is intentionally deferred.
- The reviewer prompt must be inspection-only and anti-nit by default.
- Remote comments are written only for findings; clean passes stay quiet.
