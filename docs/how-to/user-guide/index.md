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
minimal user-owned config, `.chc/user/`, and
`.chc/user/runtime/agent/Dockerfile`.

Set an active operator id for workflows that need project-local userspace
ownership:

```sh
export CHC_OPERATOR_ID=adrian
```

With direnv, the same export may live in `.envrc`.

## Configure Runtime Execution

Runtime defaults are owned by `chc`. A new project does not need to repeat
default runtime paths in `.chc/user/config.yml`.

Override runtime settings only when the project has a real reason to do so:

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/user/runtime/agent/Dockerfile
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

## Configure Work Readiness

Work readiness protects the project before ChrysoCoord starts agent work. The
default policy expects a clean `dev` branch with `origin` as the canonical
remote, but projects can configure their own branch posture in
`.chc/user/config.yml`.

Example:

```yml
work:
  mode: integration_branch
  integration_branch: develop
  canonical_remote: upstream
  require_clean_worktree: true
  require_remote_tracking: true
  allow_safe_repair: true
  allow_branch_creation: false
```

Inspect readiness:

```sh
chc work doctor
```

Apply safe repair when ChrysoCoord reports one:

```sh
chc work doctor --fix-safe
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

## Open the TUI Console

Open the local operational console from the resolved project:

```sh
chc
```

The explicit form is:

```sh
chc tui
```

Open a specific project directory:

```sh
chc tui --dir /path/to/project
```

The TUI is the main interactive surface for project context, use cases, dumps,
quizzes, specs, plans, tickets, and execution activity. It loads durable
workspace files from disk and writes normal Markdown artifacts back to the
core.

## Navigate the TUI

Use the function-key tabs:

| Key | View |
| --- | --- |
| `F1` | Project |
| `F2` | Cases |
| `F3` | Dumps |
| `F4` | Quizzes |
| `F5` | Specs |
| `F6` | Plans |
| `F7` | Board |
| `F8` | Dash |
| `F9` | Help |

Common navigation:

| Key | Behavior |
| --- | --- |
| `tab` | Move to the next pane. |
| `shift+tab` | Move to the previous pane. |
| `up` / `k` | Move to the previous selected row or field when the focused pane supports selection. |
| `down` / `j` | Move to the next selected row or field when the focused pane supports selection. |
| `alt+m` | Toggle maximized mode for the focused pane. |
| `esc` | Return focus to the primary pane, cancel an active prompt, or close an active mode. |
| `q`, `ctrl+c`, `F10` | Quit the TUI. |
| `:` | Open the command palette. |
| `/` | Open filtering where the current view supports filtering. |

The footer tabs are clickable. The TUI also supports mouse resizing between
panes: drag near the vertical divider, or near the horizontal divider in stacked
right panes.

## Use the Command Palette

Press `:` to open the command palette.

Inside the palette:

| Key | Behavior |
| --- | --- |
| type text | Filter commands by id, title, or description. |
| `up` / `down` | Select a command. |
| `enter` | Run the selected command. |
| `backspace` | Remove one filter character. |
| `esc` | Close the palette. |

The palette includes navigation, refresh, selected-source inspection, artifact
creation, dump, quiz, spec, acceptance feature, plan, ticket, and work commands.
Some commands are view-gated. Spec and acceptance feature commands appear only
in Specs. Plan commands appear only in Plans, and plan tracking appears only
when the selected plan exists.

Palette commands that change files open the same confirmation prompt used by
direct keybindings. Confirm with `enter` or cancel with `esc`.

## Edit Project Context

Open the Project view with `F1`.

The Project view contains project name, description, long-form context,
location, and artifact counts. The context here is project-wide material that
feeds later spec and plan generation.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select the project field in the primary pane. |
| `enter` or `e` | Edit the selected field. |
| `enter` while editing | Apply the field edit locally. |
| `esc` while editing | Cancel the field edit. |
| `s` or `ctrl+s` | Save project metadata to disk. |
| `r` | Refresh the workspace graph from disk. |

Refresh is blocked while project edits are active or unsaved.

## Manage Use Cases

Open the Cases view with `F2`.

The system is use-case based. The list represents use cases. Use-case metadata
and context become part of the material used by dumps, quizzes, specs, plans,
and tickets.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` in the list pane | Select a use case. |
| `enter` in the list pane | Move focus into the use-case detail pane. |
| `up` / `down` or `k` / `j` in the detail pane | Select a use-case field. |
| `enter` or `e` in the detail pane | Edit the selected use-case field. |
| `n` | Create a new use case entry. |
| `/` in the list pane | Filter use cases. |
| `enter` while filtering | Apply the filter. |
| `esc` while filtering | Cancel filter editing. |
| `esc` in the list pane | Clear the case filter. |
| `s` or `ctrl+s` | Save the selected use case. |
| `r` | Refresh the workspace graph from disk. |

Refresh is blocked while use-case edits are active or unsaved.

## Capture Dumps in the TUI

Open the Dumps view with `F3`.

Dumps are free-form Markdown context. They can be project-wide or tied to a use
case. Creating a dump writes a curated primary document plus a raw capture
sibling. The TUI lists curated dumps by use case, shows the selected curated
content in the detail pane, and shows the raw sibling path when one exists.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a dump. |
| `c` | Create a dump for the selected project or use case after confirmation. |
| `enter` | Edit the selected dump in the TUI editor. |
| `alt+e` | Open the selected dump in the external editor. |
| `r` | Refresh dumps from disk. |
| `ctrl+s` while editing | Save the TUI editor content. |
| `esc` while editing | Cancel the TUI editor. |

Direct edit maximizes the detail pane. External edit starts the configured
editor process and refreshes the dump list when the editor exits.

Equivalent CLI commands:

```sh
chc dump list
chc dump new --title "Billing notes" --content "Initial context."
chc dump new --title "Billing notes" --content "Use-case context." --use-case billing
chc dump show .chc/user/dumps/billing-notes.md
chc dump save .chc/user/dumps/billing-notes.md --content "# Billing notes"
chc dump append .chc/user/dumps/billing-notes.md --content "Additional context."
chc dump open .chc/user/dumps/billing-notes.md
```

## Run Quizzes in the TUI

Open the Quizzes view with `F4`.

Quizzes collect missing project or use-case knowledge through bounded questions.
The TUI shows quiz state, active question, recorded answers, and curated notes.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a quiz. |
| `s` | Start or restart the selected quiz. |
| `enter` | Answer the active question. |
| `r` | Refresh quiz state from disk. |

When answering a quiz:

| Key | Behavior |
| --- | --- |
| `enter` | Submit the answer. |
| `shift+enter` or `ctrl+j` | Insert a newline. |
| `alt+s` | Skip the active question. |
| `alt+d` | Mark the quiz done. |
| `esc` | Cancel answer entry. |

The command palette also exposes resume, skip, complete, cancel, deep mode, and
fast mode actions for the selected quiz.

Equivalent CLI commands:

```sh
chc quiz list
chc quiz start --use-case billing --mode deep
chc quiz resume --use-case billing
chc quiz answer --use-case billing --answer "Retries must be idempotent."
chc quiz skip --use-case billing
chc quiz complete --use-case billing
chc quiz cancel --use-case billing
chc quiz mode fast --use-case billing
```

## Generate Specs in the TUI

Open the Specs view with `F5`.

Specs are generated from accumulated project and use-case material: project
metadata, project context, use-case metadata, use-case context, dumps, quizzes,
and related workspace documents. Specs are generated from the Specs tab.

Spec generation writes the Markdown spec only. For behavior-testable use cases,
the `Generate acceptance feature` action derives a Gherkin `.feature` artifact
from the spec, quiz, and dump context.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Generate a spec for the selected use case. |
| `f` | Generate an acceptance feature for the selected use case. |
| `r` | Refresh specs from disk. |

You can also press `:` in the Specs view and run `Generate acceptance feature`
from the command palette. The TUI labels the action as an acceptance feature;
the generated artifact is a Gherkin feature file.

Spec and acceptance feature generation are asynchronous. While either is
running, the TUI shows a working state and ignores other spec actions until the
request finishes.

If acceptance feature generation succeeds, the status line shows the generated
`.chc/user/acceptance/<case>.feature` path. If more scenario detail is needed,
the status line shows the `.chc/user/acceptance/<case>-feature-round.md` path
so you can review the default-oriented prompts before trying again.

Current limitation: those feature-round prompts are persisted, but the TUI does
not yet collect the follow-up answers interactively. Until that answer flow is
implemented, the prompt answers must be added to the feature-round note before
retrying generation.

Equivalent CLI commands:

```sh
chc spec list
chc spec generate --use-case billing
chc acceptance feature generate --use-case billing
```

## Generate Plans and Tracking in the TUI

Open the Plans view with `F6`.

Plans are generated from accumulated project and use-case material, including
existing specs. A plan decomposes implementation into reviewable slices and is
the input for later tracking and ticket creation.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Generate the selected plan only when it is missing. |
| `G` | Regenerate the selected plan after confirmation. |
| `alt+t` | Create plan tracking and tickets from the selected plan after confirmation. |
| `r` | Refresh plans from disk. |

`g` does not overwrite an existing plan. Use `G` when the existing plan should
be replaced. Regeneration archives the existing generated plan before writing
the replacement, so archived generated plans remain available on disk but are
excluded from normal document loading.

`alt+t` is available only when the selected plan exists. It creates a tracking
Markdown file and one ticket per generated slice. If tracking already exists,
the operation reports that state instead of creating duplicate tracking.

Plan generation, regeneration, and tracking creation are guarded so the same
TUI cannot start overlapping plan operations.

Equivalent CLI commands:

```sh
chc plan list
chc plan generate --use-case billing
chc plan regenerate --use-case billing
chc plan tracking create --use-case billing
```

## Use the Board

Open the Board view with `F7`.

The Board view groups tickets by status and shows details for the selected
ticket. Normal Board columns are:

- Backlog: `backlog`
- Ready For Dev: `ready_for_dev`
- In Progress: `in_progress`
- Ready For Review: `ready_for_review`
- Ready For Merge: `ready_for_merge`
- Done: `done`

Legacy ticket statuses are still accepted. `open`, `ready`, `reviewing`,
`solved`, `merged`, and `closed` map into normal Board columns. `blocked` and
unknown statuses appear in the virtual attention view instead of the normal
column cycle.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a ticket. |
| `left` / `right` or `h` / `l` | Move between normal Board columns. |
| `alt+a` | Toggle the virtual attention view. |
| `alt+m` | Maximize the focused pane. In the ticket pane, Board maximization shows all normal columns; in attention mode it shows only attention tickets. |
| `/` | Filter tickets. |
| `enter` while filtering | Apply the filter and select the first match. |
| `esc` while filtering | Cancel filter editing. |

The compact Board shows one normal column at a time and marks the active column
in the header and column strip. The detail pane includes status metadata,
synopsis, source metadata, slice metadata, checklist counts, comment counts,
evidence strategy status, bounded Markdown body preview, checklist preview,
recent comment headings, attention reasons, and the ticket path when those
fields are available.

When a ticket contains the anchored evidence strategy checklist, Board detail
shows a concise summary:

- overall evidence state: `complete`, `incomplete`, `blocked`, or `malformed`;
- strategy, Gherkin, tests, validation, and evidence-summary item states.

Tickets without an evidence strategy checklist show `missing` in Board detail.
CLI list and board output keep missing evidence compact with `-`.

Ticket and work operations are available from the command palette:

| Command | Behavior |
| --- | --- |
| Create ticket | Create a default task ticket after confirmation. |
| Add ticket comment | Add the standard `TUI action recorded.` comment to the selected ticket. |
| Move ticket to ready | Move the selected ticket to `ready`. This legacy status maps to the Ready For Dev Board column. |
| Move ticket to in_progress | Move the selected ticket to `in_progress`. |
| Move ticket to reviewing | Move the selected ticket to `reviewing`. |
| Start selected ticket | Run the configured runtime harness for the selected ticket. |
| Start ready ticket loop | Run one ready ticket through the agent loop. |
| Run all visible ready tickets | Run the Dash ready queue with the selected filters and order. |
| Apply safe readiness repair | Apply a non-destructive readiness repair when Dash reports one. |
| Request agent stop | Request daemon stop through local daemon state. |

Checklist toggle and checklist item creation commands are visible as capability
checks, but they currently return an unavailable-operation message because the
safe checklist writer has not been implemented.

Equivalent CLI commands:

```sh
chc ticket board
chc ticket list
chc ticket show CHC-1
chc ticket new "Add completed task support" --synopsis "Add task completion support."
chc ticket comment CHC-1 "TUI action recorded."
chc ticket status CHC-1 ready
chc ticket status CHC-1 in_progress
chc ticket status CHC-1 reviewing
chc ticket start CHC-1
chc ticket loop --status ready --max-iterations 1
chc work run-all
chc work doctor --fix-safe
chc daemon stop
```

## Inspect Operations

Open the Dash view with `F8`.

Dash shows agent activity and runtime evidence loaded from local state. Use it
before work to inspect readiness, during work to watch the event stream, and
after work to inspect recent execution or loop outcomes.

Dash shows:

- work readiness and safe repair hints;
- ready queue filters and ordering;
- current loop or ticket execution state;
- current or latest integration phase and event message;
- recent human-readable agent events;
- ticket, execution report, runtime log, and validation log paths when the
  event stream includes them.
- recent execution and loop outcomes.

When readiness is blocked, Dash prevents ticket execution and run-all actions
before the runtime worker starts. Press `x` to apply a safe repair when Dash
offers one, confirm the repair, and retry the work action.

For a CLI-first inspection flow, see
[Observe agent work](../observe-agent-work.md).

## Get Help in the TUI

Open Help with `F9`, or use `:` and run the Help command. The footer at the
bottom of the TUI always shows the active key helper for the current view or
modal state.

## Create and Manage Tickets

Create a ticket:

```sh
chc ticket new "Add completed task support" \
  --synopsis "Add task completion support." \
  --type feature \
  --priority high
```

Useful metadata:

```sh
chc ticket new "Fix task ID reuse" \
  --synopsis "Fix task id reuse in the runtime path." \
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

Ticket files remain Markdown under `.chc/user/tickets/`.

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

Execution reports summarize the runtime plan, result status, validation status,
changed files, and artifact paths for one execution.

When evidence strategy report copying is enabled, execution reports also copy
the ticket evidence strategy and final evidence summary so the checklist and
runtime trace can be inspected together.

## Review a Ticket Pull Request

Ticket review is optional and configured per core. It uses normal Board
statuses and ticket metadata; it does not add a separate Board column.

With default evidence strategy settings, a ticket must have completed evidence
before `ready_for_review`. Use Board detail or `chc ticket show <key>` to inspect
the checklist summary before applying a result or requesting review.

Inspect the configured mode:

```sh
chc ticket review-mode status
```

Check provider access for a pull request:

```sh
chc ticket review-mode diagnostics --url https://forge.example.test/owner/repo/pulls/123
```

Review a ticket after its branch and pull request are recorded in the ticket:

```sh
chc ticket review CHC-1 --result pass
```

Record requested changes:

```sh
chc ticket review CHC-1 \
  --result needs_changes \
  --finding "Handle the nil input path before merge."
```

When review requests changes, the ticket returns to `ready_for_dev` with repair
handoff metadata. In `agentic_review` loop mode, ChrysoCoord runs one default
repair execution for that same ticket, reviews it again, and only then merges
or blocks. Later normal tickets do not start while that repair cycle is
unfinished.

For the Forgejo setup flow, see
[Use Forgejo ticket review](../use-forgejo-ticket-review.md).

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

Loop reports summarize loop status, ticket selection, execution outcomes, and
stop reasons.

## Run All Ready Agent Work

Preview the ready queue without execution:

```sh
chc work run-all --dry-run --order priority_severity
```

Run all matching ready tickets:

```sh
chc work run-all --order priority_severity
```

Filter by use case, priority, severity, project area, tag, type, or status:

```sh
chc work run-all --dry-run --use-case billing --priority high --order key
```

`work run-all` uses the same readiness policy and execution loop as Dash. In
serialized integration modes, it completes the current ticket's development,
review policy, repair when applicable, and merge or block before selecting the
next normal ticket. It leaves loop, execution, report, log, and event-stream
evidence under the configured runtime artifact directory.

Inspect the evidence:

```sh
chc report list --kind loop
chc report list --kind execution
chc work logs <execution-id>
chc work events .chc/user/runtime/runs/<loop-id>/events.jsonl
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
chc author session show <session-handle>
```

The session handle can be the exact id, exact topic, or a unique topic slug.

Create a context-aware questionnaire:

```sh
chc author quiz new --target spec --topic "Billing sync"
```

Preview the questionnaire first:

```sh
chc author quiz new --target spec --topic "Billing sync" --dry-run
```

After editing quiz answers, use them to create or refine a spec. Keep the spec
in `status: draft` while it still needs review.

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
  --session <session-handle> \
  --fake-question "What should the spec include?"
```

Submit the answer:

```sh
chc author assisted answer \
  --session <session-handle> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id <question-id> \
  --question "What should the spec include?" \
  --answer "It should include retry behavior and idempotent writes."
```

Check progress:

```sh
chc author assisted status --session <session-handle>
```

Finalize the session into a curated summary and target artifact:

```sh
chc author assisted finalize \
  --session <session-handle> \
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

## Configure Notifications and Remote Command Intake

Inspect notification configuration:

```sh
chc notify status
```

Preview a Telegram setup-check payload without contacting Telegram:

```sh
chc notify telegram test --dry-run
```

Send the setup-check message when Telegram is configured:

```sh
chc notify telegram test
```

Parse Telegram update fixtures into remote command handoff records:

```sh
chc notify telegram poll --fixture updates.json --dry-run
chc notify telegram poll --fixture updates.json
```

Remote command intake is bounded. It stores accepted or rejected handoff records
for supported intents such as status, loop start, ticket start, and ticket
comment. It does not expose arbitrary shell command execution.

## Inspect and Recover Local State

Write a local diagnostic bundle:

```sh
chc maintenance bundle
```

Inspect database, registry, execution artifact, and cleanup-preview state:

```sh
chc maintenance db status
chc maintenance registry status
chc maintenance execution artifacts
chc maintenance execution cleanup-preview
```

Repair moved or stale registry entries explicitly:

```sh
chc maintenance registry disable example
chc maintenance registry set-path example /path/to/project
```

Maintenance previews and diagnostics are read-only unless the command name
describes an explicit repair action.

For a full recovery sequence, see
[Inspect and recover local state](../inspect-and-recover-local-state.md).

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
`/workspace/.chc/user/runtime/runs/...`.

## Current Boundaries

- Context composition runs outside Docker.
- Docker is the runtime sandbox boundary.
- Project mode mounts the user-visible checkout into Docker.
- Worktree mode prepares a ticket-specific Git worktree and mounts that
  execution workspace into Docker.
- Avoid editing the mounted execution workspace while an agent run is active.
- PAT-backed push from inside the container is not part of the current local
  runtime boundary.
- Assisted authoring prepares Markdown documents; it does not execute tickets or
  run the Docker worker.
- Assisted authoring does not include implicit conversation history. Relevant
  prior discussion must be captured in Markdown artifacts or transcript events.

## Reference

- [Project workspace](../../reference/project-workspace.md)
- [Work readiness](../../reference/work-readiness.md)
- [Agent work observability](../../reference/agent-work-observability.md)
- [TUI](../../reference/tui.md)
- [Authoring sessions](../../reference/authoring-sessions.md)
- [Project CLI](../../reference/project-cli.md)
- [Ticket execution](../../reference/ticket-execution.md)
- [Ticket loop](../../reference/ticket-loop.md)
- [Ticket result application](../../reference/ticket-result-application.md)
- [Reports](../../reference/reports.md)
- [Notifications](../../reference/notifications.md)
- [Remote commands](../../reference/remote-commands.md)
- [Daemon coordination](../../reference/daemon-coordination.md)
- [Maintenance](../../reference/maintenance.md)
