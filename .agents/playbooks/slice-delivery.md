# Slice Delivery

## Slice Workflow

1. Evaluate single-slice vs multi-slice execution.
2. Define behavior-first slice boundaries.
3. Create and commit the slice plan and tracker on `dev`.
4. Create the slice branch from `dev`.
5. Implement and validate per slice.
6. Integrate incrementally while keeping repository healthy.

## Slice Creation Gate

1. A slice suggested in conversation is only a proposal.
2. A slice becomes executable only when the repository contains a committed plan and tracker for that slice.
3. Do not write runtime code for a proposed slice before the plan and tracker are committed to `dev`.
4. If there is no plan/tracker for the next slice, stop at planning and create those artifacts first.
5. Do not present an unplanned suggestion as if it were already part of the delivery plan.

## Naming Convention

1. Use `Slice 1`, `Slice 2`, `Slice 3`, ... for slice identifiers.
2. Use `T1.1`, `T1.2`, `T2.1`, ... for task identifiers.
3. Do not use letter-based or Roman numeral variants for slices/tasks.

## PR Title

- Required: `feat(slice-<n>): <short slice name>`

## PR Body

```md
## Summary

- <functional increment 1>
- <functional increment 2>

## Changes

- <path/to/file>
  - <change detail>
- <path/to/other_file>
  - <change detail>

## Validation

- <command and result>
```

## Ordering Rules

1. `Summary`: primary behavior change first.
2. `Changes`: runtime/library code, tests, docs, ops artifacts.
3. `Validation`: targeted to broad.
4. Report validation outcomes as `pass`/`fail`.

## Traceability Contract

- Keep explicit correlation across:
  - plan artifact
  - tracker artifact
  - branch name and PR title
  - commit list mapped to slice tasks

## Blocker Handling

1. Capture failing command and exact error.
2. Attempt least-invasive compliant alternative.
3. Report blocker and stop risky changes.

## Completion Checklist

1. Scope implemented.
2. Relevant checks run.
3. Tracker/plan status updated.
4. Final status reported with concrete outcomes.

## Post-Slice Tuning

1. After planned slices and automated checks pass, decide whether tuning is needed.
2. Tuning slices are recommended, not mandatory.
3. If needed, run one or more short tuning slices (UX polish, manual verification findings, integration adjustments).
4. Keep the same delivery discipline: one branch + one PR per tuning slice, base `dev`.
5. Record the decision (`tuning needed: yes/no`) and evidence in the tracker.

## Main Alignment

1. Slice PRs target `dev`, not `main`.
2. Never work on `main` directly.
3. Align `dev -> main` only through the release-alignment playbook.
4. The `main` update is a PR merge using rebase + fast-forward from `dev`.

## Mandatory Execution Rules

1. Execute one slice at a time.
2. Commit plan and tracker before implementation.
3. Implement tasks in order.
4. Create one conventional commit per task.
5. Open one PR per slice.
6. Do not open PRs per task or per commit.
7. Complete user review.
8. Merge the approved slice PR into `dev`.
9. Return to `dev` before continuing.
10. Report implementation and validation status.
11. Start next slice only after previous slice is fully closed.
