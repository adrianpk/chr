---
key: "CHR-105"
title: "Conversion endpoint"
synopsis: "Expose explicit lead-to-prospect conversion over HTTP."
status: "in_progress"
priority: "high"
severity: "normal"
type: "task"
tags: ["conversion", "api"]
use_case: "conversion"
project_area: "api"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/conversion-plan.md"
slice: "CV-02"
created_at: "2026-05-30T15:49:00Z"
updated_at: "2026-05-30T16:02:00Z"
---

# Description

Add `POST /leads/{id}/convert` and delegate conversion rules to internal domain
code.

# Checklist

- [x] Decide route shape.
- [ ] Add handler.
- [ ] Add JSON error responses.
- [ ] Add endpoint tests.

# Acceptance

- Qualified leads can be converted through HTTP.

# Comments

## 2026-05-30T16:02:00Z chr

Route work is in progress. Domain behavior is available from the previous slice.

# Blocker

# Validation
