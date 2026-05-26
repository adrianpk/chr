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

## Command Palette Commands

| Domain | Command palette id | TUI action | CLI command | Service path | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Workspace | `refresh` | Reload workspace files into runtime graph. | `chc index rebuild`, `chc index status` partially | `service.RefreshWorkspace` | `missing_cli` | CLI index commands target ticket index, not the full TUI workspace graph refresh. |
| Help | `help` | Navigate to Help view. | `chc --help` | `internal/chctui.Model.selectView`, CLI help | `matched` | Equivalent intent, different presentation. |
| Selection | `open-source` | Show selected source path. | none | `internal/chctui.Model.selectedSourcePath` | `missing_cli` | CLI has path outputs in list/show commands, but no selected-source concept. |
| Ticket | `create-ticket` | Create default task ticket. | `chc ticket new <title>` | TUI: `service.CreateDefaultTicket`; CLI: `runTicketNew` and ticket renderer | `service_divergence` | CLI can create tickets, but TUI default creation uses a dedicated service method. |
| Dump | `create-dump` | Create selected project/use-case dump. | none | `service.CreateDump` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Dump | `edit-dump` | Open selected dump in TUI editor. | none | `service.ReadDump`, `service.SaveDump` | `missing_cli` | Service exists; CLI entrypoint missing. |
| Dump | `refresh-dumps` | Reload dump views. | none | `service.ListDumps` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-start` | Start selected runtime quiz. | none | `service.StartQuiz` | `missing_cli` | Different from authoring quiz generation. |
| Quiz | `quiz-resume` | Resume selected runtime quiz. | none | `service.ResumeQuiz` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-answer` | Answer active runtime quiz question. | none | `service.AnswerQuiz` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-skip` | Skip active runtime quiz question. | none | `service.SkipQuiz` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-complete` | Complete selected runtime quiz. | none | `service.CompleteQuiz` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-cancel` | Cancel selected runtime quiz. | none | `service.CancelQuiz` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-refresh` | Reload runtime quiz state. | none | `service.ListQuizzes` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-mode-deep` | Switch selected runtime quiz to deep mode. | none | `service.SwitchQuizMode` | `missing_cli` | CLI entrypoint missing. |
| Quiz | `quiz-mode-fast` | Switch selected runtime quiz to fast mode. | none | `service.SwitchQuizMode` | `missing_cli` | CLI entrypoint missing. |
| Spec | `generate-spec` | Generate selected spec. | none | `service.GenerateSpec` | `missing_cli` | Command is gated to Specs view. |
| Spec | `refresh-specs` | Reload spec views. | none | `service.ListSpecs` | `missing_cli` | Command is gated to Specs view. |
| Plan | `generate-plan` | Generate missing selected plan. | none | `service.GeneratePlan` | `missing_cli` | Command is gated to Plans view. |
| Plan | `regenerate-plan` | Regenerate selected plan. | none | `service.RegeneratePlan` | `missing_cli` | Command is gated to Plans view. |
| Plan | `materialize-plan-tracking` | Create plan tracking and tickets. | none | `service.MaterializePlanTracking` | `missing_cli` | Command is gated to Plans and only available for existing plans. |
| Plan | `refresh-plans` | Reload plan views. | none | `service.ListPlans` | `missing_cli` | Command is gated to Plans view. |
| Authoring | `create-authoring-session` | Create TUI default authoring session. | `chc author session new ...` | TUI: `service.CreateTUIAuthoringSession`; CLI: `runAuthorSessionNew` | `service_divergence` | Equivalent artifact family, but TUI default helper and CLI command paths differ. |
| Authoring | `create-context-quiz` | Validate context-aware quiz creation availability. | `chc author quiz new ...` | TUI returns unavailable; CLI `runAuthorContextualQuizNew` | `missing_tui` | CLI supports context-aware quiz generation; TUI action is not implemented. |
| Ticket | `ticket-add-comment` | Add default selected-ticket comment. | `chc ticket comment <key> <body>` | TUI: `service.AddSelectedTicketComment`; CLI: `runTicketComment` | `service_divergence` | CLI supports arbitrary body; TUI uses default body through service. |
| Ticket | `ticket-status-ready` | Move selected ticket to `ready`. | `chc ticket status <key> ready` | TUI: `service.MoveSelectedTicketStatus`; CLI: `runTicketStatus` | `service_divergence` | Equivalent behavior, paths differ. |
| Ticket | `ticket-status-in-progress` | Move selected ticket to `in_progress`. | `chc ticket status <key> in_progress` | TUI: `service.MoveSelectedTicketStatus`; CLI: `runTicketStatus` | `service_divergence` | Equivalent behavior, paths differ. |
| Ticket | `ticket-status-reviewing` | Move selected ticket to `reviewing`. | `chc ticket status <key> reviewing` | TUI: `service.MoveSelectedTicketStatus`; CLI: `runTicketStatus` | `service_divergence` | Equivalent behavior, paths differ. |
| Ticket | `ticket-checklist-toggle` | Validate checklist toggle availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Ticket | `ticket-checklist-add` | Validate checklist item creation availability. | none | TUI returns unavailable | `not_implemented` | Safe checklist writer is not implemented. |
| Work | `work-start-ticket` | Start selected ticket execution. | `chc ticket start <key>` | TUI: `service.StartSelectedTicket`; CLI: `runTicketStart` | `service_divergence` | Equivalent behavior, paths differ. |
| Work | `work-start-ready-loop` | Start one ready-ticket loop. | `chc ticket loop --status ready --max-iterations 1` | TUI: `service.StartReadyTicketLoop`; CLI: `runTicketLoop` | `service_divergence` | Equivalent behavior, paths differ. |
| Work | `work-request-stop` | Request daemon stop. | `chc daemon stop` | TUI: `service.RequestDaemonStop`; CLI: `runDaemonStop` | `service_divergence` | Equivalent behavior, paths differ. |

Command availability gates:

- Spec commands are available only in the Specs view.
- Plan commands are available only in the Plans view.
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
| Project Registry | `chc project add` | none | `runProjectAdd`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chc project list` | none | `runProjectList`, user registry | `missing_tui` | Registry management is CLI-only. |
| Project Registry | `chc project current` | current project topbar | `runProjectCurrent`, user registry | `matched` | TUI shows resolved project, CLI prints active project. |
| Project Registry | `chc project select` | none | `runProjectSelect`, user registry | `missing_tui` | Registry selection is CLI-only. |
| Tickets | `chc ticket new` | Palette `create-ticket` partially | CLI: `runTicketNew`; TUI: `service.CreateDefaultTicket` | `service_divergence` | CLI supports metadata; TUI creates a default task ticket. |
| Tickets | `chc ticket list` | Board view | `runTicketList`, ticket index | `matched` | TUI board and CLI list inspect indexed tickets. |
| Tickets | `chc ticket show` | Board selected ticket detail | `runTicketShow`, ticket index | `matched` | Equivalent inspection surface. |
| Tickets | `chc ticket comment` | Palette `ticket-add-comment` partially | CLI: `runTicketComment`; TUI: `service.AddSelectedTicketComment` | `service_divergence` | CLI body is user-supplied; TUI body is fixed. |
| Tickets | `chc ticket status` | Palette ticket status commands | CLI: `runTicketStatus`; TUI: `service.MoveSelectedTicketStatus` | `service_divergence` | Equivalent state change, different path. |
| Work | `chc ticket start` | Palette `work-start-ticket` | CLI: `runTicketStart`; TUI: `service.StartSelectedTicket` | `service_divergence` | Equivalent execution intent, different path. |
| Work | `chc ticket loop` | Palette `work-start-ready-loop` partially | CLI: `runTicketLoop`; TUI: `service.StartReadyTicketLoop` | `service_divergence` | CLI is configurable; TUI runs one ready-ticket iteration. |
| Work | `chc ticket apply-result` | none | `runTicketApplyResult` | `missing_tui` | Result application is CLI-only. |
| Index | `chc index rebuild` | Workspace refresh partially | `runIndexRebuild`, ticket reconciler | `missing_tui` | CLI rebuilds ticket index; TUI refresh rebuilds workspace graph/read model. |
| Index | `chc index status` | Project/Dash diagnostics partially | `runIndexStatus` | `missing_tui` | CLI-specific index diagnostics. |
| Authoring | `chc author session new` | Palette `create-authoring-session` partially | CLI: `runAuthorSessionNew`; TUI: `service.CreateTUIAuthoringSession` | `service_divergence` | CLI is configurable; TUI creates a default session. |
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
| Daemon | `chc daemon stop` | Palette `work-request-stop` | CLI: `runDaemonStop`; TUI: `service.RequestDaemonStop` | `service_divergence` | Equivalent stop intent, different path. |
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
| Project metadata | `tuiProjectMetadataExecutor.SaveProjectMetadata` -> `service.SaveProjectMetadata` | no equivalent | `missing_cli` | Service is ready for CLI exposure. |
| Use-case metadata | `tuiCaseMetadataExecutor.SaveCaseMetadata` -> `service.SaveCaseMetadata` | no equivalent | `missing_cli` | Service is ready for CLI exposure. |
| Dump operations | `tuiDumpActionExecutor` -> `service.ListDumps`, `CreateDump`, `AppendDump`, `ReadDump`, `SaveDump`, `OpenDump` | no equivalent | `missing_cli` | Service is ready for CLI exposure. |
| Runtime quiz operations | `tuiQuizActionExecutor` -> `service.ListQuizzes`, `StartQuiz`, `ResumeQuiz`, `AnswerQuiz`, `SkipQuiz`, `CompleteQuiz`, `CancelQuiz`, `SwitchQuizMode` | no equivalent | `missing_cli` | Different from `chc author quiz new`; this is the TUI runtime quiz loop. |
| Spec operations | `tuiSpecActionExecutor` -> `service.ListSpecs`, `GenerateSpec` | no equivalent | `missing_cli` | Service is ready for CLI exposure. |
| Plan operations | `tuiPlanActionExecutor` -> `service.ListPlans`, `GeneratePlan`, `RegeneratePlan`, `MaterializePlanTracking` | no equivalent | `missing_cli` | Service is ready for CLI exposure. |
| Ticket creation | `tuiArtifactActionExecutor` -> `service.CreateDefaultTicket` | `runTicketNew` -> lower-level ticket renderer/repository path | `service_divergence` | CLI behavior is richer; TUI default helper should be reconciled with shared service command shape. |
| Ticket comment | `tuiTicketActionExecutor` -> `service.AddSelectedTicketComment` -> `service.AddTicketComment` | `runTicketComment` -> lower-level commenter/reconcile path | `service_divergence` | Both write comments; CLI does not use the service method. |
| Ticket status | `tuiTicketActionExecutor` -> `service.MoveSelectedTicketStatus` -> `service.MoveTicketStatus` | `runTicketStatus` -> lower-level status updater/reconcile path | `service_divergence` | Both update status; CLI does not use the service method. |
| Work start ticket | `tuiWorkActionExecutor` -> `service.StartSelectedTicket` | `runTicketStart` | `service_divergence` | Both run one ticket; service boundary should be consolidated or explicitly separated. |
| Work ready loop | `tuiWorkActionExecutor` -> `service.StartReadyTicketLoop` | `runTicketLoop` | `service_divergence` | TUI wraps one ready iteration; CLI exposes configurable loop options. |
| Work request stop | `tuiWorkActionExecutor` -> `service.RequestDaemonStop` | `runDaemonStop` | `service_divergence` | Equivalent intent, different command path. |
| Authoring session creation | `tuiArtifactActionExecutor` -> `service.CreateTUIAuthoringSession` | `runAuthorSessionNew` | `service_divergence` | CLI is configurable; TUI creates a default spec-oriented session. |
| Context-aware authoring quiz | TUI returns unavailable | `runAuthorContextualQuizNew` | `missing_tui` | CLI path exists and should inform future TUI implementation. |
| Assisted authoring | no default TUI path | `runAuthorAssisted*` | `missing_tui` | CLI-only workflow. |
| Project registry | no TUI path | `runProject*` | `missing_tui` | User-level registry management is CLI-only. |
| Maintenance/notify/proving | no TUI path | command-specific paths | `missing_tui` | Operational/admin command families are CLI-only. |

## Gap Summary

High-priority `missing_cli` gaps:

- Project metadata save.
- Use-case metadata create/save.
- Dump list/create/read/save/open.
- Runtime quiz list/start/resume/answer/skip/complete/cancel/mode.
- Spec list/generate.
- Plan list/generate/regenerate/tracking.

High-priority `service_divergence` gaps:

- Ticket creation.
- Ticket comment.
- Ticket status.
- Ticket start.
- Ready-ticket loop.
- Daemon stop.
- Authoring session creation.

Current `not_implemented` actions:

- TUI checklist toggle.
- TUI checklist item creation.

Important `missing_tui` command families:

- Workspace initialization and runtime setup.
- Project registry management.
- CLI authoring workflows.
- Reports, maintenance, notifications, and proving-ground workflows.
