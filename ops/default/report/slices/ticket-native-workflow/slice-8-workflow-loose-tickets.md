# Slice 8: Loose-ticket integration loop

Status: delivered

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-8-workflow-loose-tickets`

PR: [#371](https://forge.adrianpk.com/adrianpk/chr/pulls/371).

## Purpose

Integrate exact independent tickets and filtered loose work through the shared
candidate, canonical PR, verified receipt and bounded continuation lifecycle.

## Delivered Behavior

An independent task is its actual integration unit. Its candidate retains the
host-validated task proof, source lane and frozen contract without constructing
a synthetic slice or set. Each ticket has its own canonical PR and immutable
receipt. Merge, rebase and squash retain their existing result proofs and review
policies. Local validation and a bound PR never mark the task done. An exact
grouped-ticket invocation validates only its selected task and cannot integrate
the parent slice partially.

Loose invocations retain an admitted order independently of their current
eligible queue. After a verified integration releases the lane, selection reads
fresh rank/status, AND filters, requested order and delivered prerequisites.
The confirmed prefix remains fixed. Excluded admissions stay visible without
counting as delivery; grouped tickets remain excluded. Successors start from
the hydrated canonical target without moving the host checkout. Later admission
is frozen off by default; accepted later admission adds matching ready work only
after confirmed progress in an unfinished invocation. Exhaustion is terminal.

A failed loose task pauses its invocation and retains consumed attempts. An
unmet dependency persists its precise prerequisite blocker before any worker
claim, without admitting outside work or bypassing it after rank changes.
Explicit resume verifies the retained target or a fresh canonical descendant,
with immutable prerequisite receipt hashes. It preserves original selection,
policy, membership and confirmed progress. Duplicate requests, 129-member paged
admission, per-call limits, terminal retries, corrupt evidence and journal
process death retain their bounded behavior. Exact-ticket manual retry remains
supported.

Explicit `work.state_authority` maps existing run-all/ready-work entrypoints to
loose native continuation and single-ticket entrypoints to one exact adopted
task. Results and previews display authority, scope and exclusions. Equivalent
unfinished starts retain their original target. `tracker_compat` retains its
unadopted legacy scope; key and actual-source guards reject adopted work before
legacy workers start. Configuration absence retains transitional defaults.

The common continuation boundary publishes only a retained validated source
branch before PR binding. Empty leases protect new refs; repaired candidates use
the bound PR's retained previous head. Retrying an exact published source is
idempotent; branch collisions, paused invocations and changed evidence reject.
A repaired source updates its existing PR before fresh review and approval.
Uncertain PR creation resumes by observation. Source publication never moves
the integration target or fabricates a delivery receipt.

## Implementation Notes

`WorkflowStore` owns selection, claims, candidate reuse, source publication,
frontier advancement and resume evidence. The application service supplies
trusted provider/Git ports and the host-supervised JSON worker in the owned
isolated lane. It delegates the same continuation operation for legacy command
aliases and never initializes the legacy runtime on the native path.

The compatibility entrypoint reads an explicit credential-free public HTTPS
canonical remote and configured integration branch. Host aliases and local
transports are not guessed into persistent repository URLs. Native and legacy
execution retain the repository's shared serial operation exclusion.

## Contracts Added or Changed

- `chr-ticket-candidate/v1` binds the actual independent ticket to retained task validation and common PR/delivery contracts.
- `loose_queue` separates admitted facts, confirmed prefix, fresh eligible ordering and excluded-member outcomes.
- Quiescent display/status writes preserve frozen admission and execution images; execution/description reassociation and active-lane edits reject.
- `prerequisite_blocked` records a resumable pause; `chr-dependency-target/v1` retains at most 128 immutable target/receipt observations in the invocation journal.
- `work.state_authority` accepts `tickets` and `tracker_compat`; absence preserves existing defaults pending versioned activation.
- Ticket-mode `--max-iterations` counts per-call accepted checkpoint steps, bounded to 1–128; zero uses the retained policy.
- `WorkflowSourceGit` supplies guarded source-only transport to common continuation. Canonical PR observation and shared receipts remain the delivery authority.

## Files of Interest

- [Independent candidates](../../../../../internal/core/workflow_ticket_candidate.go).
- [Fresh loose selection](../../../../../internal/core/workflow_loose_queue.go).
- [Prerequisite stops](../../../../../internal/core/workflow_prerequisite_stop.go).
- [Resume target proofs](../../../../../internal/core/workflow_dependency_target.go).
- [Guarded source publication](../../../../../internal/core/workflow_source.go).
- [Authority routing](../../../../../internal/app/service_work_authority.go).
- [Real entrypoint and PR repair](../../../../../internal/app/service_work_repair_test.go).
- [Execution reference](../../../../../docs/reference/workflow-execution.md).

## Validation

All four task pre-commit repository gates and expanded focused race checks
passed. Strict lint reported zero issues. Gates used the retained owned CLI
curation fixture described in the [Slice 2 report](slice-2-workflow-materialization.md)
and a scoped uncached 30-minute Go test timeout. Task commands and corrections
are recorded in the linked task tickets and tracker.

Actual bare-Git fixtures and controlled provider observations cover distinct
ticket PRs, all review modes and integration methods, all four requested
filter/order examples, guarded later admission, grouped exclusion, display-write
guards, outside prerequisites, resume target ancestry, receipt/candidate corruption,
budget exhaustion, terminal/duplicate requests and journal process-death recovery.
Shared service tests confirm native/legacy routing, exact previews, source leases,
idempotent source retry and a revised source on the same canonical PR.

The first T8.3 repository gate found an exact-ticket manual-retry regression.
The failed-task pause guard was restricted to loose invocations, then expanded
race checks and the complete gate passed. T8.4's initial routing fixtures were
corrected to expose the required publication interface and actual generated
ticket keys before its full task checks passed.

The controller passed the final clean-head checks on `f9995446e72fa0e2e53972d280001cfdccd070db`,
the exact source and merged dev head of canonical PR #371, in this order:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

Canonical PR #371 merged on 2026-10-09; Acta merge event `01M4FZER759JJQFE97B3MJW8GN`.
These candidate checks are distinct from the delivery set's later full gate.

## Risks and Follow-ups

None recorded.
