---
title: "Lead Conversion Specification"
project_id: "project-dialer-gallery"
case_slug: "conversion"
status: "accepted"
created_at: "2026-05-30T15:05:00Z"
---

# Lead Conversion Specification

Qualified leads should become prospects through an explicit conversion action.
Invalid transitions should be rejected with JSON client errors.

## Requirements

- Convert only leads in the `qualified` state.
- Reject `lost` and `do_not_call` leads.
- Preserve call outcome history.
- Expose conversion result as JSON.
- Keep prospect storage in memory.

## Acceptance

- Qualified leads convert successfully.
- Invalid states return clear JSON errors.
- Conversion behavior is covered by focused tests.
