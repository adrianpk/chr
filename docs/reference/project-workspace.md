# Project Workspace

## Root

ChrysoCoord uses `.chc/` as the project-local hidden directory.

The `chc` command creates and manages this directory inside the project where it runs. The directory is intentionally tied to the CLI name rather than the target repository name.

ChrysoCoord does not keep a global project registry. Project state is local to `.chc/` inside each managed project.

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
  notifications/
```

## Source of Truth

Markdown files are canonical assets.

Indexes, databases, runtime traces, and notification state are supporting artifacts. They may be rebuilt or regenerated when their source assets allow it.
