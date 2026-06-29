---
objective_id: proving-TODO-1
ticket_key: TODO-1
kind: bugfix
artifact_root: .chr/runtime/runs/proving/TODO-1
context_path: .chr/runtime/runs/proving/TODO-1/context.md
result_path: .chr/runtime/runs/proving/TODO-1/result.json
log_path: .chr/runtime/runs/proving/TODO-1/agent.log
validation_path: .chr/runtime/runs/proving/TODO-1/validation.log
max_attempts: 3
---

# Fix Task ID Reuse

## Objective

Fix the todolist backend so adding a task after deleting another task never
reuses an existing task ID.

The failing behavior is covered by `TestStoreDoesNotReuseDeletedTaskIDs`.

## Allowed Scope

The agent may change:

- `internal/todo/store.go`;
- `internal/todo/store_test.go` only to add legitimate coverage.

The agent must not remove or weaken existing tests.

## Validation Commands

Run these commands after each attempt:

```sh
make compile
make lint
make test
make verify
```

## Success Outcome

The objective succeeds when:

- task IDs are unique and monotonically assigned for the lifetime of a store;
- deleting a task does not cause a future task to reuse an existing ID;
- all validation commands pass.

## Failed Outcome

The objective fails when validation still fails after `max_attempts` attempts.

## Blocked Outcome

The objective is blocked when:

- Codex cannot run in the sandbox;
- the fixture cannot be validated;
- required files are missing;
- the task requires behavior outside the allowed scope.
