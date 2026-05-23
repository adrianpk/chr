# Go Conventions

This document defines coding rules for Go application code.

## Dependencies

- Prefer the Go standard library before adding dependencies.
- Ask before introducing external modules, services, generators, or build tools.
- Keep dependency additions tied to a concrete package responsibility.
- Avoid dependencies that replace small, local design decisions.

## Package boundaries

- Default to `internal/` for application code.
- Use `pkg/` only for stable contracts that are intentionally shared outside the module.
- Keep command packages under `cmd/` as assembly entrypoints.
- Keep transport adapters thin and delegate execution to package-owned services or workflows.
- Keep reusable CLI, TUI, and future transport operations centered in the
  `internal/chcapp` operation service.
- Do not duplicate an operation workflow inside an adapter when the behavior can
  be invoked from another adapter.
- Split operation service methods across files when it improves readability,
  but do not create one service implementation per entity by default.
- Keep `service*.go` files adapter-neutral. TUI-specific mappings belong in
  TUI adapter files, not in the service implementation.
- Use package-local unexported types unless a public contract is required.

## Code style

- Keep functions short and explicit.
- Return errors with useful context and avoid swallowing failure causes.
- Prefer small interfaces at package boundaries where substitution is needed.
- Avoid global mutable state except for constants or immutable package-level configuration.
- Use comments sparingly: Godoc for exported symbols when useful, and brief explanations for non-obvious logic.
- Prefer clear data flow over clever generic composition.

## Constructors and lifecycle

- Constructor functions use `New...` naming.
- Constructors should store dependencies and configuration; they should not perform failable IO.
- Put failable startup, readiness, and shutdown behavior in explicit methods such as `Start`, `Setup`, or `Stop` when the component lifecycle matters.
- Keep constructor parameter order stable within a package.

## Testing

- Use Go's standard `testing` package by default.
- Prefer table-driven tests for input/output behavior.
- Place tests beside the package they validate.
- Use fakes over broad mocks when dependency substitution is needed.
- Cover expected failures, edge cases, and package boundary contracts.
- Do not skip tests for behavior changes.

## Anti-patterns

Avoid:
- command packages that contain domain logic
- TUI, CLI, Telegram, or email adapters that own reusable workflow logic
- separate service implementations per entity when one operation service with
  file-level method grouping is sufficient
- generic `util` packages without clear ownership
- public packages created for internal convenience
- stringly typed contracts where structs or explicit types are available
- hidden global state
- broad rewrites that do not reduce current complexity

Prefer:
- explicit package contracts
- small cohesive packages
- standard library first
- deterministic tests
- local consistency over generic architecture
