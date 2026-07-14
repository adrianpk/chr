---
schema_version: 1
name: "Ticket Execution Loop"
description: "Execute one ticket or a ready queue through the sandboxed agent runtime."
---

This use case covers direct ticket execution, queue execution, result application, board updates, and execution evidence.

The user can run one ticket or let Chryso process ready tickets in order. Each attempt should preserve context, runtime details, logs, validation output, result artifacts, and changed-file evidence. The board should reflect work state without requiring the user to approve every harness pause.

Execution must remain bounded, inspectable, and recoverable.
