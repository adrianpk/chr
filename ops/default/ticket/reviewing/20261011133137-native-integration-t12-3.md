---
id: TKT-20261011133137
title: "T12.3: Publish bounded certification and denial matrix"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: ops
tags: native-integration, slice-12
source: chat
reported_at: 2026-10-11T13:31:37Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-15T01:11:09Z
reviewed_at: 2026-10-15T09:16:12Z
branch: test/slice-12-native-provider-certification
commits: []
---

# T12.3: Publish bounded certification and denial matrix

## Delivery Binding

Delivery set: native-integration-queue

Slice: 12 — Live provider certification

Task: T12.3

Plan: [Slice 12](../../plan/native-integration-queue.md#slice-12-live-provider-certification)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Deliver a repeatable native-integration QA entrypoint and profile evidence for permission denial, incompatible native review, fork/no-atomic and transport failures.

## Acceptance

All six positive paths and negative boundaries are explicit; missing live access remains a blocker rather than simulated success.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Strict lint and local protocol/capability guards passed. The full live QA run
completed on unchanged 14-file source
`435d082c526a6065bec62aee4a54ab9c4f3df988a682c6c68913feb1104c715f`: 45 of 46 real
cases passed in a 636.930 s command. Actual anonymous/native-review denials,
HTTPS method/contention cases, original-PR close/retarget races and lost-report
reconciliation passed. The live-matrix runner skipped the subsequent focused
race command after failure.

- `go test -race ./internal/core -run '^(TestNativeMergePolicy|TestWorkflowGit(Atomic|Advertisement|HTTPS|SSH|SSHUnknown)|TestNativeCapabilityDenials|TestNativeReceiptProof)$' -count=1` — an independent focused run passed in 12.113 s package time and 34.826 s command time on the same unchanged source. Acta `01M4YKF8DNFZTDZ6YFY5VJ7D0G`.

This is local development evidence; final clean-HEAD validation remains pending.

## Corrective Continuation

The authorized `github_same_pr_git_source_attribution/v2` contract preserves raw
GitHub attribution S separately from exact delivered P. Complete guarded
transaction, rebase history/tree, source attribution and original merged PR
proof remain mandatory. Historical v1 records retain their strict interpretation.

The corrective live matrix passed all 46 cases on both designated forges, all
methods and SSH/HTTPS. Source-attribution, receipt persistence, later-source and
integration-repair race checks passed; strict lint passed with zero issues.
All 46 retained live evidence files match their immutable hashes. Recovery Acta
`01M4ZARSAGBGADWYHS9D0N5M9V`.

The required pre-commit `make verify` passed on unchanged 21-file source
`15354ab5747e2c0a080f0a950dde38800aa7a4ddfda8c8e5bb6ad2c87df49648`: strict lint
0 issues, core 2919.809 s, app 192.959 s and all remaining packages passed.
Acta `01M4ZDFXDP7YC3Z1NXWAYS3YHK`. The delivery controller owns final clean-HEAD
validation and integration.
See the [Slice 12 report](../../report/slices/native-integration-queue/slice-12-provider-certification.md).
