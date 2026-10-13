---
id: TKT-20261011133117
title: "T7.3: Repair actionable integration findings"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-7
source: chat
reported_at: 2026-10-11T13:31:17Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T21:12:08Z
reviewed_at: 2026-10-12T22:17:36Z
branch: feat/slice-7-native-candidate-review
pr: 381
resolution: fixed
closed_at: 2026-10-13T06:03:13Z
commits: ["91676739a6706575df5ee1e644acf29189cfc62b"]
---

# T7.3: Repair actionable integration findings

## Delivery Binding

Delivery set: native-integration-queue

Slice: 7 — Combined validation and review

Task: T7.3

Plan: [Slice 7](../../plan/native-integration-queue.md#slice-7-combined-validation-and-review)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Route bounded combined-gate/review repair through the merger while preserving task evidence and original replay attribution.

## Acceptance

Clean passing candidates avoid extra merger calls; repaired candidates receive a new host gate and review.

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

T7.3 retains actionable native gate/review findings and routes repair through the
existing bounded merger without rewinding source tasks. Host acceptance remains
pending validation. A fresh isolated checkout receives exactly the accepted P,
the unchanged host gate and a new candidate-bound review. The new candidate
retains merger provenance and original rebase task attribution and messages.
Clean passing candidates do not reserve a merger attempt. Changed findings,
policy or scope and substituted results are rejected.

Development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; all packages passed.
- `go test ./internal/core -run '^(TestNativeCandidate|TestNativePreparationFailure)' -count=1` — passed in 204.054 s.
- `go test -race ./internal/core ./internal/app -run '^(TestNativeCandidate|TestNativeMerger|TestNativePreparation|TestNativeReview|TestWorkflowPreparation|TestWorkflowReview|TestServiceNativeReview|TestServiceWorkflowPreparation|TestServiceWorkflowReview)' -count=1` — core passed in 1324.484 s and app in 40.741 s.
- `git diff --check` — passed.

The initial native review test reproduced source-task rewind. Composed repair
checks exposed a retained finished merger claim, missing original Git objects
and a stale invocation revision after the queue transition. These were corrected;
final focused, repository and race checks passed on unchanged source. All owned
check processes exited naturally; no external fixture was started.

Implementation commit: [91676739a6706575df5ee1e644acf29189cfc62b](https://forge.adrianpk.com/adrianpk/chr/commit/91676739a6706575df5ee1e644acf29189cfc62b).

Acta commit activity: `01M4S33CF0TC96K7KTAP1TNASA`.
Canonical slice PR: [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381). The controller records ordered exact-HEAD
validation separately before integration; task status is solved.

Verified integration: [PR #381](https://forge.adrianpk.com/adrianpk/chr/pulls/381),
`79af9e38426e681e5b0d778ae4df2b0a84952d73` on `dev` at 2026-10-13T05:58:15Z. Controller Acta merge:
`01M4SXE559V2QJJB40S9H1S9QB`. Ordered exact-head checks passed; this task is solved.
