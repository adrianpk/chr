---
title: "Follow-ups Specification"
project_id: "project-dialer-gallery"
case_slug: "followups"
status: "draft"
created_at: "2026-05-30T15:10:00Z"
---

# Follow-ups Specification

Follow-ups represent callback work created from call outcomes. The first version
keeps scheduling in memory and surfaces due callbacks through the agent queue.

## Requirements

- Create a follow-up from a callback disposition.
- Store the requested callback time.
- Exclude future callbacks from the active queue.
- Return due callbacks to the queue.
- Avoid calendar or notification integrations.

## Acceptance

- Callback outcomes create follow-up state.
- Future follow-ups are hidden from active queue results.
- Due follow-ups become callable.
