# Slice 3: Dash Live Work Stream

Status: delivered
Slicing: Agent Work Observability
Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`
Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`
Branch: `feat/slice-3-dash-live-work-stream`
PR: #310

## Purpose

Replace the static foreground work overlay with a concise, human-readable event
stream while keeping the existing modal safety behavior.

## Delivered Behavior

The TUI observability read model now loads recent agent events from the latest
loop stream and latest execution stream. Events are deduplicated, sorted by
timestamp, capped to the most recent entries, and projected into terminal-ready
fields with relative runtime and validation log paths.

While a foreground work action is running, the overlay now shows:

- the current work action status,
- a navigation lock notice,
- recent agent events,
- runtime or validation log paths when recent events include them.

Dash agent-work refreshes continue while `workBusy` is active, so newly written
events can appear during long-running foreground work. Keyboard and mouse
navigation remain locked while the work action is running, including footer
clicks such as `F10`.

## Implementation Notes

`loadTUIObservability` now reads event streams from existing runtime artifact
locations and keeps event read errors best-effort. Missing or unreadable event
streams do not block TUI startup or refresh.

The work overlay renders the same modal surface as before, but the body is now
event-driven instead of a static wait message. It intentionally shows concise
lifecycle events, not raw Codex or container output.

## Contracts Added or Changed

- `tui.ObservabilityView.RecentEvents`
- `tui.AgentEventView`
- TUI work-running overlay renders recent agent events and log refs.
- TUI mouse footer actions are ignored while `workBusy` is active.

## Files of Interest

- `internal/app/app.go`
- `internal/tui/observability.go`
- `internal/tui/work_actions.go`
- `internal/tui/model.go`
- `internal/tui/model_test.go`

## Validation

- `go test ./internal/tui` (pass)
- `go test ./internal/app` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Detailed runtime/container log inspection remains a later slice.
- Event readback after interrupted runs remains a later slice.
