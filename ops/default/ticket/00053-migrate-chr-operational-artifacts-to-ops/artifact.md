---
id: TKT-20260714151004
title: Migrate chr operational artifacts to ops layout
status: solved
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
closed_at: 2026-07-14T16:13:33Z
branch: fix/slice-1-ops-artifact-boundary
pr: 361
commits:
  - c63e92d1
  - d86a25e911a650a3d8798867d4c130c6dd886aaa
resolution: fixed
---

# Migrate chr operational artifacts to ops layout

## Problem

The no-legacy path migration incorrectly treated `.chr/default/` as the final
home for too many chr-managed artifacts. That preserved the legacy pseudo-user
shape instead of completing the operational artifact migration.

## Expected Outcome

- chr-private configuration and local chr state may remain under `.chr/`.
- Operational artifacts are stored under the canonical `ops/<namespace>/...`
  layout, with `ops/default` used for no-user workspaces.
- Active code no longer writes or discovers operational artifacts under
  `.chr/default/{dumps,quizzes,specs,plans,tickets,reports,runtime}`.
- Documentation and changelog describe the boundary clearly.

## Notes

This ticket follows the same boundary used in the KAL migration: `ops/` is for
shared operational artifacts, while tool-private configuration remains outside
shared operational artifact space.

## Resolution

Closure audit on 2026-10-20 verified that [PR #361](https://forge.adrianpk.com/adrianpk/chr/pulls/361)
merged at `d86a25e911a650a3d8798867d4c130c6dd886aaa`, an ancestor of dev. The
closure timestamp is the verified merge time. The migrated artifact identity
and references remain unchanged.
