# `internal/chcapp`

## Purpose

`internal/chcapp` is the application assembly layer for Chrysocoord commands
and the local TUI.

## Boundary

The package owns CLI command construction, stream wiring, command asset wiring,
and adapter-facing service operations. It delegates durable domain behavior,
workspace indexing, persistence, ticket management, generation, and validation
to `internal/chc`.

## API

The public package entrypoint is `Run`, which accepts explicit process IO,
arguments, context, and embedded command assets. Exported service request,
result, and view types form the boundary between command/TUI adapters and the
application service methods.

## Behavior

Service operations load project context, call core `internal/chc` services, and
return view-oriented results for CLI or TUI rendering. The package should stay
adapter-oriented and avoid owning reusable domain workflows.

## Validation

Package changes should be validated with:

```sh
go test ./internal/chcapp
```
