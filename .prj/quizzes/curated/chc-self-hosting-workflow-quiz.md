# ChrysoCoord Self-Hosting Workflow Notes

Status: curated

## Paired Draft

- `.prj/specs/drafts/chc-self-hosting-workflow-draft.md`

## Context

ChrysoCoord is being developed with agent collaboration rules stored in `AGENTS.md` and `.agents/`.

The tool being built is intended to provide a related but more formal project-local operating model for Go projects. That managed model will eventually use `.chc/` in target repositories.

For this repository, `.chc/` should not be created yet because the tool is under development, not managing this repository as a target project.

## Curated Decisions

- The product name is ChrysoCoord.
- The CLI command is `chc`.
- The Go module path is `forge.adrianpk.com/chc`.
- The managed project-local workspace root will be `.chc/`.
- This repository may use `.prj/` as a temporary scaffold to emulate future ChrysoCoord artifacts.
- The first product surface is the CLI.
- A TUI may be added later, possibly in the same binary.
- The CLI should use `urfave/cli`.
- The preferred command grammar is domain-action: `chc <domain> <action>`.
- Top-level commands should be reserved for tool-level or workspace-level behavior.
- Tickets are expected to become the main operational unit.
- Execution isolation is a core concern, but the model is not decided yet.

## Preserved Tensions

- `.agents/` and `AGENTS.md` guide current repository collaboration, while future `.chc/` is product-managed runtime state.
- `.prj/` should help this repository practice the intended workflow without pretending that ChrysoCoord is already self-hosting.
- Docker appears stronger for sandboxing, but it is heavier.
- Git worktrees are useful for branch and checkout isolation, but they are not a security boundary.
- Command names should remain discoverable and mnemonic without over-crystallizing the final CLI too early.

## Open Questions

- How much of the future `.chc/` layout should `.prj/` emulate now?
- Should `.prj/` contain tickets immediately, or only specs and curated notes until the ticket model exists?
- What is the smallest useful first ticket command?
- How should `urfave/cli` be introduced without prematurely freezing the whole command tree?
- What evidence is needed before choosing Docker, worktrees, or both as execution strategies?
