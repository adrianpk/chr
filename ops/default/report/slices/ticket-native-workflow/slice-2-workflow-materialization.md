# Slice 2: Task-Ticket Materialization

Status: reviewing
Delivery set: ticket-native-workflow
Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)
Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)
Branch: `feat/slice-2-workflow-materialization`
PR: [#365](https://forge.adrianpk.com/adrianpk/chr/pulls/365)
Implementation candidate: `1585ee89b21a51c565fee54c4f87d05fddab4dac`

## Purpose

Materialize an approved ordered plan into task-owned tickets and set-owned slice
contracts, with resumable creation and explicit adoption of selected legacy work.

## Delivered Behavior

- Approved input separates the materialization request identity, stable slice
  and task identities, display labels and positive local orders. Duplicate
  identities/orders, empty slices and incomplete contracts reject.
- Two slices with three tasks each create six native-key tickets and one set.
  Each slice retains one aggregate integration contract and an ordered task
  fingerprint. Task dependencies resolve input IDs to allocated ticket keys.
- Complete namespace updates validate before replacement and use the existing
  shared transition journal. Receipts retain original allocation bindings.
  Matching retries return existing records; changed input, membership, receipt
  images and conflicting manual edits block reconciliation.
- Recovery replays matching partial creation and releases only the recorded
  materialization allocation owner. Competing callers allocate six keys once.
- Selected legacy plan envelopes and task Markdown records can be adopted with
  exact observed images and runtime ownership checks. Keys, IDs, creation times,
  locations, original bodies and compatible unknown fields survive. Ambiguous
  keys, foreign associations, active attempts and unfinished execution block.
- Legacy operation exclusion spans linked worktrees during adoption. Matching
  interrupted adoption replays without duplicating tickets; dead locks belonging
  to that adoption can be reclaimed. Ordinary reads leave legacy records alone.
- The shared plan operation accepts explicit approved workflow input. Generation
  guidance now specifies one ticket per task and one integration unit per slice.
  The existing ticket index preserves native classification in canonical bodies.

## Implementation Notes

`WorkflowStore` owns materialization and adoption under repository exclusion.
It reuses ops envelope metadata and allocation exclusion; the application service
supplies the existing native key sequence and refreshes the existing SQLite
projection. Receipt hashes bind immutable materialization evidence to the set.

Real-files fixtures cover forward references, invalid dependencies, source
conflicts, unknown metadata, duplicate keys, use-case conflicts and body retention.
Owned subprocesses exit after each created ticket and during adoption to prove
restart recovery. A real legacy execution database covers unfinished ownership.

## Contracts Added or Changed

- `GeneratedWorkflowPlan`, `GeneratedWorkflowSlice`, `GeneratedWorkflowTask`
  and `ValidateGeneratedWorkflowPlan` define explicit approved generation input.
- `WorkflowStore.MaterializePlan`, `AdoptPlanWorkflow` and
  `WorkflowMaterializationResult` expose authoritative records and receipt refs.
- `PlanTrackingRequest.Workflow`, `MaterializeWorkflowPlan` and
  `AdoptWorkflowPlan` share application operations without implicit adoption.
- Materialization receipts, immutable slice/task contract files and owned
  allocation guards extend the existing journal protocol and namespace bounds.

## Files of Interest

- `internal/core/plan_workflow.go`: approved hierarchy and validation contracts.
- `internal/core/workflow_materialization*.go`, `workflow_allocation.go`:
  canonical creation, receipts and restart recovery.
- `internal/core/workflow_adoption*.go`: selected original records and ownership
  exclusion.
- `internal/app/service_plan_workflow.go`, `service_plan.go`: shared operations.
- `internal/core/quiz_openai_model.go`: per-task generation guidance.
- [Plan workflow materialization](../../../../../docs/reference/plan-workflow-materialization.md): persistence and explicit-adoption reference.

## Validation

Successful local implementation checks:

- `make verify`
- `go test ./internal/core ./internal/app -run '^(TestWorkflowPlanAdoption|TestWorkflowMaterialization|TestServiceWorkflowMaterialization)' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowPlanAdoption|TestWorkflowMaterialization|TestServiceWorkflow)' -count=1`
- `git diff --check`

The controller records `make verify`,
`go test ./internal/core ./internal/app -count=1`,
`go test -race ./internal/core ./internal/app -count=1` and `git diff --check`
in that order against the exact clean PR-reference follow-up head. Its
`checks/<full-head>.json` and matching log retain final candidate evidence.
These are Slice 2 checks; composed integration remains in the approved final gate.

Acceptance: A01 materialization and retry behavior; A02 identity and association
validation. Slice 5 supplies the aggregate PR delivery behavior for A01.

## Risks and Follow-ups

Selected adoption supports canonical plan envelopes and backlog or
ready-for-development Markdown tickets. Other legacy lifecycle states require
explicit evidence reconciliation. An incompatible journal image or unknown lock
owner blocks recovery. Record and journal size bounds can reject large approved
plans before replacement. Shared controls and prospective defaults remain in
their recorded later slices.
