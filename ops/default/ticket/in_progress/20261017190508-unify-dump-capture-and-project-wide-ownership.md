---
id: TKT-20261017190508
title: "Unify dump capture around one evolving project-wide dump"
status: in_progress
kind: task
severity: medium
priority: normal
scope: domain
tags: dumps, project-context, authoring
source: chat
reported_at: 2026-10-17T19:05:08Z
commits: [a8bdec69d5595740a6fee1b723b1d2f1357a6ce4, f4b080aeb4a2933b87535d56ca64e9fcb74f39ef, b18c795c115463d37bd676e9968b9ae1e25e47f1, 14735ffbd247419294adcc9ef93c164fe63471bf]
branch: feat/slice-2-project-dump
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Unify dump capture around one evolving project-wide dump

## Context

Finding from [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
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

T2.1 is implemented in `a8bdec69d5595740a6fee1b723b1d2f1357a6ce4`. Project metadata now retains
one validated `project_dump_ref`, with shared authoring exclusion for metadata
saves and source-bound adoption of an unambiguous existing general dump. The
read-only inventory spans namespaces, distinguishes raw and case-scoped records,
retains original artifact bytes, and rejects stale/cross-project observations,
unsafe paths and conflicting destinations. Unknown project frontmatter survives
adoption and later descriptive metadata edits.

The [Slice 2 report](../../report/slices/web-ui-workflow-consolidation/slice-2-project-dump.md)
records successful focused race, complete affected-package lint, signature and
whitespace checks.

The T2.2 core reconciliation owner is implemented in `f4b080aeb4a2933b87535d56ca64e9fcb74f39ef`.
Preview selects by creation timestamp/path with explicit invalid-time evidence.
Apply preserves source body/metadata history, composes attributed content and
publishes aliases, the project reference and a retry receipt through the existing
versioned journal. Recovery and bounded alias/receipt behavior have focused race
coverage. Live data has not been reconciled.

The T2.2 core dump mutation owner is implemented in `b18c795c115463d37bd676e9968b9ae1e25e47f1`.
It journals first and later general captures, retains prepared input and retry
bindings, validates independent named captures, and guards full-body Save/Append
with exact document revisions. Unknown frontmatter/metadata, raw captures and
prior body/metadata history survive edits and recovery. Focused race tests cover
the replacement interruption points, retained preparation, stale/alias/root
bindings, provider ownership validation and affected existing workflows.

Application integration is implemented in `14735ffbd247419294adcc9ef93c164fe63471bf`.
The service routes capture/Save/Append through retained core writes, exposes
reconciliation, and reads canonical catalogs/aliases across namespaces. CLI edits
retain operation/revision fields, TUI retains unchanged Save bindings and Web
writes replay receipts before checking current revisions. Confirmed content is
reported with projection-stale evidence when a later index/catalog refresh fails.
The old direct file writers are removed. Focused race checks, affected-package
lint, signature and whitespace checks passed; the Slice 2 report records them.

Bound authoring integration is implemented in `ac1e17b9257ac2ee53e821b7c87876fdf5d1c4c6`.
General session/Telegram output extends the shared dump across namespaces;
refinement replaces only an observed session output and retains exact history.
Dump, project-reference, session, transcript, provenance and receipt images publish
through one existing recovery journal. Alias-aware observations guard the resolved
body and project metadata, and native work retains its own namespace. Focused race
checks cover all changed publication replacements, retries, stale/provider edits,
legacy adoption and question-only raw input; affected-package lint and signature
checks passed.

T2.2 still requires assisted synthesis routing. T2.3's single Web New capture form
with existing case choices and retained capture identity remains. Final controller
validation has not run and the slice has no ready PR.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).
