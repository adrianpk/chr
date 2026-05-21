# Slice 7: Maintenance Documentation

Status: reviewing
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-7-maintenance-docs`
PR: #132

## Purpose

Add task-oriented documentation for inspecting and recovering local `chc` state.

This slice closes the Maintenance and Recovery slicing with user-facing
documentation that connects the maintenance commands delivered in earlier
slices.

## Delivered Behavior

- Added a how-to guide for local state inspection and recovery.
- Linked the guide from the documentation index.
- Added the maintenance reference to the documentation index.
- Added maintenance workflow coverage to the user guide.
- Linked the maintenance reference back to the recovery how-to.

## Implementation Notes

The new how-to lives under `docs/how-to/` because it is procedural. Exact command
behavior remains in `docs/reference/maintenance.md`.

## Files of Interest

- `docs/how-to/inspect-and-recover-local-state.md`
- `docs/how-to/user-guide/index.md`
- `docs/index.md`
- `docs/reference/maintenance.md`

## Validation

```sh
make verify
```

The command passes.

## Risks and Follow-ups

- This slice does not add new runtime behavior.
