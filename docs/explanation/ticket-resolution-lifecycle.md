# Ticket Resolution Lifecycle

Ticket resolution is the path from a Markdown ticket to an attempted or
completed change in the project.

## Lifecycle

1. A ticket describes the work and carries metadata used for filtering.
2. The user selects the ticket directly or lets a loop select an eligible
   ticket.
3. chr creates an execution record for the attempt.
4. The host process composes the execution context.
5. The sandboxed runtime starts the harness.
6. The harness works on the ticket and runs validation.
7. The harness writes result data and evidence.
8. chr applies the result back to the ticket when requested by the
   workflow.
9. The ticket records status, checklist progress, comments, blockers, and
   handoff notes.

## Blockers and Thresholds

The agent should not loop forever on the same failing problem.

When repeated attempts fail, chr can preserve failure evidence and stop
or escalate according to configured thresholds. The exact notification transport
is separate from ticket resolution. Future integrations may send blocker or
status information through Telegram or email.

## Evidence

The lifecycle is designed around durable evidence:

- ticket metadata and comments;
- checklists;
- execution records;
- runtime and validation logs;
- result artifacts;
- changed-file summaries;
- previous-attempt context for later retries.

This makes the state of the work inspectable after the agent stops.
