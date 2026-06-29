# MVP Agent Proving Ground

## Purpose

The MVP Agent Proving Ground is a temporary direct execution path used to prove
that `chr` can run Codex inside an agent sandbox for one controlled code task.

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
.chr/proving/objectives/
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
chr proving run \
  --dir _workspace/testing/projects/todolist \
  --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md \
  --dry-run
```

The dry run prints:

- objective identity;
- runtime image;
- Dockerfile path;
- context, result, log, and validation artifact paths;
- Docker build command;
- Docker run command.

Without `--dry-run`, `chr proving run`:

- loads the objective;
- computes the runtime plan;
- writes the materialized worker context;
- runs the Docker build command;
- runs Codex inside the sandbox container.

Validation and Ralph Loop retries are provided by separate proving commands.

## Validation Runner

Run validation for one objective:

```sh
chr proving validate \
  --dir _workspace/testing/projects/todolist \
  --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md
```

The validation runner executes:

- `make compile`;
- `make lint`;
- `make test`;
- `make verify`.

It writes a Markdown validation artifact to the objective's `validation_path`.
The command exits non-zero when any validation command fails.

## Minimal Ralph Loop

Run a bounded Ralph Loop for one objective:

```sh
chr proving ralph \
  --dir _workspace/testing/projects/todolist \
  --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md
```

The loop:

- runs one sandboxed Codex invocation per attempt;
- validates after each attempt;
- feeds failed validation output into the next attempt context;
- stops on validation pass, repeated failure, blocked execution, or max
  attempts;
- writes per-attempt artifacts under the objective artifact root;
- writes a loop summary to `ralph-loop.md`.

Each Codex invocation is independent. The loop does not rely on implicit
conversation memory. Context transfer between attempts is explicit through the
generated worker context artifact.

## Agent Runtime Authentication

The proving runtime mounts local Codex credentials into the container when they
exist:

- `~/.codex/auth.json` to `/root/.codex/auth.json`;
- `~/.codex/config.toml` to `/root/.codex/config.toml`.

Both mounts are read-only. The Docker run command also passes
`OPENAI_API_KEY` by environment-variable name, so the value can be inherited by
the container without being rendered into command artifacts.

Codex runs with `--dangerously-bypass-approvals-and-sandbox` inside the
container. Docker is the sandbox boundary for this proving runtime; Codex's
inner bubblewrap sandbox is intentionally disabled because nested namespace
creation is not available in the container.

Future clone, commit, and push execution should inherit a repository-scoped PAT
through an explicit secret environment variable. That PAT is not required for
the current proving-ground flow because it works on the mounted core.

## Boundaries

This command intentionally allows fixture-specific behavior. It does not clone
repositories, push commits, create PRs, or generalize the full ticket execution
workflow.
