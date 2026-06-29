# Chryso Capability Map

Status: draft

Current alignment: partially superseded by `.prj/specs/specs/user-daemon-and-global-index.md`

## Purpose

Capture the current capability map for `chr` so product shaping can continue without prematurely locking implementation details.

This is not an implementation plan. It is a map of the areas Chryso is expected to cover.

## Evolution Note

This draft captured the early project-local model.

Later design work refined the runtime model:

- `chr` remains project-local for canonical project artifacts under `.chr/`;
- Markdown tickets remain canonical and portable;
- a user-level daemon may know multiple local projects;
- a user-level project registry is needed so one daemon can coordinate multiple projects;
- a global user SQLite database is used as a derived operational index, not as the canonical ticket source.

The early statement that Chryso should not keep a global project registry is superseded by the user-daemon model. The replacement decision is narrower: user-level state may record registered local projects and derived indexes, while canonical project state remains in each project's `.chr/`.

## Product Shape

Chryso is a Go-focused CLI tool for managing agent-assisted development workflows across local Go projects.

The tool creates and manages `.chr/` in target Go projects.

Project-local artifacts should live in `.chr/`. User-level runtime state may exist outside the project for the multi-project daemon, registered project list, and derived SQLite indexes.

## Workspace Initialization

`chr init` creates a managed project workspace.

The root generated workspace is:

```text
.chr/
```

`chr init` runs against the current directory by default.

Profiles:

- `standard`: default useful profile.
- `personal`: full opinionated workflow for the author's own use.
- `minimal`: lightweight scaffold.

## Local Artifacts

The managed workspace should support these artifact families:

- drafts,
- specs,
- plans,
- trackers,
- reports,
- tickets,
- ticket comments,
- conversation dumps,
- conversation quizzes,
- agent protocol files,
- playbooks,
- context artifacts,
- runtime traces and execution records.

One level of grouping is allowed inside areas such as drafts and specs when it helps cluster documents by use case or domain.

## Tickets

Tickets are the primary execution unit.

Expected ticket capabilities:

- create tickets,
- read ticket status,
- add comments,
- work on one specific ticket,
- work in a loop across eligible tickets,
- move tickets through lifecycle states,
- mark tickets blocked when execution cannot continue,
- record execution outcome and validation evidence.

Ticket selection should eventually support filters by:

- use case,
- project or project area,
- priority,
- severity,
- type,
- label,
- tag,
- combinations of these fields.

The exact ticket schema remains open.

## Execution Primitives

Chryso needs bounded execution primitives comparable in spirit to:

- read files,
- write files,
- apply patches,
- list files,
- search,
- run commands,
- inspect git status,
- create branches or worktrees,
- commit changes,
- produce structured action results.

Primitive execution must be constrained by an allowed project root or execution workspace.

Raw text should not execute directly. Natural language should be normalized into explicit commands, tickets, or structured execution requests before mutation.

## Agent Harness

Codex is the planned default agent harness.

The execution boundary should favor Docker for agent execution because Docker provides stronger sandboxing than a plain git worktree.

Git worktrees remain useful for branch and checkout isolation, but they are not a security boundary by themselves.

## Agent Runtime Dockerfile

Managed projects should be able to keep a Dockerfile or Dockerfile template for agent ticket execution.

This Dockerfile is not the project's production Dockerfile and not necessarily the human development Dockerfile. Its purpose is to give agents a controlled working environment that can:

- install or expose the project dependencies needed for ticket work,
- run validation commands,
- provide Codex or the selected harness,
- mount or receive a ticket-scoped workspace,
- write logs, diffs, validation output, and result envelopes back to host-readable artifact paths.

The user should be able to tune this Dockerfile near project setup time.

Chryso may provide a template and may later assist with edits, but the system should not depend on an LLM fully generating or fully owning this file. Once the user has shaped a working agent runtime, ticket execution can run with less intervention.

The expected distinction is:

```text
project Dockerfile          optional production/development concern
.chr agent Dockerfile       operational sandbox for agent ticket work
```

## Ticket Loop

The ticket loop should support:

- single-ticket execution,
- continuous loop execution,
- filtering eligible tickets,
- retry thresholds,
- configurable blocked thresholds,
- status reporting,
- validation after agent output,
- recording blocker reasons.

When a ticket cannot progress after configured thresholds, Chryso should record the blocker and notify the user when user input can unblock the work.

## Conversation Capture

Conversation context should support two complementary forms:

- dumps: free-form context capture,
- quizzes: bounded question-driven context capture.

These artifacts are durable project context and may later inform ticket execution, specs, or plans.

Raw and curated forms may both be useful.

## Messaging Integrations

Telegram and email are optional integrations.

Telegram is the more important first messaging integration because it can provide faster operational feedback and user input.

Email is useful for lower-urgency notifications, summaries, and reports.

Messaging should not replace the CLI or the `.chr/` workspace state. Messaging should:

- notify progress,
- report blockers,
- ask for user input when a ticket loop stalls,
- record user replies as comments or conversation artifacts,
- report PR or diff summaries when relevant.

## Root AGENTS.md

Managed projects should receive a root-level `AGENTS.md`.

This is required for compatibility with external agent tooling that looks for `AGENTS.md` at the project root.

The root `AGENTS.md` may point into `.chr/` for the fuller protocol, but it should remain useful as a standalone entry point.

## Open Questions

- What is the first ticket file format?
- Which artifacts belong in the `standard` profile by default?
- How much of the agent protocol belongs in `standard` versus `personal`?
- What is the first safe subset of execution primitives?
- What Docker boundary is required for the first Codex harness?
- Which ticket loop thresholds should be configurable first?
- How should Telegram replies map into ticket comments or conversation artifacts?
- What should remain CLI-only before messaging integrations exist?
