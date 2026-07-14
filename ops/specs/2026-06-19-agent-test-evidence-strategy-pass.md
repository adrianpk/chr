# Agent Test/Evidence Strategy Pass

Status: draft

Date: 2026-06-19

Quiz: `ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz.md`

Raw quiz capture:
`ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz-raw.md`

Related:

- `docs/explanation/testing-and-engineering-doctrine.md`
- `docs/reference/runtime-config.md`
- `docs/reference/review-workflow-config.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/work-readiness.md`
- `ops/specs/2026-06-18-agentic-loop-review-merge.md`

## Summary

Chryso should add a mandatory Test/Evidence Strategy pass to development
worker execution.

The pass happens after a ticket is selected and before the development worker
modifies project files. It makes the agent classify the change, decide what
evidence is required, decide whether Gherkin applies, decide whether tests are
needed and when to create them, and persist that decision in the ticket
tasklist/checklist.

The feature is evidence-first, not test-first. The worker does not blindly add
tests, Gherkin, or ceremony for every change. It must choose the smallest
evidence strategy that is adequate for the actual risk and then reconcile that
strategy before the ticket can move to `ready_for_review`.

## Goals

- Make the development worker explicitly decide its testing and evidence
  strategy before code changes.
- Persist the strategy in the ticket checklist/tasklist as the primary
  operational source.
- Keep the checklist parseable enough for Chryso and future workers to
  recognize and validate.
- Keep the checklist readable enough for humans to inspect and edit.
- Allow user-authored checklist edits without making user-authored strategy the
  normal workflow.
- Make the worker prompt explicitly read, obey, update, and reconcile the
  persisted strategy.
- Require Gherkin creation or update when Gherkin is active and the change
  affects observable behavior.
- Prevent `ready_for_review` when required evidence, validation, linting,
  compilation, or runtime proof is missing or failing.
- Keep runtime execution reports as secondary trace evidence, not the primary
  operational state.
- Provide sane defaults through configuration while documenting how projects can
  change behavior.

## Non-Goals

- Do not impose universal TDD.
- Do not require tests for docs-only, wording-only, Godoc-only, or otherwise
  non-behavioral changes.
- Do not launch a separate Codex container for strategy by default.
- Do not make the user answer a testing strategy question for every ticket.
- Do not turn tickets into long prose reports.
- Do not make Gherkin universal.
- Do not allow a worker to mark a ticket `ready_for_review` by merely explaining
  why it skipped required evidence.
- Do not replace the project `runtime.validation_command`.
- Do not make execution reports the only place where strategy decisions live.

## Terminology

`Test/Evidence Strategy`

: The worker's explicit decision about which tests, Gherkin artifacts,
  validation commands, and runtime/manual proof are required for the ticket.

`Evidence Summary`

: The worker's end-of-run reconciliation of planned evidence against produced
  evidence.

`Primary operational source`

: The artifact Chryso should inspect to determine whether the ticket has an
  evidence strategy. For this feature, the primary source is the ticket
  checklist/tasklist.

`Execution report`

: Runtime trace evidence written by the agent execution system. It may duplicate
  the strategy and summary but is not the source used by the board for ordinary
  ticket state.

`Active Gherkin`

: Project or feature configuration indicates that acceptance feature artifacts
  are enabled for behavior that benefits from Gherkin integration.

`Applies to observable behavior`

: The ticket changes behavior visible through a user workflow, CLI command, TUI
  interaction, API boundary, domain behavior, or acceptance scenario.

## Desired Lifecycle

The development worker lifecycle becomes:

```text
select ticket
-> work readiness/preflight
-> read ticket and existing checklist
-> Test/Evidence Strategy pass
-> persist strategy checklist in ticket
-> implementation
-> validation
-> evidence summary reconciliation
-> board transition or blocked outcome
```

The strategy pass is inline in the development worker by default. A separate
strategy worker may be added later for high-criticality projects, but the first
contract should not require a second container or orchestration phase.

## Ticket Checklist Contract

The strategy should be stored in a clearly bounded checklist block in the ticket.

Preferred shape:

```md
<!-- chr:evidence-strategy:v1 -->
- [ ] Evidence strategy: planned; change_kind=<kind>; risk=<risk>; test_timing=<timing>; note=<reason>
- [ ] Gherkin: <required|not_applicable|disabled|blocked>; path=<path|none>; reason=<reason>
- [ ] Tests: <none|existing|unit|table|integration|fuzz|property|mutation|runtime|mixed>; reason=<reason>
- [ ] Validation: <pending|passed|failed|blocked>; commands=<commands>; evidence=<path|summary|none>
- [ ] Evidence summary: pending
<!-- /chr:evidence-strategy -->
```

The HTML comments are machine anchors. The checklist items remain normal
Markdown so humans can read and edit them.

### Required Labels

Chryso must recognize these labels inside the strategy block:

- `Evidence strategy`
- `Gherkin`
- `Tests`
- `Validation`
- `Evidence summary`

The parser should tolerate whitespace differences, but label names are stable.
Changing labels means changing the checklist schema version.

### Checklist States

Checklist items start unchecked.

The worker may check an item only when it has completed the corresponding
obligation:

- `Evidence strategy` may be checked after the strategy decision is written.
- `Gherkin` may be checked after the required `.feature` artifact is created or
  updated, or after the strategy records `not_applicable` or `disabled`.
- `Tests` may be checked after required tests are added or the strategy records
  that no new tests are required.
- `Validation` may be checked only after required validation has passed.
- `Evidence summary` may be checked only after planned and produced evidence are
  reconciled.

If an item is `blocked`, the ticket must not move to `ready_for_review`.

### User Edits

Users may edit the checklist before or during agent work.

The worker must:

- preserve user edits unless they are syntactically invalid;
- treat user-provided strategy as strong input;
- fill missing required fields;
- report conflicts between user instructions and project policy;
- avoid silently replacing a user-authored strategy with a new one.

If the existing checklist is malformed, the worker should block with a specific
strategy parse error unless a safe repair is possible.

## Worker Prompt Contract

The Codex worker prompt must include a strategy section before implementation
instructions.

Minimum prompt obligations:

```text
Before modifying project files:
1. Read the ticket and any existing evidence-strategy checklist.
2. Classify the change kind and risk.
3. Decide whether Gherkin applies.
4. Decide whether new tests are required.
5. Decide test timing: before, during, after, or none.
6. Write or update the ticket evidence-strategy checklist.
7. Treat the checklist as an operational contract during implementation.

Before finishing:
1. Re-read the strategy checklist.
2. Reconcile planned evidence with produced evidence.
3. Update Validation and Evidence summary.
4. Do not move the ticket to ready_for_review if required evidence is missing,
   failing, or blocked.
```

The worker must not treat the checklist as optional commentary.

## Change Classification

The worker should classify each ticket with one primary `change_kind`.

Initial controlled values:

- `docs_only`
- `godoc_only`
- `metadata_only`
- `test_only`
- `refactor_no_behavior`
- `bugfix`
- `behavior`
- `cli_behavior`
- `tui_behavior`
- `api_behavior`
- `parsing_or_validation`
- `state_machine`
- `persistence_or_filesystem`
- `network_or_external_io`
- `concurrency_or_timeout`
- `configuration`
- `performance`
- `security_or_safety`
- `unknown`

If more than one applies, the worker should choose the highest-risk primary kind
and mention secondary concerns in `note=`.

## Risk Levels

Initial risk levels:

- `none`
- `low`
- `medium`
- `high`
- `critical`

Risk should increase when the ticket touches:

- behavior visible to users;
- parser or validator boundaries;
- persistence, filesystem, network, or external IO;
- state machines;
- concurrency, queues, retries, workers, timers, or timeouts;
- security, authorization, secret handling, or destructive operations;
- code with high complexity or weak existing evidence;
- runtime scaffolding or agent orchestration.

Risk should stay `none` or `low` for:

- wording-only documentation changes;
- Godoc-only changes that do not affect generated behavior;
- link-only documentation maintenance;
- cosmetic metadata changes.

## Evidence Decision Matrix

The worker should use these defaults unless project configuration or ticket
instructions override them.

| Change kind | Default evidence |
| --- | --- |
| `docs_only`, `godoc_only`, `metadata_only` | No new tests. Validate links, formatting, or repository docs checks when available. |
| `test_only` | Run the relevant test command and repository validation. No Gherkin unless observable behavior is being specified. |
| `refactor_no_behavior` | Run existing tests and validation. Add tests only when the refactor exposes an obvious missing safety net. |
| `bugfix` | Prefer a failing regression test before or during implementation. Add Gherkin when user-visible behavior changes and Gherkin is active. |
| `behavior`, `cli_behavior`, `tui_behavior`, `api_behavior` | Add or update tests for the behavior. Add or update Gherkin when active and applicable. Include runtime proof when useful. |
| `parsing_or_validation` | Prefer table-driven tests and fuzz/property-style checks when the input space warrants it. |
| `state_machine` | Prefer table-driven transition tests, invariants, and invalid-state checks. |
| `persistence_or_filesystem` | Prefer integration tests with temp directories, fixtures, and failure cases. |
| `network_or_external_io` | Prefer fakes, timeout/failure cases, and integration tests when practical. |
| `concurrency_or_timeout` | Prefer boundedness checks, timeout tests, race-oriented validation, and failure-path tests. |
| `configuration` | Test defaulting, invalid config, and override behavior. |
| `performance` | Add benchmarks or boundedness evidence only when the ticket changes performance-sensitive behavior. |
| `security_or_safety` | Require explicit tests or validation evidence unless impossible; missing evidence blocks review. |
| `unknown` | Default to conservative evidence and do not move to review without a clear strategy. |

## Test Timing

The strategy must record `test_timing`.

Allowed values:

- `before`
- `during`
- `after`
- `none`

TDD is encouraged but not mandatory.

Prefer `before` when:

- fixing a reproducible bug;
- protecting logic with clear input/output behavior;
- capturing a regression;
- changing a parser, validator, or state machine with known expected outputs.

Prefer `during` when:

- the correct test shape becomes clearer while implementing;
- integration boundaries need small code discovery;
- fixtures or fakes must be shaped alongside implementation.

Prefer `after` when:

- UI or workflow proof is easier after the behavior exists;
- the risk is low but validation evidence is still needed;
- the ticket mainly wires existing behavior.

Use `none` only when:

- no behavior changes;
- existing validation is sufficient;
- the strategy records a concrete reason.

## Gherkin Policy

Gherkin is not universal.

If Gherkin is active and the ticket affects observable behavior, the worker must
create or update a `.feature` artifact.

Observable behavior includes:

- user-visible CLI behavior;
- user-visible TUI behavior;
- API behavior;
- domain behavior expressed through a use case;
- acceptance workflow changes;
- bug fixes visible to users.

Gherkin usually does not apply to:

- docs-only edits;
- Godoc-only edits;
- internal refactors without behavior change;
- purely mechanical renames;
- tests-only cleanup;
- internal implementation details that are better captured by unit, property,
  fuzz, or integration tests.

If Gherkin is active and applies, `Gherkin: not_applicable` is invalid.

If Gherkin is required but cannot be written because context is missing, the
ticket must block instead of moving to review.

## Validation and Closure Rules

A ticket may move to `ready_for_review` only when all of these are true:

- the evidence-strategy checklist exists;
- all required checklist labels are present;
- strategy parse status is valid;
- `Evidence strategy` is complete;
- required Gherkin status is complete or not applicable;
- required tests are complete or not applicable;
- validation is passed;
- evidence summary is complete;
- linting has no unresolved failures when linting is part of validation;
- compilation has no unresolved failures;
- required tests have no unresolved failures;
- required runtime proof has no unresolved failures.

The worker must not move a ticket to `ready_for_review` when:

- Docker or Codex runtime is unavailable;
- project runtime validation cannot start;
- required validation command is missing;
- required validation command fails;
- required Gherkin is missing;
- required tests are missing;
- evidence summary is missing;
- the strategy checklist is malformed and cannot be safely repaired.

## Evidence Summary

At closure, the worker must update `Evidence summary`.

The summary should be compact and parseable:

```md
- [x] Evidence summary: produced=<summary>; validation=<passed|blocked|failed>; deviations=<none|summary>; remaining_risk=<none|low|medium|high>
```

The execution report may include a longer version:

```md
## Evidence Summary

Planned evidence:
- ...

Produced evidence:
- ...

Validation:
- `make verify` passed

Deviations:
- none

Remaining risk:
- none
```

The ticket checklist remains the board-facing source. The execution report is a
trace and review aid.

## Execution Report Duplication

The runtime execution report should include:

- copied initial strategy;
- final checklist state;
- produced evidence;
- validation commands and results;
- links or paths to created tests, Gherkin, runtime logs, validation logs, and
  PR evidence when available.

The report must not be the only place where strategy exists.

If the report and ticket checklist disagree, the ticket checklist is the
operational source and Chryso should surface a diagnostic.

## Configuration

Configuration should live in `.chr/default/config.yml` under a new top-level
`evidence:` key.

Generated config may omit defaults. Omitted fields use sane defaults.

Proposed default config:

```yml
evidence:
  strategy:
    enabled: true
    mode: inline
    ticket_checklist: required
    gherkin_policy: active_when_applicable
    strict_ready_for_review: true
    allow_user_edits: true
    execution_report_copy: true
```

Fields:

- `enabled`: whether the evidence strategy pass is active.
- `mode`: `inline` for the development worker. Future values may include
  `separate_pass`.
- `ticket_checklist`: `required` or `optional`.
- `gherkin_policy`: `active_when_applicable`, `disabled`, or
  `manual_only`.
- `strict_ready_for_review`: whether missing or failing required evidence blocks
  the review transition.
- `allow_user_edits`: whether user-authored checklist values are accepted as
  strategy input.
- `execution_report_copy`: whether runtime reports duplicate strategy details.

Default behavior:

- strategy pass enabled;
- inline mode;
- ticket checklist required;
- Gherkin required when active and applicable;
- strict review transition;
- user edits allowed;
- execution report copy enabled.

## Board Behavior

The board should expose evidence readiness without requiring the user to open
runtime logs.

At minimum, ticket detail should be able to show:

- strategy missing;
- strategy planned;
- strategy blocked;
- validation pending;
- validation passed;
- evidence summary complete.

Board transitions should respect strategy state:

- a missing strategy blocks development completion;
- a blocked strategy blocks review transition;
- a failed validation blocks review transition;
- a completed summary allows normal review transition when other readiness
  checks pass.

## Compatibility

Existing tickets without evidence-strategy checklists should remain readable.

When a worker picks such a ticket, it should insert the strategy block during the
initial strategy pass.

Chryso should not require bulk migration of existing tickets before the
feature is used.

If a project disables the feature through config, existing checklist blocks
should be preserved but not enforced.

## Diagnostics

Chryso should provide clear diagnostics for blocked outcomes.

Initial diagnostic reasons:

- `evidence_strategy_missing`
- `evidence_strategy_malformed`
- `evidence_strategy_blocked`
- `gherkin_required_missing`
- `tests_required_missing`
- `validation_missing`
- `validation_failed`
- `runtime_unavailable`
- `codex_runtime_unavailable`
- `evidence_summary_missing`

Diagnostics should be short, actionable, and usable by both the TUI and CLI.

## Acceptance Criteria

- A development worker prompt requires an evidence strategy before file changes.
- A ticket without a strategy receives a parseable checklist block before
  implementation begins.
- A ticket with a user-edited strategy preserves that strategy and fills missing
  required fields.
- Docs-only tickets can complete with no new tests when the strategy records why.
- User-visible behavior tickets create or update Gherkin when Gherkin is active.
- Bugfix tickets prefer a regression test before or during implementation.
- Parsing or validation tickets prefer table-driven tests and consider fuzz or
  property-style checks.
- Runtime or infrastructure failures block the ticket instead of moving it to
  `ready_for_review`.
- Validation failures block `ready_for_review` until fixed or explicitly marked
  blocked.
- The final evidence summary reconciles planned and produced evidence.
- Execution reports duplicate strategy details but are not required as the only
  source of strategy state.
- Project configuration can override default strategy behavior.
- Documentation explains defaults and override behavior.

## Open Questions

- Exact Markdown parser implementation for checklist blocks.
- Exact relationship between checklist task completion and existing ticket state
  update commands.
- Exact command or TUI surface for repairing malformed evidence checklist
  blocks.
- Exact detection mechanism for active Gherkin in a project.
- Whether `evidence:` should remain a top-level config key or become nested under
  `work:` after implementation review.
- Whether separate-pass mode should be documented as future-only or included as
  a disabled config enum in the first implementation.

