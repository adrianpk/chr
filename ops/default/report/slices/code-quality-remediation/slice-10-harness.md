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

Daemon lifecycle values, start guards, SQLite state persistence and enabled-project
classification now live under core/harness/daemon. Application and TUI consumers
call that owner directly. The constructor uses the existing database.DataDirectory
port; SQL text, state fields, timing, stop requests and failure messages are
unchanged. Four existing test files follow their owner; the real user-environment
integration stays beside its environment, and package fixtures use production
migration assets against owned in-memory databases. The 68 moved declarations and
tests retain their source preimages after namespace qualification.

Host executable resolution, streaming fingerprints, retained command-image
validation and worker executable checks now live under core/harness/host. The
parent harness owner contains the original pure binding copy. Native and app
consumers call each owner directly; no root forwarding API remains. The existing
large, symlink, FIFO, directory and oversized executable tests follow their host
owner. Native admission still owns its distinct omitted-command request case.

Application generation binding now returns one boundGeneration value containing
selected adapters and the admitted operation, without an operation context. Its
callers derive contexts from their own parents. Model IDs, source observations,
quiz transitions, guided reads and review dispatch receive that explicit context.
The guarded quiz/spec paths no longer copy their writer context into a service.
The original binding, protocol checks, fallback behavior and finite budgets remain.
A real Pi peer regression cancels one caller and then dispatches the same admitted
operation from a fresh caller; capacity peers retain independent cancellation.

Shared operator/userspace resolution now lives in workspace, and bounded model
traces live in harness/trace. This prerequisite removes the root dependency from
host trace recording. Existing configuration, redaction, payload retention,
rotation, path validation and environment contracts are unchanged; 49 declarations
and tests match their pre-move source after qualification. The root quiz integration
observes the retained llm.ndjson filename without a test-only exported constant.

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

A separate naming commit exposes workspace.ConfigPath and WithConfigLock for
those shared filesystem operations. Original error messages, authority decisions
and configuration interpretation remain at their existing boundaries. Native
upgrade and host consumers use the shared operations directly.

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

The signature disposition ledger records 31 fixed occurrences and seven cohesive
five-input contracts. All 38 assigned occurrences now have a concrete disposition.
Together with the two additional owned resolver contracts, these changes remove
30 exact debt allowances: 207 remain, with ceiling seven and one external
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

Configuration API naming check passed:

- `go test ./internal/core/workspace ./internal/core ./internal/app -run '^Test(WorkflowUpgrade|HarnessSelection|HarnessGeneration)' -count=1 -timeout=3m`

Daemon owner move checks passed:

- `go test ./internal/core/harness/daemon ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(ValidateDaemon|EvaluateDaemon|LoadDaemon|DaemonState|NewDaemon|Daemon|Leaf)' -count=1 -timeout=3m`: TUI compiled; lifecycle, persistence, native/application and dependency checks passed.
- `make check-go-style`: 9303 declarations, ceiling seven, 208 exact debts, one external exception.

A separate naming commit exposes State, Status, Repository, StartGuardDecision,
Project and their lifecycle operations under the daemon namespace. Direct
consumers and repository implementations use those names; retained strings and
SQL remain unchanged.

- `go test ./internal/core/harness/daemon ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(ValidateDaemon|EvaluateDaemon|LoadDaemon|DaemonState|NewDaemon|Daemon|Leaf)' -count=1 -timeout=3m`: all selected checks passed; TUI compiled.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/daemon ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues.

Executable owner move checks passed:

- `go test ./internal/core/harness/... ./internal/core ./internal/app ./scripts/goquality -run '^Test(HarnessBinding|HarnessGeneration|HarnessSelection|QualityRecordWire|ServiceHarness|Leaf)' -count=1 -timeout=3m`
- `make check-go-style`: 9303 declarations, ceiling seven, 208 exact debts, one external exception.

Executable API naming check passed:

The separate naming commit exposes host.FreezeExecutable, ExecutableHash,
CheckCommandImage, CheckExecutable and ValidateBindingCommand, and the pure
harness.CloneBinding. Twelve declaration/test preimages still match; direct
consumers and the baseline disposition follow the moved operation.

- `go test ./internal/core/harness/... ./internal/core ./internal/app ./scripts/goquality -run '^Test(HarnessBinding|HarnessGeneration|HarnessSelection|QualityRecordWire|ServiceHarness|Leaf)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues.

Generation caller context checks passed:

- `go test ./internal/app -run '^Test(HarnessGenerationCallerContext|HarnessGenerationBinding|HarnessJourney|PiGeneration|GuidedUseCase(Spec|Quiz|Plan|Admission)|TelegramHarness)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(AgentWork|StartAgent|Work|FastQuiz|Quiz|Generate|Spec|Plan|Dump)' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^Test(HarnessGenerationCallerContext|PiGenerationCapacity)$' -count=1 -timeout=3m`
- `make check-go-style`: 9304 declarations, ceiling seven, 207 exact debts, one external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/app`: zero issues.

Workspace and trace owner move checks passed:

- `go test ./internal/core/workspace ./internal/core/harness/trace ./internal/core ./internal/app ./scripts/goquality -run '^Test(ResolveProjectUserspace|EnsureProjectUserspace|IsValidOperatorID|LoadLLMTrace|WriteLLMTrace|RecordLLMTrace|HarnessGeneration|QualityRecordWire|Leaf)' -count=1 -timeout=3m`
- `make check-go-style`: 9304 declarations, ceiling seven, 207 exact debts, one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workspace ./internal/core/harness/trace ./internal/core ./internal/app ./scripts/goquality`: zero issues.

## Risks and Follow-ups

The slice remains in progress; PR readiness is not established. All assigned
signature occurrences are reconciled. Remaining work includes configuration/
selection and generation admission ownership, runtime host boundaries, the
remaining lifecycle/flag/wrapper review, and final exact-HEAD controller validation.
No Slice 10 PR has been opened.
