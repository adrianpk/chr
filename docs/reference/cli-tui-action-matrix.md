# CLI TUI Action Matrix

This matrix tracks parity between the TUI and CLI operation surfaces.

Status values:

- `matched`: TUI and CLI expose equivalent operator behavior.
- `missing_cli`: the TUI exposes an operational action with no equivalent CLI.
- `missing_tui`: the CLI exposes an operational action with no equivalent TUI.
- `service_divergence`: both surfaces exist, but they do not clearly share the
  same application service path.
- `intentionally_interactive`: the action is terminal navigation, layout, or
  editing behavior that should remain TUI-only.
- `not_implemented`: the TUI advertises a capability check, but the operation is
  not implemented yet.

## Direct TUI Keybindings

| Domain | Action | TUI entrypoint | CLI command | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Navigation | Open Project view | `F1` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Cases view | `F2` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Dumps view | `F3` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Quizzes view | `F4` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Specs view | `F5` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Plans view | `F6` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Board view | `F7` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Dash view | `F8` | none | `internal/chctui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Help view | `F9` | `chc --help`, command-specific `--help` | `internal/chctui.Model.selectView`, `urfave/cli` help | `matched` | Different surfaces, same help intent. |
| Layout | Move pane focus | `tab`, `shift+tab`, `esc` | none | `internal/chctui.Model.selectNextPane`, `selectPreviousPane`, `selectPrimaryPane` | `intentionally_interactive` | TUI-only layout state. |
| Layout | Maximize focused pane | `alt+m` | none | `internal/chctui.Model.toggleMaximizedPane` | `intentionally_interactive` | TUI-only layout state. |
| Layout | Resize panes | mouse drag | none | `internal/chctui.Model.applyLayoutDrag` | `intentionally_interactive` | TUI-only layout state. |
| Shell | Quit TUI | `q`, `ctrl+c`, `F10` | process exit | Bubble Tea quit | `intentionally_interactive` | Terminal lifecycle. |
| Palette | Open command palette | `:` | none | `internal/chctui.CommandPaletteState` | `intentionally_interactive` | TUI-only command launcher. |
| Project | Edit project metadata field | `F1`, field selection, `enter` or `e` | none | `internal/chctui.ProjectEditState` | `missing_cli` | CLI lacks direct project metadata set command. |
| Project | Save project metadata | `F1`, `s` or `ctrl+s` | none | `service.SaveProjectMetadata` through `tuiProjectMetadataExecutor` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Project | Refresh workspace state | `F1`, `r` | `chc index rebuild`, `chc index status` partially | `service.RefreshWorkspace` through `tuiWorkspaceRefreshExecutor` | `missing_cli` | CLI index commands cover ticket index, not full workspace graph refresh. |
| Cases | Select/filter use cases | `F2`, arrows, `/` | none | `internal/chctui` filter state | `intentionally_interactive` | List navigation/filtering is TUI-only. |
| Cases | Create use case entry | `F2`, `n` | none | `internal/chctui.Model.newCase`, saved by `service.SaveCaseMetadata` | `missing_cli` | CLI lacks case metadata create command. |
| Cases | Edit use-case metadata field | `F2`, `enter` or `e` | none | `internal/chctui.ProjectEditState` | `missing_cli` | CLI lacks case metadata set command. |
| Cases | Save selected use case | `F2`, `s` or `ctrl+s` | none | `service.SaveCaseMetadata` through `tuiCaseMetadataExecutor` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Dumps | Create dump | `F3`, `c`, confirm | none | `service.CreateDump` through `tuiDumpActionExecutor` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Dumps | Edit dump in TUI editor | `F3`, `enter`, `ctrl+s` | none | `service.ReadDump`, `service.SaveDump` | `missing_cli` | CLI lacks read/save dump operations. |
| Dumps | Open dump in external editor | `F3`, `alt+e` | none | `service.OpenDump` | `missing_cli` | CLI lacks external edit command. |
| Dumps | Refresh dump list | `F3`, `r` | none | `service.ListDumps` | `missing_cli` | CLI lacks dump list/refresh command. |
| Quizzes | Start selected quiz | `F4`, `s` | none | `service.StartQuiz` through `tuiQuizActionExecutor` | `missing_cli` | Runtime quiz lifecycle service exists; CLI lacks command. |
| Quizzes | Answer active question | `F4`, `enter`, answer editor, `enter` | none | `service.AnswerQuiz` | `missing_cli` | CLI lacks runtime quiz answer command. |
| Quizzes | Skip active question | answer editor `alt+s` | none | `service.SkipQuiz` | `missing_cli` | CLI lacks runtime quiz skip command. |
| Quizzes | Complete selected quiz | answer editor `alt+d` | none | `service.CompleteQuiz` | `missing_cli` | CLI lacks runtime quiz complete command. |
| Quizzes | Refresh quizzes | `F4`, `r` | none | `service.ListQuizzes` | `missing_cli` | CLI lacks runtime quiz list/status command. |
| Specs | Generate spec | `F5`, `g` | none | `service.GenerateSpec` through `tuiSpecActionExecutor` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Specs | Refresh specs | `F5`, `r` | none | `service.ListSpecs` | `missing_cli` | CLI lacks spec list/show command. |
| Plans | Generate missing plan | `F6`, `g` | none | `service.GeneratePlan` through `tuiPlanActionExecutor` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Plans | Regenerate plan | `F6`, `G`, confirm | none | `service.RegeneratePlan` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Plans | Create plan tracking and tickets | `F6`, `alt+t`, confirm | none | `service.MaterializePlanTracking` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Plans | Refresh plans | `F6`, `r` | none | `service.ListPlans` | `missing_cli` | CLI lacks plan list/show command. |
| Board | Select/filter tickets | `F7`, arrows, `/` | `chc ticket list` filters | `internal/chctui` filter state, `runTicketList` | `matched` | CLI exposes ticket listing/filtering, TUI provides interactive selection. |
