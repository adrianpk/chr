# Runtime Configuration

Runtime configuration lives in the managed project's `.chc/config.yml` file
under the `runtime:` key.

When fields are omitted, ChrysoCoord uses local defaults suitable for the
project-owned runtime scaffold.

## Fields

- `image`: Docker image tag used for the agent runtime.
- `dockerfile`: project-relative path to the agent runtime Dockerfile.
- `workdir`: container working directory.
- `artifact_dir`: project-relative directory for runtime results.
- `validation_command`: command executed inside the runtime for validation.

## Default Configuration

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
```

## Validation

Preview the runtime boundary without invoking Docker:

```sh
chc runtime validate --dry-run
```

Validate the configured runtime:

```sh
chc runtime validate
```

Runtime validation checks the workspace, Dockerfile path, artifact directory,
image tag, container workdir, and planned mount. The dry-run path prints the
plan without building an image, starting a container, running the harness, or
executing the validation command.

## Runtime Artifacts

Ticket execution writes artifacts under the configured runtime artifact
directory. Execution-scoped paths include the worker context, result envelope,
runtime log, and validation log.

For the detailed execution contract, see
[Ticket execution](ticket-execution.md).
