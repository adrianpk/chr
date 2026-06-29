# Getting Started

This tutorial creates a managed project workspace, adds a ticket, and previews
the runtime boundary.

## Build the CLI

From the Chryso repository:

```sh
make build
```

The binary is written to:

```text
bin/chr
```

Use `make install` if you want `chr` on your Go binary path.

## Initialize a Project

Move to a project you want to manage:

```sh
cd /path/to/project
chr init
```

This creates:

```text
AGENTS.md
.chr/
  config.yml
  manifest.md
  tickets/
  runtime/
    agent/
      Dockerfile
```

`chr init` does not overwrite an existing runtime Dockerfile.

## Validate the Runtime Boundary

Preview the runtime plan:

```sh
chr runtime validate --dry-run
```

The command prints the effective Dockerfile, image, container workdir, mount,
artifact paths, and local path checks. Runtime defaults come from `chr` unless
the project overrides them in `.chr/user/config.yml`.

## Create a Ticket

Create a Markdown ticket:

```sh
chr ticket new "Fix the failing task workflow" \
  --synopsis "Fix the failing task workflow." \
  --type bug \
  --priority high
```

List indexed tickets:

```sh
chr ticket list
```

Show one ticket:

```sh
chr ticket show CHR-1
```

## Register the Project

Register the project for user-level coordination:

```sh
chr project add /path/to/project
chr project select /path/to/project
```

Check the active project:

```sh
chr project current
```

## Preview Ticket Execution

Prepare one ticket execution without running the worker:

```sh
chr ticket start CHR-1 --dry-run
```

This verifies the runtime boundary and prints execution-scoped artifact paths.

## Next Reading

- [User guide](../how-to/user-guide/index.md)
- [Project workspace reference](../reference/project-workspace.md)
- [Ticket execution reference](../reference/ticket-execution.md)
