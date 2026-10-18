# Slice 2: Canonical Project Dump

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-2-project-dump`
PR: pending

## Purpose

Deliver one evolving project-wide dump and one Web capture form through the
existing metadata, authoring and application owners, following the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md). T2.1–T2.3 are implemented.

## Delivered Behavior

The first general capture creates a destination in the default ops namespace or
adopts the sole existing general identity without moving it. Later captures from
Web, CLI, TUI, bound authoring/Telegram and assisted synthesis extend that same
identity independently of active namespace. Named captures validate an existing
case and remain independent documents. Raw components and superseded aliases do
not appear as extra general catalog cards.

Project metadata retains an optional canonical dump reference and unknown
frontmatter. Ordinary descriptive saves cannot bind, rebind or clear it.
Read-only observation inventories canonical files without adoption, index rebuild
or journal recovery. Conflicting candidates, missing targets and orphaned raw
captures require explicit reconciliation.

Reconciliation previews choose the earliest valid creation timestamp, then path,
retaining invalid-time evidence. Observed apply preserves every original body and
metadata image in immutable history, composes attributed content, publishes the
canonical reference and marks superseded metadata with retained aliases. Old
references resolve to the canonical body; originals remain intact.

Capture, Append and full-body Save use retained operation arguments and guarded
publication. Save preserves frontmatter, unknown metadata, raw provenance and
exact prior body/metadata history. Unchanged retries return the original receipt
before inspecting later edits or invoking preparation. Stale or cross-project
bindings and changed arguments conflict. Post-commit refresh failures retain
committed hashes and an explicit pending projection instead of hiding the write.

Bound continuations publish dump, history, project reference, session,
transcript, provenance and receipt in one journal. Alias-aware snapshots guard
the resolved body, alias images and project metadata; native work descriptions
use the resolved body while retaining the session's own namespace. A
question-only session may bootstrap its first output only from its exact observed
raw inputs. Refinement requires an observed session output before provider dispatch.

Assisted synthesis joins its guarded prepared summary/session with the same
canonical publication. Root-bound semantic identity ignores preparation/update
timestamps and accumulated outputs/provenance. Committed images remain in
immutable history so bounded receipts can return original content after later
edits. CLI finalization resumes prepared curation or a confirmed receipt before
attempting initial curation.

The Web catalog now exposes one New action. Its modal offers Project-wide and
bounded existing case names, title and content. Each live form receives a capture
identity; unchanged submissions retain it, while edited submitted arguments create
a new intent. Validation and transport failures retain the form. Confirmed
success restores focus and refreshes the catalog; pending projections retain the
committed result and permit an unchanged refresh retry. Reads and scope selection
do not allocate artifacts or start generation.

## Implementation Notes

- T2.1 metadata/adoption: `a8bdec69d5595740a6fee1b723b1d2f1357a6ce4`.
- T2.2 reconciliation: `f4b080aeb4a2933b87535d56ca64e9fcb74f39ef`.
- T2.2 retained mutation owner: `b18c795c115463d37bd676e9968b9ae1e25e47f1`.
- T2.2 application/adapters: `14735ffbd247419294adcc9ef93c164fe63471bf`.
- T2.2 bound authoring: `ac1e17b9257ac2ee53e821b7c87876fdf5d1c4c6`.
- T2.2 assisted synthesis: `df5013c7629bd54d0ac2ecf49c55363a1b9c99ed`.
- T2.3 Web capture, documentation and changelog: `ce10f273f308a10faf48d72b5bebed9de733659d`.

One existing authoring owner and recovery journal govern publication; no nested
standalone dump transaction publishes ahead of its session. Catalogs read bounded
canonical inventory across namespaces instead of treating indexes as authority.
The original direct application dump writers are removed. Migration was exercised
only on isolated fixtures, never applied to live project data.

## Contracts Added or Changed

`ProjectMetadata.ProjectDumpRef`, `SaveContext`, `ObserveProjectDump`,
`AdoptProjectDump`, `PreviewProjectDumpReconciliation`, `ReconcileProjectDump`,
`ResolveProjectDumpReference`, `ReadDumpDocument` and `MutateDump` extend existing
owners. Application requests carry operation identity and observed revisions;
committed results retain body/metadata proof and projection state. Web capture
requires an identity; CLI edits retain `--operation-id` and `--revision`.

Journal v2 retains v1 decoding and its original limits. V2 allows the fixed
project metadata and content-addressed dump history paths, at most 128 changes
and 24 MiB encoded; each image remains bounded to 1 MiB. Inventory permits 4096
entries and 8 MiB aggregate images. Reconciliation accepts 2–32 general sources,
uses a 30-second timeout and bounds composition to 1 MiB. Alias lookup allows
16 links and rejects cycles, missing targets and raw/scoped destinations.

Dump preparation reserves root-bound arguments and references, caches successful
preparation and permits five attempts across restart, each bounded to 60 seconds.
A mutation has a 90-second deadline. Capture input is at most 64 KiB, edited bodies
1 MiB and titles 512 bytes. New operations respect shared request/receipt capacity;
confirmed retries replay first. A general capture revision binds full observation,
a named revision binds its case, and edits bind root/reference/body/metadata.

Assisted synthesis uses `chr-assisted-dump-receipt/v1` in the shared receipt
directory and a 90-second deadline. Its source guard is an unchanged journal
image, not a rewrite. The result's `Mutation` retains committed proof. Other
synthesis targets keep their existing writers.

## Files of Interest

- `internal/core/project_metadata*.go`, `project_dump*.go` and `dump_mutation*.go`: reference, inventory, reconciliation, aliases, retained preparation, edits and history.
- `internal/core/authoring_dump*.go` and `assisted_dump*.go`: session-integrated shared publication and retry receipts.
- `internal/app/service_dump*.go`, `service_project_dump.go`, `dump_cli.go`, `author_assisted_dump.go` and TUI adapters: application ownership and retained bindings.
- `internal/app/web_actions.go`, `web_artifacts.go` and `internal/webui`: one capture form, bounded scope choices and failure/pending-result retention.
- `docs/how-to/user-guide/index.md`, `docs/reference/authoring-sessions.md`, `docs/reference/web-ui-action-matrix.md` and `CHANGELOG.md`: user behavior and recovery contracts.

## Validation

Task development checks passed before their corresponding implementation commits:

- `go test -race ./internal/core ./internal/app -run '^(TestDumpMutation.*|TestProjectDump.*|TestProjectMetadata.*|TestParseProjectMetadata.*|TestService(ProjectDump|SaveProjectMetadata)|TestAuthoringSessionRepository.*|TestTelegramAuthoring(Owner.*|RespondRecovery|CommitGuard|ModelReceipt|InitialRollback)|TestService.*Dump.*)$' -count=1` — core mutation, metadata, reconciliation, history and affected recovery passed.
- `go test -race ./internal/core ./internal/app ./internal/tui ./internal/webui -run '^(TestDump.*|TestService.*Dump.*|TestProjectDump.*|TestTUIDump.*|TestModel.*Dump.*|TestDefaultCommandRegistryExposesDumpWorkflows|TestMouseClickSelectsDumpTextPane|TestWebServe(Artifact.*|WriteAction.*)|TestWebDump.*|TestTelegramAuthoring(BridgeLostAck|PreparedLostAck|ServiceObservedEdit|Work.*))$' -count=1` — application, adapters, projections and existing authoring replay passed.
- `go test -race ./internal/core ./internal/app -run '^(TestAuthoringDump.*|TestTelegramAuthoring.*)$' -count=1` — bound publication, concurrency, alias/source guards and native preparation passed. Subsequent focused core runs covered adoption and refinement without an output.
- `go test -race ./internal/core ./internal/app -run '^(TestAssistedDump.*|TestSynthesizeAssistedArtifact.*|TestAuthorAssisted.*|TestAuthoringDump.*)$' -count=1` — assisted publication, every changed replacement, source divergence, copied-root receipts and actual CLI retry passed.
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test -race ./internal/app ./internal/webui -run '^(TestWebDumpCapture.*|TestDumpCaptureResult|TestHandlerDumpsCreateUsesModalTemplates|TestDumpModalWrites)$' -count=1` — actual desktop/narrow modal, viewport bounds, lost acknowledgement, retained input/identity, edited intent, scope independence, focus restoration and pending projection responses passed.
- `go test -race ./internal/app ./internal/webui -run '^(TestWebDumpCapture|TestDumpCaptureResult|TestHandlerDumpsCreateUsesModalTemplates|TestDumpModalWrites|TestWebServeWriteActionCreatesDump)$' -count=1` — final capture service/HTTP regressions passed after fencing the existing capture fixture's provider.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/app ./internal/webui/...` — final Web task lint passed with zero issues. Core/application task lint also passed before their commits.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

An initial browser fixture inherited the default curation backend and timed out.
The corrected fixture substitutes deterministic curation while retaining real
HTTP, publication and unchanged assertions; the passing browser command above
uses that boundary. Owned browser/server fixtures stop after their checks.

Final candidate validation runs through Looper on the clean PR head, in this
order: `make verify`, focused race regressions, the owned capture/Board browser
checks and `git diff --check`. Looper retains the exact HEAD and ordered command
list; the canonical PR's Validation section records the successful final commands.
Development evidence does not establish scheduled exhaustive/nightly certification.

## Risks and Follow-ups

No live data reconciliation was performed. Existing ambiguous inventories require
explicit reconciliation. The controller owns integration into dev. Subsequent
editor/grid slices remain in their approved order and are not part of this PR.
