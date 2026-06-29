---
title: "Lead Queue Specification"
project_id: "project-dialer-gallery"
case_slug: "lead-queue"
status: "accepted"
created_at: "2026-05-30T15:00:00Z"
---

# Lead Queue Specification

Agents need a reliable list of leads to call now. The queue is scoped to one
agent, sorted by lead score, and filtered for callability.

## Requirements

- Return leads assigned to the requested agent.
- Sort callable leads by score, highest first.
- Exclude do-not-call leads.
- Exclude future callbacks until due.
- Return JSON responses.

## Acceptance

- Agent queues never include do-not-call leads.
- Higher-scored callable leads appear first.
- Future callbacks stay out of the active queue.
- Queue behavior is covered by table-driven tests.
