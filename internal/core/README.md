# `internal/core`

## Purpose

`internal/core` contains the core chr application behavior.

## Boundary

The package owns workspace configuration, project metadata, ticket workflows,
authoring artifacts, quizzes, dumps, specs, plans, persistence repositories,
workspace indexes, gate validation, agent execution records, and release
evidence models.

Adapters such as `internal/app` and `internal/tui` may call this package,
but this package should not import UI, CLI, or transport adapters.

Agent callers depend on `internal/harness` contracts. App injects the concrete
generation runner and worker command adapter. Core capability owners retain prompts, schemas and semantic validation.
`core/harness` owns binding/profile/selection identity. Its `host`, `runtime`,
`daemon` and `trace` children own retained generation execution, runtime plans,
lifecycle state and bounded tracing respectively. Native configuration and
workflow composition retain their actual revision, authorization and execution
authority; consumers import the capability owners directly.

Ticket and indexed records, document rendering/parsing and field/status rules
live in `core/ticket`. Native `WorkflowTicketIndex` composes ticket storage with
workflow projections on one SQLite handle; source/cache authority and guarded
ticket mutations remain with their actual native owners.

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
