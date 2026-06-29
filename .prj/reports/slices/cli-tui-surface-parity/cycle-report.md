# CLI TUI Surface Parity Cycle Report

Status: completed
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`

## Purpose

Summarize the completed five-slice parity cycle that aligned Chryso CLI
and TUI operation surfaces.

## Delivered Slices

| Slice | PR | Result |
| --- | --- | --- |
| Slice 1: Inventory | #204 | Created the CLI/TUI action matrix. |
| Slice 2: Service Boundaries | #205 | Aligned audited shared service paths. |
| Slice 3: Missing CLI Actions | #206 | Added service-backed CLI commands for TUI operations. |
| Slice 4: Parity Tests | #207 | Added CLI and TUI adapter parity tests. |
| Slice 5: Parity Docs | #208 | Published user-facing and reference parity documentation. |

## Current Contract

- CLI and TUI are first-class operation surfaces.
- Shared actions route through application services.
- The TUI remains responsible for interactive navigation, editing, focus,
  confirmation, and busy states.
- The CLI remains responsible for scripted and agent-operable workflows.
- The parity matrix records matched, missing, intentionally interactive, and
  not-yet-implemented actions.

## Validation

- `make verify`: pass.

## Remaining Gaps

- Workspace graph refresh remains only partially covered by CLI index commands.
- TUI checklist item operations remain unavailable until a safe checklist writer
  exists.

## Tuning Decision

Post-slice tuning needed: no.
