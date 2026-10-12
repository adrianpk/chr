---
id: TKT-20261011133109
title: "T5.3: Enforce non-no-op and signing policy"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-5
source: chat
reported_at: 2026-10-11T13:31:09Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T10:21:53Z
reviewed_at: 2026-10-12T10:56:07Z
branch: feat/slice-5-native-method-preparation
commits: [134ebaf9c8041ca92bdd184ee48f6e521a293e07]
---

# T5.3: Enforce non-no-op and signing policy

## Delivery Binding

Delivery set: native-integration-queue

Slice: 5 — Method preparation and provenance

Task: T5.3

Plan: [Slice 5](../../plan/native-integration-queue.md#slice-5-method-preparation-and-provenance)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Reprepare source-equal candidates within the approved bound and sign rewritten commits through the trusted host where required.

## Acceptance

No artificial unrelated commit is inserted; missing certified signing capability blocks before validation/publication.

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

The implementation records preparation identity, initial/effective host committer
dates and a maximum of three identity retries. It retains source authors,
messages, trees and commit count without adding an unrelated commit.

Required signatures are bound to the current capability profile. An explicitly
supplied host SSH signer creates and verifies every rewritten object and declared
integration fix. Missing or mismatched signing capability blocks before a
validation lane is attached. Public profile and per-object signature hashes enter
provenance; private signing paths and key material are excluded.

Successful task checks:

- `make verify` — zero lint issues; core 1345.889 seconds, app 114.726 seconds and all other packages passed.
- `go test ./internal/core -run '^TestNativePreparation|^TestNativeSigning|^TestNativeSignedFix|^TestNativeCollision|^TestNativeReplay|^TestNativeCombined' -count=1` — pass, 26.664 seconds.
- `go test -race ./internal/core -run '^TestNative|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit|^TestWorkflowSubmission' -count=1` — pass, 705.730 seconds.
- `git diff --check` — pass.

Task commit: [134ebaf9c8041ca92bdd184ee48f6e521a293e07](https://forge.adrianpk.com/adrianpk/chr/commit/134ebaf9c8041ca92bdd184ee48f6e521a293e07).
Acta task record: `01M4QW3D3M4DHF1MHP5B7X11WR`.
T5.4, the canonical PR and final exact-HEAD controller checks remain pending.
This task is validated locally; it is not integrated or finalized.
