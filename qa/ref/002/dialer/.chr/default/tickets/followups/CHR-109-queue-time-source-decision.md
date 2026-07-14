---
key: "CHR-109"
title: "Queue time source decision"
synopsis: "Decide how queue code receives current time for due callbacks."
status: "blocked"
priority: "normal"
severity: "normal"
type: "task"
tags: ["follow-up", "queue", "time"]
use_case: "followups"
project_area: "architecture"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/followups-plan.md"
slice: "FU-02"
created_at: "2026-05-30T15:53:00Z"
updated_at: "2026-05-30T16:03:00Z"
---

# Description

Queue tests need deterministic time. Decide whether to inject a clock or pass
`now` into queue selection.

# Checklist

- [x] Identify flaky time dependency.
- [ ] Pick clock strategy.
- [ ] Update follow-up queue tests.

# Acceptance

- Queue due-time tests are deterministic.

# Comments

## 2026-05-30T16:03:00Z chr

Blocked on a small design decision: injected clock versus explicit `now`
parameter.

# Blocker

## 2026-05-30T16:03:00Z exec-gallery-109

- kind: decision
- reason: time source must be chosen before implementation continues
- details: see runtime artifacts

# Validation
