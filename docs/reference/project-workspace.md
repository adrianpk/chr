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
  authoring/
    sessions/
  dumps/
  quizzes/
  drafts/
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

## Notification Config

`.chc/config.yml` may also define notification settings under
`notifications:`.

Notifications are disabled by default. Telegram can be enabled with:

```yml
notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
```

The token value is read from the named environment variable by later dispatch
code. It is not stored in project configuration.

## Runtime Template

`chc init` creates `.chc/runtime/agent/Dockerfile` for new managed projects.
This file is the project-local agent runtime template. It is owned by the
managed project after initialization and may be edited to install project
dependencies, agent tooling, or validation prerequisites.

`chc init` does not overwrite an existing runtime Dockerfile.

Existing managed projects can install missing runtime defaults without running
full workspace initialization:

```sh
chc runtime init
```

Preview runtime initialization without writing files:

```sh
chc runtime init --dry-run
```

Initialize runtime defaults for another project directory:

```sh
chc runtime init --dir /path/to/project
```

## Runtime Plan

The runtime planner computes a side-effect-free Docker boundary plan from the
project directory and `.chc/config.yml` runtime fields.

Path semantics:

- `project_root`: absolute path to the managed project.
- `workspace_root`: absolute path to `.chc/` inside the project.
- `dockerfile`: absolute path resolved from `runtime.dockerfile`.
- `artifact_dir`: absolute path resolved from `runtime.artifact_dir`.
- `run_artifact_dir`: `artifact_dir` or `artifact_dir/<run_id>` when a run ID is supplied.
- `result_path`: `run_artifact_dir/result.json`.
- `log_path`: `run_artifact_dir/runtime.log`.

Docker mount contract:

- the project root is mounted read-write into the container;
- the container target is `runtime.workdir`;
- relative container workdirs are normalized to absolute paths.

The planner only computes paths and mount metadata. It does not create
directories, build images, run Docker, or execute validation commands.

## Runtime Validation

Validate the configured runtime boundary:

```sh
chc runtime validate
```

Preview the validation plan without invoking Docker or running the configured
validation command:

```sh
chc runtime validate --dry-run
```

Validate another managed project:

```sh
chc runtime validate --dir /path/to/project --dry-run
```

Runtime validation loads `.chc/config.yml`, computes the runtime plan, prints
the Docker build/run inputs, and checks required local paths:

- `.chc/` workspace directory;
- configured runtime Dockerfile;
- configured runtime artifact directory.

Validation output includes:

- image tag;
- project-relative Dockerfile path when the file is inside the managed project;
- container workdir;
- project-relative artifact, result, and log paths when they are inside the
  managed project;
- project mount summary;
- a compact `checks:` section for required local paths.

This command does not build Docker images, run containers, execute agents, or
run `runtime.validation_command`.

## Go Package Boundary

Project workspace behavior lives in the core `internal/chc` package.

The CLI entrypoint under `cmd/chc` wires commands and delegates workspace behavior to `chc.Workspace`.
