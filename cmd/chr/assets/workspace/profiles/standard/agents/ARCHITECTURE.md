# Architecture

This project is managed by chr.

## Repository Shape

Prefer a conventional Go repository layout:

```text
cmd/          executable entrypoints
internal/     private application code
pkg/          public packages only when external reuse is intentional
docs/         durable project documentation
AGENTS.md     root agent entrypoint
```

chr stores its own project-local workspace under `.chr/`. Treat that
directory as tool-managed project state, not as part of the application's
package layout.

## Boundaries

- Keep command packages thin.
- Put application behavior in `internal/` by default.
- Use `pkg/` only for stable external contracts.
- Keep IO, persistence, transport, and execution adapters separate from domain decisions.
