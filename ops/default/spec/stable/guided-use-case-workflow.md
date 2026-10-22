# Guided use-case workflow

Status: approved
Approved: maintainer, 2026-10-22; specification and delivery planning scope accepted.
Source: maintainer discussion, 2026-10-22
Model: [Guided use-case workflow model](guided-use-case-workflow-model.md)

## Purpose

Provide the canonical guided path from a use-case idea to completed development. chr owns the sequence and carries context between stages; the user supplies a name, a knowledge dump, and quiz answers. After clarification, chr generates and approves the specification, prepares the plan, tracker and native tickets, and starts executing that work automatically.

The same persistent workflow is available through Web, Telegram and TUI. Delivery starts with the shared application behavior and Web, then Telegram, then TUI. This is the intended delivery order, not an implementation plan.

Independent operations remain available: users may create tickets directly, capture dumps without answering quizzes, write specifications manually, or generate plans without starting execution.

## Scope and agreed decisions

1. The guided sequence is: name, dump, quiz, specification, work preparation, execution, results.
2. The generated specification is approved by default inside this workflow. There is no routine approval prompt between clarification and planning.
3. Work preparation produces an inspectable plan, its tracker and executable native tickets. Execution starts automatically after successful preparation and native admission.
4. The initiating action communicates that this journey proceeds through development automatically. Existing project execution and publication policies remain effective.
5. The workflow asks for human input when information, a decision, or required authority is actually missing. Ordinary stage completion does not introduce another confirmation.
6. Progress and recovery belong to the shared application workflow, independently of the active interface or connection.
7. Initial scope supports new use cases and resumption of guided runs. Adoption of manually started work is deferred.
8. Each completed authoring or work-preparation stage is sealed with a Git commit. Downstream stages consume the recorded committed versions.
9. Later ideas are retained for review after the current development completes. They do not automatically change, pause or regenerate the committed scope.

This document specifies a new capability. Existing authoring operations do not acquire automatic approval or execution behavior merely because the guided workflow composes them.

## Actors, inputs and outputs

The initiating user selects the project and namespace and supplies the use-case name, dump and quiz answers. chr coordinates existing authoring, planning, materialization and native execution owners. The configured model provider generates authoring content; it cannot grant authority or declare native delivery complete.

Inputs include the selected project context, explicitly referenced material, current project policy, provider readiness and execution limits. Unrelated conversations, cases and tickets are not implicit inputs.

Outputs are a retained guided run, its authoring session, dump, answered quiz, approved specification, plan, tracker, materialization receipt, native ticket references, work invocation and delivery evidence. Existing owners retain those artifacts; the guided run links them.

## Guided behavior

### Start and name

Starting the workflow establishes its project, namespace, initiating actor and automatic progression intent. Show the selected project, target and effective execution/publication policy in understandable terms. Check available authoring and execution prerequisites; report a known missing provider, host or project capability without claiming that work has started.

Ask for a use-case name when absent. Resolve it through the canonical use-case owner and retain a stable identity separately from the display name. A naming collision must not silently reuse or replace another case. Ask for project or case selection only when the current context is genuinely ambiguous.

The start action authorizes automatic progression within this use case and the effective project policy and limits. Once exact native scope is available, the workflow binds that intent to the previewed snapshot through the native admission owner. It does not broaden the scope to other work or bypass a policy that requires additional authority.

### Dump

Prompt the user to provide everything known about the use case. Capture multiple messages or edits until the user explicitly finishes, through a labeled interface action or a natural-language equivalent such as “That is all.” Inactivity is not completion.

Keep the captured content and explicit references in the canonical authoring session. Completing an empty capture must request meaningful input rather than manufacture requirements or continue as if context had been supplied.

### Quiz

Generate a focused quiz from the committed name, dump and explicit references. Guide the user through the current question, retain answers, and explain what information is needed. Avoid asking the user to identify an internal session or remember the next command.

Advance automatically when the quiz is complete and the owner confirms sufficient input. Missing required answers remain visible. Failed or invalid generation, including an unexpectedly empty result, does not satisfy this stage. Question bindings identify the committed inputs and current answer revision; edits to the working copies do not replace those inputs.

### Specification

Generate the specification from the retained clarification inputs. Save its lifecycle approval through the canonical specification mutation owner, bound to the exact generated revision and content hash. Retain the initiating actor, guided-run identity and default-approval policy as provenance.

Default approval applies only to the specification generated for this guided run. A model statement, body heading or readiness verdict is not approval authority. Other specifications and manual generation keep their existing lifecycle rules. Unresolved requirements that prevent valid planning remain a blocker; approval does not invent their answers.

Make the complete original specification available for inspection without requiring inspection as an extra step before progression.

### Work preparation

Generate the plan from the approved specification, then use the existing native workflow materialization path to create its ordered work, tickets, contracts and tracker. Present this as one understandable preparation stage while retaining each canonical artifact and receipt.

The tracker represents ordering and progress; it does not own execution selection, ticket success or delivery. The materialized native membership and contracts define the executable scope.

Bind execution to exactly the work materialized for this run. Older work in the same use case, later tickets, and unrelated backlog items are excluded. Missing contracts or incompatible policy block admission with a useful explanation. A partial preparation failure resumes the existing operation instead of allocating another ticket set.

### Committed stage inputs

Seal the completed dump, answered quiz, approved specification and prepared plan/tickets with stage commits. Save each artifact through its canonical owner before committing it; specification approval is included in the sealed specification version. Commit completed stages rather than every chat message or individual answer.

Retain the repository identity, full commit ID and repository-relative path for each input, together with its content hash and source references. The quiz consumes the committed dump; the specification consumes the committed dump and quiz; planning consumes the committed specification. Admission retains the complete input chain and committed initial work contracts.

Downstream operations must read the recorded Git versions, not substitute the current files at those paths. Capture source material needed by the operation in the checkpoint so that an untracked or externally mutable input is not mistaken for a retained version.

Each commit includes only the stage-owned changes and preserves unrelated working-tree and index changes. Retain the referenced commits while the run and its evidence require them. An interrupted commit operation is reconciled before retrying; a stage is not sealed until its commit reference is retained.

### Later ideas

Once an input is sealed, later edits do not replace that version for the current run. During implementation, a new request or changed idea is retained separately from the admitted scope. Explain that chr is implementing the agreed version and has saved the new idea to review after completion.

Continue the admitted work without automatic replanning, additional approval prompts or pauses caused solely by the new idea or a changed working copy. Explicit pause and cancel commands retain their normal meaning, and native execution failures or policy controls remain effective.

At completion, surface the retained ideas for review. Their presence does not authorize another development run or extend the completed invocation. Incorporating them requires a separately authorized revision or run.

### Execution

Admit the prepared scope and start the existing native host automatically. The user does not need to move every ticket manually, issue a separate start command, or supervise transitions between ready tickets.

Use the configured harness, target, publication policy, timeouts, attempt limits and retained budgets. This workflow does not establish an unlimited retry loop. Pausing, reconnecting or recovering does not reset limits or create another invocation.

If native execution requires review, additional authority, capacity, a repaired prerequisite or a decision, retain that state and request the concrete action needed. Default specification approval does not grant publication authority or waive required integration review.

### Results

A successful terminal result requires the canonical native completion evidence for all admitted work, including the required closing and integration receipts. A worker exit, passing validation, an open pull request or completed tracker checkboxes alone does not establish completion.

Present what changed, the repository and target where the code is available, relevant branch or pull-request links, validation results, and confirmed integration status. A partial or failed run identifies remaining work and the next useful action. A run waiting for review remains waiting.

Include any retained later ideas as pending review, clearly separated from the delivered scope. Preserve them if the run fails or is cancelled.

## Interface contract

| Interface | Guided interaction |
| --- | --- |
| Web | One resumable journey showing the current stage, required input, retained artifacts, execution progress and results. Reloading the page restores the same run. |
| Telegram | A natural-language entry starts the journey; subsequent messages and buttons answer the current step. The user does not need slash commands, artifact paths or opaque identifiers to complete the normal path. |
| TUI | The same stages, input requirements, progress and controls through terminal interaction, backed by the same run. |

All interfaces invoke shared application commands and queries. Adapters own rendering and transport interaction, not separate progression logic. Queries, page rendering and polling do not advance stages or start hosts.

The current context and explicit reply binding select the run whenever unambiguous. If several runs could match, ask using recognizable use-case names. Another interface can inspect or continue the same run without creating a new conversation-specific execution.

Show the use-case name, current stage, whether chr is working or waiting, and any required action. Coalesce unchanged progress. In Telegram, use concise text and the existing original-document attachment behavior for long artifacts; internal metadata and full documents do not become routine chat output.

## State, recovery and controls

The [model companion](guided-use-case-workflow-model.md) defines the retained state and ownership boundaries.

1. Persist stage progress, source observations and pending operation identities before crossing a side-effect boundary. Reconcile the original owner's result or receipt after an interrupted call or lost response before retrying.
2. Repeated messages, duplicate callbacks, browser reloads and process restarts cannot create duplicate artifacts, approvals, ticket sets, invocations or publications.
3. Mutations use the current run revision and, where applicable, the exact question and input binding. Concurrent or stale actions return the current state without overwriting newer input.
4. A disconnected client does not cancel the authorized run. Progress continues under the existing host lifecycle, and another interface can recover it.
5. Pause stops further stage dispatch and delegates already admitted work to its native control owner. Resume reconciles pending effects and continues within retained scope and limits. Cancel preserves artifacts, evidence and already delivered changes; it does not undo publication.
6. Consume the committed source references for each closed stage. Preserve later edits and ideas separately; their existence does not invalidate those references or redirect an admitted invocation. An unavailable or invalid committed input is a real recovery blocker.
7. A blocker records its cause, affected stage and concrete next action. Recovery continues the same run where its original contracts still apply. It does not silently start a replacement run or claim unobserved success.

## Existing boundaries to reuse

| Evidence | Constraint on the new behavior |
| --- | --- |
| [Authoring sessions](../../../../docs/reference/authoring-sessions.md) | Retain source observations, question bindings and artifact provenance through the existing authoring owner. |
| [Shared plan generation](../../../../internal/app/service_plan.go) and [readiness](../../../../internal/app/service_plan_readiness.go) | Keep standalone plan generation authoring-only; saved specification lifecycle remains authoritative. |
| [Native plan materialization](../../../../docs/reference/plan-workflow-materialization.md) | Reuse revision-bound materialization, stable ticket identities, journals and receipts. |
| [Workflow trackers](../../../../docs/reference/workflow-trackers.md) | Keep the tracker a projection rather than an execution authority. |
| [Workflow controls](../../../../docs/reference/workflow-controls.md) and [configuration](../../../../docs/reference/workflow-configuration.md) | Preserve native admission, controls, budgets, review and publication ownership. |

The shared application service coordinates these operations; domain and persistence owners enforce their invariants. The workspace index may project run state but is rebuildable and cannot be its authority. This specification does not prescribe a new general-purpose workflow engine, database, or duplicate execution subsystem.

## Acceptance criteria

1. **Canonical journey:** given an eligible project, starting the guided flow requests a missing name, captures a nonempty dump and conducts the quiz in that order. The user never needs to know the next internal operation.
2. **Automatic continuation:** after sufficient quiz answers, chr generates and saves the specification as approved, prepares the plan, tracker and native tickets, then starts their execution without routine intermediate confirmations.
3. **Approval boundary:** approval is bound to the generated revision and retained policy provenance. Standalone and pre-existing specifications are not automatically approved.
4. **Exact scope:** execution includes exactly the work materialized for this run. Other tickets in the project or use case are not admitted by association alone.
5. **Preserved manual use:** independent dump, quiz, specification, plan and ticket operations retain their existing semantics, including authoring without execution.
6. **Shared interaction:** Web, Telegram and TUI can inspect or continue the same run, with equivalent stages and controls. A stale action cannot alter another run or supersede a newer answer.
7. **Useful presentation:** progress identifies the use case and required action without opaque identifiers. Unchanged state does not produce repeated notifications; complete documents remain inspectable.
8. **Input failure:** an empty dump, unresolved required quiz answer or failed model response cannot be reported as successful clarification or produce invented completion evidence.
9. **Interrupted preparation:** failure after some artifacts or tickets are written, followed by recovery or a duplicate request, preserves their identities and finishes or reports the existing operation without duplication.
10. **Lost acknowledgement:** after native admission or publication succeeds but its response is lost, recovery reconciles the original invocation or receipt rather than submitting replacement work.
11. **Connection independence:** closing a browser, switching interfaces or restarting the process restores the retained stage and pending operation without resetting work or limits.
12. **Policy and controls:** review requirements, exhausted budgets and missing host/publication prerequisites produce truthful waiting or blocked states. Pause, resume and cancel preserve native control semantics and completed evidence.
13. **Committed inputs:** each completed authoring/preparation stage retains its commit and artifact paths. Downstream operations read those versions even when working copies change. Approval and admission remain bound to the recorded content.
14. **Evidence-based result:** the workflow reports completion only with the native evidence required for every admitted unit. The result tells the user where the code is and distinguishes validated, published, integrated and remaining work.
15. **Later ideas:** a new idea received during implementation is retained and acknowledged for review after completion. It does not pause execution, regenerate artifacts or expand the admitted work. The result preserves that distinction and does not start another run automatically.
16. **Interrupted checkpoint:** if committing a completed stage succeeds but its acknowledgement is lost, recovery finds the original commit and retains its reference without duplicating the stage or including unrelated changes.

## Resolved product decisions

1. Start with new use cases and resumption of runs created by this workflow. Adoption of an existing manually authored, partially completed case is useful future scope.
2. Stage commits fix the inputs used by the current run. Later changes remain available for a subsequent review; they do not introduce an automatic stop-and-regenerate cycle.

The specification and companion model are approved as the contract for delivery planning. Implementation starts after approval of the slice map and tracker.
