# chr Operational Artifacts to Ops Plan

Status: planned
Ticket: `ops/tickets/open/20260714151004-migrate-chr-operational-artifacts-to-ops.md`
Tracker: `ops/trackers/2026-07-14-chr-operational-artifacts-to-ops-tracker.md`

## Purpose

Complete the chr operational artifact migration by moving active artifact
readers/writers from the intermediate `.chr/default/...` pseudo-user tree to the
canonical `ops/<namespace>/...` layout, while preserving `.chr/` for CHR-private
configuration and local state.

## Boundary

### Keep in `.chr/`

- CHR-private configuration such as `.chr/default/config.yml` and
  `.chr/default/config/gates.yml`.
- CHR-private agent protocol/profile assets when they are configuration or
  local tool state.
- Local caches or implementation-private state that is not a shared operational
  artifact.

### Move to `ops/<namespace>/...`

- Dumps, quizzes, specs, plans, tickets, reports, evidence examples, acceptance
  artifacts, runtime run evidence, worktrees, generated operational envelopes,
  and comparable CHR-managed operational outputs.
- No-user workspaces use `ops/default`.

## Slice Map

### Slice 1: Ops artifact boundary

- Branch: `fix/slice-1-ops-artifact-boundary`
- PR title: `feat(slice-1): ops artifact boundary`
- Report: `.prj/reports/slices/chr-operational-artifacts-to-ops/slice-1-ops-artifact-boundary.md`
- Prerequisites: PRs `#358`, `#359`, and release PR `#360` delivered the
  intermediate migration and lint gate cleanup.

## Slice 1 Tasks

- T1.1 Define code-level path helpers for chr private state versus operational
  artifacts.
- T1.2 Retarget active operational artifact producers/readers to
  `ops/<namespace>/<type>/...` where existing behavior expects filesystem-backed
  Markdown artifacts.
- T1.3 Preserve `.chr/` configuration paths and update docs/scripts/tests to
  reflect the boundary.
- T1.4 Validate with focused legacy-path scans, `make verify`, ops layout
  validation, and diff hygiene.

## Out of Scope

- Removing `.chr/` entirely.
- Migrating unrelated external `.chc` sample projects outside this repository.
- Changing the shared ops layout contract.
