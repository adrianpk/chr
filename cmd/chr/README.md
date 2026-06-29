# `cmd/chr`

## Purpose

`cmd/chr` is the executable assembly point for the Chryso command-line
application.

## Boundary

The package embeds command assets, wires process IO, and delegates application
execution to `internal/app`. It should not contain domain, persistence, TUI,
or workflow logic.

## Behavior

`main` passes command arguments, standard streams, and embedded assets into
`app.Run`. Errors are written to standard error and terminate the process
with a non-zero status.
