# Codex Artifact Generation Slicing Plan

Status: active

## Purpose

Normalize long-form artifact generation around the shared Codex agent runner
while keeping interactive quiz turns on the direct LLM path.

The goal is to use Codex for artifact synthesis workloads that benefit from a
long-running execution wrapper, user-space run bundles, and reproducible
evidence, without adding latency to one-question-at-a-time quiz interaction.

## Source

Primary sources:

- `.agents/playbooks/planned-delivery-workflow.md`
- `.agents/playbooks/slice-delivery.md`
- Existing `internal/chc` `AgentRunner` and `CodexAgentRunner`
- Existing `internal/chcapp` dump, spec, plan, plan tracking, and ticket flows
- Existing `.chc/user/tmp/agent-runs` bundle behavior

Conversation decisions:

- Long-form artifact generation should use Codex runner.
- Interactive quiz turns should remain direct LLM calls because they are short,
  latency-sensitive, and already reliable.
- Codex means the Codex CLI execution wrapper, not Docker.
- Codex should synthesize structured artifacts; Go code should validate and
  write files.
- Dump, spec, plan, and plan tracking are artifact generation workloads.
- Plan tracking includes generation of ticket payloads, but local Go code should
  remain responsible for writing ticket files.

## Current State

Delivered on `dev`:

- Shared `AgentRunner` and `CodexAgentRunner`.
- Codex-backed spec generation when `CHC_SPEC_GENERATION_BACKEND=codex`.
- Codex-backed single-pass plan generation when
  `CHC_PLAN_GENERATION_BACKEND=codex`.
- Codex-backed authoring quiz provider still exists, but interactive quiz turns
  remain direct model calls.

In progress outside this slicing map:

- `feat/codex-dump-generation` contains a prototype/default move of dump
  curation to Codex runner.

Remaining gaps:

- Plan tracking still synthesizes tracking JSON through the default quiz model
  path.
- Ticket payload generation inside plan tracking is not yet explicitly bounded
  as "Codex produces payload, Go writes files".
- Backend defaults and configuration are not yet documented as a stable policy.
- End-to-end bundle audit across dump, spec, plan, and tracking is not yet
  captured.

## Target Model

Long-form artifact generators use Codex runner:

- dump curation;
- spec generation;
- plan generation;
- plan tracking and ticket payload synthesis.

Interactive quiz turns stay direct:

- one question per turn;
- low latency;
- no Codex wrapper overhead;
- no requirement for run bundles.

Codex must not directly write workspace artifacts. Codex returns structured
JSON. Go validates the response, writes canonical files under `.chc/user/...`,
and records trace/bundle evidence.

## Slice Map

Slicing name: Codex Artifact Generation

Active slice number: Slice 1

Active slice short name: plan tracking codex backend

Expected branch name: `feat/slice-1-plan-tracking-codex-backend`

Expected PR title: `feat(slice-1): plan tracking codex backend`

Expected report path:
`.prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md`

Prerequisite PRs or delivered work:

- `dev` includes shared Codex agent runner.
- `dev` includes Codex routing for spec and plan generation.
- `feat/codex-dump-generation` exists as a dump curation implementation branch
  and must be reconciled before or during default-policy cleanup.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | plan tracking codex backend | `feat/slice-1-plan-tracking-codex-backend` | `feat(slice-1): plan tracking codex backend` | `.prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md` |
| Slice 2 | ticket payload boundary | `feat/slice-2-ticket-payload-boundary` | `feat(slice-2): ticket payload boundary` | `.prj/reports/slices/codex-artifact-generation/slice-2-ticket-payload-boundary.md` |
| Slice 3 | artifact backend defaults | `feat/slice-3-artifact-backend-defaults` | `feat(slice-3): artifact backend defaults` | `.prj/reports/slices/codex-artifact-generation/slice-3-artifact-backend-defaults.md` |
| Slice 4 | artifact run bundle audit | `feat/slice-4-artifact-run-bundle-audit` | `feat(slice-4): artifact run bundle audit` | `.prj/reports/slices/codex-artifact-generation/slice-4-artifact-run-bundle-audit.md` |

## Slice 1: Plan Tracking Codex Backend

Route plan tracking synthesis through the Codex runner.

Expected behavior:

- `QuizModelActionMaterializePlanTracking` is supported by
  `CodexQuizModelAdapter`.
- The Codex adapter uses the existing strict plan tracking JSON schema.
- A dedicated backend setting can route plan tracking to Codex, or the service
  can use Codex by default if the default policy has already been approved.
- `MaterializePlanTracking` still owns validation, ticket creation, and file
  writes in Go.
- Failed Codex execution, parse errors, and validation failures do not create
  partial tracking or ticket artifacts.
- Trace records `provider: codex` for the model synthesis step.
- Run bundles are written under `.chc/user/tmp/agent-runs/...`.

Candidate tasks:

- T1.1 Add plan tracking schema support to the Codex quiz model adapter.
- T1.2 Add plan tracking routing/configuration in the service.
- T1.3 Add tests that plan tracking can use the Codex route.
- T1.4 Add failure-path tests for parse/validation failure without writes.
- T1.5 Add slice implementation report and run quality gates.

Out of scope:

- Changing ticket file rendering.
- Changing existing plan tracking Markdown format.
- Making all artifact generators Codex by default.

## Slice 2: Ticket Payload Boundary

Make the generation/materialization boundary explicit for tickets created from
plan tracking.

Expected behavior:

- Codex may only produce structured `GeneratedPlanTracking` ticket payloads.
- Go validates ticket payloads before creating ticket files.
- Go remains responsible for ticket keys, paths, status, metadata, and body
  rendering.
- Ticket creation remains all-or-nothing at the service boundary where current
  behavior allows it.
- Tests cover invalid payloads and confirm no direct Codex file writes are
  required for ticket materialization.

Candidate tasks:

- T2.1 Audit current plan tracking ticket payload validation.
- T2.2 Add or strengthen validation around ticket metadata and body fields.
- T2.3 Add tests proving invalid generated ticket payloads are rejected before
  file creation.
- T2.4 Add tests proving Go owns ticket paths and keys.
- T2.5 Add slice implementation report and run quality gates.

Out of scope:

- Replacing the ticket repository.
- Changing ticket workflow semantics.
- Adding manual ticket editing features.

## Slice 3: Artifact Backend Defaults

Normalize and document backend defaults for artifact generation.

Expected behavior:

- Long-form artifact generation policy is explicit:
  - dump curation uses Codex runner;
  - spec generation uses Codex runner;
  - plan generation uses Codex runner;
  - plan tracking uses Codex runner;
  - interactive quiz turns remain direct LLM.
- Existing env vars are either preserved as overrides or replaced with one
  coherent artifact generation backend setting.
- Service construction makes the routing policy easy to inspect.
- Docs explain the split between artifact generation and interactive quiz
  generation.
- Tests cover default routing and relevant override behavior.

Candidate tasks:

- T3.1 Decide and implement default routing for dump/spec/plan/tracking.
- T3.2 Preserve or simplify env var override behavior.
- T3.3 Reconcile the `feat/codex-dump-generation` implementation with the
      committed slice map.
- T3.4 Update docs/config reference for artifact generation backends.
- T3.5 Add routing tests for default and override cases.
- T3.6 Add slice implementation report and run quality gates.

Out of scope:

- Changing quiz interaction behavior.
- Removing the OpenAI direct adapters if they remain useful as fallbacks.
- Changing Codex authentication/configuration.

## Slice 4: Artifact Run Bundle Audit

Run a controlled end-to-end probe and record evidence for Codex-backed artifact
generation.

Expected behavior:

- A reference workspace run exercises dump, spec, plan, and plan tracking.
- Each generated artifact has a corresponding Codex run bundle under
  `.chc/user/tmp/agent-runs/...`.
- LLM trace records `provider: codex` for long-form artifact generation.
- Generated artifacts live under `.chc/user/...`.
- The audit records latency, output quality notes, avoidable errors, and any
  remaining trace visibility gaps.
- No reference workspace artifacts outside the intended test scope are
  overwritten without explicit test setup.

Candidate tasks:

- T4.1 Create or update a temporary probe script for the reference workspace.
- T4.2 Run dump/spec/plan/tracking generation with clean trace and bundles.
- T4.3 Inspect generated artifacts and trace evidence.
- T4.4 Record results in the slice report.
- T4.5 Add any missing lightweight docs for manual reproduction.
- T4.6 Run quality gates.

Out of scope:

- Implementing the reference blog fixture itself.
- Automating remote PR merges.
- Treating generated fixture artifacts as product code.

## Delivery Order

Implement slices sequentially:

1. Slice 1 makes plan tracking use Codex for synthesis.
2. Slice 2 hardens the ticket payload boundary.
3. Slice 3 normalizes defaults and documents the policy.
4. Slice 4 proves the full artifact generation flow with trace/bundle evidence.

No implementation slice should start before this plan and tracker are committed
to `dev`.
