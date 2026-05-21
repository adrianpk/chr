# ChrysoCoord

ChrysoCoord is a local coordination tool for agent-driven software work.

It gives a project a durable `.chc/` workspace, keeps ticket files in Markdown,
tracks local projects in a user-level registry, and can run ticket execution
through a Docker-based Codex worker boundary.

The current system is intentionally local. Project state lives in the managed
workspace. User-level state stores indexes, project registration, daemon state,
loop records, and execution records.

## What It Does Today

- Initializes a project-local `.chc/` workspace.
- Creates and indexes Markdown tickets.
- Registers local projects for user-level coordination.
- Resolves an active project or an explicit project directory.
- Validates the project runtime boundary.
- Prepares and runs ticket execution records.
- Invokes Codex inside the configured Docker runtime.
- Captures result, runtime log, and validation artifacts.
- Applies execution results back to tickets.
- Runs foreground ticket loops over eligible tickets.
- Coordinates registered projects through a local foreground daemon.
- Builds a bounded execution context package for agent work.
- Creates and promotes Markdown-first authoring sessions.

## Install

Build the CLI locally:

```sh
make build
```

Run it from the repository:

```sh
./bin/chc --help
```

Install it into the active Go environment:

```sh
make install
```

## First Project

Inside a project you want ChrysoCoord to manage:

```sh
chc init
chc runtime validate --dry-run
chc ticket new "Fix the failing task workflow" --type bug --priority high
chc ticket list
```

To use the project registry:

```sh
chc project add /path/to/project
chc project select /path/to/project
chc project current
```

## Agent Runtime

`chc init` creates `.chc/runtime/agent/Dockerfile`. That Dockerfile defines the
runtime used for agent execution.

The project runtime config lives in `.chc/config.yml`. By default, validation
inside the runtime is:

```sh
make verify
```

Ticket execution uses Docker as the sandbox boundary. The Codex worker receives
its instructions through `context.md`, writes `result.json`, and leaves runtime
and validation logs under `.chc/runtime/runs/`.

## Documentation

Start with the user guide:

- [User guide](docs/how-to/user-guide/index.md)
- [Authoring sessions reference](docs/reference/authoring-sessions.md)

Documentation is organized with Diataxis:

- [Documentation index](docs/index.md)
- [Tutorials](docs/tutorials/getting-started.md)
- [How-to guides](docs/how-to/user-guide/index.md)
- [Reference](docs/reference/project-workspace.md)
- [Explanation](docs/explanation/architecture.md)

## Development

Common local checks:

```sh
make fmt
make test
make verify
```

Run the CLI without installing:

```sh
make run RUN_ARGS="--help"
```

The main Go package is `forge.adrianpk.com/chc/internal/chc`. The CLI entrypoint
is `cmd/chc`.
