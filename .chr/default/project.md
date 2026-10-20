---
schema_version: 1
name: "chr"
description: "Local Go-oriented software-development orchestrator for agentic work."
---

chr is a local software-development orchestrator for agentic work in Go projects.

It turns rough intent and project context into durable Markdown-first artifacts: dumps, quizzes, specs, plans, trackers, tickets, execution reports, and review evidence. The user remains in a project-manager role: shaping context, reviewing generated work, choosing what runs, and inspecting outcomes.

The product exposes both a TUI and CLI. The TUI is the operational console for projects, cases, dumps, quizzes, specs, plans, tickets, board state, runtime activity, and review evidence. The CLI mirrors those workflows for scripts and for agents that need to interact with the project workflow directly.

Ticket execution runs through a Docker-based sandbox boundary. The host process resolves project state, selects work, composes context, invokes the runtime, and stores artifacts. The runtime works against the mounted project and writes evidence that chr can later apply and inspect.

The project is intentionally Markdown-first. Project material should remain inspectable and editable as files. SQLite and indexes are derived coordination state, not the primary source of truth.

This repository now also has a project-local `.chr/default/` workspace so chr can manage chr itself. The existing `.prj/` and `ops/` material remains historical delivery context and is not removed by this adoption.
