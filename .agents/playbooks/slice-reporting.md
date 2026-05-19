# Slice Reporting

Use this playbook when executing any planned implementation slice.

## Objective

Create one persistent Markdown implementation report per slice.

The report gives reviewers and future maintainers a semantic summary of what
changed, why it matters, and which contracts were introduced or changed.

## Timing

1. Create or update the report while implementation context is fresh.
2. Before opening the PR, set report status to `reviewing`.
3. Include the report path in the PR `Changes` section.
4. After merge, update report status to `delivered` and record the PR number
   during slice closure on `dev`.

## Location

Use this path shape:

```text
.prj/reports/slices/<slicing-slug>/slice-<n>-<slice-slug>.md
```

The slicing slug and slice slug must match the committed slicing plan.

## Template

```md
# Slice N: <Slice Name>

Status: drafting
Slicing: <Slicing Name>
Plan: `<path>`
Tracker: `<path>`
Branch: `<branch>`
PR: pending

## Purpose

## Delivered Behavior

## Implementation Notes

## Contracts Added or Changed

## Files of Interest

## Validation

## Risks and Follow-ups
```

## Content Rules

- Keep the report in English.
- Prefer concise prose and short bullets.
- Link the plan and tracker.
- Mention high-signal files only.
- Record commands exactly as run in `Validation`.
- Do not paste raw diffs.
- Do not duplicate the tracker task table.
- Use `None recorded.` for empty risk or follow-up sections.

## PR Integration

The PR body keeps the standard sections:

```md
## Summary

## Changes

## Validation
```

In `Changes`, include the report path as an operational artifact:

```md
- .prj/reports/slices/<slicing-slug>/slice-<n>-<slice-slug>.md
  - records semantic implementation context for this slice
```

## Closure

After the PR is merged and `dev` is synced:

1. Set report status to `delivered`.
2. Replace `PR: pending` with `PR: #<number>`.
3. Ensure the tracker records the report path.
4. Commit the report closure with the normal slice closure planning update.

## HTML Rendering

HTML rendering is optional and derived.

Do not use HTML as the canonical artifact unless a later tool needs rendered
reports for browsing. Markdown remains the source of truth.
