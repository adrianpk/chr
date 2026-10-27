# Slice 11: Ticket, readiness and execution boundaries

Status: delivered
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-11-quality-tickets`
PR: [#448](https://forge.adrianpk.com/adrianpk/chr/pulls/448)

## Purpose

Give ticket records/rules and execution/readiness their actual owners while
preserving durable contracts, guarded mutation and native workflow authority.

## Delivered Behavior

`core/ticket` owns ticket models, rendering/parsing/validation, SQL rows/tags,
indexing/discovery/reconciliation, key reservation and file creation, review-repair
rules and merge-handoff invariants. `core/execution` owns execution/loop records,
repositories, runner/events, artifact paths, changed-file discovery, readiness,
explicit safe repair and ticket worktree lifecycle. Actual consumers call those
owners directly; no root aliases or forwarding APIs were added.

Ticket lifecycle, review state/result and merge state/result have separate types.
Execution/loop machines retain their types; cleanup and retained execution status
are typed at the execution owner. Automatic/manual preparation is a private typed
choice. Existing JSON/Markdown/SQL strings, field order, optionality and transition
guards remain intact.

## Implementation Notes

D03 settings normalization was removed before splitting readiness into check,
repair, Git, status and diagnostic concerns. Check/repair settings share the same
normalization; only explicit repair has a Git mutation port. User checkout,
ticket worktree, preview and authoritative readiness remain distinct.

Native indexing composes ticket and workflow/integration projections on the same
SQLite handle and transaction. Native workflow source locks, metadata, cache
invalidation, recovery and guarded ticket writes keep their original authority.
File creation receives the existing durable key allocator; no second key source
or database pool is introduced.

The [owner map](../../../plan/code-quality-remediation-core-owners.csv) records
40 concrete native composition/authority retentions: board/document projections,
work selection, guarded ticket mutations, remote review/merge/control and evidence
report/admission policy. Those operations compose multiple domains or own native
representation/receipt authority; they do not become ticket or execution internals.
Component tests sit beside their owner; native integrations retain private test
fixtures. `timePtr` and the default preparation test wrapper are test-only.

## Contracts Added or Changed

- [789 exact declaration/test preimages](../../../plan/code-quality-remediation-slice-11-moves.csv) reconcile relocation, naming and subsequent explicit contract changes.
- [59 assigned signature/watch dispositions](../../../plan/code-quality-remediation-slice-11-dispositions.csv): 47 fixed, 12 retained cohesive contracts; every original baseline identity remains traceable.
- [13 control/legacy/preparation boundaries](../../../plan/code-quality-remediation-slice-11-boundaries.csv) preserve presence, applied/changed, dry-run, read-only observation, authority and retained compatibility meanings.
- The parameter ceiling is six: 9306 production declarations, 162 exact debts, one existing external exception. [All 18 above-six declarations](../../../plan/code-quality-remediation-slice-11-parameter-phase.csv) have exact signatures and approved Slice 12–15 owners.
- SQL column/value order, NULL timestamps, non-nil/empty lists, bounded pagination, command ordering, error matching, resource cleanup and existing wire/hash fixtures are preserved.

## Files of Interest

- `internal/core/ticket/`, especially `index_row.go`, `review_state.go`, `merge_state.go`, `review_repair.go` and `merge_handoff.go`.
- `internal/core/execution/`, especially readiness/worktree files, `runner.go`, `agent_event_stream.go` and `changed_files.go`.
- `internal/core/workflow_index_storage.go`, `workflow_execution_prepare.go` and native guarded ticket/result operations.
- The move, owner, signature, duplicate, boundary and parameter-phase ledgers linked above and from the approved plan.

## Validation

Controller validation `3XHQGI6UV6MWJFH5CWULZLBTCE` passed the exact ordered commands
below on clean PR head `d0d65c114cf8ddffcd0492f7fc0b746a832adc0d`, now verified
integrated into canonical dev through PR #448.

- `make verify`: pass.
- `go test ./internal/core/ticket/... ./internal/core/execution/... -count=1 -timeout=3m`: pass.
- `go test ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(Ticket|RunTicket|ValidateTicket|Board|WorkSelectionPreflight|WorkDoctor|Evidence|ParseEvidence|RenderEvidence|UpsertEvidence|SummarizeEvidence|ExecutionReport|NewExecutionReport|RenderExecutionReport|LoopReport|NewLoopReport|RenderLoopReport|RemoteCommand|ParseRemoteCommand|DetectReviewProviderReference|RunReviewProviderDiagnostics|RemoteReviewAdapter|ApplyAgenticReview|ParseAgenticReview|TrustedAutoMerge|QualityRecordWire|QualityStorage|Leaf)' -count=1 -timeout=3m`: pass.
- `go test ./internal/core -run '^TestWorkflow(Record|Index|ReviewModes|ReviewDecisionChecks|ReviewPolicy|SelectionNativeTies|TaskGateOrder|ManualExact)$' -count=1 -timeout=3m`: pass.
- `git diff --check`: pass.

One earlier combined selection included every WorkflowSelection/Review/Task/Manual
test and exhausted the package's three-minute limit while entering
TestWorkflowReviewModes/trusted. The bounded ticket/state and selected native
checks passed separately; the timeout is not successful evidence.

## Risks and Follow-ups

Remaining workflow parameter debt is explicitly assigned to the approved
Slice 12–15 tasks. Slice 11 of 16 is delivered through verified canonical PR #448 integration.
