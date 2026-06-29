# Slice 1: Execution Record Model

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-1-execution-record-model`
PR: `#34`

## Purpose

Define the durable domain model for one ticket execution attempt.

## Delivered Behavior

The core package now has execution record types, status values, runtime
references, artifact references, status validation, record validation, and a
helper for copying runtime plan references into an execution record.

## Implementation Notes

The model is storage-neutral. Persistence, migrations, CLI creation, and
execution mutation belong to later slices in this slicing.

## Contracts Added or Changed

New execution statuses:

- `pending`
- `running`
- `succeeded`
- `failed`
- `blocked`

New core types and functions:

- `ExecutionStatus`
- `ExecutionRecord`
- `ExecutionRuntimeRef`
- `ExecutionArtifacts`
- `ExecutionRecordFromRuntimePlan`
- `ValidateExecutionRecord`

Required record fields:

- execution ID;
- project ID;
- ticket key;
- status;
- created timestamp;
- updated timestamp.

## Files of Interest

- `internal/chr/execution.go`
- `internal/chr/execution_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

The execution record is not persisted yet. SQLite storage and repository APIs
are planned for Slice 2.
