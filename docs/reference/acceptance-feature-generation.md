# Acceptance Feature Generation

Acceptance feature generation is becoming a separate workflow from spec
generation.

When a use case is naturally behavior-testable, ChrysoCoord should generate a
linked Gherkin `.feature` artifact from the canonical Markdown spec and related
quiz/dump context. The spec remains the source of truth. The feature file is
supporting acceptance evidence.

## Workflow

The target workflow is:

```text
dump -> quiz -> spec -> acceptance feature round -> .feature
```

The general quiz gathers requirements for the use case. A later acceptance
feature round should use the spec, quiz, and dump context to decide whether a
feature can be generated or whether default-oriented follow-up questions are
needed.

The acceptance feature workflow can carry a readiness decision:

- `generate`: enough behavior context exists to attempt a feature artifact;
- `needs_more_info`: the feature round should ask focused default-confirmation
  questions for missing scenario details;
- `not_applicable`: the topic should continue through normal dump/spec
  clarification without Gherkin-oriented artifacts.

Readiness slots are semantic targets, not fixed user-visible questions:

- actor
- trigger
- initial state
- observable outcome
- success path
- failure path
- boundaries

The model may phrase feature-round questions naturally, using the conversation
language, while preserving the selected scenario gap.

## Generated Artifact

Feature generation should write:

```text
.chc/user/acceptance/<scope>.feature
```

Spec generation does not write this file. It generates only the Markdown spec.

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

If static checks fail, the feature must not be written as acceptance evidence.

## Godog Status

This iteration does not execute `godog` and does not generate step definitions.

A later workflow can run generated `.feature` files through `godog`, fail on
undefined or pending steps, and record `godog` output as normal validation
evidence. Until then, generated feature files are acceptance artifacts, not
executable test results.
