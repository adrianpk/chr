---
key: "CHC-103"
title: "Queue summary metadata"
synopsis: "Return queue counts for callable, deferred, and blocked leads."
status: "ready_for_dev"
priority: "normal"
severity: "low"
type: "task"
tags: ["queue", "metadata"]
use_case: "lead-queue"
project_area: "api"
source_kind: "plan_tracking"
source_ref: ".chc/user/plans/lead-queue-plan.md"
slice: "LQ-03"
created_at: "2026-05-30T15:47:00Z"
updated_at: "2026-05-30T15:47:00Z"
---

# Description

Add summary metadata to queue responses so the TUI and API can show why work is
or is not available.

# Checklist

- [ ] Define response shape.
- [ ] Count callable leads.
- [ ] Count deferred callbacks.
- [ ] Count compliance-blocked leads.

# Acceptance

- Queue response includes useful summary counts.

# Comments

# Blocker

# Validation
