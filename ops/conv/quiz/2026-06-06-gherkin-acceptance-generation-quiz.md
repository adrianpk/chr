# Gherkin Acceptance Generation Quiz Notes

Status: curated

## Source

This quiz note captures user alignment from the June 6, 2026 discussion about
using dump and quiz context to generate a specification plus an associated
Gherkin acceptance artifact.

Related draft spec:

- `ops/specs/drafts/2026-06-06-gherkin-acceptance-generation.md`

## Summary

The user wants ChrysoCoord to bias early quiz generation toward discovering
whether a use case is naturally suitable for Gherkin-style acceptance scenarios.
The quiz should not force a fixed set of hardcoded questions or a fixed language.
Instead, it should use the initial dump and existing context to infer whether
the topic has behavior-testable elements.

If the topic is suitable, the quiz should ask for missing scenario inputs such
as actors, triggers, initial state, observable outcomes, success paths, failure
paths, and boundaries. If the topic is not suitable, the quiz should continue as
a more dump-oriented clarification flow without making the user answer
irrelevant acceptance-test questions.

The generated specification should be able to produce a linked Gherkin feature
artifact. A later implementation iteration can use that feature artifact to
drive `godog` acceptance tests.

## Decisions

- Keep the main specification as the canonical behavior contract.
- Treat Gherkin as an additional acceptance artifact, not as a replacement for
  the specification.
- Do not force Gherkin generation for every use case.
- Use an LLM-based suitability decision with deterministic validation around
  accepted outputs.
- Use plain module names such as quiz, spec, acceptance, feature, check, and
  run.
- Prefer `.feature` for executable Gherkin files rather than Markdown.
- Keep `godog` execution as a follow-up implementation iteration.

## Tensions

- A quiz should gather enough acceptance-test information without becoming
  tedious for non-Gherkinizable topics.
- The LLM can infer behavior structure from a dump, but the system must avoid
  letting it invent actors, triggers, or outcomes when the material is vague.
- Generated scenarios can improve clarity, but they can also create false
  confidence if they are not executable or linked to real requirements.

## Open Questions

- What exact metadata should connect a feature file to its source spec and use
  case?
- Should acceptance readiness be stored as a separate JSON artifact, embedded in
  the spec result, or both?
- Which existing spec generation service should own feature generation, and
  which package should own static Gherkin checks?
- Should quiz language follow the user's session language, project language
  configuration, or provider default?
