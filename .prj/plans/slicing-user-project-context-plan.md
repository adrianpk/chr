# User Project Context Slicing Plan

Status: active

## Purpose

Plan the `User Project Context` slicing from the feature slicing set map.

This slicing covers the use case:

> A user registers local projects and selects which project `chc` should operate on by default.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related specifications:

- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`

Related delivered plans:

- `.prj/plans/slice-7-user-environment-paths-plan.md`
- `.prj/plans/slice-9-project-registry-repository-plan.md`
- `.prj/plans/slice-10-project-registry-cli-plan.md`

## Current State

The user-level project foundation is partially delivered.

Already delivered:

- user environment path resolution;
- development-mode path redirection;
- project registry database;
- project registry repository;
- `chc project add <dir>`;
- `chc project list`;
- project path validation for registration.

Still missing:

- active project persistence;
- `chc project select`;
- `chc project current`;
- a reusable project context resolver for commands that can operate on the active project;
- recovery behavior when the active project is missing, disabled, or no longer has a `.chc/` workspace.

## Slicing Scope

This slicing includes:

- user-level project registry access;
- active project selection;
- active project inspection;
- project context resolution for commands;
- validation and recovery behavior for registered project paths;
- development-mode coverage for user-level project context.

## Out of Scope

This slicing does not include:

- ticket list or ticket show;
- ticket indexing behavior beyond project context needs;
- automatic ticket reconciliation;
- Docker runtime setup;
- agent execution;
- daemon process lifecycle;
- notifications.

## Slices

### User Environment Paths

Status: delivered.

Delivery:

- normal-mode user paths;
- development-mode path redirection;
- CLI construction of `chc.UserEnvironment`.

### Project Registry Persistence

Status: delivered.

Delivery:

- SQLite-backed project registry;
- project migration;
- add and list repository operations;
- lookup by project name.

### Project Registry CLI

Status: delivered.

Delivery:

- `chc project add <dir>`;
- `chc project list`;
- project name defaulting;
- stable project id generation;
- `.chc/` requirement for registered projects.

### Active Project Persistence

Status: pending.

Goal:

Persist the user's selected project in user-level state.

Candidate scope:

- define where active project state lives;
- persist selected project id or name;
- read active project state;
- clear or replace active project state;
- table-based tests using fakes or an isolated database;
- development-mode behavior under `_workspace`.

Expected outcome:

- core code can store and load the active project independently from CLI commands.

### Project Select and Current CLI

Status: pending.

Goal:

Expose active project selection and inspection through CLI.

Candidate scope:

- `chc project select <name-or-path>`;
- accept a registered project name;
- accept a registered project path when unambiguous;
- reject unknown or disabled projects clearly;
- `chc project current`;
- stable, parseable output;
- focused CLI tests.

Expected outcome:

- the user can choose an active project once and inspect it later.

### Project Context Resolver

Status: pending.

Goal:

Provide one reusable command boundary for resolving the project a command should operate on.

Candidate scope:

- resolve explicit `--project` or `--dir` when present;
- fall back to active project state;
- fall back to current directory only for commands whose semantics allow it;
- validate `.chc/` exists before returning a project context;
- return clear errors when no project can be resolved;
- avoid duplicating project resolution logic across commands.

Expected outcome:

- later ticket, execution, and loop commands can use the same project context behavior.

### Project Validation and Recovery

Status: pending.

Goal:

Make stale registered project paths understandable and recoverable.

Candidate scope:

- detect missing project paths;
- detect registered paths without `.chc/`;
- report disabled projects distinctly from missing projects;
- provide recovery-oriented errors;
- decide whether disable/remove commands belong in this slicing or a later maintenance slicing.

Expected outcome:

- project context failures are actionable and do not silently select the wrong project.

## Acceptance Criteria

The slicing is complete when:

- a project can be registered;
- registered projects can be listed;
- one registered project can be selected as active;
- the active project can be inspected;
- command code has a reusable way to resolve project context;
- invalid, missing, or stale project paths produce clear errors;
- development-mode paths keep tests and local use isolated from real user state;
- validation passes with `make verify`.

## Validation

Required validation for every implementation slice in this slicing:

```sh
make verify
```

Manual smoke checks for the completed slicing:

```sh
make install
tmpdir="$(mktemp -d)"
chc init --dir "$tmpdir"
chc project add "$tmpdir" --name sample --dev-root _workspace
chc project select sample --dev-root _workspace
chc project current --dev-root _workspace
```

## Risks

- Active project state can become stale if the project directory is moved or deleted.
- Using names and paths for selection can create ambiguity if output and errors are not precise.
- Project context resolution can become inconsistent if each command resolves project paths independently.
- Development-mode path behavior must not write into the user's real `~/.chc` or `~/.config/chc`.

## Notes

- No quiz was needed for this slicing because the target behavior follows directly from the operational MVP path.
- Manual sync or ticket indexing behavior belongs to later slicings.
