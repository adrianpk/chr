# Project Workspace

## Root

ChrysoCoord uses `.chc/` as the project-local hidden directory.

The `chc` command creates and manages this directory inside the project where it runs. The directory is intentionally tied to the CLI name rather than the target repository name.

Project ticket files and project protocol assets are local to `.chc/` inside each managed project.

ChrysoCoord may also keep user-level state outside the project for daemon operation, project registration, and derived indexes. User-level state does not replace the project-local `.chc/` source files.

Board and ticket operations use both layers:

- Markdown tickets under `.chc/user/tickets/` are the durable ticket records.
- The user-level SQLite index stores derived ticket metadata for fast listing,
  Board columns, attention detection, and ticket key allocation.
- `.chc/user/` is the project-local active userspace boundary for workflows
  that need an operator-owned location.

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
  acceptance/
    <case>.feature
    <case>-feature-round.md
  playbooks/
  reports/
  indexes/
  user/
  runtime/
    agent/
      Dockerfile
  notifications/
```

## Active Userspace

`.chc/user/` is the current project-local active userspace.

The MVP layout is intentionally single-active-operator. Workflows that need to
bind a write to an operator resolve the operator id from:

```sh
CHC_OPERATOR_ID
```

For direnv users, a project may set it in `.envrc`:

```sh
export CHC_OPERATOR_ID=adrian
```

Users who do not use direnv should export the same variable in their shell or
automation environment before running workflows that require operator
ownership.

The future multi-operator shape is expected to use an operator namespace such
as `.chc/operators/<operator-id>/...`. The current `.chc/user/` path prepares
the boundary without relocating existing artifacts.

## Source of Truth

Markdown files are the human-readable project-local records.

Workspace metadata, project descriptions, cases, tickets, dumps, quizzes, specs,
plans, trackers, reports, and runtime evidence are durable as files in the
managed project.

Runtime indexes, caches, user-level databases, runtime traces, and notification
state are supporting artifacts. They may be rebuilt or regenerated when their
source assets allow it.

The workspace index is an in-memory structural model that may be cached under
`.chc/user/tmp/`. That cache is discardable and must not become the durable source of
truth.

The `.chc/` directory and Markdown-first artifact model are ChrysoCoord
invariants. They are not configured by generated workspace files.

The generated files under `.chc/user/agents/` guide how an agent should work inside
the managed project. They do not define or override ChrysoCoord's operational
artifact layout. Directory placement for tickets, runtime evidence, reports,
indexes, and authoring artifacts is a product contract owned by `chc`.

## Runtime Config

`.chc/user/config.yml` may define the project-local agent runtime contract under
`runtime:`.

Supported runtime fields:

- `image`: Docker image tag used for the agent runtime.
- `dockerfile`: project-relative path to the agent runtime Dockerfile.
- `workdir`: container working directory used by future execution commands.
- `artifact_dir`: project-relative directory for host-readable runtime results.
- `validation_command`: command future runtime validation uses inside the configured boundary.

Generated config omits runtime defaults. When runtime fields are absent, `chc`
uses safe local defaults:

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/user/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/user/runtime/runs
  validation_command: make verify
```

## Work Config

`.chc/user/config.yml` may define the project-local work readiness policy under
`work:`.

Supported work fields:

- `mode`: one of `integration_branch`, `current_branch`, or
  `no_git_enforcement`.
- `integration_branch`: branch expected by `integration_branch` mode and used
  as the default ticket merge base.
- `canonical_remote`: remote used for remote tracking checks.
- `require_clean_worktree`: whether visible changes block agent work.
- `require_remote_tracking`: whether the configured branch must track the
  configured remote ref.
- `allow_safe_repair`: whether ChrysoCoord may run non-destructive readiness
  repairs.
- `allow_branch_creation`: whether safe repair may create the configured local
  branch.
- `workspace_mode`: workspace execution mode. Defaults to `project`.
- `worktree_root`: project-relative or absolute root for ticket worktrees.
- `worktree_base`: base selection policy for ticket worktree creation.
  Supported values are `remote_integration_branch`, `local_integration_branch`,
  and `current_branch`.
- `worktree_cleanup`: cleanup policy for clean merged ticket worktrees.
  Supported values are `after_merge`, `manual`, and `never`.

Example:

```yml
work:
  mode: integration_branch
  integration_branch: dev
  canonical_remote: origin
  require_clean_worktree: true
  require_remote_tracking: true
  allow_safe_repair: true
  allow_branch_creation: false
  workspace_mode: project
  worktree_root: .chc/user/runtime/worktrees
  worktree_base: remote_integration_branch
  worktree_cleanup: after_merge
```

See [Work readiness](work-readiness.md) for diagnostics, repair behavior, and
Dash integration.

## Notification Config

`.chc/user/config.yml` may also define notification settings under
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

`chc init` creates `.chc/user/runtime/agent/Dockerfile` for new managed projects.
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
project directory and `.chc/user/config.yml` runtime fields.

Path semantics:

- `project_root`: absolute path to the managed project.
- `project_checkout`: absolute path to the user-visible checkout.
- `execution_workspace`: absolute path where agent execution is expected to
  operate. In `project` mode this is the same path as `project_checkout`. In
  `worktree` mode this is the prepared ticket worktree.
- `workspace_mode`: effective work workspace mode.
- `workspace_root`: absolute path to `.chc/` inside the project.
- `dockerfile`: absolute path resolved from `runtime.dockerfile`.
- `artifact_dir`: absolute path resolved from `runtime.artifact_dir`.
- `run_artifact_dir`: `artifact_dir` or `artifact_dir/<run_id>` when a run ID is supplied.
- `result_path`: `run_artifact_dir/result.json`.
- `log_path`: `run_artifact_dir/runtime.log`.

Docker mount contract:

- the execution workspace is mounted read-write into the container;
- the container target is `runtime.workdir`;
- relative container workdirs are normalized to absolute paths.

Project mode remains fully compatible with existing managed projects because
the execution workspace is the user-visible project checkout. Worktree mode
changes only the execution workspace: ChrysoCoord prepares a ticket-specific Git
worktree and mounts that path into Docker.

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

Runtime validation loads `.chc/user/config.yml`, computes the runtime plan, prints
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

Project workspace behavior lives in the core `internal/core` package.

The CLI entrypoint under `cmd/chc` wires commands and delegates workspace behavior to `chc.Workspace`.
