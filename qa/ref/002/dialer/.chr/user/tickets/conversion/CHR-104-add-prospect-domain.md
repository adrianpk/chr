---
key: "CHR-104"
title: "Add prospect domain model"
synopsis: "Create internal prospect model for converted leads."
status: "done"
priority: "normal"
severity: "low"
type: "task"
tags: ["conversion", "domain"]
use_case: "conversion"
project_area: "dialer"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/conversion-plan.md"
slice: "CV-01"
created_at: "2026-05-30T15:48:00Z"
updated_at: "2026-05-30T16:08:00Z"
---

# Description

Add the internal prospect representation used by conversion work.

# Checklist

- [x] Add prospect type.
- [x] Add conversion test fixture.
- [x] Document invalid source states.

# Acceptance

- Prospect model supports the conversion endpoint.

# Comments

## 2026-05-30T16:08:00Z chr

Domain model landed and was accepted for gallery state.

# Blocker

# Validation

## 2026-05-30T16:08:00Z exec-gallery-104

- command: go test ./...
- status: succeeded
