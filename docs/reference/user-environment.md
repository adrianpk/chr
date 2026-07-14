# User Environment

## Purpose

The user environment resolves user-level `chr` paths.

These paths are separate from project-local `.chr/` directories. Project workspaces contain canonical project files such as Markdown tickets. User-level paths contain local tool state such as future project registry data, global indexes, caches, and daemon state.

Project-local active userspace is separate from these user-level paths. It
lives at `.chr/default/` in a managed project and uses `CHR_OPERATOR_ID` to bind
operator-owned workflows to a local operator id.

## Normal Mode

Normal mode resolves paths from XDG-style directories when they are provided.

```text
config -> $XDG_CONFIG_HOME/chr
data   -> $XDG_DATA_HOME/chr
cache  -> $XDG_CACHE_HOME/chr
state  -> $XDG_STATE_HOME/chr
```

When XDG directories are not provided, paths fall back to the user's home directory.

```text
config -> ~/.config/chr
data   -> ~/.local/share/chr
cache  -> ~/.cache/chr
state  -> ~/.local/state/chr
```

## Development Mode

Development mode redirects all user-level paths into a caller-supplied sandbox root.

```text
config -> <dev-root>/config/chr
data   -> <dev-root>/data/chr
cache  -> <dev-root>/cache/chr
state  -> <dev-root>/state/chr
```

Development mode is intended to keep work on `chr` from writing to real user configuration or data directories.

## Construction

Core code uses `chr.NewUserEnvironment(...)` to construct a `chr.UserEnvironment`.

Construction resolves paths only. It does not create directories or write files.

## Active Operator

`CHR_OPERATOR_ID` identifies the active local operator for project-local
userspace workflows.

Valid operator ids use letters, digits, `.`, `_`, and `-`. Path separators and
empty values are rejected. The value is not a secret.

Example:

```sh
export CHR_OPERATOR_ID=adrian
```

A project may put the export in `.envrc` when direnv is available. Without
direnv, set the same variable in the shell or command runner.
