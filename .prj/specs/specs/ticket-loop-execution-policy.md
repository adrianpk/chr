# Ticket Loop Execution Policy

Status: initial

Sources:

- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-loop.md`

## Purpose

Define policy for selecting, ordering, executing, and stopping ticket loops.

The loop is not arbitrary agent autonomy. It is metadata-driven execution over
eligible tickets.

## Loop Modes

Supported conceptual modes:

- single ticket;
- simple loop;
- compound filtered loop;
- dry-run loop;
- plan or slice-oriented loop;
- resume loop.

The current implementation supports foreground loop execution over one managed
project. Future work may add richer modes.

## Selection

Loop selection may filter by:

- status;
- priority;
- severity;
- tags;
- project area;
- ticket type;
- use case.

Filters inside one dimension are alternatives unless otherwise specified. Tag
filters may require all requested tags.

## Ordering

Ordering must be deterministic.

Candidate ordering dimensions:

- ticket key;
- created time;
- priority;
- severity;
- combined priority and severity;
- explicit plan or slice order.

Tie-breakers must eventually fall back to ticket key.

## Execution

Each selected ticket is executed through the same single-ticket execution path.

The loop should not duplicate:

- runtime validation;
- execution record creation;
- harness invocation;
- result capture;
- ticket result application.

## Stop Conditions

The loop must stop when:

- there are no eligible tickets;
- `--max-iterations` is reached;
- a ticket execution fails;
- a ticket result is blocked;
- duplicate ticket keys are encountered in one selected run;
- stop is requested by daemon coordination.

Skipped tickets are eligible tickets not attempted because a stop condition
ended the run first.

## Blocked Tickets

Blocked tickets should stop the run unless policy explicitly allows continuing.

The blocker reason should be preserved in the ticket and in loop evidence.

## Checklist Handling

Checklist items may provide execution hints and completion evidence.

The system should avoid blindly marking checklist items complete unless the
execution result or validation evidence supports the change.

## Resume

A resume mode should use durable loop run and execution records to avoid
repeating completed attempts unnecessarily.

## Open Questions

- What query language is enough for compound loop selection?
- Should blocked tickets always stop the loop?
- Should loop continuation policy differ for failures versus blockers?
- How should plan or slice order map onto ticket order?
