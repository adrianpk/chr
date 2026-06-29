# Ralph Loop

Status: initial

Sources:

- `/home/adrian/Downloads/chr_ticket_execution_context_and_ralph_loop.md`

## Purpose

Define the controlled retry loop used when an agent execution fails validation
or produces incomplete work.

The Ralph Loop is not an infinite retry mechanism. It is a bounded pipeline that
uses failures as additional context and decides whether another attempt is
worthwhile.

## Loop Inputs

Each iteration may use:

- current ticket context;
- previous agent output;
- changed files;
- validation command output;
- runtime errors;
- logs;
- code graph context;
- project protocol documents;
- previous blocker comments.

## Iteration Flow

One iteration should:

1. Run the configured validation command or validation set.
2. Capture stdout, stderr, exit status, and relevant artifacts.
3. Classify the failure.
4. Decide whether retry is allowed.
5. Build a focused repair context.
6. Invoke the agent with the repair context.
7. Re-run validation.
8. Record evidence for the attempt.

## Failure Classes

Useful failure classes include:

- test failure;
- lint failure;
- compile failure;
- formatting failure;
- missing tool;
- runtime container failure;
- secret or credential failure;
- ambiguous requirement;
- scope conflict;
- repeated unchanged failure.

Only failures that provide actionable information should trigger another
automated attempt.

## Stop Conditions

The Ralph Loop must stop when:

- validation passes;
- max attempts is reached;
- the same failure repeats without meaningful diff;
- required tools are unavailable;
- secrets are missing;
- the ticket is ambiguous enough to require human input;
- the agent tries to expand beyond ticket scope;
- the workspace becomes unsafe to continue automatically.

## Scope Discipline

Each retry must stay inside the ticket scope.

The repair context should emphasize:

- the exact failure;
- the changed files;
- the relevant code paths;
- the validation command;
- the expected minimal correction.

It should not invite unrelated refactors.

## Evidence

Each iteration should preserve:

- attempt number;
- validation commands;
- validation outcomes;
- failure classification;
- changed file summary;
- whether retry was allowed;
- final stop reason.

## Human Intervention

Human intervention is not a failure.

The loop should request or record human input when it lacks enough reliable
context to proceed safely.

## Open Questions

- What default max-attempt value should be used?
- Should max attempts vary by failure class?
- Which failures should be considered non-retryable by default?
- How much validation output should be embedded in the next agent prompt?
