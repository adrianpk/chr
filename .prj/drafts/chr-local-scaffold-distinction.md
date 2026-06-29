# Chryso Local Scaffold Distinction

Status: draft

## Context

Chryso grew out of the practical `AGENTS.md` and `.agents/` repository protocol used to coordinate human-agent development work.

The goal is to formalize part of that repeated operating intelligence into a Go-focused tool. Unlike broader systems, Chryso initially targets Go projects and uses the `chr` CLI as its first operating surface.

## Local Layers

This repository intentionally has three similar-looking concepts:

- `AGENTS.md` and `.agents/`: the current repository collaboration protocol.
- `.prj/`: the durable local scaffold used to develop Chryso itself.
- `.chr/`: the product-owned workspace directory that `chr` creates and manages in target Go projects.

## Why `.prj/` Exists Here

The Chryso repository should not store its durable design notes in `.chr/` yet.

Future tests of the `chr` binary may create, overwrite, reset, or delete `.chr/` while validating product behavior. Keeping real design context in `.chr/` would risk losing those docs during a failed or destructive test.

`.prj/` is therefore a manual scaffold and contextual memory area for developing `chr`.

## Operating Rule

In this repository:

```text
.prj/ is durable development context.
.chr/ is product output and may be used as disposable test state.
AGENTS.md and .agents/ remain the active collaboration rules.
```

If Chryso later becomes capable enough to manage its own development, this distinction can be revisited through an explicit self-hosting decision.
