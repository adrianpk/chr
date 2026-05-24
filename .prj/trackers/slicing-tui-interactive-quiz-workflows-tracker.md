# TUI Interactive Quiz Workflows Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`

## Active Slice

Slicing name: TUI Interactive Quiz Workflows

Active slice number: Slice 1

Active slice short name: quiz runtime model

Expected branch name: `feat/slice-1-quiz-runtime-model`

Expected PR title: `feat(slice-1): quiz runtime model`

Expected report path:
`.prj/reports/slices/tui-interactive-quiz-workflows/slice-1-quiz-runtime-model.md`

Prerequisite PRs or delivered work:

- Application operation service consolidation slicing.
- TUI Dump Workflows through PR #181.
- External dump editor shell-launch fix through PR #182.
- Workspace index and semantic cache slicing through PR #172.

## Slicing Checklist

- [x] Capture interactive quiz workflow direction.
- [x] Capture service-boundary requirement.
- [x] Capture canonical raw, curated, and runtime-state artifact contract.
- [x] Create stable spec.
- [x] Create model contract.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Quiz Runtime Model | Planned | `feat/slice-1-quiz-runtime-model` | `feat(slice-1): quiz runtime model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-1-quiz-runtime-model.md` | Core runtime model and validation only. |
| Slice 2: Quiz Artifact Repository | Planned | `feat/slice-2-quiz-artifact-repository` | `feat(slice-2): quiz artifact repository` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-2-quiz-artifact-repository.md` | File-backed raw, curated, and runtime-state artifacts. |
| Slice 3: Quiz Service Operations | Planned | `feat/slice-3-quiz-service-operations` | `feat(slice-3): quiz service operations` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-3-quiz-service-operations.md` | Service start/resume/answer/skip/done/cancel/mode operations. |
| Slice 4: Scoped Quiz Read Model | Planned | `feat/slice-4-scoped-quiz-read-model` | `feat(slice-4): scoped quiz read model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-4-scoped-quiz-read-model.md` | TUI read model and scoped quiz listing. |
| Slice 5: TUI Deep Quiz Loop | Planned | `feat/slice-5-tui-deep-quiz-loop` | `feat(slice-5): tui deep quiz loop` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-5-tui-deep-quiz-loop.md` | Interactive one-question-at-a-time deep mode. |
| Slice 6: Fast Quiz Candidates | Planned | `feat/slice-6-fast-quiz-candidates` | `feat(slice-6): fast quiz candidates` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-6-fast-quiz-candidates.md` | Internal candidate buffering and mode switch. |
| Slice 7: Ariadne Trail Recovery | Planned | `feat/slice-7-ariadne-trail-recovery` | `feat(slice-7): ariadne trail recovery` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-7-ariadne-trail-recovery.md` | Trail display, depth cap, branch return, resume recovery. |
| Slice 8: Quiz Curation Polish | Planned | `feat/slice-8-quiz-curation-polish` | `feat(slice-8): quiz curation polish` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-8-quiz-curation-polish.md` | Curated notes and documentation alignment. |

## Slice 1: Quiz Runtime Model

Candidate tasks:

- [ ] Add quiz scope, status, mode, question, answer, coverage, trail, and
      candidate model types.
- [ ] Add validation for runtime state, question depth, answers, skips,
      modes, statuses, and budget limits.
- [ ] Add tests for valid project-wide and use-case-scoped runtime state.
- [ ] Add tests for invalid depth, invalid answer/skip state, invalid mode,
      invalid status, and budget overflow.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 2: Quiz Artifact Repository

Candidate tasks:

- [ ] Add canonical quiz artifact path resolution for project and use-case
      scopes.
- [ ] Add raw capture rendering and append behavior.
- [ ] Add curated quiz rendering from distilled entries.
- [ ] Add runtime JSON state save/load behavior.
- [ ] Add candidate persistence only inside runtime state.
- [ ] Add repository tests for paths, rendering, resume, and candidate
      isolation.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 3: Quiz Service Operations

Candidate tasks:

- [ ] Add service quiz operation interface and request/result shapes.
- [ ] Implement list, start, resume, answer, skip, complete, cancel, switch
      mode, and refresh over the application service.
- [ ] Ensure answer and skip use service-owned active question state.
- [ ] Wire deep mode generation through the existing model/provider boundary.
- [ ] Persist raw capture and runtime state after each operation.
- [ ] Add service tests for control semantics and scoped artifacts.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

## Slice 4: Scoped Quiz Read Model

Candidate tasks:

- [ ] Add TUI quiz view model carrying scope, status, mode, paths, counts,
      updated time, active question summary, trail summary, and curated preview.
- [ ] Add adapter mapping from service quiz results to TUI quiz views.
- [ ] Render project-wide and use-case-scoped quizzes in `F4 Quizzes`.
- [ ] Add empty, idle, active, done, cancelled, and blocked detail states.
- [ ] Add focused TUI tests for scoped listing and detail rendering.
- [ ] Add slice implementation report.
- [ ] Run targeted tests and `make verify`.

Task commits:

- Pending.

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
