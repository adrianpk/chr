# `internal/core`

## Purpose

`internal/core` contains the core Chryso application behavior.

## Boundary

The package owns workspace configuration, project metadata, ticket workflows,
authoring artifacts, quizzes, dumps, specs, plans, persistence repositories,
workspace indexes, gate validation, agent execution records, and release
evidence models.

Adapters such as `internal/app` and `internal/tui` may call this package,
but this package should not import UI, CLI, or transport adapters.

## API

Exported types are internal application contracts shared across adapters and
package-local services. Constructors generally use `New...`, validation helpers
use `Validate...`, and render/parse/decode helpers expose stable document or
payload boundaries.

## Behavior

Services in this package operate on explicit project directories, user
environment paths, repositories, and configured assets. Persistence-backed
repositories expose closeable interfaces where callers must release database
resources.

## Validation

Package changes should be validated with:

```sh
go test ./internal/core
```
