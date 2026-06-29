# Gallery

Visual examples of Chryso in use.

This gallery follows a seeded reference project through context, specification,
planning, board state, ticket execution, review state, and runtime proof.

## Context

![Project context](../../assets/img/gallery/chr-agent-work-03-project-context.png)

The project view keeps workspace context and generated artifacts visible.

![Use case CLI gameplay](../../assets/img/gallery/chr-agent-work-04-use-case-cli-gameplay.png)

Use cases keep feature work scoped before it becomes implementation work.

![Dump CLI gameplay context](../../assets/img/gallery/chr-agent-work-05-dump-cli-gameplay-context.png)

Dumps preserve rough context in natural language.

![Quizzes list](../../assets/img/gallery/chr-agent-work-06-quizzes-list.png)

Quizzes capture focused answers without turning the user into a spec machine.

## Specification

![Spec CLI gameplay](../../assets/img/gallery/chr-agent-work-07-spec-cli-gameplay.png)

Specs turn captured context into expected behavior and acceptance boundaries.

![Plan sliced implementation](../../assets/img/gallery/chr-agent-work-08-plan-sliced-implementation.png)

Plans slice the work into tickets that the agent loop can process.

## Board

![Board empty backlog column](../../assets/img/gallery/chr-agent-work-09-board-empty-backlog-column.png)

The board exposes ticket state as the implementation queue changes.

![Board ready dev columns](../../assets/img/gallery/chr-agent-work-10-board-ready-dev-columns.png)

Generated tickets appear on the board once planning has produced implementable
work.

## Dash

![Dash ready queue](../../assets/img/gallery/chr-agent-work-02-dash-ready-queue.png)

Dash shows the ready queue before execution starts.

![Dash ready queue before run](../../assets/img/gallery/chr-agent-work-11-dash-ready-queue-before-run.png)

Dash shows the queue before execution begins.

![Help placeholder](../../assets/img/gallery/chr-agent-work-12-help-placeholder.png)

The runtime surface includes a place for operational help and hints.

![Dash work running start](../../assets/img/gallery/chr-agent-work-14-dash-work-running-start.png)

Once running, Dash reports the active task and current execution state.

![Dash work running later](../../assets/img/gallery/chr-agent-work-15-dash-work-running-later.png)

The execution surface keeps progress visible during a long agent run.

![Dash loop succeeded](../../assets/img/gallery/chr-agent-work-16-dash-loop-succeeded.png)

Successful completion returns the loop result to the operator.

## Review

![Board ready for review columns](../../assets/img/gallery/chr-agent-work-17-board-ready-for-review-columns.png)

Completed work moves into review state on the board.

## Runtime Proof

![Terminal game in progress](../../assets/img/gallery/chr-agent-work-20-terminal-game-in-progress.png)

Runtime behavior provides evidence beyond a static diff.

![Terminal game win](../../assets/img/gallery/chr-agent-work-21-terminal-game-win.png)

The completed reference app reaches a successful gameplay state.

## Previous Version

The earlier gallery is preserved in [Old Gallery](old-gallery.md).
