# TUI Reference

The TUI opens with `chc` or `chc tui`.

```sh
chc tui --dir /path/to/project
```

The TUI loads project state from disk and writes durable Markdown artifacts back
to the workspace.

See [CLI TUI action matrix](cli-tui-action-matrix.md) for current parity
between TUI actions and command-line operations.

## Views

| Key | View | Purpose |
| --- | --- | --- |
| `F1` | Project | Project metadata, project context, location, and artifact counts. |
| `F2` | Cases | Use-case list, use-case context, and related artifact counts. |
| `F3` | Dumps | Free-form Markdown context dumps. |
| `F4` | Quizzes | Bounded question sessions and curated quiz notes. |
| `F5` | Specs | Generated specifications by use case. |
| `F6` | Plans | Generated implementation plans and plan tracking creation. |
| `F7` | Board | Ticket board grouped by status. |
| `F8` | Dash | Agent activity and runtime evidence. |
| `F9` | Help | Help view and command palette target. |

## Global Keys

| Key | Behavior |
| --- | --- |
| `F1` to `F9` | Change view. |
| `tab` | Move to the next pane. |
| `shift+tab` | Move to the previous pane. |
| `up` / `k` | Select the previous row or field in the focused selection pane. |
| `down` / `j` | Select the next row or field in the focused selection pane. |
| `alt+m` | Toggle maximized mode for the focused pane. |
| `:` | Open the command palette. |
| `/` | Start filtering when the current view supports filtering. |
| `esc` | Cancel the active mode or return focus to the primary pane. |
| `q`, `ctrl+c`, `F10` | Quit. |

The footer tabs are clickable. Mouse dragging near pane dividers resizes the
layout.

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

- Navigation to every TUI view.
- Workspace refresh.
- Selected source path inspection.
- Default ticket creation.
- Dump creation, edit, and refresh.
- Quiz start, resume, answer, skip, complete, cancel, refresh, deep mode, and
  fast mode.
- Spec generation and refresh, only in the Specs view.
- Plan generation, regeneration, tracking creation, and refresh, only in the
  Plans view.
- Ticket comment and selected ticket status transitions.
- Work start, ready-loop start, and agent stop request.

`Create context-aware quiz`, `Toggle ticket checklist item`, and `Add ticket
checklist item` are capability checks in the current TUI. They report an
unavailable-operation error until their safe writers or provider-backed paths
are implemented.

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

## Dumps View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a dump. |
| `c` | Plan dump creation. Confirm with `enter`. |
| `enter` | Edit the selected dump in the TUI editor. |
| `alt+e` | Open the selected dump in the external editor. |
| `r` | Refresh dumps from disk. |
| `ctrl+s` while editing | Save the TUI editor content. |
| `esc` while editing | Cancel TUI editing. |

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

## Specs View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a use case. |
| `g` | Generate the selected spec. |
| `r` | Refresh specs from disk. |

Spec generation uses project context, use-case context, dumps, quizzes, and
related workspace documents.

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

## Board View

| Key | Behavior |
| --- | --- |
| `up` / `down` or `k` / `j` | Select a ticket. |
| `/` | Filter tickets. |
| `enter` while filtering | Apply the filter and select the first match. |
| `esc` while filtering | Cancel filter editing. |

Ticket and work actions are available from the command palette and use
confirmation prompts.

## Confirmation Modes

Confirmation prompts are used for actions that write files, change ticket
state, or start work.

| Key | Behavior |
| --- | --- |
| `enter` | Confirm. |
| `esc` | Cancel. |

Busy spec, plan, and quiz states block overlapping actions until the running
operation finishes.
