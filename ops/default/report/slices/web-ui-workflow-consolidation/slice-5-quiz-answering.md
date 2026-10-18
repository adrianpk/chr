# Slice 5: Focused Quiz Answering

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-5-quiz-answering`
PR: `#402`

## Purpose

Deliver the focused answering and retained Deep/Fast behavior in the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 4 was
integrated through PR #401 at `54fb4f6eeba8407c05893a77c4b54c74668018e6`.

## Delivered Behavior

The focused Web screen shows the scope title, a Deep/Fast selector, one ordinary
read-only question, one multiline answer field and a compact Skip/Answer/Cancel/Done
row. Confirmed answers appear newest first by answered time and stable entry ID;
skips are excluded. Start and Resume open this screen from the existing catalog.

Drafts remain in the browser tab through refresh, mode changes and request errors.
A confirmed Answer clears only that question's draft. Stale reload keeps an earlier
unsent draft separate from the current field. Dirty Skip, Cancel, Done and links
away require explicit discard. Done completes the session; Cancel preserves its
history with cancelled status.

Fast generates one successful flat batch and retains its queue, source hash and
consumed identities. Mode changes retain the current question and suspend/resume
Fast work. Questions covered while Deep is selected are excluded from pending
Fast work on return. Exhaustion permits Done without another generation. Deep retains the
existing bounded adaptive follow-ups. A continuation failure publishes the
confirmed answer with resumable blocked state; retrying its operation cannot
append another entry. Reused answered question identities are rejected.

## Implementation Notes

- T5.1: `c00bbcce643fbc1958bc14184e38b73e63f14bd7` — guarded observations, retained schema and recoverable actions.
- T5.2: `0257336d0c738ad9211047f1e8afde951b1745fa` — shared queue policy and observed TUI bindings.
- T5.3: `793621d7828bb45217c8c9517a8bcb93a3db06a2` — focused Web projection, guarded forms, draft recovery, real HTTP/browser checks and captures.

The quiz repository uses the existing process-safe authoring owner and journal.
Runtime, raw entries, curated notes, metadata and the action receipt publish in
one recovery boundary. Reads reject a pending publication without recovering it;
an explicit action recovers under the owner before comparing its binding.
Web request validation permits that recovery before reading retained runtime.

Schema 0 has an explicit read-only legacy interpretation. Schema 2 retains session
identity and mode work. Raw turn records preserve exact entry IDs, timestamps and
multiline content alongside readable material. Metadata keeps unknown fields.
Each image remains bounded to 1 MiB, inventories to 4096 records and Fast queues
to the 50-question budget. Provider preparation has a 90-second context. Only
journal v2 admits the specific quiz-runtime path.

## Contracts Added or Changed

Every shared interactive mutation compares scope, session, active question and
complete source revision under the owner. Operation IDs bind immutable arguments
and replay the original acknowledgement; stale or changed requests fail.
CLI commands observe before entering the owner; TUI editing/confirmation retains
the displayed binding. The Web requires all binding fields and uses the shared
service. Draft storage is separated by project/scope context, session and question;
it cannot submit or mark an answer automatically.

## Files of Interest

- `internal/core/quiz_action*.go`, `quiz_observation.go`, `quiz_legacy_write.go` and quiz repository/runtime files: versioning, guarded publication and recovery.
- `internal/app/service_quiz*.go` and TUI adapters: shared actions, retained queues and bounded observations.
- `internal/app/web_quiz*.go`, `internal/webui/quiz.go`, `assets/templates/quiz.html` and `assets/static/quiz.js`: focused view, bound forms and transient draft behavior.
- `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: delivered user behavior.

## Validation

Successful development checks:

- `go test -race ./internal/core -run '^TestQuiz(Action|EntryExact|Observation|Artifact|Runtime)' -count=1` — persistence, every journal-image interruption, competing answers, receipt replay, schema/size/symlink guards and exact raw-entry round trips passed.
- `go test -race ./internal/core ./internal/app ./internal/tui ./internal/webui -run 'Quiz' -count=1` — shared action, queue, model, adapter and real HTTP regressions passed.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_QUIZ_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-5" go test ./internal/app -run '^(TestWebQuizBrowser|TestWebDumpEditorBrowser|TestWebBoardBrowser)$' -count=1 -v` — owned quiz journey/captures and affected shared editor/Board journeys passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/webui` — passed with zero issues.
- `git diff --check` — passed.

The browser checks lost acknowledgement and unchanged retry, failed mode request,
refresh, both mode switches, rejected dirty Skip/Done, stale Answer, separate old
draft recovery, newest-first answered history, narrow controls and independent
Done/Cancel. Fast makes one model request across the project journey. Each fixture
owns and stops its isolated Chromium and HTTP server after its checks finish.

The first mode regression found an HTMX disabled-control selector naming an absent
button; selectors now select existing controls. Narrow measurements distinguish a
scrollbar-reduced document width from horizontal page overflow.

The first controller attempt at `b21ce52d6229a3f162de1eea637b206666784e14`
stopped at the function-signature rule before lint or tests. The renamed raw
capture, curated-note and runtime helpers now keep their parameters on the
declaration line. The rule reports one declaration per run and checks
the committed branch diff, so its correction is validated after committing.

Final clean-HEAD development and focused results are recorded through
`looper validate` after the PR-reference follow-up. The controller's check record
binds that exact HEAD and ordered command list.

## Captured Views

- [Desktop Fast](files/slice-5/desktop-fast.png): 1440 × 1100 CSS px, DPR 1, active question, unsent answer and four answered-history cards.
- [Desktop Deep](files/slice-5/desktop-deep.png): same viewport, question, draft and history after switching modes.
- [Narrow Deep](files/slice-5/narrow-deep.png): 390 × 844 CSS px, DPR 1, usable answer controls and one-column history.

Adjacent JSON files retain viewport, DPR, document width, mode, history count and
question measurements. Browser page scale is 1. Narrow document width is 380 px
within the 390 px viewport.

## Risks and Follow-ups

None recorded.
