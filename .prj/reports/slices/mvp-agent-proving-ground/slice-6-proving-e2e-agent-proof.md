# Slice 6: End-to-End Agent Proof

Status: delivered
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-6-proving-e2e-agent-proof`
PR: #77

## Purpose

Demonstrate that `chr` can run a sandboxed Codex agent against the controlled
todolist fixture, validate the result, and leave the project working.

## Delivered Behavior

The proving runtime now supports the real local agent execution path:

- Codex credentials are mounted read-only into the container when present.
- `OPENAI_API_KEY` is inherited by name into `docker run`.
- Codex runs without its inner bubblewrap sandbox because Docker is the sandbox
  boundary for this runtime.

## Proof Run

The fixture was materialized with:

```sh
./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh
```

The initial TODO-1 validation failed as expected:

- `make compile`: pass
- `make lint`: pass
- `make test`: fail
- `make verify`: fail

The bugfix objective was executed with:

```sh
go run ./cmd/chr proving ralph \
  --dir _workspace/testing/projects/todolist \
  --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md
```

Result:

- status: `passed`
- attempts: `1`
- stop reason: `validation_passed`
- summary: `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/ralph-loop.md`

The feature objective was executed with:

```sh
go run ./cmd/chr proving ralph \
  --dir _workspace/testing/projects/todolist \
  --objective .chr/proving/objectives/TODO-2-complete-task.md
```

Result:

- status: `passed`
- attempts: `1`
- stop reason: `validation_passed`
- summary: `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/ralph-loop.md`

Final fixture validation:

```sh
make verify
```

Result: pass.

## Evidence Artifacts

Artifacts were written under the ignored fixture workspace:

- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/attempt-1/context.md`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/attempt-1/result.json`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/attempt-1/agent.log`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/attempt-1/validation.log`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-1/ralph-loop.md`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/attempt-1/context.md`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/attempt-1/result.json`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/attempt-1/agent.log`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/attempt-1/validation.log`
- `_workspace/testing/projects/todolist/.chr/runtime/runs/proving/TODO-2/ralph-loop.md`

## Limitations

TODO-1 validation runs the full fixture suite. In this proof run, the TODO-1
agent fixed both the task ID bug and `CompleteTask`, so TODO-2 confirmed that
the final feature state was already correct rather than producing a second code
change.

The agent's internal validation saw a limited PATH in one attempt, but `chr`
validation ran outside the agent process and passed. The proof criterion is the
`chr` validation result.

Future clone, commit, and push execution should pass a repository-scoped PAT
through an explicit secret environment variable. This is operationally required
for remote worktree flows, but it was not required for this mounted-workspace
proof.

## Validation

- `go test ./internal/chr -run 'TestNewHarnessInvocationRendersDockerRunCommand|TestNewCodexWorkerCommandDefinesCodexInvocation|TestRenderCodexWorkerShellScriptQuotesInputs|TestCodexCredentialMounts'`: pass
- `go run ./cmd/chr proving run --dir _workspace/testing/projects/todolist --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md --dry-run`: pass
- `go run ./cmd/chr proving ralph --dir _workspace/testing/projects/todolist --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md`: pass
- `go run ./cmd/chr proving ralph --dir _workspace/testing/projects/todolist --objective .chr/proving/objectives/TODO-2-complete-task.md`: pass
- `make verify` in `_workspace/testing/projects/todolist`: pass
- `make verify`: pass
