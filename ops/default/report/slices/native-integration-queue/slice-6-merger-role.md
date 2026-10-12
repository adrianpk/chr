# Slice 6: Bounded merger role

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-6-native-merger-role`
PR: pending

## Purpose

Resolve integration conflicts through an isolated, finite merger role whose
proposals require independent host verification before candidate acceptance.

## Delivered Behavior

T6.1 defines protocol version 1 input and untrusted result contracts. The input
seals entry/unit identity, candidate revision/hash, claim generation, canonical
source/target refs, exact S/B, configured method, scope and opaque owned-workspace
identity. It includes bounded source/target intent, original history, conflict
and finding references, permitted paths/validation, replay checkpoint and finite
attempt/time/token allowances.

The parser rejects stale input or host fences, malformed/oversized output,
ambiguous JSON, escaped evidence paths, workflow/contract scope edits, incomplete
preserved requirements and invented approval/publication fields. Results retain
per-conflict rationale, bounded decision questions and supplementary local
observations without granting candidate or publication authority. Unknown usage
remains unknown.

## Implementation Notes

Protocol ownership stays in core beside existing preparation and integration
records. Actual evidence is read through the existing contained workflow store;
changed hashes, symlink traversal, foreign source manifests and incorrect replay
positions fail before acceptance. Existing task/review/receipt schemas are
unchanged. No dependency is added.

## Contracts Added or Changed

`WorkflowMergerInput`, `WorkflowMergerFence` and `WorkflowMergerResult` define the
role boundary. Input/output images use the existing 1 MiB file bound, lists use
128-entry limits and JSON nesting is limited to 32 levels. Attempts are positive
and capped at 10; runtime uses the existing worker-policy ceiling. Optional token
allowances require whole-call accounting and the exact admitted deadline.

`SealWorkflowMergerInput`, `CheckWorkflowMergerEvidence` and
`ParseWorkflowMergerResult` validate immutable inputs, actual retained evidence
and proposals. They neither launch an agent nor acquire capacity, reserve budget,
accept Git changes, run a gate or create review/publication approval.

## Files of Interest

- [Merger contracts](../../../../../internal/core/workflow_merger.go).
- [Input validation](../../../../../internal/core/workflow_merger_input.go).
- [Contained evidence](../../../../../internal/core/workflow_merger_evidence.go).
- [Result parsing](../../../../../internal/core/workflow_merger_result.go).
- [Protocol reference](../../../../../docs/reference/workflow-execution.md#merger-proposal-protocol).

## Validation

T6.1 focused checks:

- `go test ./internal/core -run '^TestNativeMerger' -count=1` — pass, 0.132 seconds.
- `go test -race ./internal/core -run '^TestNativeMerger' -count=1` — pass, 1.263 seconds.
- `go test ./internal/core -run '^$' -fuzz '^FuzzNativeMergerJSON$' -fuzztime=10s -parallel=2` — pass, 37022 executions, 11.044 seconds.

The initial strict lint check found two complexity and thirteen spacing findings.
Cohesive validators and formatting corrected those findings; strict lint then
reported zero issues. Required T6.1 pre-commit checks passed:

- `make verify` — zero lint issues; core 1282.104 seconds, app 94.013 seconds and all other packages passed.
- `git diff --check` — pass.

Owned fixtures completed naturally and their disposable data was cleaned by the
test lifecycle.

## Risks and Follow-ups

T6.2 owned harness execution, T6.3 semantic/replay repair and T6.4 host acceptance
remain pending in this slice. The protocol alone does not deliver automatic
conflict repair. No canonical Slice 6 PR is open yet. Required ordered final
candidate checks use the controller ledger after implementation and PR-reference
commits are pushed. No full delivery-set gate has run.
