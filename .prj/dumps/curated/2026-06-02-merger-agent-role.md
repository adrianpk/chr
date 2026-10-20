# Merger Agent Role

Date: 2026-06-02
Raw source: `.prj/dumps/raw/2026-06-02-merger-agent-role-raw.md`

## Central Topic

Formalize a small `@merger` responsibility in the agent workflow after ticket
review.

## Summary

`@merger` is a proposed operational role, not a new quality-review phase. Its
job is to verify merge preconditions and execute integration after `@reviewer`
has passed a ticket pull request.

The value is mainly formal responsibility separation, auditability, and a future
extension point. It is acceptable that the same underlying model may operate
under different roles because the role boundary constrains what the agent is
allowed to decide.

## Key Concepts

- `@dev` produces branch and pull request work.
- `@reviewer` checks for findings and either passes review or returns the
  ticket to repair.
- `@merger` performs merge readiness checks and executes integration.
- `ready_for_merge` is the pickup status for the merger role.
- The merger role should be mechanical, concise, and low-noise.

## Decisions And Strong Preferences

- Do not make `@merger` a second or third code reviewer.
- Do not add a new Board column solely for merger failures.
- Do not run a separate quiz before specification; the requirement is clear.
- Use existing ticket state, especially `ready_for_merge`, as the workflow
  signal.
- Add local ticket comments only when the merge is blocked or needs attention.
- Keep successful merges quiet except for the status transition and metadata
  update.
- Treat the role as an audit boundary and future extension point.
- Treat a passed review as merge authorization, subject only to objective merge
  preflight failures.
- Do not send a passed ticket back to review or dev unless preflight exposes an
  objective blocker that requires that path.

## Expected Preflight

- Ticket status is `ready_for_merge`.
- Ticket has branch and pull request metadata.
- Pull request exists and is open.
- Pull request targets the expected base branch.
- Pull request is mergeable.
- Configured review policy is satisfied.
- No earlier repair ticket blocks the serial workflow.
- Merge strategy is configured and allowed.

## Handoff Rule

The `@reviewer -> @merger` handoff should be formal and deterministic:

```text
@reviewer pass
  -> ticket status: ready_for_merge
  -> review outcome metadata: passed
  -> pull request reference remains authoritative

@merger pickup
  -> selects ready_for_merge tickets
  -> verifies objective merge preflight
  -> merges the pull request
  -> marks the ticket done
```

The merger must not reinterpret a passed review. It can block only on objective
integration failures such as stale branch, wrong base, missing metadata,
closed pull request, merge conflict, missing credentials, or serial repair
blockers.

## Tensions And Tradeoffs

- Automerge could solve the mechanical part, but it would not produce the same
  explicit role boundary or audit trail.
- Since the same model may operate every role, the distinction is mostly
  protocol-level rather than intelligence-level.
- The role is useful only if kept narrow; expanding it into review would add
  latency and noise.

## Assumptions

- chr continues to operate serially with one active agent role in the
  repository at a time.
- Ticket metadata remains the source of truth for workflow state.
- Remote pull request providers remain the place for provider-native merge and
  review state.

## Derived Threads

- Merger role prompt and allowed actions.
- Merge preflight command or service operation.
- Ticket metadata for merge outcome and blocked merge reason.
- Optional audit report for merge decisions.

## Possible Downstream Artifacts

- Spec for merger role semantics and handoff.
- Ticket review workflow update including `@merger`.
- CLI command design for merge preflight and execution.
- Implementation plan and sliced tracker for merger workflow support.

## Open Questions

- Should `@merger` be optional per workspace configuration, like review mode,
  or should it be the default path when review mode is enabled?
- Should successful merges write explicit metadata beyond ticket status?
- Which provider-specific merge checks should be normalized first?
