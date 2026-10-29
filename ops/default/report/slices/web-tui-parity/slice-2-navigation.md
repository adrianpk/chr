# Slice 2: Case-oriented shell and contextual help

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-2-tui-navigation`
PR: [#456](https://forge.adrianpk.com/adrianpk/chr/pulls/456)
Base: `1ac51bac2aa5d307cc7400e7c07462c7e0542eda` on dev, after verified integration of PR #455

## Purpose

Resolve G-001, G-002, G-003, G-033 and I-002 with case-oriented navigation,
contextual Help, useful footer actions and reachable existing capabilities.

## Delivered Behavior

- Cases opens first. F1-F7 and footer clicks select Cases, Dumps, Quizzes,
  Specs, Plans, Board and Dash by view identity; F10 remains Exit.
- Project appears in header context and opens with alt+p or the palette. `?`
  opens real Help for the active view; Esc restores its originating pane.
  Text editors, filters and pending operations keep their own input handling.
- Opening a case or observing its retained run binds document lists, operation
  scopes and Board to that case. Tab navigation retains the run and selected
  artifacts without saving documents or dispatching work. The header names the
  bound case; a removed case stays unavailable with refresh, replacement and
  explicit palette clearing as recovery paths.
- The actual shell footer shows local actions before common navigation hints.
  Useful auxiliary views remain in the palette without adding another tab bar.

## Implementation Notes

T2.1-T2.3 share `f2d7961bebc10139851cdaaeae4ac28b3fafa725`: shell selection,
case-scoped getters and contextual return state depend on one model transition.
Acceptance and capability inventory are recorded in
`5dbd0c2ba271465b3e47d66143f86658b802c3c8`; the task ledger also records
PR preparation and the pending controller integration boundary.

The first navigation regression reproduced the previous Project startup and
nine-tab mapping. Existing fixtures now select view IDs rather than old numeric
positions. The 32 catalog frames were updated for the header, footer and body
geometry; all nonempty non-chrome body lines were compared before replacement
and remained identical. Full-frame assertions remain enabled.

The new owned PTY fixture loads actual persisted cases and dumps through the
root command's production loader, chooses the second case, traverses Dumps,
Help, Project and Runtime, joins its child, and compares all canonical ops file
hashes. Its initial setup failure was the missing required ticket directory;
creating that empty directory fixes setup without changing the assertions or
time limits. The retained-run integration test supplies terminal geometry to
check the real header, traverses all bound tabs and verifies the unchanged run
revision. Passive run observations preserve unsent local case text and selection.

Useful-capability inventory:

| Existing capability | Retained destination |
| --- | --- |
| Project metadata, context and Harness defaults | Header alt+p, Project palette entry, Project h |
| Artifact counts and unresolved case references | Project and Cases artifact panels and case list |
| Workspace-index health and runtime checks | Go to Runtime in the palette |
| Additional documents, authoring sessions and agent observations | Go to Documents, Authoring and Agent in the palette |
| Seventeen-field ticket filters | Board or palette Tickets, `/` |
| External source editing | Existing alt+e actions and source command |
| Acceptance-feature generation | Specs f and its view-scoped palette command |
| Safe readiness repair | Dash x and existing confirmed palette command |
| Queue filters, ordering and Run all ready | Dash p, o and a; existing work palette commands |
| Running-work events and log locations | Dash details and palette Agent/Runtime |
| Later ideas, committed originals and result JSON | Existing guided i, 1-5 and v controls |

The navigation changes retain these owners and command handlers. Existing short
regressions cover their controls; the new PTY explicitly exercises auxiliary
reachability. This inventory does not claim a fresh execution of every provider
or external editor.

## Contracts Added or Changed

Transient navigation state carries case slug, observed run ID, auxiliary view
and Help return pane. Scoped projections preserve original artifact indexes;
missing cases return no document or ticket instead of falling back to another
scope. `GuidedView.Cases` carries a bounded application read of canonical case
metadata after a run observation; nil leaves the projection unchanged, while
active or unsaved edits prevent replacement. Domain operations and durable
schemas remain owned by the existing services.

## Files of Interest

- `internal/tui/shell_shortcuts.go`, `help.go`, `shell_chrome.go`,
  `command_palette.go`: shell actions, contextual Help and auxiliary navigation.
- `internal/tui/case_navigation.go`, `case_navigation_views.go`, `guided.go`:
  case/run binding, scoped projections and passive observation handling.
- `internal/app/tui_adapters_guided.go`: bounded canonical case observations.
- `internal/tui/shell_navigation_test.go`, `case_navigation_test.go`,
  `internal/app/tui_navigation_terminal_linux_test.go`, `tui_guided_test.go`:
  keyboard, mouse, draft, missing-case, terminal and retained-run regressions.
- `docs/how-to/user-guide/index.md`, `docs/reference/tui.md`,
  `docs/reference/cli-tui-action-matrix.md`, `CHANGELOG.md`: updated user controls.

## Validation

Implementation regressions passed locally:

- `go test ./internal/tui -count=1` — passed, including all 32 catalog frames.
- `go test ./internal/tui -run '^Test(Shell.*|CaseNavigation.*)$' -count=1` — passed.
- `go test ./internal/app -run '^Test(GuidedUseCaseTUIJourney|TUINavigationTerminal)$' -count=1 -timeout=2m` — passed with owned filesystem, SQLite, Git and PTY fixtures.
- `git diff --check` — passed.

After the PR-reference follow-up, Looper records the clean final candidate and
this ordered final batch in its canonical validation evidence:

- `make verify`
- `go test -race ./internal/tui -run '^Test(Shell.*|CaseNavigation.*|DocumentCatalogFrames|GuidedUseCaseTUI(StaleInput|Replay|RetainedStep)|Model(Ownership|AsyncRefresh|WorkflowBinding))$' -count=1`
- `go test -race ./internal/app -run '^Test(TUINavigationTerminal|GuidedUseCaseTUIJourney|GuidedUseCaseTUITerminal|TUIBusyTerminal|TUIDashTarget)$' -count=1 -timeout=3m`
- `git diff --check`

The PR Validation section records successful final commands. Canonical dev
integration remains pending until the controller verifies its merge.

## Risks and Follow-ups

None recorded.
