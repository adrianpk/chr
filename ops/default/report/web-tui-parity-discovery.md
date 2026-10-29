# Web and TUI parity discovery

Status: assessed for implementation planning
Assessed: 2026-10-29T10:11:36Z
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on `dev`; clean working tree.
Source: `relevamiento-paridad-web-tui.md`, independent static draft supplied by the maintainer.
Source SHA-256: `a24e6494bf9a100f5240b01ff2f3f9f2b4481dc1ffc213b70e768caa55432964`

## Evidence boundary

The original draft was read in full and preserved byte-for-byte outside the repository. This English inventory preserves all 38 stable G identifiers, the four incidental findings as I-001 through I-004, the decided navigation, acceptable terminal differences, and useful TUI extras. It is not a second independent audit and does not claim that either interface was executed during planning. Paths identify production evidence; each implementation slice must add and record focused behavioral evidence.

Verified priority counts from the individual G rows: P0: 2, P1: 14, P2: 16, P3: 6. Incidental priorities below are delivery triage, not additions to the draft's 38 gaps.

## Validated qualifications

- Both adapters already use the central application service. Parity requires adapter presentation and missing command/query exposure, not another domain engine.
- G-034 is supported by the local Dash queue and shared `selectedTicket` path: a distinct visible queue key must own start and preview.
- G-026 concerns more than labels. `ValidStatus` accepts legacy aliases; `ticketStatusUpdater.UpdateStatus` writes the supplied accepted value. Review guards compare canonical review states, so the new terminal path must use the same canonical, revision-bound guarded move as Web. Do not ban legacy input globally or claim runtime exploitation was performed.
- G-001 is already decided in the source. Normalize G-002's old F3-F7 positional wording to the new view IDs: Dumps through Board are F2-F6. Project and Help leave the tab bar, not the product.
- `DashPageView` and `completeTicketEditor` are Web-specific projections. Reuse their neutral application owners and canonical options, not their Web DTOs in terminal/domain packages.
- A behavior-only change does not require a persistence model companion. This set adds no durable entities, migration, generic result framework, package-wide rewrite or new dependencies.
- G-008, G-013, G-014, G-018 and G-024 remain runtime-unverified at planning. Their owning slices and final acceptance use disposable owned fixtures to resolve them.
- Useful legacy capabilities need a reachable destination; unreachable view declarations are not permission for blanket deletion. Unsupported checklist writers must be unavailable honestly rather than replaced by a new unsafe writer.

## Complete finding inventory

| ID | Priority | Owning slice | Required outcome | Baseline paths | Evidence boundary |
| --- | --- | --- | --- | --- | --- |
| G-001 | P2 | Slice 2 | Cases starts the seven-tab shell; Project and Help remain accessible outside the tab bar. | internal/tui/model.go; internal/tui/navigation.go; internal/tui/shell_chrome.go | Static evidence; runtime acceptance required |
| G-002 | P1 | Slice 2 | Case and retained-run context follows views by View ID, without carrying hidden form mutations. | internal/tui/model_state.go; internal/tui/navigation.go; internal/app/web_guided_use_case.go | Static evidence; runtime acceptance required |
| G-003 | P2 | Slice 2 | The actual shell footer exposes active-view actions with the existing visual style. | internal/tui/model_view.go; internal/tui/shell_chrome.go | Static evidence; runtime acceptance required |
| G-004 | P3 | Slice 8 | Successful generation offers Open document and View tickets when the result supports them. | internal/app/tui_adapters_spec_plan_ticket.go; internal/app/web_plan.go | Static evidence; runtime acceptance required |
| G-005 | P1 | Slice 11 | Cases exposes Wizard and Open workflow for the selected case. | internal/tui/guided.go; internal/app/tui_adapters_guided.go | Static evidence; runtime acceptance required |
| G-006 | P1 | Slice 11 | One Name/Create operation calls StartNamedGuidedUseCase. | internal/app/tui_adapters_guided.go; internal/app/service_guided_start.go | Static evidence; runtime acceptance required |
| G-007 | P1 | Slice 11 | The wizard traverses the existing tabs and reuses the same forms. | internal/tui/guided.go; internal/tui/model.go | Static evidence; runtime acceptance required |
| G-008 | P1 | Slice 11 | Continue to quiz submits the current dump once and retains the original pending operation for retry. | internal/app/tui_adapters_guided.go; internal/tui/guided.go | Runtime verification required |
| G-009 | P1 | Slice 12 | Wizard quizzes expose the shared mode, history, skip, answer, cancel and done controls with exact question binding. | internal/tui/guided.go; internal/app/tui_adapters_dump_quiz.go | Static evidence; runtime acceptance required |
| G-010 | P2 | Slice 12 | Each processing or recovery state gives a stage-specific explanation and the eligible retained action. | internal/tui/guided.go; internal/app/tui_adapters_guided.go | Static evidence; runtime acceptance required |
| G-011 | P2 | Slice 13 | Wizard execution uses the case-filtered Board and shared progress, review and lifecycle controls. | internal/tui/guided.go; internal/tui/work_actions.go | Static evidence; runtime acceptance required |
| G-012 | P2 | Slice 13 | Results expose human-readable deliveries, validation commands and exit codes, with JSON as an extra. | internal/app/tui_adapters_guided.go; internal/tui/guided.go | Static evidence; runtime acceptance required |
| G-013 | P2 | Slice 13 | Cancellation confirms that saved documents and completed changes will be kept. | internal/tui/guided.go; internal/app/tui_adapters_guided.go | Runtime verification required |
| G-014 | P2 | Slice 13 | Passive polling preserves unsent text, selection and navigation without advancing work. | internal/tui/guided.go; internal/tui/model_update.go | Runtime verification required |
| G-015 | P3 | Slice 5 | Manual case creation collects Name, Description and Context before one Create. | internal/tui/case_metadata_actions.go; internal/app/tui_adapters_artifact.go | Static evidence; runtime acceptance required |
| G-016 | P3 | Slice 10 | Cases visibly exposes work preview and retained workflows. | internal/tui/workflow_commands.go; internal/tui/command_palette.go | Static evidence; runtime acceptance required |
| G-017 | P1 | Slice 5 | Dump creation collects scope, title and content rather than fixed placeholder content. | internal/app/tui_adapters_dump_quiz.go; internal/tui/dumps.go | Static evidence; runtime acceptance required |
| G-018 | P2 | Slice 5 | Stale dump saves retain the draft and require confirmation before reload or discard. | internal/app/tui_adapters_dump_quiz.go; internal/tui/dump_actions.go | Runtime verification required |
| G-019 | P3 | Slice 5 | External-editor bindings and hints agree inside and outside the textarea. | internal/tui/dump_actions.go; internal/tui/navigation.go | Static evidence; runtime acceptance required |
| G-020 | P2 | Slice 7 | Deep/Fast selection is visible before a quiz starts. | internal/app/tui_adapters_dump_quiz.go; internal/tui/quizzes.go | Static evidence; runtime acceptance required |
| G-021 | P2 | Slice 7 | Answer history remains visible with the active quiz question. | internal/tui/quizzes.go; internal/app/tui_adapters_dump_quiz.go | Static evidence; runtime acceptance required |
| G-022 | P3 | Slice 7 | Quiz targets needing answers are identified in the list. | internal/tui/quizzes.go; internal/app/tui_adapters_dump_quiz.go | Static evidence; runtime acceptance required |
| G-023 | P1 | Slice 6 | Specs supports manual create and revision-checked editing with title, case, body and draft state. | internal/app/web_spec.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |
| G-024 | P2 | Slice 6 | The Spec catalog includes manual, generated and project-wide records. | internal/app/web_spec.go; internal/app/tui_adapters_spec_plan_ticket.go | Runtime verification required |
| G-025 | P1 | Slice 8 | Generate Plan calls GeneratePlanDelivery and reconciles partial ticket preparation without duplicates. | internal/app/web_plan.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |
| G-026 | P0 | Slice 1 | Moves use all six canonical Board states and the same revision, ownership and evidence guards as Web. | internal/tui/ticket_actions.go; internal/app/service_artifact_ticket.go; internal/core/workflow/ticket_status.go; internal/core/ticket/status.go | Static evidence; runtime acceptance required |
| G-027 | P1 | Slice 9 | Ticket creation collects user values and canonical metadata options. | internal/app/web_ticket_form.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |
| G-028 | P1 | Slice 10 | Dash offers paginated retained workflows, preview filters, evidence, eligible controls, allowance and integration decisions. | internal/tui/workflow_commands.go; internal/app/web_workflow.go | Static evidence; runtime acceptance required |
| G-029 | P1 | Slice 9 | Ticket editing exposes canonical fields, phase/slice and reasons for read-only values. | internal/app/web_ticket_form.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |
| G-030 | P1 | Slice 9 | Tickets accept free-text comments and show comment history. | internal/app/tui_adapters_spec_plan_ticket.go; internal/tui/ticket_actions.go | Static evidence; runtime acceptance required |
| G-031 | P2 | Slice 10 | Wide Board shows all columns and Attention; narrow Board remains usable. | internal/tui/ticket_board.go; internal/tui/model_view.go | Static evidence; runtime acceptance required |
| G-032 | P2 | Slice 4 | Refresh preserves Board selection by ticket key independently from Dash selection. | internal/tui/agent_work_refresh.go; internal/tui/dash.go | Static evidence; runtime acceptance required |
| G-033 | P3 | Slice 2 | Contextual Help is real, reachable with ? outside text input, and remains in the palette. | internal/tui/model.go; internal/tui/navigation.go | Static evidence; runtime acceptance required |
| G-034 | P0 | Slice 1 | Dash owns its selected queue key; preview and execution act on that displayed target. | internal/tui/dash.go; internal/tui/work_actions.go; internal/app/service_work.go | Static evidence; runtime acceptance required |
| G-035 | P1 | Slice 4 | Dash uses ObserveWork eligibility, counts, reasons, exclusions and pagination. | internal/app/web_dash.go; internal/tui/dash.go | Static evidence; runtime acceptance required |
| G-036 | P2 | Slice 4 | Dash shows recent activity and results without requiring the running-work overlay. | internal/app/web_dash.go; internal/tui/dash.go | Static evidence; runtime acceptance required |
| G-037 | P2 | Slice 4 | Preview Work is visible, bound to its scope and policy, and requires confirmation. | internal/tui/workflow_commands.go; internal/app/web_dash.go | Static evidence; runtime acceptance required |
| G-038 | P2 | Slice 3 | A reusable operation-result panel distinguishes ok, validation_failed, stale and blocked with findings and reload. | internal/tui/status_messages.go; internal/webui/views.go | Static evidence; runtime acceptance required |
| I-001 | P1 | Slice 1 | Ctrl+C can quit a busy UI through the existing graceful host lifecycle without silently cancelling retained work. | internal/tui/model_update.go | Static evidence; runtime acceptance required |
| I-002 | P2 | Slice 2 | Useful legacy-view capabilities remain reachable through Project, Dash or the palette. | internal/tui/model.go; internal/tui/command_palette.go | Static evidence; runtime acceptance required |
| I-003 | P2 | Slice 14 | Palette actions either invoke a supported operation or clearly explain unavailability before confirmation. | internal/tui/command_palette.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |
| I-004 | P2 | Slice 9 | AppendDump and AttachReviewRef have reachable user-input actions; dump wiring is delivered in Slice 5. | internal/app/tui_adapters_dump_quiz.go; internal/app/tui_adapters_spec_plan_ticket.go | Static evidence; runtime acceptance required |

## Preserved terminal conventions and extras

Keep palette, borders, F-key navigation, keyboard focus, footer clicks, pane resizing, maximize, terminal overlays and the existing polling mechanism. Parity means the same available actions, location, order and outcomes, not pixel matching or HTML inside a terminal.

Keep Project metadata and Harness selection; unresolved references and artifact counts; 17-field ticket filtering; external editing; acceptance-feature generation; safe readiness repair; queue filters/order and Run all ready; running-work events and log locations; later ideas, committed original documents and JSON exports. These extras must not become mandatory fields or foreground sections in the normal wizard.

## Traceability

[Requirements](../spec/web-tui-parity.md), [delivery plan](../plan/web-tui-parity.md), [coverage ledger](../plan/web-tui-parity-coverage.csv), and [tracker](../tracker/web-tui-parity.md) own implementation scope and delivery evidence. The source draft is an input, not executable instructions.
