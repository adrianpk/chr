# Web and TUI workflow parity

Status: accepted
Date: 2026-10-29
Authorization: maintainer requested conversion of the supplied parity draft into specifications, slicing, tickets and Looper execution.
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on `dev`.

## Purpose

Make chr's terminal interface expose the same functional workflows as Web while preserving its established terminal interaction and visual language. Close all 38 source gaps and four incidental findings through one bounded delivery set.

Source: [assessed inventory](../report/web-tui-parity-discovery.md).
Execution: [plan](../plan/web-tui-parity.md), [tracker](../tracker/web-tui-parity.md), [coverage ledger](../plan/web-tui-parity-coverage.csv).

## Architecture and scope

The central application service remains the command/query owner. Terminal and Web adapters map neutral values to their respective views; neither adapter imports the other. Existing ticket, quiz, authoring, guided-state, work-selection and workflow-persistence owners retain their invariants. The terminal must not duplicate guided progression, queue eligibility, durable approvals, publication policy or retry accounting.

No new durable domain model or persistence schema is introduced, so no model companion is required. Case/run selection and editor drafts are presentation state. Keep current stored IDs internally for precise binding; present case names and functional identifiers to users. Do not require namespace or UUID entry for these workflows.

The source's tab decision is adopted. Web is the behavior reference; its markup, Web DTOs and incidental presentation are not terminal contracts. Existing TUI palette, borders, focus, F-keys, overlays and pane behavior remain the visual reference. No redesign of Web or Telegram is in scope.

## Requirements

### Requirement: Safe visible action targets

The TUI SHALL bind queue selection, previews and execution to the displayed Dash ticket key and captured observation. Board selection SHALL remain independent. Moves SHALL offer `backlog`, `ready_for_dev`, `in_progress`, `ready_for_review`, `ready_for_merge` and `done` through the same guarded revision-bound operation used by Web. Legacy input compatibility elsewhere SHALL remain intact. Native-managed fields, review evidence and integration authority MUST NOT be bypassed.

#### Scenario: Board and Dash select different tickets

- GIVEN Board selected ticket A and Dash displays selected queue ticket B
- WHEN the user previews or starts the selected Dash item
- THEN the scope is B, and any changed snapshot requires a fresh confirmation.

#### Scenario: Review evidence is missing

- GIVEN a ticket lacks the canonical review or merge evidence
- WHEN the user requests the corresponding Board move
- THEN the operation is rejected with an explanation and no alternate legacy spelling bypass.

### Requirement: Graceful busy interaction

Ctrl+C SHALL remain available during generation and other busy terminal modes through the existing host shutdown lifecycle. Quitting the UI MUST NOT imply cancellation of retained work or reset attempts, budgets or ownership.

#### Scenario: Quit while a command is pending

- GIVEN a busy editor or work observation
- WHEN Ctrl+C is pressed
- THEN the terminal exits cleanly, transient resources are released and retained execution state remains governed by its host.

### Requirement: Canonical navigation and help

The initial view SHALL be Cases. The tab order SHALL be F1 Cases, F2 Dumps, F3 Quizzes, F4 Specs, F5 Plans, F6 Board, F7 Dash, with F10 Exit. View identity SHALL own routing rather than hard-coded positional assumptions. Project SHALL be header context with alt+p and palette access; Help SHALL be functional through ? outside text entry and the palette. The actual shell footer SHALL show relevant view actions.

#### Scenario: Typing punctuation in a form

- GIVEN a text editor has focus
- WHEN ? or another printable character is typed
- THEN it remains text and does not invoke global Help or switch a view.

#### Scenario: Project leaves the tab bar

- GIVEN the seven-tab shell
- WHEN Project is opened
- THEN metadata editing, Harness selection, workspace health and useful legacy runtime information remain reachable.

### Requirement: Case-bound navigation

Opening a case or its retained workflow SHALL bind the corresponding document tabs and Board to that functional case. Navigation and observation SHALL not mutate documents or dispatch work. Invalid or removed bindings SHALL show a recoverable explanation rather than selecting a different case silently.

#### Scenario: Navigate within a retained case

- GIVEN a case-bound run is open
- WHEN the user changes among Dumps, Quizzes, Specs, Plans and Board
- THEN the same case remains selected and no hidden write occurs.

### Requirement: Shared operation feedback and draft safety

Writes SHALL distinguish ok, validation_failed, stale and blocked with summary, relevant findings, changed files and eligible reload or retry. The existing TUI Action overlay and established styling SHALL carry this feedback. User drafts SHALL survive errors and delayed/out-of-order observations. Reload/discard SHALL require confirmation when text would be lost.

#### Scenario: Concurrent save changes a document

- GIVEN a user has unsent changes and another writer advances the stored revision
- WHEN save is attempted
- THEN no overwrite occurs, the draft remains available and confirmed reload is offered.

#### Scenario: A late command finishes after focus changed

- GIVEN a newer form or selected record is active
- WHEN a result for an older captured interaction arrives
- THEN it cannot replace the current draft or execute the new selection.

### Requirement: Complete Case and Dump forms

Manual case creation SHALL collect Name, Description and Context before one Create. Dump creation SHALL collect project/case scope, title and body. Editing, appending and external editing SHALL use canonical document operations with consistent shortcut hints and revision checks. Existing case metadata, reference and artifact displays SHALL remain available.

#### Scenario: Cancel a new document form

- GIVEN entered but unsubmitted case or dump fields
- WHEN creation is cancelled
- THEN no placeholder case or fixed-content dump is created.

#### Scenario: Append to an existing dump

- GIVEN a selected dump and free-text append input
- WHEN the append is confirmed
- THEN only that dump receives the submitted content under its current revision contract.

### Requirement: Complete Spec catalog and manual editing

Specs SHALL list canonical manual, generated and project-wide records. Users SHALL create and edit title, body and draft state with create-time case binding and canonical validation. Generation and acceptance-feature generation SHALL remain available. These editors SHALL be the same components used when a wizard displays that stage.

#### Scenario: Inspect all supported Spec origins

- GIVEN owned manual, generated and project-wide Spec fixtures
- WHEN the catalog is opened
- THEN all appear with stable selection and can be inspected through their canonical source.

### Requirement: Complete Quiz interaction

Quizzes SHALL offer visible Deep/Fast selection before start, pending-target markers, current question and retained answer history. Answer, Skip, Done, Cancel, Resume and mode changes SHALL use the existing quiz owner. Submissions SHALL bind the exact session, current question, revision and input digest; stale or duplicate submission MUST NOT answer a different question. Manual and wizard mode SHALL use the same component.

#### Scenario: A question advances before a delayed answer

- GIVEN an answer bound to an earlier question
- WHEN it reaches the service after the question changed
- THEN the service rejects or reconciles that original operation without applying it to the current question, and the UI preserves useful input.

### Requirement: Plan generation prepares tickets

Manual Generate Plan SHALL invoke GeneratePlanDelivery to retain the plan and prepare its native tickets as one user operation. Retry SHALL reconcile the retained partial result without duplicate tickets or unnecessary regeneration. Explicit Regenerate SHALL retain confirmation and source constraints. Manual preparation SHALL NOT start execution. Successful Spec/Plan generation SHALL offer appropriate Open document and View tickets actions.

#### Scenario: Ticket preparation fails after a plan is saved

- GIVEN a retained plan and incomplete ticket preparation
- WHEN the user retries
- THEN the original operation completes the intended ticket set without a second plan or duplicate tickets.

### Requirement: Complete Ticket interactions

Ticket create/edit forms SHALL use user content, canonical metadata options, case binding and existing validation. Phase/slice and read-only reasons SHALL be visible where applicable. The TUI SHALL accept free-text comments, show comment history, attach review references and preserve external editing. Shared rules SHALL live behind neutral application operations rather than importing Web-specific forms.

#### Scenario: Edit a managed or stale ticket

- GIVEN managed read-only fields or a concurrent revision
- WHEN a user submits changes
- THEN the same ownership/revision restrictions as Web apply, with field findings and retained input.

### Requirement: Responsive Board and stable selection

At adequate width, Board SHALL show all six columns and Attention together. Narrow terminals SHALL retain usable column navigation and focus. Refresh/reordering SHALL preserve selected ticket keys with a deterministic visible fallback only when the selected record disappears. Filtering, pane resizing and maximize SHALL remain available.

#### Scenario: Background refresh reorders cards

- GIVEN a selected ticket still exists after a refresh
- WHEN the order changes
- THEN focus remains on that key rather than the previous index or first ticket.

### Requirement: Shared Dash selection and activity

Dash SHALL use service.ObserveWork for eligibility, totals, readiness reasons, exclusions and pagination. It SHALL expose current work, recent activity, results, Preview Work and workflow details. Existing queue filters/order, Run all ready, safe readiness repair and running-work details SHALL remain available through the same canonical selection rules.

#### Scenario: A ready-looking ticket is excluded by policy

- GIVEN a ticket with a ready status but failing canonical eligibility
- WHEN Dash is rendered
- THEN the observation explains the exclusion and does not locally advertise it as executable.

### Requirement: Retained Workflow interactions

Dash SHALL provide a paginated retained workflow list and detail with members, evidence and eligible controls. Preview SHALL support ticket, slice, set, case and loose extents with canonical filters/order. Allowed allowance grants and integration-decision answers SHALL be exposed with existing revision, authorization and finite-budget guards. Cases and tickets SHALL offer visible entry points.

#### Scenario: An allowance or decision changes before confirmation

- GIVEN a bound workflow control form
- WHEN the retained revision or decision has changed
- THEN the stale operation cannot grant another allowance or answer another decision.

### Requirement: Wizard entry in the shared shell

Cases SHALL expose Wizard and Open workflow. Starting SHALL take one Name/Create action through StartNamedGuidedUseCase. The wizard SHALL use the existing tabs and shared forms, with case name and stage progress. It SHALL NOT add a second navigation menu, duplicate near-identical forms, require namespace input or foreground technical metadata.

| Domain stage | Existing tab | Interaction |
| --- | --- | --- |
| name | Cases | Name and Create |
| dump | Dumps | Shared Dump editor and Continue to quiz |
| quiz | Quizzes | Shared Quiz component |
| specification | Specs | Shared document presentation and domain-owned preparation |
| preparation | Plans | Shared plan/ticket preparation presentation |
| execution | Board | Case-filtered work and review |
| results | Board | Receipt-backed delivery and validation results |

#### Scenario: Browse an earlier wizard stage

- GIVEN a run has advanced beyond its dump
- WHEN Dumps is opened for the case
- THEN the committed original is read-only and Continue returns to the current stage without resubmission.

#### Scenario: Continue with edited dump content

- GIVEN unsent content in the shared wizard Dump editor
- WHEN Continue to quiz is selected
- THEN the current content is submitted exactly once through the retained finish operation, not lost between separate append/finish actions.

### Requirement: Mode-specific controls

Manual forms SHALL expose manual save/generate actions. Wizard forms SHALL expose the current stage's progression actions. Actions from the other mode SHALL not accidentally save, terminate or fork the current flow. Shared component ownership SHALL be demonstrable in code and runtime tests; visual resemblance alone is insufficient.

#### Scenario: Wizard dump is being edited

- GIVEN wizard mode is active
- WHEN the Dump actions are rendered
- THEN Continue to quiz is the progression action and an independent manual Save Dump cannot break the flow.

### Requirement: Domain-owned wizard progression and recovery

The UI SHALL observe existing guided state and submit existing guided commands. Domain-owned specification auto-approval, committed original inputs, ticket preparation, dispatch and finite retries SHALL remain unchanged. Processing, suspended, pending-operation and blocker states SHALL explain what happened and expose only eligible actions. Retry SHALL reuse the original retained operation and inputs.

#### Scenario: Restart after an interrupted preparation

- GIVEN a retained pending operation
- WHEN the UI restarts and Retry is confirmed
- THEN the original result is reconciled without duplicate artifacts, new budgets or UI-owned stage orchestration.

### Requirement: Wizard execution and evidence-backed results

Execution SHALL use the shared case-filtered Board. Progress SHALL distinguish validation, confirmed delivery and remaining work. Results SHALL expose summary, branch/PR, delivery receipts, validation commands and exit codes. A successful worker, validation or open PR SHALL NOT be represented as confirmed integration. JSON export, committed originals and later ideas SHALL remain secondary existing capabilities.

#### Scenario: Validation succeeds but integration is absent

- GIVEN validated tasks and an open PR without retained integration receipts
- WHEN progress or results are shown
- THEN the UI reports pending delivery and the eligible next action, not completion.

### Requirement: Safe cancellation and passive observation

Cancel workflow SHALL require confirmation with: "Saved documents and completed changes will be kept." Pause SHALL appear only where existing domain policy permits it, particularly active execution. Polling SHALL be read-only and preserve unsent text, focus and the user-selected view. Pending text SHALL prevent automatic navigation and show that it remains available.

#### Scenario: Polling observes a newer stage while the user types

- GIVEN unsent text in a stage form
- WHEN the retained run changes
- THEN polling neither replaces the draft nor redirects nor advances work; it offers a context-preserving next action.

### Requirement: Honest palette and retained capabilities

Every advertised action SHALL invoke a supported operation or explain its unavailability before confirmation. The set SHALL not invent a new unsafe checklist writer to make an existing stub appear functional. Useful legacy capabilities and all extras listed in the discovery inventory SHALL retain reachable, coherent destinations.

#### Scenario: No safe checklist writer exists

- GIVEN the application cannot safely change a checklist
- WHEN the user discovers the action
- THEN it is unavailable with a clear reason instead of asking for confirmation and always failing.

### Requirement: Complete parity evidence

All 38 G identifiers and four I identifiers SHALL have an owning slice, acceptance evidence and final disposition. G-008, G-013, G-014, G-018 and G-024 SHALL receive owned runtime verification rather than inferred static closure. Tests SHALL exercise production adapters and real terminal/HTTP/persistence boundaries, with deterministic external model/provider responses as appropriate. Existing gate and signature constraints SHALL remain enabled.

#### Scenario: A flow passes only a reducer test

- GIVEN a parity flow has unit coverage
- WHEN delivery is assessed
- THEN a real Bubble Tea PTY journey and relevant Web reference fixtures still establish binding, focus, persistence and recovery behavior before final acceptance.

## Non-goals and approval boundary

Do not change guided domain semantics, storage formats, publication policies, branch topology, authentication requirements or provider dependencies. Do not implement imports/adoption of partly manual use cases as part of this set. Do not add a generic UI framework, terminal transport of Web DTOs, new entities or a second execution controller. Do not promote main, publish mirrors, tag, deploy production or mutate unrelated active runs. Changes needed beyond this bounded parity scope require maintainer approval.
