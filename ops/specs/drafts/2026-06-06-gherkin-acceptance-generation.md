# Gherkin Acceptance Generation

Status: draft

## Summary

ChrysoCoord should extend the dump, quiz, and spec workflow so that behavior-testable
use cases can produce an associated Gherkin feature artifact alongside the
generated specification.

The workflow must avoid forcing every topic into Gherkin. The quiz should first
use the initial dump and available project context to decide whether the topic
is naturally suitable for Gherkin-style acceptance scenarios. If it is suitable,
the early quiz questions should bias toward discovering actors, triggers, initial
state, observable outcomes, success paths, failure paths, and boundaries. If it
is not suitable, the quiz should continue as a more dump-oriented clarification
flow.

Related quiz notes:

- `ops/conv/quiz/2026-06-06-gherkin-acceptance-generation-quiz.md`

## Problem

Generated specifications currently make expected behavior explicit, but they do
not produce a parallel executable acceptance artifact. This leaves two gaps:

- the generated spec may contain vague behavior that is not forced into
  observable Given/When/Then form;
- the implementation workflow does not receive acceptance scenarios that can
  later be run against completed work.

Adding Gherkin generation could close both gaps, but only when the source use
case is suitable. Forcing Gherkin onto internal refactors, architecture changes,
documentation-only work, or vague product direction would add ceremony and false
confidence.

## Goals

- Let the quiz detect whether a use case is naturally Gherkinizable.
- Bias early LLM-generated quiz questions toward Gherkin readiness when useful.
- Avoid hardcoded question lists and avoid fixed-language assumptions in the
  user-facing quiz flow.
- Use the initial dump and existing context as input to the suitability decision.
- Generate a Gherkin `.feature` file alongside the spec when readiness is high
  enough.
- Link the generated feature artifact to its source spec and use case.
- Preserve a path for a later iteration to use `godog` for executable acceptance
  tests.

## Non-Goals

- Do not make Gherkin generation mandatory for every spec.
- Do not replace the prose specification with Gherkin.
- Do not implement `godog` execution in the first iteration.
- Do not require step definition generation in the first iteration.
- Do not introduce elaborate domain names for simple workflow modules.
- Do not make generated scenarios pass validation merely because they parse.

## Workflow

The target authoring flow is:

```text
dump -> quiz suitability -> targeted quiz questions -> spec + optional feature
```

The quiz starts from a dump or existing context. Before asking normal
clarification questions, the LLM evaluates whether the topic has behavior that
can be expressed as acceptance scenarios.

The suitability decision has three outcomes:

- `generate`: enough information exists to produce useful Gherkin scenarios;
- `needs_more_info`: the topic appears suitable, but key scenario inputs are
  missing;
- `not_applicable`: Gherkin is not the right artifact for this topic.

When the decision is `generate`, the quiz can continue with normal spec
clarification and preserve inferred acceptance elements as strong defaults.

When the decision is `needs_more_info`, the quiz should ask the minimum useful
questions needed to fill missing acceptance elements.

When the decision is `not_applicable`, the quiz should avoid Gherkin-oriented
questions and continue as a dump-oriented or spec-oriented clarification flow.

## Suitability Inputs

The LLM suitability check should consider:

- actor: user, CLI caller, API client, worker, scheduler, reviewer, or other
  initiator;
- trigger: command, request, event, state transition, user action, or scheduled
  action;
- initial state: relevant existing project, ticket, file, configuration, record,
  or runtime state;
- observable outcome: output, file, report, status, API response, database row,
  or error;
- success path: expected behavior when inputs are valid;
- failure path: expected behavior when inputs are invalid or a dependency fails;
- boundaries: what is explicitly in scope or out of scope.

The check may infer these elements from the dump, but inferred values must be
treated as defaults or suggestions unless the source material clearly supports
them.

## Suitability Output

The suitability result should be structured data that can be checked before any
feature file is accepted.

Example shape:

```json
{
  "decision": "generate",
  "confidence": 0.82,
  "reasons": [
    "The use case has a CLI actor, command trigger, and observable report output."
  ],
  "missing": [],
  "inferred_elements": {
    "actors": ["CLI user"],
    "triggers": ["run validation command"],
    "observable_outcomes": ["validation report is written"]
  },
  "candidate_scenarios": [
    "Validation succeeds and writes a report",
    "Invalid configuration fails before running validation"
  ]
}
```

The system should reject a `generate` decision when required fields are empty or
contradict the source material.

## Quiz Behavior

The quiz provider should receive instructions to:

- inspect the initial dump and available project context before asking the first
  user-facing question;
- decide whether the topic is acceptance-testable;
- ask only the minimum Gherkin-readiness questions when the topic appears
  suitable;
- skip Gherkin-readiness questions when the topic is internal, architectural,
  documentation-only, or not externally observable;
- keep inferred actors, triggers, and outcomes visible to the model as defaults,
  but avoid presenting them as confirmed facts;
- continue with normal dump/spec clarification after the readiness gap is
  resolved or marked not applicable.

The quiz should not use a fixed question list. The prompt should bias the LLM
toward the required information categories while allowing it to choose natural
questions for the current topic and conversation language.

## Generated Artifacts

When suitable, spec generation should produce:

- the canonical spec Markdown file;
- a linked Gherkin feature file.

The feature file should use `.feature`, not `.md`, because the artifact is meant
to be executable by Gherkin tooling.

Suggested project-local path:

```text
.chc/user/acceptance/<topic>.feature
```

The feature should include tags that connect it to its source material:

```gherkin
@spec:<spec-slug> @use-case:<use-case-slug>
Feature: Example behavior
```

Each scenario should map to a requirement, acceptance criterion, or explicit
behavior in the spec.

## Static Checks

The first iteration should validate generated feature artifacts without running
implementation acceptance tests.

Static checks should require:

- the feature parses as Gherkin;
- each scenario has Given, When, and Then steps;
- each scenario links to a spec requirement or acceptance criterion;
- no scenario depends on unresolved open questions;
- no step relies on vague assertions such as "works correctly";
- generated scenarios describe observable behavior rather than internal
  implementation details.

If static checks fail, spec generation should report the feature artifact as
rejected or incomplete rather than treating it as acceptance evidence.

## Future Godog Integration

A later iteration should add executable acceptance validation with `godog`.

That iteration should:

- run selected `.feature` files through `godog`;
- fail validation on undefined or pending steps;
- record `godog` output in the normal validation artifacts;
- keep Go unit and integration tests as the baseline validation layer;
- treat Gherkin acceptance tests as high-level behavior checks, not as a
  replacement for package-level tests.

## Acceptance Criteria

- The quiz can produce a structured Gherkin suitability result from dump and
  context material.
- The suitability result supports `generate`, `needs_more_info`, and
  `not_applicable`.
- The quiz asks Gherkin-oriented questions only when the topic appears suitable
  or close to suitable.
- The quiz continues as a dump/spec clarification flow when Gherkin is not
  applicable.
- Inferred actor, trigger, initial state, outcome, success path, failure path,
  and boundary data can be passed into spec generation as defaults.
- Spec generation can emit a linked `.feature` artifact when suitability is
  `generate`.
- The generated feature file includes tags or metadata linking it to the source
  spec and use case.
- Static checks reject feature artifacts that do not parse or contain vague,
  unlinked, or non-observable scenarios.
- The first implementation iteration does not require `godog`.
- The spec clearly reserves `godog` acceptance execution for a later iteration.

## Validation

Implementation validation should include:

- focused tests for suitability decisions over generate, needs_more_info, and
  not_applicable cases;
- tests proving non-Gherkinizable topics do not receive forced
  Gherkin-readiness questions;
- tests proving Gherkinizable topics bias early questions toward missing
  acceptance elements;
- tests proving spec generation can link a feature artifact to its source spec;
- tests for static feature checks.

The later `godog` iteration should add validation that runs generated feature
files against project-specific step definitions.

## Open Questions

- Should `.chc/user/acceptance/` become the canonical feature artifact location,
  or should acceptance artifacts live beside specs?
- Should suitability output be stored as a separate JSON artifact, embedded in
  spec generation metadata, or both?
- Which service boundary should own static Gherkin parsing and checks?
- Should rejected feature generation block spec generation, or should the spec
  still be written with a warning?
- How should user-facing quiz language be selected when the workspace artifacts
  themselves must remain English?
