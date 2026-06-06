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
| Layout | Maximize focused pane | `alt+m` | none | `internal/chctui.Model.toggleMaximizedPane` | `intentionally_interactive` | TUI-only layout state. Board ticket-pane maximization renders all normal Board columns. |
| Layout | Resize panes | mouse drag | none | `internal/chctui.Model.applyLayoutDrag` | `intentionally_interactive` | TUI-only layout state. |
| Shell | Quit TUI | `q`, `ctrl+c`, `F10` | process exit | Bubble Tea quit | `intentionally_interactive` | Terminal lifecycle. |
| Palette | Open command palette | `:` | none | `internal/chctui.CommandPaletteState` | `intentionally_interactive` | TUI-only command launcher. |
| Project | Edit project metadata field | `F1`, field selection, `enter` or `e` | `chc metadata project set --name ... --description ... --context ...` | `internal/chctui.ProjectEditState`, `service.SaveProjectMetadata` | `matched` | CLI edits by flag instead of interactive field focus. |
| Project | Save project metadata | `F1`, `s` or `ctrl+s` | `chc metadata project set ...` | `service.SaveProjectMetadata` through `tuiProjectMetadataExecutor` and CLI command | `matched` | Same metadata save service. |
| Project | Refresh workspace state | `F1`, `r` | `chc index rebuild`, `chc index status` partially | `service.RefreshWorkspace` through `tuiWorkspaceRefreshExecutor` | `missing_cli` | CLI index commands cover ticket index, not full workspace graph refresh. |
| Cases | Select/filter use cases | `F2`, arrows, `/` | none | `internal/chctui` filter state | `intentionally_interactive` | List navigation/filtering is TUI-only. |
| Cases | Create use case entry | `F2`, `n` | `chc metadata case set <slug> --name ... --description ... --context ...` | `internal/chctui.Model.newCase`, saved by `service.SaveCaseMetadata` | `matched` | CLI creates by slug and metadata flags. |
| Cases | Edit use-case metadata field | `F2`, `enter` or `e` | `chc metadata case set <slug> ...` | `internal/chctui.ProjectEditState`, `service.SaveCaseMetadata` | `matched` | CLI edits by flag instead of interactive field focus. |
| Cases | Save selected use case | `F2`, `s` or `ctrl+s` | `chc metadata case set <slug> ...` | `service.SaveCaseMetadata` through `tuiCaseMetadataExecutor` and CLI command | `matched` | Same metadata save service. |
| Dumps | Create dump | `F3`, `c`, confirm | `chc dump new --title ... --content ... [--use-case ...]` | `service.CreateDump` through `tuiDumpActionExecutor` and CLI command | `matched` | Same dump creation service. |
| Dumps | Edit dump in TUI editor | `F3`, `enter`, `ctrl+s` | `chc dump show <path>`, `chc dump save <path> --content ...`, `chc dump append <path> --content ...` | `service.ReadDump`, `service.SaveDump`, `service.AppendDump` | `matched` | CLI is non-interactive and path-based. |
| Dumps | Open dump in external editor | `F3`, `alt+e` | `chc dump open <path>` | `service.OpenDump` | `matched` | CLI prints the resolved editor command. |
| Dumps | Refresh dump list | `F3`, `r` | `chc dump list` | `service.ListDumps` | `matched` | Same dump list service. |
| Quizzes | Start selected quiz | `F4`, `s` | `chc quiz start [--use-case ...] [--mode deep|fast]` | `service.StartQuiz` through `tuiQuizActionExecutor` and CLI command | `matched` | Same runtime quiz service. |
| Quizzes | Answer active question | `F4`, `enter`, answer editor, `enter` | `chc quiz answer --answer ... [--use-case ...]` | `service.AnswerQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Skip active question | answer editor `alt+s` | `chc quiz skip [--use-case ...]` | `service.SkipQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Complete selected quiz | answer editor `alt+d` | `chc quiz complete [--use-case ...]` | `service.CompleteQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Refresh quizzes | `F4`, `r` | `chc quiz list` | `service.ListQuizzes` | `matched` | Same runtime quiz list service. |
| Specs | Generate spec | `F5`, `g` | `chc spec generate [--use-case ...]` | `service.GenerateSpec` through `tuiSpecActionExecutor` and CLI command | `matched` | Same spec generation service. |
| Specs | Generate Gherkin acceptance feature | `F5`, `f` | `chc acceptance feature generate [--use-case ...]` | `service.GenerateAcceptanceFeature` through `tuiSpecActionExecutor` and CLI command | `partial` | Same generation service. If more information is needed, both surfaces report the feature-round note path; interactive answer collection is not implemented yet. |
| Specs | Refresh specs | `F5`, `r` | `chc spec list` | `service.ListSpecs` | `matched` | Same spec list service. |
| Plans | Generate missing plan | `F6`, `g` | `chc plan generate [--use-case ...]` | `service.GeneratePlan` through `tuiPlanActionExecutor` and CLI command | `matched` | Same plan generation service. |
| Plans | Regenerate plan | `F6`, `G`, confirm | `chc plan regenerate [--use-case ...]` | `service.RegeneratePlan` | `matched` | Same plan regeneration service. |
| Plans | Create plan tracking and tickets | `F6`, `alt+t`, confirm | `chc plan tracking create [--use-case ...]` | `service.MaterializePlanTracking` | `matched` | Same plan tracking materialization service. |
| Plans | Refresh plans | `F6`, `r` | `chc plan list` | `service.ListPlans` | `matched` | Same plan list service. |
| Board | Select/filter tickets | `F7`, arrows, `/` | `chc ticket list` filters | `internal/chctui` filter state, `runTicketList` | `matched` | CLI exposes ticket listing/filtering, TUI provides interactive selection. |
| Board | Move between normal Board columns | `F7`, `left` / `right` or `h` / `l` | `chc ticket board` | `internal/chctui.Model.selectBoardColumn`, `service.Board` | `matched` | CLI prints all Board columns; TUI navigates one compact column at a time. |
| Board | Toggle attention view | `F7`, `alt+a` | `chc ticket board` attention summary | `internal/chctui.Model.toggleBoardAttention`, `chc.BoardAttentionReasons` | `matched` | TUI shows attention tickets separately; CLI prints attention count through the same Board service. |
| Board | Maximize Board ticket pane | `F7`, ticket pane, `alt+m` | `chc ticket board` | `internal/chctui.renderMaximizedBoardWorkspace`, `service.Board` | `matched` | TUI renders a multi-column Board; CLI prints equivalent column data. |
| Board | Selected ticket detail | `F7`, selected ticket detail pane | `chc ticket show <key>` | `internal/chctui.Model.renderSelectedTicketDetail`, `runTicketShow` | `matched` | TUI detail includes indexed metadata and bounded Markdown context; CLI shows the selected ticket file. |

## Command Palette Commands

| Domain | Command palette id | TUI action | CLI command | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Workspace | `refresh` | Reload workspace files into runtime graph. | `chc index rebuild`, `chc index status` partially | `service.RefreshWorkspace` | `missing_cli` | CLI index commands target ticket index, not the full TUI workspace graph refresh. |
| Help | `help` | Navigate to Help view. | `chc --help` | `internal/chctui.Model.selectView`, CLI help | `matched` | Equivalent intent, different presentation. |
| Selection | `open-source` | Show selected source path. | none | `internal/chctui.Model.selectedSourcePath` | `missing_cli` | CLI has path outputs in list/show commands, but no selected-source concept. |
| Ticket | `create-ticket` | Create default task ticket. | `chc ticket new <title> --synopsis ...` | `service.CreateTicket`; TUI also wraps defaults with `service.CreateDefaultTicket` | `matched` | CLI supports full metadata; TUI uses default metadata through the same ticket creation service. |
| Dump | `create-dump` | Create selected project/use-case dump. | `chc dump new ...` | `service.CreateDump` | `matched` | Same service path. |
| Dump | `edit-dump` | Open selected dump in TUI editor. | `chc dump show`, `chc dump save`, `chc dump append` | `service.ReadDump`, `service.SaveDump`, `service.AppendDump` | `matched` | CLI is non-interactive and path-based. |
| Dump | `refresh-dumps` | Reload dump views. | `chc dump list` | `service.ListDumps` | `matched` | Same service path. |
| Quiz | `quiz-start` | Start selected runtime quiz. | `chc quiz start ...` | `service.StartQuiz` | `matched` | Different from authoring quiz generation. |
| Quiz | `quiz-resume` | Resume selected runtime quiz. | `chc quiz resume ...` | `service.ResumeQuiz` | `matched` | Same service path. |
| Quiz | `quiz-answer` | Answer active runtime quiz question. | `chc quiz answer ...` | `service.AnswerQuiz` | `matched` | Same service path. |
| Quiz | `quiz-skip` | Skip active runtime quiz question. | `chc quiz skip ...` | `service.SkipQuiz` | `matched` | Same service path. |
| Quiz | `quiz-complete` | Complete selected runtime quiz. | `chc quiz complete ...` | `service.CompleteQuiz` | `matched` | Same service path. |
| Quiz | `quiz-cancel` | Cancel selected runtime quiz. | `chc quiz cancel ...` | `service.CancelQuiz` | `matched` | Same service path. |
| Quiz | `quiz-refresh` | Reload runtime quiz state. | `chc quiz list` | `service.ListQuizzes` | `matched` | Same service path. |
| Quiz | `quiz-mode-deep` | Switch selected runtime quiz to deep mode. | `chc quiz mode deep ...` | `service.SwitchQuizMode` | `matched` | Same service path. |
| Quiz | `quiz-mode-fast` | Switch selected runtime quiz to fast mode. | `chc quiz mode fast ...` | `service.SwitchQuizMode` | `matched` | Same service path. |
| Spec | `generate-spec` | Generate selected spec. | `chc spec generate ...` | `service.GenerateSpec` | `matched` | Command is gated to Specs view. |
| Acceptance | `generate-acceptance-feature` | Generate selected Gherkin acceptance feature. | `chc acceptance feature generate ...` | `service.GenerateAcceptanceFeature` | `partial` | Command is gated to Specs view. Missing scenario detail is persisted as feature-round notes; TUI answer collection is not implemented yet. |
| Spec | `refresh-specs` | Reload spec views. | `chc spec list` | `service.ListSpecs` | `matched` | Command is gated to Specs view. |
| Plan | `generate-plan` | Generate missing selected plan. | `chc plan generate ...` | `service.GeneratePlan` | `matched` | Command is gated to Plans view. |
| Plan | `regenerate-plan` | Regenerate selected plan. | `chc plan regenerate ...` | `service.RegeneratePlan` | `matched` | Command is gated to Plans view. |
| Plan | `materialize-plan-tracking` | Create plan tracking and tickets. | `chc plan tracking create ...` | `service.MaterializePlanTracking` | `matched` | Command is gated to Plans and only available for existing plans. |
| Plan | `refresh-plans` | Reload plan views. | `chc plan list` | `service.ListPlans` | `matched` | Command is gated to Plans view. |
| Authoring | `create-authoring-session` | Create TUI default authoring session. | `chc author session new ...` | `service.CreateAuthoringSession`; TUI wraps default session metadata with `service.CreateTUIAuthoringSession` | `matched` | CLI is configurable; TUI creates a default spec-oriented session through the same service. |
| Authoring | `create-context-quiz` | Validate context-aware quiz creation availability. | `chc author quiz new ...` | TUI returns unavailable; CLI `runAuthorContextualQuizNew` | `missing_tui` | CLI supports context-aware quiz generation; TUI action is not implemented. |
| Ticket | `ticket-add-comment` | Add default selected-ticket comment. | `chc ticket comment <key> <body>` | `service.AddTicketComment`; TUI wraps default body with `service.AddSelectedTicketComment` | `matched` | CLI supports arbitrary body; TUI supplies the standard body through the same service. |
| Ticket | `ticket-status-ready` | Move selected ticket to `ready`. | `chc ticket status <key> ready` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-status-in-progress` | Move selected ticket to `in_progress`. | `chc ticket status <key> in_progress` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-status-reviewing` | Move selected ticket to `reviewing`. | `chc ticket status <key> reviewing` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-checklist-toggle` | Validate checklist toggle availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Ticket | `ticket-checklist-add` | Validate checklist item creation availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Work | `work-start-ticket` | Start selected ticket execution. | `chc ticket start <key>` | shared `buildTicketStartPreflightForTicket` and `runTicketStartExecution` | `matched` | CLI owns preflight/output; TUI service wraps selected-ticket execution. |
| Work | `work-start-ready-loop` | Start one ready-ticket loop. | `chc ticket loop --status ready_for_dev --max-iterations 1` | shared `runTicketLoopExecution` | `matched` | CLI supports configurable query/run options; TUI wraps one ready-ticket iteration. |
| Work | `work-run-all-ready` | Run all visible ready tickets with Dash filters and order. | `chc work run-all [filters]` | `service.StartAgentWorkLoop` -> `runTicketLoopExecution` | `matched` | CLI and TUI use the configurable agent-work facade. |
| Work | `work-request-stop` | Request daemon stop. | `chc daemon stop` | shared `requestDaemonStopState` | `matched` | CLI owns output and dry-run; TUI reloads work state after the shared stop request. |

Command availability gates:

- Spec commands are available only in the Specs view.
- Plan commands are available only in the Plans view.
- Dash refreshes agent-work observability from persisted execution and loop
  records while selected.
- `materialize-plan-tracking` is additionally available only when the selected
  plan exists and contains content.

## CLI Command Inventory

| Domain | CLI command | TUI equivalent | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Shell | `chc` | Opens TUI | `runDefaultTUIAction`, `chctui.Run` | `matched` | Root command launches the operational console. |
| Shell | `chc tui --dir <path>` | Opens TUI | `runTUICommandAction`, `chctui.Run` | `matched` | Explicit TUI command. |
| Workspace | `chc init` | none | `chc.Workspace.Init` | `missing_tui` | Workspace bootstrap is CLI-only. |
| Runtime | `chc runtime init` | none | `chc.Workspace.InitRuntime` through runtime command | `missing_tui` | Runtime asset installation is CLI-only. |
| Runtime | `chc runtime validate` | Dash displays runtime evidence only | `runRuntimeValidate` | `missing_tui` | TUI does not run runtime validation directly. |
| Project Metadata | `chc metadata project set` | Project metadata save | `service.SaveProjectMetadata` | `matched` | CLI edits by flag; TUI edits interactively. |
| Use-Case Metadata | `chc metadata case set` | Cases metadata save | `service.SaveCaseMetadata` | `matched` | CLI creates or updates by slug. |
| Project Registry | `chc project add` | none | `runProjectAdd`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chc project list` | none | `runProjectList`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chc project current` | current project topbar | `runProjectCurrent`, user registry | `matched` | TUI shows resolved project, CLI prints active project. |
| Project Registry | `chc project select` | none | `runProjectSelect`, user registry | `missing_tui` | Registry selection is CLI-only. |
| Tickets | `chc ticket new` | Palette `create-ticket` partially | `service.CreateTicket`; TUI also wraps defaults with `service.CreateDefaultTicket` | `matched` | CLI supports metadata; TUI creates a default task ticket through the same service. |
| Tickets | `chc ticket list` | Board view | `runTicketList`, ticket index | `matched` | TUI board and CLI list inspect indexed tickets. |
| Tickets | `chc ticket board` | Board view and maximized Board | `service.Board` | `matched` | CLI prints Board totals and columns; TUI renders compact and maximized Board views. |
| Tickets | `chc ticket show` | Board selected ticket detail | `runTicketShow`, ticket index | `matched` | Equivalent inspection surface. |
| Tickets | `chc ticket comment` | Palette `ticket-add-comment` partially | `service.AddTicketComment`; TUI wraps default body with `service.AddSelectedTicketComment` | `matched` | CLI body is user-supplied; TUI body is fixed. |
| Tickets | `chc ticket status` | Palette ticket status commands | `service.MoveTicketStatus`; TUI wraps selected ticket with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Dumps | `chc dump list` | Dumps refresh | `service.ListDumps` | `matched` | Same service path. |
| Dumps | `chc dump new` | Dumps create | `service.CreateDump` | `matched` | CLI creates project-wide or use-case scoped dumps. |
| Dumps | `chc dump show` | Dumps editor load | `service.ReadDump` | `matched` | CLI prints content for a path. |
| Dumps | `chc dump save` | Dumps editor save | `service.SaveDump` | `matched` | CLI replaces content for a path. |
| Dumps | `chc dump append` | Dump append action | `service.AppendDump` | `matched` | CLI appends content for a path. |
| Dumps | `chc dump open` | Dumps external editor | `service.OpenDump` | `matched` | CLI prints the resolved editor command. |
| Runtime Quizzes | `chc quiz list` | Quizzes refresh | `service.ListQuizzes` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz start` | Quizzes start | `service.StartQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz resume` | Quizzes resume | `service.ResumeQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz answer` | Quizzes answer | `service.AnswerQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz skip` | Quizzes skip | `service.SkipQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz complete` | Quizzes complete | `service.CompleteQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz cancel` | Quizzes cancel | `service.CancelQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chc quiz mode` | Quizzes mode switch | `service.SwitchQuizMode` | `matched` | Same service path. |
| Specs | `chc spec list` | Specs refresh | `service.ListSpecs` | `matched` | Same service path. |
| Specs | `chc spec generate` | Specs generate | `service.GenerateSpec` | `matched` | Same service path. |
| Plans | `chc plan list` | Plans refresh | `service.ListPlans` | `matched` | Same service path. |
| Plans | `chc plan generate` | Plans generate | `service.GeneratePlan` | `matched` | Same service path. |
| Plans | `chc plan regenerate` | Plans regenerate | `service.RegeneratePlan` | `matched` | Same service path. |
| Plans | `chc plan tracking create` | Plan tracking materialization | `service.MaterializePlanTracking` | `matched` | Same service path. |
| Work | `chc ticket start` | Palette `work-start-ticket` | shared `buildTicketStartPreflightForTicket` and `runTicketStartExecution` | `matched` | Shared core execution path. |
| Work | `chc ticket loop` | Palette `work-start-ready-loop` partially | shared `runTicketLoopExecution` | `matched` | CLI is configurable; TUI runs one ready-ticket iteration. |
| Work | `chc ticket apply-result` | none | `runTicketApplyResult` | `missing_tui` | Result application is CLI-only. |
| Index | `chc index rebuild` | Workspace refresh partially | `runIndexRebuild`, ticket reconciler | `missing_tui` | CLI rebuilds ticket index; TUI refresh rebuilds workspace graph/read model. |
| Index | `chc index status` | Project/Dash diagnostics partially | `runIndexStatus` | `missing_tui` | CLI-specific index diagnostics. |
| Authoring | `chc author session new` | Palette `create-authoring-session` partially | `service.CreateAuthoringSession`; TUI wraps default session metadata with `service.CreateTUIAuthoringSession` | `matched` | CLI is configurable; TUI creates a default session through the same service. |
| Authoring | `chc author session list` | legacy Authoring view only | `runAuthorSessionList` | `missing_tui` | Authoring is not part of default TUI tabs. |
| Authoring | `chc author session show` | legacy Authoring view only | `runAuthorSessionShow` | `missing_tui` | Authoring is not part of default TUI tabs. |
| Authoring | `chc author quiz new` | Palette `create-context-quiz` not implemented | `runAuthorContextualQuizNew` | `missing_tui` | CLI supports context-aware authoring quiz generation. |
| Authoring | `chc author promote quiz-to-draft` | none | `runAuthorPromoteQuizToDraft`, `chc.PlanQuizToDraftPromotion` | `missing_tui` | Active CLI compatibility command; no TUI equivalent. |
| Assisted Authoring | `chc author assisted start` | none | `runAuthorAssistedStart` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chc author assisted resume` | none | `runAuthorAssistedResume` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chc author assisted answer` | none | `runAuthorAssistedAnswer` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chc author assisted status` | none | `runAuthorAssistedStatus` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chc author assisted finalize` | none | `runAuthorAssistedFinalize` | `missing_tui` | CLI-only assisted workflow. |
| Daemon | `chc daemon status` | Dash displays observability partially | `runDaemonStatus` | `missing_tui` | TUI displays state but does not expose full daemon status command. |
| Daemon | `chc daemon start` | Palette `work-start-ready-loop` partially | `runDaemonStart` | `missing_tui` | Daemon foreground loop is CLI-only. |
| Daemon | `chc daemon stop` | Palette `work-request-stop` | shared `requestDaemonStopState` | `matched` | CLI owns output and dry-run; TUI reloads work state after the shared stop request. |
| Reports | `chc report list` | Dash displays recent evidence partially | `runReportList` | `missing_tui` | Report browsing is CLI-only. |
| Reports | `chc report show` | Dash displays recent evidence partially | `runReportShow` | `missing_tui` | Report browsing is CLI-only. |
| Maintenance | `chc maintenance bundle` | none | `runMaintenanceBundle` | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chc maintenance db status` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chc maintenance execution artifacts` | Dash evidence partially | maintenance command path | `missing_tui` | Detailed artifact drift reporting is CLI-only. |
| Maintenance | `chc maintenance execution cleanup-preview` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chc maintenance registry status` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chc maintenance registry disable` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chc maintenance registry set-path` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Notify | `chc notify status` | none | `runNotifyStatus` | `missing_tui` | Notification inspection is CLI-only. |
| Notify | `chc notify telegram test` | none | `runNotifyTelegramTest` | `missing_tui` | Notification setup test is CLI-only. |
| Notify | `chc notify telegram poll` | none | `runNotifyTelegramPoll` | `missing_tui` | Remote command intake is CLI-only. |
| Proving | `chc proving validate` | none | `runProvingValidate` | `missing_tui` | Proving-ground workflows are CLI-only. |
| Proving | `chc proving run` | none | `runProvingRun` | `missing_tui` | Proving-ground workflows are CLI-only. |
| Proving | `chc proving ralph` | none | `runProvingRalph` | `missing_tui` | Proving-ground workflows are CLI-only. |

## Service Path Summary

| Service area | TUI path | CLI path | Status | Notes |
| --- | --- | --- | --- | --- |
| Workspace graph refresh | `tuiWorkspaceRefreshExecutor.RefreshWorkspace` -> `service.RefreshWorkspace` | no equivalent | `missing_cli` | CLI index commands do not refresh the complete workspace graph/read model. |
| Project metadata | `tuiProjectMetadataExecutor.SaveProjectMetadata` -> `service.SaveProjectMetadata` | `chc metadata project set` -> `service.SaveProjectMetadata` | `matched` | CLI edits by flag; TUI edits interactively. |
| Use-case metadata | `tuiCaseMetadataExecutor.SaveCaseMetadata` -> `service.SaveCaseMetadata` | `chc metadata case set` -> `service.SaveCaseMetadata` | `matched` | CLI creates or updates by slug. |
| Dump operations | `tuiDumpActionExecutor` -> `service.ListDumps`, `CreateDump`, `AppendDump`, `ReadDump`, `SaveDump`, `OpenDump` | `chc dump *` -> same service methods | `matched` | CLI is path-based and non-interactive. |
| Runtime quiz operations | `tuiQuizActionExecutor` -> `service.ListQuizzes`, `StartQuiz`, `ResumeQuiz`, `AnswerQuiz`, `SkipQuiz`, `CompleteQuiz`, `CancelQuiz`, `SwitchQuizMode` | `chc quiz *` -> same service methods | `matched` | Different from `chc author quiz new`; this is the runtime quiz loop. |
| Spec operations | `tuiSpecActionExecutor` -> `service.ListSpecs`, `GenerateSpec` | `chc spec list`, `chc spec generate` -> same service methods | `matched` | Same service path. |
| Plan operations | `tuiPlanActionExecutor` -> `service.ListPlans`, `GeneratePlan`, `RegeneratePlan`, `MaterializePlanTracking` | `chc plan *` -> same service methods | `matched` | Same service path. |
| Ticket creation | `tuiArtifactActionExecutor` -> `service.CreateDefaultTicket` -> `service.CreateTicket` | `runTicketNew` -> `service.CreateTicket` | `matched` | CLI behavior is richer; both use the same ticket creation service. |
| Board query | TUI Board views -> indexed ticket views | `runTicketBoard` -> `service.Board` | `matched` | CLI and TUI read the same Board columns and attention classification. |
| Ticket comment | `tuiTicketActionExecutor` -> `service.AddSelectedTicketComment` -> `service.AddTicketComment` | `runTicketComment` -> `service.AddTicketComment` | `matched` | CLI body is user-supplied; TUI supplies a default body. |
| Ticket status | `tuiTicketActionExecutor` -> `service.MoveSelectedTicketStatus` -> `service.MoveTicketStatus` | `runTicketStatus` -> `service.MoveTicketStatus` | `matched` | Same status service. |
| Work start ticket | `tuiWorkActionExecutor` -> `service.StartSelectedTicket` -> shared execution helpers | `runTicketStart` -> shared execution helpers | `matched` | TUI wraps selection and message reload; CLI owns dry-run and output. |
| Work ready loop | `tuiWorkActionExecutor` -> `service.StartReadyTicketLoop` -> `runTicketLoopExecution` | `runTicketLoop` -> `runTicketLoopExecution` | `matched` | TUI uses fixed ready/max-one selection; CLI is configurable. |
| Work run all ready | `tuiWorkActionExecutor` -> `service.StartAgentWorkLoop` -> `runTicketLoopExecution` | `chc work run-all` -> `service.StartAgentWorkLoop` | `matched` | Both paths use the agent-work facade and shared filter/order selection contract. |
| Work request stop | `tuiWorkActionExecutor` -> `service.RequestDaemonStop` -> `requestDaemonStopState` | `runDaemonStop` -> `requestDaemonStopState` | `matched` | CLI owns output and dry-run; TUI reloads work state after the shared stop request. |
| Authoring session creation | `tuiArtifactActionExecutor` -> `service.CreateTUIAuthoringSession` -> `service.CreateAuthoringSession` | `runAuthorSessionNew` -> `service.CreateAuthoringSession` | `matched` | CLI is configurable; TUI creates a default spec-oriented session. |
| Context-aware authoring quiz | TUI returns unavailable | `runAuthorContextualQuizNew` | `missing_tui` | CLI path exists and should inform future TUI implementation. |
| Assisted authoring | no default TUI path | `runAuthorAssisted*` | `missing_tui` | CLI-only workflow. |
| Project registry | no TUI path | `runProject*` | `missing_tui` | User-level registry management is CLI-only. |
| Maintenance/notify/proving | no TUI path | command-specific paths | `missing_tui` | Operational/admin command families are CLI-only. |

## Test Evidence

| Surface | Evidence | Coverage |
| --- | --- | --- |
| CLI command registration | `TestRootCommandRegistersParitySubcommands` | Verifies metadata, dump, quiz, spec, plan, and plan tracking command paths are registered. |
| CLI metadata and document commands | `TestMetadataProjectSetSavesProjectMetadata`, `TestMetadataCaseSetCreatesUseCaseMetadata`, `TestDumpNewAndShowUseServiceCommands`, `TestSpecAndPlanListCommandsUseServiceViews` | Verifies representative CLI commands write or read through project services. |
| TUI dump adapter | `TestTUIDumpAdapterUsesDumpServices` | Verifies TUI dump create/edit/save dispatch reaches dump services and preserves selected paths. |
| TUI runtime quiz adapter | `TestTUIQuizAdapterDispatchesRuntimeQuizActions` | Verifies TUI quiz start/answer/mode dispatch reaches runtime quiz services. |
| TUI plan adapter | `TestTUIPlanGenerateReturnsExistingPlanForGeneratedScope`, `TestTUIPlanGenerateReturnsRefusalMessageForExistingPlan` | Verifies TUI plan generation dispatch and existing-plan guard. |
| Board CLI and TUI | `TestTicketBoardPrintsBoardColumns`, `TestModelRendersMaximizedBoardColumns`, `TestModelRendersMaximizedBoardAttentionSeparately` | Verifies CLI Board output, maximized TUI Board columns, and separate attention rendering. |
| Service generation paths | `TestServiceGenerateSpecUsesUseCaseContext`, `TestServiceMaterializePlanTrackingCreatesTrackingAndTickets` | Verifies spec generation and plan tracking services used by both surfaces. |

## Gap Summary

High-priority `missing_cli` gaps:

- Workspace graph refresh remains only partially covered by `chc index rebuild`
  and `chc index status`.

Current documentation status:

- Metadata, dump, runtime quiz, spec, plan, plan tracking, Board, ticket, and
  work operations have documented CLI/TUI parity where both surfaces exist.
- Remaining gaps are explicit below and are either intentionally interactive,
  CLI-only administration, TUI-only navigation, or not-yet-implemented checklist
  capabilities.

High-priority `service_divergence` gaps:

- None currently identified in the audited surfaces.

Current `not_implemented` actions:

- TUI checklist toggle.
- TUI checklist item creation.

Important `missing_tui` command families:

- Workspace initialization and runtime setup.
- Project registry management.
- CLI authoring workflows.
- Reports, maintenance, notifications, and proving-ground workflows.
