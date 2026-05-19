# Slice Reporting Contract

This document defines persistent implementation reporting for sliced work.

## Purpose

Every completed slice must produce a persistent, human-readable implementation
report for:

- fast human review;
- long-term traceability;
- pattern validation;
- architecture auditing;
- future onboarding and context reconstruction;
- correlation between implementation decisions, plans, trackers, commits, and PRs.

The report does not replace the PR diff. It surfaces the semantically important
parts of the slice so reviewers do not need to reconstruct intent from raw
diffs alone.

Reports optimize for:

- signal over completeness;
- structural understanding over line-by-line detail;
- future readability over immediate convenience.

## Scope

This contract applies to sliced work derived from plans, specs, and trackers.

A report is required:

- once per completed slice;
- after implementation is finalized;
- before opening the slice PR;
- whether the slice was autonomous or supervised.

This contract does not apply by default to small ad-hoc operations committed
directly on `dev`.

## Artifact Format

The canonical report format is Markdown.

Markdown is required because it is easy to diff, search, version, link, and
later index semantically. It also supports embedded code excerpts when needed.

Generated HTML may exist as a secondary rendering, but Markdown remains the
source of truth.

## Artifact Location

Slice reports live under:

```text
.prj/reports/slices/
  <slicing-slug>/
    slice-<n>-<slice-slug>.md
```

Example:

```text
.prj/reports/slices/automatic-reconciliation/
  slice-7-reconciliation-execution-boundary.md
```

Each slice must have exactly one canonical report.

## Required Metadata

Each report starts with:

```md
# Slice N: <Slice Name>

Status: drafting | reviewing | delivered
Slicing: <Slicing Name>
Plan: `<path>`
Tracker: `<path>`
Branch: `<branch>`
PR: pending | `#<number>`
```

Status semantics:

- `drafting`: implementation is still in progress.
- `reviewing`: implementation is complete and the PR is open or ready.
- `delivered`: the PR was merged and the slice was closed in the tracker.

## Required Sections

Each report must include:

```md
## Purpose

## Delivered Behavior

## Implementation Notes

## Contracts Added or Changed

## Files of Interest

## Validation

## Risks and Follow-ups
```

Section intent:

- `Purpose`: why this slice exists in the slicing.
- `Delivered Behavior`: user-facing or caller-facing behavior now available.
- `Implementation Notes`: important design decisions, tradeoffs, and boundaries.
- `Contracts Added or Changed`: APIs, commands, schemas, file formats, config,
  lifecycle rules, or operational behavior.
- `Files of Interest`: high-signal paths only, not a full diff inventory.
- `Validation`: commands run and outcomes.
- `Risks and Follow-ups`: known residual risks, explicit non-goals, or follow-up
  candidates. Use `None recorded.` when empty.

## PR Relationship

The report must exist before opening the PR.

The PR body should mention the report in `Changes`, but must not paste the full
report.

The report should be updated after merge with the PR number and `delivered`
status as part of slice closure.

## Tracker Relationship

Trackers record task completion, commits, and validation.

Slice reports record semantic implementation context.

Do not duplicate the tracker table in the report. Link the tracker and summarize
the implementation meaning.

## Size Discipline

Reports should remain concise.

If a report becomes a line-by-line changelog, it is duplicating the diff.

If a report cannot be read quickly, split incidental detail out and keep the
slice report focused on the decisions and contracts needed for future readers.
