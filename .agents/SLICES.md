# Slice-Based Development Strategy

This document defines delivery strategy for non-trivial work.
Use the Slice Delivery playbook for the step-by-step execution procedure.

## 1. Definitions

- Task: high-level requested work unit
- Slice: smallest meaningful, testable behavior increment
- Commit: atomic versioning unit

## 1.1 Naming Convention (Mandatory)

Use one naming scheme consistently across plans, trackers, tickets, PR notes, and reports.

- Slice labels: `Slice 1`, `Slice 2`, `Slice 3`, ...
- Task labels: `T1.1`, `T1.2`, `T2.1`, ...
- Completion reporting: `Slice N of M completed`

Slice numbers are scoped to one slicing plan. They are not repository-global,
they are not PR numbers, and they are not inferred from previously merged PRs.

Each slicing plan must define its own ordered slice table beginning at
`Slice 1`. A later slicing starts again at `Slice 1`.

For example:

- Slicing A: `Slice 1`, `Slice 2`, ..., `Slice N`
- Slicing B: `Slice 1`, `Slice 2`, ..., `Slice N`

Do not mix alternative numbering styles for these entities:

- no alphabetic slice/task labels (`Slice A`, `Task B2`)
- no Roman numerals (`Slice IV`, `Task II.3`)
- no repository-global continuation (`Slice 9` after another slicing ended at `Slice 8`)

## 2. Decomposition rules

Use a single slice only when scope and integration risk are low.
Otherwise decompose into multiple behavior-first vertical slices.

Slices are not created conversationally or invented on the fly during implementation.
A proposed next slice becomes executable only after its plan and tracker exist in the repository and are committed to `dev`.

Before executing a slice, the slicing plan must explicitly identify:

- the slicing name;
- the full ordered slice list for that slicing;
- the active slice number and short name;
- the expected branch name;
- the expected PR title;
- whether any older PRs are prerequisites rather than part of the slicing.

If the plan and tracker do not agree on this mapping, stop and fix the planning
artifacts before opening or editing a branch or PR.

The normal discovery-to-delivery chain is:

1. dump and/or quiz when useful
2. draft and/or spec when useful
3. plan
4. tracker
5. one branch and one PR per slice
6. one conventional commit per tracker task
7. user review and merge
8. return to `dev` before the next slice

Any early phase may be skipped when the work is already sufficiently specified, but plan and tracker are mandatory before implementation.

## 3. Slice rules

Each slice must:
- deliver a functional increment
- be testable in isolation
- integrate without breaking existing behavior
- have a versioned plan before code changes begin
- have a versioned tracker before code changes begin

## 4. Commit strategy

- Keep commits cohesive and scoped.
- Do not mix unrelated slice concerns in one commit.
- Keep commit subjects aligned with slice task intent.
- Use conventional commits.
- For approved slice execution, use one commit per tracker task.

## 5. Integration

- Integrate slices in safe order through PRs targeting `dev`.
- Keep repository passing checks after each integrated slice.
- Do not integrate slices directly into `main`.
- Align `dev -> main` only through the release-alignment playbook.

## 6. Traceable execution

For approved multi-slice work:
1. Commit the slice plan and tracker to `dev`.
2. Create a dedicated branch from `dev`: `feat/slice-<n>-<slug>` or another approved type prefix, where `<n>` comes from the active slicing plan.
3. Implement only that slice's tracker tasks.
4. Keep a one-to-one mapping between task IDs and commits.
5. Open one PR scoped to that slice.
6. Never open PRs per task or per commit.
7. Update trackers with tasks, commits, and validation outcomes.
8. Complete user review before merge.
9. Merge the approved slice PR into `dev`.
10. Return to `dev`, sync, and validate before the next slice.

## 7. Execution gate

For non-trivial work:
- approve slice plan first
- execute only approved slices
- capture extra ideas as backlog candidates
- do not implement a slice that exists only as a suggestion in conversation
- do not assign or rewrite slice numbers from the PR list
- do not open or rename a PR until the expected PR title is present in the plan or tracker
