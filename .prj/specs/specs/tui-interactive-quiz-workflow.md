# TUI Interactive Quiz Workflow

Status: accepted

Model: `.prj/specs/specs/tui-interactive-quiz-workflow-model.md`

Sources:

- `.agents/playbooks/quiz.md`
- `.prj/specs/specs/context-composition-pipeline.md`
- `.prj/specs/specs/tui-authoring-and-quiz-workflow.md`
- `docs/reference/authoring-sessions.md`
- `internal/chc/assisted_authoring.go`
- `internal/chc/assisted_quiz_loop.go`
- `internal/chc/authoring_contextual_quiz.go`
- `internal/chcapp` operation service boundary
- `internal/chctui` Quizzes, Authoring, command palette, and action surfaces

## Purpose

Define the exact interactive quiz behavior that the TUI must provide for
project-wide and use-case-scoped knowledge extraction.

Quizzes are not external Markdown editing flows. A quiz is a bounded,
stateful, one-question-at-a-time interaction that captures answers, preserves a
recoverable raw working record, and produces a curated Markdown artifact.

## Scope Model

Quiz scope follows the same product rule as dumps.

- A quiz with no associated use case belongs directly to the project.
- A quiz with a use-case slug belongs to that use case.
- Each project has one main project quiz.
- Each use case has one main use-case quiz.
- The TUI must treat project and use-case quizzes as first-class scoped
  records, not as generic documents.

## Goals

- Provide a TUI-native interactive quiz flow.
- Keep quiz logic in the application service, not in the TUI model.
- Preserve the same behavior through future CLI, Telegram, Email, or other
  adapters.
- Ask exactly one visible question at a time.
- Support deep adaptive questioning as the default.
- Support fast mode without exposing candidate generation mechanics to users.
- Preserve bounded Ariadne-style follow-up and trunk-return behavior.
- Persist enough state to resume after accidental TUI closure.
- Generate raw and curated Markdown quiz artifacts.
- Keep speculative fast candidates out of final curated quiz material until
  activated.
- Keep all model-backed generation behind the existing model/provider boundary.
- Avoid database-backed quiz persistence.

## Non-Goals

- Do not use an external editor as the primary quiz workflow.
- Do not expose batch, tree, or candidate strategy selection to users.
- Do not create a separate service implementation for quizzes.
- Do not move quiz routing, scope, file layout, or curation rules into the TUI.
- Do not introduce database persistence.
- Do not turn quizzes into raw transcripts.
- Do not require first-class multi-user or permission modeling.
- Do not redesign the whole TUI navigation model.

## Interaction Contract

### Session Initialization

Starting a quiz initializes a service-owned runtime state.

The service must define:

- quiz scope;
- mode;
- maximum question budget;
- maximum follow-up depth;
- active topic or coverage group;
- coverage distribution;
- active trunk question;
- active follow-up depth;
- completion criteria;
- current question;
- raw capture path;
- curated quiz path;
- runtime state cache path.

The default maximum question budget is 50 questions.

The default maximum follow-up depth is 3 levels:

- level 1: primary question;
- level 2: follow-up;
- level 3: deep clarification.

No further nested follow-up may be asked after level 3. The runtime must return
to trunk coverage or terminate.

### Visible Question Flow

The TUI must show one active question.

The user may perform these actions:

- answer;
- skip;
- done;
- cancel;
- switch future generation mode;
- refresh/resume state.

The TUI must not pre-render a list of future questions as the normal
interaction. Future questions may exist internally only as runtime candidates.

### Answer

Answer records the active question and the user's answer.

After answer:

- the raw working artifact is updated;
- the runtime updates coverage, depth, and completion state;
- the service decides whether to ask a follow-up, return to trunk coverage,
  switch topic, or terminate;
- the TUI receives the next runtime view.

Answer must operate on the service-owned active question. It must not trust
question text copied through TUI state as the source of truth.

### Skip

Skip records the active question as intentionally skipped.

After skip:

- the raw working artifact records the skipped question;
- the skipped question may appear in curated notes only when it matters as a
  decision or unresolved issue;
- the runtime advances using the same routing rules as answer.

### Done

Done means the current quiz has enough useful context for now.

After done:

- the runtime stops the active loop;
- the curated quiz artifact is generated or refreshed;
- the quiz remains inspectable and resumable later;
- the status reflects sufficient-for-now rather than permanently closed.

### Cancel

Cancel stops the current loop without claiming sufficient context.

After cancel:

- answered and skipped material already persisted remains valid;
- no new curated decision should be inferred from cancellation;
- runtime state remains resumable unless explicitly discarded later.

### Resume

Resume reconstructs the current quiz view from service-owned state.

The runtime must be recoverable from project-local files after TUI closure.
If runtime state is missing but raw or curated quiz artifacts exist, the service
must degrade to an inspectable quiz state and report that active runtime state
cannot be resumed.

## Modes

### Deep Mode

Deep mode is the default.

Deep mode asks the model/provider boundary for the next question after each
answer, skip, done-relevant transition, or routing point.

Deep mode is authoritative when answers are:

- free text;
- ambiguous;
- unexpected;
- rich enough to require interpretation;
- likely to change follow-up routing.

### Fast Mode

Fast mode is a user-facing speed preference.

Fast mode must not expose internal strategy details. The user chooses fast
versus deep, not batch versus tree.

Fast mode may internally prepare candidate questions. Candidates remain runtime
state until activated. Discarded candidates must not appear in curated quiz
notes, raw answered/skipped entries, retrieval tags, or final context.

Fast mode must fall back to deep mode when deterministic routing is not safe.

### Mode Switching

Mode switching affects future question generation only.

Already recorded answers, skips, raw capture entries, and curated notes remain
stable.

Switching from fast to deep must invalidate or ignore unused fast candidates
immediately. The runtime must not force the user to exhaust fast candidates.

Switching from deep to fast may generate candidates at the next generation
point.

## Ariadne Routing

The runtime must preserve a visible trail between trunk coverage and nested
follow-ups.

The state must distinguish:

- trunk questions;
- follow-up questions;
- deep clarification questions;
- the parent question that opened a follow-up chain;
- the topic or coverage group being explored;
- the reason the runtime returned to trunk coverage.

The runtime must prevent local overfitting. A branch must close when:

- the branch reaches maximum depth;
- the answer is sufficiently clear;
- additional follow-up would not materially improve the artifact;
- broader coverage has become more valuable than local detail.

## Artifact Contract

Quiz artifacts are project-local files under `.chc/`.

The canonical visible artifacts are:

```text
.chc/quizzes/project-quiz-raw.md
.chc/quizzes/project-quiz.md
.chc/quizzes/<case-slug>-quiz-raw.md
.chc/quizzes/<case-slug>-quiz.md
```

The raw file is working memory.

The curated file is the durable quiz artifact used by downstream spec,
planning, indexing, and context assembly workflows.

The runtime state cache lives outside the curated artifact:

```text
.chc/tmp/quiz-runtime/project.json
.chc/tmp/quiz-runtime/<case-slug>.json
```

Model request/response evidence should reuse the existing authoring model
evidence layout when a model call is made:

```text
.chc/authoring/sessions/<session-id>/model/<request-id>/request.md
.chc/authoring/sessions/<session-id>/model/<request-id>/response.md
```

The exact authoring session id is service-owned. The TUI must not construct it.

## Raw Capture Rules

The raw artifact must preserve enough execution context to prevent information
loss during long sessions.

It may include:

- quiz scope;
- active mode at the time of the turn;
- question id;
- question text;
- answer text;
- skipped marker;
- topic or coverage group;
- follow-up depth;
- parent question id;
- created or answered timestamp;
- runtime notes needed for later curation.

The raw artifact is not the final product. It must not be rendered as if it
were the curated quiz.

## Curated Quiz Rules

The curated artifact must contain distilled, decision-relevant quiz notes.

It must not map raw questions one-to-one into the final file.

It should preserve:

- scope;
- topic;
- stable decisions;
- constraints;
- preferences;
- unresolved questions;
- useful implementation implications;
- source provenance to the raw quiz artifact.

It must exclude:

- conversational filler;
- speculative candidates that were never activated;
- discarded branches;
- repeated low-signal Q/A pairs;
- internal routing mechanics unless they explain a real decision.

## Service Boundary

The application service owns all quiz behavior.

The service owns:

- scope resolution;
- path selection;
- runtime state persistence;
- raw capture writes;
- curated artifact writes;
- active question generation;
- fast candidate lifecycle;
- answer, skip, done, cancel, resume, and mode-switch semantics;
- reload/index refresh after writes.

The TUI owns:

- rendering;
- focus state;
- local text input state before submit;
- shortcut handling;
- intent dispatch to the service;
- status display.

The TUI must not parse Markdown to infer quiz runtime state.

## TUI Behavior

The Quizzes screen must show:

- project-wide quiz;
- use-case-scoped quizzes;
- selected quiz status;
- selected scope;
- active mode;
- active question when a loop is running;
- answer input;
- answered count;
- skipped count;
- remaining budget;
- follow-up depth and trail;
- curated preview or full curated body when no active question is running.

The detail panel must support:

- starting a project quiz;
- starting a use-case quiz;
- resuming an active quiz;
- answering the active question;
- skipping the active question;
- marking done;
- cancelling the active loop;
- switching future generation mode;
- refreshing state from files.

The primary editing surface is the TUI answer input. External editor shortcuts
are not part of quiz completion.

## Fast Candidate Rules

Fast candidates are internal runtime data.

They may carry:

- candidate id;
- prompt;
- route id;
- parent id;
- depth;
- topic;
- internal hints;
- generated timestamp.

They must not carry final retrieval tags.

A candidate becomes durable quiz material only when activated and then answered
or skipped.

Unused candidates must be cleaned opportunistically:

- when a sibling branch is selected;
- when mode switches to deep;
- when a quiz is completed;
- when a new candidate set replaces an older set;
- during normal resume/refresh when stale candidates are detected.

## Model Boundary

All model-backed quiz question generation must use the existing model/provider
boundary.

Generation requests must include:

- action;
- strategy;
- scope;
- session id when available;
- target artifact intent;
- current mode;
- active topic;
- coverage state;
- follow-up depth;
- prior answered/skipped material;
- relevant project context;
- output contract.

Generation responses must be structured and validated before they affect
runtime state.

Invalid model responses must block or fail the operation with a clear service
error. The service must not silently replace invalid model output with generic
template questions.

## Acceptance Criteria

- The TUI can start a project-wide quiz.
- The TUI can start a use-case-scoped quiz.
- The TUI asks one visible question at a time.
- The TUI can answer the active question.
- The TUI can skip the active question.
- The TUI can mark a quiz done without making it permanently closed.
- The TUI can cancel a quiz without claiming sufficient context.
- The TUI can resume an interrupted quiz from project-local files.
- Deep mode is the default.
- Fast mode is user-visible only as a speed preference.
- Internal fast candidates do not appear in raw or curated quiz artifacts until
  activated.
- Follow-up depth is capped at 3.
- The runtime returns from follow-up branches to broader coverage.
- Project and use-case quizzes generate the canonical raw and curated files.
- The curated quiz artifact contains distilled insights, not a transcript.
- No quiz behavior requires database persistence.
- TUI code dispatches service intents and does not own runtime semantics.

## Open Questions

- Whether the first fast implementation should generate ten candidates or use a
  smaller default for TUI responsiveness.
- Whether a project can keep archived historical quiz snapshots in addition to
  the one main quiz per scope.
- Whether curated quiz generation should happen synchronously on done or as a
  separate explicit refine action when model cost is high.
