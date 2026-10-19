# Project Workspace

## Root

Chryso uses `.chr/` as the project-local hidden directory.

The `chr` command creates and manages this directory inside the project where it runs. The directory is intentionally tied to the CLI name rather than the target repository name.

Project ticket files and project protocol assets are local to `.chr/` inside each managed project.

Chryso may also keep user-level state outside the project for daemon operation, project registration, and derived indexes. User-level state does not replace the project-local `.chr/` source files.

Board and ticket operations use both layers:

- Markdown tickets under `.chr/default/tickets/` are the durable ticket records.
- The user-level SQLite index stores derived ticket metadata for fast listing,
  Board columns, attention detection, and ticket key allocation.
- `.chr/default/` is the project-local active userspace boundary for workflows
  that need an operator-owned location.

Managed projects also receive a root-level `AGENTS.md` for compatibility with agent tooling that looks for this file at the repository root.

## Initialization

Create the workspace in the current directory:

```sh
chr init
```

`chr init` uses the `standard` profile by default.

Create the workspace with a specific profile:

```sh
chr init --profile minimal
```

Supported initial profiles:

- `standard`: default useful profile with agent protocol files.
- `personal`: full opinionated workflow for the author's own use.
- `minimal`: lightweight scaffold.

Create the workspace in another project directory:

```sh
chr init --dir /path/to/project
```

`--dir` is an override for scripts or external initialization. The normal workflow is to change into the project and run `chr init`.

Preview initialization without writing files:

```sh
chr init --dry-run
```

## Initial Layout

```text
AGENTS.md
.chr/
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

`.chr/default/` is the current project-local active userspace.

The MVP layout is intentionally single-active-operator. Workflows that need to
bind a write to an operator resolve the operator id from:

```sh
CHR_OPERATOR_ID
```

For direnv users, a project may set it in `.envrc`:

```sh
export CHR_OPERATOR_ID=adrian
```

Users who do not use direnv should export the same variable in their shell or
automation environment before running workflows that require operator
ownership.

The future multi-operator shape is expected to use an operator namespace such
as `.chr/operators/<operator-id>/...`. The current `.chr/default/` path prepares
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
`.chr/default/tmp/`. That cache is discardable and must not become the durable source of
truth.

The `.chr/` directory and Markdown-first artifact model are Chryso
invariants. They are not configured by generated workspace files.

The generated files under `.chr/default/agents/` guide how an agent should work inside
the managed project. They do not define or override Chryso's operational
artifact layout. Directory placement for tickets, runtime evidence, reports,
indexes, and authoring artifacts is a product contract owned by `chr`.

## Runtime Config

`.chr/default/config.yml` may define the project-local agent runtime contract under
`runtime:`.

Supported runtime fields:

- `image`: Docker image tag used for the agent runtime.
- `dockerfile`: project-relative path to the agent runtime Dockerfile.
- `workdir`: container working directory used by future execution commands.
- `artifact_dir`: project-relative directory for host-readable runtime results.
- `validation_command`: command future runtime validation uses inside the configured boundary.

Generated config omits runtime defaults. When runtime fields are absent, `chr`
uses safe local defaults:

```yml
runtime:
  image: chr-agent-runtime:local
  dockerfile: .chr/default/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: ops/default/run
  validation_command: make verify
```

## Work Config

`.chr/default/config.yml` may define the project-local work readiness policy under
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
- `allow_safe_repair`: whether Chryso may run non-destructive readiness
  repairs.
- `allow_branch_creation`: whether safe repair may create the configured local
  branch.
- `workspace_mode`: workspace execution mode. Defaults to `project`.
- `worktree_root`: required canonical `ops/default/worktree` root for ticket worktrees; the active namespace replaces `default` at runtime. Custom and legacy roots are rejected.
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
  worktree_root: ops/default/worktree
  worktree_base: remote_integration_branch
  worktree_cleanup: after_merge
```

See [Work readiness](work-readiness.md) for diagnostics, repair behavior, and
Dash integration.

## Web Config

`.chr/default/config.yml` may define the optional local Web UI contract under
`web:`.

The Web UI is disabled by default as a background behavior. Project commands,
the default TUI, agent work, ticket loops, and normal CLI operations do not
start a server. The server starts only through:

```sh
chr web serve
```

Supported web fields:

- `enabled`: operator intent marker for project configuration. The server still
  requires the explicit `chr web serve` command.
- `host`: bind host. Defaults to `127.0.0.1`.
- `port`: bind port. Defaults to `7331`.
- `open_browser`: whether startup should request opening a browser.
- `require_token`: one of `auto`, `true`, `false`, or `always`.
- `token_env`: environment variable that contains the Web UI token when token
  protection is required. Defaults to `CHR_WEB_TOKEN`.
- `allowed_origins`: comma-separated exact origins allowed for cross-origin
  browser requests. Empty means same-origin only.
- `refresh_interval_seconds`: polling interval for browser refresh fragments.
  Defaults to `2`.

Example local-only config:

```yml
web:
  enabled: true
  host: 127.0.0.1
  port: 7331
  open_browser: false
  require_token: auto
  token_env: CHR_WEB_TOKEN
  allowed_origins: ""
  refresh_interval_seconds: 2
```

Example private-network or tunnel config:

```yml
web:
  enabled: true
  host: 0.0.0.0
  port: 7331
  open_browser: false
  require_token: auto
  token_env: CHR_WEB_TOKEN
  allowed_origins: "https://chr.example.internal"
  refresh_interval_seconds: 2
```

Non-loopback binds print an exposure warning. In `auto` mode, non-loopback binds
also require the token value to be present in `token_env`.

## Web Harness Defaults

Open **Harnesses** in the Web UI project context to inspect the worker and
content-generation defaults. Each named profile shows its harness, model and
local readiness for that role. Local readiness does not test remote
authentication. Prepare executables, runtimes and credentials privately before
selecting a profile.

Choose a profile for either role and select **Save defaults**. **Keep current
default** leaves that role unchanged. **Restore legacy dispatch** clears its
selection: workers use the existing Codex default and generation keeps its
operation-specific dispatch. Changes are persisted for the current project and
become visible to new operations without restarting the Web service.

Running, paused and resumed operations keep their admitted bindings. The chooser
lists the first eight retained worker invocations separately from current
defaults; **View workflows** opens their existing status view.

Saves use the current project, configuration revision and a retained request
identity through the existing Web authentication and origin guards. A stale or
invalid save keeps the attempted choices. Use **Reload current defaults** to
inspect a newer configuration before saving again. If an acknowledgement is
lost, retry the same form to recover the original receipt. Closing the chooser
discards unsent choices. The control edits named selections only; private paths,
credentials and runtime setup remain outside the form.

## TUI Harness Defaults

In Project (`F1`), press `h` to open the harness picker. It reads the same
persisted project selections and prepared named profiles as the Web control.
Worker and generation choices are separate from the retained worker bindings.
Local readiness does not test remote authentication.

Use Tab or Shift+Tab to select a role and arrows (or `j`/`k`) to cycle its
profiles. **Keep current default** leaves that role unchanged. Press `c` to
choose **Restore legacy dispatch**, then Enter or Ctrl+S to save. Both changed
roles are published together using the observed revision, exact choices and
request identity. The application resolves the active local operator.

A rejected or unacknowledged save keeps its choices and original request.
Retry unchanged to recover a committed receipt. If another process changed the
configuration, press `r` to discard the preview and reload before saving.
Esc discards unsent choices; an in-flight save finishes before the picker can
close. Changes affect new operations without restarting the TUI or service.
Admitted work keeps its retained binding.

PgUp/PgDn scroll the compact content while controls remain visible. The picker
shows the first four retained worker invocations. For further inspection, close
the picker and use **Inspect retained workflow status** in the command palette.
Profile inspection and selection do not provision credentials, install runtimes
or call a model.

## Telegram Harness Defaults

After `/project <name>`, send `/harness` to inspect the same persisted defaults,
locally prepared profiles and retained worker bindings. Send
`/harness worker=PROFILE generation=PROFILE` to preview an exact change, then
confirm its retained prompt. Omit a role to keep it; use `-` to restore legacy
dispatch. Both changed roles publish together under the mapped operator and
observed configuration revision. Other interfaces see the update on their next
operation; admitted work keeps its original binding.

See [Telegram harness defaults](telegram-receiver.md#harness-defaults) for
confirmation expiry, stale requests and recovery. Private profile preparation
remains local; inspection and selection invoke no model.

## Notification Config

`.chr/default/config.yml` may also define notification settings under
`notifications:`.

Notifications are disabled by default. Telegram can be enabled with:

```yml
notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHR_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
```

The token value is read from the named environment variable by later dispatch
code. It is not stored in project configuration.

## Runtime Template

`chr init` creates `.chr/default/runtime/agent/Dockerfile` for new managed projects.
This file is the project-local agent runtime template. It is owned by the
managed project after initialization and may be edited to install project
dependencies, agent tooling, or validation prerequisites.

`chr init` does not overwrite an existing runtime Dockerfile.

Existing managed projects can install missing runtime defaults without running
full workspace initialization:

```sh
chr runtime init
```

Preview runtime initialization without writing files:

```sh
chr runtime init --dry-run
```

Initialize runtime defaults for another project directory:

```sh
chr runtime init --dir /path/to/project
```

### Pi runtime assets

All bundled profiles also install `.chr/default/runtime/agent/pi/` with a
sibling Dockerfile, minimal npm manifest and complete dependency lock. Existing
projects add these missing assets with `chr runtime init`; this explicit upgrade
preserves every existing runtime file and the configured Codex image/Dockerfile.
Installing assets does not install Pi, build an image or change role selections.

Build the Pi image explicitly, using that directory as its build context:

```sh
docker build --file .chr/default/runtime/agent/pi/Dockerfile \
  --tag chr-agent-pi:1.3.0 .chr/default/runtime/agent/pi
```

The image pins the Go base manifest, verifies the Node 22.23.3 Linux amd64/arm64
archive checksum and installs Pi 1.3.0 through `npm ci` with the committed lock.
Dependency lifecycle scripts are disabled. Other Linux architectures fail the
build with a preparation error. Codex-only use does not probe or build Pi.
Runtime inspection validates the Linux platform and Pi version labels and
returns the immutable Docker image ID for worker admission.

### Prepared Pi worker profile

A Pi worker profile in the private user `harnesses.yml` names the Linux CHR
executable that supplies its bridge, the prepared image and explicit resources.

Build that Linux bridge with `CGO_ENABLED=0 go build -o /opt/chr/bin/chr ./cmd/chr`
from the selected source revision, using a writable installation path. Readiness
inspects its build metadata without executing it and requires the host's Linux
architecture. Generation uses a separate native Pi executable profile. For example:

```yaml
profiles:
  pi-worker:
    harness: pi
    command: /opt/chr/bin/chr
    model_provider: example-provider
    model: exact-model
    thinking: off
    credential_ref: pi-worker/auth.json
    model_config_ref: pi-worker-models.json
    runtime_ref: chr-agent-pi:1.3.0
```

Resource references are relative to that private configuration directory unless
absolute. Keep `harnesses.yml`, authentication and model settings private. The
authentication file contains only the selected provider. For OAuth, its parent
must be a private directory containing only that auth file and the managed
refresh lock; keep model settings outside that directory. The bridge serializes
refreshes across processes and publishes valid refreshed credentials atomically.
A static API-key source is mounted read-only.

Worker admission freezes the bridge executable and immutable image ID. A later
profile change or image retag cannot rebind admitted work. Credentials may rotate
at their retained reference; changed model definitions or credential types fail
instead of changing the operation. Prepare a native Pi executable profile for
generation separately from the container worker bridge profile.

Each worker receives its execution checkout and Git metadata, the selected
resources and private writable state. It receives no host home, Docker socket,
SSH configuration or blanket environment forwarding. Pi runs from private state;
only the host-owned extension binds its seven builtin tools to the execution
checkout. Project instructions are supplied explicitly. Discovered Pi settings,
extensions, MCP servers, skills and context files cannot replace the selection.

Custom endpoints must be reachable from the container. Literal loopback endpoints
are rejected. Explicit `host.docker.internal` endpoints add Docker's host-gateway
mapping; host firewall and listener policy still apply. A container-network
endpoint is another valid choice. The bridge does not enable host networking.
Local readiness checks preparation; they do not prove remote authentication.

Workers retain the existing invocation, claims, retries, shared capacity and host
validation. The runtime holds the same admitted capacity reservation until its
processes stop, including after a host disconnect. Pi's observed usage does not
certify a finite whole-call budget or qualify it as a merger broker.

## Runtime Plan

The runtime planner computes a side-effect-free Docker boundary plan from the
project directory and `.chr/default/config.yml` runtime fields.

Path semantics:

- `project_root`: absolute path to the managed project.
- `project_checkout`: absolute path to the user-visible checkout.
- `execution_workspace`: absolute path where agent execution is expected to
  operate. In `project` mode this is the same path as `project_checkout`. In
  `worktree` mode this is the prepared ticket worktree.
- `workspace_mode`: effective work workspace mode.
- `workspace_root`: absolute path to `.chr/` inside the project.
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
changes only the execution workspace: Chryso prepares a ticket-specific Git
worktree and mounts that path into Docker.

The planner only computes paths and mount metadata. It does not create
directories, build images, run Docker, or execute validation commands.

## Runtime Validation

Validate the configured runtime boundary:

```sh
chr runtime validate
```

Preview the validation plan without invoking Docker or running the configured
validation command:

```sh
chr runtime validate --dry-run
```

Validate another managed project:

```sh
chr runtime validate --dir /path/to/project --dry-run
```

Runtime validation loads `.chr/default/config.yml`, computes the runtime plan, prints
the Docker build/run inputs, and checks required local paths:

- `.chr/` workspace directory;
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

The CLI entrypoint under `cmd/chr` wires commands and delegates workspace behavior to `chr.Workspace`.
