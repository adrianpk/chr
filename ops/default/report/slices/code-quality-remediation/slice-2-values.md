# Slice 2: Shared boundary values and residual core helpers

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-2-quality-values`
PR: `#439`

## Purpose

Replace the assigned positional contracts with cohesive values and establish
shared ownership before later capability moves.

## Delivered Behavior

Configuration parsing, migration identities/SQL sections, named artifact paths,
context sources and proving operations return named values. Stage sealing and
workspace creation accept cohesive operation inputs. Retained-policy, guided
admission, report, cache and profile lookups preserve explicit found, hit and
included outcomes alongside their original error behavior.

All 15 distinct assigned signature violations are fixed. Eight legal
five-parameter watch items have explicit cohesion dispositions. The exact
temporary debt manifest shrinks from 532 to 517 entries; the seven-parameter
phase and external-interface exception stay unchanged.

Proving objectives, sandbox execution, validation and finite retry orchestration
have separate files. Hashing, project paths, atomic project writes, frontmatter
parsing and exit-code interpretation have explicit shared owners. The
declaration/file ledger assigns all 678 remaining root production files.

Proving sandbox log-close failures are now observable alongside command errors.
A successful command result remains available when cleanup fails. This fixes
an owned write-resource suppression without changing read-only cleanup policy.

## Implementation Notes

Concrete callers change with their inputs/results. Existing authoring helpers
retain their outward contracts until their assigned slices. Normal absence,
cache miss/staleness, inclusion filtering, pending guided admission, config
revision matching and partial proving attempts keep their original branches.
Contexts stay explicit. No package, generic options object, forwarding facade
or migration interface is introduced.

The proving split preserves all 49 moved declaration bodies/types exactly;
only function-header layout changes in that mechanical commit. Cleanup error
handling is a separate task commit. Later package moves must keep durable
transactions at the store and stage policy behind consumer-owned operations.

## Contracts Added or Changed

Named result values replace the assigned tuples while preserving errors.Is
identity, config parsing semantics and original on-disk/wire representations.
The mutable proving log joins its cleanup and command errors with errors.Join.
The final responsibility ledger distinguishes signature tasks, package moves,
shared primitives and store/stage splits.

## Files of Interest

- `internal/core/config*.go`, `migration.go`, `ops_named_artifact.go`:
  parser and artifact values with adapted callers.
- `internal/core/workspace*.go`, `guided_control_admission.go`,
  `guided_use_case_progress.go`, `report_listing.go`: input and lookup values.
- `internal/core/proving_*.go`: objective, sandbox, validation and retry owners.
- `internal/core/content_hash.go`, `project_paths.go`, `project_files.go`,
  `frontmatter.go`, `process_result.go`: unchanged shared primitives.
- `internal/core/proving_cleanup_test.go`: reproduced command/cleanup failures.
- [All assigned dispositions](../../../plan/code-quality-remediation-slice-2-dispositions.csv).
- [Root file/declaration owners](../../../plan/code-quality-remediation-core-owners.csv).
- [Shared dependency decisions](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./internal/core -run '^Test(WorkflowConfig|LoadWorkspaceConfig|Init|Proving|WorkspaceContext|WorkspaceGraphBuilder|ReportLister|GuidedUseCase|ReadSQLiteMigrations|Migrator)' -timeout=3m`
  — passed for configuration, real SQLite/filesystem, cache, report and guided boundaries.
- `go test ./internal/core -run '^Test(WorkflowConfig|Workspace|GuidedUseCase|ProjectMetadata|Proving|RunProving|PlanProving|ParseProving)' -timeout=3m`
  — passed after shared primitive extraction.
- `go test ./internal/core -run '^Test(ParseProving|LoadProving|RunProving|PlanProving|RenderProving|Proving|ProjectMetadata|Workspace|WorkflowConfig)' -timeout=3m`
  — passed after the mechanical proving split.
- `go test ./internal/core -run '^Test(ProvingLogFailure|RunProving|PlanProving|ParseProving|RenderProving)' -timeout=3m`
  — passed; the new cleanup regression failed before the correction.
- `go run ./scripts/goquality` — passed with 517 exact debt entries and one exception.
- Focused strict lint passed with zero issues for the changed core package.

Final clean-head validation runs through the controller after the PR-reference
commit: make verify, the PR's explicit focused core regression command without
short mode, then git diff --check. The controller retains the exact HEAD,
ordered commands, outcomes and timings. That record and PR Validation own final
candidate results. The original review and detections remain unchanged.

## Risks and Follow-ups

The root-owner ledger is a concrete handoff for the already approved later
slices, not evidence that their package moves are complete. Mixed store/stage
files still require their recorded transaction and recovery checks before a
move. No Slice 2 signature debt remains.
