# Slice 5: Adaptive Quiz Loop

Status: delivered
Slicing: Agent-Assisted Authoring Sessions
Plan: `.prj/plans/slicing-agent-assisted-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`
Branch: `feat/slice-5-assisted-quiz-loop`
PR: #103

## Purpose

Run a bounded adaptive question-and-answer loop for one assisted authoring
session.

## Delivered Behavior

The core package now supports:

- asking the next assisted quiz question through the model adapter boundary;
- rendering the current authoring context into the model request;
- recording model request and response artifacts;
- persisting asked, answered, completed, and blocked turns as transcript events;
- accepting one user answer for a previously asked question;
- stopping when the user stops, the turn limit is reached, the follow-up limit
  is reached, or the model returns a blocked response.

## Implementation Notes

This slice keeps the loop host-side and CLI-free. The loop works from explicit
session, config, context package, transcript events, repository, and adapter
inputs. It does not rely on implicit conversation history.

Turn IDs and request IDs are deterministic for testability:

```text
turn-001
request-001
```

Model invocation artifacts are written through the Slice 4 contract under the
session model directory.

## Contracts Added or Changed

New internal contracts:

- `AssistedQuizLoopOptions`
- `AssistedQuizAskResult`
- `AssistedQuizAnswerOptions`
- `AssistedQuizAnswerResult`
- `AssistedQuizLoopStopReason`
- `AskNextAssistedQuizQuestion`
- `AnswerAssistedQuizQuestion`

## Files of Interest

- `internal/chc/assisted_quiz_loop.go`
- `internal/chc/assisted_quiz_loop_test.go`
- `.prj/trackers/slicing-agent-assisted-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestAskNextAssistedQuizQuestion|TestAnswerAssistedQuizQuestion'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice does not curate answers, synthesize final artifacts, expose CLI
commands, or call a real provider. Those remain assigned to later slices in the
agent-assisted authoring slicing set.
