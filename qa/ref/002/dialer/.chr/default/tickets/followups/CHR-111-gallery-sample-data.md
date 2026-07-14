---
key: "CHR-111"
title: "Gallery sample data"
synopsis: "Seed representative leads, outcomes, and follow-ups for screenshots."
status: "ready_for_merge"
priority: "low"
severity: "low"
type: "task"
tags: ["gallery", "fixture"]
use_case: "followups"
project_area: "qa"
source_kind: "plan_tracking"
source_ref: "ops/default/plan/<id>/artifact.mdfollowups-plan.md"
slice: "FU-04"
created_at: "2026-05-30T15:55:00Z"
updated_at: "2026-05-30T16:09:00Z"
---

# Description

Add enough representative sample data to make the TUI gallery screenshots useful.

# Checklist

- [x] Seed high-score lead.
- [x] Seed do-not-call lead.
- [x] Seed future callback.

# Acceptance

- Screenshots show a realistic queue and board.

# Comments

## 2026-05-30T16:09:00Z chr

Ready to merge into the gallery fixture.

# Blocker

# Validation

## 2026-05-30T16:09:00Z exec-gallery-111

- command: go test ./...
- status: succeeded
