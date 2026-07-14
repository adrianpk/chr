# Ticket Reconciliation

## Purpose

Ticket reconciliation is the reusable boundary that aligns one project's derived
ticket index with canonical Markdown ticket files before execution code selects
or mutates tickets.

Future single-ticket execution, ticket loops, and daemon flows should call this
boundary before they rely on indexed ticket state.

## Go Contract

The core contract lives in `internal/core`:

```go
type TicketReconciler interface {
    ReconcileProjectTickets(ctx context.Context, opts ReconcileProjectTicketsOptions) (TicketReconciliationResult, error)
    ReconcileProjectTicket(ctx context.Context, opts ReconcileProjectTicketOptions) (TicketReconciliationResult, error)
}
```

Create a reconciler from an opened `TicketIndex`:

```go
reconciler := chr.NewTicketReconciler(index)
```

## Options

```go
type ReconcileProjectTicketsOptions struct {
    ProjectID    string
    ProjectDir   string
    ReconciledAt time.Time
}

type ReconcileProjectTicketOptions struct {
    ProjectID    string
    ProjectDir   string
    TicketPath   string
    ReconciledAt time.Time
}
```

- `ProjectID` identifies the registered project in the user-level index.
- `ProjectDir` points to the project root that owns `.chr/default/tickets`.
- `TicketPath` points to one affected Markdown ticket file.
- `ReconciledAt` records the refresh time used for indexed rows.

Use `ReconcileProjectTickets` when a full project refresh is required. Use
`ReconcileProjectTicket` after a result application mutates one known ticket.

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

For single-ticket reconciliation, parse errors are returned as blockers instead
of hard errors. Storage failures remain hard errors.

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

Manual `chr index rebuild` and `chr index status` remain maintenance and
diagnostic commands. They are not the execution precondition.

## Result Application Consistency

Result application writers mutate canonical Markdown ticket files. After those
writes complete, callers should reconcile the affected ticket before reading the
derived SQLite index again:

```go
result, err := reconciler.ReconcileProjectTicket(ctx, chr.ReconcileProjectTicketOptions{
    ProjectID:    application.ProjectID,
    ProjectDir:   projectDir,
    TicketPath:   application.TicketPath,
    ReconciledAt: time.Now().UTC(),
})
```

If `result.HasBlockers()` is true, the caller should surface the blocker and
avoid treating the index as current for that ticket.
