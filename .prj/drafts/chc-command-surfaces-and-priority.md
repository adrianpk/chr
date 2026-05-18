# ChrysoCoord Command Surfaces and Priority

Status: draft

## Purpose

Capture the current product finding about how a user can command and observe `chc`, and use that finding to prioritize later slicing.

This is not an executable slice plan. A slice becomes executable only after a plan and tracker are committed to `dev`.

## Core Finding

`chc` should support several user command surfaces:

- CLI commands;
- direct file edits inside the managed project workspace;
- Telegram commands or replies;
- email commands or replies.

These surfaces should not imply that the user sends messages to themselves from the terminal. Manual notification commands are only useful as setup and validation tools.

The main value of messaging appears when execution is decoupled from the active terminal session:

- a daemon is running;
- an agent is working on a ticket;
- a ticket loop is processing eligible work;
- the system needs a user decision to unblock progress.

## Command Surface Model

### CLI

The CLI is the primary explicit control surface.

Expected responsibilities:

- initialize `.chc/` in a project;
- register and select local projects;
- create, list, show, and start tickets;
- create, list, and promote drafts and specs;
- start dumps and quizzes;
- start slicing from a spec;
- start one ticket or a filtered ticket loop;
- rebuild or inspect derived indexes;
- configure and test external integrations.

The CLI should remain useful even when no daemon is running.

### Files

Files are the canonical durable project surface.

The user should be able to edit managed Markdown artifacts directly when that is more ergonomic than issuing CLI commands.

Expected responsibilities:

- tickets;
- drafts;
- specs;
- plans;
- trackers;
- dumps;
- quizzes;
- comments or execution notes when represented as file content.

Direct file editing creates a synchronization requirement. `chc` needs explicit or automatic reconciliation so derived indexes do not become stale.

### Telegram

Telegram is an optional remote interaction surface.

Expected responsibilities:

- receive blocker prompts;
- reply with ticket comments or decisions;
- ask for status;
- select or start a ticket when away from the terminal;
- approve or reject the next loop step when human input is required.

Telegram should complement CLI and file workflows. It should not become the canonical state store.

### Email

Email is an optional lower-urgency interaction surface.

Expected responsibilities:

- summaries;
- reports;
- lower-priority blocker notifications;
- review reminders;
- asynchronous replies that can become comments or decisions.

Email is less urgent than Telegram and can come later.

### Daemon and Internal Calls

The daemon is the natural owner of asynchronous execution.

Expected responsibilities:

- watch or reconcile registered projects;
- serve Telegram and email integrations;
- run ticket loops;
- dispatch agent work;
- emit notifications when a threshold or state transition requires user attention;
- maintain the global user-level derived index.

When the daemon and CLI live in the same binary, the CLI may call daemon functionality directly through internal process boundaries or shared packages. The implementation detail is open, but the product model should avoid requiring multiple user-facing binaries.

## Notification Finding

`notify` is not a primary user workflow.

Manual commands such as the following are setup checks:

```text
chc notify telegram test
chc notify email test
```

The real value of notification comes from system-triggered events:

- ticket blocked;
- user decision required;
- loop finished;
- PR ready for review;
- validation failed repeatedly;
- too many iterations happened without progress;
- daemon detected a state that needs intervention.

Notification should therefore be treated as daemon and loop infrastructure, not as an early standalone product surface.

## Candidate Command Families

Current implemented families:

```text
chc init
chc project add
chc project list
chc ticket new
```

Candidate future families:

```text
chc project select
chc project current
chc project sync

chc draft list
chc draft new
chc draft promote

chc spec list
chc spec new
chc spec from-draft

chc dump start
chc dump append
chc dump curate

chc quiz start
chc quiz answer
chc quiz close

chc ticket list
chc ticket show
chc ticket status
chc ticket start
chc ticket sync

chc slice start
chc slice plan
chc slice track
chc slice status

chc loop start
chc loop status
chc loop stop

chc index rebuild
chc index status

chc daemon start
chc daemon status
chc daemon stop

chc notify telegram test
chc notify email test
```

## Filtering and Selection

Ticket and loop commands should eventually support filters:

```text
--project
--use-case
--priority
--severity
--area
--tag
--type
--status
--limit
--order-by
```

Filtering should support both direct ticket start and loop execution.

Examples:

```text
chc ticket start CHC-12
chc loop start --priority high --severity max
chc loop start --use-case onboarding --tag uxui
chc loop start --order-by priority,severity,updated_at
```

## Priority Guidance for Later Slicing

### Priority 1: Project and Document Discovery

The next usable layer should make existing artifacts visible and selectable.

Candidate scope:

- `project select`;
- `project current`;
- list tickets;
- show one ticket;
- list drafts;
- list specs.

Reason: before richer execution exists, the user needs to see and select the operating context.

### Priority 2: Document Creation Flows

After discovery, add first-class document creation.

Candidate scope:

- create draft, with optional use-case grouping;
- create spec, with optional use-case grouping;
- start dump;
- start quiz;
- promote draft to spec.

Reason: this captures the workflow that already exists manually and makes it repeatable through `chc`.

### Priority 3: Reconciliation and Indexing

Before daemon or messaging, make file edits and SQLite agree.

Candidate scope:

- project sync;
- ticket sync;
- index rebuild;
- index status;
- stale index detection.

Reason: the user can edit files directly, so the derived database must be trustworthy or explicitly marked stale.

### Priority 4: Ticket Start and Single-Ticket Execution

Add the first execution entrypoint after the model is visible and reconcilable.

Candidate scope:

- start a specific ticket;
- create an execution record;
- enforce project root boundaries;
- record result status and validation output.

Reason: single-ticket execution is the smallest useful unit before loops.

### Priority 5: Filtered Loop

Add loop execution only after single-ticket execution exists.

Candidate scope:

- loop over all eligible tickets;
- loop with filters;
- loop ordering;
- blocked threshold;
- iteration threshold;
- status reporting.

Reason: loop execution amplifies mistakes, so it should rely on already working ticket selection, indexing, and single-ticket execution.

### Priority 6: Daemon

Add daemon support when asynchronous execution or messaging becomes necessary.

Candidate scope:

- daemon start/status/stop;
- registered project awareness;
- reconciliation scheduling;
- loop ownership;
- event emission.

Reason: the daemon is needed for long-running loops and integrations, but it should sit on top of already clear project, ticket, and index semantics.

### Priority 7: Messaging

Add messaging after daemon events exist.

Candidate scope:

- Telegram setup and test;
- Telegram blocker prompts;
- Telegram replies as comments or decisions;
- email setup and test;
- email reports.

Reason: messaging is valuable when triggered by daemon, loop, or ticket execution events, not as a standalone terminal-driven workflow.

## Open Questions

- Should `project select` write active project state into user config or into the user-level database?
- Should direct file edits be detected by an explicit `sync` command first, or should the daemon watch registered project directories from the beginning?
- Which document families need SQLite indexing in the first pass: tickets only, or drafts and specs too?
- Should Telegram commands map to the same command handlers as CLI commands, or to a narrower interaction API?
- Which loop filters are required before the first usable loop: priority, severity, area, tag, use case, or status?
