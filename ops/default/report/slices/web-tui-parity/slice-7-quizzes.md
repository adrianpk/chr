# Slice 7: Complete Quiz interaction

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-7-tui-quizzes`
PR: [#461](https://forge.adrianpk.com/adrianpk/chr/pulls/461)
Base: `754ee017880be3e2dfb08b3c282898da0356456a` on dev, after verified PR #460 integration and Slice 6 closure

## Purpose

Resolve G-020, G-021 and G-022 through visible quiz modes, retained answer
history, guarded submissions and truthful contextual-authoring availability.

## Delivered Behavior

- Deep/Fast is visible before Start. `m` selects the target's mode without
  creating a session; Start submits that exact choice. Pending-question markers
  identify targets awaiting answers, within the functional case filter.
- Saved question/answer history appears newest first beside the captured current
  question and input. Skipped questions are excluded. PgUp/PgDn pages complete
  retained text while resize and navigation preserve input. Selecting another
  quiz resets its history offset.
- Answer, Skip, Done, Cancel, Resume and mode changes use existing service owners.
  `alt+x` confirms cancellation even with an empty answer; saved history remains.
  `u` resumes retained work. Skip, Done and mode changes ask before discarding
  entered text. Esc confirms discarding only the current draft.
- Every TUI mutation carries the captured session, question, composite revision
  and operation identity. Input changes, including newline edits, rotate intent;
  retries keep the original request. Stale submissions retain their old prompt
  and text, and confirmed reload installs the current question. Duplicate
  execution reconciles once without answering another question.
- The separate contextual-authoring entry explains its TUI unavailability before
  any confirmation or executor call. It names the existing authoring CLI and
  supported F3 project/case quiz alternatives.

## Implementation Notes

T7.1 is `6d12928d88911fa896fba7ddb1c2dc025fe536c8`. T7.2 is
`552751e3b451550946d842cc36027aad39b85c50`, corrected by
`4fa89c658afee30b98e5499cd352633988ab5aec` and `8958f0200940db15a1e17b19d3a20591897fc788`.
T7.3 is `15fbe9b97c82dec1f13599edf1b4142080be992b`; T7.4 is `8466be11507a83418ca18b1d5b8bbefc7c9cbfc0`.

The first regressions reproduced Fast being silently forced to Deep and an
absent session lacking its observation revision. The latter permitted the
adapter to fall back to observing a newer state during execution. The read model
now captures an absent revision, and the terminal adapter rejects unbound
mutations rather than using that legacy service fallback.

Active question, saved history and binding come from one bounded canonical
observation. A neutral application projection owns answer filtering and ordering;
Web and TUI map it independently. One terminal answer component captures its
question and history when editing begins. This is the component owner for later
guided binding in the recorded plan; no parallel wizard editor is added here.
The existing operation context hashes input and binding, while canonical receipts
digest the complete action and answer. A saved action whose catalog refresh
fails remains an eligible original retry. No persistence schema changes.

A narrow terminal regression exposed a final history page containing only a
separator. Its unchanged final-answer/input assertion passes after trimming that
separator. The initial task commit ran despite the failure; the correction is
recorded separately, and subsequent validation/commit sequences stop on failure.
The contextual palette PTY initially pasted the colon and command together;
waiting for the real palette before entering its filter corrects driver mechanics
without changing assertions or existing PTY deadlines.

T7.4 uses the approved pre-action unavailable alternative. The existing
`runAuthorContextualQuizNew` CLI owns a distinct topic/target/provider-backed
contextual document and authoring session pipeline. The central TUI artifact
executor has no supported bridge for it. It is not relabelled as a runtime quiz;
its CLI capability remains, and no new generator or domain workflow is invented.

## Contracts Added or Changed

Start and controls retain explicit observation bindings; delayed commands capture
immutable requests rather than re-reading selection. The canonical owner still
checks revision/session/question and idempotent receipt digest. The terminal
answer buffer is bounded to the existing document capacity, with canonical byte
limits enforced on writes. History reads retain the existing bounded observation
and source images. Adapters remain thin, and service files remain transport-neutral.

## Files of Interest

- `internal/tui/quiz_mode.go`, `quizzes.go`: visible mode, pending markers,
  captured Start and scoped navigation.
- `internal/app/service_quiz_read.go`, `quiz_answer_history.go`,
  `tui_quiz_history.go`, `web_quiz.go`: consistent observation and neutral history.
- `internal/tui/quiz_actions.go`, `quiz_history.go`, `draft_recovery.go`: captured
  editor, paging, controls and eligible confirmed recovery.
- `internal/app/tui_adapters_dump_quiz.go`, `tui_operation_feedback.go`:
  explicit binding guard and acknowledged-action retry mapping.
- `internal/app/tui_quiz_*test.go`, `internal/tui/quiz_*test.go`: real retained
  IO, queued conflict/duplicate execution, input identity and production PTY.
- User Guide, TUI reference, action matrix and changelog: delivered controls.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed.
- `go test ./internal/app -run '^TestTUIQuizStartMode$' -count=1 -timeout=3m` — passed after reproducing and correcting forced Deep.
- `go test ./internal/app -run '^Test(TUIQuizHistory|WebQuizHistory|ServiceListQuizzesReturnsScopedReadModels)$' -count=1 -timeout=3m` — passed with retained entries and shared order.
- `go test ./internal/tui -run '^Test(QuizHistoryPaging|Model.*Quiz.*|Operation.*)$' -count=1` — passed after the final-page correction.
- `go test ./internal/app -run '^Test(TUIQuizDraft|TUIQuizBinding|TUIQuizAdapterDispatchesRuntimeQuizActions)$' -count=1 -timeout=3m` — passed with actual conflicts and duplicate reconciliation.
- `go test ./internal/app -run '^Test(TUIQuiz.*|WebQuiz(Actions|History|StartResume|ReadOnly))$' -count=1 -timeout=3m` — passed before the final unavailable-entry extension.
- `go test ./internal/app -run '^TestTUIQuizTerminal$' -count=1 -timeout=3m` — passed after the palette driver correction, including unavailable-entry no-write assertions and the complete scoped quiz journey.
- `go test ./internal/tui -run '^Test(Quiz.*|Model.*Quiz.*|Operation.*)$' -count=1` — passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.

The real PTY uses the production root loader and managed host with deterministic
model output/runtime binding only. Revision checks, terminal input, canonical
files, cancellation and process cleanup are real. Assertions retain all saved
answers, unchanged files after cancelled/stale actions and no writes to other
quiz scopes. No acceptance assertions or resource limits were weakened.

Controller validation `JRP44LHIFOTCAKMSHR3ELP542E` passed all four ordered checks
on clean HEAD `abc766b3e15ed14cd16d215ac823d1a16d25e582`:

- `make verify`
- `go test -race ./internal/tui -run '^Test(Quiz.*|Model.*Quiz.*|Operation.*|BusyQuit)$' -count=1`
- `go test -race ./internal/app -run '^Test(TUIQuiz(StartMode|History|Binding|Draft|Terminal|AdapterDispatchesRuntimeQuizActions)|WebQuiz(Actions|History|StartResume|ReadOnly|ContinuationFailure|ObservationProjection)|TUIOperation.*|ServiceListQuizzesReturnsScopedReadModels)$' -count=1 -timeout=3m`
- `git diff --check`

Canonical Forgejo and Acta event `01M640NYJZXJJPGES9QRX8QM8R` confirm
PR #461 integrated at the validated commit. No checks were repeated for closure.

## Risks and Follow-ups

None recorded.
