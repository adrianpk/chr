# TUI Reference

The TUI opens with `chr` or `chr tui`.

```sh
chr tui --dir /path/to/project
```

The TUI loads project state from disk and writes durable Markdown artifacts back
to the core.

See [CLI TUI action matrix](cli-tui-action-matrix.md) for current parity
between TUI actions and command-line operations.

Most TUI write actions have equivalent CLI commands. Use the CLI for scripted
operation and the TUI for interactive inspection, editing, and confirmation.

## Views

| Key | View | Purpose |
| --- | --- | --- |
| `F1` | Cases | Use-case list, use-case context, and related artifact counts. |
| `F2` | Dumps | Free-form Markdown context dumps. |
| `F3` | Quizzes | Bounded question sessions and curated quiz notes. |
| `F4` | Specs | Generated specifications by use case. |
| `F5` | Plans | Generated implementation plans and plan tracking creation. |
| `F6` | Board | Ticket board grouped by status. |
| `F7` | Dash | Agent activity and runtime evidence. |
| `F10` | Exit | Graceful terminal shutdown. |

## Global Keys

| Key | Behavior |
| --- | --- |
| `F1` to `F7` | Change view outside text entry or pending operations. |
| `alt+p` | Open Project metadata, context, artifact counts and Harness defaults. |
| `?` | Open contextual Help outside text entry; Esc returns to its originating pane. |
| `tab` | Move to the next pane. |
| `shift+tab` | Move to the previous pane. |
| `up` / `k` | Select the previous row or field in the focused selection pane. |
| `down` / `j` | Select the next row or field in the focused selection pane. |
| `alt+m` | Toggle maximized mode for the focused pane. |
| `:` | Open the command palette. |
| `/` | Start filtering when the current view supports filtering. |
| `esc` | Cancel the active mode or return focus to the primary pane. |
| `q`, `ctrl+c`, `F10` | Quit. |

Cases is the initial view. Selecting a case binds Dumps, Quizzes, Specs, Plans
and Board to it; tab changes retain the observed run and never submit work.
The header names the case. An unavailable case remains explicitly bound until
refreshed or replaced; **Clear case context** in the palette returns to all
scopes. The shell footer shows the active view's actions.

The footer tabs are clickable. Mouse dragging near pane dividers resizes the
layout.

## Guided Use Cases

From Cases (`F1`), `g` opens the guided catalog. The **Guided use cases** command
palette entry opens the same modal screen. The selected project and resolved
namespace stay bound to all requests. Catalog entries use recognizable names;
Enter inspects the selected retained run without starting or advancing it.

| Key | Guided behavior |
| --- | --- |
| `n` | Start a new retained development request under the displayed project policy. |
| `e` / `enter` | Edit the current required input; Enter in the editor saves its complete text. |
| `ctrl+j` | Insert a newline while editing. |
| `f` | Explicitly finish nonempty capture when offered. |
| `p` / `u` / `x` | Apply an available pause, resume or cancel through shared/native control owners. |
| `a` | Inspect required native approval through the existing preview/confirmation flow. |
| `i` | Edit a separately retained later idea after an input checkpoint exists. |
| `1` / `2` / `3` / `4` / `5` | Read the complete committed dump, quiz, specification, plan or tracker. |
| `v` | Read the current verified result export, including deferred ideas. |
| arrows / `pgup` / `pgdown` | Select catalog rows or scroll complete run/document content. |
| `r` | Read current canonical state without advancing the workflow. |
| `y` | Retry the exact original failed request after leaving the editor. |
| `c` | Reconcile the displayed retained operation when available. |
| `d` | Discard the local draft and its old binding; canonical input stays saved. |
| `b` | Return to the named run catalog. |
| `esc` | Leave the editor or guided screen. |
| `ctrl+c` | Quit the TUI and join its owned managed host. |

The editor retains its original run revision, question/input binding and operation
through asynchronous refresh. Rejected saves preserve the full draft. Read the
current state and explicitly open a new editor before submitting a revised
answer; a newer observation cannot silently rebind an old answer. Obsolete
responses from a previously closed screen cannot replace a reopened screen.

Startup attaches the shared bounded recovery/execution host. Inspection, rendering
and periodic refresh remain read-only. Leaving the modal does not cancel its run;
process exit stops owned jobs, preserving the canonical stage and original limits
for later recovery. A paused run stays paused after reopening.

Progress distinguishes native validation, confirmed delivery, remaining work,
review waits, blockers and pending native stop acknowledgement. Result completion
requires verified delivery and closing evidence for all admitted work. The export
contains curated delivery/validation/closing receipts and later ideas; raw native
state cannot substitute for those proofs. Committed documents carry their full
commit and repository-relative source, and remain stable through refresh. The
current result export is refreshed from canonical evidence rather than retained
as an immutable stage input.

## Command Palette

Open with `:`.

| Key | Behavior |
| --- | --- |
| text input | Filter by command id, title, or description. |
| `up` / `down` | Select a command. |
| `enter` | Execute the selected command. |
| `backspace` | Delete one filter character. |
| `esc` | Close the palette. |

Available command families:

- Navigation to the seven tabs and Project, Help, Tickets, Documents, Authoring,
  Agent and Runtime. Runtime includes workspace-index health and readiness.
- Clear case context without writing documents.
- Workspace refresh.
- Guided use-case capture, clarification, execution controls and results.
- Selected source path inspection.
- Default ticket creation.
- Dump creation, edit, and refresh.
- Quiz start, resume, answer, skip, complete, cancel, refresh, deep mode, and
  fast mode.
- Spec generation, Gherkin acceptance feature generation, and refresh, only in
  the Specs view.
- Plan generation, regeneration, tracking creation, and refresh, only in the
  Plans view.
- Ticket comment and selected ticket status transitions.
- Work start, ready-loop start, run-all-ready, and agent stop request.

`Create context-aware quiz`, `Toggle ticket checklist item`, and `Add ticket
checklist item` are capability checks in the current TUI. They report an
unavailable-operation error until their safe writers or provider-backed paths
are implemented.

Acceptance feature generation can return `needs_more_info`. In that case the
TUI reports the feature-round note path. Interactive collection of those
feature-round answers is not implemented in the current TUI.

## Project View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select project name, description, or context. |
| `enter` or `e` | Edit the selected field. |
| `enter` while editing | Apply the edit locally. |
| `esc` while editing | Cancel the edit. |
| `s` or `ctrl+s` | Save project metadata. |
| `r` | Refresh workspace state. |

Refresh is blocked while project edits are active or unsaved.

CLI equivalent:

```sh
chr metadata project set --name "Project" --description "Description" --context "Context"
```

## Cases View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` in pane 0 | Select a use case. |
| `enter` in pane 0 | Focus the use-case detail pane. |
| `up` / `down` or `k` / `j` in pane 1 | Select a use-case field. |
| `enter` or `e` in pane 1 | Edit the selected field. |
| `n` | Create a new use case. |
| `/` in pane 0 | Filter use cases. |
| `enter` while filtering | Apply the filter. |
| `esc` while filtering | Cancel filter editing. |
| `esc` in pane 0 | Clear the case filter. |
| `s` or `ctrl+s` | Save the selected use case. |
| `r` | Refresh workspace state. |

Refresh is blocked while use-case edits are active or unsaved.

CLI equivalent:

```sh
chr metadata case set billing --name "Billing" --description "Description" --context "Context"
```

## Dumps View

The Dumps view shows curated primary dumps. Raw capture siblings are retained on
disk and shown as metadata for the selected curated dump when available.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a dump. |
| `c` | Plan dump creation. Confirm with `enter`. |
| `enter` | Edit the selected dump in the TUI editor. |
| `alt+e` | Open the selected dump in the external editor. |
| `r` | Refresh dumps from disk. |
| `ctrl+s` while editing | Save the TUI editor content. |
| `esc` while editing | Cancel TUI editing. |

CLI equivalents:

```sh
chr dump list
chr dump new --title "Notes" --content "Context." --use-case billing
chr dump show ops/default/dump/00000-notes/artifact.md
chr dump save ops/default/dump/00000-notes/artifact.md --content "# Notes"
chr dump append ops/default/dump/00000-notes/artifact.md --content "More context."
chr dump open ops/default/dump/00000-notes/artifact.md
```

## Quizzes View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a quiz. |
| `s` | Start or restart the selected quiz. |
| `enter` | Enter an answer for the active question. |
| `r` | Refresh quizzes from disk. |

Answer entry:

| Key | Behavior |
| --- | --- |
| `enter` | Submit the answer. |
| `shift+enter` or `ctrl+j` | Insert a newline. |
| `alt+s` | Skip the active question. |
| `alt+d` | Mark the quiz done. |
| `esc` | Cancel answer entry. |

CLI equivalents:

```sh
chr quiz list
chr quiz start --use-case billing --mode deep
chr quiz resume --use-case billing
chr quiz answer --use-case billing --answer "Answer text."
chr quiz skip --use-case billing
chr quiz complete --use-case billing
chr quiz cancel --use-case billing
chr quiz mode fast --use-case billing
```

## Specs View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Generate the selected spec. |
| `r` | Refresh specs from disk. |

Spec generation uses project context, use-case context, dumps, quizzes, and
related workspace documents.

CLI equivalents:

```sh
chr spec list
chr spec generate --use-case billing
```

## Plans View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Generate the selected plan only if it is missing. |
| `G` | Regenerate the selected plan after confirmation. |
| `alt+t` | Create plan tracking and tickets after confirmation. |
| `r` | Refresh plans from disk. |

Plan generation is guarded against overlapping requests. `g` does not overwrite
an existing plan. `G` archives the existing generated plan before writing a new
one. `alt+t` is available only when the selected plan exists and contains
content.

CLI equivalents:

```sh
chr plan list
chr plan generate --use-case billing
chr plan regenerate --use-case billing
chr plan tracking create --use-case billing
```

## Board View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a ticket. |
| `left` / `right` or `h` / `l` | Move between normal Board columns. |
| `alt+a` | Toggle the virtual attention view. |
| `alt+m` | Maximize the focused pane. In the ticket pane, maximized Board shows all normal columns; in attention mode it shows only attention tickets. |
| `/` | Filter tickets. |
| `enter` while filtering | Apply the filter and select the first match. |
| `esc` while filtering | Cancel filter editing. |

Normal Board columns are `backlog`, `ready_for_dev`, `in_progress`,
`ready_for_review`, `ready_for_merge`, and `done`. Legacy statuses still parse:
`open` maps to Backlog, `ready` maps to Ready For Dev, `reviewing` maps to
Ready For Review, `solved` maps to Ready For Merge, and `merged` or `closed`
map to Done. `blocked` and statuses outside normal Board columns appear in the
attention view.

The selected ticket detail includes synopsis, source metadata, slice metadata,
checklist counts, comment counts, bounded Markdown body context, checklist
preview, evidence strategy summary, recent comment headings, attention reasons,
and the source path when available.

Evidence summary renders the overall checklist state plus strategy, Gherkin,
tests, validation, and evidence-summary item states. Malformed evidence blocks
show the parse error. Tickets without the anchored evidence checklist show
`missing` in Board detail; CLI list and board output render missing evidence as
`-`.

The compact Board header marks the active column explicitly and also shows a
column strip with the active normal column or attention view in brackets.

Ticket and work actions are available from the command palette and use
confirmation prompts.

CLI equivalents:

```sh
chr ticket board
chr ticket list
chr ticket show CHR-1
chr ticket status CHR-1 ready
chr ticket status CHR-1 in_progress
chr ticket status CHR-1 reviewing
chr ticket start CHR-1
chr ticket loop --status ready_for_dev --max-iterations 1
chr work ticket CHR-1
chr work run-all
```

## Dash View

The work-readiness panel shows the effective workspace mode, current branch,
expected integration branch and selected ticket worktree state. In worktree
mode, user checkout changes are informational; dirty or unsafe execution
worktrees block work. The expected branch identifies the integration policy
and does not require switching a user's unrelated checkout in worktree mode.
Use `chr work doctor --ticket <ticket-key>` for the detailed workspace paths,
and inspect the execution report or worker context to find where the agent
writes. See [Work readiness](work-readiness.md).

Dash shows agent-work operations:

- work readiness status and repair guidance;
- ready queue;
- current loop or execution state;
- recent human-readable agent event messages;
- ticket, execution report, runtime log, and validation log paths when event
  refs are available;
- recent execution and loop outcomes.

When Dash is selected, the TUI periodically reloads ticket and observability
state from persisted execution, loop, daemon, runtime, and work-readiness
records. Refresh does not start work or mutate ticket files.

Dash work actions use confirmation prompts:

- `enter` starts the selected ticket;
- `p` cycles the ready-queue priority filter;
- `o` cycles the ready-queue order;
- `a` or `R` runs all visible `ready_for_dev` tickets with the selected
  priority and order;
- `x` applies safe work-readiness repairs when Dash reports one is available;
- `s` requests agent stop.

When work readiness is blocked, Dash blocks ticket execution and run-all actions
before starting agent work. If a safe repair is available, apply `x`, confirm
the repair, and retry the work action. If Dash reports manual action required,
use the diagnostic text or `chr work doctor --dir <project>` to inspect the
same readiness report outside the TUI.

Event stream rendering, runtime logs, report lookup, and CLI readback are
documented in [Agent work observability](agent-work-observability.md).

## Confirmation Modes

Confirmation prompts are used for actions that write files, change ticket
state, or start work.

| Key | Behavior |
| --- | --- |
| `enter` | Confirm. |
| `esc` | Cancel. |

Busy spec, plan, and quiz states block overlapping actions until the running
operation finishes.
