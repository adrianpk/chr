# Slice 4: Runtime Workspace Mount

Status: delivered
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-4-runtime-workspace-mount`
PR: #331

## Purpose

Make the runtime harness mount the execution workspace as the writable
container workspace when worktree mode is active.

## Delivered Behavior

- `RuntimePlan` now records ticket branch and workspace base metadata alongside
  project checkout and execution workspace paths.
- Project mode still treats the project checkout as the execution workspace and
  keeps the writable `/workspace` mount source unchanged.
- Worktree mode uses `ExecutionWorkspaceDir` as the writable `/workspace` mount
  source for worker and validation containers.
- Runtime artifacts are resolved under the execution workspace in worktree
  mode, so `context.md`, `result.json`, `runtime.log`, and `validation.log`
  are visible inside the mounted container workspace.
- Container path rendering now derives paths from the host path mounted at the
  container workdir.
- Codex worker context now includes workspace mode, project checkout path,
  execution workspace path, ticket branch, workspace base ref, and base SHA.
- Runtime validation output now shows workspace mode and execution workspace
  paths, and validates the execution workspace directory.

## Implementation Notes

- Docker build still uses the project checkout and configured runtime
  Dockerfile.
- This slice does not prepare ticket worktrees. Later branch lifecycle work will
  pass the prepared ticket worktree path into the runtime plan.
- Execution persistence still stores the existing runtime/artifact fields. The
  durable execution metadata expansion remains owned by Slice 5.

## Contracts Added or Changed

- `RuntimePlanOptions` accepts `TicketBranch`, `WorkspaceBaseRef`, and
  `WorkspaceBaseSHA`.
- `RuntimePlan` carries `TicketKey`, `TicketBranch`, `WorkspaceBaseRef`, and
  `WorkspaceBaseSHA`.
- `RuntimePlan.ContainerPathHostRoot()` identifies the host root mounted at the
  container workdir.
- `ValidateRuntime` checks `execution_workspace`.
- `printRuntimeValidateResult` prints `workspace_mode`,
  `project_checkout_dir`, and `execution_workspace_dir`.

## Files of Interest

- `internal/core/runtime_plan.go`
- `internal/core/harness_invocation.go`
- `internal/core/codex_worker_context.go`
- `internal/core/context_materializer.go`
- `internal/core/runtime_validate.go`
- `internal/app/runtime_proving.go`
- `internal/core/runtime_plan_test.go`
- `internal/core/harness_invocation_test.go`
- `internal/core/codex_worker_context_test.go`
- `internal/core/context_materializer_test.go`
- `internal/core/runtime_validate_test.go`
- `internal/app/cli_test.go`

## Validation

- `go test ./internal/core -run 'TestPlanRuntime|TestNewHarnessInvocation|TestMaterializeCodexWorkerContextForExecution|TestRenderMainContextPackage|TestValidateRuntime' && go test ./internal/app -run 'TestRuntimeValidatePrintsRuntimePlan'` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Worktree execution depends on later slices preparing and passing a real ticket
  worktree path before runtime materialization.
- Execution records and reports do not yet persist workspace mode or execution
  workspace metadata. Slice 5 owns that expansion.
- Branch creation, host-side push, and PR lifecycle are still owned by Slice 6.
