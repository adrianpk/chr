# Gherkin Acceptance Generation Tracker

Status: planned

Spec: `ops/specs/2026-06-06-gherkin-acceptance-generation.md`
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | acceptance readiness contract | delivered | `feat/slice-1-acceptance-readiness-contract` | #289 | `.prj/reports/slices/gherkin-acceptance-generation/slice-1-acceptance-readiness-contract.md` |
| Slice 2 | readiness-aware quiz questions | reviewing | `feat/slice-2-readiness-aware-quiz-questions` | pending | `.prj/reports/slices/gherkin-acceptance-generation/slice-2-readiness-aware-quiz-questions.md` |
| Slice 3 | feature artifact checks | planned | `feat/slice-3-feature-artifact-checks` | pending | `.prj/reports/slices/gherkin-acceptance-generation/slice-3-feature-artifact-checks.md` |
| Slice 4 | spec feature generation | planned | `feat/slice-4-spec-feature-generation` | pending | `.prj/reports/slices/gherkin-acceptance-generation/slice-4-spec-feature-generation.md` |
| Slice 5 | acceptance workflow docs | planned | `feat/slice-5-acceptance-workflow-docs` | pending | `.prj/reports/slices/gherkin-acceptance-generation/slice-5-acceptance-workflow-docs.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add acceptance readiness types and validation. | done | `0563555` | Added readiness decisions, slots, normalization, and validation. |
| T1.2 | Add semantic prompt intent rendering for acceptance readiness slots. | done | `0563555` | Added slot-specific semantic contracts and forbidden-slot metadata. |
| T1.3 | Add tests for valid, missing, contradictory, and non-applicable readiness payloads. | done | `0563555` | Added focused package tests for readiness decisions and prompt intent construction. |
| T1.4 | Add slice report and tracker updates. | done | this commit | Added Slice 1 report, tracker status, and validation log. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Extend quiz model request payloads with acceptance readiness context. | done | `4bd7660` | Added optional readiness payload, validation, context section, and service request derivation. |
| T2.2 | Update quiz question prompt construction to include readiness guidance. | done | `8300c83` | Added provider payload exposure, routing-state prompt intent, and deep/fast question guidance. |
| T2.3 | Persist readiness decisions and inferred elements with quiz runtime or curated quiz artifacts. | done | `10c1fce` | Added curated `Acceptance Readiness` notes derived from persisted prompt intent provenance. |
| T2.4 | Add tests for suitable, needs-more-info, and not-applicable quiz flows. | done | `afde96d` | Covered generate, needs-more-info, and not-applicable service quiz request flows. |
| T2.5 | Add slice report and tracker updates. | done | this commit | Added Slice 2 report, tracker status, and validation log. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add acceptance feature artifact types and layout helpers. | pending | pending | |
| T3.2 | Add deterministic static feature checks. | pending | pending | |
| T3.3 | Add tests for valid features, missing Given/When/Then, vague steps, missing traceability, and open-question scenarios. | pending | pending | |
| T3.4 | Add slice report and tracker updates. | pending | pending | |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Extend spec generation model contracts for optional feature output. | pending | pending | |
| T4.2 | Write accepted feature artifacts to the project workspace. | pending | pending | |
| T4.3 | Report feature generation status in spec generation results and logs. | pending | pending | |
| T4.4 | Add tests for generated, rejected, and not-applicable feature output. | pending | pending | |
| T4.5 | Add slice report and tracker updates. | pending | pending | |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add command and workflow reference docs for acceptance feature generation. | pending | pending | |
| T5.2 | Update user guide or explanation docs for Gherkin suitability behavior. | pending | pending | |
| T5.3 | Document the deferred `godog` acceptance-test execution path. | pending | pending | |
| T5.4 | Run `make verify`. | pending | pending | |
| T5.5 | Add final cycle report and tracker closure. | pending | pending | |

## Validation Log

- Slice 1: `go test ./internal/chc -run AcceptanceReadiness` (pass)
- Slice 1: `go test ./internal/chc` (pass)
- Slice 2: `go test ./internal/chc -run 'QuizModelProviderPayloadCarriesAcceptanceReadiness|QuizModelQuestionInstructions|QuizModelRoutingStateCarriesAcceptanceReadiness|OpenAIQuizModelRequestBodySerializesFocusedNextQuestionPayload'` (pass)
- Slice 2: `go test ./internal/chc -run 'QuizArtifactRepositoryWritesCuratedNotes|CurateQuizEntriesPreservesAcceptanceReadiness|QuizArtifactRepositoryLoadsRawCaptureEntries'` (pass)
- Slice 2: `go test ./internal/chcapp -run 'NewQuizModelRequest.*Acceptance'` (pass)
- Slice 2: `go test ./internal/chc` (pass)
- Slice 2: `go test ./internal/chcapp` (pass)

## Notes

- `godog` execution is intentionally deferred.
- Feature files are acceptance artifacts and do not replace specs.
- Plan and tracker must be committed to `dev` before Slice 1 implementation
  starts.
