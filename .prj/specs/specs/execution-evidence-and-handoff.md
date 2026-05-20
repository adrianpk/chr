# Execution Evidence and Handoff

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-result-application.md`

## Purpose

Define the evidence and handoff artifacts expected after ticket execution.

The goal is for a human or future agent to understand what happened without
reconstructing the work from raw diffs alone.

## Evidence Types

Ticket execution may produce:

- execution record;
- runtime log;
- worker result artifact;
- validation artifact;
- changed file list;
- changed file summary;
- ticket comments;
- blocker reason;
- checklist updates;
- loop run summary;
- PR body or handoff summary.

## Result Contract

The worker result should be machine-readable enough for `chc` to apply ticket
updates.

It should capture:

- status;
- summary;
- changed files;
- validation commands;
- validation outcomes;
- blocker reason when blocked;
- follow-up notes when useful.

## Ticket Update Policy

Ticket state should be updated from execution evidence, not from optimistic
assumption.

Examples:

- successful validation may support moving a ticket to a completed state;
- missing context should write a blocker;
- partial work should record a comment or evidence without pretending the ticket
  is complete.

## Machine-Readable Comments

Ticket comments may eventually carry structured status blocks that humans can
read and `chc` can parse.

Useful machine-readable comment data includes:

- agent status;
- blocker reason;
- validation summary;
- changed file summary;
- execution ID;
- artifact references;
- next required human decision.

Machine-readable comments should remain Markdown-friendly and should not replace
the execution record.

## PR Handoff

PR handoff should summarize:

- what changed;
- why it changed;
- how it was validated;
- where detailed evidence lives.

For this repository, PR bodies should use exactly:

- `Summary`;
- `Changes`;
- `Validation`.

## Auditability

Evidence should connect:

- project ID;
- ticket key;
- execution ID;
- artifact paths;
- commit SHA;
- PR number when available.

## Noise Control

Evidence should avoid:

- full raw logs unless needed;
- excessive diffs;
- secrets;
- unrelated file listings;
- speculative conclusions.

## Open Questions

- Which evidence artifacts are mandatory for every execution?
- Should PR creation consume the same evidence model as ticket result
  application?
- How should validation evidence be summarized when commands are flaky?
- Should loop run evidence aggregate all ticket evidence or link to each
  execution record?
