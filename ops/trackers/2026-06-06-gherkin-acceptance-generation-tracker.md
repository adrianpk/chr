# Gherkin Acceptance Generation Tracker

Status: planned

Spec: `ops/specs/2026-06-06-gherkin-acceptance-generation.md`
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | acceptance readiness contract | delivered | `feat/slice-1-acceptance-readiness-contract` | #289 | `.prj/reports/slices/gherkin-acceptance-generation/slice-1-acceptance-readiness-contract.md` |
| Slice 2 | readiness-aware quiz questions | delivered | `feat/slice-2-readiness-aware-quiz-questions` | #290 | `.prj/reports/slices/gherkin-acceptance-generation/slice-2-readiness-aware-quiz-questions.md` |
| Slice 3 | feature artifact checks | delivered | `feat/slice-3-feature-artifact-checks` | #291 | `.prj/reports/slices/gherkin-acceptance-generation/slice-3-feature-artifact-checks.md` |
| Slice 4 | spec feature generation | delivered | `feat/slice-4-spec-feature-generation` | #292 | `.prj/reports/slices/gherkin-acceptance-generation/slice-4-spec-feature-generation.md` |
| Slice 5 | acceptance workflow docs | reviewing | `feat/slice-5-acceptance-workflow-docs` | pending | `.prj/reports/slices/gherkin-acceptance-generation/slice-5-acceptance-workflow-docs.md` |

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
| T3.1 | Add acceptance feature artifact types and layout helpers. | done | `2f8f245` | Added feature artifact contract, canonical `.chc/user/acceptance/<scope>.feature` layout, and layout validation. |
| T3.2 | Add deterministic static feature checks. | done | `ffa196e` | Added line-oriented feature structure, step, traceability, open-question, vague-step, and implementation-detail checks. |
| T3.3 | Add tests for valid features, missing Given/When/Then, vague steps, missing traceability, and open-question scenarios. | done | `298a671` | Added focused coverage for valid and rejected feature bodies. |
| T3.4 | Add slice report and tracker updates. | done | this commit | Added Slice 3 report, tracker status, and validation log. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Extend spec generation model contracts for optional feature output. | done | `348c80a` | Added optional model `feature` output, validation, schema, response transport, and prompt guidance. |
| T4.2 | Write accepted feature artifacts to the project workspace. | done | `e100a1e` | Added checked feature writing and connected spec generation to `.chc/user/acceptance/<scope>.feature`. |
| T4.3 | Report feature generation status in spec generation results and logs. | done | `ff19c25` | Added explicit final feature status logging for spec generation. |
| T4.4 | Add tests for generated, rejected, and not-applicable feature output. | done | `9bafb1d` | Covered generated feature writes, invalid feature rejection, and not-applicable status. |
| T4.5 | Add slice report and tracker updates. | done | this commit | Added Slice 4 report, tracker status, and validation log. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add command and workflow reference docs for acceptance feature generation. | done | `2553dc8` | Added reference docs, command reference, workspace layout, and index entry. |
| T5.2 | Update user guide or explanation docs for Gherkin suitability behavior. | done | `2553dc8` | Updated user guide and document generation explanation. |
| T5.3 | Document the deferred `godog` acceptance-test execution path. | done | `2553dc8` | Documented `godog` as deferred, not implemented execution evidence. |
| T5.4 | Run `make verify`. | done | `1bd65c8` | Fixed strict lint findings and verified lint plus tests. |
| T5.5 | Add final cycle report and tracker closure. | done | this commit | Added Slice 5 report and tracker status for review. |

## Validation Log

- Slice 1: `go test ./internal/chc -run AcceptanceReadiness` (pass)
- Slice 1: `go test ./internal/chc` (pass)
- Slice 2: `go test ./internal/chc -run 'QuizModelProviderPayloadCarriesAcceptanceReadiness|QuizModelQuestionInstructions|QuizModelRoutingStateCarriesAcceptanceReadiness|OpenAIQuizModelRequestBodySerializesFocusedNextQuestionPayload'` (pass)
- Slice 2: `go test ./internal/chc -run 'QuizArtifactRepositoryWritesCuratedNotes|CurateQuizEntriesPreservesAcceptanceReadiness|QuizArtifactRepositoryLoadsRawCaptureEntries'` (pass)
- Slice 2: `go test ./internal/chcapp -run 'NewQuizModelRequest.*Acceptance'` (pass)
- Slice 2: `go test ./internal/chc` (pass)
- Slice 2: `go test ./internal/chcapp` (pass)
- Slice 2 post-merge: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 3: `go test ./internal/chc -run 'AcceptanceFeatureLayout|ValidateAcceptanceFeatureArtifact'` (pass)
- Slice 3: `go test ./internal/chc -run 'AcceptanceFeature|CheckAcceptanceFeature'` (pass)
- Slice 3: `go test ./internal/chc` (pass)
- Slice 3 post-merge: `go test ./internal/chc` (pass)
- Slice 4: `go test ./internal/chc ./internal/chcapp -run 'GeneratedSpecFeatureSchema|SpecGenerationInstructionsDescribeOptionalFeature|GeneratedSpecOutputCarriesFeature|WriteAcceptanceFeature|ServiceGenerateSpec.*Feature|ServiceGenerateSpecRejectsInvalidGeneratedFeature|ServiceGenerateSpecReportsNotApplicableFeature'` (pass)
- Slice 4: `go test ./internal/chc` (pass)
- Slice 4: `go test ./internal/chcapp` (pass)
- Slice 4 post-merge: `go test ./internal/chc ./internal/chcapp` (pass)
- Slice 5: `make verify` (pass)

## Notes

- `godog` execution is intentionally deferred.
- Feature files are acceptance artifacts and do not replace specs.
- Plan and tracker must be committed to `dev` before Slice 1 implementation
  starts.
