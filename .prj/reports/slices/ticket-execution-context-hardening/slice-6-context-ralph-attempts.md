# Slice 6: Ralph Attempt Context

Status: reviewing
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-6-context-ralph-attempts`
PR: #83

## Purpose

Feed previous attempt evidence into Ralph Loop retries through explicit
materialized artifacts instead of hidden conversation memory.

## Delivered Behavior

The repository now has previous-attempt context support that can:

- load the previous validation artifact when available;
- load the previous result artifact when available;
- include previous execution status, failure kind, failure reason, and
  validation command;
- include previous changed files reported by `result.json`;
- summarize validation excerpts and failure signatures;
- identify repeated failure signatures from validation output;
- render `previous-attempt.md`;
- render a reusable link section for the next main `context.md`;
- explicitly exclude implicit conversation history.

## Implementation Notes

Missing validation or result artifacts do not fail context construction. They
are recorded as skipped source inventory entries so a retry can still proceed
with the evidence that exists.

The renderer keeps validation output bounded. It includes excerpts and
signatures rather than dumping the entire validation log into the next prompt.

## Contracts Added or Changed

New internal contracts:

- `PreviousAttemptContextOptions`
- `PreviousAttemptContext`
- `LoadPreviousAttemptContext`
- `RenderPreviousAttemptContext`
- `RenderPreviousAttemptContextLink`

## Files of Interest

- `internal/chc/context_previous_attempt.go`
- `internal/chc/context_previous_attempt_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestLoadPreviousAttemptContext|TestRenderPreviousAttemptContextLink'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice does not yet decide retry eligibility or write `previous-attempt.md`
to disk. Docker-compatible materialization and readiness checks remain planned
for later slices.
