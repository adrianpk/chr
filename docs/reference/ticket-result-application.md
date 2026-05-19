# Ticket Result Application

Ticket result application turns one execution attempt into ticket-facing state.

The model is side-effect free. It does not mutate Markdown, update SQLite, or
write report artifacts. Later writers consume the model to update durable state.

## Inputs

Result application requires:

- an execution record;
- an execution result envelope.

The execution record and result envelope must refer to the same execution ID and
ticket key.

## Ticket Outcome

The model maps execution state to ticket-facing status:

- succeeded execution with succeeded result: `reviewing`;
- any failed, blocked, or failed-result outcome: `blocked`.

Later writer slices are responsible for applying this status to ticket Markdown.

## Application Fields

The application contract carries:

- execution ID, project ID, ticket key, and ticket path;
- target ticket status;
- execution status, failure kind, and failure reason;
- summary;
- generated comment author and body;
- validation command and artifact paths;
- changed-file summaries;
- blocked reason details.

## Evidence

Evidence records the execution artifact references already captured by single
ticket execution:

- result path;
- runtime log path;
- validation path;
- validation command.

## Blocked Reason

Blocked applications carry the failure kind, failure reason, and diagnostic
artifact paths. Successful reviewing applications leave blocked reason fields
empty.
