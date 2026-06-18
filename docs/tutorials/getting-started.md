# Getting Started

This tutorial creates a managed project workspace, adds a ticket, and previews
the runtime boundary.

## Build the CLI

From the ChrysoCoord repository:

```sh
make build
```

The binary is written to:

```text
bin/chc
```

Use `make install` if you want `chc` on your Go binary path.

## Initialize a Project

Move to a project you want to manage:

```sh
cd /path/to/project
chc init
```

This creates:

```text
AGENTS.md
.chc/
  config.yml
  manifest.md
  tickets/
  runtime/
    agent/
      Dockerfile
```

`chc init` does not overwrite an existing runtime Dockerfile.

## Validate the Runtime Boundary

Preview the runtime plan:

```sh
chc runtime validate --dry-run
```

The command prints the effective Dockerfile, image, container workdir, mount,
artifact paths, and local path checks. Runtime defaults come from `chc` unless
the project overrides them in `.chc/user/config.yml`.

## Create a Ticket

Create a Markdown ticket:

```sh
chc ticket new "Fix the failing task workflow" \
  --synopsis "Fix the failing task workflow." \
  --type bug \
  --priority high
```

List indexed tickets:

```sh
chc ticket list
```

Show one ticket:

```sh
chc ticket show CHC-1
```

## Register the Project

Register the project for user-level coordination:

```sh
chc project add /path/to/project
chc project select /path/to/project
```

Check the active project:

```sh
chc project current
```

## Preview Ticket Execution

Prepare one ticket execution without running the worker:

```sh
chc ticket start CHC-1 --dry-run
```

This verifies the runtime boundary and prints execution-scoped artifact paths.

## Next Reading

- [User guide](../how-to/user-guide/index.md)
- [Project workspace reference](../reference/project-workspace.md)
- [Ticket execution reference](../reference/ticket-execution.md)
