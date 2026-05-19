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

## Ticket State Writer

The ticket result state writer applies only the ticket status from a result
application.

It updates ticket frontmatter:

- `status`;
- `updated_at`.

It preserves existing ticket metadata, body sections, comments, blocker text,
and validation text.

The writer maps the already-computed application status directly:

- `reviewing` for successful review-ready applications;
- `blocked` for failed or blocked applications.

Comment, validation evidence, changed-file summary, and blocked reason writing
are handled by later slices.

## Execution Comment Writer

The execution comment writer appends one comment to the ticket `# Comments`
section using the result application comment payload.

The generated comment includes:

- execution ID;
- execution status;
- result summary;
- result artifact path;
- runtime log artifact path;
- validation artifact path;
- failure reason when present.

The writer preserves existing comments and ticket body sections. It also updates
`updated_at` through the existing ticket comment writer.

Validation evidence, changed-file summaries, and blocked reason section content
are still handled by later slices.

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
