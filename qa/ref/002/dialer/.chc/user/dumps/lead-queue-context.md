---
title: "Lead Queue Context"
status: "curated"
type: "dump"
case: "lead-queue"
---

# Lead Queue Context Dump

The queue should answer one practical question: who should this agent call next?

The first version should rank by score and exclude records that should not be
called. Do-not-call leads are never callable. Future callbacks should not appear
until their scheduled time. Leads owned by another agent should not appear in
the current agent queue.

The queue is a good place to show a seeded bug, because the impact is easy to
understand in the TUI: a do-not-call lead appearing in the queue is clearly
wrong and important.
