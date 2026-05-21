# Slicing Workflow

Slicing is the practice of breaking larger implementation work into small,
reviewable units.

ChrysoCoord uses slicing so the user can understand how a feature will be built
before the agent starts making changes. A good slice is small enough to review
and validate, but large enough to move the system forward.

## Flow

The normal flow is:

1. Describe a use case or feature.
2. Produce or refine a specification.
3. Promote the specification into a plan.
4. Create a tracker for execution.
5. Divide the work into slices.
6. Define tasks inside each slice.
7. Associate tasks with commits where useful.
8. Associate slices with branches and pull requests.
9. Review and merge each pull request manually.
10. Use auto-merge only in trusted workflows where that policy is explicit.

The plan and tracker should make the expected work visible. The user should be
able to inspect how many slices are expected, which tasks belong to each slice,
and how implementation will be distributed across commits and pull requests.

## Slices and Tasks

A slice groups related implementation tasks under one reviewable goal.

Tasks are the smaller pieces of work inside a slice. In ChrysoCoord's delivery
style, tasks can correspond to commits. That makes the implementation trail
easier to audit: a tracker can show what was planned, and the commit history can
show how that work landed.

## Branches and Pull Requests

Slices usually map to dedicated branches and pull requests. The pull request is
the review surface for that slice.

Manual review remains the normal path. Auto-merge can exist in trusted setups,
but it should be treated as an explicit workflow policy rather than the default
assumption.

## Relationship to Tickets

Tickets are the agent's task surface during execution. Slicing is the planning
and delivery structure around larger work. A ticket can represent a slice, a
task, or another bounded unit depending on the project workflow.
