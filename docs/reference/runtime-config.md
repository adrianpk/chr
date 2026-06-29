# Runtime Configuration

Runtime configuration may live in the managed project's `.chr/user/config.yml` file
under the `runtime:` key.

When fields are omitted, Chryso uses local defaults suitable for the
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
- `dockerfile`: `.chr/user/runtime/agent/Dockerfile`
- `workdir`: `/workspace`
- `artifact_dir`: `.chr/user/runtime/runs`
- `validation_command`: `make verify`

Example override:

```yml
runtime:
  image: chr-agent-runtime:local
  dockerfile: .chr/user/runtime/agent/Dockerfile
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

- `workspace_mode`: `project` by default.
- `project_checkout`: the user-visible project checkout.
- `execution_workspace`: the path intended for agent work. In `project` mode
  this is the project checkout. In `worktree` mode this is the prepared ticket
  worktree.

Existing projects keep project-mounted execution unless they explicitly opt
into `work.workspace_mode: worktree`. In worktree mode, ticket execution
prepares the isolated Git worktree first and the runtime mount points Docker at
that execution workspace.

Human output uses project-relative paths for files inside the managed project
and groups path checks under `checks:`.

## Runtime Artifacts

Ticket execution writes artifacts under the configured runtime artifact
directory. Execution-scoped paths include the worker context, result envelope,
runtime log, and validation log.

For the detailed execution contract, see
[Ticket execution](ticket-execution.md).
