---
key: TODO-1
title: Fix task ID reuse after delete
status: ready
priority: high
severity: high
type: bug
tags:
  - backend
  - ids
project_area: todo-store
---

# Fix task ID reuse after delete

The todo store currently derives new task IDs from the current task slice
length. After a task is deleted, adding another task can reuse an existing ID.

## Acceptance

- Adding a task after deletion must produce a new unique ID.
- Existing task IDs must remain stable.
- `make verify` must pass.
