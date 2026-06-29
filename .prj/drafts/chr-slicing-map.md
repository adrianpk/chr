# Chryso Slicing Map

Status: draft

Current alignment: historical draft, superseded by committed slice plans and runtime specs

## Purpose

Define an initial implementation slicing map for Chryso.

This is not a formal delivery plan. It is a working sequence to help turn the capability map into small implementation slices.

## Evolution Note

This draft predates the committed slice workflow and the user-daemon/global-index decision.

The repository now treats a slice as executable only after a plan and tracker are committed to `dev`. The authoritative slice record is the set of committed files under:

```text
.prj/plans/
.prj/trackers/
```

The sequence below remains historical context for the early product direction. It should not be read as the current executable roadmap.

Important later changes:

- root-level `AGENTS.md`, asset profiles, and protocol defaults were delivered through committed slices;
- the ticket model was formalized in `.prj/specs/specs/ticket-template.md`;
- the runtime model moved from project-only state to a user-centric daemon with a local project registry;
- the global SQLite database is a derived operational index for registered projects and tickets;
- slices after the early workspace work followed the committed plan/tracker flow, not this initial draft numbering.

## Slice 1: Asset Profiles

Implement workspace profiles:

- `standard`,
- `personal`,
- `minimal`.

`chr init` uses `standard` by default.

The implementation should:

- embed profile assets from `cmd/chr/assets`,
- support `chr init --profile <name>`,
- copy profile files recursively into `.chr/`,
- return a clear error for unknown profiles.

## Slice 2: Root AGENTS.md

Materialize root-level `AGENTS.md` in managed projects.

The file is required because external agent tooling may look for it at the project root.

The root file should point to `.chr/` while remaining useful as a standalone entry point.

## Slice 3: Protocol Defaults

Define the first `standard` protocol files.

Candidate files:

- `.chr/agents/ARCHITECTURE.md`,
- `.chr/agents/CONVENTIONS.md`,
- `.chr/agents/DOCUMENTING.md`,
- `.chr/agents/INVARIANTS.md`,
- `.chr/agents/LIFECYCLE.md`,
- `.chr/agents/LINTING.md`,
- `.chr/agents/PLAYBOOKS.md`,
- `.chr/agents/ROUTING.md`,
- `.chr/agents/SLICES.md`,
- `.chr/agents/VERSIONING.md`,
- `.chr/agents/playbooks/*`.

The goal is a coherent default, not a mechanical copy of this repository's current `.agents/`.

## Slice 4: Ticket File Model

Define file-based tickets as the primary execution unit.

Include:

- file location by status,
- metadata fields,
- key or slug rules,
- comments,
- blocked reason,
- validation evidence.

## Slice 5: Ticket CLI Basics

Add basic ticket commands before any agent execution:

- create ticket,
- list tickets,
- show ticket status,
- add comments.

The exact command names should follow the domain-action grammar.

## Slice 6: Selection Filters

Add ticket selection filters for listing and later execution:

- use case,
- project area,
- priority,
- severity,
- type,
- label,
- tag.

Filters should compose.

## Slice 7: Conversation Artifacts

Support durable conversation context:

- dumps,
- quizzes,
- raw and curated forms.

These artifacts may later feed ticket context, specs, plans, and decisions.

## Slice 8: Execution Primitive Boundary

Implement bounded primitives:

- file read,
- file write,
- patch apply,
- file listing,
- search,
- command execution,
- git status and related read operations.

All mutating primitives must enforce an allowed root or execution workspace.

## Slice 9: Agent Runtime Dockerfile

Generate or install an agent-runtime Dockerfile template.

This Dockerfile is for agent ticket execution only. It is not the production Dockerfile and not the human development Dockerfile.

The user should be able to tune it during project setup.

The slice should also define where runtime metadata and validation commands live under `.chr/`.

## Slice 10: Codex Harness Spike

Run a simple ticket-like task through Codex inside the chosen runtime boundary.

The output should include:

- result status,
- changed files,
- diff,
- validation output,
- logs,
- structured result envelope.

## Slice 11: Ticket Execution Loop

Execute one ticket first.

Then execute loops over eligible tickets with filters and thresholds.

Record ticket state transitions, blockers, retries, validation, and result artifacts.

## Slice 12: Messaging Notifications

Add optional notifications.

Telegram comes first for operational feedback and unblock prompts.

Email comes later for lower-urgency summaries and reports.
