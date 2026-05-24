# TUI Interactive Quiz Workflows Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`

## Active Slice

Slicing name: TUI Interactive Quiz Workflows

Active slice number: Slice 5

Active slice short name: tui deep quiz loop

Expected branch name: `feat/slice-5-tui-deep-quiz-loop`

Expected PR title: `feat(slice-5): tui deep quiz loop`

Expected report path:
`.prj/reports/slices/tui-interactive-quiz-workflows/slice-5-tui-deep-quiz-loop.md`

Prerequisite PRs or delivered work:

- Application operation service consolidation slicing.
- TUI Dump Workflows through PR #181.
- External dump editor shell-launch fix through PR #182.
- Workspace index and semantic cache slicing through PR #172.
- Slice 1 quiz runtime model through PR #183.
- Slice 2 quiz artifact repository through PR #184.
- Slice 3 quiz service operations through PR #185.
- Slice 4 scoped quiz read model through PR #186.

## Slicing Checklist

- [x] Capture interactive quiz workflow direction.
- [x] Capture service-boundary requirement.
- [x] Capture canonical raw, curated, and runtime-state artifact contract.
- [x] Capture augmented question-generation context contract.
- [x] Capture dump plus complete quiz to draft spec promotion contract.
- [x] Capture formal model request/response boundary requirement.
- [x] Create stable spec.
- [x] Create model contract.
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
| Slice 1: Quiz Runtime Model | Delivered | `feat/slice-1-quiz-runtime-model` | `feat(slice-1): quiz runtime model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-1-quiz-runtime-model.md` | PR #183. |
| Slice 2: Quiz Artifact Repository | Delivered | `feat/slice-2-quiz-artifact-repository` | `feat(slice-2): quiz artifact repository` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-2-quiz-artifact-repository.md` | PR #184. |
| Slice 3: Quiz Service Operations | Delivered | `feat/slice-3-quiz-service-operations` | `feat(slice-3): quiz service operations` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-3-quiz-service-operations.md` | PR #185. |
| Slice 4: Scoped Quiz Read Model | Delivered | `feat/slice-4-scoped-quiz-read-model` | `feat(slice-4): scoped quiz read model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-4-scoped-quiz-read-model.md` | PR #186. |
| Slice 5: TUI Deep Quiz Loop | Planned | `feat/slice-5-tui-deep-quiz-loop` | `feat(slice-5): tui deep quiz loop` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-5-tui-deep-quiz-loop.md` | Interactive one-question-at-a-time deep mode. |
| Slice 6: Fast Quiz Candidates | Planned | `feat/slice-6-fast-quiz-candidates` | `feat(slice-6): fast quiz candidates` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-6-fast-quiz-candidates.md` | Internal candidate buffering and mode switch. |
| Slice 7: Ariadne Trail Recovery | Planned | `feat/slice-7-ariadne-trail-recovery` | `feat(slice-7): ariadne trail recovery` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-7-ariadne-trail-recovery.md` | Trail display, depth cap, branch return, resume recovery. |
| Slice 8: Quiz Curation Polish | Planned | `feat/slice-8-quiz-curation-polish` | `feat(slice-8): quiz curation polish` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-8-quiz-curation-polish.md` | Curated notes and documentation alignment. |

## Slice 1: Quiz Runtime Model

Candidate tasks:

- [x] Add quiz scope, status, mode, question, answer, coverage, trail, and
      candidate model types.
- [x] Add validation for runtime state, question depth, answers, skips,
      modes, statuses, and budget limits.
- [x] Add tests for valid project-wide and use-case-scoped runtime state.
- [x] Add tests for invalid depth, invalid answer/skip state, invalid mode,
      invalid status, and budget overflow.
- [x] Add slice implementation report.
- [x] Run targeted tests and `make verify`.

Task commits:

- T1.1: `549161e` adds quiz runtime model types.
- T1.2: `55e7896` adds runtime validation.
- T1.3: `2bb3a0d` covers valid project-wide and use-case runtime scopes.
- T1.4: `cb06763` covers invalid runtime and entry states.
- T1.5: `863ca6d` refines quiz context, promotion, and model boundary docs.

Validation:

- `go test ./internal/chc`
- `make verify`

## Slice 2: Quiz Artifact Repository

Candidate tasks:

- [x] Add canonical quiz artifact path resolution for project and use-case
      scopes.
- [x] Add raw capture rendering and append behavior.
- [x] Add curated quiz rendering from distilled entries.
- [x] Add runtime JSON state save/load behavior.
- [x] Add candidate persistence only inside runtime state.
- [x] Add repository tests for paths, rendering, resume, and candidate
      isolation.
- [x] Add slice implementation report.
- [x] Run targeted tests and `make verify`.

Task commits:

- T2.1: `4181472` adds canonical quiz artifact paths, raw capture append,
  curated note writes, runtime state JSON save/load, and tests.

Validation:

- `go test ./internal/chc`
- `make verify`

## Slice 3: Quiz Service Operations

Candidate tasks:

- [x] Add service quiz operation interface and request/result shapes.
- [x] Implement list, start, resume, answer, skip, complete, cancel, switch
      mode, and refresh over the application service.
- [x] Ensure answer and skip use service-owned active question state.
- [x] Wire deep mode generation through the existing model/provider boundary.
- [x] Add formal semantic request and response payload contracts for
      model-backed quiz operations.
- [x] Include agent protocol, project, use-case, and relevant dump context in
      model-backed question generation.
- [x] Persist raw capture and runtime state after each operation.
- [x] Add service tests for control semantics and scoped artifacts.
- [x] Add slice implementation report.
- [x] Run targeted tests and `make verify`.

Task commits:

- T3.1: `f03cfa0` adds the formal quiz model request and response boundary.
- T3.2: `341ffbc` adds service-owned quiz operations and tests.

Validation:

- `go test ./internal/chc ./internal/chcapp`
- `make verify`

## Slice 4: Scoped Quiz Read Model

Candidate tasks:

- [x] Add TUI quiz view model carrying scope, status, mode, paths, counts,
      updated time, active question summary, trail summary, and curated preview.
- [x] Add adapter mapping from service quiz results to TUI quiz views.
- [x] Render project-wide and use-case-scoped quizzes in `F4 Quizzes`.
- [x] Add empty, idle, active, done, cancelled, and blocked detail states.
- [x] Add focused TUI tests for scoped listing and detail rendering.
- [x] Add slice implementation report.
- [x] Run targeted tests and `make verify`.

Task commits:

- T4.1: `ce57db4` adds service and TUI scoped quiz read models.

Validation:

- `go test ./internal/chcapp ./internal/chctui`
- `make verify`


## Slice 5: TUI Deep Quiz Loop

Candidate tasks:

- [ ] Add TUI action state for start, resume, answer, skip, done, cancel, and
      refresh.
- [ ] Add TUI-native answer input for the active question.
- [ ] Dispatch all quiz actions through the service executor.
- [ ] Preserve selected quiz and focus state across operations.
- [ ] Render budget, counts, status messages, and active question state.
- [ ] Add TUI tests for keyboard behavior and service intent dispatch.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 6: Fast Quiz Candidates

Candidate tasks:

- [ ] Add mode switch operation to the TUI and service result flow.
- [ ] Add internal fast candidate generation and activation behavior.
- [ ] Route fast candidate generation through the formal model request/response
      boundary.
- [ ] Ensure ambiguous answers fall back to deep generation.
- [ ] Ensure discarded candidates stay out of raw capture and curated notes.
- [ ] Add candidate cleanup during branch selection, mode switch, completion,
      replacement, and stale resume.
- [ ] Add tests for candidate activation, discard, cleanup, and mode switching.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 7: Ariadne Trail Recovery

Candidate tasks:

- [ ] Persist trunk id, parent id, topic, depth, trail, and branch return
      reason in runtime state.
- [ ] Render compact trail information in the Quizzes detail panel.
- [ ] Enforce maximum follow-up depth of 3.
- [ ] Return to broader coverage after branch closure.
- [ ] Reconstruct active question and trail from runtime state on resume.
- [ ] Add tests for depth cap, trunk return, trail rendering, and missing-state
      recovery.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 8: Quiz Curation Polish

Candidate tasks:

- [ ] Generate or refresh curated quiz notes when the user marks a quiz done.
- [ ] Render curated body in the Quizzes detail panel when no active question
      is running.
- [ ] Ensure curated notes distill raw turns instead of mirroring them
      one-to-one.
- [ ] Add promotion support from complete quiz plus relevant dump context to a
      spec artifact with `status: draft`.
- [ ] Route draft spec synthesis through the formal model request/response
      boundary.
- [ ] Align docs with the implemented interactive quiz contract.
- [ ] Add tests for curated artifact shape and TUI curated-body rendering.
- [ ] Add final slicing report or completion notes.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Open Decisions

- Whether the first fast implementation should generate ten candidates or use a
  smaller default for TUI responsiveness.
- Whether completed quiz snapshots should be archived in addition to one main
  quiz per scope.
- Whether curation on done should always invoke the model/provider boundary or
  initially use deterministic extraction from entries.

## Completion Notes

- Pending.
