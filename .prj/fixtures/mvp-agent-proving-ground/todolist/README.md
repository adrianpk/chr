# Todolist Proving Ground Fixture

This fixture is the controlled project used by the MVP Agent Proving Ground
slicing.

Materialize it into the ignored local workspace with:

```sh
./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh
```

Target path:

```text
_workspace/testing/projects/todolist
```

The fixture is intentionally not green at first.

It contains:

- a deliberate backend bug: task IDs are reused after deletion;
- a missing feature: marking a task as completed is not implemented;
- direct proving objectives under `.chr/proving/objectives/`;
- validation commands for compile, lint, and tests;
- a hand-authored `.chr/default/runtime/agent/Dockerfile` for sandboxed agent work.

The goal of the proving ground is for `chr` to run Codex inside the fixture's
agent sandbox, iterate through validation failures, and leave the project green.
