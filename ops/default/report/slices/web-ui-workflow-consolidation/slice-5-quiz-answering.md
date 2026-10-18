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

Implementation in progress.

## Implementation Notes

Reuse the quiz repository, existing model boundary and shared authoring owner.
Retain versioned session/question and action bindings, entries and mode queues
in one recovery boundary. Model and browser fixtures remain deterministic.

## Contracts Added or Changed

Pending guarded runtime and action contracts.

## Files of Interest

- `internal/core/quiz_runtime.go` and quiz repository owners: retained state and recovery.
- `internal/app/service_quiz.go`: shared actions and generation policy.
- `internal/webui`: focused server-rendered interaction and transient draft retention.

## Validation

Pending focused persistence, concurrency, service and browser evidence, followed
by Looper's exact clean-HEAD development gate.

## Risks and Follow-ups

The controller owns integration into dev. No live quiz or model is invoked.
