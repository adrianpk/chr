# Asset Embedding Approach

Status: draft

## Purpose

Define the local ChrysoCoord approach for embedding default assets into the `chc` binary while keeping those assets editable as normal repository files.

## Current Direction

The executable entrypoint owns a single embedded filesystem.

Internal packages should receive that filesystem through constructors instead of importing a global asset package.

This keeps runtime assembly visible in `main` and lets internal packages stay explicit about the assets they need.

## Asset Source Layout

Default assets should live as non-binarized files under the command asset tree.

Current candidate:

```text
cmd/chc/assets/
  workspace/
    profiles/
      standard/
      personal/
      minimal/
```

The binary embeds the top-level `assets` directory once.

## Init Profiles

ChrysoCoord supports three initial workspace profiles:

- `standard`: the default profile used by `chc init`.
- `personal`: the full opinionated workflow for the author's own use.
- `minimal`: a lightweight scaffold with the least practical default content.

The command shape is:

```sh
chc init
chc init --profile standard
chc init --profile personal
chc init --profile minimal
```

`standard` is selected when no profile is provided.

## Profile Layout

The profile asset layout is:

```text
cmd/chc/assets/workspace/profiles/
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

Profiles are copied recursively into the target `.chc/` workspace.

Users may edit generated `.chc/` files freely after initialization.

## Root AGENTS.md

`AGENTS.md` is a required root-level file in managed projects.

Even though most ChrysoCoord-managed state lives under `.chc/`, `AGENTS.md` must be materialized at the project root because external tools may look for it there.

The root file should point to the project-local ChrysoCoord protocol under `.chc/` while still being useful to tools that only read `AGENTS.md`.

## Default Agent Harness

Codex is the planned default agent harness.

The execution boundary still needs implementation design, but Docker is the preferred direction for safe harness execution because it provides a stronger sandbox than a plain git worktree.

Git worktrees remain useful for branch and checkout isolation, but they should not be treated as a security sandbox by themselves.

## Runtime Assembly

`cmd/chc/main.go` declares the embedded filesystem:

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

Those files become the defaults written to a managed project's `.chc/` directory.

## Open Questions

- Should all command-owned assets stay under `cmd/chc/assets/`, or should the project later move to a root `assets/` directory?
- What exact file set belongs in each profile?
- Should tests use `testing/fstest.MapFS` for focused asset fixtures, or embed testdata assets per package?
