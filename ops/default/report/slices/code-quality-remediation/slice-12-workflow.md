# Slice 12: Workflow state, store and coordinator package

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-12-quality-workflow`
PR: pending

## Purpose

Establish the workflow state and persistence boundary with one store/locking
and recovery authority, direct consumers and cohesive stage contracts.

## Delivered Behavior

`internal/core/workflow` now owns the actual `Store` and all 908 receiver methods.
Core composition, application operations and adapters use that package directly.
The repository/file locks, source-image and revision guards, journals, atomic
multi-record writes and recovery remain at one owner. Legacy consumers call the
real guarded read/replace operations; no type bridge or test-only export is added.

Shared review/cycle values live in workflow/state. Ticket queries, Board/document
projections and pure frontmatter edits use their existing ticket owners. The
complete parsed configuration image, raw revision, explicit-key presence,
defaults and guarded harness-selection recovery use workspace/config.

## Implementation Notes

T12.1, the compiling T12.2 relocation and T12.3 stage operation boundaries are implemented. The relocation commit
is `20ca6c3e508b7c0964baaef4f345c87199e3d81b`. The
[authority map](../../../plan/code-quality-remediation-slice-12-boundaries.md)
retains the frozen predecessor identities/hashes and records actual current
owners. All 79 direct root dependencies and 100 authority/recovery entries resolve;
all 908 methods move with their Go type. Final stage ownership remains recorded
separately for the subsequent approved extraction tasks.

The T12.2 mechanical unit moves 2,710 production declarations across 353 files and
1,666 test declarations across 276 files. Compiler-resolved naming removes package
stutter from 236 public types/functions and updates implicit embedded fields.
All 4,342 moved declaration groups retain exact tokens after those names and owner
qualification. JSON field/tag order, enum strings, nil/empty behavior, error branches
and journal/lock order stay unchanged. Thirty-three shared fixture declarations
without private owner references remain in both test packages; private workflow
checks move beside their owner. Frozen wire fixture bytes are unchanged.

T12.3 implementation is `9c74db45a790d03f7d710b2ffac2e7dcdc4d5e88`. It moves 60 shared protocol types and both evidence-enumeration methods into
workflow/state. All 62 declaration token sequences match after qualification and
the EvidenceRefs operation name; public type/embedded-field names, field/tag order,
wire strings and nil/empty representations stay unchanged. Six cross-package
NativeUnsupported literals now name their original Condition/Decision fields with
the same values and evaluation order.

Five consumer-owned contracts are connected to actual production calls:
integrate.Validation, merge.Evidence, review.Decisions, forge.Responses and
closeout.Gate. Private workflow implementations retain the claim/prepared/provider,
closing claim or journaled invocation/submission identity; the store implements
merger evidence checking directly. The [port map](../../../plan/code-quality-remediation-slice-12-ports.csv)
records each binding and its approved subsequent policy extraction task. Locks,
source/revision checks, cleanup, error joins and atomic acceptance remain in the
existing operations. Stage interfaces contain no coordinator or mutable-store type.

T12.4's first batch (`fe0df936e6152c0ad346e1e680ad2dd246825e60`) replaces ten assigned declarations (11 baseline rows) with
named mutation normalization/preparation/lookup, materialization replay and task
results. Set construction takes an explicit input and artifact number/time form
one allocation value; task numbering, timestamps, nil/empty results and replay
presence on errors retain their existing meaning. Authoring and amendment/adoption
callers consume the actual named contracts directly.

D08 shares only contained receipt reads, absence and typed JSON decoding. Record
and control schemas, checksum field order, request/operator/revision guards and
receipt writers remain separate at the original owner. Named receipt/reservation
states retain the exact `reserved` and `confirmed` bytes. Ten new table cases
exercise both real readers against confirmed, missing, malformed, null and the
other operation's correctly checksummed receipt. The four assigned mutation tuple
allowances and six materialization/adoption allowances are removed explicitly.
Current policy: 9,317 declarations, ceiling six, 152 debts and one external exception.
Thirty-seven assigned declarations and the retained cohesion-watch reconciliation
still need T12.4 work; this batch does not complete that task.

## Contracts Added or Changed

`ReadLegacyTicketFile` refuses adopted sources. `ReplaceLegacyTicketFile` takes
the existing repository or non-Git file lock, rechecks adoption and source equality,
and atomically replaces the whole image. These are actual legacy production
operations; recovery callbacks and mutable store internals remain private.

The workflow import guard permits its existing neutral dependencies and five
consumer contracts, and rejects root core and adapters. Each stage contract permits
only shared workflow/state values; 20 new AST cases reject coordinator, filesystem
mutation and configuration imports while retaining the valid data dependency. The more-specific workflow/state guard retains its narrow
bounds. Existing debt identities and owner-qualified signatures are explicitly
rekeyed without adding allowances: 9,316 production declarations, ceiling six,
162 exact debts and one external exception. The 71 assigned baseline rows still
map to 68 declarations; their signature remediation remains T12.4 work.

Acceptance/fault manifests retain their IDs, order and 78 distinct tests while 82
bindings follow the moved sources. Native/live command selections target the
actual workflow package. The frozen controller gate is unchanged.

## Files of Interest

- `internal/core/workflow/workflow_store.go`, `workflow_lock.go` and `workflow_transition.go`.
- `internal/core/workflow/workflow_legacy_write.go` and colocated recovery tests.
- `internal/core/workflow/state/` and `internal/core/workspace/config/`.
- The method, boundary, dependency, signature and move CSVs linked from the authority map.

## Validation

The T12.4 receipt/materialization batch passes:

- `go test ./internal/core/workflow ./internal/app -run '^Test(OperationReceiptGuards|TelegramOperationOwner[A-Za-z]+|TelegramControlReceiptMissing|TelegramControlJournal|TelegramControlLifecycleJournal|TelegramAuthoringWork[A-Za-z]+|WorkflowMaterialization[A-Za-z]*|WorkflowPlanAdoption[A-Za-z]*|WorkflowAmendment[A-Za-z]*)$' -count=1 -timeout=20m -v`: 30 top-level checks pass through the existing fixture-owning runner; real filesystem/Git replay, crash/restart, conflict, concurrent owner, adoption/amendment, control journal and authoring guards retain their behavior. No owned descendants remain.
- `go test ./internal/core/workflow -run '^TestOperationReceiptGuards$' -count=1 -timeout=3m`: pass after separating the test fixture's image replacement from its table, preserving all ten cases.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues after test-complexity and whitespace corrections.
- `make check-go-style`: pass; unchanged ceiling six, 152 exact debts and one external exception.
- `git diff --check`: pass.

The T12.3 boundary batch passes:

- `go test ./internal/core/workflow/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|QualityStorage[A-Za-z]+|LeafDependencies|WorkflowPreparationMethods|WorkflowPreparationGuard|NativeMergerExecution|NativeMergerAdmission|WorkflowReviewModes|WorkflowReviewStop|WorkflowSubmissionMethods|WorkflowSubmissionLostResponse|WorkflowClosingEmpty|WorkflowClosingStop)$' -count=1 -timeout=20m -v`: real preparation/guard, merger, review/stop, publication/lost-response, closing, frozen wire/SQLite and dependency regressions pass through the existing fixture-owning runner; TUI compiles.
- `go test ./internal/core/workflow ./scripts/goquality -run '^Test(NativeCapabilityDenials|NativePreparationAdmission|NativeReviewPolicy|LeafDependencies)$' -count=1 -timeout=3m`: policy/capability/admission and dependency cases pass after the field-name fix.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues.
- `make check-go-style`: pass; ceiling six, 162 exact debts and one external exception remain unchanged.
- `scripts/check-package-readmes.sh`: pass.

The T12.2 mechanical batch passed before this boundary extraction:

- `go test ./internal/core/workflow ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|QualityStorage|WorkflowRoundTrip|WorkflowClaimFence|WorkflowClaimJournal|WorkflowMaterializationRecovery|WorkflowUpgradeRecovery|WorkflowLegacyEntry|WorkflowAdoption|WorkflowRelocation|Ticket|LeafDependencies)' -count=1 -timeout=3m`: frozen wire/SQLite storage, real Git/filesystem, adoption, claim fencing/journals, materialization/upgrade recovery, ticket/legacy operations, app and boundary checks pass; TUI compiles.
- `go test ./internal/core/workflow -run '^Test(WorkflowSubmissionJournal|WorkflowSliceJournal|WorkflowReviewJournal|WorkflowFinalizationCrash|WorkflowClosingJournal|WorkflowTrackerCrash|WorkflowPublicationCrash)$' -count=1 -timeout=20m -v`: all seven journal/crash-recovery checks pass through the existing QA fixture-owning runner; owned descendants finish cleanly.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues.
- `make check-go-style`: pass with the unchanged limits above.
- `python3 -m unittest discover -s scripts/qa -p 'test_nightly.py'`: eight runner/admission/evidence regressions pass; no exhaustive gate executes.
- `scripts/check-package-readmes.sh`: pass.

Compiler-resolved identity coverage and all 4,342 declaration token comparisons
pass. Acceptance/fault source validation retains the same 78 test selections.
Earlier prerequisite batches also passed ticket/configuration component suites
and native Git/filesystem/wire/storage/app regressions. A prior overlap with strict
lint hit the existing one-second configuration deadline fixture; unchanged dev and
the same candidate selection after lint passed. Its bounds/assertions are unchanged.
The seven-test recovery batch initially exceeded an aggregate three-minute test
budget; the same tests passed with the existing QA 20-minute budget. Production
deadlines and assertions are unchanged. Final exact-head controller validation
has not run.

## Risks and Follow-ups

T12.4 remaining assigned signatures/typed modes and cohesion-watch decisions,
final caller reconciliation, exact-head validation and the
canonical PR remain outstanding. Slice 12 is active and has not been integrated.
