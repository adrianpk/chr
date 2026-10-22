# Guided use-case workflow model

Status: approved
Approved: maintainer, 2026-10-22; specification and delivery planning scope accepted.
Specification: [Guided use-case workflow](guided-use-case-workflow.md)

## Ownership

A guided run coordinates one canonical use-case journey. It owns the retained progression, input requests, automatic progression intent and references needed to recover. It does not own duplicate copies of specifications, tickets, native task states or delivery receipts.

The shared application service coordinates commands. The existing authoring session owns captured context and question bindings; specification, plan, materialization and native workflow owners retain their respective artifacts and invariants. Adapters render the same state. The workspace index is a rebuildable read model.

One canonical authoring session identifies a guided run. Reuse that session identity as the run identity rather than creating a second conversation identity. Scope is the owning project and frozen namespace; use-case association is explicit.

## Logical retained record

The field names below define a proposed logical contract, not an existing API or encoded schema.

| Field group | Required meaning and constraint |
| --- | --- |
| `schema_version`, `run_id`, `revision` | Versioned representation, canonical session identity and positive mutation revision. |
| `project`, `namespace`, `case_slug`, `case_name` | Canonical owning scope and stable use-case association, with a recognizable display name. Name changes do not change identity. |
| `initiating_actor`, `started_at`, `updated_at`, `finished_at` | Actor and lifecycle provenance; a finished timestamp requires a terminal result. |
| `stage`, `status` | Current activity and progression condition, defined below. |
| `progression_intent` | Initiating actor's intent to continue through authoring, default specification approval, materialization and execution within the selected scope and project policy. |
| `source_observation` | Exact committed inputs consumed by the current operation, plus current question/answer revisions while clarification is open. Unrelated context and later ideas are excluded. |
| `current_input` | Input kind, current question reference where applicable, expected revision and source binding. Cleared only by a valid answer or explicit transition. |
| `artifact_refs` | Canonical artifact identities with owner revisions, content hashes and retained Git references: repository identity, full commit ID and repository-relative path. |
| `stage_checkpoints` | Completed stage commits, their source references and the operation identities needed to reconcile an interrupted commit. A mutable branch name alone is not a checkpoint. |
| `spec_approval_ref` | Approval owner result, exact specification revision/hash, actor, guided-run association and default-approval policy provenance. |
| `pending_operation` | Stable operation identity, stage, immutable input digest, owner request reference, reconciliation state and owner result/receipt reference. |
| `completed_operations` | References needed to prove completed transitions and replay their owner results without repeating effects. |
| `materialization_ref` | Canonical materialization receipt and exact native set/ticket membership produced from the selected plan revision. |
| `admission_ref`, `invocation_ref` | Native admission observation/receipt and invocation identity for exactly the materialized scope. Native owners retain the policy, budgets and execution details. |
| `blocker` | Stable reason code, human explanation, affected stage, owner reference and required input or action. |
| `control_intent` | Retained pause, resume or cancel request and its owner acknowledgement; a requested stop is not falsely shown as an acknowledged stop. |
| `result_refs` | Canonical completion, closing and integration evidence, plus queryable code and validation locations. |
| `later_ideas` | References to separately retained user requests or changed ideas, with provenance and pending-review state. They are excluded from the committed input chain and admitted membership. |

The logical uniqueness boundary is `(project, namespace, run_id)`. Artifact titles, chat message identifiers and tracker rows are not identities. Transport message bindings may locate the run but cannot replace its canonical identity.

## Persistence mapping

Extend the existing project-local authoring session store rather than introducing a separate ticket database or an interface-specific run store.

| Logical data | Physical owner and representation |
| --- | --- |
| Run identity, scope, guided progression, intent and operation references | Structured, versioned guided state associated with `.chr/authoring/sessions/<session-id>/session.md`, written through the canonical revision-guarded authoring owner. |
| Captured user input and explicit references | Existing session transcript and curated context files while capture is open; completed inputs are sealed in stage artifacts. |
| Current question and answers | Canonical authoring/quiz records with their existing session, revision and question bindings. |
| Generated content and approval | Existing canonical artifact records under the selected `ops/<namespace>/` layout and the specification mutation owner; the run retains references. |
| Plans, trackers and native ticket membership | Existing plan publication and native materialization stores, journals and receipts. |
| Closed-stage inputs | Git objects referenced by full commit ID and repository-relative path, with retained content hashes. The current working copy is not a substitute. |
| Execution state, budgets, publication and integration | Existing native invocation and receipt stores; the run retains owner references and projects their current state. |
| Later ideas | Separate retained entries through the existing project authoring owner, linked to the guided run and shown as pending review. They do not replace frozen artifacts or extend native scope. |
| Search and interface summaries | Rebuildable workspace/query projections, derived from the owners above. |

The exact guided-state encoding and schema migration within the session store are implementation design work. They must preserve existing session revisions, hashes and compatibility with independently authored sessions. No implicit migration adopts an existing manual session into automatic execution.

## Progression

`stage` is one of `name`, `dump`, `quiz`, `specification`, `preparation`, `execution`, `results`. Preparation retains separate plan-generation and materialization operation references even though the user sees one stage.

`status` is one of `waiting_input`, `running`, `blocked`, `paused`, `completed`, `cancelled`, `failed`. Stage and status are independent: for example, execution may be blocked on native review while preserving the admitted invocation.

| Current stage | Successful transition evidence | Next stage |
| --- | --- | --- |
| Name | Valid canonical case association in the selected scope. | Dump, waiting for input. |
| Dump | Explicit capture completion with meaningful retained context and its stage commit. | Quiz, generating or awaiting the current answer. |
| Quiz | Bound answers, an owner-confirmed complete clarification result and its stage commit. | Specification. |
| Specification | Canonical generated artifact, saved approval for that exact revision/hash and its stage commit. | Preparation. |
| Preparation | Committed plan and initial work contracts, completed native materialization receipt, exact selection and successful native admission/start. | Execution. |
| Execution | Native completion evidence for all admitted work and required closing/integration receipts. | Results, completed. |

Known missing input yields `waiting_input`; an unmet prerequisite or required decision yields `blocked`. Recoverable interruptions retain their pending operation and can be reconciled. `failed` represents an ended unsuccessful run, not a transient error automatically followed by a new run. A blocked run becomes failed only through the owning lifecycle's explicit terminal decision.

Pause preserves the stage and pending-operation identity and prevents new downstream dispatch. During native execution, the native owner determines whether work has actually stopped. Resume reconciles in-flight effects before continuing. Cancel records the terminal outcome after required owner acknowledgement; it preserves previously completed artifacts and delivered work.

Completed, failed and cancelled runs cannot resume automatic dispatch. Further work requires an explicit new run or the relevant owner's existing recovery contract. Diagnostic inspection remains available.

## Application operations

These are semantic contracts; implementation may reuse existing commands rather than adding matching public method names.

| Operation | Contract |
| --- | --- |
| Start | Bind the initiating actor, project, namespace and automatic progression intent; establish the canonical run and report prerequisite readiness. |
| Supply name | Associate a valid use case under the current run revision without silently adopting conflicting content. |
| Append or edit dump | Persist input through the authoring owner while capture is open. Changes after the stage is sealed do not replace its committed input. |
| Finish dump | Require meaningful content, seal the completed input with a stage commit and schedule quiz generation from that version. |
| Answer current question | Validate run revision, question identity and source binding; retain the answer and advance only when clarification is complete. |
| Retain later idea | Save the request separately with provenance, acknowledge pending review after completion, and leave current progression and admitted work unchanged. |
| Advance | Reconcile or execute the next authorized stage operation; preserve stable operation identities and native limits. Owned by the host/application lifecycle, never by rendering or polling. |
| Pause, resume, cancel | Retain intent, delegate relevant native controls, and report their actual acknowledgement. |
| Inspect or list | Return stage, status, current input/action and owner-backed artifact/execution/result views without mutation or host startup. |

## Consistency and recovery rules

1. Run mutations use compare-and-swap against the canonical revision. Accepting a stale transport callback must not overwrite a newer mutation or answer another question.
2. Serialize progression through the existing project authoring/runtime ownership and exclusion mechanisms. Multiple adapters are clients of that owner, not independent schedulers. Its ownership does not replace native worker claims or leases.
3. Reserve a stable operation identity and immutable input observation before provider calls, artifact writes, materialization or admission. Retries use that identity and the original owner's reconciliation contract.
4. A transition is durable only after its owning operation supplies retained evidence. Closing an authoring/preparation stage also requires a retained Git checkpoint. There is no assumed global transaction across model execution, filesystem artifacts, Git and native work.
5. If an effect may have succeeded, reconcile the retained request, journal, artifact or receipt before repeating it. A timeout, lost response or process death is not proof that no effect occurred.
6. Specification approval must match the consumed committed revision/hash. Materialization must match the retained plan revision and use case. Native admission must match that materialization's exact membership and effective policy.
7. Read closed-stage inputs from their recorded Git versions. Later working-copy edits and retained ideas do not invalidate that chain or require a pause. Policy changes and native controls retain their existing authority; unavailable or corrupt checkpoint content requires recovery rather than substitution with a newer file.
8. A successful native invocation cannot be inferred from model output, presentation state, tracker rows or a stale query projection. Completion is derived from the native owner's receipts.
9. Stop intent prevents new progression even if an in-flight call later returns successfully. Reconcile that result and retain evidence without starting the next stage.
10. Query projections and notification delivery cannot drive transitions. Coalescing unchanged notifications or losing a client connection must not change the run's authority or progress.
11. Stage commits include only stage-owned changes and preserve unrelated working-tree and index state. Reconcile successful but unacknowledged commits before retrying. Keep referenced commits available for recovery and evidence retention.
12. Later ideas cannot be included in an in-flight generation input, native ticket set or invocation by association alone. Present them for review at completion; executing them requires separately authorized work.

## Product decisions and implementation details

1. Initial scope supports new use cases and resumption of guided runs. Adoption of manually started cases is deferred.
2. Commit-based checkpoints fix the current run's inputs. New ideas remain pending review after completion rather than triggering an automatic stop or regeneration.

The guided-state encoding, schema upgrade details, scoped commit mechanism and concrete command signatures must be designed against the existing owners before implementation. They do not change the behavior or authority boundaries specified here.
