# Project Workspace

## Root

ChrysoCoord uses `.chc/` as the project-local hidden directory.

The `chc` command creates and manages this directory inside a target project. The directory is intentionally tied to the CLI name rather than the target repository name.

## Initial Layout

```text
.chc/
  manifest.md
  config.yml
  agents/
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
