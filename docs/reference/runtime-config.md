# Runtime Configuration

Runtime configuration may live in the managed project's `.chr/default/config.yml` file
under the `runtime:` key.

When fields are omitted, chr uses local defaults suitable for the
project-owned runtime scaffold.

New workspaces omit runtime defaults from generated config so users only edit
runtime values when they are making a real project decision.

## Fields

- `image`: Docker image tag used for the agent runtime.
- `dockerfile`: project-relative path to the agent runtime Dockerfile.
- `workdir`: container working directory.
- `artifact_dir`: project-relative directory for runtime results.
- `validation_command`: command executed inside the runtime for validation.

## Default Values

- `image`: `chr-agent-runtime:local`
- `dockerfile`: `.chr/default/runtime/agent/Dockerfile`
- `workdir`: `/workspace`
- `artifact_dir`: `ops/default/run`
- `validation_command`: `make verify`

Example override:

```yml
runtime:
  image: chr-agent-runtime:local
  dockerfile: .chr/default/runtime/agent/Dockerfile
  validation_command: make verify
```

## Validation

Preview the runtime boundary without invoking Docker:

```sh
chr runtime validate --dry-run
```

Validate the configured runtime:

```sh
chr runtime validate
```

Runtime validation checks the workspace, Dockerfile path, artifact directory,
image tag, container workdir, and planned mount. The dry-run path prints the
plan without building an image, starting a container, running the harness, or
executing the validation command.

The runtime plan also records the effective workspace execution metadata from
`work:`:

- `workspace_mode`: `worktree` with `workflow.version: 1`; otherwise the legacy
  default is `project`. Explicit project configuration overrides the default.
- `project_checkout`: the user-visible project checkout.
- `execution_workspace`: the path intended for agent work. In `project` mode
  this is the project checkout. In `worktree` mode this is the prepared ticket
  worktree.

New projects initialized with `workflow.version: 1` use isolated worktrees.
Existing unversioned projects keep project-mounted execution unless they
explicitly opt into `work.workspace_mode: worktree` or deliberately adopt the
versioned policy. In worktree mode, ticket execution
prepares the isolated Git worktree first and the runtime mount points Docker at
that execution workspace.

Human output uses project-relative paths for files inside the managed project
and groups path checks under `checks:`.

## Runtime Artifacts

`chr ticket start` creates a shared operational artifact envelope in the
project checkout. It does not use `runtime.artifact_dir` for ticket execution
outputs:

```text
ops/<namespace>/run/<artifact-id>/
  artifact.md
  metadata.json
  files/context.md
  evidence/result.json
  evidence/runtime.log
  evidence/validation.log
  evidence/events.jsonl
  evidence/execution-report.md
```

`<namespace>` is `default` when `OPS_USER` is unset. Set `OPS_USER` to a
lowercase, hyphen-separated namespace to isolate one user's committed
operational artifacts without changing the rest of the runtime behavior.

`runtime.artifact_dir` defaults to `ops/default/run` for runtime routines that
have not yet been allocated a shared operational artifact envelope. Older
`.chr/default/runtime/runs` configuration is treated as a migration input and
planned through the active `ops/<namespace>/run` root.

For the detailed execution contract, see
[Ticket execution](ticket-execution.md).
