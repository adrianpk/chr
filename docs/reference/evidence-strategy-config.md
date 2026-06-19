# Evidence Strategy Configuration

Evidence strategy configuration controls how ChrysoCoord prepares ticket
workers to decide and record testing, validation, Gherkin, and runtime evidence.

The first implementation defines the configuration contract and defaults. Later
slices use these values to write ticket checklists, inject worker context, and
gate review transitions.

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

## Validation

Invalid enum values are configuration errors. Examples:

- unknown `mode`;
- unknown `ticket_checklist` policy;
- unknown `gherkin_policy`.

Boolean fields accept the same truthy values as other workspace config fields:
`true`, `yes`, `1`, and `on`.
