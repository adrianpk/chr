# Acceptance Feature Generation

When a use case is naturally behavior-testable, ChrysoCoord should generate a
linked Gherkin `.feature` artifact from the canonical Markdown spec and related
quiz/dump context. The spec remains the source of truth. The feature file is
supporting acceptance evidence.

## Workflow

The target workflow is:

```text
dump -> quiz -> spec -> acceptance feature round -> .feature
```

The general quiz gathers requirements for the use case. Acceptance feature
generation uses the spec, quiz, dump, and existing feature-round context to
decide whether a feature can be generated or whether default-oriented follow-up
questions are needed.

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

Feature generation writes:

```text
.chc/user/acceptance/<scope>.feature
```

Spec generation does not write this file. It generates only the Markdown spec.

When generation needs more information, ChrysoCoord writes feature-round notes:

```text
.chc/user/acceptance/<scope>-feature-round.md
```

Those notes contain focused prompts with inferred defaults and empty answer
slots. A later generation attempt reads the notes as acceptance context.

Current limitation: ChrysoCoord writes the feature-round prompts, but the TUI
does not yet collect their answers with the quiz-style interaction loop. Until
that flow exists, the answers must be recorded in the feature-round note before
retrying generation.

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

## Evidence Strategy Interaction

The Test/Evidence Strategy pass treats Gherkin as active when the project has
existing acceptance feature artifacts under:

```text
.chc/user/acceptance/*.feature
```

With the default `evidence.strategy.gherkin_policy: active_when_applicable`,
active Gherkin becomes required for tickets that change observable behavior.
Observable behavior includes CLI behavior, TUI behavior, API behavior, domain
use cases, acceptance workflow behavior, and user-visible bug fixes.

For those tickets, the worker context tells the agent:

- whether Gherkin is active for the project;
- whether it applies to the ticket;
- which `.feature` path should be updated first;
- which evidence decision matrix row applies.

If active Gherkin applies, the ticket checklist must record:

```text
Gherkin: required; path=<feature path>; reason=<observable behavior reason>
```

The item can be checked only after the `.feature` artifact is created or
updated. Marking it `not_applicable` for an active, applicable behavior ticket
blocks `ready_for_review` with `gherkin_required_missing`.

Gherkin remains non-universal. Docs-only, Godoc-only, metadata-only, internal
refactor, test-only cleanup, and low-level parser or validator work can still
record Gherkin as `not_applicable` when no observable behavior scenario is
being specified.

Set `evidence.strategy.gherkin_policy: disabled` to remove Gherkin from the
evidence gate. Set `manual_only` when a project wants feature artifacts only
after explicit human request.

## Godog Status

This iteration does not execute `godog` and does not generate step definitions.

A planned later workflow should let an agent use generated `.feature` files as
acceptance-test source material. For Go projects, that workflow will likely use
[`godog`](https://github.com/cucumber/godog), the Cucumber BDD framework for
Go, to connect Gherkin scenarios with executable step definitions.

That later workflow can generate or update acceptance tests from the feature
artifact, run selected `.feature` files through `godog`, fail on undefined or
pending steps, and record `godog` output as normal validation evidence. Until
then, generated feature files are acceptance artifacts, not executable test
results.
