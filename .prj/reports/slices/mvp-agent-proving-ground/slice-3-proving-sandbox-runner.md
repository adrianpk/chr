# Slice 3: Sandbox Codex Runner

Status: delivered
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-3-proving-sandbox-runner`
PR: #74

## Purpose

Run one direct proving objective inside the fixture's agent sandbox.

## Delivered Behavior

`chr` now exposes a temporary proving command:

```sh
chr proving run --dir <project> --objective <objective> [--dry-run]
```

The command loads a direct proving objective, computes the fixture runtime plan,
renders Docker build/run commands, and materializes a worker context for Codex.

In non-dry-run mode it runs:

- Docker build for the fixture agent image;
- Docker run with the project mounted writable;
- `codex exec` inside the sandbox container.

Validation and Ralph Loop retries remain later slices.

## Implementation Notes

The runner reuses the existing runtime plan, harness invocation, Codex worker
command, and OS harness executor boundaries.

The new core proving package parses the fixture objective frontmatter and
resolves the related fixture ticket path.

Each proving attempt is an independent `codex exec` invocation. Context between
future Ralph Loop attempts must be materialized explicitly through artifacts,
not through hidden conversational memory.

## Contracts Added or Changed

New temporary CLI command:

- `chr proving run`
- `--dir`
- `--objective`
- `--dry-run`

New core behavior:

- `LoadProvingObjective`
- `ParseProvingObjective`
- `PlanProvingSandboxRun`
- `RunProvingSandbox`
- `RenderProvingWorkerContext`

New reference doc:

- `docs/reference/mvp-agent-proving-ground.md`

## Files of Interest

- `cmd/chr/main.go`
- `internal/chr/proving.go`
- `internal/chr/proving_test.go`
- `docs/reference/mvp-agent-proving-ground.md`

## Validation

- `go test ./internal/chr -run 'TestParseProving|TestPlanProving|TestRunProving|TestRenderProving'`: pass
- `./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh`: pass
- `go run ./cmd/chr proving run --dir _workspace/testing/projects/todolist --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md --dry-run`: pass
- `go run ./cmd/chr proving run --dir _workspace/testing/projects/todolist --objective .chr/proving/objectives/TODO-2-complete-task.md --dry-run`: pass
- `docker --version`: pass
- `make verify`: pass

## Risks and Follow-ups

Non-dry-run execution depends on local Docker and Codex credentials inside the
fixture runtime image. Slice 4 adds validation capture, and Slice 5 adds Ralph
Loop retries.
