# User Guide

This guide covers the chr workflow that exists today.

For native plan execution, retained controls, supervised manual handoffs and
optional tracker export/repair, follow
[Run ticket-native workflows](../run-ticket-workflows.md).

chr is local-first. A managed project has a `.chr/` directory, while
user-level state stores project registration, indexes, execution records, loop
records, and daemon state.

## Initialize a Workspace

Create project-local chr state:

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
defaults select ticket authority, isolated worktrees, trusted automatic review
and one worker. Native publication stays disabled until explicitly selected;
[Configure native integration](../configure-native-integration.md) explains the
source rewrite and support checks before work admission. Use an explicit
`review.mode: human_review` when the project
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

## Select Worker and Generation Harnesses

Projects may use Codex or Pi independently for workers and content generation.
Prepare named profiles in private user configuration, select the registered
project with `chr project use`, and inspect `chr harness status --json` before
changing either default. The [harness setup and recovery guide](../select-harnesses.md)
covers native Pi generation, the separate Docker worker bridge, credentials,
custom endpoints and exact revision/request handling.

Web Harnesses, TUI Project `h`, Telegram `/harness` and the CLI share those
persisted defaults. A switch affects new operations; pending, running, paused
and resumed work keeps its original binding. Local readiness does not test
remote authentication or certify the selected model's tool/structured-output
capabilities. Missing usage remains unknown, and observed usage does not certify
hard token-budget enforcement.

## Configure Work Readiness

Work readiness protects the project before chr starts agent work. The
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

Apply safe repair when chr reports one:

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

To upgrade an existing project, retain the revision from `work config --json`
and choose its authority explicitly:

```sh
chr work upgrade --config-revision <revision> --state-authority tickets --json
```

The operation requires quiescent work, saves the previous effective policy and
preserves its review mode and explicit native settings. An absent publication
profile stays disabled. Supply `--review-mode` only when deliberately changing
that choice. Legacy records remain intact unless `--adopt-file` selects one
approved plan and its exact task map. Retry an interrupted upgrade with the same
revision, choices and file; see the configuration reference for the envelope and
reconciliation rules.

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

In **Project** (`F1`), press `h` to inspect harness defaults. Tab selects the
worker or generation role; arrows choose a named profile and `c` restores legacy
dispatch for that role. Press Enter to save or Esc to discard unsent choices.
Stale or failed saves retain the choices. Press `r` to discard them and reload
current defaults, including changes from another interface or process. Running,
paused and resumed work keeps its original binding. See [TUI harness defaults](../../reference/project-workspace.md#tui-harness-defaults)
for readiness, retry and scrolling controls.

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

For token-free browser access from another computer on a trusted internal
network, run `chr web serve --allow-lan`. LAN access defaults to disabled;
`web.allow_lan: true` in the project's configuration retains the explicit
choice. See [Expose the local Web UI](../expose-local-web-ui.md) for bind,
token-policy and origin behavior.

The browser UI can manage the Board and the same project artifact families used
by the TUI and CLI: Cases, Dumps, Quizzes, Specs, Plans, Board, and Dash. It
uses project-local files as the source of truth and reloads from shared state
after CLI, TUI, or Web UI changes.

Open **Harnesses** in the project context to choose named worker and generation
defaults. Each role shows its local readiness separately from retained worker
invocations. Select **Save defaults** to persist choices without restarting the
service, or **Restore legacy dispatch** to clear a role. These changes apply to
new operations; running, paused and resumed work keeps its original binding.
Rejected saves retain your choices. Reload current defaults before retrying a
stale form. See [Web harness defaults](../../reference/project-workspace.md#web-harness-defaults)
for readiness, retry and private-setup boundaries.

Open a named case to generate its plan. The planning panel shows Generate Plan
when no plan exists and Regenerate Plan for the displayed existing revision.
Save case edits first: planning uses saved context. Follow Add discovery context
or Answer this case's quiz when more information is needed. The Plans catalog
has no separate slug-based generation forms.

Plans shows saved plan documents as fixed-size cards with opening body lines.
Open a card to read that exact document in full. The detail has no manual Edit,
Save or Append controls. Internal order trackers stay in project files and
cannot be browsed or created through the Web UI.

Planning requires useful discovery and enough information about the intended
outcome, scope, constraints and verification. A complete dump alone, a short
focused dump or sufficient quiz answers can qualify. A manual spec alone does
not replace this discovery requirement. Missing information appears in the case
panel; generation does not fill material gaps with assumptions. Spec generation
and approval remain separate operations, and generating a plan does not start
work. Failed preparation or changed sources leave the current plan in place;
reload the case before retrying a stale operation.

Open **Dash** to observe the queue, each active ticket context and retained
results. The queue sits at left, current work at upper right and activity/results
below; these panels stack on narrow screens. Use the queue page links for more
than 128 pending items. Readiness and exclusions explain why pending work cannot
start. Native host-validated tasks and confirmed tickets are separate counts;
a successful legacy harness run does not itself confirm integration.

Dash refreshes while the tab is visible. Quiet validation remains active even
when there is no new event. Unchanged observations keep the displayed content;
changed observations retain scroll position, focused links and selected text
that still exists. If a read fails, a stale notice preserves the last observation
and its time until recovery. Reopening Dash reads retained state without
starting work. Preview Work and Stop Daemon keep their confirmations, with
results below the controls while monitoring remains visible. Stop Daemon requests
a daemon stop; it does not cancel an individual ticket.

This first Web UI iteration is intended to be usable, not final. Expect visual
and ergonomic refinement to continue while the local server, shared state, and
action boundaries remain explicit.

Keep it local unless you deliberately expose it through operator-owned private
networking or a tunnel. See [Run the local Web
UI](../run-local-web-ui.md) and [Expose the local Web UI
deliberately](../expose-local-web-ui.md).

## Complete a Web Workflow

1. Capture project context in Dumps. Choose a named case for details that belong
   only to that case; later Project-wide captures extend the shared document.
2. Open Quizzes to answer unresolved questions. Choose Deep for adaptive
   follow-ups or Fast for a retained flat batch. Done preserves earlier answers.
3. Create or edit individual Specs. Manual specs need no discovery or model
   access; editing a generated spec uses the same editor.
4. Open the named case to Generate or Regenerate its plan from saved, sufficient
   discovery. Missing context produces guidance and leaves an earlier plan intact.
5. Read the full result from Plans. Plan generation does not approve execution
   or allocate task tickets. Internal order maps remain outside Web browsing;
   approved native materialization binds their tickets through the workflow owner.
6. Open Board to edit allowed ticket metadata or add comments. Execution-owned
   fields explain their guards; pending edits remain separate from comments.
7. Open Dash to observe the same ticket identities, current work and retained
   outcomes. Opening or refreshing a view does not start work.

## Manage Tickets in the Web UI

Board has six status lanes and a separate attention area for blocked or unknown
states. Larger projects expose page links; horizontal lane scrolling keeps cards
readable on narrow screens. Choose **New** for an empty ticket form or open a card
to edit the selected ticket. Maximize and restore keep the same pending input.

The form groups classification, ownership and delivery fields with one description
editor. Save applies the displayed revision. Status selection and dragging reject
unsupported or stale moves with an explanation; reload before replacing newer
changes. Existing tickets have an independent comment composer, which does not
save or discard pending metadata.

Approved native plan order owns bound task titles, membership and case association;
the execution contract owns their description. These fields show their guard in
the editor. Use the existing approved amendment operation for contract changes.
Active claims and verified delivery receipts retain their own editing restrictions.
Unbound titles and other permitted metadata remain editable.

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
included when chr builds the total agent context: constraints, decisions,
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

## Capture Project Context in the Web UI

Open **Dumps**, select **New**, and choose **Project-wide** or an existing case
name. Enter a title and content, then select **Add context**. The first
project-wide capture creates the shared dump; later captures extend that same
document, including captures made through CLI, TUI or authoring conversations.
Case captures remain separate documents associated with the selected case.

The catalog fits more fixed-size cards across wider windows. Short or long
content keeps the same card and inset preview size; previews show only opening
text. Narrow windows reduce the column count. Open a card for the full body.

A failed request retains the form. Keep the same fields when retrying an uncertain
request; changing them submits a new capture. If the capture is committed but its
catalog refresh is pending, the form shows that result and an unchanged retry
refreshes the view without capturing the material again.

Raw captures and earlier body versions remain preserved. Conflicting older
general dumps require explicit reconciliation; opening a catalog never merges or
rewrites them. See the [authoring reference](../../reference/authoring-sessions.md#project-dump-reconciliation).

## Edit Dumps in the Web UI

Open a dump card, edit its **Body**, and select **Save**. Correct, remove,
reorganize or add text in that one editor. Save replaces the body of the same
document; frontmatter, raw captures and earlier versions remain preserved.

A failed or uncertain request retains your edits. Retry Save with the same body
to recover a confirmed result. If someone changed the document after you opened
it, Save reports a conflict and keeps your text. **Reload stored content** asks
before discarding your pending edits and loading the current body. Catalog
refreshes keep an open editor intact, and the full-page editor does not refresh
automatically. A pending catalog refresh after Save is reported as already saved;
an unchanged retry refreshes the view.

## Capture Dumps in the TUI

Open the Dumps view with `F3`.

Dumps are free-form Markdown context. They can be project-wide or tied to a use
case. A dump is a document type for material the user already knows or
remembers; quizzes complement dumps by prompting for missing or latent context.
The first project-wide capture creates a curated document and raw capture. Later
project-wide captures extend that document and retain their raw material. A
use-case capture creates an independent scoped document.
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

## Answer Quizzes in the Web UI

Quizzes shows Project Quiz first, including **No answers yet** when empty, then
named cases with recorded answers. Each fixed-size card previews the opening
question and answer. Open a card to read its history or continue unfinished work.

**New Quiz** appears only while a project or existing case has no recorded answers.
Open its compact menu and choose a name. An active selection keeps its current
question; a cancelled session offers **Resume**. A new scope offers **Start Quiz**.
After Done, **Start new session** explicitly creates another session and retains
earlier entries. Skipped or completed empty quizzes remain pending. When every
scope has an answer, New Quiz disappears; adding a case makes it available again.

The focused screen shows the current question, one answer field and the Deep/Fast
selector. Recorded answers appear newest first; skipped questions are excluded.

Answer records the response and advances. Skip advances without an answer. Done
completes the session, including when questions remain; Cancel leaves it cancelled
and preserves earlier answers. With an unsent answer, Skip, Cancel, Done and links
away from the screen require an explicit discard decision.

Deep generates adaptive follow-ups within the question/depth budget. Fast generates
one successful flat batch and then advances without further per-question model
calls. Switching modes preserves the active question and unsent answer. Deep
suspends the retained Fast queue; returning to Fast resumes that queue. When Fast
is exhausted, choose Done or switch to Deep and Resume to continue.

Unsent drafts survive refresh and mode changes in the current browser tab. A
successful Answer clears only its question's draft. Request failures keep the
input available for retry. If another client changes the question, reload the
current state; an earlier unsent draft remains separate from the new answer field.
If question generation fails after an answer, the recorded answer stays in history
and Resume retries continuation without recording it again.

## Run Quizzes in the TUI

Open the Quizzes view with `F4`.

Quizzes collect missing project or use-case knowledge through bounded questions.
The TUI shows quiz state, active question, recorded answers, and curated notes.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a quiz. |
| `s` | Start a quiz, or restart a completed/cancelled session. |
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

## Write and Edit Specs in the Web UI

Open Specs and choose **New**. Enter a title, choose **Project-wide** or an
existing named use case, and write the document in Body. Draft is on by default.
Leave it on to create a draft, or turn it off to create an approved spec directly.
Choose **Create**. Manual authoring needs no dump, quiz or model request.

Each card represents one stored spec and previews its opening body text. Several
specs can share one scope. Open the card to edit that exact document, including a
generated spec. The editor prefills its title, body and Draft state. Its use case
remains fixed. **Save** retains its identity, scope, creation time and provenance.

Errors keep the form available for correction or retry. If another client changes
the spec, Save rejects the stale revision. **Reload spec** explicitly discards the
pending form and loads the current document. Background catalog refresh does not
replace an open editor. An approved spec can be edited without first making a
separate draft.

Generated specs start in draft. Generation uses project context, the canonical
general dump and matching optional use-case discovery. It excludes unrelated
cases. An edit made while generation is preparing output prevents that output
from replacing the edited document.

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
| `alt+t` | Create legacy tracking for a historical plan without a revision-bound companion. |
| `r` | Refresh plans from disk. |

`g` does not overwrite an existing plan. Use `G` when the existing plan should
be replaced. Both actions use the same discovery readiness check as the Web UI.
Regeneration prepares and validates the replacement first, checks that sources
and the existing plan still match, then publishes the archive and replacement
together. Archived generated plans remain on disk and are excluded from normal
document loading. Provider errors, cancellation and stale inputs preserve the
current plan.

Each successful generation retains its own internal order map. Generating a new
revision preserves earlier maps and ticket history, even if the plan text is
unchanged. Explicit operation retries reuse the committed revision. The initial
map has planned identities and order; it allocates no tickets and starts no work.
Approving a prepared native workflow binds canonical task tickets to that revision.
Later approved changes use guarded amendment and retain completed work. The
[materialization reference](../../reference/plan-workflow-materialization.md)
describes the approved input contract.

Workers receive the current order map and use live ticket state for progress.
Ticket progress leaves the map unchanged. If a required map is missing or edited,
repair it with the [native tracker commands](../../reference/workflow-trackers.md)
before continuing. Replacing edited output requires its exact observed hash.

`alt+t` remains available for historical body-only plans and retains its existing
tracking behavior. For newly generated plans, it reports that order is already
retained and requires native approval before ticket allocation.

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
handoff metadata. In `agentic_review` loop mode, chr runs one default
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

## Prepare Work Through Telegram

Configure the dedicated receiver with a local operator and registered project,
then start it:

```sh
chr telegram setup --user-id 100001 --operator-id maintainer --project-id example-project
chr telegram doctor --check-connection
chr telegram serve
```

Setup reads the bot token privately. In Telegram, select the project and ask an
explanation or start a preparation-only conversation:

```text
/project example-project
/explain How are retries bounded?
/prepare Retain a retry ceiling
```

Inspect or change harness defaults for that selected project without restarting
the receiver:

```text
/harness
/harness worker=code generation=writing
```

Replace `code` and `writing` with locally prepared profile names. Read the exact
project and choices in the confirmation, then reply `confirm` or `cancel` (or use
`/answer <token> confirm`). Omitted roles stay unchanged; `generation=-` restores
legacy generation dispatch. Confirmations expire after ten minutes and reject
changed revisions or authority. New operations see saved defaults; running,
paused and resumed work keeps its original binding. Configuration and private
setup cannot be authorized by a model proposal. See
[Telegram harness defaults](../../reference/telegram-receiver.md#harness-defaults).

Send the requirements as several messages. Use `/sessions` to inspect or select
the returned conversation ID when several conversations are active. `/done`
generates its canonical dump. `/quiz` asks a real clarification; reply to its
prompt or use `/reply <conversation> <answer>`. Request `/spec` or `/plan` when
those artifacts are useful. `/refine spec <conversation> <instructions>` explicitly
replaces an observed specification.

Use `/prepare-work <conversation>` to prepare a native backlog ticket from its
dump or spec. The response identifies the actual canonical record and receipt.
An unchanged preparation returns that result, so a retry creates no extra ticket.
Preparation starts no work invocation.

Preview an exact native extent with `/work ticket <ticket-key>`. Use
`/work --conversation <conversation> ticket <ticket-key>` to keep an existing
conversation's project and namespace. Other extents are `slice`, `slice-set`,
`use-case` and `loose`; loose work omits its reference. Inspect the displayed
membership, target, execution limits and publication policy, then press its
start button, reply `start`, or send `/answer <action-token> start`.

Leave `chr telegram serve` running after confirmation. It continues the retained
extent without another message. Automatic policy advances through review and
integration; human policy waits on the exact displayed candidate. Routine
progress arrives at most once per thirty seconds. Status remains available
during work. Native pauses, capacity and other owners remain authoritative.

Each response identifies the invocation, project and namespace. Open that same
invocation in the TUI or Web, or run `chr work status --json <invocation>` in its
registered project with the matching `OPS_USER`. Host-validated tasks, open PRs,
confirmed receipts and remaining tickets are separate observations. A final
confirmed receipt establishes integration. If notification delivery is uncertain,
a labelled retry repeats the retained notice and preserves the original work.

Start with `/develop <topic>` to retain explicit development intent. Its first
prepared scope can use that instruction. Missing authority or changed requirements
produce an exact preview; use its button, reply `develop`, or send
`/answer <action-token> develop`. Development authority applies only to that
displayed scope. Ordinary-language requests use the same operations, but a model
proposal cannot grant authority. Restart preserves the project binding, artifacts,
questions and completed results.

See [Telegram receiver](../../reference/telegram-receiver.md) for configuration,
limits and the supported controls.

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

Before agent work, chr can compose a context package with:

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
