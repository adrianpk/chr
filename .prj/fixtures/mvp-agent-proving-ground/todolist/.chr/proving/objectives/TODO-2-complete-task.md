---
objective_id: proving-TODO-2
ticket_key: TODO-2
kind: feature
artifact_root: ops/default/run/00002-proving-todo-2
context_path: ops/default/run/00002-proving-todo-2/files/context.md
result_path: ops/default/run/00002-proving-todo-2/evidence/result.json
log_path: ops/default/run/00002-proving-todo-2/evidence/agent.log
validation_path: ops/default/run/00002-proving-todo-2/evidence/validation.log
max_attempts: 3
---

# Complete Task Feature

## Objective

Implement task completion in the todolist backend.

`Store.CompleteTask(id)` should mark an existing task as completed and return
`ErrNotFound` when the task does not exist.

The failing behavior is covered by:

- `TestStoreCompletesTask`;
- `TestStoreCompleteTaskReturnsNotFound`.

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

- completing an existing task sets `Completed` to `true`;
- completing a missing task returns `ErrNotFound`;
- all validation commands pass.

## Failed Outcome

The objective fails when validation still fails after `max_attempts` attempts.

## Blocked Outcome

The objective is blocked when:

- Codex cannot run in the sandbox;
- the fixture cannot be validated;
- required files are missing;
- the task requires behavior outside the allowed scope.
