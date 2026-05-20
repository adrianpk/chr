# Ticket Loop

## Purpose

The ticket loop is a foreground runner that processes eligible tickets from one
managed project.

The first loop boundary is ticket selection. Execution, result application,
status reporting, and daemon coordination are separate concerns.

## Eligible Ticket Query

The eligible ticket query reads from the existing ticket index. It does not
scan ticket files, refresh the index, update ticket state, create execution
records, or mutate the workspace.

Required input:

- project ID.

Optional input:

- eligible ticket statuses.

When no statuses are provided, the default eligible status is:

```text
ready
```

Callers may provide explicit statuses to widen or narrow selection. Status
values must be valid ticket lifecycle statuses.

Returned tickets are sorted by ticket key so repeated reads are deterministic.

## Scope

The eligible ticket query is a core package boundary. CLI parsing, metadata
filters, ordering policies, loop run persistence, repeated execution, stop
conditions, and reporting are implemented by later slices.
