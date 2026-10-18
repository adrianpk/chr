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
T2.1, the T2.2 core owners and application/CLI/TUI/Web edit integration are
implemented, along with bound authoring/Telegram dump continuations. Assisted
synthesis and T2.3 remain.

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

The core dump mutation API resolves the first general capture in the default
namespace or adopts the sole existing general identity. First publication includes
body, raw component, metadata, project reference and receipt in one recoverable
journal. Later general captures extend that identity under authoring exclusion.
Named captures validate existing cases, retain the selected namespace and create
independent documents. Legacy `case_slug` declarations remain scoped records.

The owner reserves operation arguments and resolved references before preparation,
retains a successful prepared capture before publication, and returns confirmed
receipts before inspecting later edits. Full-body Save and Append compare an
observed document revision inside exclusion. Save replaces editable content while
retaining frontmatter, unknown metadata, capture provenance and the exact prior
body/metadata history. Alias edits target the canonical document while retaining
the superseded original. The central application service now routes capture, Save and Append through
these operations. Bound authoring and Telegram continuations also use the destination. Assisted
synthesis remains to be routed.

Project metadata saves share the authoring owner and compare source hashes after
exclusion. Name/context edits retain the canonical reference and unknown
frontmatter. Ordinary metadata saves cannot bind, rebind or clear the reference.
Adoption rejects stale observations, including changed candidate membership,
body/metadata images and default project context derived from README. The revision
also binds the project root so byte-identical copies cannot reuse an observation.

The application exposes reconciliation preview/apply and reads catalogs directly
from bounded canonical inventory across namespaces. Raw components and superseded
aliases are hidden; independent scoped documents remain visible. Old references
resolve to the canonical body in CLI, TUI and Web reads. Metadata update timestamps
support catalog display without depending on a particular index namespace.

Application writes return committed identity and body/metadata hashes even when
post-commit index/catalog refresh fails. A ten-second refresh deadline and
projection-stale result preserve the confirmed write; an identical retry refreshes
without another capture or provider invocation. Returned content/revision matches
the committed hashes, retaining later independent edits on receipt replay.

CLI edit callers can retain `--operation-id` with `--revision`, obtained from
`dump show --revision`; fresh edits observe before dispatch. TUI editors retain
operation/revision bindings across an unchanged failed Save and allocate a new
intent when the text changes. Web detail forms bind Save/Append to the observed
document and look up receipts before stale-source checks. The two Web capture
forms still await T2.3's single New flow.

Bound authoring and Telegram dump outputs join canonical content, retained raw
inputs, history, project reference, session, transcript, provenance and receipt
in one v2 journal. Other artifact targets retain their v1 publication format.
Later sessions append to the shared identity across namespaces; explicit
refinement replaces its observed editable body and retains the previous images.
No nested standalone dump operation publishes ahead of the session transaction.

Session observations resolve retained aliases to the canonical body while guarding
both original alias images and the destination. Their source hash also binds
project metadata. Providers and proposed native work use the resolved body while
session-owned raw captures, quizzes, specs, plans and records retain their own
namespace. An independent shared edit requires a fresh observation. The owner
also rechecks the complete dump inventory after provider preparation.

A question-only creation's bound continuation may bootstrap its first dump only
when all raw-only components are its exact observed inputs. Unrelated orphan
captures still require explicit reconciliation. Original raw bytes remain intact.
Refinement without a session output fails before provider dispatch.

## Implementation Notes

T2.1 commit: `a8bdec69d5595740a6fee1b723b1d2f1357a6ce4`.

T2.2 core reconciliation commit: `f4b080aeb4a2933b87535d56ca64e9fcb74f39ef`.
T2.2 core mutation commit: `b18c795c115463d37bd676e9968b9ae1e25e47f1`.
T2.2 application routing commit: `14735ffbd247419294adcc9ef93c164fe63471bf`. The task remains open.
T2.2 bound authoring integration commit: `ac1e17b9257ac2ee53e821b7c87876fdf5d1c4c6`.

The prior dump service serialized create/append/save but allocated separate
general pairs and performed direct file writes. Those application writers are
removed. Capture curation now runs inside the retained core preparation callback;
its original input and successful output survive retry. Named captures retain
existing case validation and namespace selection. No reconciliation has been
applied to live project data.

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

Dump mutation requests share the existing request/receipt directories and v2
journal. A retained typed request binds the root, caller input and resolved
destination; successful preparation is cached before publication. The five-attempt preparation
budget survives reopening; each preparation has a 60-second timeout and each
mutation has a 90-second timeout.
Request JSON is bounded to 8 MiB, capture input to 64 KiB and edited bodies to
1 MiB; canonical images retain the existing 1 MiB limit. A supplied general capture
revision binds the full project observation, a named capture revision binds its
case metadata, and Save/Append require a root/document/body/metadata revision.

## Contracts Added or Changed

`ProjectMetadata.ProjectDumpRef` and `ProjectMetadataService.SaveContext` extend the
existing metadata owner. Its explicit `AdoptProjectDump` operation compares a
complete observation while holding the authoring owner. The reference's namespace
and artifact identity derive from its validated body path.

`ObserveProjectDump` returns the project record, general candidates, raw components,
selected destination, retained aliases, reconciliation reason and semantic revision. `ReadProjectDump`
and `AdoptProjectDump` expose these operations through the central application
service. The application catalog and write operations share this observation and owner.

`PreviewProjectDumpReconciliation`, `ReconcileProjectDump` and
`ResolveProjectDumpReference` extend the core authoring owner. Reconciliation
requires operation ID, expected observation revision and exact proposed target.
Its receipt binds the project root and arguments to immutable committed body and
metadata hashes. Pending authoring requests reserve their operation identities.
The central application service exposes preview/apply and committed-result
projection handling.

`MutateDump`, `DumpMutationRequest`, `DumpCapturePreparer` and `ReadDumpDocument`
provide the shared core capture/edit boundary. The mutation rejects changed
arguments, stale or cross-project document bindings, unresolved general ownership,
unknown cases, raw edits and provider scope/type/component-link changes. It can
be nested under an existing authoring claim; bound continuations still need to
include the prepared changes in their own session/receipt journal. This integration
is implemented for bound continuations; assisted synthesis remains pending.

## Files of Interest

- `internal/core/project_metadata.go` and `project_metadata_fields.go`: persisted reference, source checks, serialization and preserved frontmatter.
- `internal/core/project_dump.go`: reference validation, observation and guarded adoption.
- `internal/core/project_dump_read.go` and `project_dump_artifact.go`: bounded cross-namespace inventory and component classification.
- `internal/core/project_dump_reconciliation.go`, `project_dump_composition.go`, `project_dump_receipt.go`, `project_dump_alias.go` and `project_dump_journal.go`: deterministic migration, immutable history/receipts and bounded journal/alias contracts.
- `internal/core/dump_mutation*.go`: retained capture preparation, destination selection, source-bound edits and recoverable content/history publication.
- `internal/core/authoring_dump.go`, `authoring_dump_snapshot.go` and the existing continuation owner: joined session/dump publication, alias-aware context and shared-source guards.
- `docs/reference/authoring-sessions.md`: project dump and recovery format reference.
- `internal/app/service_dump_owner.go`, `service_dump_catalog.go`, `service_project_dump.go`, `dump_cli.go` and `tui_adapters_dump_quiz.go`: application mutation, catalog, explicit reconciliation and adapter binding.
- `internal/app/web_actions.go`, `web_artifacts.go`, `internal/tui/dump_actions.go` and the existing Web detail forms: revision-bound edits and receipt-aware results.
- `internal/core/project_dump_test.go`, `project_dump_validation_test.go` and `internal/app/service_project_dump_test.go`: isolated persistence, ownership and service regressions.

## Validation

Latest core task development checks passed on the implementation tree before
`b18c795c115463d37bd676e9968b9ae1e25e47f1`:

- `go test -race ./internal/core ./internal/app -run '^(TestDumpMutation.*|TestProjectDump.*|TestProjectMetadata.*|TestParseProjectMetadata.*|TestService(ProjectDump|SaveProjectMetadata)|TestAuthoringSessionRepository.*|TestTelegramAuthoring(Owner.*|RespondRecovery|CommitGuard|ModelReceipt|InitialRollback)|TestService.*Dump.*)$' -count=1` — passed. Covers first/repeated/concurrent core capture, namespace changes, existing adoption, independent named captures and stale case bindings; every first/capture/append/Save replacement interruption; retained preparation/cancellation/budget; stale/cross-project/alias edits; preserved originals and unknown fields; deterministic reconciliation, divergent recovery and bounded aliases/receipts; affected v1 authoring and existing application dump regressions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Application integration development checks passed before `14735ffbd247419294adcc9ef93c164fe63471bf`:

- `go test -race ./internal/core ./internal/app ./internal/tui ./internal/webui -run '^(TestDump.*|TestService.*Dump.*|TestProjectDump.*|TestTUIDump.*|TestModel.*Dump.*|TestDefaultCommandRegistryExposesDumpWorkflows|TestMouseClickSelectsDumpTextPane|TestWebServe(Artifact.*|WriteAction.*)|TestWebDump.*|TestTelegramAuthoring(BridgeLostAck|PreparedLostAck|ServiceObservedEdit|Work.*))$' -count=1` — passed. Covers shared application concurrent/retried capture, canonical catalog/alias reads, explicit reconciliation, projection failure/retry, CLI revision handling, TUI binding retention, Web receipt-before-stale behavior and affected authoring replay/work guards. Authoring fixture edits retain the exact persisted image and verify it survives replay.
- `go test -race ./internal/app -run '^(TestServiceDump.*|TestWebDumpReceipt|TestTUIDumpAdapterUsesDumpServices|TestDumpCLIWriteBinding)$' -count=1` — passed after adding the final post-commit refresh deadline.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app ./internal/tui ./internal/webui/...` — passed with zero issues.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Bound authoring development checks passed before the integration commit:

- `go test -race ./internal/core ./internal/app -run '^(TestAuthoringDump.*|TestTelegramAuthoring.*)$' -count=1` — passed. Covers all affected bound authoring and Telegram continuations, native preparation, authority guards and receipt replay, with new concurrent namespace, alias, stale-source, provider-edit and publication-fault regressions.
- `go test -race ./internal/core -run '^(TestAuthoringDump.*|TestTelegramAuthoringOwner.*)$' -count=1` — passed after adding legacy adoption and tightening refinement ownership.
- `go test -race ./internal/core -run '^TestAuthoringDump' -count=1` — passed after adding refinement-without-output rejection. Publication faults cover every changed image of first capture, later capture and refinement; unchanged transcript images do not perform a replacement.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app` — passed with zero issues on the final task tree.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Final clean-HEAD slice validation through Looper, including `make verify`, has not
run. These task checks do not establish complete Slice 2 or nightly evidence.

## Risks and Follow-ups

T2.2 remains open for assisted authoring synthesis. Its artifact, session,
provenance and receipt changes must join the shared dump transaction. Application,
CLI, TUI, Web edits and bound authoring/Telegram continuations now use the canonical
owner. T2.3 remains: one Web New capture form with Project-wide and existing named
cases, including a retained per-form capture identity. Final controller validation
and the canonical slice PR remain pending. No reconciliation has been applied to
live data. The source ticket stays in progress until all Slice 2 contracts are
delivered.
