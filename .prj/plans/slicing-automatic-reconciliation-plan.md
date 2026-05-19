# Automatic Reconciliation Slicing Plan

Status: active

## Purpose

Plan the `Automatic Reconciliation` slicing from the feature slicing set map.

This slicing covers the use case:

> A user can edit ticket files directly without remembering manual synchronization commands.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related documents:

- `.prj/drafts/chc-command-surfaces-and-priority.md`
- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`
- `docs/reference/ticket-index.md`
- `docs/reference/ticket-parser.md`
- `docs/reference/ticket-validation.md`

Prerequisite delivered work:

- `User Project Context` slicing.
- `Ticket Foundation` slicing.

## Current State

Already delivered:

- registered project persistence;
- active project context resolution;
- Markdown ticket parser;
- indexed ticket model;
- SQLite ticket index repository;
- ticket list, show, comment, status, and validation behavior.

Still missing:

- scanning project ticket files into the global index on demand;
- detecting stale, missing, deleted, or invalid indexed tickets;
- refreshing the index before user-facing ticket reads;
- maintenance commands for index status and rebuild;
- parse-error reporting for malformed ticket files;
- rehydration after cloning a project with `.chc/` files.

## Slicing Scope

This slicing includes:

- project ticket file discovery;
- ticket file fingerprint comparison against indexed rows;
- automatic refresh before ticket read commands;
- manual index status for debugging;
- manual index rebuild for recovery;
- parse-error reporting;
- cloned-project rehydration through existing project registration and rebuild behavior;
- preparation for future execution commands to request a fresh ticket index before selecting work.

## Out of Scope

This slicing does not include:

- filesystem watchers;
- daemon lifecycle;
- agent runtime setup;
- ticket execution;
- filtered execution loops;
- Telegram or email notifications.

Manual synchronization commands are maintenance and recovery tools. Normal ticket reads should not depend on the user remembering to run them.

## Slice Numbering

This slicing uses local slice numbering. The slice number is scoped to this
plan, starts at `Slice 1`, and is the number used in branch names, PR titles,
trackers, and reports for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Ticket File Discovery

Status: delivered.

Expected branch: `feat/slice-1-ticket-file-discovery`
Expected PR title: `feat(slice-1): ticket file discovery`

Goal:

Add a core primitive that discovers Markdown ticket files for one project.

Candidate scope:

- resolve the project ticket directory;
- walk nested use-case directories;
- return stable relative paths;
- ignore non-Markdown files;
- report missing ticket directories clearly.

Expected outcome:

- reconciliation code can enumerate canonical ticket files without duplicating filesystem walking logic.

Delivery:

- Repository PR: `#22`
- PR title: `feat(slice-1): ticket file discovery`

### Slice 2: Ticket Index Rebuild

Status: delivered.

Expected branch: `feat/slice-2-ticket-index-rebuild`
Expected PR title: `feat(slice-2): ticket index rebuild`

Goal:

Rebuild one project's ticket index from Markdown ticket files.

Candidate scope:

- combine ticket discovery, parsing, and ticket index repository writes;
- replace the indexed ticket set for one project;
- preserve parser validation errors as structured rebuild results;
- add focused tests for nested tickets and deleted indexed rows.

Expected outcome:

- a project can rehydrate its derived ticket index from `.chc/tickets/*.md`.

Delivery:

- Repository PR: `#23`
- PR title: `feat(slice-2): ticket index rebuild`

### Slice 3: Index Status

Status: delivered.

Expected branch: `feat/slice-3-index-status`
Expected PR title: `feat(slice-3): index status`

Goal:

Report whether the indexed ticket rows match canonical Markdown files.

Candidate scope:

- compare indexed rows with discovered files;
- detect missing index rows;
- detect stale rows using fingerprint fields;
- detect orphaned index rows whose files were deleted;
- surface parse errors separately from stale rows.

Expected outcome:

- a user or agent can inspect index health without mutating state.

Delivery:

- Repository PR: `#24`
- PR title: `feat(slice-3): index status`

### Slice 4: Index Rebuild CLI

Status: delivered.

Expected branch: `feat/slice-4-index-rebuild-cli`
Expected PR title: `feat(slice-4): index rebuild cli`

Goal:

Expose manual project index rebuild as a recovery command.

Candidate scope:

- add an index command group if needed;
- add `chc index rebuild`;
- resolve project context through the existing resolver;
- print concise rebuild results;
- document command behavior.

Expected outcome:

- a cloned or drifted project can rebuild derived ticket index state explicitly.

Delivery:

- Repository PR: `#25`
- PR title: `feat(slice-4): index rebuild cli`

### Slice 5: Index Status CLI

Status: delivered.

Expected branch: `feat/slice-5-index-status-cli`
Expected PR title: `feat(slice-5): index status cli`

Goal:

Expose index health inspection as a user-facing command.

Candidate scope:

- add `chc index status`;
- resolve project context through the existing resolver;
- print clean status rows or a summary;
- return non-zero when the index is stale or invalid;
- document command behavior.

Expected outcome:

- advanced users and agents can diagnose stale derived state without rebuilding blindly.

Delivery:

- Repository PR: `#26`
- PR title: `feat(slice-5): index status cli`

### Slice 6: Refresh Before Ticket Reads

Status: delivered.

Expected branch: `feat/slice-6-refresh-before-ticket-reads`
Expected PR title: `feat(slice-6): refresh before ticket reads`

Goal:

Automatically refresh the ticket index before read-oriented ticket commands.

Candidate scope:

- refresh before `ticket list`;
- refresh before `ticket show`;
- refresh before `ticket comment`;
- refresh before `ticket status`;
- keep refresh scoped to the resolved project;
- avoid requiring manual rebuild for normal CLI usage.

Expected outcome:

- ticket commands observe direct Markdown edits without user babysitting.

Delivery:

- Repository PR: `#27`
- PR title: `feat(slice-6): refresh before ticket reads`

### Slice 7: Reconciliation Boundary for Execution

Status: reviewing.

Expected branch: `feat/slice-7-reconciliation-execution-boundary`
Expected PR title: `feat(slice-7): reconciliation execution boundary`

Goal:

Provide a reusable boundary that future ticket execution can call before selecting or mutating tickets.

Candidate scope:

- expose a core reconciliation service suitable for command and future daemon use;
- define refresh result semantics for future execution blockers;
- document that execution must reconcile before ticket selection;
- add focused tests for successful refresh and parse-error outcomes.

Expected outcome:

- future single-ticket execution and loop slicings can depend on one reconciliation contract instead of rebuilding their own sync logic.

Delivery:

- Repository PR: pending review.
- PR title: `feat(slice-7): reconciliation execution boundary`

## Acceptance Criteria

The slicing is complete when:

- project ticket files can be discovered deterministically;
- one project index can be rebuilt from Markdown tickets;
- stale, missing, orphaned, and invalid index state can be reported;
- manual rebuild and status commands exist for recovery and diagnostics;
- normal ticket read commands refresh the index automatically;
- future execution code has a reusable reconciliation boundary;
- validation passes with `make verify`.

## Validation

Required validation for every implementation slice in this slicing:

```sh
make verify
```

Manual smoke checks for the completed slicing:

```sh
make install
tmpdir="$(mktemp -d)"
chc init --dir "$tmpdir"
chc ticket new --dir "$tmpdir" "Do something"
sed -i 's/title: "Do something"/title: "Do something edited"/' "$tmpdir/.chc/tickets/CHC-1-Do-something.md"
chc ticket show --dir "$tmpdir" CHC-1
chc index status --dir "$tmpdir"
chc index rebuild --dir "$tmpdir"
```

## Risks

- Automatic refresh must not hide parse errors.
- Refresh before reads must stay scoped to one resolved project.
- Rebuild must not treat derived SQLite rows as canonical.
- Status commands should support debugging without becoming required babysitting.

## Notes

- Files remain canonical.
- SQLite remains derived local state.
- Watcher and daemon reconciliation can build on this slicing later.
