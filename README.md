# Chryso

<p align="center">
  <img src="assets/img/home/chryso-hero.png" alt="Chryso hero" width="760">
</p>

[Gallery](docs/explanation/gallery.md)

Chryso is a **local software factory built in Go**. It turns development goals
into specifications and tickets, then coordinates AI coding agents to execute
them in isolated environments.

A local TUI and Telegram interface provide access to the work queue, agent
execution, and results. Each execution leaves behind artifacts that can be
inspected and reviewed.

## Workflow

Idea → Dump and quiz → Spec → Plan → Tickets → Start loop → Execution → Review evidence

You provide the initial intent, project context, and answers to focused
questions. The agent does the heavy work after that. It formalizes the feature,
plans the implementation, creates tickets, executes them one by one or as a
ready queue, then records what happened.

Ticket work runs through a Docker-based sandboxed runtime boundary. When
enabled, ticket worktrees keep agent edits out of the user's active checkout
while keeping execution explicit and inspectable.

## Control Surface

Chryso provides three interfaces: a TUI, a CLI, and an optional local Web UI.
All three operate on the same project state.

The TUI and Web UI provide access to specifications, plans, tickets, the Kanban
board, agent activity, and execution results. The CLI exposes the same
operations for scripting and agent use.

The workflow is organized around a few artifacts:

- **Dumps and quizzes** capture and clarify context.
- **Specs and Gherkin features** define expected behavior.
- **Plans** break work into manageable steps.
- **Tickets** define executable units of work.
- **Kanban board** tracks progress and execution state.

Agents execute ready tickets in isolated environments. Execution can be
triggered for a single ticket or run continuously through the queue. Users can
stop execution, inspect results, and review work before integration.

Review happens at integration checkpoints rather than after every small task.
It focuses on consistency, scope, contracts, and execution evidence. See the
[Agent review model](https://forge.adrianpk.com/adrianpk/chr/src/branch/main/docs/explanation/agent-review-model.md).

The Web UI runs only when explicitly started with `chr web serve`. It binds to
localhost by default and can be exposed through private networking. It is
functional but still undergoing refinement.

## Install

Build the CLI locally:

```sh
make build
```

Run it from the repository:

```sh
./bin/chr --help
```

Install it into the active Go environment:

```sh
make install
```

## First Project

Inside a project you want Chryso to manage:

```sh
chr init
chr runtime validate --dry-run
chr ticket new "Fix the failing task workflow" --type bug --priority high
chr ticket list
```

This initializes the project workspace, checks the configured runtime boundary,
creates a first ticket, and lists the tickets available to the agent.

To use the project registry:

```sh
chr project add /path/to/project
chr project select /path/to/project
chr project current
```

## Documentation

- [User Guide](docs/how-to/user-guide/index.md)
- [Documentation index](docs/index.md)
- [Changelog](CHANGELOG.md)

## Repos

- Official: https://forge.adrianpk.com/adrianpk/chr
- Codeberg: https://codeberg.org/adrianpk/chr
- GitHub: https://github.com/adrianpk/chr

## License

Chryso is licensed under the [Apache License, Version 2.0](LICENSE).
