---
key: "CHR-102"
title: "Add callback visibility tests"
synopsis: "Cover future callback exclusion and due callback return behavior."
status: "in_progress"
priority: "normal"
severity: "normal"
type: "task"
tags: ["queue", "follow-up", "test"]
use_case: "lead-queue"
project_area: "dialer"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/lead-queue-plan.md"
slice: "LQ-02"
created_at: "2026-05-30T15:46:00Z"
updated_at: "2026-05-30T16:01:00Z"
---

# Description

Add table-driven tests for not-yet-due callbacks and due callbacks in the agent
queue.

# Checklist

- [x] Identify queue time boundary cases.
- [ ] Add test cases for future callbacks.
- [ ] Add test cases for due callbacks.

# Acceptance

- Queue tests explain callback visibility rules.

# Comments

## 2026-05-30T16:01:00Z chr

Work started. The next step is to make queue time dependency testable.

# Blocker

# Validation
