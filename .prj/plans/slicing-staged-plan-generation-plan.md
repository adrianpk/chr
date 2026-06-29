# Staged Plan Generation Slicing Plan

Status: active

## Purpose

Plan the implementation of staged plan generation so large implementation
plans are generated through smaller, validated model calls without losing the
global slice map.

The goal is to replace the fragile one-response planning path with an outline
plus slice-detail workflow and deterministic assembly while preserving the
existing generated plan artifact contract.

## Source

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.prj/specs/drafts/staged-plan-generation-draft.md`
- `.prj/plans/slicing-plan-generation-safety-plan.md`
- Existing `internal/chr` quiz model and plan generation contracts
- Existing `internal/chrapp` plan generation service flow

Conversation decisions:

- Monolithic plan generation can become slow or fragile with large context.
- Slice-by-slice generation without a global outline risks incoherent plans.
- The preferred model is outline first, then slice detail expansion, then
  deterministic assembly.
- Existing generated plans remain valid and do not need migration.
- The initial implementation should preserve the current canonical plan
  document layout.

## Current State

Plan generation currently invokes one model action, `plan_generation`, and
expects one strict JSON response containing a complete Markdown plan body.

Safety hardening already exists:

- long generation timeout;
- service-level same-scope generation guard;
- existing-plan generate refusal;
- archive-before-replace regeneration;
- archived context exclusion.

Remaining limitations:

- one model response must contain the entire plan;
- timeout risk grows with context and plan size;
- validation can only inspect the final plan body as one large string;
- retries cannot target a single failed slice expansion;
- the model owns final Markdown structure and slice detail wording at once.

## Target Model

Staged plan generation has three implementation stages:

1. Generate a structured outline with the authoritative slice map.
2. Expand each outline slice into structured implementation detail.
3. Assemble the final Markdown plan deterministically.

The final user-facing artifact remains a generated plan with status
`planned`.

The service owns:

- stage orchestration;
- outline and slice detail validation;
- final Markdown section order;
- provenance across staged model calls;
- all-or-nothing final plan writes.

## Slice Map

Slicing name: Staged Plan Generation

Active slice number: Slice 1

Active slice short name: outline generation contract

Expected branch name: `feat/slice-1-outline-generation-contract`

Expected PR title: `feat(slice-1): outline generation contract`

Expected report path:
`.prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md`

Prerequisite PRs or delivered work:

- Plan Generation Safety slicing delivered.
- Model-backed spec and plan generation flow.
- Staged Plan Generation draft spec.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | outline generation contract | `feat/slice-1-outline-generation-contract` | `feat(slice-1): outline generation contract` | `.prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md` |
| Slice 2 | slice detail expansion | `feat/slice-2-slice-detail-expansion` | `feat(slice-2): slice detail expansion` | `.prj/reports/slices/staged-plan-generation/slice-2-slice-detail-expansion.md` |
| Slice 3 | deterministic plan assembly | `feat/slice-3-deterministic-plan-assembly` | `feat(slice-3): deterministic plan assembly` | `.prj/reports/slices/staged-plan-generation/slice-3-deterministic-plan-assembly.md` |

## Slice 1: Outline Generation Contract

Add the model action, response schema, payload types, validation, and request
construction for plan outline generation.

Expected behavior:

- A new model action represents plan outline generation.
- The OpenAI response format uses a strict JSON schema for outline payloads.
- Outline payloads contain title, status, summary, assumptions, open
  questions, risks, validation strategy, ordered slices, and dependencies.
- Slice labels start at `Slice 1` and are sequential inside one outline.
- Each outline slice records short name, objective, rationale, dependencies,
  branch metadata, PR metadata, report path, complexity, and boundary summary.
- Request construction reuses the same planning context sections as current
  plan generation.

Out of scope:

- Calling outline generation from `GeneratePlan`.
- Slice detail expansion.
- Final Markdown assembly.
- Persistence of intermediate outline artifacts.

## Slice 2: Slice Detail Expansion

Add the model action, response schema, payload types, validation, and request
construction for expanding one outline slice.

Expected behavior:

- A new model action represents slice detail expansion.
- Each request includes the full outline and the target slice identifier.
- The response contains details for exactly one slice.
- The service validates that the returned slice ID and short name match the
  outline.
- The service rejects unknown slice IDs, renamed slices, and slice details that
  attempt to alter outline ordering.
- Slice detail payloads include implementation notes, tasks, acceptance
  criteria, ticket metadata, branch metadata, PR expectations, validation
  expectations, dependencies, and non-goals.

Out of scope:

- Final plan assembly.
- Retrying failed slice detail generation.
- Persisting intermediate slice detail artifacts.

## Slice 3: Deterministic Plan Assembly

Connect staged generation into the plan generation service and assemble the
final Markdown plan from validated staged payloads.

Expected behavior:

- `GeneratePlan` can run the staged path.
- The service generates outline, then expands each outline slice.
- The final plan Markdown is assembled deterministically from outline and slice
  details.
- The final plan keeps the existing generated plan layout and status contract.
- Existing generate/regenerate/archive/failure safety rules still apply.
- No final plan is written if outline generation fails.
- No final plan is written if any slice detail generation fails.
- Provenance records staged request/response IDs in an inspectable form.

Out of scope:

- Migrating existing generated plans.
- Background queue support.
- Resumable intermediate artifacts.
- Ticket materialization changes.

## Delivery Order

Implement slices sequentially:

1. Slice 1 creates the outline contract and tests it without changing runtime
   behavior.
2. Slice 2 creates slice detail expansion and validation.
3. Slice 3 integrates orchestration and deterministic assembly into plan
   generation.

No implementation slice should start before the committed plan and tracker are
approved for execution.
