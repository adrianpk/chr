---
id: TKT-20261011133108
title: "T5.2: Replay attributed rebase histories"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-5
source: chat
reported_at: 2026-10-11T13:31:08Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T08:20:13Z
reviewed_at: 2026-10-12T09:57:20Z
branch: feat/slice-5-native-method-preparation
commits: [ed92e270edd489332191667c1c80201536363ba2]
---

# T5.2: Replay attributed rebase histories

## Delivery Binding

Delivery set: native-integration-queue

Slice: 5 — Method preparation and provenance

Task: T5.2

Plan: [Slice 5](../../plan/native-integration-queue.md#slice-5-method-preparation-and-provenance)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Preserve ordered source authors/messages/dates, empty replay mappings and explicit attributed integration-fix commits.

## Acceptance

No silent squash, dropped duplicate intent or merge-containing history bypass is accepted.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through Looper.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only Looper submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implemented attributed linear replays with exact authors/dates/messages/encoding,
retained empty and duplicate intent, declared integration fixes, and canonical
host verification of retained history versions. Unsupported merge histories and
wholly empty integration units block before lane attachment.

Task checks passed:

- `make verify` — zero lint issues; core 1352.893 seconds, app 124.414 seconds and all other packages passed.
- `go test -race ./internal/core -run '^TestNative|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit|^TestWorkflowSubmission' -count=1` — pass, 729.202 seconds.
- `go test ./internal/core -run '^TestNativeSubmission|^TestNativeReceipt|^TestNativeReplay|^TestNativeObservation|^TestNativeLaterSource|^TestWorkflowSubmission' -count=1` — pass, 167.016 seconds.
- `go test ./internal/core -run '^TestWorkflowPublication' -count=1` — pass, 165.846 seconds.
- `git diff --check` — pass.

Development failures and corrections are retained in the [Slice 5 report](../../report/slices/native-integration-queue/slice-5-method-preparation.md).
Task commit: `ed92e270edd489332191667c1c80201536363ba2`. Acta task record: `01M4QRQTNWK814YQMJG4Q5F53K`. T5.3/T5.4,
the canonical slice PR and final exact-HEAD controller checks remain pending.
