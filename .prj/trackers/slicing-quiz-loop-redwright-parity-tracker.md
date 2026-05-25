# Quiz Loop RedWright Parity Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`

## Active Slice

Slicing name: Quiz Loop RedWright Parity

Active slice number: Slice 4

Active slice short name: quiz parity lab validation

Expected branch name: `feat/slice-4-quiz-parity-lab-validation`

Expected PR title: `feat(slice-4): quiz parity lab validation`

Expected report path:
`.prj/reports/slices/quiz-loop-redwright-parity/slice-4-quiz-parity-lab-validation.md`

Prerequisite PRs or delivered work:

- TUI Interactive Quiz Workflows through PR #190.
- Post-slice TUI quiz UX fixes through the latest `dev -> main` alignment.
- Service-owned quiz operations, file-backed runtime state, scoped read models,
  direct LLM question generation, and TUI answer flow already delivered.
- Slice 1 quiz provider parity fixtures through PR #193.
- Slice 2 quiz prompt intent persistence through PR #194.
- Slice 3 quiz routing parity control through PR #195.

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
| Slice 1: Quiz Provider Parity Fixtures | Delivered | `feat/slice-1-quiz-provider-parity-fixtures` | `feat(slice-1): quiz provider parity fixtures` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-1-quiz-provider-parity-fixtures.md` | PR #193. |
| Slice 2: Quiz Prompt Intent Persistence | Delivered | `feat/slice-2-quiz-prompt-intent-persistence` | `feat(slice-2): quiz prompt intent persistence` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-2-quiz-prompt-intent-persistence.md` | PR #194. |
| Slice 3: Quiz Routing Parity Control | Delivered | `feat/slice-3-quiz-routing-parity-control` | `feat(slice-3): quiz routing parity control` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-3-quiz-routing-parity-control.md` | PR #195. |
| Slice 4: Quiz Parity Lab Validation | Planned | `feat/slice-4-quiz-parity-lab-validation` | `feat(slice-4): quiz parity lab validation` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-4-quiz-parity-lab-validation.md` | Active next slice. |

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

- [x] Add runtime storage for formal prompt intent metadata returned by the
      model.
- [x] Preserve follow-up depth, tags, completion hints, parent linkage, and
      topic intent across save/load.
- [x] Rehydrate next-question payloads from persisted intent metadata instead
      of display-oriented approximations.
- [x] Add tests for resume and entry-based reconstruction.
- [x] Add slice implementation report.
- [x] Run targeted tests and broader validation as needed.

Task commits:

- T2.1: `81664c8` persists quiz prompt intent.
- T2.2: `45dbc0d` covers prompt intent rehydration.
- T2.3: `23f45c9` satisfies strict lint whitespace.

Validation:

- `go test ./internal/chc ./internal/chcapp`
- `go test ./internal/chcapp`
- `git diff --check`
- `make verify`

## Slice 3: Quiz Routing Parity Control

Candidate tasks:

- [x] Align branch follow-up decisions with persisted prompt intent metadata.
- [x] Enforce maximum follow-up depth from the persisted model intent.
- [x] Return skipped and exhausted branches to broader coverage.
- [x] Add tests for vague answer, specific answer, skip, completion, and trunk
      return flows.
- [x] Add slice implementation report.
- [x] Run targeted tests and broader validation as needed.

Task commits:

- T3.1: `3ca138a` returns skipped quiz branches to trunk.

Validation:

- `go test ./internal/chc ./internal/chcapp`
- `git diff --check`
- `make verify`

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
