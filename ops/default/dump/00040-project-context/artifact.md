---
title: "chr Project Context"
status: "curated"
type: "dump"
---

# chr Project Context

chr is a local orchestrator for agentic software development in Go projects.

Its core workflow is:

1. capture project context through dumps and quizzes;
2. generate or curate a spec;
3. generate a plan and tracking document;
4. materialize tickets;
5. run one ticket or a queue through a sandboxed agent runtime;
6. update board state and preserve execution evidence;
7. review accumulated evidence before promotion.

The product is operated through a TUI and mirrored CLI commands. The TUI is the primary interactive console. The CLI exists for scripting, automation, and agent self-service.

The project favors Markdown-first artifacts under `.chr/default/`. SQLite, indexes, runtime logs, and traces are derived operational state.

Important source documents:

- `README.md`
- `docs/explanation/architecture.md`
- `docs/explanation/document-generation-model.md`
- `docs/explanation/ticket-resolution-lifecycle.md`
- `docs/explanation/agent-review-model.md`
- `docs/reference/commands.md`
- `docs/reference/tui.md`
