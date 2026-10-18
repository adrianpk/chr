# Slice 5: Focused Quiz Answering

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-5-quiz-answering`
PR: pending

## Purpose

Deliver focused answering and retained Deep/Fast behavior from the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
Slice 4 is integrated. T5.1 establishes guarded, recoverable quiz state and actions;
T5.2 applies the queue policy; T5.3 exposes the focused Web interaction.

## Delivered Behavior

T5.1 implements versioned runtime observations, immutable action bindings and
recoverable publication of runtime, raw capture, curated notes, metadata and
operation receipts under the existing authoring owner. T5.2 retains one successful Fast batch across mode switches and process reloads,
keeps the active question through switches, and consumes the queue without
per-question generation. Deep retains adaptive follow-ups. Provider continuation
failure retains the confirmed answer in a resumable blocked state. Cancel and
Done preserve history with distinct statuses. The Web interaction remains in
progress.

## Implementation Notes

T5.1: `c00bbcce643f` — versioned persistence, guarded actions and journal recovery.
T5.2: shared service queue policy and observed TUI action bindings.


Reuse the quiz repository, existing model boundary and shared authoring owner.
Retain versioned session/question and action bindings, entries and mode queues
in one recovery boundary. Model and browser fixtures remain deterministic.

## Contracts Added or Changed

Quiz runtime schema 2 retains session identity and suspended Fast work. Schema 0
reads retain the legacy interpretation and leave stored bytes unchanged. Actions
compare scope, session, question and complete source revision under the shared
owner; identical retries replay the committed receipt, while stale or changed
arguments fail. Only journal v2 admits the bounded quiz-runtime path. Raw turn
records retain exact entry IDs, timestamps and multiline content.

## Files of Interest

- `internal/core/quiz_runtime.go` and quiz repository owners: retained state and recovery.
- `internal/app/service_quiz.go`: shared actions and generation policy.
- `internal/webui`: focused server-rendered interaction and transient draft retention.

## Validation

Task validation:

- `go test -race ./internal/core -run '^TestQuiz(Action|EntryExact|Observation|Artifact|Runtime)' -count=1` — passed, including every journal-image interruption, competing answers, receipt replay, legacy read stability, unsupported schema, size and symlink rejection.

- `go test -race ./internal/core ./internal/app ./internal/tui -run 'Quiz' -count=1` — passed for shared actions, queues, continuation failure, Deep routing, model contracts and adapter regressions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/tui` — passed with zero issues.

Web/browser checks and the final exact clean-HEAD development gate remain pending.

## Risks and Follow-ups

The controller owns integration into dev. No live quiz or model is invoked.
