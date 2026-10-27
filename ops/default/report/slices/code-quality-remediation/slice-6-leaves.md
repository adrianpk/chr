# Slice 6: Gate, notifications and project packages

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-6-quality-leaves`
PR: `#443`

## Purpose

Extract gate, notification persistence and registered project identity behind
narrow dependencies while preserving retained records and operation authority.

## Delivered Behavior

Gate owns configuration parsing, Git evidence, configured validation,
classification and report encoding. Explicit config paths and a loaded-report
port remove root storage dependencies; application composition still allocates
report artifacts and refreshes metadata. Named outcomes retain parsed-line
state, distinct refs, partial range evidence and ordered missing-report evidence.

Project owns the real registry, active selection and diagnostics. Directory
capabilities and the shared database owner preserve schema, timestamps, ordering,
active-selection bytes and atomic replacement. Root binding retains registered
and resolved paths in the same digest. Broader host/workspace configuration and
authoring transactions keep their existing responsibilities.

Notification owns events/configuration, SQLite admission, frozen lookup,
conditional reserve/completion and retained-document cleanup. Named lookup and
completion values preserve legacy/frozen distinctions, target binding,
acknowledgement, safe failures, uncertainty and original dispatch allowances.
Receiver storage and HTTP sending remain direct consumers. Shared immutable
message values, content hashing, namespace spelling and the existing conflict
sentinel resolve the concrete dependency graph without root facades.

All 19 assigned baseline occurrences have precise dispositions: 16 fixed and
three cohesive five-input contracts. Six authoring tuple results now name byte
counts, file images, partial changes, receipt absence and parser outcomes; receipt
callbacks keep the outer result/error under the original mutation lock. Temporary
signature debt decreases from 425 to 410 with ceiling seven and the same external
exception. Production import bounds now enforce the actual leaf dependencies.

## Implementation Notes

Separate mechanical commits preserve 125 gate, 39 project and 76 notification
complete declaration texts before qualified-name and operation/result changes.
Twenty-seven shared message declarations have normalized-token proofs including
literals/tags. Distinct follow-up commits change signatures and direct callers.
The payload prerequisite shares pure rendering/validation; it moves no receiver
claim, transport or storage transaction into notification persistence.

The shared database package retains original migration assets and failure handling.
Its static cross-table document quota predicate executes inside each owner's
existing transaction. Public repository fixtures retain their owned SQL connection;
the retention fixture triggers real admission and preserves every original image,
receipt and hash assertion. Frozen wire and old SQLite fixtures remain intact.

The finite file ledger accounts for 699 current production files under core.
Actual receivers establish project-dump/metadata transactions as authoring behavior
for the recorded Slice 9 move; their assigned Slice 6 signatures are resolved here.
The source review, detections and database schema assets are unchanged.

## Contracts Added or Changed

Direct gate, project and notify APIs replace root implementations. Locally owned
notification interfaces change with implementations. Context stays an explicit
operation argument. Typed gate report/action and event/status/source/severity
values preserve original spellings. Read-only-close suppressions retain checked
scanner/iteration errors; normal missing results remain separate from failures.
Reserve's plain flag retains the single definite-format fallback contract.

The existing AST tool checks production imports across generated files, every
build variant and descendants. Alias/dot/blank imports cannot bypass owner bounds.
Focused fixtures reject root/coordinator/adapter imports and nested bypasses,
while allowing actual shared values, external drivers and application composition.

## Files of Interest

- `internal/core/gate/`, `internal/app/gate_artifacts.go`: explicit gate inputs,
  real child/Git behavior and application-owned report composition.
- `internal/core/project/`, `internal/core/database/`: registry/selection and
  unchanged durable schema, configuration and transaction failure handling.
- `internal/core/notify/`, `internal/core/telegram/message/`: conditional event
  delivery, retained payload contracts and independent transport consumers.
- `internal/core/project_dump_results.go`, receipt and mutation callbacks:
  precise authoring outcomes under the existing durable owner.
- `scripts/goquality/boundaries.go`, `boundaries_test.go`: enforced import bounds.
- [All 19 dispositions](../../../plan/code-quality-remediation-slice-6-dispositions.csv),
  [mechanical moves](../../../plan/code-quality-remediation-slice-6-moves.csv),
  [payload proofs](../../../plan/code-quality-remediation-slice-6-payloads.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-6-signatures.csv).
- [Current owner ledger](../../../plan/code-quality-remediation-core-owners.csv),
  [dependency/glossary decisions](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./internal/core/notify ./internal/core ./internal/app -run '^Test(Notification|NewNotification|TelegramPresentation|TelegramNotification|QualityStorage|QualityRecordWire)' -count=1 -timeout=3m`
  — passed after the mechanical notification move; final signature changes also
  passed the core/app selections, covering real SQLite/HTTP and retained records.
- `go test ./internal/core -run '^Test(ProjectDump|ProjectMetadata|CaseMetadata|WorkspaceGraph|AuthoringDump|DumpMutation|QualityRecordWire)' -count=1 -timeout=3m`
  — passed for the authoring result conversions and original recovery/receipt checks.
- `go test ./scripts/goquality -count=1` — passed including dependency fixtures.
- `go run ./scripts/goquality` — passed with 410 debts and one external exception.
- `make check-go-style` — passed after the notification signature changes.
- Pinned strict lint passed for core/app/extracted packages and quality tooling.
  Original focused gate/Git/child, registry/selection, SQL and maintenance checks
  passed at their task checkpoints.

Initial controller `make verify` on `7a78164e6632720d4d38b87b15aa28b57475c827`
stopped at four import/type-layout formatting errors. Those files are corrected;
pinned core/TUI lint passes. The retained controller diagnostics identify that
failed candidate, and the following final batch validates its replacement.

Final validation runs through the controller after the PR-reference commit:
make verify, focused non-short leaf/core/application regressions, focused retained
notification/authoring concurrency and recovery race checks, then git diff --check.
The controller retains the exact clean HEAD, ordered commands, results and logs;
that record and PR Validation own final results. Integration remains pending.

## Risks and Follow-ups

None recorded.
