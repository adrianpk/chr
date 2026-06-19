# Evidence Strategy Configuration

Evidence strategy configuration controls how ChrysoCoord prepares ticket
workers to decide and record testing, validation, Gherkin, and runtime evidence.

The configuration controls ticket checklist writing, worker context, execution
report copying, and review-transition gating.

## Location

Evidence strategy settings live in `.chc/user/config.yml` under
`evidence.strategy`.

Generated config may omit these fields. Omitted fields use the defaults listed
below.

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

## Fields

- `enabled`: whether the Test/Evidence Strategy pass is active.
- `mode`: how strategy is produced.
- `ticket_checklist`: whether the ticket checklist is the required operational
  source for strategy state.
- `gherkin_policy`: how Gherkin requirements interact with strategy decisions.
- `strict_ready_for_review`: whether missing or failing required evidence blocks
  `ready_for_review`.
- `allow_user_edits`: whether user-authored checklist values are accepted as
  strategy input.
- `execution_report_copy`: whether runtime reports duplicate strategy details.

## Values

`mode` supports:

- `inline`: the development worker performs strategy inside the same worker run.
- `separate_pass`: reserved for a future separate strategy worker.

`ticket_checklist` supports:

- `required`: the ticket checklist is the required operational source.
- `optional`: strategy may be advisory and should not block by itself.

`gherkin_policy` supports:

- `active_when_applicable`: Gherkin is required when active and applicable to
  observable behavior.
- `disabled`: Gherkin is not required by evidence strategy.
- `manual_only`: Gherkin is created only when explicitly requested.

Gherkin is considered active when `gherkin_policy` is `active_when_applicable`
and the project already has `.feature` artifacts under
`.chc/user/acceptance/`. Active Gherkin applies to tickets that change
observable CLI, TUI, API, domain, acceptance workflow, or user-visible bugfix
behavior.

## Defaults

Defaults are intentionally strict enough to make agent work inspectable:

- strategy enabled;
- inline mode;
- ticket checklist required;
- Gherkin required when active and applicable;
- strict `ready_for_review` gating;
- user edits allowed;
- execution report copy enabled.

The defaults are designed to require no user decision for normal projects while
still allowing projects to relax or redirect the behavior through config.

## Review Transition Gate

When `enabled: true`, `ticket_checklist: required`, and
`strict_ready_for_review: true`, ChrysoCoord blocks `ready_for_review` until the
ticket evidence is ready.

The gate runs when:

- `chc ticket apply-result <execution-id>` applies a successful worker result;
- `chc ticket status <key> ready_for_review` is used directly.

Required ticket evidence:

- an anchored `chc:evidence-strategy:v1` checklist exists and parses;
- `Evidence strategy` is checked and `complete`;
- `Gherkin` is checked when its status is `required`;
- active/applicable Gherkin cannot be marked `not_applicable`;
- `Tests` is checked unless its status is `none`;
- `Validation` is checked and `passed`;
- `Evidence summary` is checked and `complete`.

For execution result application, the result envelope must also include a
complete `evidence_summary` with `strategy`, `produced`, and `validation`
values.

Blocking diagnostics use stable codes:

- `evidence_strategy_missing`;
- `evidence_strategy_malformed`;
- `evidence_strategy_blocked`;
- `gherkin_required_missing`;
- `tests_required_missing`;
- `validation_missing`;
- `validation_failed`;
- `runtime_unavailable`;
- `codex_runtime_unavailable`;
- `evidence_summary_missing`.

If the gate fails during result application, the ticket is moved to `blocked`
and the blocker section records the diagnostic. If the gate fails during direct
status change, the command returns an error and leaves the ticket unchanged.

Projects can relax this behavior by setting `ticket_checklist: optional`,
`strict_ready_for_review: false`, or `enabled: false`.

## Validation

Invalid enum values are configuration errors. Examples:

- unknown `mode`;
- unknown `ticket_checklist` policy;
- unknown `gherkin_policy`.

Boolean fields accept the same truthy values as other workspace config fields:
`true`, `yes`, `1`, and `on`.
