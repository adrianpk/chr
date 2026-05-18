# ChrysoCoord Local Scaffold Distinction

Status: draft

## Context

ChrysoCoord grew out of the practical `AGENTS.md` and `.agents/` repository protocol used to coordinate human-agent development work.

The goal is to formalize part of that repeated operating intelligence into a Go-focused tool. Unlike broader systems, ChrysoCoord initially targets Go projects and uses the `chc` CLI as its first operating surface.

## Local Layers

This repository intentionally has three similar-looking concepts:

- `AGENTS.md` and `.agents/`: the current repository collaboration protocol.
- `.prj/`: the durable local scaffold used to develop ChrysoCoord itself.
- `.chc/`: the product-owned workspace directory that `chc` creates and manages in target Go projects.

## Why `.prj/` Exists Here

The ChrysoCoord repository should not store its durable design notes in `.chc/` yet.

Future tests of the `chc` binary may create, overwrite, reset, or delete `.chc/` while validating product behavior. Keeping real design context in `.chc/` would risk losing those docs during a failed or destructive test.

`.prj/` is therefore a manual scaffold and contextual memory area for developing `chc`.

## Operating Rule

In this repository:

```text
.prj/ is durable development context.
.chc/ is product output and may be used as disposable test state.
AGENTS.md and .agents/ remain the active collaboration rules.
```

If ChrysoCoord later becomes capable enough to manage its own development, this distinction can be revisited through an explicit self-hosting decision.
