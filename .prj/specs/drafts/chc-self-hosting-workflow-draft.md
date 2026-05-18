# ChrysoCoord Self-Hosting Workflow

Status: draft

## Purpose

Capture the early workflow for developing ChrysoCoord with a project-local asset directory that emulates the future `.chc/` workspace without using `.chc/` in this repository yet.

This draft exists because ChrysoCoord is not usable enough to manage its own development directly, but the project should still exercise the same operating ideas that the tool is meant to provide later.

## Paired Notes

- `.prj/quizzes/curated/chc-self-hosting-workflow-quiz.md`

## Current Decision

Use `.prj/` as a temporary project-local operating directory for this repository.

`.prj/` is not the ChrysoCoord runtime directory. It is a development scaffold used to test the approach before `chc` can create and manage `.chc/` workspaces itself.

The eventual managed workspace root remains `.chc/`.

## Rationale

The repository already has `AGENTS.md` and `.agents/` for human-agent collaboration rules. Those files guide how work is performed in this repository, but they are not the product runtime surface.

`.prj/` gives the project a controlled place to store draft specifications, curated discussion, and early operating artifacts while keeping `.chc/` reserved for workspaces created by the `chc` binary.

If ChrysoCoord becomes capable enough, the project can later use `chc` to manage its own development through a real `.chc/` workspace.

## Scope

This draft covers:

- the temporary `.prj/` scaffold,
- the relationship between `.prj/`, `.agents/`, and future `.chc/`,
- the early CLI-first direction,
- the current command grammar preference,
- ticket-oriented execution as a design direction,
- sandboxing and worktree isolation as open architecture concerns.

This draft does not define:

- the final ticket schema,
- the final command tree,
- Docker execution behavior,
- worktree execution behavior,
- the self-hosting migration path from `.prj/` to `.chc/`.

## Early Direction

ChrysoCoord starts as a CLI-first Go tool.

The future TUI may live in the same binary, but it is not the first official surface.

The CLI should use `urfave/cli` for modular command organization.

The preferred command grammar is:

```text
chc <domain> <action>
```

Top-level commands are reserved for tool-level or workspace-level behavior.

Examples are illustrative, not stable contracts:

```text
chc init
chc ticket new
chc ticket status
chc ticket loop
```

## Ticket Direction

Tickets are expected to become the main operational unit.

ChrysoCoord should eventually support:

- working on a specific ticket,
- working in a loop across multiple eligible tickets,
- selecting tickets by metadata such as priority, severity, tags, or combinations of those fields.

The exact ticket format and selection rules remain open.

## Execution Isolation Direction

Ticket execution probably needs an isolated workspace.

Two candidate approaches are preserved:

- Docker for stronger process and filesystem sandboxing.
- Git worktrees for lightweight change isolation.

Git worktrees should not be treated as a security sandbox. They may isolate repository changes, but they do not by themselves prevent a worker process from touching files outside the worktree.

The project should investigate whether worker execution can be reliably constrained to a worktree root before treating worktrees as anything stronger than operational isolation.

## Open Questions

- What is the minimal `.prj/` layout needed for the current repository?
- When should `.prj/` be migrated, discarded, or converted into `.chc/`?
- Should `.prj/` mirror the future `.chc/` layout exactly or stay smaller?
- What is the first useful ticket artifact format?
- When should the current stdlib CLI be replaced with `urfave/cli`?
- What is the first safe execution mode: local, worktree, Docker, or no execution loop at all?
