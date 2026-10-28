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

T12.1 and the compiling T12.2 relocation are implemented. The
[authority map](../../../plan/code-quality-remediation-slice-12-boundaries.md)
retains the frozen predecessor identities/hashes and records actual current
owners. All 79 direct root dependencies and 100 authority/recovery entries resolve;
all 908 methods move with their Go type. Final stage ownership remains recorded
separately for the subsequent approved extraction tasks.

The mechanical unit moves 2,710 production declarations across 353 files and
1,666 test declarations across 276 files. Compiler-resolved naming removes package
stutter from 236 public types/functions and updates implicit embedded fields.
All 4,342 moved declaration groups retain exact tokens after those names and owner
qualification. JSON field/tag order, enum strings, nil/empty behavior, error branches
and journal/lock order stay unchanged. Thirty-three shared fixture declarations
without private owner references remain in both test packages; private workflow
checks move beside their owner. Frozen wire fixture bytes are unchanged.

## Contracts Added or Changed

`ReadLegacyTicketFile` refuses adopted sources. `ReplaceLegacyTicketFile` takes
the existing repository or non-Git file lock, rechecks adoption and source equality,
and atomically replaces the whole image. These are actual legacy production
operations; recovery callbacks and mutable store internals remain private.

The workflow import guard permits its 19 actual neutral dependencies and rejects
root core and adapters. The more-specific workflow/state guard retains its narrow
bounds. Existing debt identities and owner-qualified signatures are explicitly
rekeyed without adding allowances: 9,305 production declarations, ceiling six,
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

The current mechanical batch passes:

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

T12.3 stage ports and policy separation, T12.4 assigned signatures/typed modes and
D08 receipt semantics, final caller reconciliation, exact-head validation and the
canonical PR remain outstanding. Slice 12 is active and has not been integrated.
