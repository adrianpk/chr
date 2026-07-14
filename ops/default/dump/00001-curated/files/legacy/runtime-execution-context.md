---
title: "Runtime Execution Context"
status: "curated"
type: "dump"
case_slug: "ticket-execution-loop"
---

# Runtime Execution Context

Ticket execution uses Docker as the sandbox boundary.

The host-side `chr` process selects work, composes bounded context, plans the runtime, invokes the container, and records artifacts. The runtime works against the mounted project and writes result data, logs, validation evidence, and changed-file summaries.

The goal is autonomous work with observable boundaries:

- execution records identify the attempt;
- context artifacts show what the agent received;
- validation output proves what was run;
- result artifacts describe the outcome;
- ticket status and comments preserve handoff state.

Primary sources:

- `docs/explanation/agent-sandbox-and-runtime.md`
- `docs/explanation/ticket-resolution-lifecycle.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop.md`
