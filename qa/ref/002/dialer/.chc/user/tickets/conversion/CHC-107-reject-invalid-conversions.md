---
key: "CHC-107"
title: "Reject invalid conversions"
synopsis: "Return clear client errors for lost and do-not-call leads."
status: "ready_for_dev"
priority: "high"
severity: "normal"
type: "task"
tags: ["conversion", "validation"]
use_case: "conversion"
project_area: "api"
source_kind: "plan_tracking"
source_ref: ".chc/user/plans/conversion-plan.md"
slice: "CV-04"
created_at: "2026-05-30T15:51:00Z"
updated_at: "2026-05-30T15:51:00Z"
---

# Description

Invalid conversion attempts should fail predictably with JSON client errors.

# Checklist

- [ ] Reject `lost`.
- [ ] Reject `do_not_call`.
- [ ] Add table-driven transition tests.

# Acceptance

- Invalid transitions do not create prospects.

# Comments

# Blocker

# Validation
