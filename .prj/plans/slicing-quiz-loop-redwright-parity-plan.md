# Quiz Loop RedWright Parity Slicing Plan

Status: active

## Purpose

Plan a short corrective slicing round for the TUI quiz loop so question
generation, routing, persistence, and validation follow the established
RedWright-style behavior instead of drifting into heavyweight, specification-like
questions.

This slicing must preserve the existing service-owned architecture: adapters
emit user intent, `internal/chcapp` coordinates operations, and `internal/chc`
owns quiz runtime, artifacts, model payloads, and provider interpretation.

## Source

Primary sources:

- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.agents/SLICES.md`
- `.agents/CONVENTIONS.md`
- `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
- `internal/chc` quiz runtime, model payload, repository, and provider code
- `internal/chcapp` quiz service operations
- `internal/chctui` quiz interaction surface
- RedWright quiz behavior and artifacts as the parity reference

Conversation decisions:

- The TUI must ask one useful question at a time.
- The default flow must not pre-generate a large batch of questions.
- The model must not push the user into writing a complete spec in the first
  answer.
- The fix must not rely on brittle bans such as "do not use parentheses" or
  "do not enumerate".
- Question generation must use formal request and response payloads.
- Follow-up routing, depth, branch return, and completion semantics must be
  preserved across turns.
- Dumps, project context, use-case context, and agent context remain valid
  induction sources, but the next-question request must stay focused.
- Fast mode remains out of this corrective pass unless required by validation
  evidence.

## Current State

The TUI quiz loop is usable, but generated questions still diverge from the
expected behavior.

Observed problems:

- Questions can become verbose and heavy to read.
- Questions can overuse parenthetical qualifiers and option-like structures.
- The first question can ask for broad implementation detail instead of opening
  a focused discovery path.
- Existing runtime state can make it appear that many questions were loaded in
  advance.
- Follow-up depth and routing metadata may not survive rehydration with the
  same fidelity as the original model response.

Important constraints:

- No database-backed persistence exists.
- Quiz state and artifacts remain file-backed.
- Adapters must not own quiz business logic.
- Runtime code changes must stay behind the service boundary.
- The implementation must be proven with deterministic tests before manual TUI
  verification.

## Target Model

The target model keeps a strict model boundary:

- A next-question operation builds one formal semantic request payload.
- The provider receives only the context required for that operation.
- The provider returns one formal semantic response payload.
- The response preserves intent metadata needed for later routing decisions.
- Runtime state persists the active question and the routing metadata needed to
  decide whether to follow up or return to the trunk.
- Subsequent turns reconstruct the same semantic state instead of approximating
  it from display-oriented fields.

The target user behavior is:

- Select a project or use-case quiz.
- Start answering without reading operational noise.
- Submit one answer.
- See clear activity while the model is working.
- Receive the next focused question.
- Avoid large preloaded question batches in the default flow.
- Complete the session into curated quiz artifacts and later draft-spec
  generation.

## Slice Map

Slicing name: Quiz Loop RedWright Parity

Active slice number: Slice 3

Active slice short name: quiz routing parity control

Expected branch name: `feat/slice-3-quiz-routing-parity-control`

Expected PR title: `feat(slice-3): quiz routing parity control`

Expected report path:
`.prj/reports/slices/quiz-loop-redwright-parity/slice-3-quiz-routing-parity-control.md`

Prerequisite PRs or delivered work:

- TUI Interactive Quiz Workflows through PR #190.
- Post-slice TUI quiz UX fixes through the latest `dev -> main` alignment.
- Service-owned quiz operations, file-backed runtime state, scoped read models,
  direct LLM question generation, and TUI answer flow already delivered.
- Slice 1 quiz provider parity fixtures through PR #193.
- Slice 2 quiz prompt intent persistence through PR #194.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | quiz provider parity fixtures | `feat/slice-1-quiz-provider-parity-fixtures` | `feat(slice-1): quiz provider parity fixtures` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-1-quiz-provider-parity-fixtures.md` |
| Slice 2 | quiz prompt intent persistence | `feat/slice-2-quiz-prompt-intent-persistence` | `feat(slice-2): quiz prompt intent persistence` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-2-quiz-prompt-intent-persistence.md` |
| Slice 3 | quiz routing parity control | `feat/slice-3-quiz-routing-parity-control` | `feat(slice-3): quiz routing parity control` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-3-quiz-routing-parity-control.md` |
| Slice 4 | quiz parity lab validation | `feat/slice-4-quiz-parity-lab-validation` | `feat(slice-4): quiz parity lab validation` | `.prj/reports/slices/quiz-loop-redwright-parity/slice-4-quiz-parity-lab-validation.md` |

## Slice 1: Quiz Provider Parity Fixtures

Add deterministic fixtures and tests that capture the intended provider
request/response contract for the one-question-at-a-time quiz loop.

Expected behavior:

- Next-question payloads are serialized as focused semantic requests.
- Draft-spec payloads are the only quiz model requests that include broader
  context section material in the provider payload.
- Provider responses decode one active question plus semantic prompt intent.
- Tests cover first question, subsequent question, skip, and completion-shaped
  responses.
- The implementation report records the actual payload contracts used by later
  slices.

Out of scope:

- Changing runtime routing decisions.
- Changing TUI behavior.
- Adding fast mode.

## Slice 2: Quiz Prompt Intent Persistence

Persist and reload the model's prompt intent metadata instead of reconstructing
it from display-oriented runtime fields.

Expected behavior:

- Active question state keeps the formal prompt intent returned by the model.
- Raw capture or runtime state stores the routing metadata required for resume.
- Rehydrated next-question payloads preserve follow-up depth, tags, completion
  hints, parent linkage, and topic intent.
- Tests cover resume from saved runtime state and rehydration from existing quiz
  entries.

Out of scope:

- UI redesign.
- Provider prompt changes not required by the persistence contract.

## Slice 3: Quiz Routing Parity Control

Align the next-question routing decision with the established branch/trunk
behavior.

Expected behavior:

- Vague answers can receive a bounded follow-up.
- Specific answers can advance coverage instead of drilling indefinitely.
- Skipped questions return cleanly to broader coverage.
- Maximum follow-up depth is enforced from persisted intent metadata.
- Tests cover vague answers, specific answers, skip, completion, and trunk
  return behavior.

Out of scope:

- Bulk question generation.
- Fast mode candidate buffering.

## Slice 4: Quiz Parity Lab Validation

Run a focused validation pass against representative project and use-case
contexts and record evidence for whether the loop now behaves like the reference
pattern.

Expected behavior:

- Test-project runtime state can be reset cleanly.
- Project-level and use-case-level quiz starts produce one focused first
  question.
- Answer submission produces a visible activity state and then one next
  question.
- No default path preloads a large batch of questions.
- The report captures payload examples, generated-question samples, validation
  commands, and remaining risks.

Out of scope:

- New product features.
- Redesigning the dashboard analytics surface.

## Validation Strategy

Run targeted tests per slice, expanding only when touched code crosses package
boundaries:

- `go test ./internal/chc`
- `go test ./internal/chcapp`
- `go test ./internal/chctui`
- `make verify`

Manual TUI validation belongs in Slice 4 after deterministic tests prove the
contract.

## Risks and Boundaries

- Provider output quality depends on input context quality, but this slicing
  must first prove the semantic request and routing contracts.
- The work must not hardcode project-specific names or negative prompt bans.
- The work must not introduce a database or a second service implementation.
- The work must not bypass the application service from the TUI.
