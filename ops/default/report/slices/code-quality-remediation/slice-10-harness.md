# Slice 10: Harness contracts and runtime lifecycle

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-10-quality-harness`
PR: pending

## Purpose

Separate adapter-neutral execution from project admission, binding and host
lifecycle policy while preserving frozen identities and bounded child resources.

## Delivered Behavior

Binding sealing and validation now live beside the retained binding values in
core/harness. Capability and binding hashes use the original JSON representation;
profile-name and selection-digest checks have one owner. Root and app consumers
call the owner directly. Native command fencing retains its original workflow
errors and execution checks.

Runtime planning consumes its existing named artifact path set directly; the
seven-result facade is removed. Canonical layout selection has a named outcome.
Pi preparation returns distinct credential, call-resource, model-mount, container
and publication records. Expected absence and publication stability remain
separate from failure. Output limits, digest guards, credential refresh exclusion,
cleanup ordering and finite deadlines are unchanged.

The bridge derives and defers its own cancellation before input reading. The
input watcher receives that exact cancellation handle and returns prompt/error;
there is no stored operation context or context hidden inside a result record.
The real malformed-input test now verifies cancellation on every error branch.
The process exchange groups its two owned spools and retains explicit context,
cancellation and protocol handler arguments.

Executable freezing now returns one host-only value pairing its retained public
reference with its private resolved command. Binding fields and serialized
executable/argv hashes are unchanged; the new value is never serialized into a
retained binding. Every native and application consumer uses its explicit fields.

Profile parsing, role references, readiness/selection projections and generation
binding sealing/validation now live under core/harness. Workspace configuration
field interpretation and whole-image Revision stay with their actual configuration
owner. Three profile tests move beside the parser. The 29 moved declarations and
tests match their source preimages after namespace qualification and equivalent
standard-library enum membership. No serialized field or profile hash changes.

Generation selection now projects its ID and binding without private host data.
Receipt lookup retains Found separately from read/decode/conflict errors. Measured
usage retains unknown/partial/known status and absent counters, while checklist
observation retains a nonfatal parse problem. Their callers consume named fields
at the same original observation and guarded mutation points.

Generation freezing now accepts one request for project/operation identity and
host routing, with context and resolution kept explicit. The original zero-or-one
location check, retained lookup before current selection, nested configuration
locks, config revision/profile guards, executable checks and artifact write order
remain. Preparation and its owned resolver return one binding/launch value;
Codex, Pi and legacy API paths keep their distinct readiness and dispatch rules.
No operation context is stored in either new record.

Application selection now receives one coherent selection input; the complete
configuration remains available to the original legacy branch. Pi readiness and
worker launch return named local readiness/image and command/runtime/protocol
values. Original role, profile, image/architecture, private-resource and error
checks remain in order. Invocation helpers reuse the existing complete
HarnessInvocationOptions instead of repeating its five fields positionally.

The baseline prepareWorkflowHarness result was already a conforming typed
WorkflowHarnessResolution from an earlier approved prerequisite; current direct
binding/launch consumers and native regressions confirm that disposition.

The original configuration path guard and config.lock owner are shared workspace
filesystem operations. Native upgrade, selection changes and generation admission
call that owner directly, preserving path checks, permissions, locking order,
callback timing and joined unlock/close errors. Three exact function preimages
are recorded. Workspace dependency guards admit only the concrete file owner;
coordinator imports remain rejected.

## Implementation Notes

The approved inventory assigns 38 signature occurrences and 30 initial root
files to this slice. Each current owner and consumer is checked before moving.
The seven binding and 29 profile/selection declaration/test preimages in
slice-10-moves.csv match after namespace
qualification and the equivalent standard-library membership predicate.
Dependency guards allow only the existing neutral capability and exact digest
owner; workflow/coordinator imports remain forbidden.

A separate naming commit shortens the moved types and operations to Profile,
Selection, Readiness, SelectionConfig, GenerationBinding, LoadProfiles,
ProfileRevision, SealGenerationBinding and ValidateGenerationBinding. Native
configuration-only files now use configuration names. Direct consumers retain
all field names and JSON tags; the preimage ledger follows each current name.

## Contracts Added or Changed

Pure binding validation and lexical identity predicates moved without a root
forwarding API. A separate naming commit shortens the public contracts to
harness.SealBinding, ValidateBinding and ValidProfileName; the focused binding,
selection and wire regressions passed again after that rename.

## Files of Interest

- `internal/core/harness/`: existing binding values and planned project policy.
- `internal/harness/`: existing adapter-neutral execution and process owners.

## Validation

Slice 9 closure uses its existing
three-command controller evidence at `0ce066b1599b55f6efc3f151bb960a74b75d5e91`.

Focused implementation checks passed:

- `go test ./internal/core/harness ./internal/core ./internal/app -run '^Test(HarnessBinding|HarnessSelection|QualityRecordWire)' -count=1 -timeout=3m`
- `go test ./scripts/goquality -run '^Test' -count=1`
- `make check-go-style`: 9304 declarations, ceiling seven, 237 debts, one external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app`

The signature disposition ledger records 30 fixed occurrences and seven cohesive
five-input contracts. Only service.bindGeneration remains pending. The fixed signatures
and two owned resolver contracts remove 29 exact debt allowances: 208 remain, with ceiling seven and one external
exception unchanged. No migration debt or broader allowance is introduced.

Focused runtime/resource and concurrency checks passed:

- `go test ./internal/harness/pi ./internal/harness/process ./internal/core ./internal/app -run '^Test(Pi|AgentRunProcess|RuntimePlan|QualityRecordWire)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app ./internal/harness/pi ./internal/harness/codex -run '^Test(PlanRuntime|ValidateRuntime|RuntimeArtifactCleanup|HarnessBinding|QualityRecordWire|PiWorkerInput|PiWorkerPreparation|PiGenerationResources|PiGenerationResourceIdentity|Codex)' -count=1 -timeout=3m`: Codex compiled with no matching tests; runtime paths, resources and bridge preparation exercised directly.
- `go test -race ./internal/harness/pi ./internal/harness/process -run '^Test(PiWorkerConnection(Bounds)?|PiCredentialRefresh(Failure|Reject|Deadline|Resources|Canceled|Home)?|PiTransportDriver|AgentRunProcess(Streams|Cancellation))$' -count=1 -timeout=3m`
- `make check-go-style`: 9303 production declarations, 227 debts, unchanged ceiling and external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/harness/... ./internal/core ./internal/app`: zero issues.

Executable preparation checks passed:

- `go test ./internal/core ./internal/app -run '^Test(HarnessBinding|HarnessGenerationBinding|HarnessGenerationIdentity|HarnessGenerationCommand|ServiceHarness|PiWorkerPreparation|QualityRecordWire)' -count=1 -timeout=3m`

- `make check-go-style`: 9303 declarations, ceiling seven, 226 exact debts, one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app`: zero issues.

Profile owner move checks passed:

- `go test ./internal/core/harness ./internal/core ./internal/app ./internal/tui ./internal/webui -run '^Test(Harness|ServiceHarness|QualityRecordWire|TelegramHarness|WebHarness)' -count=1 -timeout=3m`: TUI compiled with no matching tests; domain, native, app and Web checks passed.

Profile naming checks passed:

- `go test ./internal/core/harness ./internal/core ./internal/app -run '^Test(HarnessSelection|HarnessGenerationBinding|QualityRecordWire|ServiceHarness)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app ./internal/tui ./internal/webui`: zero issues.

Observation checks passed:

- `go test ./internal/core/harness ./internal/core ./internal/app -run '^Test(HarnessGeneration|HarnessSelection|HarnessBinding|HarnessRun|WorkerContext|BuildWorker|QualityRecordWire|ServiceHarness|TelegramHarness)' -count=1 -timeout=3m`
- `make check-go-style`: 9303 declarations, ceiling seven, 222 debts, one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app`: zero issues.

Generation input/preparation checks passed:

- `go test ./internal/core ./internal/app -run '^Test(HarnessGeneration|HarnessBinding|QualityRecordWire|ServiceHarness|PiGeneration|PlanContextualQuiz|HarnessAuthoringQuiz|TelegramHarness)' -count=1 -timeout=3m`
- `make check-go-style`: 9303 declarations, ceiling seven, 213 debts, one external exception. The first checks identified two resolved resolver/callback allowances; their exact stale entries were removed and recorded in slice-10-contracts.csv.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app`: zero issues.

Selection and invocation checks passed:

- `go test ./internal/core ./internal/app -run '^Test(HarnessSelection|HarnessBinding|HarnessInvocation|HarnessGeneration|ServiceHarness|PiWorkerInput|PiWorkerPreparation|PiWorkerContainer|QualityRecordWire|TelegramHarness|PlanContextualQuiz|HarnessAuthoringQuiz)' -count=1 -timeout=3m`
- `go test ./internal/core -run '^TestNewHarnessInvocation' -count=1 -timeout=3m`
- `make check-go-style`: 9303 declarations, ceiling seven, 208 debts, one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`: zero issues.

Shared configuration filesystem checks passed:

- `go test ./internal/core/workspace ./internal/core ./internal/app ./scripts/goquality -run '^Test(WorkflowUpgrade|WorkspaceConfig|HarnessSelection|HarnessGeneration|HarnessBinding|QualityRecordWire|ServiceHarness|Leaf|Dependency)' -count=1 -timeout=3m`
- `make check-go-style`: 9303 declarations, ceiling seven, 208 exact debts, one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workspace ./internal/core ./internal/app ./scripts/goquality`: zero issues.

## Risks and Follow-ups

T10.1-T10.6 remain in progress or pending; PR readiness is not established.
Remaining work includes configuration/selection and generation admission owners,
the service.bindGeneration signature, host lifecycle/flag review and final exact-head
controller validation. No Slice 10 PR has been opened.
