---
key: "CHR-108"
title: "Follow-up domain model"
synopsis: "Represent scheduled callbacks created from call outcomes."
status: "ready_for_review"
priority: "normal"
severity: "normal"
type: "task"
tags: ["follow-up", "domain"]
use_case: "followups"
project_area: "dialer"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/followups-plan.md"
slice: "FU-01"
created_at: "2026-05-30T15:52:00Z"
updated_at: "2026-05-30T16:06:00Z"
---

# Description

Add an in-memory follow-up model for callback scheduling.

# Checklist

- [x] Add follow-up type.
- [x] Add schedule test.
- [ ] Review naming before merge.

# Acceptance

- Callback outcomes can create follow-up records.

# Comments

## 2026-05-30T16:06:00Z chr

Follow-up model is ready for review.

# Blocker

# Validation

## 2026-05-30T16:06:00Z exec-gallery-108

- command: go test ./...
- status: succeeded
