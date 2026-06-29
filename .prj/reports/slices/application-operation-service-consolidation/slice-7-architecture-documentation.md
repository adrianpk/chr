# Slice 7: Architecture Documentation

Status: delivered
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `docs/slice-7-architecture-documentation`
PR: #162

## Purpose

Document the application operation service boundary and adapter
responsibilities after the consolidation slices.

The guidance keeps future CLI, TUI, Telegram, email, and other adapter work
centered on the shared operation service instead of duplicating reusable
workflow logic.

## Delivered Behavior

- Documented `internal/chrapp` as the owner of the application operation
  service boundary.
- Documented that adapters translate inputs, handle interaction or transport
  concerns, and map operation results.
- Documented that adapters should not own reusable workflow logic.
- Documented that one central operation service may be split across capability
  files without becoming separate entity services.
- Documented the adapter-neutral `service*.go` boundary and TUI adapter wrapper
  placement.
- Added convention-level anti-patterns for adapter-owned workflow logic and
  unnecessary per-entity service implementations.

## Implementation Notes

The documentation follows the architecture implemented by the previous slices:
one central `internal/chrapp` service type, split across operation files for
readability, with TUI-specific mappings isolated in adapter files.

The guidance keeps narrow interfaces available for adapter contracts while
making the central service or a thin wrapper the expected implementation path.

## Contracts Added or Changed

- Added workspace architecture guidance for the application operation service.
- Added coding convention guidance for keeping reusable operations out of
  adapters.
- Added explicit guidance against splitting the operation layer into separate
  Project, Ticket, Authoring, or similar service implementations by default.

## Files of Interest

- `.agents/ARCHITECTURE.md`
- `.agents/CONVENTIONS.md`
- `.prj/reports/slices/application-operation-service-consolidation/slice-7-architecture-documentation.md`
- `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`

## Validation

- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

None recorded.
