# TUI Interactive Quiz Workflows Slicing Plan

Status: active

## Purpose

Plan the implementation of service-owned, TUI-native interactive quiz workflows
for project-wide and use-case-scoped knowledge extraction.

The goal is to reproduce the established quiz behavior and artifact semantics:
one visible question at a time, bounded depth, raw working capture, curated
quiz notes, deep and fast generation modes, Ariadne-style branch/trunk routing,
and resumable runtime state.

## Source

Primary sources:

- `.agents/playbooks/quiz.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.agents/SLICES.md`
- `.prj/specs/specs/tui-interactive-quiz-workflow.md`
- `.prj/specs/specs/tui-interactive-quiz-workflow-model.md`
- `.prj/specs/specs/context-composition-pipeline.md`
- `docs/reference/authoring-sessions.md`
- `internal/chc` assisted authoring and contextual quiz packages
- `internal/chcapp` operation service boundary
- `internal/chctui` Quizzes, Authoring, command palette, and action surfaces

Conversation decisions:

- Quizzes are not an external editor workflow.
- Quiz handling must go through the application service.
- The TUI must not own routing, scope resolution, file layout, curation, or
  question generation.
- Project-level quiz scope has no use case.
- Use-case-level quiz scope uses the selected use-case slug.
- Fast mode is a speed preference, not a visible batch/tree strategy selector.
- Deep mode is the default and remains authoritative for ambiguous free text.
- Runtime state must survive accidental TUI closure through project-local files.
- Quiz persistence must remain file-backed.
- LLM-backed question generation must receive agent protocol context, project
  context, use-case context when scoped, and relevant dump context, including a
  use-case dump when one exists.
- A complete quiz plus relevant dump context must be promotable into a spec
  artifact with `status: draft`.
- Model-backed quiz and synthesis operations must use one formal request and
  response boundary instead of ad hoc feature-specific LLM call paths.

## Current State

The TUI has a `Quizzes` view, but it currently renders generic document-family
content rather than a quiz-specific workflow.

Important limitations:

- There is no quiz-specific read model.
- There is no service-owned interactive quiz operation interface.
- The TUI cannot start or resume a scoped quiz loop.
- The TUI cannot answer, skip, complete, or cancel an active quiz.
- Current context-aware quiz generation creates Markdown question sets, but it
  does not provide the required one-question-at-a-time TUI runtime.
- Existing assisted authoring primitives provide useful turn, transcript, and
  model-boundary concepts, but they do not yet expose the full scoped quiz
  runtime required by this spec.
- The Quizzes screen cannot distinguish raw capture, curated quiz notes,
  runtime state, candidates, or active questions.

## Target Model

The target model is service-centered:

- `internal/chcapp.service` owns quiz operations.
- `internal/chc` owns quiz runtime types, validation, file-backed state, raw
  capture rendering, curated artifact rendering, and candidate lifecycle.
- The service resolves project-wide versus use-case-scoped quiz requests.
- The service owns canonical `.chc/quizzes` and `.chc/tmp/quiz-runtime` paths.
- The service returns quiz read models and operation results to adapters.
- TUI adapters translate keyboard and command-palette input into service calls.
- The TUI renders service-provided quiz state and emits user intent only.

The target TUI behavior is:

- `F4 Quizzes` shows project-wide and use-case-scoped quizzes.
- The selected quiz shows status, scope, mode, budget, answered/skipped counts,
  trail, active question, or curated notes.
- The user can start or resume the selected quiz.
- The user can answer or skip exactly one active question at a time.
- The user can complete or cancel a loop with distinct semantics.
- The user can switch future generation mode between deep and fast.
- Refresh reflects direct filesystem edits and runtime-state changes.

## Slice Map

Slicing name: TUI Interactive Quiz Workflows

Active slice number: Slice 3

Active slice short name: quiz service operations

Expected branch name: `feat/slice-3-quiz-service-operations`

Expected PR title: `feat(slice-3): quiz service operations`

Expected report path:
`.prj/reports/slices/tui-interactive-quiz-workflows/slice-3-quiz-service-operations.md`

Prerequisite PRs or delivered work:

- Application operation service consolidation slicing.
- TUI Dump Workflows through PR #181.
- External dump editor shell-launch fix through PR #182.
- Workspace index and semantic cache slicing through PR #172.
- Slice 1 quiz runtime model through PR #183.
- Slice 2 quiz artifact repository through PR #184.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | quiz runtime model | `feat/slice-1-quiz-runtime-model` | `feat(slice-1): quiz runtime model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-1-quiz-runtime-model.md` |
| Slice 2 | quiz artifact repository | `feat/slice-2-quiz-artifact-repository` | `feat(slice-2): quiz artifact repository` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-2-quiz-artifact-repository.md` |
| Slice 3 | quiz service operations | `feat/slice-3-quiz-service-operations` | `feat(slice-3): quiz service operations` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-3-quiz-service-operations.md` |
| Slice 4 | scoped quiz read model | `feat/slice-4-scoped-quiz-read-model` | `feat(slice-4): scoped quiz read model` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-4-scoped-quiz-read-model.md` |
| Slice 5 | tui deep quiz loop | `feat/slice-5-tui-deep-quiz-loop` | `feat(slice-5): tui deep quiz loop` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-5-tui-deep-quiz-loop.md` |
| Slice 6 | fast quiz candidates | `feat/slice-6-fast-quiz-candidates` | `feat(slice-6): fast quiz candidates` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-6-fast-quiz-candidates.md` |
| Slice 7 | ariadne trail recovery | `feat/slice-7-ariadne-trail-recovery` | `feat(slice-7): ariadne trail recovery` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-7-ariadne-trail-recovery.md` |
| Slice 8 | quiz curation polish | `feat/slice-8-quiz-curation-polish` | `feat(slice-8): quiz curation polish` | `.prj/reports/slices/tui-interactive-quiz-workflows/slice-8-quiz-curation-polish.md` |

## Slice 1: Quiz Runtime Model

Add the core quiz runtime models and validation rules.

Expected behavior:

- Runtime state represents scope, mode, status, active question, budget,
  coverage, Ariadne trail, and fast candidates.
- Scope rules distinguish project-wide and use-case-scoped quizzes.
- Validation rejects invalid status, mode, depth, budget, and answer state.
- No persistence or TUI behavior changes are introduced yet.

Out of scope:

- File writes.
- Service operations.
- TUI rendering.
- Model/provider integration.

## Slice 2: Quiz Artifact Repository

Add file-backed raw, curated, and runtime-state artifact handling.

Expected behavior:

- Project and use-case scopes resolve to canonical `.chc/quizzes` paths.
- Runtime state resolves to canonical `.chc/tmp/quiz-runtime` paths.
- Raw capture can be created, appended, and loaded.
- Curated quiz notes can be written from distilled quiz entries.
- Runtime JSON state can be saved and loaded.
- Speculative candidates remain runtime state only.

Out of scope:

- TUI interaction.
- Real model-backed next-question generation.
- Full service orchestration.

## Slice 3: Quiz Service Operations

Expose service-owned quiz operations over the application service boundary.

Expected behavior:

- The service can list, start, resume, answer, skip, complete, cancel, switch
  mode, and refresh quizzes.
- Answer and skip operate on service-owned active question state.
- Done and cancel have distinct semantics.
- Deep mode is the default.
- The service writes raw capture and runtime state after each operation.
- The service refreshes quiz read state after writes.
- Model-backed question generation receives augmented context from agent
  protocol files, project context, scoped use-case context, and relevant dumps.
- Model-backed question generation uses formal semantic request and response
  payloads.

Out of scope:

- Full TUI interactive panel.
- Fast candidate generation.
- Final UX polish.

## Slice 4: Scoped Quiz Read Model

Add a quiz-specific read model for the TUI.

Expected behavior:

- Quizzes are listed by project and use-case scope.
- The read model carries title, status, scope, mode, paths, counts, updated
  time, active question summary, and curated preview.
- The Quizzes screen renders project-wide and use-case-scoped quizzes without
  parsing Markdown in the TUI.
- Empty states distinguish no quiz, idle quiz, active quiz, and blocked quiz.

Out of scope:

- Answer input.
- Fast mode controls.
- Curation polish.

## Slice 5: TUI Deep Quiz Loop

Add the first interactive one-question-at-a-time TUI quiz loop in deep mode.

Expected behavior:

- The user can start or resume a selected scoped quiz.
- The active question renders in the detail panel.
- The answer input is TUI-native.
- The user can answer, skip, done, cancel, and refresh through service intents.
- The TUI displays budget, answered/skipped counts, status, and service
  messages.
- The TUI does not own routing decisions.

Out of scope:

- Fast candidate generation.
- Advanced trail visualization.
- External editor support.

## Slice 6: Fast Quiz Candidates

Add fast mode as a service-owned future-generation preference.

Expected behavior:

- The user can switch future generation mode between deep and fast.
- Fast candidate generation uses the same formal model request/response
  boundary as deep generation.
- Fast mode can prepare internal candidates.
- Activated candidates become normal active questions.
- Discarded candidates do not appear in raw capture, curated notes, or tags.
- Ambiguous/free-text answers fall back to deep generation.
- Candidate cleanup is local and opportunistic.

Out of scope:

- Exposing batch/tree strategy selection.
- Context ranking based on fast/deep mode.
- Background workers.

## Slice 7: Ariadne Trail Recovery

Make branch/trunk state and interruption recovery explicit in the TUI.

Expected behavior:

- The runtime records trunk question, parent question, depth, topic, and return
  reason.
- The TUI renders a compact trail for nested follow-ups.
- Follow-up depth is capped at 3.
- The runtime returns to broader coverage after branch closure.
- Resume reconstructs active question and trail from runtime state.
- Missing runtime state degrades to inspectable raw/curated quiz state.

Out of scope:

- Changing the overall TUI navigation model.
- Adding historical archive snapshots.

## Slice 8: Quiz Curation Polish

Finalize curated quiz notes and downstream artifact readiness.

Expected behavior:

- Done refreshes curated quiz notes from raw capture and entries.
- Curated notes retain decisions, constraints, preferences, open questions, and
  implementation implications.
- Curated notes do not mirror raw turns one-to-one.
- The Quizzes screen can show the curated body when no active question is
  running.
- The generated artifacts are ready for spec creation and context assembly.
- A complete quiz plus relevant dump context can generate a spec artifact with
  `status: draft`.
- Draft spec generation uses the same formal model request/response boundary as
  quiz question generation.
- Documentation is aligned with the implemented quiz contract.

Out of scope:

- Automatic accepted/final spec generation.
- Historical quiz archive UX.
- Telegram or Email adapters.

## Global Exit Criteria

- Project-wide and use-case-scoped quizzes work from the TUI.
- Quiz logic is owned by the application service.
- The TUI asks one visible question at a time.
- Raw and curated quiz artifacts are generated in canonical paths.
- Runtime state survives TUI closure through project-local files.
- Deep mode is the default.
- Fast mode works as a speed preference without exposing internals.
- Ariadne-style follow-up and trunk-return behavior is represented and bounded.
- No database persistence is introduced.
- All slice PRs target `dev` and are reviewed independently.

## Validation Command Context

- Targeted Go tests should be run for touched packages.
- Broad validation should use `make verify` before each slice PR.
- PR validation entries must list only the commands executed.
- PR validation entries must not use `pass:` or `(pass)`.

## Non-Goals

- Do not implement more than one slice per PR.
- Do not open a PR before the slice report exists and is set to `reviewing`.
- Do not add external-editor quiz behavior.
- Do not add database-backed quiz persistence.
