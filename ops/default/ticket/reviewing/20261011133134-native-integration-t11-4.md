---
id: TKT-20261011133134
title: "T11.4: Document and verify profile setup"
status: reviewing
reviewed_at: 2026-10-14T22:46:23Z
started_at: 2026-10-14T22:43:28Z
branch: feat/slice-11-native-integration-surfaces
kind: task
severity: unclassified
priority: normal
scope: api
tags: native-integration, slice-11
source: chat
reported_at: 2026-10-11T13:31:34Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T11.4: Document and verify profile setup

## Delivery Binding

Delivery set: native-integration-queue

Slice: 11 — Shared operation surfaces

Task: T11.4

Plan: [Slice 11](../../plan/native-integration-queue.md#slice-11-shared-operation-surfaces)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add user-facing setup/reference guidance for source rewriting, policy support and optional bounded parallel/token operation.

## Acceptance

Upgrade preserves old configuration, explicit opt-in is checked before work admission and supported boundaries match the executable.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Setup and reference guidance now explains explicit source rewriting, current
publication policy support, shared parallel-role capacity, service admission
limits and optional enforced whole-call token accounting. Existing configuration
and upgrade review/profile choices remain preserved. Current unsupported YAML
keys, protected-branch/rule combinations and unmetered finite-token calls remain
explicitly rejected.

Focused configuration, upgrade, capability, admission-limit and token checks
passed: core 14.878 s and app 0.919 s. Documentation relative links, anchors and
whitespace passed. Acta `01M4Y9EP3EKTFGEZ30S6PTBGXA`. Runtime and documentation hashes remained
unchanged and the owned validation service exited successfully.

Exact development command:

- `go test ./internal/core ./internal/app -run '^(TestNativeProfileConfig|TestNativeProfileUpgrade|TestNativeLegacySnapshot|TestNativePreparedSnapshot|TestNativeCapability|TestNativeCapabilityDenials|TestNativePreparationAdmission|TestNativePreparedBound|TestNativeRolePool|TestNativeLimitPolicy|TestNativeAgentAdmission|TestNativeAgentCeiling|TestNativeAgentLateCeiling|TestNativeAgentCapability|TestWorkflowUpgradePolicy|TestWorkflowUpgradePreserves|TestWorkflowUpgradeRejects|TestWorkflowConfigDefaults|TestWorkflowConfigBounds|TestWorkflowConfigErrors|TestWorkflowConfigCLI|TestWorkflowConfigRestriction|TestWorkflowUpgradeCLI|TestWorkflowUpgradeCLIRejects)$' -count=1`

The [Slice 11 report](../../report/slices/native-integration-queue/slice-11-integration-surfaces.md)
records the executable boundaries and retained evidence. Task commit, canonical
PR, final ordered clean-HEAD checks and verified integration remain pending.
