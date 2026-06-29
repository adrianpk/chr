# Asset Embedding Approach

Status: draft

## Purpose

Define the local Chryso approach for embedding default assets into the `chr` binary while keeping those assets editable as normal repository files.

## Current Direction

The executable entrypoint owns a single embedded filesystem.

Internal packages should receive that filesystem through constructors instead of importing a global asset package.

This keeps runtime assembly visible in `main` and lets internal packages stay explicit about the assets they need.

## Asset Source Layout

Default assets should live as non-binarized files under the command asset tree.

Current candidate:

```text
cmd/chr/assets/
  workspace/
    profiles/
      standard/
      personal/
      minimal/
```

The binary embeds the top-level `assets` directory once.

## Init Profiles

Chryso supports three initial workspace profiles:

- `standard`: the default profile used by `chr init`.
- `personal`: the full opinionated workflow for the author's own use.
- `minimal`: a lightweight scaffold with the least practical default content.

The command shape is:

```sh
chr init
chr init --profile standard
chr init --profile personal
chr init --profile minimal
```

`standard` is selected when no profile is provided.

## Profile Layout

The profile asset layout is:

```text
cmd/chr/assets/workspace/profiles/
  standard/
    manifest.md
    config.yml
    agents/
    playbooks/
  personal/
    manifest.md
    config.yml
    agents/
    playbooks/
  minimal/
    manifest.md
    config.yml
    agents/
    playbooks/
```

Profiles are copied recursively into the target `.chr/` workspace.

Users may edit generated `.chr/` files freely after initialization.

## Root AGENTS.md

`AGENTS.md` is a required root-level file in managed projects.

Even though most Chryso-managed state lives under `.chr/`, `AGENTS.md` must be materialized at the project root because external tools may look for it there.

The root file should point to the project-local Chryso protocol under `.chr/` while still being useful to tools that only read `AGENTS.md`.

## Default Agent Harness

Codex is the planned default agent harness.

The execution boundary still needs implementation design, but Docker is the preferred direction for safe harness execution because it provides a stronger sandbox than a plain git worktree.

Git worktrees remain useful for branch and checkout isolation, but they should not be treated as a security sandbox by themselves.

## Runtime Assembly

`cmd/chr/main.go` declares the embedded filesystem:

```go
//go:embed assets
var assetsFS embed.FS
```

The CLI assembly then passes `assetsFS` to internal services through constructors.

## Internal Package Rule

Internal packages should not own the root embedded filesystem.

They may:

- accept `fs.FS` or `embed.FS` through constructors,
- read only their owned subdirectories,
- expose package behavior without knowing how the binary embedded the files.

## Initial Workspace Use

The workspace initializer reads defaults from:

```text
assets/workspace/profiles/<profile>/
```

Those files become the defaults written to a managed project's `.chr/` directory.

## Open Questions

- Should all command-owned assets stay under `cmd/chr/assets/`, or should the project later move to a root `assets/` directory?
- What exact file set belongs in each profile?
- Should tests use `testing/fstest.MapFS` for focused asset fixtures, or embed testdata assets per package?
