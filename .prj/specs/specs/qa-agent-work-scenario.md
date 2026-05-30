# QA Agent Work Scenario

Status: accepted

## Purpose

Define a manually reproducible QA scenario for validating the ChrysoCoord flow
from seeded project context through spec generation, plan generation, tracking
and ticket materialization, single-ticket execution, and agent-work loop
execution.

The scenario exists to test ChrysoCoord behavior, not to be a complex product
sample. The reference project must be realistic enough to exercise project
structure, tests, documentation, `.chc` artifacts, and Git branching without
making the domain implementation the source of most failures.

## Workspace Layout

The QA workspace lives inside the `chc` repository under `qa/`.

```text
qa/
  ref/
    001/
      blog/
        ...
  project/
    blog/
        ...
  tmp/
    blog/
        ...
```

`qa/ref/001/blog` is versioned in the `chc` repository and is the canonical
reference input.

`qa/project/blog` is a generated active project copy used for manual testing.
It must be ignored by the `chc` repository.

`qa/tmp/blog` is an ignored temporary location used to move aside the previous
active project during reset. The reset flow may replace the previous
`qa/tmp/blog` state instead of keeping timestamped archives.

## Reset Contract

The reset operation must be destructive for the active project copy.

`make qa-reset-blog` must:

- remove or replace the previous `qa/tmp/blog`;
- move the existing `qa/project/blog` to `qa/tmp/blog` when it exists;
- copy `qa/ref/001/blog` to `qa/project/blog`;
- initialize `qa/project/blog` as an independent Git repository;
- create a clean base commit in that repository;
- leave `qa/project/blog` ready for `chc` to run from inside it.

The active project Git repository must not interfere with the parent `chc`
repository. Scripts must use explicit paths and `git -C qa/project/blog` when
operating on the active project repository.

## Make Targets

The repository should expose generic QA targets and convenience aliases.

Required targets:

- `qa-reset` with a scenario parameter;
- `qa-test` with a scenario parameter;
- `qa-reset-blog` as the blog-specific alias;
- `qa-test-blog` as the blog-specific validation alias.

The heavy shell logic should live under `scripts/qa/` when the Make targets
would otherwise become procedural.

## Reference Project

The first reference project is `blog`.

The project should be a small but realistic Go HTTP service using only the
standard library. It should use modern `http.ServeMux` method/path patterns
instead of a third-party router.

The code should look like the initial stage of a real project, with simple
internal package structure rather than a single-file toy implementation.

The baseline must include:

- blog post listing;
- single post retrieval by slug;
- an intentionally failing test for a known bug;
- no implemented comments feature;
- enough structure for a later comments implementation to be natural.

The seeded bug is:

- `GET /posts/{slug}` returns `500` for an unknown slug instead of `404`.

The seeded test suite must fail immediately after reset because of this bug.
`make qa-test-blog` must run `go test ./...` in the active project and expose
that expected failure.

## Seeded ChrysoCoord Context

The reference project must include a `.chc` tree that respects the current
project artifact layout. Generated or user-facing ChrysoCoord artifacts must
live under `.chc/user/...` except for root files that CHC requires for project
configuration or manifest compatibility.

Seeded context must include:

- project metadata;
- a `comments` use case;
- dumps describing the comments feature need;
- curated quiz material that clarifies the comments behavior;
- a manual bug ticket in `ready_for_dev` for the unknown-slug bug;
- no generated comments spec;
- no generated comments plan;
- no generated comments tracking document;
- no generated comments tickets.

The manual bug ticket must:

- be `ready_for_dev`;
- use type `bug`;
- use high priority and high severity;
- reference `go test ./...`;
- instruct the agent to observe the existing failing test before changing code.

## Project Documentation

Architecture decision records are project documentation, not hidden
ChrysoCoord-only context.

The blog reference project should use Diataxis documentation layout. ADRs should
live under:

```text
docs/explanation/adr/
```

At least one ADR should influence how the comments feature is specified,
planned, tracked, or implemented. Example decisions may cover in-memory storage,
HTTP error response shape, package boundaries, or initial comment ownership by
post slug.

The reference project should also include a concise `README.md` explaining the
manual QA flow without verbose product prose.

## Expected Manual Flow

A user should be able to run:

```sh
make qa-reset-blog
make qa-test-blog
```

Then from `qa/project/blog`:

1. Run `chc`.
2. Inspect the seeded bug ticket.
3. Start single-ticket execution for the bug ticket.
4. Validate that the test suite can become green after the ticket is resolved.
5. Generate a comments spec from seeded dumps, quiz notes, project metadata, and
   ADRs.
6. Generate a comments implementation plan from the spec.
7. Materialize plan tracking and tickets.
8. Run the agent-work loop for the generated comments tickets.
9. Validate the final project with `make qa-test-blog`.

The first, second, or later runs may fail. The scenario exists to make those
failures easy to reproduce, inspect, reset, and retry.

## Testing Preference

The reference project and ChrysoCoord itself should express a test-based
engineering preference:

- reproduce bugs with a failing test when feasible;
- use table-driven tests as the default style for input/output behavior;
- implement the smallest fix that makes the relevant tests pass;
- run and report focused validation commands;
- consider property, fuzz, or other advanced testing approaches only when the
  use case justifies them.

This preference should appear in the project agent conventions and should also
influence generated tracking and ticket content so agents receive validation
expectations inside the work units they execute.

## Non-Goals

The first scenario does not need:

- a second reference case for bugs without existing tests;
- comments moderation;
- a database;
- external dependencies;
- remote Git hosting;
- automatic execution of the full QA flow from one command;
- formal ADR entities in the CHC model;
- configurable ADR paths.

## Open Questions

- Whether future scenarios should cover bugs that require the agent to create
  the first failing test.
- Whether CHC should eventually initialize managed project Git repositories on
  demand.
- Whether larger blog variants should reuse the same reference project lineage
  or live as separate numbered scenarios.
