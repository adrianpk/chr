# Daemon Coordination Slicing Plan

Status: completed

## Purpose

Plan the `Daemon Coordination` slicing from the feature slicing set map.

This slicing covers the use case:

> A single user-level `chr` process coordinates multiple registered projects and
> long-running work.

## Source

Primary source:

- `.prj/drafts/chr-feature-slicing-set-map.md`

Related documents:

- `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- `.prj/plans/slicing-automatic-reconciliation-plan.md`
- `docs/reference/project-registry.md`
- `docs/reference/ticket-loop.md`
- `docs/reference/ticket-reconciliation.md`

Prerequisite delivered work:

- `User Project Context` slicing.
- `Automatic Reconciliation` slicing.
- `Codex Runtime Worker` slicing.
- `Filtered Ticket Loop` slicing.

## Current State

Delivered prerequisites:

- user-level project registry exists;
- registered projects can be resolved from commands;
- ticket indexes can be refreshed on command boundaries;
- foreground filtered ticket loops can execute eligible tickets;
- loop run records persist foreground loop outcomes.

Delivered by this slicing:

- daemon command surface;
- daemon lifecycle state;
- single-instance ownership;
- project loading for daemon runs;
- daemon-owned foreground loop execution;
- daemon status reporting;
- graceful stop and stale process recovery.

## Slicing Scope

This slicing includes:

- daemon command skeleton;
- daemon lifecycle state model;
- daemon state persistence;
- single-instance guard;
- registered project loading;
- foreground-safe daemon run boundaries;
- daemon status reporting;
- graceful stop request handling;
- stale process recovery.

## Out of Scope

This slicing does not include:

- notification or remote command surfaces;
- Telegram, email, or external transport setup;
- hosted execution services;
- multi-user coordination;
- forge or PR automation;
- OS service manager installation.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Daemon Command Skeleton

Status: delivered.

Expected branch: `feat/slice-1-daemon-command-skeleton`
Expected PR title: `feat(slice-1): daemon command skeleton`
Expected report: `.prj/reports/slices/daemon-coordination/slice-1-daemon-command-skeleton.md`

Goal:

Add the user-facing daemon command group without starting background ownership
yet.

Candidate scope:

- add `chr daemon` command group;
- add `daemon status`, `daemon start --dry-run`, and `daemon stop --dry-run`;
- resolve user environment for daemon commands;
- keep daemon commands side-effect free in this slice;
- add focused CLI tests;
- document the daemon command boundary.

Expected outcome:

- later slices can attach lifecycle state and process ownership behind a stable
  command surface.

### Slice 2: Daemon Lifecycle State Model

Status: delivered.

Expected branch: `feat/slice-2-daemon-lifecycle-state`
Expected PR title: `feat(slice-2): daemon lifecycle state`
Expected report: `.prj/reports/slices/daemon-coordination/slice-2-daemon-lifecycle-state.md`

Goal:

Define durable daemon lifecycle state independent from process startup.

Candidate scope:

- add daemon state model;
- track status, PID, started time, heartbeat time, and stop request time;
- validate lifecycle transitions;
- add focused model tests;
- document lifecycle fields.

Expected outcome:

- daemon state has a clear contract before persistence and process ownership.

### Slice 3: Daemon State Repository

Status: delivered.

Expected branch: `feat/slice-3-daemon-state-repository`
Expected PR title: `feat(slice-3): daemon state repository`
Expected report: `.prj/reports/slices/daemon-coordination/slice-3-daemon-state-repository.md`

Goal:

Persist daemon lifecycle state in the user-level database.

Candidate scope:

- add SQLite migration;
- add daemon state repository;
- support get, upsert, clear, and heartbeat updates;
- add repository tests.

Expected outcome:

- daemon commands can read and update durable lifecycle state.

### Slice 4: Daemon Status Reporting

Status: delivered.

Expected branch: `feat/slice-4-daemon-status-reporting`
Expected PR title: `feat(slice-4): daemon status reporting`
Expected report: `.prj/reports/slices/daemon-coordination/slice-4-daemon-status-reporting.md`

Goal:

Expose daemon lifecycle state through `chr daemon status`.

Candidate scope:

- read daemon state from the repository;
- print stopped, running, stale, and stopping states;
- include PID and heartbeat details when available;
- add focused CLI tests.

Expected outcome:

- users can inspect daemon state before start and stop behavior is active.

### Slice 5: Single-Instance Guard

Status: delivered.

Expected branch: `feat/slice-5-daemon-single-instance`
Expected PR title: `feat(slice-5): daemon single instance`
Expected report: `.prj/reports/slices/daemon-coordination/slice-5-daemon-single-instance.md`

Goal:

Prevent two local daemon owners from running for the same user environment.

Candidate scope:

- add a user-level lock or ownership guard;
- detect existing running daemon state;
- reject duplicate start attempts;
- add stale-state classification hooks;
- add focused guard tests.

Expected outcome:

- daemon startup has a safe ownership boundary.

### Slice 6: Registered Project Loading

Status: delivered.

Expected branch: `feat/slice-6-daemon-project-loading`
Expected PR title: `feat(slice-6): daemon project loading`
Expected report: `.prj/reports/slices/daemon-coordination/slice-6-daemon-project-loading.md`

Goal:

Load registered projects for daemon coordination.

Candidate scope:

- list enabled registered projects;
- validate project paths before daemon work;
- classify missing or invalid project paths;
- add focused project-loading tests.

Expected outcome:

- daemon runs can select project work without relying on a current directory.

### Slice 7: Daemon Foreground Run Loop

Status: delivered.

Expected branch: `feat/slice-7-daemon-foreground-run-loop`
Expected PR title: `feat(slice-7): daemon foreground run loop`
Expected report: `.prj/reports/slices/daemon-coordination/slice-7-daemon-foreground-run-loop.md`

Goal:

Run daemon coordination in the foreground while preserving process ownership
semantics.

Candidate scope:

- add foreground `daemon start` execution path;
- update heartbeat state while running;
- periodically refresh registered projects;
- invoke existing foreground-safe loop boundaries;
- support a small testable iteration limit;
- add focused run-loop tests.

Expected outcome:

- daemon coordination can run as a supervised foreground process before any OS
  service integration exists.

### Slice 8: Graceful Stop and Stale Recovery

Status: delivered.

Expected branch: `feat/slice-8-daemon-stop-recovery`
Expected PR title: `feat(slice-8): daemon stop recovery`
Expected report: `.prj/reports/slices/daemon-coordination/slice-8-daemon-stop-recovery.md`

Goal:

Let users request daemon stop and recover from stale daemon state.

Candidate scope:

- persist stop requests;
- make the run loop observe stop requests;
- mark stopped daemon state on graceful exit;
- classify stale process state;
- expose recovery behavior in status output;
- add focused stop and stale recovery tests.

Expected outcome:

- users can stop daemon coordination and recover from interrupted daemon runs.

## Validation Record

- Planning validation: documentation diff reviewed.
- Completed slicing validation: `make verify` passed on `dev` after PR #71 merge.

## Notes

- This slicing keeps daemon coordination inside the `chr` binary.
- Notification and remote command surfaces remain post-daemon work.
- Slice numbers are local to this slicing and start at `Slice 1`.
