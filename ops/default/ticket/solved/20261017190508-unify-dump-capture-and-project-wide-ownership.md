---
id: TKT-20261017190508
title: "Unify dump capture around one evolving project-wide dump"
status: solved
kind: task
severity: medium
priority: normal
scope: domain
tags: dumps, project-context, authoring
source: chat
reported_at: 2026-10-17T19:05:08Z
commits: [7f4b82f629696b39dc2f1a4326f90a11459db1d3, 78b9165e07963d74ab856b68a223034f39e4a609, 7bd9c09ad018e279d04726844e62896b0c365908, 67dfdbd3937618b1003c914784ee952418f895ba]
branch: feat/slice-2-project-dump
pr: "#399"
resolution: fixed
reviewed_at: 2026-10-18T16:20:03Z
closed_at: 2026-10-18T16:20:03Z
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Unify dump capture around one evolving project-wide dump

## Context

Finding from [the Web UI review](../in_progress/20261017105432-review-draft-web-ui-usability.md).
The catalog currently supports many separate project-wide dumps and exposes
separate New Dump and For Use Case buttons. The maintainer wants project-wide
context to accumulate in one evolving dump rather than fragment into many
small general captures. Material for an individual use case belongs to that
case's scope.

This is a shared application behavior change with a matching Web UI change.
Capturing this ticket does not authorize implementation, data migration,
service changes or creation of a specification or delivery plan.

## Required Behavior

1. Each project has one canonical logical project-wide dump. The first general
   capture creates it; subsequent general captures append to or extend that
   existing dump rather than allocate another project-wide dump.
2. Editing changes the same existing dump. Preserve its identity and references;
   do not interpret saving a complete edited body as appending a duplicate copy
   of that body or as creating another dump.
3. Keep project-wide and use-case-specific material in their selected scopes.
   A use-case capture selects an existing use case and retains that association.
   Do not create a use case implicitly or merge its material into project-wide
   context merely because that general dump already exists.
4. Replace New Dump and For Use Case with one Web UI action labeled New.
5. The New action opens one capture form with a scope selector offering
   Project-wide and the project's existing use cases.
6. The selected scope determines the destination. Selecting Project-wide must
   reuse the general dump when present, even though the entry action is New.
7. Apply the uniqueness and destination rule at the shared authoring/write
   boundary. Hiding a Web button alone must not leave other entrypoints able
   to create additional project-wide dumps.
8. Reuse existing mutation ownership and serialization so overlapping captures
   cannot create two general dumps or lose another contribution. Keep the same
   canonical identity across application restarts and index rebuilds.

## Existing Paths to Review

- `internal/app/service_dump.go`: CreateDump, AppendDump and SaveDump requests;
  creation currently allocates dump-pair files for a supplied title and scope.
- `internal/app/service_dump_owner.go`: the existing authoring mutation boundary.
- `internal/app/dump_cli.go` and `internal/app/tui_adapters_dump_quiz.go`:
  CLI and TUI creation, append and editing adapters.
- `internal/app/web_actions.go`: Web dump creation, save and append operations.
- `internal/webui/assets/templates/shell.html`: separate creation buttons,
  dialogs and scope fields.
- Authoring and Telegram generation/continuation writers: determine how they
  produce or refine dumps and ensure they cannot bypass the canonical general
  destination. Their detailed behavior has not yet been assessed for this ticket.

## Existing Data and Model Decisions

Multiple general dumps already exist. Before implementation, define how to
select or establish the canonical project-wide dump and reconcile older general
captures without losing content, provenance or incoming references. Do not
silently choose the newest record, delete the others or overwrite existing
content as an incidental UI operation.

Resolve canonical identity through the existing operational artifact ownership
model. Document how the project-level rule relates to active ops namespaces;
do not silently weaken one dump per project into an unrelated UI-only rule.

Distinguish one logical dump from raw capture, curated text and retained
history that may be components of the same artifact. Preserve their ownership
and linkage rather than presenting implementation components as independent
project-wide captures.

The maintainer's explicit singleton requirement in this finding concerns the
project-wide dump. Verify the existing per-use-case lifecycle and resolve any
remaining ambiguity during refinement rather than silently changing its rules.

## Acceptance

1. One New action offers Project-wide and existing use cases in the same form;
   the two old entry buttons are absent.
2. With no general dump, the first Project-wide capture establishes the
   canonical dump. Further captures update that same logical dump and the
   catalog does not gain additional general entries.
3. A general capture initiated through any supported adapter follows the same
   ownership rule. Overlapping requests retain both contributions in one dump.
4. Editing preserves identity and content semantics. Restart and index rebuild
   preserve the same canonical destination.
5. Choosing an existing use case keeps the capture in that case's scope.
6. Existing general material remains available under the explicitly chosen
   reconciliation strategy, with content and references preserved.
7. Record focused service, persistence and adapter validation when this work is
   authorized. This capture runs no implementation or migration checks.

## Implementation Progress

T2.1 is implemented in `7f4b82f629696b39dc2f1a4326f90a11459db1d3`. Project metadata now retains
one validated `project_dump_ref`, with shared authoring exclusion for metadata
saves and source-bound adoption of an unambiguous existing general dump. The
read-only inventory spans namespaces, distinguishes raw and case-scoped records,
retains original artifact bytes, and rejects stale/cross-project observations,
unsafe paths and conflicting destinations. Unknown project frontmatter survives
adoption and later descriptive metadata edits.

The [Slice 2 report](../../report/slices/web-ui-workflow-consolidation/slice-2-project-dump.md)
records successful focused race, complete affected-package lint, signature and
whitespace checks.

The T2.2 core reconciliation owner is implemented in `78b9165e07963d74ab856b68a223034f39e4a609`.
Preview selects by creation timestamp/path with explicit invalid-time evidence.
Apply preserves source body/metadata history, composes attributed content and
publishes aliases, the project reference and a retry receipt through the existing
versioned journal. Recovery and bounded alias/receipt behavior have focused race
coverage. Live data has not been reconciled.

The T2.2 core dump mutation owner is implemented in `7bd9c09ad018e279d04726844e62896b0c365908`.
It journals first and later general captures, retains prepared input and retry
bindings, validates independent named captures, and guards full-body Save/Append
with exact document revisions. Unknown frontmatter/metadata, raw captures and
prior body/metadata history survive edits and recovery. Focused race tests cover
the replacement interruption points, retained preparation, stale/alias/root
bindings, provider ownership validation and affected existing workflows.

Application integration is implemented in `67dfdbd3937618b1003c914784ee952418f895ba`.
The service routes capture/Save/Append through retained core writes, exposes
reconciliation, and reads canonical catalogs/aliases across namespaces. CLI edits
retain operation/revision fields, TUI retains unchanged Save bindings and Web
writes replay receipts before checking current revisions. Confirmed content is
reported with projection-stale evidence when a later index/catalog refresh fails.
The old direct file writers are removed. Focused race checks, affected-package
lint, signature and whitespace checks passed; the Slice 2 report records them.

Bound authoring integration is implemented in `cdd413c861723380a13b56ccc90bc1f7e27d092d`.
General session/Telegram output extends the shared dump across namespaces;
refinement replaces only an observed session output and retains exact history.
Dump, project-reference, session, transcript, provenance and receipt images publish
through one existing recovery journal. Alias-aware observations guard the resolved
body and project metadata, and native work retains its own namespace. Focused race
checks cover all changed publication replacements, retries, stale/provider edits,
legacy adoption and question-only raw input; affected-package lint and signature
checks passed.

Assisted synthesis is implemented in `467d2e083c3fb2d66c6de7bbf2ed8aad33e571e4`.
Its guarded source/session, canonical content, provenance, history, pointer and
receipt share one recoverable publication. Semantic retries preserve original
content after later writes; CLI finalization resumes preparation or its receipt.
Focused race tests, affected-package lint and signature checks passed. T2.2 is
complete. T2.3 is implemented in `2c553693b3fe1efb8f40af7a17f619f88ab188de`. One New
capture modal offers Project-wide and existing case names. Unchanged retries keep
their identity and input; changed submitted arguments create a new intent.
Failures and pending projections retain the form, and confirmed success restores
focus and refreshes the catalog. Deterministic desktop/narrow browser, lost-ack,
HTTP, scope and lint checks passed; the source report records exact commands and
corrected fixture evidence. User Guide, reference and Unreleased are updated.
All Slice 2 tasks are implemented. The report is reviewing in canonical PR
[#399](https://forge.adrianpk.com/adrianpk/chr/pulls/399). Exact clean-HEAD
validation is recorded through Looper before handoff. The controller merged PR #399 into dev at
`aadb6b2bd5060334e0290b441472a93bf952ea92` on 2026-10-18. All four
ordered exact-HEAD checks passed through Looper; the ticket is solved.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).
