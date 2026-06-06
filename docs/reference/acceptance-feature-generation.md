# Acceptance Feature Generation

Acceptance feature generation is an optional extension of spec generation.

When a use case is naturally behavior-testable, ChrysoCoord can generate a
linked Gherkin `.feature` artifact next to the canonical Markdown spec. The spec
remains the source of truth. The feature file is supporting acceptance evidence.

## Workflow

The workflow is:

```text
dump -> quiz readiness -> targeted quiz questions -> spec + optional feature
```

The quiz and spec generation flow can carry an acceptance readiness decision:

- `generate`: enough behavior context exists to attempt a feature artifact;
- `needs_more_info`: the quiz should ask focused questions for missing behavior
  slots;
- `not_applicable`: the topic should continue through normal dump/spec
  clarification without Gherkin-oriented questions.

Readiness slots are semantic targets, not fixed user-visible questions:

- actor
- trigger
- initial state
- observable outcome
- success path
- failure path
- boundaries

The model may phrase readiness questions naturally, using the conversation
language, while preserving the selected slot.

## Generated Artifact

When spec generation receives generated feature content and static checks pass,
ChrysoCoord writes:

```text
.chc/user/acceptance/<scope>.feature
```

Generated features should include traceability tags such as:

```gherkin
@spec:billing-sync-spec @use-case:billing-sync
Feature: Billing sync
```

Each scenario must describe observable behavior with concrete Given, When, and
Then steps.

## Static Checks

Feature files are checked before they are written as acceptance artifacts.

Static checks require:

- a `Feature:` heading;
- at least one `Scenario:` or `Scenario Outline:`;
- Given, When, and Then steps for every scenario;
- traceability to spec behavior through supported tags;
- no dependency on open questions;
- no vague steps such as "works correctly";
- no internal implementation details as scenario behavior.

If static checks fail, the Markdown spec still writes successfully and feature
status is reported as `rejected`.

## Godog Status

This iteration does not execute `godog` and does not generate step definitions.

A later workflow can run generated `.feature` files through `godog`, fail on
undefined or pending steps, and record `godog` output as normal validation
evidence. Until then, generated feature files are acceptance artifacts, not
executable test results.
