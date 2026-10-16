# Telegram Record Operations

The receiver exposes deterministic record operations through the central
application service and the existing workflow and use-case metadata owners.
`/help` lists the commands and their scope options.

## Commands

`/tickets`, `/ticket`, `/slices`, `/slice-sets` and `/use-cases` accept an action
followed by strict JSON. The default action is `list`. Unknown fields and
trailing JSON values are rejected.

```text
/tickets list {"limit":20,"offset":0}
/ticket show {"identity":"PRJ-1"}
/tickets create {"title":"A task","synopsis":"A useful outcome","body":"Requirements"}
/ticket update {"identity":"PRJ-1","revision":1,"source_hash":"<observed full hash>","title":"Updated title"}
/use-cases create {"title":"Example case","synopsis":"Description","context":"Context"}
```

Scope options precede the action:

```text
/tickets --project example-project --namespace default list {"limit":20}
/use-cases --conversation <session-id> list
```

An exact reply or explicit conversation retains its project, physical registry
root and namespace. A conflicting explicit scope fails. Otherwise an explicit
project, selected peer project or unambiguous default resolves through the
current authorized registry. A mutation freezes this scope before preparing its
owner input. Changing `OPS_USER` during restart cannot redirect that request.
Current actor mapping, project enablement, root binding and any observed
conversation revision are checked again before application.

## Record authority and guards

Native tickets require configured ticket keys. Creation sets the normal pending
lifecycle and does not start work. Ticket display updates preserve their body
requirements. Slice sets are proposed plan records; creation requires the
existing plan reference, full plan hash and explicit closing validation policy.

Slices belong to a slice set. List/show may select `parent_id`. Slice creation
requires that exact proposed parent, its observed revision/source hash and a
slice descriptor with ID, order, title and contract reference/hash. The owner
checks the real contract and task membership at application. Display updates
require a proposed parent and pending slice. Lifecycle transitions, execution,
review and integration use their dedicated owners.

Native updates require both the positive observed `revision` and full
`source_hash`. Use-case updates require the exact slug in `identity` and full
`source_hash`; they preserve omitted context. Unsupported owner operations and
stale observations return diagnostics rather than inferred success.

Lists contain at most 50 records, defaulting to 20, with offsets bounded by
4,096. List results omit record bodies and expose `has_more` and `next_offset`.
Show returns one exact record and its observed guards and canonical reference.

## Retained application and delivery

A source/ordinal-derived request ID permanently binds the configured operator,
operation, normalized typed arguments and frozen target. Preparation retains
serialized owner input and its checksum before an application-intent checkpoint.
No SQLite transaction spans the project-owner call. A successful owner result
then commits its original receipt reference and bounded result facts with the
source acknowledgement and outbox in one SQLite transaction.

Restart after a lost owner acknowledgement observes that same request and
owner receipt. It does not prepare against a newly observed record. A conflicting
owner image blocks the request; another error preserves application intent for
recovery. Confirmed and rejected payloads remain terminal. A new instruction has
a new source identity.

Outbox parts are ordered and retained independently of send acknowledgement.
Each plain-text part fits 3,900 Unicode code points and UTF-16 units. Multi-part
output carries the same reference; output beyond eight parts becomes a bounded
summary with a canonical or paged reference. An uncertain retry says that the
notice may repeat. Its persisted attempt count and original application request
remain unchanged by notification retries.
