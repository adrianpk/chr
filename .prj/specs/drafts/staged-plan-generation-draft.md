# Staged Plan Generation

Status: draft

Sources:

- `.prj/specs/specs/context-composition-pipeline.md`
- `.prj/plans/slicing-plan-generation-safety-plan.md`
- `internal/chr/quiz_openai_model.go`
- `internal/chrapp/service_plan.go`
- Conversation decision: replace monolithic plan generation with a staged flow
  only if it reduces timeout and consistency risk without invalidating
  existing generated plans.

## Purpose

Define a staged plan generation workflow that reduces timeout risk and large
single-response fragility while preserving a coherent implementation plan.

Current plan generation asks the model to synthesize all project/use-case
context and return one complete Markdown plan body in a single response. The
current implementation already uses long generation timeouts and concurrency
guards, but large context sets and long plans can still make the operation
slow, fragile, or difficult to recover.

Staged plan generation splits the operation into:

1. a global plan outline;
2. one detail expansion per planned slice;
3. deterministic Markdown assembly.

The global outline preserves cross-slice coherence. Slice detail expansion
keeps each model response smaller and easier to validate.

## Goals

- Reduce timeout risk for large project-wide or use-case plan generation.
- Preserve global slice ordering, dependency analysis, and plan coherence.
- Keep generated plan output compatible with the existing plan document model.
- Make each model call smaller and easier to retry.
- Preserve provenance across all staged requests and responses.
- Avoid forcing migration or retrofit of already generated plans.
- Keep final Markdown assembly deterministic and service-owned.

## Non-Goals

- Do not migrate existing generated plans.
- Do not require staged generation for every plan immediately.
- Do not change the user-facing plan artifact path or canonical plan document
  format unless a later implementation slice explicitly does so.
- Do not generate tickets directly from the outline stage.
- Do not let slice expansion create, remove, reorder, or rename slices from the
  approved outline.
- Do not introduce a background job queue in this draft.

## Current Behavior

Plan generation currently uses one model request with action
`plan_generation`.

The request includes:

- project and quiz payload data;
- current runtime state;
- output contract;
- context sections containing project metadata, use-case metadata when scoped,
  dumps, quizzes, specs, and related plans;
- project-wide planning baseline/readiness sections when applicable.

The model returns a strict JSON object with:

- `title`;
- `body`;
- `status: planned`.

The `body` is a full Markdown implementation plan. The schema is simple, but
the body can be large and the prompt asks for broad synthesis, slice
decomposition, task generation, branch strategy, PR strategy, and validation
criteria in one response.

## Target Workflow

### Stage 1: Plan Outline

The service invokes a new outline generation action.

The model receives the same planning context as current plan generation, but
the requested output is only the global structure:

- plan title;
- plan status;
- plan scope summary;
- current implementation baseline summary when applicable;
- assumptions;
- low-confidence items;
- ordered slice list;
- slice dependency graph;
- global validation strategy;
- global risks and open questions.

Each outline slice must include:

- stable slice ID, such as `Slice 1`;
- short name;
- objective;
- rationale;
- dependency references;
- expected branch name;
- expected PR title;
- expected report path;
- complexity estimate;
- implementation boundary summary.

The outline is the authoritative slice map for the rest of the generation
cycle.

### Stage 2: Slice Detail Expansion

For each slice in the outline, the service invokes a slice detail action.

Each detail request receives:

- the complete outline;
- the selected slice ID and short name;
- relevant planning context sections;
- the global assumptions and risks;
- any previously generated slice details when needed for sequencing.

The model returns details only for the selected slice.

Each slice detail must include:

- slice ID matching the outline exactly;
- short name matching the outline exactly;
- objective;
- implementation notes;
- acceptance criteria;
- task list;
- ticket metadata;
- branch metadata;
- PR expectations;
- validation expectations;
- dependencies;
- explicit non-goals.

The service must reject slice details that introduce unknown slice IDs, change
the selected slice ID, rename the selected slice, or contradict the outline's
ordering/dependency map.

### Stage 3: Deterministic Assembly

The service assembles the final plan Markdown from:

- the outline;
- all accepted slice detail payloads;
- deterministic section templates;
- provenance from all staged requests and responses.

The final plan should preserve the same conceptual sections expected by current
plan generation:

- High-level implementation strategy;
- Slice decomposition;
- Slice ordering;
- Slice dependency analysis;
- Tasks inside slices;
- Ticket generation data;
- Branch strategy;
- Pull request strategy;
- Merge sequencing;
- Review requirements;
- Completion validation criteria.

The renderer, not the model, owns final heading order and document shape.

## Contracts

### Plan Outline Model Action

Add a model action for plan outline generation.

The response payload should be structured JSON, not a Markdown body.

Required fields:

- `title`;
- `status`;
- `summary`;
- `assumptions`;
- `open_questions`;
- `risks`;
- `validation_strategy`;
- `slices`;
- `dependencies`.

`status` must be `planned`.

`slices` must be an ordered array. Slice labels must start at `Slice 1` and
increase sequentially within the generated outline.

### Slice Detail Model Action

Add a model action for one slice detail expansion.

Required fields:

- `slice_id`;
- `short_name`;
- `objective`;
- `implementation_notes`;
- `tasks`;
- `acceptance_criteria`;
- `ticket_metadata`;
- `branch_metadata`;
- `pr_expectations`;
- `validation_expectations`;
- `dependencies`;
- `non_goals`;

The service must validate that `slice_id` and `short_name` match the outline.

### Provenance

The generated plan should record:

- outline request ID;
- outline response ID;
- slice detail request IDs;
- slice detail response IDs;
- source context section paths.

The stored plan may keep one primary `SourceRequestID` and `SourceResponseID`
for compatibility, but staged provenance must remain inspectable in the plan
body or a related metadata structure.

## Compatibility

Existing generated plans remain valid.

No migration is required for already generated plans. The staged workflow
applies to new generations after implementation.

The service may keep the current monolithic generation path as an initial
fallback while staged generation is introduced. If both paths exist, the
selection rule must be explicit and test-covered.

## Failure Handling

Outline failure must not write a plan.

Slice detail failure must not write a partial final plan unless a later design
introduces resumable staged artifacts. The initial implementation should prefer
all-or-nothing final plan writes.

If one slice expansion fails, the service should report:

- failed stage;
- slice ID when applicable;
- request ID when available;
- provider/model error;
- whether any final plan was written.

Existing canonical plan safety rules still apply:

- generating an already existing canonical plan must not overwrite it;
- regeneration must archive the previous canonical plan before replacement;
- timeout or provider failure leaves existing plans untouched.

## Validation Requirements

Implementation tests should cover:

- outline request construction;
- outline schema validation;
- slice detail request construction;
- slice detail schema validation;
- rejection of slice detail that changes slice ID or short name;
- deterministic assembly from outline plus slice details;
- no final plan write on outline failure;
- no final plan write on slice detail failure;
- compatibility with existing plan layout and provenance;
- fallback or feature-selection behavior if monolithic generation remains.

## Risks

- Staging increases the number of model calls.
- More model calls can increase cost and total wall-clock time for small plans.
- Slice expansion may drift from the outline without strict validation.
- Retrying individual stages requires careful provenance and error reporting.
- Deterministic assembly may initially produce less natural prose than a single
  model-generated Markdown body.

## Open Questions

- Should staged generation be the default immediately or guarded behind an
  explicit option first?
- Should outline and slice detail payloads be persisted as recoverable
  intermediate artifacts?
- Should slice details receive all context sections or only context selected by
  the outline?
- Should previous slice details be included in later slice expansion requests?
- How should staged provenance be represented without bloating the final plan
  front matter?
