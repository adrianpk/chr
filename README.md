# ChrysoCoord

ChrysoCoord organizes agent work inside a software project.

It gives the project a local workspace where tickets, specifications, plans,
execution evidence, and handoff material can live together. The agent can use
that workspace to receive work, reason from project documents, execute tickets,
track progress, and leave behind a traceable record of what happened.

ChrysoCoord is implemented in Go and is currently optimized for Go projects.

## What ChrysoCoord Does

Working with an agent on a real project becomes difficult when the agent has no
durable project-local memory, no structured task workflow, no clear execution
boundary, and no stable record of attempts or blockers.

ChrysoCoord gives that work a shape. It keeps project-specific agent state close
to the code, uses Markdown documents as editable workflow artifacts, and lets the
agent work from tickets instead of from loose chat context.

## How It Works

At a high level, the workflow is:

1. A project gets a hidden local workspace.
2. The user creates or imports project knowledge such as drafts,
   specifications, plans, trackers, quizzes, dumps, and tickets.
3. The agent receives work through Markdown tickets.
4. Ticket execution runs inside a sandboxed Docker runtime.
5. The harness runs the agent workflow, including Ralph-loop based ticket
   resolution where applicable.
6. Results are written back as ticket updates, comments, checklists, logs,
   validation evidence, and changed-file summaries.
7. The user inspects, reviews, continues, or promotes the work.

## Project Workspace

ChrysoCoord maintains project-specific state in a hidden workspace inside the
project. That workspace can contain drafts, specs, plans, trackers, tickets,
reports, quizzes, fixtures, execution artifacts, and validation evidence.

These files are intentionally project-local. They are inspectable, editable, and
connected to the code they describe. Generated documents are not locked away in a
closed database; a technical user can edit the Markdown directly when that is
the clearest path.

See [Project workspace](docs/reference/project-workspace.md) for the exact
layout and contracts.

## Tickets and Agent Work

Tickets are the project-local task surface used by the agent. A ticket describes
work the agent can execute and can carry metadata such as priority, severity,
type, labels, use case, or project area.

The agent can work on one explicit ticket, or ChrysoCoord can run a loop over
eligible tickets. Loops can be filtered and ordered so the agent only picks work
that matches the intended scope.

Tickets include checklists for subtasks and comment areas for progress,
blockers, failed attempts, evidence, and handoff notes. The agent is not expected
to spin forever when stuck: repeated failures can be recorded, and future
notification channels can surface blockers outside the CLI.

See [Ticket loop](docs/reference/ticket-loop.md),
[Ticket execution](docs/reference/ticket-execution.md), and
[Ticket comments](docs/reference/ticket-comments.md) for detailed behavior.

## Documents, Specs, Plans, and Trackers

ChrysoCoord is Markdown-first. Commands can create base project documents, but
the documents remain normal files that can be reviewed and edited directly.

The intended document progression is simple:

- a dump or quiz captures rough project knowledge;
- a draft turns that material into a first structured document;
- a specification defines expected behavior;
- a plan defines implementation work;
- a tracker records progress through that work.

The current CLI supports authoring sessions, static quiz generation, quiz to
draft promotion, and assisted authoring flows that persist their state as
Markdown.

See [Authoring sessions](docs/reference/authoring-sessions.md) and
[Document generation model](docs/explanation/document-generation-model.md).

## Slicing

ChrysoCoord is built around slicing larger work into smaller implementation
units.

A feature or use case can be described, turned into specifications and plans,
and then split into slices. A slice can contain tasks, and tasks can correspond
to commits. Slices can be associated with branches and pull requests, so the
user can review the work manually. In trusted workflows, auto-merge may be
configured later.

This matters because a plan should make implementation visible before the agent
starts writing code. The user should be able to inspect how many slices and
tasks are expected, how the commits are distributed, and where review will
happen.

See [Slicing workflow](docs/explanation/slicing-workflow.md).

## Knowledge Capture

ChrysoCoord's document model is designed around two knowledge-capture paths.

A dump captures broad project or use-case context. It is useful when the user
needs to preserve rough knowledge before turning it into a draft, specification,
plan, or ticket.

A quiz is a bounded question session. The agent asks focused questions to
extract missing knowledge, then stores the raw and curated material as reusable
documentation input. It is structured knowledge extraction, not disposable chat
history.

The current CLI exposes quiz-based authoring and assisted authoring sessions.
The dump command surface belongs to the same Markdown-first authoring direction.

See [Document generation model](docs/explanation/document-generation-model.md)
and [User guide](docs/how-to/user-guide/index.md).

## Runtime

When the agent works on a ticket, ChrysoCoord can start a Docker container and
run the configured harness inside it. The harness receives the execution
context, works on the ticket, runs validation, and leaves behind result data,
logs, validation evidence, and changed-file information.

The Docker container is the sandboxed runtime boundary. ChrysoCoord prepares the
context on the host side and passes explicit artifacts into the runtime instead
of relying on hidden conversation state.

See [Agent sandbox and runtime](docs/explanation/agent-sandbox-and-runtime.md)
and [Runtime configuration](docs/reference/runtime-config.md).

## Current Capabilities

Project workspace:

- initializes a project-local workspace;
- stores project-specific agent material;
- keeps durable state close to the project.

Documents:

- creates drafts, quizzes, reports, and related Markdown artifacts;
- supports manual editing;
- supports promotion from quiz material into drafts;
- supports assisted authoring sessions with durable transcripts.

Tickets:

- creates and indexes Markdown tickets;
- lets the agent execute explicit tickets or filtered ticket loops;
- records progress, comments, checklists, and execution outcomes.

Runtime:

- validates the project runtime boundary;
- runs ticket execution inside Docker;
- captures logs, results, validation evidence, and changed-file summaries.

Coordination:

- registers local projects;
- resolves active or explicit project directories;
- coordinates foreground loops and local daemon-style operation.

Planned integrations include Telegram command handling and status delivery
through Telegram or email. Those messaging channels are future-facing and are
not required for the local CLI workflow that exists today.

## Install

Build the CLI locally:

```sh
make build
```

Run it from the repository:

```sh
./bin/chc --help
```

Install it into the active Go environment:

```sh
make install
```

## First Project

Inside a project you want ChrysoCoord to manage:

```sh
chc init
chc runtime validate --dry-run
chc ticket new "Fix the failing task workflow" --type bug --priority high
chc ticket list
```

This initializes the project workspace, checks the configured runtime boundary,
creates a first ticket, and lists the tickets available to the agent.

To use the project registry:

```sh
chc project add /path/to/project
chc project select /path/to/project
chc project current
```

## Documentation

The README is the front door. Detailed technical material lives under
`docs/`, organized with Diataxis:

- [Documentation index](docs/index.md)
- [Getting started tutorial](docs/tutorials/getting-started.md)
- [User guide](docs/how-to/user-guide/index.md)
- [Reference](docs/reference/project-workspace.md)
- [Explanation](docs/explanation/architecture.md)

Use reference pages for exact command behavior, workspace layout, configuration,
ticket schema, and runtime contracts. Use explanation pages for architecture,
runtime design, slicing, ticket lifecycle, and document-generation rationale.

## Development

Common local checks:

```sh
make fmt
make test
make verify
```

Run the CLI without installing:

```sh
make run RUN_ARGS="--help"
```
