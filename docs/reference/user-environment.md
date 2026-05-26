# User Environment

## Purpose

The user environment resolves user-level `chc` paths.

These paths are separate from project-local `.chc/` directories. Project workspaces contain canonical project files such as Markdown tickets. User-level paths contain local tool state such as future project registry data, global indexes, caches, and daemon state.

Project-local active userspace is separate from these user-level paths. It
lives at `.chc/user/` in a managed project and uses `CHC_OPERATOR_ID` to bind
operator-owned workflows to a local operator id.

## Normal Mode

Normal mode resolves paths from XDG-style directories when they are provided.

```text
config -> $XDG_CONFIG_HOME/chc
data   -> $XDG_DATA_HOME/chc
cache  -> $XDG_CACHE_HOME/chc
state  -> $XDG_STATE_HOME/chc
```

When XDG directories are not provided, paths fall back to the user's home directory.

```text
config -> ~/.config/chc
data   -> ~/.local/share/chc
cache  -> ~/.cache/chc
state  -> ~/.local/state/chc
```

## Development Mode

Development mode redirects all user-level paths into a caller-supplied sandbox root.

```text
config -> <dev-root>/config/chc
data   -> <dev-root>/data/chc
cache  -> <dev-root>/cache/chc
state  -> <dev-root>/state/chc
```

Development mode is intended to keep work on `chc` from writing to real user configuration or data directories.

## Construction

Core code uses `chc.NewUserEnvironment(...)` to construct a `chc.UserEnvironment`.

Construction resolves paths only. It does not create directories or write files.

## Active Operator

`CHC_OPERATOR_ID` identifies the active local operator for project-local
userspace workflows.

Valid operator ids use letters, digits, `.`, `_`, and `-`. Path separators and
empty values are rejected. The value is not a secret.

Example:

```sh
export CHC_OPERATOR_ID=adrian
```

A project may put the export in `.envrc` when direnv is available. Without
direnv, set the same variable in the shell or command runner.
