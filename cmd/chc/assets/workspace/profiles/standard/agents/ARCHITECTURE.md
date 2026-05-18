# Architecture

This project is managed by ChrysoCoord.

## Repository Shape

Prefer a conventional Go repository layout:

```text
cmd/          executable entrypoints
internal/     private application code
pkg/          public packages only when external reuse is intentional
docs/         durable project documentation
.chc/         ChrysoCoord project-local state
AGENTS.md     root agent entrypoint
```

## Boundaries

- Keep command packages thin.
- Put application behavior in `internal/` by default.
- Use `pkg/` only for stable external contracts.
- Keep IO, persistence, transport, and execution adapters separate from domain decisions.
- Keep generated or operational ChrysoCoord state under `.chc/`.
