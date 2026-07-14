---
id: TKT-20260714151004
title: Migrate CHR operational artifacts to ops layout
status: open
kind: task
severity: high
priority: high
scope: persistence
tags: ops-layout, chr-artifacts, migration
source: chat
reported_at: 2026-07-14T15:10:04Z
ready_at:
started_at:
reviewed_at:
closed_at:
branch:
pr:
commits: []
resolution:
---

# Migrate CHR operational artifacts to ops layout

## Problem

The no-legacy path migration incorrectly treated `.chr/default/` as the final
home for too many CHR-managed artifacts. That preserved the legacy pseudo-user
shape instead of completing the operational artifact migration.

## Expected Outcome

- CHR-private configuration and local CHR state may remain under `.chr/`.
- Operational artifacts are stored under the canonical `ops/<namespace>/...`
  layout, with `ops/default` used for no-user workspaces.
- Active code no longer writes or discovers operational artifacts under
  `.chr/default/{dumps,quizzes,specs,plans,tickets,reports,runtime}`.
- Documentation and changelog describe the boundary clearly.

## Notes

This ticket follows the same boundary used in the KAL migration: `ops/` is for
shared operational artifacts, while tool-private configuration remains outside
shared operational artifact space.
