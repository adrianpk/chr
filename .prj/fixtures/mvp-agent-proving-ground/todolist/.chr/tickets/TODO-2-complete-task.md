---
key: TODO-2
title: Implement task completion
status: ready
priority: high
severity: normal
type: feature
tags:
  - backend
  - completion
project_area: todo-store
---

# Implement task completion

The todo store exposes `CompleteTask`, but the method is intentionally not
implemented.

## Acceptance

- Completing an existing task marks it completed.
- Completing a missing task returns `ErrNotFound`.
- `make verify` must pass.
