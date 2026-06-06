# Gherkin Acceptance Generation Plan

Status: planned

Active Slice: Slice 4 - spec feature generation

Spec: `ops/specs/2026-06-06-gherkin-acceptance-generation.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`

## Purpose

Implement Gherkin acceptance generation as an optional extension to the existing
dump, quiz, and spec workflow. The implementation must bias early quiz questions
toward acceptance-test readiness only when a use case is naturally suitable, and
it must generate linked `.feature` artifacts only when the available context is
strong enough.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice must update its implementation report under
  `.prj/reports/slices/gherkin-acceptance-generation/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before continuing.
- `godog` execution is out of scope for this slicing and remains a follow-up
  implementation candidate.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | acceptance readiness contract | `feat/slice-1-acceptance-readiness-contract` | `feat(slice-1): acceptance readiness contract` | `.prj/reports/slices/gherkin-acceptance-generation/slice-1-acceptance-readiness-contract.md` |
| Slice 2 | readiness-aware quiz questions | `feat/slice-2-readiness-aware-quiz-questions` | `feat(slice-2): readiness-aware quiz questions` | `.prj/reports/slices/gherkin-acceptance-generation/slice-2-readiness-aware-quiz-questions.md` |
| Slice 3 | feature artifact checks | `feat/slice-3-feature-artifact-checks` | `feat(slice-3): feature artifact checks` | `.prj/reports/slices/gherkin-acceptance-generation/slice-3-feature-artifact-checks.md` |
| Slice 4 | spec feature generation | `feat/slice-4-spec-feature-generation` | `feat(slice-4): spec feature generation` | `.prj/reports/slices/gherkin-acceptance-generation/slice-4-spec-feature-generation.md` |
| Slice 5 | acceptance workflow docs | `feat/slice-5-acceptance-workflow-docs` | `feat(slice-5): acceptance workflow docs` | `.prj/reports/slices/gherkin-acceptance-generation/slice-5-acceptance-workflow-docs.md` |

## Slice 1: Acceptance Readiness Contract

Goal: add the core domain contract for Gherkin suitability without changing quiz
runtime behavior yet.

Behavior:

- ChrysoCoord has explicit acceptance readiness decisions:
  `generate`, `needs_more_info`, and `not_applicable`.
- Readiness captures stable acceptance slots: actor, trigger, initial state,
  observable outcome, success path, failure path, and boundaries.
- Readiness prompt intents describe semantic targets and forbidden slots without
  hardcoding user-facing question text.
- Invalid readiness payloads fail validation before they can influence quiz or
  spec generation.

Tasks:

- T1.1 Add acceptance readiness types and validation.
- T1.2 Add semantic prompt intent rendering for acceptance readiness slots.
- T1.3 Add tests for valid, missing, contradictory, and non-applicable
  readiness payloads.
- T1.4 Add slice report and tracker updates.

## Slice 2: Readiness-Aware Quiz Questions

Goal: make quiz generation use acceptance readiness when the topic is suitable
or close to suitable.

Behavior:

- Quiz start evaluates the initial dump and context for acceptance suitability.
- The first quiz questions are biased toward missing acceptance slots only when
  readiness is `generate` or `needs_more_info`.
- Topics marked `not_applicable` continue through normal dump/spec
  clarification without Gherkin-oriented questions.
- Inferred readiness elements are stored as structured quiz context and treated
  as defaults rather than confirmed facts.

Tasks:

- T2.1 Extend quiz model request payloads with acceptance readiness context.
- T2.2 Update quiz question prompt construction to include readiness guidance.
- T2.3 Persist readiness decisions and inferred elements with quiz runtime or
  curated quiz artifacts.
- T2.4 Add tests for suitable, needs-more-info, and not-applicable quiz flows.
- T2.5 Add slice report and tracker updates.

## Slice 3: Feature Artifact Checks

Goal: add local validation for generated Gherkin feature content before it can
be treated as an acceptance artifact.

Behavior:

- ChrysoCoord can validate a `.feature` body with deterministic static checks.
- Feature checks require Feature/Scenario structure and Given/When/Then steps.
- Scenarios must link to source spec behavior and must not rely on open
  questions.
- Vague steps such as "works correctly" are rejected.
- Static feature checks do not require `godog`.

Tasks:

- T3.1 Add acceptance feature artifact types and layout helpers.
- T3.2 Add deterministic static feature checks.
- T3.3 Add tests for valid features, missing Given/When/Then, vague steps,
  missing traceability, and open-question scenarios.
- T3.4 Add slice report and tracker updates.

## Slice 4: Spec Feature Generation

Goal: generate a linked `.feature` artifact alongside the canonical spec when
readiness allows it.

Behavior:

- Spec generation can request and receive an optional feature artifact from the
  configured generation backend.
- A feature is written only when readiness is `generate` and static checks pass.
- Feature files are written under the selected acceptance artifact directory.
- Generated feature metadata links back to the spec path, use case, and source
  context.
- Spec generation still writes the spec when feature generation is
  not-applicable or rejected, while reporting the feature status.

Tasks:

- T4.1 Extend spec generation model contracts for optional feature output.
- T4.2 Write accepted feature artifacts to the project workspace.
- T4.3 Report feature generation status in spec generation results and logs.
- T4.4 Add tests for generated, rejected, and not-applicable feature output.
- T4.5 Add slice report and tracker updates.

## Slice 5: Acceptance Workflow Docs

Goal: document the completed first iteration and reserve `godog` execution for
a later slice set.

Behavior:

- Reference docs describe acceptance readiness, generated feature artifacts, and
  static checks.
- User-facing docs explain when Gherkin generation applies and when it is
  skipped.
- The future `godog` path is documented as a follow-up without implying it is
  implemented.
- Repository validation confirms the feature is integrated cleanly.

Tasks:

- T5.1 Add command and workflow reference docs for acceptance feature
  generation.
- T5.2 Update user guide or explanation docs for Gherkin suitability behavior.
- T5.3 Document the deferred `godog` acceptance-test execution path.
- T5.4 Run `make verify`.
- T5.5 Add final cycle report and tracker closure.

## Open Implementation Notes

- Prefer standard library checks unless a Gherkin parser dependency is explicitly
  approved.
- The spec remains canonical; feature files are acceptance artifacts.
- The first implementation must not execute `godog` or generate step
  definitions.
- If feature generation fails static checks, the failure should be visible but
  should not silently replace spec generation.
