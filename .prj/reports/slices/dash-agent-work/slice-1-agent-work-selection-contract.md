# Slice 1: Agent Work Selection Contract

Status: reviewing

## Purpose

Define the shared default ticket selection contract for agent work before
adding new CLI and Dash surfaces.

## Delivered

- Changed the default loop eligible status to `ready_for_dev`.
- Preserved explicit legacy `ready` selection through the existing status
  validation path.
- Changed the default loop order to `created_at`, oldest first, with existing
  deterministic tie-breakers.
- Updated the TUI ready-loop service path to use the same `ready_for_dev` and
  `created_at` contract.
- Updated CLI and core tests for default `ready_for_dev` selection and
  oldest-first loop behavior.
- Updated ticket loop and related command reference examples.

## Validation

- `go test ./internal/chr ./internal/chrapp`
- `make verify`
- `git diff --check`
