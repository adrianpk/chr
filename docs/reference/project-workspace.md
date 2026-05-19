# Project Workspace

## Root

ChrysoCoord uses `.chc/` as the project-local hidden directory.

The `chc` command creates and manages this directory inside the project where it runs. The directory is intentionally tied to the CLI name rather than the target repository name.

Project ticket files and project protocol assets are local to `.chc/` inside each managed project.

ChrysoCoord may also keep user-level state outside the project for daemon operation, project registration, and derived indexes. User-level state does not replace the project-local `.chc/` source files.

Managed projects also receive a root-level `AGENTS.md` for compatibility with agent tooling that looks for this file at the repository root.

## Initialization

Create the workspace in the current directory:

```sh
chc init
```

`chc init` uses the `standard` profile by default.

Create the workspace with a specific profile:

```sh
chc init --profile minimal
```

Supported initial profiles:

- `standard`: default useful profile with agent protocol files.
- `personal`: full opinionated workflow for the author's own use.
- `minimal`: lightweight scaffold.

Create the workspace in another project directory:

```sh
chc init --dir /path/to/project
```

`--dir` is an override for scripts or external initialization. The normal workflow is to change into the project and run `chc init`.

Preview initialization without writing files:

```sh
chc init --dry-run
```

## Initial Layout

```text
AGENTS.md
.chc/
  manifest.md
  config.yml
  agents/
    ARCHITECTURE.md
    CONVENTIONS.md
    DOCUMENTING.md
    INVARIANTS.md
    LIFECYCLE.md
    LINTING.md
    PLAYBOOKS.md
    ROUTING.md
    SLICES.md
    VERSIONING.md
    playbooks/
  docs/
  specs/
  tickets/
  comments/
  context/
  dumps/
  quizzes/
  playbooks/
  reports/
  indexes/
  runtime/
    agent/
      Dockerfile
  notifications/
```

## Source of Truth

Markdown files are canonical assets.

Indexes, databases, runtime traces, and notification state are supporting artifacts. They may be rebuilt or regenerated when their source assets allow it.

## Runtime Config

`.chc/config.yml` may define the project-local agent runtime contract under
`runtime:`.

Supported runtime fields:

- `image`: Docker image tag used for the agent runtime.
- `dockerfile`: project-relative path to the agent runtime Dockerfile.
- `workdir`: container working directory used by future execution commands.
- `artifact_dir`: project-relative directory for host-readable runtime results.
- `validation_command`: command future runtime validation uses inside the configured boundary.

When runtime fields are absent, `chc` uses safe local defaults:

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
```

## Runtime Template

`chc init` creates `.chc/runtime/agent/Dockerfile` for new managed projects.
This file is the project-local agent runtime template. It is owned by the
managed project after initialization and may be edited to install project
dependencies, agent tooling, or validation prerequisites.

`chc init` does not overwrite an existing runtime Dockerfile.

## Go Package Boundary

Project workspace behavior lives in the core `internal/chc` package.

The CLI entrypoint under `cmd/chc` wires commands and delegates workspace behavior to `chc.Workspace`.
