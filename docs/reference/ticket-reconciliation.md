# Ticket Reconciliation

## Purpose

Ticket reconciliation is the reusable boundary that aligns one project's derived
ticket index with canonical Markdown ticket files before execution code selects
or mutates tickets.

Future single-ticket execution, ticket loops, and daemon flows should call this
boundary before they rely on indexed ticket state.

## Go Contract

The core contract lives in `internal/chc`:

```go
type TicketReconciler interface {
    ReconcileProjectTickets(ctx context.Context, opts ReconcileProjectTicketsOptions) (TicketReconciliationResult, error)
}
```

Create a reconciler from an opened `TicketIndex`:

```go
reconciler := chc.NewTicketReconciler(index)
```

## Options

```go
type ReconcileProjectTicketsOptions struct {
    ProjectID    string
    ProjectDir   string
    ReconciledAt time.Time
}
```

- `ProjectID` identifies the registered project in the user-level index.
- `ProjectDir` points to the project root that owns `.chc/tickets`.
- `ReconciledAt` records the refresh time used for indexed rows.

## Result

```go
type TicketReconciliationResult struct {
    IndexedTickets int
    Blockers       []TicketReconciliationBlocker
}
```

`IndexedTickets` is the number of ticket files successfully indexed during the
reconciliation pass.

`Blockers` contains conditions that should prevent ticket execution from
continuing.

Use `HasBlockers` and `FirstBlocker` when callers only need execution gating:

```go
if result.HasBlockers() {
    blocker, _ := result.FirstBlocker()
    return blocker
}
```

## Blockers

```go
type TicketReconciliationBlocker struct {
    Kind         TicketReconciliationBlockerKind
    Path         string
    RelativePath string
    Err          error
}
```

Supported blocker kinds:

- `parse_error`: a Markdown ticket file cannot be parsed or validated.

Parse-error blockers preserve the absolute path, relative ticket path, and
underlying parser error.

## Execution Precondition

Execution code must reconcile before selecting work from the index.

If reconciliation returns blockers, execution should stop and surface the
blocker details to the user or notification channel.

Manual `chc index rebuild` and `chc index status` remain maintenance and
diagnostic commands. They are not the execution precondition.
