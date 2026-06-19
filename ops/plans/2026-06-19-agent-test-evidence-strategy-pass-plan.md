# Agent Test/Evidence Strategy Pass Plan

Status: planned

Active Slice: none

Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`

Spec: `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

Quiz: `ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz.md`

## Purpose

Add a mandatory evidence-first strategy stage to ticket development so a worker
decides, persists, follows, and reconciles testing evidence before a ticket can
move to review.

The delivery should make this behavior operational without adding a second
Codex container by default. The first version uses an inline strategy pass in
the development worker, persisted in the ticket checklist/tasklist, with strict
review-transition enforcement and documented sane defaults.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/agent-test-evidence-strategy-pass/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before follow-up work.
- Do not start worker-prompt or review-gate enforcement before the checklist
  parser/writer foundation exists.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | evidence strategy config | `feat/slice-1-evidence-strategy-config` | `feat(slice-1): evidence strategy config` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-1-evidence-strategy-config.md` |
| Slice 2 | evidence checklist writer | `feat/slice-2-evidence-checklist-writer` | `feat(slice-2): evidence checklist writer` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-2-evidence-checklist-writer.md` |
| Slice 3 | worker strategy context | `feat/slice-3-worker-strategy-context` | `feat(slice-3): worker strategy context` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-3-worker-strategy-context.md` |
| Slice 4 | review evidence gate | `feat/slice-4-review-evidence-gate` | `feat(slice-4): review evidence gate` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-4-review-evidence-gate.md` |
| Slice 5 | gherkin test policy | `feat/slice-5-gherkin-test-policy` | `feat(slice-5): gherkin test policy` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-5-gherkin-test-policy.md` |
| Slice 6 | evidence visibility docs | `feat/slice-6-evidence-visibility-docs` | `docs(slice-6): evidence visibility docs` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-6-evidence-visibility-docs.md` |

## Slice 1: Evidence Strategy Config

Goal: introduce the configuration and core model for evidence strategy without
changing ticket execution behavior.

Behavior:

- Workspace config loads an `evidence.strategy` section with sane defaults.
- Core defines controlled values for strategy mode, checklist requirement,
  Gherkin policy, strict review gating, user edits, and execution report copy.
- Invalid evidence config produces classified configuration errors.
- Existing work, runtime, review, and gate configuration behavior is unchanged.

Tasks:

- T1.1 Add evidence strategy config structs, defaults, and parser support under
  `.chc/user/config.yml`.
- T1.2 Add enum validation for strategy mode, checklist policy, Gherkin policy,
  strict gating, user edits, and report copy behavior.
- T1.3 Add core strategy status and diagnostic value types used by later slices.
- T1.4 Add config tests for omitted defaults, explicit overrides, and invalid
  values.
- T1.5 Document the config contract in reference docs.
- T1.6 Add the slice implementation report and validation evidence.

## Slice 2: Evidence Checklist Writer

Goal: add safe parse/render/update support for evidence-strategy checklist
blocks in ticket Markdown.

Behavior:

- Core can parse an anchored `chc:evidence-strategy:v1` checklist block.
- Core can insert a missing strategy block into a ticket without damaging
  frontmatter, comments, blockers, or existing body content.
- Core can update known checklist labels while preserving user-authored values
  that are still valid.
- Malformed strategy blocks produce actionable diagnostics instead of silent
  overwrites.
- Existing Board checklist preview remains compatible.

Tasks:

- T2.1 Add parser coverage for strategy block anchors, required labels, checked
  state, status values, and malformed blocks.
- T2.2 Add a renderer for the v1 strategy checklist block.
- T2.3 Add a safe ticket Markdown updater that inserts or updates only the
  evidence strategy block.
- T2.4 Wire a core/app service path for evidence checklist write operations.
- T2.5 Add CLI-level tests or service tests that prove user content is
  preserved.
- T2.6 Add the slice implementation report and validation evidence.

## Slice 3: Worker Strategy Context

Goal: make ticket execution inject the strategy contract into the Codex worker
context and copy strategy evidence into execution artifacts.

Behavior:

- Worker context instructs Codex to create or update the ticket checklist before
  modifying project files.
- Worker context includes the current evidence config and any existing strategy
  checklist.
- Worker context includes the change classification, Gherkin, tests,
  validation, and evidence-summary obligations from the spec.
- Execution reports duplicate the initial strategy and final evidence summary
  when `execution_report_copy` is enabled.
- The default remains one inline worker run; no separate strategy container is
  launched.

Tasks:

- T3.1 Extend context materialization with an Evidence Strategy section.
- T3.2 Include current evidence config and parsed checklist state in the worker
  context.
- T3.3 Update the worker result contract to require evidence summary metadata.
- T3.4 Copy strategy and evidence summary into runtime reports when enabled.
- T3.5 Add context and result-contract tests for docs-only and behavior-change
  tickets.
- T3.6 Add the slice implementation report and validation evidence.

## Slice 4: Review Evidence Gate

Goal: block `ready_for_review` transitions when required evidence is missing,
malformed, blocked, or failing.

Behavior:

- Ticket result application inspects evidence strategy state before moving a
  ticket to `ready_for_review`.
- Direct ticket status changes to `ready_for_review` use the same evidence
  gate when strict mode is enabled.
- Runtime infrastructure failures, validation failures, missing evidence
  summary, malformed strategy, and required missing evidence produce classified
  blockers.
- Existing review-ref requirements continue to apply.
- Disabled or optional strategy config preserves current behavior.

Tasks:

- T4.1 Add an evidence readiness evaluator over parsed ticket strategy state,
  execution result, validation result, and review refs.
- T4.2 Integrate the evaluator into ticket result application.
- T4.3 Integrate strict evidence gating into ticket status transitions to
  `ready_for_review`.
- T4.4 Add tests for passed, missing, malformed, blocked, validation-failed, and
  optional-strategy paths.
- T4.5 Update reference docs for review-transition evidence requirements.
- T4.6 Add the slice implementation report and validation evidence.

## Slice 5: Gherkin Test Policy

Goal: make the strategy pass understand when Gherkin and stronger test evidence
are required or intentionally not applicable.

Behavior:

- The worker context includes the evidence decision matrix from the spec.
- Gherkin active/applicable detection is explicit and testable.
- User-visible behavior tickets require Gherkin creation or update when Gherkin
  is active.
- Docs-only, Godoc-only, and non-behavioral refactor tickets can record that no
  new tests or Gherkin are needed.
- Bugfix, parser/validator, state-machine, IO, concurrency, and configuration
  tickets receive stronger default testing guidance.

Tasks:

- T5.1 Add project-level active Gherkin detection based on existing acceptance
  feature configuration/artifacts.
- T5.2 Add strategy guidance rendering for change kind, risk, test timing, and
  evidence matrix.
- T5.3 Add required Gherkin checks for active/applicable behavior tickets.
- T5.4 Add tests for docs-only no-test strategy, user-visible Gherkin-required
  strategy, bugfix regression guidance, and parser table/fuzz guidance.
- T5.5 Update acceptance/Gherkin docs to explain the evidence strategy
  interaction.
- T5.6 Add the slice implementation report and validation evidence.

## Slice 6: Evidence Visibility Docs

Goal: make evidence strategy visible and operable from the user-facing surfaces
without overloading tickets or runtime logs.

Behavior:

- Board ticket detail surfaces strategy status, validation state, and evidence
  summary in concise form.
- Dash work-running and completion output can point to evidence checklist and
  execution report paths.
- CLI commands that show tickets include evidence strategy status when present.
- User guide, command reference, ticket execution docs, and testing doctrine
  explain the default behavior and config overrides.
- QA/reference scenarios cover docs-only, behavior-change, and blocked
  validation examples.

Tasks:

- T6.1 Add evidence status to indexed ticket detail or a derived Board read
  model.
- T6.2 Render evidence status in Board detail and relevant Dash messages.
- T6.3 Update ticket show/list/board CLI output where appropriate.
- T6.4 Update user guide, command reference, runtime config, ticket execution,
  ticket review workflow, and testing doctrine docs.
- T6.5 Add or update QA fixture coverage for no-test docs work, behavior work
  with Gherkin, and blocked validation.
- T6.6 Add the slice implementation report and validation evidence.

