# Gherkin Acceptance Generation

Status: draft

## Summary

ChrysoCoord should extend the dump, quiz, and spec workflow so that
behavior-testable use cases can generate an associated Gherkin `.feature`
artifact alongside the canonical specification.

The quiz should not force every topic through Gherkin-oriented questions.
Instead, it should use the initial dump and available context to determine
whether the topic is naturally acceptance-testable. When it is, early quiz
questions should be biased toward the missing elements needed for useful
Gherkin scenarios. When it is not, the quiz should continue as a normal
dump/spec clarification flow.

Related artifacts:

- Draft: `ops/specs/drafts/2026-06-06-gherkin-acceptance-generation.md`
- Quiz notes: `ops/conv/quiz/2026-06-06-gherkin-acceptance-generation-quiz.md`

## Problem

Spec generation can turn dump and quiz material into expected behavior, but it
does not currently produce an executable acceptance artifact. This misses an
opportunity to validate whether the requirements are expressed in clear,
observable behavior terms and to create acceptance scenarios that can later run
against implementation work.

Blindly generating Gherkin for every spec would create noise. Some topics are
internal, architectural, documentation-only, or too vague to become useful
acceptance scenarios. The workflow needs a suitability step before feature file
generation.

## Goals

- Add a Gherkin suitability step to quiz generation.
- Use LLM-generated questions without hardcoding exact wording.
- Bias the first quiz questions toward actor, trigger, initial state,
  observable outcome, success path, failure path, and boundaries when useful.
- Let the LLM infer missing acceptance elements from the initial dump when the
  source material supports them.
- Treat inferred values as defaults or strong suggestions, not confirmed facts.
- Skip Gherkin-oriented questions when the topic is not acceptance-testable.
- Generate a linked `.feature` file alongside the spec when suitability is
  `generate`.
- Reserve executable `godog` acceptance test integration for a later iteration.

## Non-Goals

- Do not make Gherkin mandatory for every spec.
- Do not replace the Markdown spec with Gherkin.
- Do not implement `godog` execution in the first iteration.
- Do not generate step definitions in the first iteration.
- Do not introduce elaborate module names for this behavior.
- Do not create a fixed, user-visible questionnaire that every user must answer.

## Question Selection Pattern

The quiz should separate what information is needed from how the question is
worded. The useful pattern is:

- stable slots selected by the system;
- flexible prompt wording;
- `locale: auto` for user-facing questions;
- semantic contracts describing what each slot means;
- forbidden-slot metadata that keeps a question focused on one topic;
- bounded readiness questions that persist answers into reusable context.

ChrysoCoord should use this pattern for acceptance readiness:

- keep stable acceptance slots internally;
- let the LLM choose natural question wording;
- keep question language adaptive to the conversation;
- prevent a question from asking for unrelated fields;
- persist the readiness decision and extracted elements as structured context.

The acceptance readiness slots are:

- `actor`
- `trigger`
- `initial_state`
- `observable_outcome`
- `success_path`
- `failure_path`
- `boundaries`

These are not hardcoded question strings. They are semantic targets for the LLM
question provider.

## Workflow

The target flow is:

```text
dump -> quiz readiness -> targeted quiz questions -> spec + optional feature
```

The quiz receives the initial dump and context before asking the first
user-facing question.

It then produces a structured readiness decision:

- `generate`: enough acceptance elements exist to create useful Gherkin;
- `needs_more_info`: the topic appears suitable, but one or more acceptance
  elements are missing;
- `not_applicable`: Gherkin is not the right artifact for this topic.

If the decision is `generate`, the quiz continues normally and passes the
inferred acceptance elements forward.

If the decision is `needs_more_info`, the next questions should target only the
missing acceptance elements.

If the decision is `not_applicable`, the quiz should avoid Gherkin-oriented
questions and continue as dump/spec clarification.

## Prompt Contract

The quiz question provider should receive a semantic prompt intent rather than a
fixed English question.

Example intent:

```json
{
  "kind": "acceptance_readiness_question",
  "slot": "observable_outcome",
  "locale": "auto",
  "llm_required": true,
  "metadata": {
    "semantic_contract": {
      "slot": "observable_outcome",
      "intent": "Ask what result can be observed after the behavior runs.",
      "expected_answer_kind": "observable output, state, file, report, response, or error",
      "question_guidance": "Ask exactly one concise question for this slot.",
      "forbidden_slots": [
        "actor",
        "trigger",
        "initial_state",
        "success_path",
        "failure_path",
        "boundaries"
      ]
    }
  }
}
```

The provider may phrase the question naturally, in the user's conversation
language, but it must preserve the semantic target.

## Suitability Output

The readiness result should be structured:

```json
{
  "decision": "needs_more_info",
  "confidence": 0.74,
  "reasons": [
    "The topic describes a CLI workflow with observable validation output."
  ],
  "missing": ["failure_path"],
  "inferred_elements": {
    "actor": ["CLI user"],
    "trigger": ["runs the validation command"],
    "observable_outcome": ["validation report is written"]
  },
  "candidate_scenarios": [
    "Validation succeeds and writes a report"
  ]
}
```

The system should reject `generate` when required elements are empty,
contradictory, or based on unresolved open questions.

## Generated Artifacts

When readiness is `generate`, spec generation should emit:

- the canonical Markdown spec;
- a linked Gherkin `.feature` file.

Suggested path:

```text
.chc/user/acceptance/<topic>.feature
```

The feature should include traceability tags:

```gherkin
@spec:<spec-slug> @use-case:<use-case-slug>
Feature: Example behavior
```

Each scenario must map to a requirement, acceptance criterion, or explicit
behavior in the spec.

## Static Checks

The first implementation should validate generated feature files without
executing them.

Checks should require:

- the feature parses as Gherkin;
- each scenario has Given, When, and Then steps;
- each scenario links to spec behavior;
- no scenario depends on open questions;
- no vague steps such as "works correctly";
- scenarios describe observable behavior, not internal implementation details.

Feature generation failure should not silently become acceptance evidence.

## Future Godog Integration

A later iteration should use generated `.feature` files as executable
acceptance tests with `godog`.

That iteration should:

- run selected feature files through `godog`;
- fail validation on undefined or pending steps;
- record `godog` output in normal validation artifacts;
- keep Go unit and integration tests as the baseline validation layer.

## Acceptance Criteria

- Quiz generation can produce an acceptance readiness decision from dump and
  context.
- The readiness decision supports `generate`, `needs_more_info`, and
  `not_applicable`.
- The quiz uses stable semantic slots without fixed user-facing question text.
- The first readiness questions are biased toward missing acceptance elements
  only when the topic is Gherkin-suitable or close to suitable.
- Non-Gherkinizable topics continue through dump/spec clarification without
  forced acceptance questions.
- Inferred acceptance elements can be passed into spec generation as defaults.
- Spec generation can emit a linked `.feature` file when readiness is
  `generate`.
- Static checks reject malformed, vague, unlinked, or non-observable scenarios.
- The first implementation does not require `godog`.

## Validation

Implementation validation should include:

- tests for all readiness decisions;
- tests proving non-Gherkinizable topics skip acceptance-slot questions;
- tests proving suitable topics ask for missing acceptance elements first;
- tests proving semantic prompt intents use `locale: auto` and forbid unrelated
  slots;
- tests proving feature artifacts link back to source specs;
- tests for static Gherkin checks.

## Open Questions

- Should `.chc/user/acceptance/` be the canonical artifact directory?
- Should readiness output be stored as a JSON artifact, embedded in generation
  metadata, or both?
- Should failed feature generation block spec generation or produce a warning?
- Which package should own static Gherkin parsing and checks?
