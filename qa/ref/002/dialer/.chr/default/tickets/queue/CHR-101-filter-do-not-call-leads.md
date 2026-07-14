---
key: "CHR-101"
title: "Filter do-not-call leads from queue"
synopsis: "Prevent do-not-call leads from appearing in agent queues."
status: "ready_for_review"
priority: "high"
severity: "high"
type: "bug"
tags: ["queue", "compliance"]
use_case: "lead-queue"
project_area: "dialer"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/lead-queue-plan.md"
slice: "LQ-01"
created_at: "2026-05-30T15:45:00Z"
updated_at: "2026-05-30T16:05:00Z"
---

# Description

Do-not-call leads must never appear in an agent queue, even when they have high
score or a matching owner.

# Checklist

- [x] Add queue regression test.
- [x] Exclude `do_not_call` from queue selection.
- [ ] Review queue ordering after filter change.

# Acceptance

- `agent-1` queue excludes `lead-102`.
- Existing callable leads remain visible.

# Comments

## 2026-05-30T16:05:00Z chr

Implementation is ready for review. The queue now filters compliance-blocked
leads before sorting.

# Blocker

# Validation

## 2026-05-30T16:05:00Z exec-gallery-101

- command: go test ./...
- status: succeeded
