---
id: TKT-20261017193927
title: "Provide a focused quiz answering screen with modes and answer history"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: quizzes, interaction, quiz-modes
source: chat
reported_at: 2026-10-17T19:39:27Z
commits: [c00bbcce643fbc1958bc14184e38b73e63f14bd7, 0257336d0c738ad9211047f1e8afde951b1745fa, 793621d7828bb45217c8c9517a8bcb93a3db06a2]
branch: feat/slice-5-quiz-answering
pr: "#402"
resolution: fixed
reviewed_at: 2026-10-18T17:54:53Z
closed_at: 2026-10-18T17:54:53Z
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Provide a focused quiz answering screen with modes and answer history

## Context and Authority

Part of [the Web UI review](../in_progress/20261017105432-review-draft-web-ui-usability.md).
The maintainer supplied a visual example and explicitly described the desired
conceptual layout and behavior for CHR's quiz answering screen. The target
below is self-contained. It does not require another application as a design
source, a matching palette or identical decorative details.

CHR's catalog has been visually reviewed; its detailed answering view has not
yet been demonstrated in this walkthrough. This ticket records the requested
target and verified adapter capabilities, rather than inventing observations
about an unseen CHR detail screen.

Approved implementation is Slice 5 of
[web-ui-workflow-consolidation](../../plan/web-ui-workflow-consolidation.md).
The [Slice 5 report](../../report/slices/web-ui-workflow-consolidation/slice-5-quiz-answering.md)
records retained actions, mode queues, focused interaction and validation.

Catalog entry, card membership and pending scope selection are owned by
[the quiz catalog ticket](../in_progress/20261017192913-show-quiz-cards-and-pending-scope-action.md).

## Conceptual Layout

```text
Quizzes
Selected quiz or scope                          [ Deep | Fast ]

+------------------------------------------------------------+
| Current question                                           |
+------------------------------------------------------------+
|                                                            |
| Answer...                                                  |
| Large editable response area                               |
|                                                            |
|                         [ Skip ] [ Answer ] [ Cancel ] [ Done ]
+------------------------------------------------------------+

[ Most recent answered question ] [ Previous answered question ]
[ Older answered question ... ]
```

1. Show a clear quiz or scope title and a simple route back to the catalog.
2. Put the Deep/Fast selector at the top right, aligned with the title region.
   Show the selected mode clearly and allow changing it directly in place.
3. Present one current question in a readable, non-editable question area.
   Show the question as ordinary text rather than a file or model envelope.
4. Place one large multiline answer input below that question. The question
   and answer have distinct roles and contents; do not duplicate the same
   body in an additional read-only preview.
5. Group Skip, Answer, Cancel and Done in one compact action row below the
   answer area, aligned toward the right. Avoid separate operation forms
   spread across the page or requiring the scope to be entered again.
6. Keep the interaction focused on the current question and typed response.
   Scope, retained session and required mutation bindings come from the
   selected quiz, not a manually typed case slug.

## Action Semantics

| Action | Required user-visible result |
| --- | --- |
| Skip | Pass the current question without recording the response field as an answer; advance according to the quiz's remaining questions or generation policy. |
| Answer | Submit the text entered for the current question, retain that answered question, and move to the next question or an applicable follow-up. |
| Cancel | Exit the quiz interaction. Do not mark it completed merely because the user left. Preserve already recorded answers under the existing cancellation contract. |
| Done | Explicitly close the quiz as done even when more questions could be asked or remain queued. Do not require exhausting the question set first. |

Cancel and Done have distinct domain outcomes. Neither action should silently
submit an unsent draft as though Answer had been selected. Resolve the handling
of a typed but unsent draft explicitly during refinement rather than silently
losing it through a mode change or another routine operation.

## Answered Question History

1. As questions are answered, display them below the current-question panel
   as visual cards, newest answer first and oldest last.
2. Each history card shows the question and its recorded answer. Keep the
   structure and size consistent enough to scan previous responses.
3. The current question remains in the main panel. Do not mix pending questions
   or model metadata into the answered-history cards.
4. Retain the history after navigation, page reload and continuation through
   the existing durable quiz owner. Browser-only state is insufficient.
5. Preserve previous answers when changing modes or closing the quiz.

## Deep and Fast Behavior

1. Deep supports adaptive follow-up questions and nested exploration when the
   existing answers require clarification or more depth.
2. Fast generates the flat question set in one LLM generation request, without
   nested follow-ups. Moving through an already generated Fast queue must not
   request generation again for each answered or skipped question.
3. Changing mode is a direct interaction with the selector at the top right.
   Do not require returning to the catalog, filling a separate mode form or
   restarting the quiz.
4. Mode is an actual quiz behavior, not only a visual label. Apply it through
   the shared quiz service and preserve the selected scope and answer history.
5. Define the handling of the active question, queued questions and unsent
   response when switching modes. Preserve completed work and avoid silently
   regenerating or discarding answered questions. The exact queue transition
   policy remains a refinement item.

## Existing Application Boundaries

Code inspection confirms that the shared operation service already exposes
AnswerQuiz, SkipQuiz, CompleteQuiz, CancelQuiz and SwitchQuizMode, while the
current Web action adapter wires quiz start and resume. The implementation
should connect the answering screen to shared operations and verify their
behavior, rather than duplicate the quiz lifecycle in a Web-only controller
or invent a parallel browser-owned session.

Relevant owners:

- `internal/app/service_quiz.go`: quiz actions, mode transitions, generation
  and completion.
- `internal/app/tui_adapters_dump_quiz.go`: existing mappings from quiz actions
  to shared operations.
- `internal/core/quiz_runtime.go` and model/repository owners: retained quiz
  state, entries, question queues and model behavior.
- `internal/app/web_actions.go`: current Web start/resume wiring and shared
  write result handling.
- `internal/webui/assets/templates/shell.html`: current quiz forms and shared
  presentation shell.

Reuse the normal server-side access, mutation and stale-state safeguards.
Keep pending answers understandable while a request or model generation is in
progress, and prevent repeated submission of the same question from producing
duplicate answered entries.

## Acceptance

1. Opening or continuing a selected quiz reaches the focused answering screen
   with its own title, current question, one answer field and the four actions.
2. The top-right selector switches Deep/Fast directly in the active quiz and
   reflects the retained mode after reload.
3. Answer retains the correct question/response and updates history below the
   form, ordered newest first. Skip advances without inventing an answer.
4. Cancel exits with cancellation semantics; Done closes successfully while
   unasked or queued questions still exist. Recorded history survives both.
5. Validate Deep follow-up behavior and instrument Fast question generation to
   establish one request for its generated flat set. A selector label alone
   is not evidence of either behavior.
6. Validate mode changes, page reload, continuation, pending-draft handling and
   repeated/stale submissions through the shared operation boundary.
7. Capture visual evidence of an active question with an answer draft, both
   selected modes and several history cards. Verify the conceptual hierarchy
   and action placement above; palette matching is not an acceptance condition.
8. Keep this work separate from catalog membership and presentation. Run
   focused adapter/service and browser checks only when implementation is
   explicitly authorized.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).

## Integration

PR #402 merged into dev at `303db92f10651f5eb36d730e25a742ff78d06cd6` on 2026-10-18. Controller validation `PFMCWI6B2IAURCUFL76I2UB6DZ` passed `make verify`, focused quiz race checks, the quiz/editor/Board browser selection and `git diff --check` at that exact HEAD.
