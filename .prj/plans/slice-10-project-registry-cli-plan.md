# Slice 10 Project Registry CLI Plan

Status: delivered

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`
Tracker: `.prj/trackers/slice-10-project-registry-cli-tracker.md`

## Purpose

Expose the project registry through CLI commands.

This slice turns the core project registry repository into usable one-shot commands while keeping daemon, watcher, Telegram, and ticket indexing out of scope.

## Scope

This slice covers:

- `chr project add <dir>`,
- `chr project list`,
- user environment construction from process paths,
- development-mode path redirection for CLI tests and local development,
- stable local project id generation,
- project name defaulting from the project directory basename,
- focused command tests.

## Task Breakdown

### T10.1 CLI User Environment Wiring

Add CLI helpers for constructing `chr.UserEnvironment`.

Expected behavior:

- normal mode reads home and XDG directories from the process environment,
- development mode uses an explicit dev root flag,
- construction does not write files by itself.

### T10.2 Project Add Command

Add `chr project add <dir>`.

Expected behavior:

- resolve the project path to an absolute path,
- require an existing `.chr/` directory,
- default project name from the project directory basename,
- allow overriding the name with `--name`,
- generate a stable local id from the absolute path,
- register the project in the user-level SQLite database.

### T10.3 Project List Command

Add `chr project list`.

Expected behavior:

- read projects from the user-level SQLite database,
- print stable, parseable rows,
- support development-mode path redirection,
- update documentation and tracker,
- run validation.

## Out of Scope

This slice does not cover:

- project reindex,
- ticket indexing,
- daemon command,
- filesystem watcher,
- Telegram integration.

## Validation

Required checks:

```sh
make verify
```
