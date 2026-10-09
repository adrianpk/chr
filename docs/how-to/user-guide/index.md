# User Guide

This guide covers the Chryso workflow that exists today.

Chryso is local-first. A managed project has a `.chr/` directory, while
user-level state stores project registration, indexes, execution records, loop
records, and daemon state.

## Initialize a Workspace

Create project-local Chryso state:

```sh
chr init
```

Use a different profile when needed:

```sh
chr init --profile minimal
chr init --profile personal
```

Preview initialization:

```sh
chr init --dry-run
```

Initialize another project directory:

```sh
chr init --dir /path/to/project
```

The initialized workspace includes ticket directories, project protocol files,
minimal user-owned config, `.chr/default/`, and
`.chr/default/runtime/agent/Dockerfile`.

All three profiles initialize new projects with `workflow.version: 1`. Its
defaults select ticket authority, isolated worktrees, trusted automatic integration
and one worker. Use an explicit `review.mode: human_review` when the project
requires human approval. See [Workflow configuration](../../reference/workflow-configuration.md)
for effective values, sources and limits.

Repeated initialization preserves existing configuration. In an existing `.chr/`
workspace with no config file, it reports `preserve absent .chr/default/config.yml`
and leaves legacy policy in effect. Runtime initialization installs runtime assets
without changing project policy.

Set an active operator id for workflows that need project-local userspace
ownership:

```sh
export CHR_OPERATOR_ID=adrian
```

With direnv, the same export may live in `.envrc`.

## Configure Runtime Execution

Runtime defaults are owned by `chr`. A new project does not need to repeat
default runtime paths in `.chr/default/config.yml`.

Override runtime settings only when the project has a real reason to do so:

```yml
runtime:
  image: chr-agent-runtime:local
  dockerfile: .chr/default/runtime/agent/Dockerfile
  validation_command: make verify
```

Install missing runtime defaults in an existing workspace:

```sh
chr runtime init
```

Preview runtime initialization:

```sh
chr runtime init --dry-run
```

Validate the runtime boundary:

```sh
chr runtime validate
```

Preview validation without running Docker:

```sh
chr runtime validate --dry-run
```

## Configure Work Readiness

Work readiness protects the project before Chryso starts agent work. The
default policy expects a clean `dev` branch with `origin` as the canonical
remote, but projects can configure their own branch posture in
`.chr/default/config.yml`.

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
chr work doctor
```

Apply safe repair when Chryso reports one:

```sh
chr work doctor --fix-safe
```

## Inspect Workflow Policy

Inspect project workflow defaults and explicit values before starting work:

```sh
chr work config
chr work config --json
```

The nested `workflow.version: 1` marker selects tickets authority, isolated
worktrees and trusted automatic review, with one active worker. Projects without
that marker retain legacy defaults and explicit policy. Reading the config does
not upgrade a project. See [Workflow configuration](../../reference/workflow-configuration.md)
for accepted keys, bounds and value sources.

To require human approval for a native invocation under automatic project policy,
start its extent with `--human-review`:

```sh
chr work slice --human-review slice-1
```

That restriction stays with the retained invocation. Execution limits freeze at
admission; publication rechecks the current project policy and exact approval
binding.

## Register and Select Projects

Register a managed project:

```sh
chr project add /path/to/project
```

Override the registry name:

```sh
chr project add --name api-service /path/to/project
```

List registered projects:

```sh
chr project list
```

Select the active project:

```sh
chr project select api-service
```

Print the active project:

```sh
chr project current
```

All project registry commands support `--dev-root` for isolated development
state.

## Open the TUI Console

Open the local operational console from the resolved project:

```sh
chr
```

The explicit form is:

```sh
chr tui
```

Open a specific project directory:

```sh
chr tui --dir /path/to/project
```

The TUI is the main interactive surface for project context, use cases, dumps,
quizzes, specs, plans, tickets, and execution activity. It loads durable
workspace files from disk and writes normal Markdown artifacts back to the
core.

## Open the Local Web UI

The Web UI is a browser surface over the same project state. It does
not start from `chr`, `chr tui`, ticket loops, agent work, or other normal
commands.

Start it explicitly from a managed project:

```sh
chr web serve
```

Open:

```text
http://127.0.0.1:7331
```

Preview the resolved server configuration without listening on a socket:

```sh
chr web serve --dry-run
```

Open a specific project directory:

```sh
chr web serve --dir /path/to/project
```

The browser UI can manage the Board and the same project artifact families used
by the TUI and CLI: Cases, Dumps, Quizzes, Specs, Plans, Board, and Dash. It
uses project-local files as the source of truth and reloads from shared state
after CLI, TUI, or Web UI changes.

This first Web UI iteration is intended to be usable, not final. Expect visual
and ergonomic refinement to continue while the local server, shared state, and
action boundaries remain explicit.

Keep it local unless you deliberately expose it through operator-owned private
networking or a tunnel. See [Run the local Web
UI](../run-local-web-ui.md) and [Expose the local Web UI
deliberately](../expose-local-web-ui.md).

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

Use the description as the short human-facing summary for lists, selectors, and
previews. Use the context field for longer use-case material that should be
included when Chryso builds the total agent context: constraints, decisions,
intent, edge cases, and notes that do not belong in a one-line summary.

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
case. A dump is a document type for material the user already knows or
remembers; quizzes complement dumps by prompting for missing or latent context.
Creating a dump writes a curated primary document plus a raw capture sibling.
The TUI lists curated dumps by use case, shows the selected curated content in
the detail pane, and shows the raw sibling path when one exists.

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
chr dump list
chr dump new --title "Billing notes" --content "Initial context."
chr dump new --title "Billing notes" --content "Use-case context." --use-case billing
chr dump show ops/default/dump/00000-billing-notes/artifact.md
chr dump save ops/default/dump/00000-billing-notes/artifact.md --content "# Billing notes"
chr dump append ops/default/dump/00000-billing-notes/artifact.md --content "Additional context."
chr dump open ops/default/dump/00000-billing-notes/artifact.md
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
chr quiz list
chr quiz start --use-case billing --mode deep
chr quiz resume --use-case billing
chr quiz answer --use-case billing --answer "Retries must be idempotent."
chr quiz skip --use-case billing
chr quiz complete --use-case billing
chr quiz cancel --use-case billing
chr quiz mode fast --use-case billing
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
`ops/default/spec/00000-<case>-acceptance-feature/artifact.md` path. If more scenario detail is needed,
the status line shows the `ops/default/spec/00000-<case>-acceptance-feature-round/artifact.md` path
so you can review the default-oriented prompts before trying again.

Current limitation: those feature-round prompts are persisted, but the TUI does
not yet collect the follow-up answers interactively. Until that answer flow is
implemented, the prompt answers must be added to the feature-round note before
retrying generation.

Equivalent CLI commands:

```sh
chr spec list
chr spec generate --use-case billing
chr acceptance feature generate --use-case billing
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
chr plan list
chr plan generate --use-case billing
chr plan regenerate --use-case billing
chr plan tracking create --use-case billing
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
chr ticket board
chr ticket list
chr ticket show CHR-1
chr ticket new "Add completed task support" --synopsis "Add task completion support."
chr ticket comment CHR-1 "TUI action recorded."
chr ticket status CHR-1 ready
chr ticket status CHR-1 in_progress
chr ticket status CHR-1 reviewing
chr ticket start CHR-1
chr ticket loop --status ready --max-iterations 1
chr work run-all
chr work doctor --fix-safe
chr daemon stop
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
chr ticket new "Add completed task support" \
  --synopsis "Add task completion support." \
  --type feature \
  --priority high
```

Useful metadata:

```sh
chr ticket new "Fix task ID reuse" \
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
chr ticket list
```

Show one ticket:

```sh
chr ticket show CHR-1
```

Append a comment:

```sh
chr ticket comment CHR-1 "Validation fails in the completion path."
```

Update status:

```sh
chr ticket status CHR-1 ready
```

Ticket files remain Markdown under `.chr/default/tickets/`.

## Keep the Ticket Index Current

Rebuild the ticket index:

```sh
chr index rebuild
```

Check index status:

```sh
chr index status
```

Ticket commands refresh the index where they need indexed ticket data, but the
index commands are useful when inspecting drift explicitly.

## Run Independent Ticket Work

For a project with `work.state_authority: tickets`, preview an exact extent
before starting it:

```sh
chr work ticket --preview CHR-1
chr work slice --preview slice-1
chr work slice-set --preview 00001-delivery
chr work use-case --preview delivery
chr work loose --preview --priority high
```

Run the selected extent by omitting `--preview`. Save its printed execution ID
to inspect progress and operate on that retained selection:

```sh
chr work status execution-id
chr work continue --max-steps 1 execution-id
chr work pause execution-id
chr work resume execution-id
chr work cancel execution-id
```

`status` distinguishes task validation from confirmed ticket integration. When
human review is required, read the displayed approval binding and pass that
exact hash to `chr work approve --binding <hash> execution-id`. Approval grants
the current candidate; subsequent continuation handles publication through the
configured policy. A changed candidate, PR head, target or policy rejects stale
approval. Cancellation is terminal.

Every operation accepts `--preview` to observe its retained inputs without
execution or state changes. See [Workflow controls](../../reference/workflow-controls.md)
for pagination, JSON output, reconciliation and manual harness commands.

In the TUI command palette, select native work for a ticket, its slice or slice
set, the selected use case, or independent tickets. Press Enter to inspect the
preview, then Enter again to confirm its exact snapshot and selected revision.
The Dash distinguishes validated tasks from confirmed integration and shows
retained evidence, waits and blockers. Inspect retained workflow status and use
the next-workflow or next-page commands to select an invocation. Continue,
pause/cancel, resume, exact-binding approval and ownership reconciliation use
that invocation's displayed revision. A stale action shows its failure with
current status; inspect again before retrying. Status and stop controls remain
available during native execution.

In the Web UI, open a native ticket on the Board to preview its ticket, slice,
set or use-case extent. Case details offer a use-case preview. Confirm the
displayed snapshot to start one checkpoint, then use its retained workflow view
to continue, pause/cancel, resume, approve the exact binding or reconcile
ownership. The Dash links to retained workflows and loose-work filters. A stale
action shows current progress alongside its failure. Refresh status before
retrying; host validation and confirmed integration remain separate counts.


Ticket-native work defaults to one active worker. To run independent lanes in
parallel, set the capacity in `.chr/default/config.yml`:

```yaml
work:
  max_active_workers: 2
  ready_queue_capacity: 64
```

The worker limit accepts 1–8. The ready queue accepts 1–1024 unfinished
integration lanes and defaults to 64. Use the same values in every linked host
checkout; ops namespaces share the pool. Each worker edits its own isolated
checkout. Task order within a slice and ordered set/use-case progress stay
serial, and competing publications wait for the canonical target.

If an invocation reports `waiting_capacity`, retry its claim or continuation
after capacity is available. Its selection, evidence and attempt budget remain
intact. A failed independent lane does not stop another lane. A surviving child
still occupies capacity after its lease expires. Legacy execution and adoption
remain exclusive across the pool.

This applies to explicitly enabled ticket-native authority on the registered
local coordinator. See [Workflow execution](../../reference/workflow-execution.md)
for ownership, recovery and manual-runner operations.

## Run One Ticket

Preview execution:

```sh
chr ticket start CHR-1 --dry-run
```

Run execution:

```sh
chr ticket start CHR-1
```

Execution uses the configured Docker runtime. The worker receives
`CHR_WORKER_CONTEXT_PATH`, reads `context.md`, writes `result.json`, and leaves
runtime and validation logs under the execution run directory.

Apply an execution result to the ticket:

```sh
chr ticket apply-result <execution-id>
```

Preview result application:

```sh
chr ticket apply-result <execution-id> --dry-run
```

Inspect execution evidence:

```sh
chr report list --ticket CHR-1
chr report show --execution <execution-id>
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
before `ready_for_review`. Use Board detail or `chr ticket show <key>` to inspect
the checklist summary before applying a result or requesting review.

Inspect the configured mode:

```sh
chr ticket review-mode status
```

Check provider access for a pull request:

```sh
chr ticket review-mode diagnostics --url https://forge.example.test/owner/repo/pulls/123
```

Review a ticket after its branch and pull request are recorded in the ticket:

```sh
chr ticket review CHR-1 --result pass
```

Record requested changes:

```sh
chr ticket review CHR-1 \
  --result needs_changes \
  --finding "Handle the nil input path before merge."
```

When review requests changes, the ticket returns to `ready_for_dev` with repair
handoff metadata. In `agentic_review` loop mode, Chryso runs one default
repair execution for that same ticket, reviews it again, and only then merges
or blocks. Later normal tickets do not start while that repair cycle is
unfinished.

For the Forgejo setup flow, see
[Use Forgejo ticket review](../use-forgejo-ticket-review.md).

## Run a Foreground Ticket Loop

Preview eligible tickets:

```sh
chr ticket loop --dry-run
```

Filter the loop:

```sh
chr ticket loop --dry-run \
  --status ready \
  --priority high \
  --tag runtime \
  --order priority_severity \
  --max-iterations 3
```

Run the foreground loop:

```sh
chr ticket loop --max-iterations 3
```

The loop executes eligible tickets one at a time and stops on the first blocked
or failed ticket.

Inspect loop evidence:

```sh
chr report list --kind loop
chr report show --loop <loop-id>
```

Loop reports summarize loop status, ticket selection, execution outcomes, and
stop reasons.

## Run All Ready Agent Work

Preview the ready queue without execution:

```sh
chr work run-all --dry-run --order priority_severity
```

Run all matching ready tickets:

```sh
chr work run-all --order priority_severity
```

Filter by use case, priority, severity, project area, tag, type, or status:

```sh
chr work run-all --dry-run --use-case billing --priority high --order key
```

`work run-all` uses the same readiness policy and execution loop as Dash. In
serialized integration modes, it completes the current ticket's development,
review policy, repair when applicable, and merge or block before selecting the
next normal ticket. It leaves loop, execution, report, log, and event-stream
evidence under the configured runtime artifact directory.

Inspect the evidence:

```sh
chr report list --kind loop
chr report list --kind execution
chr work logs <execution-id>
chr work events ops/default/run/<loop-artifact-id>/evidence/events.jsonl
```

## Capture Authoring Context

Create an authoring session:

```sh
chr author session new --target spec --topic "Billing sync"
```

List authoring sessions:

```sh
chr author session list
```

Show one session:

```sh
chr author session show <session-handle>
```

The session handle can be the exact id, exact topic, or a unique topic slug.

Create a context-aware questionnaire:

```sh
chr author quiz new --target spec --topic "Billing sync"
```

Preview the questionnaire first:

```sh
chr author quiz new --target spec --topic "Billing sync" --dry-run
```

After editing quiz answers, use them to create or refine a spec. Keep the spec
in `status: draft` while it still needs review.

Generated authoring artifacts remain Markdown under `.chr/`. The commands
initialize and connect artifacts; they do not replace direct editing.

Run an assisted authoring session when you want the CLI to drive one question at
a time while preserving every step as Markdown:

```sh
chr author assisted start --target spec --topic "Billing sync"
```

Ask the next question:

```sh
chr author assisted resume \
  --session <session-handle> \
  --fake-question "What should the spec include?"
```

Submit the answer:

```sh
chr author assisted answer \
  --session <session-handle> \
  --turn turn-001 \
  --turn-index 1 \
  --question-id <question-id> \
  --question "What should the spec include?" \
  --answer "It should include retry behavior and idempotent writes."
```

Check progress:

```sh
chr author assisted status --session <session-handle>
```

Finalize the session into a curated summary and target artifact:

```sh
chr author assisted finalize \
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
chr daemon status
```

Run one foreground coordination iteration:

```sh
chr daemon start --max-iterations 1
```

Request stop:

```sh
chr daemon stop
```

The daemon continues retained ticket-native invocations in enabled registered
projects. First authorize the exact extent through `chr work` or a confirmed
TUI/Web preview. Startup does not admit the ready backlog. Each project iteration
attempts at most 16 runnable invocations and one checkpoint per invocation.

Inspect or stop one retained invocation:

```sh
chr daemon workflow status --dir /path/to/project --json execution-id
chr daemon workflow pause --dir /path/to/project execution-id
chr daemon workflow cancel --dir /path/to/project execution-id
chr daemon workflow resume --dir /path/to/project execution-id
```

These controls use the same lifecycle and evidence as `chr work`. Operator pauses,
terminal cancellations and human review waits remain in place across daemon
startup. Manual execution appears in the same status views after host validation.

`daemon stop` yields between checkpoints and preserves invocation authorization.
Use the invocation's pause/cancel control to stop its owned worker. Read
[Daemon coordination](../../reference/daemon-coordination.md) for limits, recovery
and coordinator state.

## Configure Notifications and Remote Command Intake

Inspect notification configuration:

```sh
chr notify status
```

Preview a Telegram setup-check payload without contacting Telegram:

```sh
chr notify telegram test --dry-run
```

Send the setup-check message when Telegram is configured:

```sh
chr notify telegram test
```

Parse Telegram update fixtures into remote command handoff records:

```sh
chr notify telegram poll --fixture updates.json --dry-run
chr notify telegram poll --fixture updates.json
```

Remote command intake is bounded. It stores accepted or rejected handoff records
for supported intents such as status, loop start, ticket start, and ticket
comment. It does not expose arbitrary shell command execution.

Telegram notifications remain informational. Intake handoffs do not execute
ticket-native work controls or grant approval. Invoke admission, continuation,
pause/cancel and reconciliation through the CLI, TUI, Web or daemon controls;
inspect their retained evidence to establish progress.

## Inspect and Recover Local State

Write a local diagnostic bundle:

```sh
chr maintenance bundle
```

Inspect database, registry, execution artifact, and cleanup-preview state:

```sh
chr maintenance db status
chr maintenance registry status
chr maintenance execution artifacts
chr maintenance execution cleanup-preview
```

Repair moved or stale registry entries explicitly:

```sh
chr maintenance registry disable example
chr maintenance registry set-path example /path/to/project
```

Maintenance previews and diagnostics are read-only unless the command name
describes an explicit repair action.

For a full recovery sequence, see
[Inspect and recover local state](../inspect-and-recover-local-state.md).

## Proving-Ground Workflows

The proving-ground commands exercise direct sandbox behavior against controlled
objectives:

```sh
chr proving validate --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md
chr proving run --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md --dry-run
chr proving ralph --objective .chr/proving/objectives/TODO-1-fix-task-id-reuse.md --max-attempts 3
```

Use these commands for local runtime proofing, not as the main ticket workflow.

## Execution Context

Before agent work, Chryso can compose a context package with:

- agent identity;
- applicable `AGENTS.md` and `.chr/` protocol files;
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
`/workspace/ops/default/run/...`.

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
