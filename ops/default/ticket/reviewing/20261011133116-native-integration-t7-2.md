---
id: TKT-20261011133116
title: "T7.2: Bind review to the current candidate"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-7
source: chat
reported_at: 2026-10-11T13:31:16Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T20:15:16Z
reviewed_at: 2026-10-12T21:10:46Z
branch: feat/slice-7-native-candidate-review
commits: ["c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59"]
---

# T7.2: Bind review to the current candidate

## Delivery Binding

Delivery set: native-integration-queue

Slice: 7 — Combined validation and review

Task: T7.2

Plan: [Slice 7](../../plan/native-integration-queue.md#slice-7-combined-validation-and-review)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Extend review grants with capability, source rewrite, history manifest and required policy fingerprints across every entrypoint.

## Acceptance

Any relevant S/B/P/contract/method/policy change invalidates previous publication approval.

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

T7.2 binds human, trusted and agentic review to the immutable native candidate,
queue generation, exact original/published source, history manifest and full
trusted capability. Review, approval, grant and publication checks re-observe
current inputs through the shared core path and central application service.
Changed policy, ownership, credential identity, transport capability, S/B/P,
method, profile, contract or history cannot reuse the retained grant.

Development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; all packages passed.
- `go test ./internal/core ./internal/app -run '^(TestNativeReview|TestNativeSubmission|TestServiceNativeReview)' -count=1` — passed.
- `go test ./internal/core -run '^TestNativeReviewFence$' -count=1` — passed.
- `go test ./internal/core -run '^TestNativeReviewPolicyIO$' -count=1` — passed.
- `go test -race ./internal/core ./internal/app -run '^(TestNativeReview|TestNativeSubmission|TestNativePreparation|TestWorkflowReview|TestWorkflowPreparation|TestServiceNativeReview|TestServiceWorkflowReview|TestServiceWorkflowPreparation)' -count=1` — passed.
- `git diff --check` — passed.

Source-only native approval, stale policy reuse and lost reviewer ownership on
queue movement were reproduced before correction. A capacity-fixture argument
error in the initial checks was corrected; the repository and focused normal/race
checks then passed. All owned check processes exited naturally.

Implementation commit: [c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59](https://forge.adrianpk.com/adrianpk/chr/commit/c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59).

Acta commit activity: `01M4RZ9E4YV2JFRSJFAAG9TPTH`.

Canonical slice PR and final exact-HEAD checks remain pending.
