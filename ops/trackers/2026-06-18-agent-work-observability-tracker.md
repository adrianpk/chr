# Agent Work Observability Tracker

Status: planned

Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`

Spec: `ops/specs/2026-06-18-agent-work-observability.md`

Model: `ops/specs/2026-06-18-agent-work-observability-model.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | agent event model | delivered | `feat/slice-1-agent-event-model` | `#308` | `.prj/reports/slices/agent-work-observability/slice-1-agent-event-model.md` |
| Slice 2 | work lifecycle events | delivered | `feat/slice-2-work-lifecycle-events` | `#309` | `.prj/reports/slices/agent-work-observability/slice-2-work-lifecycle-events.md` |
| Slice 3 | dash live work stream | planned | `feat/slice-3-dash-live-work-stream` | pending | `.prj/reports/slices/agent-work-observability/slice-3-dash-live-work-stream.md` |
| Slice 4 | execution log lookup | planned | `feat/slice-4-execution-log-lookup` | pending | `.prj/reports/slices/agent-work-observability/slice-4-execution-log-lookup.md` |
| Slice 5 | event readback recovery | planned | `feat/slice-5-event-readback-recovery` | pending | `.prj/reports/slices/agent-work-observability/slice-5-event-readback-recovery.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add the `AgentEvent` model and validation helpers. | done | `689ac63` | Core model only. |
| T1.2 | Add JSONL append/read primitives for runtime event streams. | done | `b9c6239` | Append-only file storage. |
| T1.3 | Add unit tests for persistence, ordering, and malformed-line handling. | done | `3ff06cf` | Reader tolerates partial corruption. |
| T1.4 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Wire an event sink into work loop and ticket execution paths. | done | `7078ff2` | Preserve existing execution semantics. |
| T2.2 | Emit loop and ticket lifecycle events at high-signal boundaries. | done | `6ae5ed5` | No raw Codex stream in primary event feed. |
| T2.3 | Cover lifecycle emission with service or CLI tests. | done | `6d1b194` | Includes run-all path. |
| T2.4 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add a TUI read model for recent agent events. | planned | pending | Reads durable event streams. |
| T3.2 | Render the work-running screen as a live event stream with lock notice. | planned | pending | Replace static wait text. |
| T3.3 | Refresh event state while foreground work is active. | planned | pending | Keep navigation locked. |
| T3.4 | Add TUI tests for modal lock behavior and event rendering. | planned | pending | Avoid regressions around `workBusy`. |
| T3.5 | Add the slice implementation report and validation evidence. | planned | pending | Report path defined in slice map. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add a log lookup command or extend an existing report/runtime command. | planned | pending | CLI-first is acceptable. |
| T4.2 | Connect event refs to lookup output or hint text. | planned | pending | Avoid hardcoding container ids when unavailable. |
| T4.3 | Add command tests for successful and missing-log lookup. | planned | pending | Include missing execution path. |
| T4.4 | Add the slice implementation report and validation evidence. | planned | pending | Report path defined in slice map. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add event stream summary/readback behavior. | planned | pending | Latest known loop/ticket/execution state. |
| T5.2 | Surface latest known work state in CLI or Dash diagnostics. | planned | pending | Degraded summaries for partial streams. |
| T5.3 | Document recovery inspection for interrupted runs. | planned | pending | User-facing how-to or reference update. |
| T5.4 | Add the slice implementation report and validation evidence. | planned | pending | Report path defined in slice map. |

## Validation Log

- `go test ./internal/core` (pass)
- `make verify` (pass)
- `go test ./internal/app` (pass)
- `make verify` (pass)

## Notes

- Foreground work remains modal during this slicing.
- The primary UI surface should show concise lifecycle events, not raw runtime
  logs.
- Detailed runtime/container logs can remain CLI-first until the event stream
  contract is in place.
