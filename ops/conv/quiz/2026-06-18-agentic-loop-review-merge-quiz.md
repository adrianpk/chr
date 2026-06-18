# Agentic Loop Review and Merge Quiz

Status: complete

Date: 2026-06-18

Topic: Agentic loop review and merge workflow.

Raw capture: `ops/conv/quiz/2026-06-18-agentic-loop-review-merge-quiz-raw.md`

Spec: `ops/specs/2026-06-18-agentic-loop-review-merge.md`

## Summary

The default ChrysoCoord loop should be a steady sequential cycle that develops a
ticket, creates pull request evidence, validates the work, merges when policy
allows it, and then continues to the next ticket.

The default path is automatic merge. The more formal reviewer/merger role cycle
is a supported configurable alternative, not the default. That formal mode should
be documented clearly and tested heavily because some projects will prefer a
traditional review workflow.

## Decisions

- The first integration unit is the ticket.
- One ticket maps to one branch, and one branch maps to one pull request.
- Slice naming may appear in ticket, branch, or pull request names, but slice is
  not a separate integration unit in the first workflow.
- Automatic merge is the default method for new projects.
- Pull requests remain useful as inspectable evidence even when automatic merge
  is enabled.
- The default automatic path does not launch separate reviewer and merger agent
  roles.
- The formal reviewer/merger role cycle is configurable and supported.
- If review requests changes, the same ticket, branch, and pull request are
  reused.
- The default maximum review repair allowance is one cycle.
- A blocked ticket stops the automatic loop; ChrysoCoord should not skip it and
  continue later slicing work.
- The merger role should not re-run validation only to add ceremony.

## Validation Direction

The immediate enforcement mechanism is the project `validation_command`.

ChrysoCoord should influence the initial project baseline toward the strongest
practical validation command the project can support. Depending on stack and
context, that may include:

- unit tests;
- table-driven tests;
- fuzz or property-style tests;
- stronger quality checks such as CRAP or mutation-oriented checks when feasible;
- linting;
- runability checks.

Automatic merge is allowed only when implementation, tests, linting, validation,
and relevant runability evidence are satisfied.

## Formal Review Mode

The formal reviewer may run read-only validation after verifying the expected
branch, but must not mutate code created by the development role.

Findings should be local-first. Local ticket comments or metadata should stay
compact and point to review artifacts rather than embedding long review text
directly in the ticket.

Provider-backed review remains valuable for a more traditional workflow.
ChrysoCoord should aim to support:

- Forgejo through `tea`;
- GitHub through `gh`.

Provider comments, remote diffs, and `@` mentions are useful integrations, but
local ChrysoCoord artifacts remain the handoff record.

## Artifact Policy

Operational ChrysoCoord artifacts should live under `./.chc`.

Review artifact versioning should be configurable, with versioned review
artifacts as the default.

## Future Consideration

A git worktree model should be considered later. It would allow agent roles to
operate in isolated workspaces while the user remains free to interact with the
original repository.

## Open Questions

- Exact configuration key names for loop mode and automatic merge.
- Exact `.chc` path for review artifacts.
- Whether versioned review artifacts are committed automatically or only made
  available as versionable files.
- How much validation baseline generation belongs in `chc init` versus a
  separate project-hardening command.
- When provider comments should become default behavior for formal review mode.
