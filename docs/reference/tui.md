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

## Operation Results

Manual operation results distinguish `ok`, `validation_failed`, `stale` and
`blocked`. Failures open the Action panel with the summary, validation findings
and blocking reason. Successful writes show their summary in the status line.
Press `alt+o` to inspect the latest result, including known changed files; use
arrows or Home/End to scroll. Esc returns to the complete retained input.

The panel offers `r` only when a reload is available and `y` only for an eligible
retry of the original operation. Reload asks for confirmation before replacing
unsaved text. Enter confirms; Esc keeps it. Recovery for an older interaction
cannot submit or replace the current editor's input.

Project and Case field editors, Dump editing and Quiz answers ask before
discarding entered text. Quiz Skip and Done also ask when an answer has not
been submitted. A submitted answer stays visible until acknowledgement; a
failed save or late response cannot clear a newer draft. Case and Project
metadata reloads read the captured source, and Dump and Quiz reloads retain
their original scope until confirmation succeeds.

## Guided Use Cases

Cases (`F1`) exposes `w` **Wizard** (`g` alias) and `o` **Open workflow**. Wizard
collects one Name/Create action through the named guided owner. Open workflow
uses the selected case, prefers its already bound run and otherwise selects the
latest retained run with a stable identity tie-break. A case must be selected
when multiple runs are present. The palette also offers Open workflow. No run ID
or namespace is requested from the user.

The existing shell shows the bound case and stage progress. Stages use Cases,
Dumps, Quizzes, Specs, Plans and Board; F1–F7 remain the navigation menu. Earlier
and future stages are read-only. `c` returns to the current stage without
resubmitting its input. Passive refresh can update the stage while preserving
the tab being read. Manual keyboard, palette and source-edit commands cannot
write through a wizard-bound tab; clear case context explicitly for manual work.

| Key | Wizard behavior |
| --- | --- |
| `e` / `enter` at the current stage | Open its eligible input in a shared form/editor. |
| `ctrl+s` in Dump | Continue to quiz using the complete current body in one retained finish operation. |
| `enter` in an editor | Insert a newline; use Ctrl+S or its submit button for progression. |
| `alt+e` while editing Dump | Edit the unsent body through the existing owned draft editor. |
| `p` / `u` / `x` at the current stage | Apply an available pause, resume or cancel through shared owners. |
| `a` | Inspect an available native approval through the existing exact confirmation. |
| `i` | Enter a later idea separately when an input checkpoint exists. |
| `1`–`5` / `v` | Read committed originals or the retained result export. |
| arrows / `pgup` / `pgdown` | Scroll the complete stage/document. |
| `r` | Read canonical state without advancing it. |
| `y` outside an editor | Retry the exact original failed request. |
| `c` | Return to the current tab, or reconcile an offered retained operation when already there. |
| `esc` in an editor | Leave after confirming unsent changes; the case/run binding remains. |
| `ctrl+c` | Quit and join the owned managed host. |

The Dump stage uses the same body editor and draft/external-editor path as
manual Dump editing. Canonical wizard submissions are limited to 65,536 bytes. Its
progression action replaces manual Save; editing does not dispatch a separate
append. Repeating unchanged Continue after a lost acknowledgement reuses the
same operation, revision and content. Invalid or stale input retains the draft.
Generic input forms also freeze run, question and revision through refresh.

Inspection reads committed bytes with their actual reference/commit binding,
independent of later working-copy replacements. Controls remain guarded by the
guided/native owners. Pauses and process exit retain consumed work and original
limits. Validated tasks, review waits, confirmed delivery and closing evidence
remain distinct; a worker exit or an open PR cannot establish completion.

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
- Complete ticket creation and editing with canonical metadata choices.
- Dump creation, edit, and refresh.
- Quiz start, resume, answer, skip, complete, cancel, refresh, deep mode, and
  fast mode.
- Spec generation, Gherkin acceptance feature generation, and refresh, only in
  the Specs view.
- Plan generation, regeneration, tracking creation, and refresh, only in the
  Plans view.
- Free-text ticket comments, retained discussion, review-reference attachment and selected ticket status transitions.
- Work start, ready-loop start, run-all-ready, and agent stop request.

`Context-aware quiz (unavailable)` explains its separate contextual authoring CLI
alternative before any confirmation or executor call. Use `chr author quiz new`
for that authoring document, or F3 for supported project/case runtime questions.
`Toggle ticket checklist item` and `Add ticket checklist item` remain capability
checks for their safe writers.

Acceptance feature generation can return `needs_more_info`. In that case the
TUI reports the feature-round note path. Interactive collection of those
feature-round answers is not implemented in the current TUI.

## Project View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select project name, description, or context. |
| `enter` or `e` | Edit the selected field. |
| `enter` while editing | Apply the edit locally. |
| `esc` while editing | Confirm discarding changed text. |
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
| `n` | Open the Name, Description and Context form. |
| `/` in pane 0 | Filter use cases. |
| `enter` while filtering | Apply the filter. |
| `esc` while filtering | Cancel filter editing. |
| `esc` in pane 0 | Clear the case filter. |
| `s` or `ctrl+s` | Save the selected use case. |
| `r` | Refresh workspace state. |

Press `n`, enter **Name**, **Description** and **Context**, then submit once
with `ctrl+s` or Enter on **Create**. Tab and Shift+Tab move between fields;
Enter adds a line in Description and Context. Nothing is created before submission.
Validation failures retain all entered fields. Esc cancels an empty form or asks
before discarding entered text. Existing metadata editing and artifact references
remain available after creation.

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
| `c` | Open the Scope, Title and Body form. |
| `enter` | Edit the selected dump in the TUI editor. |
| `a` | Enter free text to append to the selected dump. |
| `alt+e` | Open the selected dump externally; while editing, round-trip the unsaved body. |
| `r` | Refresh dumps from disk. |
| `ctrl+s` while editing | Save the TUI editor content. |
| `esc` while editing | Confirm discarding unsaved text. |

The creation form defaults to the selected case. Use left/right in **Scope**
to choose **Project-wide** or a named case, then enter your title and body.
Tab and Shift+Tab choose a field; `ctrl+s` or Enter on **Create** submits once.
The Append form accepts your text for its captured document and revision.
Esc confirms discarding entered text, and failed submissions retain the form.

The body editor shows the existing title and scope as retained metadata.
`alt+e` inside it opens a private copy containing the current unsaved text.
Returning from the external editor retains the result in the TUI; `ctrl+s`
saves it to the document. Editor errors retain the previous draft. Outside the
body editor, `alt+e` opens the selected stored document and refreshes after exit.

A stale Save preserves your text. Open the Action panel with `alt+o`, choose
Reload with `r` and confirm only when ready to replace it with stored content.
Cancelling reload keeps the draft; refresh never silently overwrites it.

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
| `s` | Start or restart with the displayed Deep/Fast choice. |
| `m` | Choose Deep/Fast before Start; change an open quiz through its service. |
| `u` | Resume retained work. |
| `x` | Confirm cancellation, retaining saved answers. |
| `enter` | Enter an answer for the active question. |
| `r` | Refresh quizzes from disk. |

Answer entry:

| Key | Behavior |
| --- | --- |
| `enter` | Submit the answer. |
| `shift+enter` or `ctrl+j` | Insert a newline. |
| `alt+s` | Skip the active question. |
| `alt+d` | Mark the quiz done. |
| `alt+x` | Confirm cancellation, retaining saved answers. |
| `alt+m` | Change Deep/Fast mode; confirm discarding unsent text first. |
| `PgUp` / `PgDn` | Page complete saved question/answer history. |
| `esc` | Confirm discarding an unsent answer. |


The selected case scopes available quiz targets. `[answer pending]` marks a target
with an active question. The displayed mode is submitted by Start; changing it
before Start writes no session. Saved answers appear newest first beside the
current question and input. Skipped questions are excluded from answer history.
Paging history and resizing keep the current prompt and unsent answer intact.

Each answer retains its observed session, question and revision. Failed or stale
submissions preserve the draft and its original question. The Action panel
(`alt+o`) offers confirmed Reload or an unchanged retry when eligible. A read-only
refresh cannot pair an old draft with a newer prompt. If generation fails after
an answer was saved, its history remains and `u` resumes continuation. Cancel
retains saved answers; Skip, Done and mode changes ask before discarding unsent
text. Esc discards only the current input after confirmation.

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

The catalog lists individual canonical manual, generated and project-wide
specifications, including several documents in the same case. It preserves the
selected path across refresh and shows its body and status. Case context filters
the list; clear it in the palette to see all scopes.

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a specification. |
| `c` | Create with Scope, Title, Body and Draft/Approved state. |
| `e` | Edit the selected document's title, body and state; retain its case. |
| `g` | Generate a spec for the selected document's case or current context. |
| `f` | Generate an acceptance feature through the existing scoped operation. |
| `r` | Refresh canonical specifications from disk. |
| `alt+e` | Open the selected canonical source externally. |

Forms use Tab/Shift+Tab for fields and Ctrl+S or Enter on Create/Save to submit.
Enter adds a line in Body. Scope is selectable only during creation. Errors keep
all inputs; stale Save cannot overwrite newer content. The Action panel offers
confirmed Reload for the captured source. Esc asks before discarding input.
Unchanged retries keep the original request, including after a saved result's
catalog refresh failed.

Manual creation needs no generation material. Spec generation retains project
context, use-case context, dumps, quizzes and related workspace documents.

Related CLI commands (listing retains the latest-in-scope CLI inventory):

```sh
chr spec list
chr spec generate --use-case billing
chr acceptance feature generate --use-case billing
```

## Plans View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Prepare the case plan and native tickets together; resume retained preparation. |
| `G` | Regenerate the selected plan after confirmation. |
| `r` | Refresh plans from disk. |

Generate requires an existing named case with a saved Spec. It uses
`GeneratePlanDelivery` with the captured operation ID and target revision.
Existing plans are retained; partial ticket preparation resumes without duplicate
plans or tickets. The result's `y` retries the exact original request. A stale
result's `r` reloads the captured case. Neither preparation nor result navigation
starts execution.

After successful Spec/Plan generation, `o` opens the saved document in its tab.
A combined plan/ticket result also offers `t` to open the case-filtered Board and
a populated column. These actions appear only for retained outputs and remain
bound to the current result. `alt+o` reopens that result.

`G` archives and replaces the generated plan through the existing publication
owner, preserving captured target revision and source checks. Cancelled or stale
replacement does not write. The scoped catalog prefers the canonical target to
retained native copies. Concurrent plan actions are blocked.

Independent tracking remains a CLI capability. It is not a second terminal step:

```sh
chr plan list
chr plan generate --use-case billing
chr plan regenerate --use-case billing
chr plan tracking create --use-case billing
```

## Board View

| Key | Behavior |
| --- | --- |
| `c` | Create a ticket with title, description, case and metadata. |
| `e` in the ticket pane | Edit the selected ticket. In the detail pane, `e` retains external source editing. |
| `a` | Enter a comment and read retained history. |
| `v` | Attach an entered review reference. |
| `r` | Refresh ticket observations. |
| `alt+e` | Open the selected canonical source in the external editor. |
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

At width 132 and height 27 or greater, the default Board shows six normal
columns in two rows and a separate Attention area, beside ticket detail.
Narrower layouts retain the compact column view and maximize. The header and
column strip mark the active column or Attention. Resizing and focus preserve
the selected ticket key and canonical column.

Create ticket, Edit ticket, Add ticket comment and Attach ticket review reference
open the same forms from the palette or Board/Tickets keys. Tab/Shift+Tab moves
through every field, including fields beyond the current viewport; left/right
chooses canonical options. Ctrl+S submits, and Esc confirms discarding entered
input. Edit shows the phase, slice and owner-provided read-only reasons.

Comments are separate from metadata/body edits. Shift+Tab from Comment focuses
retained history; arrow keys scroll it without modifying saved discussion.
Review references collect type, provider, HTTP(S) URL, optional ID, status and
branch. Existing references and checklist content remain intact.

A failed or stale submission retains all fields. From its result, r requests a
confirmed reload; Esc keeps the complete draft. A changed case binding requires
explicit reopening. Active claims and verified delivery values stay protected.
If a write succeeded but refresh failed, close the retained form and refresh
before editing again; the form cannot resubmit that committed write.

Other ticket status and work actions retain their existing confirmations and
canonical evidence requirements.

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

Dash reads the same application observation as Web: eligible and total queue
counts, readiness reasons, exclusions, current work, recent activity and retained
results. A ticket's local status cannot override policy or workspace readiness.
Native counters distinguish host-validated tasks from confirmed delivery;
a successful harness run remains labelled as a harness result.

Queue pages contain at most 128 rows. Priority and order apply before pagination.
Board keeps its separate catalog and selection; paging Dash does not truncate
Board. Both selections follow their key through refresh and reorder. A removed
key falls back to the first visible row, and an empty page has no target.

When Dash is selected, periodic observation reads retained state without starting
or resuming work. It preserves document drafts and ignores responses for older
queue queries. Current and Recent panes scroll with their normal pane controls;
`alt+m` maximizes the focused pane. Current contexts and events are bounded at
128 each, and results at 16. Running-work details retain available ticket,
execution report, runtime log and validation log paths.

Dash actions:

- `enter` opens the selected eligible ticket's work confirmation;
- `v` loads **Preview Work** for the selected native ticket without executing it;
- `b` opens **Preview Work Scope** for ticket, slice, set, case or loose work;
- `w` opens **Retained workflows** with independent list/detail focus;
- `[` and `]` load the previous and next queue pages;
- `p` cycles the priority filter, and `o` cycles the order, returning to page one;
- `a` or `R` runs all ready work with the selected priority and order; native
  authority previews the independent-ticket extent and its exclusions first;
- `x` confirms safe work-readiness repair when one is available;
- `s` confirms an agent-stop request.

Native preview shows the exact scope, snapshot, target, policy and frozen
execution limits. Enter separately confirms at most one checkpoint; Esc cancels
without admission. Changed scope, ticket revision, snapshot or policy requires
a fresh preview. Preview Work and retained Workflow details require native
ticket authority; existing legacy work actions remain available under their
configured authority.

The scope form selects a named case or a ticket key. Loose work supports
priority, severity and project-area filters and canonical ordering; other scopes
retain hierarchy order. Preview shows actual later-ticket policy and frozen
limits and binds the complete selection even when only a bounded member window
is visible.

In retained workflows, up/down selects the invocation, Enter reads current
status, `r` refreshes the focused list/detail, `[` / `]` pages through invocations,
and `n` / `m` pages through revision-bound members. Tab focuses detail and arrows
scroll evidence. Esc or `w` returns to the queue. Controls are:

| Key | Behavior |
| --- | --- |
| `c`, `p`, `x`, `u` | Continue one checkpoint, pause, cancel or resume. |
| `a`, `R` | Approve the exact displayed candidate or reconcile interrupted ownership. |
| `g` | Preview an allowance extension with kind, amount, actor and reason. |
| `d` | Answer a retained question with its displayed option number and resolve/defer; enter actor, reason and optional original commit/fix metadata. |
| `t`, `C` | Open the linked ticket editor or linked case. |
| `b` | Open bounded scope preview. |

Unavailable controls display reasons. Mutations require separate confirmation
and recheck canonical invocation, entry, budget, candidate, approval and decision
fences. A stale form submission retains its input and original guards. Reopen
from current status before authorizing changed state. Allowance consumption is
preserved; decision answers do not approve or publish work. Terminal states stay
read-only. No worker exit or open PR establishes confirmed delivery.

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
