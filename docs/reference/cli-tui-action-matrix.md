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
| Navigation | Open Project view | `alt+p` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Cases view | `F1` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Dumps view | `F2` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Quizzes view | `F3` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Specs view | `F4` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Plans view | `F5` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Board view | `F6` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Dash view | `F7` | none | `internal/tui.Model.selectView` | `intentionally_interactive` | TUI view navigation. |
| Navigation | Open Help view | `?` | `chr --help`, command-specific `--help` | `internal/tui.Model.selectView`, `urfave/cli` help | `matched` | Different surfaces, same help intent. |
| Layout | Move pane focus | `tab`, `shift+tab`, `esc` | none | `internal/tui.Model.selectNextPane`, `selectPreviousPane`, `selectPrimaryPane` | `intentionally_interactive` | TUI-only layout state. |
| Layout | Maximize focused pane | `alt+m` | none | `internal/tui.Model.toggleMaximizedPane` | `intentionally_interactive` | TUI-only layout state. Board ticket-pane maximization renders all normal Board columns. |
| Layout | Resize panes | mouse drag | none | `internal/tui.Model.applyLayoutDrag` | `intentionally_interactive` | TUI-only layout state. |
| Shell | Quit TUI | `q`, `ctrl+c`, `F10` | process exit | Bubble Tea quit | `intentionally_interactive` | Terminal lifecycle. |
| Shell | Inspect latest operation result | `alt+o`, arrows, Home/End | none | `internal/tui.Model.renderOperationFeedback` | `intentionally_interactive` | Structured outcomes, captured recovery and confirmed draft replacement. |
| Palette | Open command palette | `:` | none | `internal/tui.CommandPaletteState` | `intentionally_interactive` | TUI-only command launcher. |
| Project | Edit project metadata field | `alt+p`, field selection, `enter` or `e` | `chr metadata project set --name ... --description ... --context ...` | `internal/tui.ProjectEditState`, `service.SaveProjectMetadata` | `matched` | CLI edits by flag instead of interactive field focus. |
| Project | Save project metadata | `alt+p`, `s` or `ctrl+s` | `chr metadata project set ...` | `service.SaveProjectMetadata` through `tuiProjectMetadataExecutor` and CLI command | `matched` | Same metadata save service. |
| Project | Refresh workspace state | `alt+p`, `r` | `chr index rebuild`, `chr index status` partially | `service.RefreshWorkspace` through `tuiWorkspaceRefreshExecutor` | `missing_cli` | CLI index commands cover ticket index, not full workspace graph refresh. |
| Guided use cases | Start/resume, answer, control and inspect results | `F1`, `g`, or Guided use cases palette entry | none | `service.StartGuidedUseCase`, `ChangeGuidedUseCase`, `InspectGuidedUseCase`, `ReadGuidedUseCaseOriginal` | `missing_cli` | TUI, Web and Telegram share one retained run and native policy; standalone CLI authoring keeps its own semantics. |
| Cases | Select/filter use cases | `F1`, arrows, `/` | none | `internal/tui` filter state | `intentionally_interactive` | List navigation/filtering is TUI-only. |
| Cases | Create use case entry | `F1`, `n` | `chr metadata case set <slug> --name ... --description ... --context ...` | `internal/tui.Model.newCase`, saved by `service.SaveCaseMetadata` | `matched` | CLI creates by slug and metadata flags. |
| Cases | Edit use-case metadata field | `F1`, `enter` or `e` | `chr metadata case set <slug> ...` | `internal/tui.ProjectEditState`, `service.SaveCaseMetadata` | `matched` | CLI edits by flag instead of interactive field focus. |
| Cases | Save selected use case | `F1`, `s` or `ctrl+s` | `chr metadata case set <slug> ...` | `service.SaveCaseMetadata` through `tuiCaseMetadataExecutor` and CLI command | `matched` | Same metadata save service. |
| Dumps | Create dump | `F2`, `c`, confirm | `chr dump new --title ... --content ... [--use-case ...]` | `service.CreateDump` through `tuiDumpActionExecutor` and CLI command | `matched` | Same dump creation service. |
| Dumps | Edit dump in TUI editor | `F2`, `enter`, `ctrl+s` | `chr dump show <path>`, `chr dump save <path> --content ...`, `chr dump append <path> --content ...` | `service.ReadDump`, `service.SaveDump`, `service.AppendDump` | `matched` | CLI is non-interactive and path-based. |
| Dumps | Open dump in external editor | `F2`, `alt+e` | `chr dump open <path>` | `service.OpenDump` | `matched` | CLI prints the resolved editor command. |
| Dumps | Refresh dump list | `F2`, `r` | `chr dump list` | `service.ListDumps` | `matched` | Same dump list service. |
| Quizzes | Start selected quiz | `F3`, `s` | `chr quiz start [--use-case ...] [--mode deep|fast]` | `service.StartQuiz` through `tuiQuizActionExecutor` and CLI command | `matched` | Same runtime quiz service. |
| Quizzes | Answer active question | `F3`, `enter`, answer editor, `enter` | `chr quiz answer --answer ... [--use-case ...]` | `service.AnswerQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Skip active question | answer editor `alt+s` | `chr quiz skip [--use-case ...]` | `service.SkipQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Complete selected quiz | answer editor `alt+d` | `chr quiz complete [--use-case ...]` | `service.CompleteQuiz` | `matched` | Same runtime quiz service. |
| Quizzes | Refresh quizzes | `F3`, `r` | `chr quiz list` | `service.ListQuizzes` | `matched` | Same runtime quiz list service. |
| Specs | Generate spec | `F4`, `g` | `chr spec generate [--use-case ...]` | `service.GenerateSpec` through `tuiSpecActionExecutor` and CLI command | `matched` | Same spec generation service. |
| Specs | Generate Gherkin acceptance feature | `F4`, `f` | `chr acceptance feature generate [--use-case ...]` | `service.GenerateAcceptanceFeature` through `tuiSpecActionExecutor` and CLI command | `partial` | Same generation service. If more information is needed, both surfaces report the feature-round note path; interactive answer collection is not implemented yet. |
| Specs | Refresh specs | `F4`, `r` | `chr spec list` | `service.ListSpecs` | `matched` | Same spec list service. |
| Plans | Generate missing plan | `F5`, `g` | `chr plan generate [--use-case ...]` | `service.GeneratePlan` through `tuiPlanActionExecutor` and CLI command | `matched` | Same plan generation service. |
| Plans | Regenerate plan | `F5`, `G`, confirm | `chr plan regenerate [--use-case ...]` | `service.RegeneratePlan` | `matched` | Same plan regeneration service. |
| Plans | Create plan tracking and tickets | `F5`, `alt+t`, confirm | `chr plan tracking create [--use-case ...]` | `service.MaterializePlanTracking` | `matched` | Same plan tracking materialization service. |
| Plans | Refresh plans | `F5`, `r` | `chr plan list` | `service.ListPlans` | `matched` | Same plan list service. |
| Board | Select/filter tickets | `F6`, arrows, `/` | `chr ticket list` filters | `internal/tui` filter state, `runTicketList` | `matched` | CLI exposes ticket listing/filtering, TUI provides interactive selection. |
| Board | Move between normal Board columns | `F6`, `left` / `right` or `h` / `l` | `chr ticket board` | `internal/tui.Model.selectBoardColumn`, `service.Board` | `matched` | CLI prints all Board columns; TUI navigates one compact column at a time. |
| Board | Toggle attention view | `F6`, `alt+a` | `chr ticket board` attention summary | `internal/tui.Model.toggleBoardAttention`, `core.BoardAttentionReasons` | `matched` | TUI shows attention tickets separately; CLI prints attention count through the same Board service. |
| Board | Maximize Board ticket pane | `F6`, ticket pane, `alt+m` | `chr ticket board` | `internal/tui.renderMaximizedBoardWorkspace`, `service.Board` | `matched` | TUI renders a multi-column Board; CLI prints equivalent column data. |
| Board | Selected ticket detail | `F6`, selected ticket detail pane | `chr ticket show <key>` | `internal/tui.Model.renderSelectedTicketDetail`, `runTicketShow` | `matched` | TUI detail includes indexed metadata and bounded Markdown context; CLI shows the selected ticket file. |

## Command Palette Commands

| Domain | Command palette id | TUI action | CLI command | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Workspace | `refresh` | Reload workspace files into runtime graph. | `chr index rebuild`, `chr index status` partially | `service.RefreshWorkspace` | `missing_cli` | CLI index commands target ticket index, not the full TUI workspace graph refresh. |
| Help | `help` | Navigate to Help view. | `chr --help` | `internal/tui.Model.selectView`, CLI help | `matched` | Equivalent intent, different presentation. |
| Selection | `open-source` | Show selected source path. | none | `internal/tui.Model.selectedSourcePath` | `missing_cli` | CLI has path outputs in list/show commands, but no selected-source concept. |
| Ticket | `create-ticket` | Create default task ticket. | `chr ticket new <title> --synopsis ...` | `service.CreateTicket`; TUI also wraps defaults with `service.CreateDefaultTicket` | `matched` | CLI supports full metadata; TUI uses default metadata through the same ticket creation service. |
| Dump | `create-dump` | Create selected project/use-case dump. | `chr dump new ...` | `service.CreateDump` | `matched` | Same service path. |
| Dump | `edit-dump` | Open selected dump in TUI editor. | `chr dump show`, `chr dump save`, `chr dump append` | `service.ReadDump`, `service.SaveDump`, `service.AppendDump` | `matched` | CLI is non-interactive and path-based. |
| Dump | `refresh-dumps` | Reload dump views. | `chr dump list` | `service.ListDumps` | `matched` | Same service path. |
| Quiz | `quiz-start` | Start selected runtime quiz. | `chr quiz start ...` | `service.StartQuiz` | `matched` | Different from authoring quiz generation. |
| Quiz | `quiz-resume` | Resume selected runtime quiz. | `chr quiz resume ...` | `service.ResumeQuiz` | `matched` | Same service path. |
| Quiz | `quiz-answer` | Answer active runtime quiz question. | `chr quiz answer ...` | `service.AnswerQuiz` | `matched` | Same service path. |
| Quiz | `quiz-skip` | Skip active runtime quiz question. | `chr quiz skip ...` | `service.SkipQuiz` | `matched` | Same service path. |
| Quiz | `quiz-complete` | Complete selected runtime quiz. | `chr quiz complete ...` | `service.CompleteQuiz` | `matched` | Same service path. |
| Quiz | `quiz-cancel` | Cancel selected runtime quiz. | `chr quiz cancel ...` | `service.CancelQuiz` | `matched` | Same service path. |
| Quiz | `quiz-refresh` | Reload runtime quiz state. | `chr quiz list` | `service.ListQuizzes` | `matched` | Same service path. |
| Quiz | `quiz-mode-deep` | Switch selected runtime quiz to deep mode. | `chr quiz mode deep ...` | `service.SwitchQuizMode` | `matched` | Same service path. |
| Quiz | `quiz-mode-fast` | Switch selected runtime quiz to fast mode. | `chr quiz mode fast ...` | `service.SwitchQuizMode` | `matched` | Same service path. |
| Spec | `generate-spec` | Generate selected spec. | `chr spec generate ...` | `service.GenerateSpec` | `matched` | Command is gated to Specs view. |
| Acceptance | `generate-acceptance-feature` | Generate selected Gherkin acceptance feature. | `chr acceptance feature generate ...` | `service.GenerateAcceptanceFeature` | `partial` | Command is gated to Specs view. Missing scenario detail is persisted as feature-round notes; TUI answer collection is not implemented yet. |
| Spec | `refresh-specs` | Reload spec views. | `chr spec list` | `service.ListSpecs` | `matched` | Command is gated to Specs view. |
| Plan | `generate-plan` | Generate missing selected plan. | `chr plan generate ...` | `service.GeneratePlan` | `matched` | Command is gated to Plans view. |
| Plan | `regenerate-plan` | Regenerate selected plan. | `chr plan regenerate ...` | `service.RegeneratePlan` | `matched` | Command is gated to Plans view. |
| Plan | `materialize-plan-tracking` | Create plan tracking and tickets. | `chr plan tracking create ...` | `service.MaterializePlanTracking` | `matched` | Command is gated to Plans and only available for existing plans. |
| Plan | `refresh-plans` | Reload plan views. | `chr plan list` | `service.ListPlans` | `matched` | Command is gated to Plans view. |
| Authoring | `create-authoring-session` | Create TUI default authoring session. | `chr author session new ...` | `service.CreateAuthoringSession`; TUI wraps default session metadata with `service.CreateTUIAuthoringSession` | `matched` | CLI is configurable; TUI creates a default spec-oriented session through the same service. |
| Authoring | `create-context-quiz` | Validate context-aware quiz creation availability. | `chr author quiz new ...` | TUI returns unavailable; CLI `runAuthorContextualQuizNew` | `missing_tui` | CLI supports context-aware quiz generation; TUI action is not implemented. |
| Ticket | `ticket-add-comment` | Add default selected-ticket comment. | `chr ticket comment <key> <body>` | `service.AddTicketComment`; TUI wraps default body with `service.AddSelectedTicketComment` | `matched` | CLI supports arbitrary body; TUI supplies the standard body through the same service. |
| Ticket | `ticket-status-ready` | Move selected ticket to `ready`. | `chr ticket status <key> ready` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-status-in-progress` | Move selected ticket to `in_progress`. | `chr ticket status <key> in_progress` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-status-reviewing` | Move selected ticket to `reviewing`. | `chr ticket status <key> reviewing` | `service.MoveTicketStatus`; TUI wraps selection with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Ticket | `ticket-checklist-toggle` | Validate checklist toggle availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Ticket | `ticket-checklist-add` | Validate checklist item creation availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Work | `work-start-ticket` | Start selected ticket execution. | `chr ticket start <key>` | shared `buildTicketStartPreflightForTicket` and `runTicketStartExecution` | `matched` | CLI owns preflight/output; TUI service wraps selected-ticket execution. |
| Work | `work-start-ready-loop` | Start one ready-ticket loop. | `chr ticket loop --status ready_for_dev --max-iterations 1` | shared `runTicketLoopExecution` | `matched` | CLI supports configurable query/run options; TUI wraps one ready-ticket iteration. |
| Work | `work-run-all-ready` | Run all visible ready tickets with Dash filters and order. | `chr work run-all [filters]` | `service.StartAgentWorkLoop` -> `runTicketLoopExecution` | `matched` | CLI and TUI use the configurable agent-work facade. |
| Work | Dash queue observation and `[` / `]` | Observe eligibility, reasons, exclusions and bounded pages. | `chr work status` for retained work | `service.ObserveWorkPage` | `intentionally_interactive` | Web and TUI share the observation; filters and order apply before pagination. |
| Work | Dash `v` Preview Work | Inspect selected native scope and policy before confirmation. | `chr work ticket <key> --preview` | `service.PreviewSelectedWork` -> `service.ExecuteConfirmedWorkExtent` | `matched` | Canonical admission checks the displayed snapshot, selected revision and policy before one checkpoint. |
| Work | Dash `w` Workflow details | Read retained status without continuing work. | `chr work status <id>` | `service.WorkflowStatus` | `matched` | Opens the Current pane; empty retained state remains read-only. |
| Work | `work-repair-readiness` | Apply safe work-readiness repair. | `chr work doctor --fix-safe` | `core.RepairWorkReadiness` | `matched` | TUI exposes the safe repair from Dash when readiness reports one is available. |
| Work | `work-request-stop` | Request daemon stop. | `chr daemon stop` | shared `requestDaemonStopState` | `matched` | CLI owns output and dry-run; TUI reloads work state after the shared stop request. |

Command availability gates:

- Spec commands are available only in the Specs view.
- Plan commands are available only in the Plans view.
- Dash reads shared queue eligibility, readiness, current contexts, activity and
  results while selected. Old queue-query replies cannot replace a newer page
  or selection. Board selection remains independent and follows its key.
- Dash refreshes work-readiness diagnostics while selected and blocks agent-work
  start actions when readiness is blocked.
- `materialize-plan-tracking` is additionally available only when the selected
  plan exists and contains content.

## CLI Command Inventory

| Domain | CLI command | TUI equivalent | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| Shell | `chr` | Opens TUI | `runDefaultTUIAction`, `tui.Run` | `matched` | Root command launches the operational console. |
| Shell | `chr tui --dir <path>` | Opens TUI | `runTUICommandAction`, `tui.Run` | `matched` | Explicit TUI command. |
| Workspace | `chr init` | none | `chr.Workspace.Init` | `missing_tui` | Workspace bootstrap is CLI-only. |
| Runtime | `chr runtime init` | none | `chr.Workspace.InitRuntime` through runtime command | `missing_tui` | Runtime asset installation is CLI-only. |
| Runtime | `chr runtime validate` | Dash displays runtime evidence only | `runRuntimeValidate` | `missing_tui` | TUI does not run runtime validation directly. |
| Project Metadata | `chr metadata project set` | Project metadata save | `service.SaveProjectMetadata` | `matched` | CLI edits by flag; TUI edits interactively. |
| Use-Case Metadata | `chr metadata case set` | Cases metadata save | `service.SaveCaseMetadata` | `matched` | CLI creates or updates by slug. |
| Project Registry | `chr project add` | none | `runProjectAdd`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chr project list` | none | `runProjectList`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chr project current` | current project topbar | `runProjectCurrent`, user registry | `matched` | TUI shows resolved project, CLI prints active project. |
| Project Registry | `chr project select` | none | `runProjectSelect`, user registry | `missing_tui` | Registry selection is CLI-only. |
| Tickets | `chr ticket new` | Palette `create-ticket` partially | `service.CreateTicket`; TUI also wraps defaults with `service.CreateDefaultTicket` | `matched` | CLI supports metadata; TUI creates a default task ticket through the same service. |
| Tickets | `chr ticket list` | Board view | `runTicketList`, ticket index | `matched` | TUI board and CLI list inspect indexed tickets. |
| Tickets | `chr ticket board` | Board view and maximized Board | `service.Board` | `matched` | CLI prints Board totals and columns; TUI renders compact and maximized Board views. |
| Tickets | `chr ticket show` | Board selected ticket detail | `runTicketShow`, ticket index | `matched` | Equivalent inspection surface. |
| Tickets | `chr ticket comment` | Palette `ticket-add-comment` partially | `service.AddTicketComment`; TUI wraps default body with `service.AddSelectedTicketComment` | `matched` | CLI body is user-supplied; TUI body is fixed. |
| Tickets | `chr ticket status` | Palette ticket status commands | `service.MoveTicketStatus`; TUI wraps selected ticket with `service.MoveSelectedTicketStatus` | `matched` | Same service path. |
| Dumps | `chr dump list` | Dumps refresh | `service.ListDumps` | `matched` | Same service path. |
| Dumps | `chr dump new` | Dumps create | `service.CreateDump` | `matched` | CLI creates project-wide or use-case scoped dumps. |
| Dumps | `chr dump show` | Dumps editor load | `service.ReadDump` | `matched` | CLI prints content for a path. |
| Dumps | `chr dump save` | Dumps editor save | `service.SaveDump` | `matched` | CLI replaces content for a path. |
| Dumps | `chr dump append` | Dump append action | `service.AppendDump` | `matched` | CLI appends content for a path. |
| Dumps | `chr dump open` | Dumps external editor | `service.OpenDump` | `matched` | CLI prints the resolved editor command. |
| Runtime Quizzes | `chr quiz list` | Quizzes refresh | `service.ListQuizzes` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz start` | Quizzes start | `service.StartQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz resume` | Quizzes resume | `service.ResumeQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz answer` | Quizzes answer | `service.AnswerQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz skip` | Quizzes skip | `service.SkipQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz complete` | Quizzes complete | `service.CompleteQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz cancel` | Quizzes cancel | `service.CancelQuiz` | `matched` | Same service path. |
| Runtime Quizzes | `chr quiz mode` | Quizzes mode switch | `service.SwitchQuizMode` | `matched` | Same service path. |
| Specs | `chr spec list` | Specs refresh | `service.ListSpecs` | `matched` | Same service path. |
| Specs | `chr spec generate` | Specs generate | `service.GenerateSpec` | `matched` | Same service path. |
| Plans | `chr plan list` | Plans refresh | `service.ListPlans` | `matched` | Same service path. |
| Plans | `chr plan generate` | Plans generate | `service.GeneratePlan` | `matched` | Same service path. |
| Plans | `chr plan regenerate` | Plans regenerate | `service.RegeneratePlan` | `matched` | Same service path. |
| Plans | `chr plan tracking create` | Plan tracking materialization | `service.MaterializePlanTracking` | `matched` | Same service path. |
| Work | `chr ticket start` | Palette `work-start-ticket` | shared `buildTicketStartPreflightForTicket` and `runTicketStartExecution` | `matched` | Shared core execution path. |
| Work | `chr ticket loop` | Palette `work-start-ready-loop` partially | shared `runTicketLoopExecution` | `matched` | CLI is configurable; TUI runs one ready-ticket iteration. |
| Work | `chr ticket apply-result` | none | `runTicketApplyResult` | `missing_tui` | Result application is CLI-only. |
| Index | `chr index rebuild` | Workspace refresh partially | `runIndexRebuild`, ticket reconciler | `missing_tui` | CLI rebuilds ticket index; TUI refresh rebuilds workspace graph/read model. |
| Index | `chr index status` | Project/Dash diagnostics partially | `runIndexStatus` | `missing_tui` | CLI-specific index diagnostics. |
| Authoring | `chr author session new` | Palette `create-authoring-session` partially | `service.CreateAuthoringSession`; TUI wraps default session metadata with `service.CreateTUIAuthoringSession` | `matched` | CLI is configurable; TUI creates a default session through the same service. |
| Authoring | `chr author session list` | legacy Authoring view only | `runAuthorSessionList` | `missing_tui` | Authoring is not part of default TUI tabs. |
| Authoring | `chr author session show` | legacy Authoring view only | `runAuthorSessionShow` | `missing_tui` | Authoring is not part of default TUI tabs. |
| Authoring | `chr author quiz new` | Palette `create-context-quiz` not implemented | `runAuthorContextualQuizNew` | `missing_tui` | CLI supports context-aware authoring quiz generation. |
| Authoring | `chr author promote quiz-to-draft` | none | `runAuthorPromoteQuizToDraft`, `chr.PlanQuizToDraftPromotion` | `missing_tui` | Active CLI compatibility command; no TUI equivalent. |
| Assisted Authoring | `chr author assisted start` | none | `runAuthorAssistedStart` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chr author assisted resume` | none | `runAuthorAssistedResume` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chr author assisted answer` | none | `runAuthorAssistedAnswer` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chr author assisted status` | none | `runAuthorAssistedStatus` | `missing_tui` | CLI-only assisted workflow. |
| Assisted Authoring | `chr author assisted finalize` | none | `runAuthorAssistedFinalize` | `missing_tui` | CLI-only assisted workflow. |
| Daemon | `chr daemon status` | Dash displays observability partially | `runDaemonStatus` | `missing_tui` | TUI displays state but does not expose full daemon status command. |
| Daemon | `chr daemon start` | Palette `work-start-ready-loop` partially | `runDaemonStart` | `missing_tui` | Daemon foreground loop is CLI-only. |
| Daemon | `chr daemon stop` | Palette `work-request-stop` | shared `requestDaemonStopState` | `matched` | CLI owns output and dry-run; TUI reloads work state after the shared stop request. |
| Reports | `chr report list` | Dash displays recent evidence partially | `runReportList` | `missing_tui` | Report browsing is CLI-only. |
| Reports | `chr report show` | Dash displays recent evidence partially | `runReportShow` | `missing_tui` | Report browsing is CLI-only. |
| Maintenance | `chr maintenance bundle` | none | `runMaintenanceBundle` | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chr maintenance db status` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chr maintenance execution artifacts` | Dash evidence partially | maintenance command path | `missing_tui` | Detailed artifact drift reporting is CLI-only. |
| Maintenance | `chr maintenance execution cleanup-preview` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chr maintenance registry status` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chr maintenance registry disable` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Maintenance | `chr maintenance registry set-path` | none | maintenance command path | `missing_tui` | Maintenance is CLI-only. |
| Notify | `chr notify status` | none | `runNotifyStatus` | `missing_tui` | Notification inspection is CLI-only. |
| Notify | `chr notify telegram test` | none | `runNotifyTelegramTest` | `missing_tui` | Notification setup test is CLI-only. |
| Notify | `chr notify telegram poll` | none | `runNotifyTelegramPoll` | `missing_tui` | Remote command intake is CLI-only. |
| Proving | `chr proving validate` | none | `runProvingValidate` | `missing_tui` | Proving-ground workflows are CLI-only. |
| Proving | `chr proving run` | none | `runProvingRun` | `missing_tui` | Proving-ground workflows are CLI-only. |
| Proving | `chr proving ralph` | none | `runProvingRalph` | `missing_tui` | Proving-ground workflows are CLI-only. |

## Service Path Summary

| Service area | TUI path | CLI path | Status | Notes |
| --- | --- | --- | --- | --- |
| Workspace graph refresh | `tuiWorkspaceRefreshExecutor.RefreshWorkspace` -> `service.RefreshWorkspace` | no equivalent | `missing_cli` | CLI index commands do not refresh the complete workspace graph/read model. |
| Project metadata | `tuiProjectMetadataExecutor.SaveProjectMetadata` -> `service.SaveProjectMetadata` | `chr metadata project set` -> `service.SaveProjectMetadata` | `matched` | CLI edits by flag; TUI edits interactively. |
| Use-case metadata | `tuiCaseMetadataExecutor.SaveCaseMetadata` -> `service.SaveCaseMetadata` | `chr metadata case set` -> `service.SaveCaseMetadata` | `matched` | CLI creates or updates by slug. |
| Dump operations | `tuiDumpActionExecutor` -> `service.ListDumps`, `CreateDump`, `AppendDump`, `ReadDump`, `SaveDump`, `OpenDump` | `chr dump *` -> same service methods | `matched` | CLI is path-based and non-interactive. |
| Runtime quiz operations | `tuiQuizActionExecutor` -> `service.ListQuizzes`, `StartQuiz`, `ResumeQuiz`, `AnswerQuiz`, `SkipQuiz`, `CompleteQuiz`, `CancelQuiz`, `SwitchQuizMode` | `chr quiz *` -> same service methods | `matched` | Different from `chr author quiz new`; this is the runtime quiz loop. |
| Spec operations | `tuiSpecActionExecutor` -> `service.ListSpecs`, `GenerateSpec` | `chr spec list`, `chr spec generate` -> same service methods | `matched` | Same service path. |
| Plan operations | `tuiPlanActionExecutor` -> `service.ListPlans`, `GeneratePlan`, `RegeneratePlan`, `MaterializePlanTracking` | `chr plan *` -> same service methods | `matched` | Same service path. |
| Ticket creation | `tuiArtifactActionExecutor` -> `service.CreateDefaultTicket` -> `service.CreateTicket` | `runTicketNew` -> `service.CreateTicket` | `matched` | CLI behavior is richer; both use the same ticket creation service. |
| Board query | TUI Board views -> indexed ticket views | `runTicketBoard` -> `service.Board` | `matched` | CLI and TUI read the same Board columns and attention classification. |
| Ticket comment | `tuiTicketActionExecutor` -> `service.AddSelectedTicketComment` -> `service.AddTicketComment` | `runTicketComment` -> `service.AddTicketComment` | `matched` | CLI body is user-supplied; TUI supplies a default body. |
| Ticket status | `tuiTicketActionExecutor` -> `service.MoveSelectedTicketStatus` -> `service.MoveTicketStatus` | `runTicketStatus` -> `service.MoveTicketStatus` | `matched` | Same status service. |
| Work start ticket | `tuiWorkActionExecutor` -> `service.StartSelectedTicket` -> shared execution helpers | `runTicketStart` -> shared execution helpers | `matched` | TUI wraps selection and message reload; CLI owns dry-run and output. |
| Work ready loop | `tuiWorkActionExecutor` -> `service.StartReadyTicketLoop` -> `runTicketLoopExecution` | `runTicketLoop` -> `runTicketLoopExecution` | `matched` | TUI uses fixed ready/max-one selection; CLI is configurable. |
| Work run all ready | `tuiWorkActionExecutor` -> `service.StartAgentWorkLoop` -> `runTicketLoopExecution` | `chr work run-all` -> `service.StartAgentWorkLoop` | `matched` | Both paths use the agent-work facade and shared filter/order selection contract. |
| Work readiness repair | `tuiWorkActionExecutor` -> `core.RepairWorkReadiness` | `chr work doctor --fix-safe` -> `core.RepairWorkReadiness` | `matched` | TUI confirms safe repair and reloads readiness before the user retries agent work. |
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

- Workspace graph refresh remains only partially covered by `chr index rebuild`
  and `chr index status`.

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
