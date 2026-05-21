# Maintenance and Recovery Slicing Plan

Status: delivered

## Purpose

Plan the `Maintenance and Recovery` slicing from the feature slicing set map.

This slicing covers the use case:

> A user can inspect and repair local `chc` state when files, database state,
> or runtime artifacts drift.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related delivered work:

- `docs/reference/project-registry.md`
- `docs/reference/project-workspace.md`
- `docs/reference/ticket-index-rebuild.md`
- `docs/reference/ticket-index-status.md`
- `docs/reference/ticket-reconciliation.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/reports.md`

## Current State

Delivered maintenance capabilities:

- project resolution validates missing, disabled, and non-managed registered
  project paths;
- `chc index rebuild` can reconstruct the derived ticket index from Markdown;
- `chc index status` can report missing, stale, orphaned, and invalid ticket
  index entries;
- runtime validation can inspect the configured Docker boundary;
- execution records, loop runs, reports, and runtime artifacts preserve
  evidence;
- `chc maintenance db status` reports migration status;
- `chc maintenance registry status`, `disable`, and `set-path` expose registry
  diagnostics and explicit repair;
- `chc maintenance execution artifacts` detects execution artifact drift;
- `chc maintenance execution cleanup-preview` previews unreferenced runtime
  artifact directories;
- `chc maintenance bundle` writes a local diagnostic bundle;
- maintenance workflows are documented under `docs/how-to/` and
  `docs/reference/`.

## Slicing Scope

This slicing includes:

- local maintenance command surfaces;
- read-only diagnostics before destructive repair;
- explicit dry-run or preview behavior for cleanup;
- documentation that explains recovery paths and safety boundaries.

## Out of Scope

This slicing does not include:

- remote repair over Telegram or email;
- automatic destructive cleanup without user opt-in;
- changing ticket execution, loop, or worker contracts;
- hosted telemetry or cloud diagnostics;
- multi-user repair coordination.

## Design Direction

Maintenance commands should be conservative:

- prefer status and diagnostics before mutation;
- report exact files, records, and paths involved;
- avoid deleting user-authored Markdown;
- keep derived-state repair explicit;
- use dry-run output for cleanup and bundle creation where practical.

Commands should make local drift understandable without requiring direct SQLite
inspection.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Database Migration Status

Status: delivered.

Expected branch: `feat/slice-1-db-migration-status`
Expected PR title: `feat(slice-1): db migration status`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-1-db-migration-status.md`

Goal:

Expose local SQLite migration state through the CLI.

Candidate scope:

- add migration status model derived from migration assets and
  `schema_migrations`;
- report applied, pending, and unknown database migrations;
- add `chc maintenance db status`;
- add tests for clean, pending, and unknown migration states;
- document the command.

Expected outcome:

- users can inspect schema state without opening SQLite manually.

### Slice 2: Project Registry Diagnostics

Status: delivered.

Expected branch: `feat/slice-2-project-registry-diagnostics`
Expected PR title: `feat(slice-2): project registry diagnostics`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-2-project-registry-diagnostics.md`

Goal:

Expose registry drift across registered projects.

Candidate scope:

- list registered projects with path existence, `.chc/` presence, enabled
  state, and active marker;
- report actionable diagnostics for missing or unmanaged paths;
- keep the command read-only;
- document registry diagnostic behavior.

Expected outcome:

- users can see which registered projects need attention before commands fail.

### Slice 3: Project Registry Repair

Status: delivered.

Expected branch: `feat/slice-3-project-registry-repair`
Expected PR title: `feat(slice-3): project registry repair`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-3-project-registry-repair.md`

Goal:

Add explicit repair operations for stale registry entries.

Candidate scope:

- support disabling stale project registrations;
- support updating a registered project path;
- require explicit project names;
- add dry-run output for repair actions;
- document repair boundaries.

Expected outcome:

- users can recover moved or deleted projects without editing the database.

### Slice 4: Execution Artifact Drift

Status: delivered.

Expected branch: `feat/slice-4-execution-artifact-drift`
Expected PR title: `feat(slice-4): execution artifact drift`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-4-execution-artifact-drift.md`

Goal:

Detect drift between execution records and runtime artifacts.

Candidate scope:

- inspect execution records for one project;
- report missing run directories, result artifacts, logs, validation artifacts,
  and reports;
- keep the command read-only;
- add focused tests with temporary databases and workspaces.

Expected outcome:

- users can identify incomplete or damaged execution evidence.

### Slice 5: Runtime Artifact Cleanup Preview

Status: delivered.

Expected branch: `feat/slice-5-runtime-artifact-cleanup-preview`
Expected PR title: `feat(slice-5): runtime artifact cleanup preview`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-5-runtime-artifact-cleanup-preview.md`

Goal:

Preview cleanup of stale runtime artifacts.

Candidate scope:

- identify runtime artifact directories not referenced by current execution
  records;
- print planned cleanup actions without deleting by default;
- keep user-authored Markdown outside cleanup scope;
- document cleanup safety boundaries.

Expected outcome:

- users can understand cleanup impact before any destructive command exists.

### Slice 6: Diagnostic Bundle

Status: delivered.

Expected branch: `feat/slice-6-diagnostic-bundle`
Expected PR title: `feat(slice-6): diagnostic bundle`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-6-diagnostic-bundle.md`

Goal:

Create a local diagnostic bundle for handoff.

Candidate scope:

- gather command-readable state summaries;
- include selected reports and maintenance diagnostics;
- avoid copying secrets or credentials;
- write a deterministic Markdown bundle under `.chc/maintenance/`;
- document bundle contents.

Expected outcome:

- users can hand off local state without manually collecting scattered files.

### Slice 7: Maintenance Documentation

Status: delivered.

Expected branch: `feat/slice-7-maintenance-docs`
Expected PR title: `feat(slice-7): maintenance docs`
Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-7-maintenance-docs.md`

Goal:

Document maintenance workflows and recovery boundaries.

Candidate scope:

- add a how-to guide for inspecting and recovering local state;
- update command reference;
- link existing index, runtime, report, and registry references;
- add slice implementation report.

Expected outcome:

- maintenance behavior is discoverable without reading implementation code.

## Completion Criteria

The slicing is complete when:

- database migration state can be inspected;
- project registry drift can be diagnosed and repaired explicitly;
- execution artifact drift can be detected;
- runtime artifact cleanup can be previewed safely;
- diagnostic bundles can be generated locally;
- documentation explains maintenance workflows and recovery boundaries;
- `make verify` passes.

## Validation

Required validation for every implementation slice:

```sh
make verify
```

Focused tests should be added for each new model, command, repository, and
filesystem boundary.
