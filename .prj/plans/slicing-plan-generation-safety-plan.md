# Plan Generation Safety Slicing Plan

Status: active

## Purpose

Plan the hardening of project-wide and use-case plan generation so expensive
model-backed writes are explicit, recoverable, concurrency-safe, and protected
from stale archived context.

The goal is to make plan generation safe enough for Markdown-first operation:
creating a missing plan remains simple, replacing an existing plan requires
explicit intent, replaced plans are archived, archived plans do not influence
future model context, and long-running generation cannot accidentally fan out
into duplicate concurrent requests.

## Source

Primary sources:

- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.agents/SLICES.md`
- Existing `internal/chcapp` plan generation service code
- Existing `internal/chctui` `F6 Plans` action surface
- Existing `.chc/plans/<scope>-plan.md` generated plan layout
- Conversation decisions about plan overwrite safety, archived plan visibility,
  expensive generation, and project-wide baseline-aware planning

Conversation decisions:

- `g` should generate a missing plan only.
- Regenerating an existing plan must be a separate explicit action.
- Repeated keypresses must not create concurrent model requests.
- Existing plans may include manual edits and must not be overwritten silently.
- Replacing a plan should archive the previous canonical plan first.
- Archived plans are human-recoverable history, not active context.
- Archived plans must not be indexed, listed as active plans, or used as model
  context for future generation.
- Plan and spec generation are expensive operations and need a longer timeout
  than next-question quiz generation.
- Failure messages must make clear when the UI is showing the last saved plan
  rather than a newly generated result.

## Current State

Plan generation is implemented, but the operation is still too easy to repeat
and too destructive conceptually.

Important limitations:

- `g` attempts generation even when a canonical plan already exists.
- There is no explicit regenerate action.
- Existing plan replacement does not archive the previous plan.
- Archived plan semantics do not exist yet.
- Archived files are not explicitly excluded from workspace context.
- Timeout behavior can fail long plan generations.
- Failure state can be visually confusing because the old plan remains visible.
- TUI busy state reduces accidental repeats in-process, but there is no
  service-level generation guard.

## Target Model

The target model distinguishes cheap reads from expensive writes:

- `r` refreshes plan state from disk and never writes.
- `g` creates a plan only when no canonical plan exists for the selected scope.
- Regenerate is a separate explicit action that replaces the canonical plan
  after archiving the current one.
- The service owns overwrite semantics, archive semantics, and generation
  locking.
- The TUI emits intent only and renders service-provided results.
- Archived plans live under `.chc/plans/archive/` and are ignored by default
  operational context.
- Generation failures leave existing plans untouched and report that the UI is
  showing the last saved plan.

## Slice Map

Slicing name: Plan Generation Safety

Active slice number: Slice 4

Active slice short name: generation runtime safety

Expected branch name: `feat/slice-4-generation-runtime-safety`

Expected PR title: `feat(slice-4): generation runtime safety`

Expected report path:
`.prj/reports/slices/plan-generation-safety/slice-4-generation-runtime-safety.md`

Prerequisite PRs or delivered work:

- TUI use-case-based Dumps, Quizzes, Specs, and Plans surfaces.
- Model-backed spec generation flow.
- Model-backed plan generation flow.
- Project-wide planning baseline and readiness context.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | plan action semantics | `feat/slice-1-plan-action-semantics` | `feat(slice-1): plan action semantics` | `.prj/reports/slices/plan-generation-safety/slice-1-plan-action-semantics.md` |
| Slice 2 | archive before replace | `feat/slice-2-archive-before-replace` | `feat(slice-2): archive before replace` | `.prj/reports/slices/plan-generation-safety/slice-2-archive-before-replace.md` |
| Slice 3 | archive context exclusion | `feat/slice-3-archive-context-exclusion` | `feat(slice-3): archive context exclusion` | `.prj/reports/slices/plan-generation-safety/slice-3-archive-context-exclusion.md` |
| Slice 4 | generation runtime safety | `feat/slice-4-generation-runtime-safety` | `feat(slice-4): generation runtime safety` | `.prj/reports/slices/plan-generation-safety/slice-4-generation-runtime-safety.md` |

## Slice 1: Plan Action Semantics

Separate generate, regenerate, and refresh semantics in the TUI and service
adapter.

Expected behavior:

- `g` creates a plan only when the selected scope has no canonical plan.
- `r` refreshes from disk and never invokes the model.
- Regenerate is exposed as a distinct action through a separate key or command
  palette command.
- Pressing `g` on a scope with an existing plan does not write and reports that
  regeneration is explicit.
- Repeated plan generation input while busy is ignored with a clear status.

Out of scope:

- Archive file creation.
- Context index exclusion.
- Service-level lockfiles.
- Timeout tuning.

## Slice 2: Archive Before Replace

Preserve the current canonical plan before any regenerate write.

Expected behavior:

- Regenerating an existing plan first copies it to
  `.chc/plans/archive/<scope>-plan-<timestamp>.md`.
- The canonical plan is replaced only after archive succeeds.
- Regenerate of a missing plan behaves like create.
- Archive metadata identifies the original canonical path and archive time.
- Archive failure leaves the canonical plan untouched.

Out of scope:

- Hiding archived plans from context.
- TUI history browsing.
- Deleting archived plans.

## Slice 3: Archive Context Exclusion

Ensure archived plans are recoverable by humans but invisible to active
planning context.

Expected behavior:

- `.chc/plans/archive/**` is excluded from active workspace indexing or from
  context section composition.
- `.chc/specs/archive/**` is excluded if the same archive convention is present.
- TUI plan lists show only canonical plans.
- Plan generation provenance does not include archived paths.

Out of scope:

- Archive browsing UI.
- Retention policies.
- Manual recovery commands.

## Slice 4: Generation Runtime Safety

Harden expensive long-running plan/spec generation behavior.

Expected behavior:

- Plan/spec generation uses a longer provider timeout than next-question quiz
  generation.
- Service-level generation guard prevents concurrent requests for the same
  scope.
- Failure messages state that the visible content is the last saved plan when
  generation fails.
- Diagnostic logging covers start, success, failure, archive, and guard events.
- Existing plans remain untouched on timeout or provider failure.

Out of scope:

- Provider retry policy beyond one explicit user action.
- Background job queue.
- Ticket generation from plans.

## Delivery Order

Implement slices sequentially:

1. Slice 1 establishes safe user intent semantics.
2. Slice 2 makes replacement recoverable.
3. Slice 3 prevents archived content from biasing future generation.
4. Slice 4 hardens long-running provider behavior and concurrent execution.

No slice should start before the previous slice is reviewed, merged to `dev`,
and validated.
