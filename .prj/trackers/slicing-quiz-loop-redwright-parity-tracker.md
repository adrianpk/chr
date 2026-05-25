# Quiz Loop RedWright Parity Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`

## Active Slice

Slicing name: Quiz Loop RedWright Parity

Active slice number: Slice 1

Active slice short name: quiz provider parity fixtures

Expected branch name: `feat/slice-1-quiz-provider-parity-fixtures`

Expected PR title: `feat(slice-1): quiz provider parity fixtures`

Expected report path:
`.prj/reports/slices/quiz-loop-redwright-parity/slice-1-quiz-provider-parity-fixtures.md`

Prerequisite PRs or delivered work:

- TUI Interactive Quiz Workflows through PR #190.
- Post-slice TUI quiz UX fixes through the latest `dev -> main` alignment.
- Service-owned quiz operations, file-backed runtime state, scoped read models,
  direct LLM question generation, and TUI answer flow already delivered.

## Slicing Checklist

- [x] Capture corrective quiz parity direction.
- [x] Capture service-boundary requirement.
- [x] Capture formal request/response payload requirement.
- [x] Capture prompt intent persistence requirement.
- [x] Capture routing parity requirement.
- [x] Capture lab validation requirement.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Quiz Provider Parity Fixtures | Reviewing | `feat/slice-1-quiz-provider-parity-fixtures` | `feat(slice-1): quiz provider parity fixtures` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-1-quiz-provider-parity-fixtures.md` | PR pending. |
| Slice 2: Quiz Prompt Intent Persistence | Planned | `feat/slice-2-quiz-prompt-intent-persistence` | `feat(slice-2): quiz prompt intent persistence` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-2-quiz-prompt-intent-persistence.md` | Starts after Slice 1 merge. |
| Slice 3: Quiz Routing Parity Control | Planned | `feat/slice-3-quiz-routing-parity-control` | `feat(slice-3): quiz routing parity control` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-3-quiz-routing-parity-control.md` | Starts after Slice 2 merge. |
| Slice 4: Quiz Parity Lab Validation | Planned | `feat/slice-4-quiz-parity-lab-validation` | `feat(slice-4): quiz parity lab validation` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-4-quiz-parity-lab-validation.md` | Starts after Slice 3 merge. |

## Slice 1: Quiz Provider Parity Fixtures

Candidate tasks:

- [x] Add fixture coverage for next-question provider payload shape.
- [x] Add fixture coverage for draft-spec payload shape.
- [x] Add response decoding tests for question, skip, and completion-shaped
      provider responses.
- [x] Record the provider payload contract in the slice report.
- [x] Run targeted tests and broader validation as needed.

Task commits:

- T1.1: `5bb651a` locks provider payload and response parity fixtures.

Validation:

- `go test ./internal/chc`
- `git diff --check`
- `make verify`

## Slice 2: Quiz Prompt Intent Persistence

Candidate tasks:

- [ ] Add runtime storage for formal prompt intent metadata returned by the
      model.
- [ ] Preserve follow-up depth, tags, completion hints, parent linkage, and
      topic intent across save/load.
- [ ] Rehydrate next-question payloads from persisted intent metadata instead
      of display-oriented approximations.
- [ ] Add tests for resume and entry-based reconstruction.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and broader validation as needed.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 3: Quiz Routing Parity Control

Candidate tasks:

- [ ] Align branch follow-up decisions with persisted prompt intent metadata.
- [ ] Enforce maximum follow-up depth from the persisted model intent.
- [ ] Return skipped and exhausted branches to broader coverage.
- [ ] Add tests for vague answer, specific answer, skip, completion, and trunk
      return flows.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and broader validation as needed.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 4: Quiz Parity Lab Validation

Candidate tasks:

- [ ] Add or update reset guidance for representative quiz runtime state.
- [ ] Run project-level and use-case-level quiz validation against the test
      project.
- [ ] Record payload examples and generated-question samples.
- [ ] Confirm the default flow asks one question at a time without preloading a
      large batch.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- Pending.

Validation:

- Pending.

## Open Risks

- Provider output quality can still be affected by weak project metadata.
- Manual TUI validation remains required after deterministic contract tests.
- Fast mode behavior may need a later tuning slice after the base loop is
  corrected.
