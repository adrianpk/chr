# MVP Agent Proving Ground

## Purpose

The MVP Agent Proving Ground is a temporary direct execution path used to prove
that `chc` can run Codex inside an agent sandbox for one controlled code task.

It is post-cutoff MVP hardening. It is not the final ticket workflow.

## Fixture

The committed fixture source lives under:

```text
.prj/fixtures/mvp-agent-proving-ground/todolist
```

Materialize it into the ignored local workspace with:

```sh
./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh
```

The materialized project path is:

```text
_workspace/testing/projects/todolist
```

## Direct Objective

Direct proving objectives live under the fixture project:

```text
.chc/proving/objectives/
```

Each objective declares:

- objective ID;
- related ticket key;
- artifact paths;
- max attempts;
- validation commands;
- success, failed, and blocked outcomes.

## Sandbox Runner

Preview the sandbox run without executing Docker:

```sh
chc proving run \
  --dir _workspace/testing/projects/todolist \
  --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md \
  --dry-run
```

The dry run prints:

- objective identity;
- runtime image;
- Dockerfile path;
- context, result, log, and validation artifact paths;
- Docker build command;
- Docker run command.

Without `--dry-run`, `chc proving run`:

- loads the objective;
- computes the runtime plan;
- writes the materialized worker context;
- runs the Docker build command;
- runs Codex inside the sandbox container.

Validation and Ralph Loop retries are delivered by later proving-ground slices.

## Validation Runner

Run validation for one objective:

```sh
chc proving validate \
  --dir _workspace/testing/projects/todolist \
  --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md
```

The validation runner executes:

- `make compile`;
- `make lint`;
- `make test`;
- `make verify`.

It writes a Markdown validation artifact to the objective's `validation_path`.
The command exits non-zero when any validation command fails.

## Boundaries

This command intentionally allows fixture-specific behavior. It does not clone
repositories, push commits, create PRs, or generalize the full ticket execution
workflow.
