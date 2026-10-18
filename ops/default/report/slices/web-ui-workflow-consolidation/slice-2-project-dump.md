# Slice 2: Canonical Project Dump

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-2-project-dump`
PR: pending

## Purpose

Deliver one evolving project-wide dump through the existing authoring and
metadata owners, following the [plan](../../../plan/web-ui-workflow-consolidation.md)
and [ordered task map](../../../tracker/web-ui-workflow-consolidation.md).
T2.1 and the T2.2 core reconciliation owner are implemented. Shared capture
routing, application/adapter integration and T2.3 remain.

## Delivered Behavior

Project metadata supports optional `project_dump_ref`, a canonical project-relative
ops dump body reference. A read-only observation inventories dump artifacts across
namespaces, distinguishes raw components and case-scoped documents from general
dumps, and identifies an unambiguous existing destination. It reports missing,
raw, scoped or conflicting bound targets and retains invalid timestamp evidence.
It does not create a pointer, rebuild an index or recover a pending journal.

Explicit adoption binds the sole existing general dump under project authoring
exclusion. It retains the artifact's identity and namespace and leaves its body,
metadata, raw capture and history bytes intact. Multiple general dumps, orphaned
raw captures and missing targets require reconciliation. An empty project has no
existing dump to adopt; first-capture allocation belongs to T2.2.

Explicit core reconciliation previews propose the earliest valid creation time,
then relative path, while retaining missing/invalid timestamp evidence. Apply
compares the complete observation under authoring exclusion and retains every
source body and metadata image byte for byte in immutable artifact history. The
chosen identity receives the attributed composition; superseded bodies remain
unchanged with canonical-reference relationships in metadata. Reads resolve
bounded aliases and exclude them from the general candidate list.

A versioned extension of the existing authoring journal publishes the composed
body, history, alias metadata, project reference and immutable result receipt
together. A restarted owner recovers interrupted publication. Identical retries
return the original receipt, including after later edits or receipt capacity;
changed arguments or a reserved authoring identity conflict. Divergent source or
history bytes are retained and block publication/recovery.

Project metadata saves share the authoring owner and compare source hashes after
exclusion. Name/context edits retain the canonical reference and unknown
frontmatter. Ordinary metadata saves cannot bind, rebind or clear the reference.
Adoption rejects stale observations, including changed candidate membership,
body/metadata images and default project context derived from README. The revision
also binds the project root so byte-identical copies cannot reuse an observation.

## Implementation Notes

T2.1 commit: `a8bdec69d5595740a6fee1b723b1d2f1357a6ce4`.

T2.2 core reconciliation commit: `f4b080aeb4a2933b87535d56ca64e9fcb74f39ef`. The task remains open.

The prior dump service already serialized create/append/save through
`WithAuthoringMutation`, but project metadata saved outside that owner. Existing
general captures still allocate separate curated/raw pairs; the complete routing
change is T2.2. No reconciliation has been applied to live project data.

Observation uses canonical files and validated envelope identities. It rejects
symlink paths, non-regular files, malformed frontmatter, identity mismatches and
unsupported loose dump documents. Bounds are ten seconds, 4096 enumerated
namespace/artifact entries, 1 MiB per source file and 8 MiB of body/metadata images.
A pending authoring journal produces a conflict without read-side recovery.

The optional field retains schema version 1. Existing project fields, source-hash
checks and accepted spacing around frontmatter keys are preserved. Nested unknown
fields cannot shadow root project fields. Adoption publishes project metadata atomically. Reconciliation uses journal
`chr-authoring-journal/v2`, retaining v1 decoding and its original limits. Version
2 permits the fixed project metadata file and content-addressed dump history
paths, at most 128 changes and 24 MiB encoded. Individual images stay bounded to
1 MiB. Preview/apply accept 2–32 general sources; the composed body is bounded to
1 MiB. Alias lookup accepts at most 16 links and rejects cycles, missing targets
and raw/scoped destinations. Reconciliation times out after 30 seconds. New dump
operations require fewer than 4096 existing receipts; confirmed retries replay
first. The shared bounded directory reader treats an empty directory as capacity.

## Contracts Added or Changed

`ProjectMetadata.ProjectDumpRef` and `ProjectMetadataService.SaveContext` extend the
existing metadata owner. Its explicit `AdoptProjectDump` operation compares a
complete observation while holding the authoring owner. The reference's namespace
and artifact identity derive from its validated body path.

`ObserveProjectDump` returns the project record, general candidates, raw components,
selected destination, retained aliases, reconciliation reason and semantic revision. `ReadProjectDump`
and `AdoptProjectDump` expose these operations through the central application
service. Existing capture adapters have not yet been routed through them.

`PreviewProjectDumpReconciliation`, `ReconcileProjectDump` and
`ResolveProjectDumpReference` extend the core authoring owner. Reconciliation
requires operation ID, expected observation revision and exact proposed target.
Its receipt binds the project root and arguments to immutable committed body and
metadata hashes. Pending authoring requests reserve their operation identities.
These operations are not yet exposed by the central application service.

## Files of Interest

- `internal/core/project_metadata.go` and `project_metadata_fields.go`: persisted reference, source checks, serialization and preserved frontmatter.
- `internal/core/project_dump.go`: reference validation, observation and guarded adoption.
- `internal/core/project_dump_read.go` and `project_dump_artifact.go`: bounded cross-namespace inventory and component classification.
- `internal/core/project_dump_reconciliation.go`, `project_dump_composition.go`, `project_dump_receipt.go`, `project_dump_alias.go` and `project_dump_journal.go`: deterministic migration, immutable history/receipts and bounded journal/alias contracts.
- `docs/reference/authoring-sessions.md`: project dump and recovery format reference.
- `internal/app/service_project_dump.go` and `service_metadata.go`: shared application entrypoints and context propagation.
- `internal/core/project_dump_test.go`, `project_dump_validation_test.go` and `internal/app/service_project_dump_test.go`: isolated persistence, ownership and service regressions.

## Validation

T2.1 task development checks passed on the implementation tree before its commit:

- `go test -race ./internal/core ./internal/app -run '^(TestProjectDump.*|TestProjectMetadata.*|TestParseProjectMetadata.*|TestService(ProjectDump|SaveProjectMetadata)|TestAuthoringSessionRepository.*|TestTelegramAuthoringOwner.*|TestService.*Dump.*)$' -count=1` — passed. Covers adoption across namespaces, raw/history preservation, read-only and pending-journal behavior, stable revisions, stale and cross-project binding, concurrent metadata writes, unsafe/invalid sources, cancellation and affected authoring/dump regressions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

T2.2 core task development checks passed on its implementation tree before commit:

- `go test -race ./internal/core ./internal/app -run '^(TestProjectDump.*|TestProjectMetadata.*|TestParseProjectMetadata.*|TestService(ProjectDump|SaveProjectMetadata)|TestAuthoringSessionRepository.*|TestTelegramAuthoring(Owner.*|RespondRecovery|CommitGuard|ModelReceipt|InitialRollback)|TestService.*Dump.*)$' -count=1` — passed. Covers deterministic timestamp ordering, original/unknown-field preservation, all twelve replacement interruption points, large multi-image recovery, divergent recovery, stale sources/targets/history, concurrent confirmations, alias cycles/depth/ownership, receipt conflicts/capacity and affected v1 authoring regressions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Final clean-HEAD slice validation through Looper, including `make verify`, has not
run. These task checks do not establish complete Slice 2 or nightly evidence.

## Risks and Follow-ups

T2.2 remains open for recoverable first-capture publication, general capture and
Save request/receipt binding, all writer routing (including assisted authoring
and Telegram continuations), shared application reconciliation operations and
catalog/old-reference integration. The core observation hides aliases, but the
existing application catalog and readers still require routing to those rules.
Post-commit index refresh must return committed-content/projection-stale evidence.
T2.3 remains: one Web New capture form with Project-wide and existing named cases.
No reconciliation has been applied to live project data. The source ticket stays
in progress until all Slice 2 contracts are delivered.
