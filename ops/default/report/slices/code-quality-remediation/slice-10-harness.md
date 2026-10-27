# Slice 10: Harness contracts and runtime lifecycle

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-10-quality-harness`
PR: `#447`

## Purpose

Separate adapter-neutral runners from project binding, configuration admission
and host execution policy while preserving retained identities and resources.

## Delivered Behavior

- Core harness owns exact binding/profile/selection values, semantic revisions,
  capability sealing and bounded private profile loading. Neutral runner/provider
  packages retain their independent contracts and dependency direction.
- Harness host owns executable fingerprints, frozen launch admission, retained
  generation operations, artifact IO, model dispatch, usage/validation metadata
  and concrete authoring/quiz host adapters. Runtime owns launch plans, artifact
  paths, Docker commands, filesystem checks, mounts and explicit instructions.
- Daemon owns lifecycle/project loading and SQLite heartbeat/stop state. Trace
  owns bounded, redacted JSONL records and rotation. Workspace owns shared config
  path/lock IO, operator userspace and execution workspace policy values.
- Generation callers retain selected adapters and operation values and derive
  contexts from their own callers. The Pi bridge owns cancellation and cleanup;
  process exchange groups its owned bounded output spools.
- Runtime retains shared references to existing admitted capacity files. Native
  workflow retains inherited-slot validation, identity/limits and coordinated
  lock conversion. Cancellation and early errors release owned resources.

## Implementation Notes

Moves and package naming are separate commits. App, native configuration,
workflow and view consumers import the actual owners directly. There are no
migration aliases or root forwarding APIs for moved capabilities. The redundant
response trace export now exposes the original implementation directly.

The [move ledger](../../../plan/code-quality-remediation-slice-10-moves.csv)
reconciles 327 declaration/test preimages and constant specifications after
owner qualification/naming. All current hashes resolve to actual declarations.
The [signature dispositions](../../../plan/code-quality-remediation-slice-10-dispositions.csv)
cover all 38 assigned occurrences: 31 fixed and seven cohesive five-input
contracts retained. Two additional owned resolver contracts are reconciled in
[the contract ledger](../../../plan/code-quality-remediation-slice-10-contracts.csv).
Production policy reports 9308 declarations, 207 exact debts, ceiling seven and
one unchanged external exception.

The [boundary ledger](../../../plan/code-quality-remediation-slice-10-boundaries.csv)
records actual flag meanings and retained native composition. Full WorkspaceConfig
parsing/revision, authorized selection changes/journals, native generation fences,
workspace template initialization, execution diagnostics, ticket/evidence context
and workflow capacity remain with their actual owners. These implementations
compose concrete types and authority; they are not compatibility facades.

## Contracts Added or Changed

Named runtime path/layout, executable, generation preparation/selection, usage,
readiness, launch command and private Pi resource values replace repeated inputs
and tuples while preserving found/unknown/partial states separately from errors.
The same parent context reaches model dispatch and guarded quiz/spec/review paths;
no operation context is stored in the bound generation result.

Bindings and generation records preserve JSON field order/tags, checksum/hash
bytes, profile/capability identity, file modes, bounds and original error meanings.
The native config bridge independently fences revision/profile/purpose under the
original lock; a preparer cannot substitute a forged successful admission callback.
An independent wire fixture verifies exact encoded bytes before semantic corruption
and checksum repair. Existing retained-operation, process-restart and selection
journal/receipt tests exercise the real host/configuration boundary.

Runtime and workspace retain enum/default strings, canonical/private artifact
paths, Docker argv and mounts. Daemon SQL/state semantics are unchanged. CLI
select/clear, file-created outcomes, API/runner dispatch and known usage retain
their distinct binary contracts and actual call-path coverage.

## Files of Interest

- `internal/core/harness/` and its host/runtime/daemon/trace children, each with
  direct dependencies and package documentation.
- `internal/core/workspace/` and `project/repository_directory.go` for shared
  configuration/identity/filesystem contracts.
- `internal/core/config_generation.go`, `config_harness_selection_*.go`,
  `workflow_runtime_reservation.go` and `internal/app/service_harness_generation.go`
  for concrete configuration, admission and caller-context composition.
- `scripts/goquality/boundaries.go` for enforced child-to-parent dependency
  direction; `scripts/goquality/debt.json` retains only exact remaining debt.

## Validation

Task checks passed for real profile/config/selection journals and receipts,
executable files, retained generation/restart/metadata, workspace/runtime paths,
SQLite daemon lifecycle, trace redaction/rotation, CLI/service adapters and
strict pinned lint. Focused race checks passed for caller cancellation/retry,
shared provider capacity, process/Pi cleanup and retained reservation lifetime.

Controller evidence is keyed to the exact clean PR HEAD and this ordered final
command list, with required repository commands first:

- `make verify`
- `go test ./internal/harness/... ./internal/core/harness/... ./internal/core/workspace ./internal/core ./internal/app ./scripts/goquality -run '^Test(Harness|AgentRun|Pi|StructuredHarness|Runtime|PlanRuntime|ValidateRuntime|NewHarnessInvocation|WorkWorkspace|WorktreePlan|WorkReadiness|WorkerContext|MaterializeWorkerContext|ResolveProjectUserspace|EnsureProjectUserspace|IsValidOperatorID|LoadLLMTrace|WriteLLMTrace|RecordLLMTrace|Daemon|ValidateDaemon|EvaluateDaemon|LoadDaemon|NewDaemon|WorkflowCoordinator|WorkflowIntegrationCoordinator|QualityRecordWire|Leaf|GuidedUseCase(Spec|Quiz|Plan|Admission))' -count=1 -timeout=3m`
- `go test -race ./internal/harness/process ./internal/harness/pi ./internal/core ./internal/app -run '^Test(AgentRunProcess(Cancellation|Streams|Bounds|Reaper)|PiTransport(Cancellation|Usage)|PiCredentialRefresh(Canceled|Deadline)|PiWorkerIsolation(Reservation|Canceled)|HarnessGenerationBindingSelectionOrder|HarnessGenerationCallerContext|PiGenerationCapacity)$' -count=1 -timeout=3m`
- `git diff --check`

## Risks and Follow-ups

Controller validation and integration into dev determine final delivery state.
The report and ticket remain reviewing until that integration is verified.
