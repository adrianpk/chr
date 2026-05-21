# User Guide

This guide covers the ChrysoCoord workflow that exists today.

ChrysoCoord is local-first. A managed project has a `.chc/` directory, while
user-level state stores project registration, indexes, execution records, loop
records, and daemon state.

## Initialize a Workspace

Create project-local ChrysoCoord state:

```sh
chc init
```

Use a different profile when needed:

```sh
chc init --profile minimal
chc init --profile personal
```

Preview initialization:

```sh
chc init --dry-run
```

Initialize another project directory:

```sh
chc init --dir /path/to/project
```

The initialized workspace includes ticket directories, project protocol files,
runtime config, and `.chc/runtime/agent/Dockerfile`.

## Configure Runtime Execution

Runtime settings live in `.chc/config.yml`:

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
```

Install missing runtime defaults in an existing workspace:

```sh
chc runtime init
```

Preview runtime initialization:

```sh
chc runtime init --dry-run
```

Validate the runtime boundary:

```sh
chc runtime validate
```

Preview validation without running Docker:

```sh
chc runtime validate --dry-run
```

## Register and Select Projects

Register a managed project:

```sh
chc project add /path/to/project
```

Override the registry name:

```sh
chc project add --name api-service /path/to/project
```

List registered projects:

```sh
chc project list
```

Select the active project:

```sh
chc project select api-service
```

Print the active project:

```sh
chc project current
```

All project registry commands support `--dev-root` for isolated development
state.

## Create and Manage Tickets

Create a ticket:

```sh
chc ticket new "Add completed task support" --type feature --priority high
```

Useful metadata:

```sh
chc ticket new "Fix task ID reuse" \
  --type bug \
  --priority high \
  --severity critical \
  --use-case todolist \
  --project-area backend \
  --tag runtime
```

List tickets:

```sh
chc ticket list
```

Show one ticket:

```sh
chc ticket show CHC-1
```

Append a comment:

```sh
chc ticket comment CHC-1 "Validation fails in the completion path."
```

Update status:

```sh
chc ticket status CHC-1 ready
```

Ticket files remain Markdown under `.chc/tickets/`.

## Keep the Ticket Index Current

Rebuild the ticket index:

```sh
chc index rebuild
```

Check index status:

```sh
chc index status
```

Ticket commands refresh the index where they need indexed ticket data, but the
index commands are useful when inspecting drift explicitly.

## Run One Ticket

Preview execution:

```sh
chc ticket start CHC-1 --dry-run
```

Run execution:

```sh
chc ticket start CHC-1
```

Execution uses the configured Docker runtime. The worker receives
`CHC_WORKER_CONTEXT_PATH`, reads `context.md`, writes `result.json`, and leaves
runtime and validation logs under the execution run directory.

Apply an execution result to the ticket:

```sh
chc ticket apply-result <execution-id>
```

Preview result application:

```sh
chc ticket apply-result <execution-id> --dry-run
```

Inspect execution evidence:

```sh
chc report list --ticket CHC-1
chc report show --execution <execution-id>
```

## Run a Foreground Ticket Loop

Preview eligible tickets:

```sh
chc ticket loop --dry-run
```

Filter the loop:

```sh
chc ticket loop --dry-run \
  --status ready \
  --priority high \
  --tag runtime \
  --order priority_severity \
  --max-iterations 3
```

Run the foreground loop:

```sh
chc ticket loop --max-iterations 3
```

The loop executes eligible tickets one at a time and stops on the first blocked
or failed ticket.

Inspect loop evidence:

```sh
chc report list --kind loop
chc report show --loop <loop-id>
```

## Capture Authoring Context

Create an authoring session:

```sh
chc author session new --target spec --topic "Billing sync"
```

List authoring sessions:

```sh
chc author session list
```

Show one session:

```sh
chc author session show <session-id>
```

Create a static questionnaire:

```sh
chc author quiz new --target spec --topic "Billing sync"
```

Preview the questionnaire first:

```sh
chc author quiz new --target spec --topic "Billing sync" --dry-run
```

After editing the quiz answers, promote the quiz to a draft:

```sh
chc author promote quiz-to-draft --session <session-id> --quiz .chc/quizzes/<quiz>.md
```

Generated authoring artifacts remain Markdown under `.chc/`. The commands
initialize and connect artifacts; they do not replace direct editing.

Run an assisted authoring session when you want the CLI to drive one question at
a time while preserving every step as Markdown:

```sh
chc author assisted start --target spec --topic "Billing sync"
```

Ask the next question:

```sh
chc author assisted resume \
  --session <session-id> \
  --fake-question "What should the spec include?"
```

Submit the answer:

```sh
chc author assisted answer \
  --session <session-id> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id <question-id> \
  --question "What should the spec include?" \
  --answer "It should include retry behavior and idempotent writes."
```

Check progress:

```sh
chc author assisted status --session <session-id>
```

Finalize the session into a curated summary and target artifact:

```sh
chc author assisted finalize \
  --session <session-id> \
  --fake-summary "Billing sync needs retry behavior and idempotent writes." \
  --target-kind spec \
  --target-name "billing sync spec"
```

The current assisted commands use explicit fake-adapter inputs. They are useful
for deterministic local authoring flows and for validating the persisted model
boundary before a real provider adapter is added.

## Use the Local Daemon Coordinator

Check daemon status:

```sh
chc daemon status
```

Run one foreground coordination iteration:

```sh
chc daemon start --max-iterations 1
```

Request stop:

```sh
chc daemon stop
```

The daemon is a local foreground coordinator. It is not a hosted service.

## Proving-Ground Workflows

The proving-ground commands exercise direct sandbox behavior against controlled
objectives:

```sh
chc proving validate --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md
chc proving run --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md --dry-run
chc proving ralph --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md --max-attempts 3
```

Use these commands for local runtime proofing, not as the main ticket workflow.

## Execution Context

Before agent work, ChrysoCoord can compose a context package with:

- agent identity;
- applicable `AGENTS.md` and `.chc/` protocol files;
- project and ticket documents;
- bounded file tree and text search context;
- optional `gograph` output for Go code context;
- previous attempt evidence for Ralph Loop retries;
- source inventory and readiness handoff.

The main input remains:

```text
<run_artifact_dir>/context.md
```

Sibling context files live under:

```text
<run_artifact_dir>/context/
```

Docker receives the same files through the project mount, usually under
`/workspace/.chc/runtime/runs/...`.

## Current Boundaries

- Context composition runs outside Docker.
- Docker is the runtime sandbox boundary.
- The mounted host workspace is the initial execution mode.
- Avoid editing a project while an agent run is active.
- Clone/worktree execution and PAT-backed push from inside the container are
  future runtime-workspace concerns.
- Assisted authoring prepares Markdown documents; it does not execute tickets or
  run the Docker worker.
- Assisted authoring does not include implicit conversation history. Relevant
  prior discussion must be captured in Markdown artifacts or transcript events.

## Reference

- [Project workspace](../../reference/project-workspace.md)
- [Authoring sessions](../../reference/authoring-sessions.md)
- [Project CLI](../../reference/project-cli.md)
- [Ticket execution](../../reference/ticket-execution.md)
- [Ticket loop](../../reference/ticket-loop.md)
- [Ticket result application](../../reference/ticket-result-application.md)
- [Daemon coordination](../../reference/daemon-coordination.md)
