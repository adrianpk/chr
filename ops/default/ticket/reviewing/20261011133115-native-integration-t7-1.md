---
id: TKT-20261011133115
title: "T7.1: Run required gates on the exact prepared result"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-7
source: chat
reported_at: 2026-10-11T13:31:15Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T19:18:41Z
reviewing_at: 2026-10-12T20:12:03Z
branch: feat/slice-7-native-candidate-review
commits: ["3f1aac6910b7e8885146a544a735db69f49c1a94"]
---

# T7.1: Run required gates on the exact prepared result

## Delivery Binding

Delivery set: native-integration-queue

Slice: 7 — Combined validation and review

Task: T7.1

Plan: [Slice 7](../../plan/native-integration-queue.md#slice-7-combined-validation-and-review)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Execute repository-owned combined validation on clean P and retain result-tree/history-bound evidence independently of agent reports.

## Acceptance

A textually clean semantic regression fails the combined gate instead of being published.

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

T7.1 connects host-owned combined validation to native candidate evidence and
queue authority. It freezes exact capability and queue generation before the
gate, rejects semantic failures and stale images, and retains stopped process
observations before shared capacity is released.

Development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; all packages passed.
- `go test ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed.
- `go test ./internal/core -run '^TestNativePreparationStop$' -count=1` — passed.
- `go test -race ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed.
- `git diff --check` — passed.

The native generation regression was reproduced before correction. The existing
host owner now survives queue projection until its gate finishes or stops.
Implementation commit: [3f1aac6910b7e8885146a544a735db69f49c1a94](https://forge.adrianpk.com/adrianpk/chr/commit/3f1aac6910b7e8885146a544a735db69f49c1a94).

Acta commit activity: `01M4RW111YN23ZPXZ87AHKG9CB`.

Canonical slice PR and final exact-HEAD checks remain pending.
