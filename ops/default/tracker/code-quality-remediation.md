# Code Quality Remediation Tracker

Status: planned
Date: 2026-10-26
Delivery set: code-quality-remediation
Approved: maintainer, 2026-10-26; exact 16-slice map, specification/model, AST checker and lint changes approved.
Implementation authority: maintainer, 2026-10-26; execute all 16 recorded slices through one Looper run.
Merge authority: maintainer, 2026-10-26; Looper may merge the 16 recorded slice PRs into dev with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 1, Inventory and enforceable quality policy; approved and ready; execution has not started.
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`

Plan: [Single-cycle plan](../plan/code-quality-remediation.md)
Spec: [Remediation requirements](../spec/code-quality-remediation.md)
Model: [Compatibility model](../spec/code-quality-remediation-model.md)
Inventory: [Baseline rows and owners](../plan/code-quality-remediation-inventory.csv)
Source: [Quality review](../report/code-quality-review-2026-10-26/report.md)

## Authority and state

The maintainer approved implementation, conventional commit/push and startup of one Looper run on 2026-10-26. Merge delegation covers the exact 16 slice PRs into dev after their required checks. No slice has been delivered yet. This tracker owns task/commit/validation evidence; slice tickets mirror verified delivery state. Existing completed/blocked sets and unrelated tickets are not modified.

## Canonical slice map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Inventory and enforceable quality policy | refactor/slice-1-quality-policy | refactor(slice-1): establish bounded quality guardrails | ops/default/report/slices/code-quality-remediation/slice-1-policy.md |
| Slice 2 | Shared boundary values and residual core helpers | refactor/slice-2-quality-values | refactor(slice-2): name shared results and operation inputs | ops/default/report/slices/code-quality-remediation/slice-2-values.md |
| Slice 3 | Stable records and positional SQL | refactor/slice-3-quality-storage | refactor(slice-3): preserve durable records through typed row bindings | ops/default/report/slices/code-quality-remediation/slice-3-storage.md |
| Slice 4 | Typed commands and central application operations | refactor/slice-4-quality-commands | refactor(slice-4): inject command dependencies and share ticket operations | ops/default/report/slices/code-quality-remediation/slice-4-commands.md |
| Slice 5 | TUI model ownership and Web view sections | refactor/slice-5-quality-views | refactor(slice-5): reduce TUI copying and separate view state | ops/default/report/slices/code-quality-remediation/slice-5-views.md |
| Slice 6 | Gate, notifications and project packages | refactor/slice-6-quality-leaves | refactor(slice-6): extract cohesive leaf capabilities | ops/default/report/slices/code-quality-remediation/slice-6-leaves.md |
| Slice 7 | Telegram transport, receiver and retained delivery | refactor/slice-7-quality-telegram | refactor(slice-7): isolate Telegram ownership and request values | ops/default/report/slices/code-quality-remediation/slice-7-telegram.md |
| Slice 8 | Quiz domain and model adapter | refactor/slice-8-quality-quiz | refactor(slice-8): separate quiz logic from the OpenAI adapter | ops/default/report/slices/code-quality-remediation/slice-8-quiz.md |
| Slice 9 | Authoring, document generation and guided preparation | refactor/slice-9-quality-authoring | refactor(slice-9): share authoring rules without merging distinct contracts | ops/default/report/slices/code-quality-remediation/slice-9-authoring.md |
| Slice 10 | Harness contracts and runtime lifecycle | refactor/slice-10-quality-harness | refactor(slice-10): separate harness identity from host workflow policy | ops/default/report/slices/code-quality-remediation/slice-10-harness.md |
| Slice 11 | Ticket, readiness and execution boundaries | refactor/slice-11-quality-tickets | refactor(slice-11): isolate ticket and readiness responsibilities | ops/default/report/slices/code-quality-remediation/slice-11-tickets.md |
| Slice 12 | Workflow state, store and coordinator package | refactor/slice-12-quality-workflow | refactor(slice-12): establish workflow state and persistence boundaries | ops/default/report/slices/code-quality-remediation/slice-12-workflow.md |
| Slice 13 | Candidate validation and merge stages | refactor/slice-13-quality-integration | refactor(slice-13): extract candidate validation and merge policy | ops/default/report/slices/code-quality-remediation/slice-13-integration.md |
| Slice 14 | Forge publication and review stages | refactor/slice-14-quality-publication | refactor(slice-14): isolate forge publication and review decisions | ops/default/report/slices/code-quality-remediation/slice-14-publication.md |
| Slice 15 | Workflow closeout and coordinator cleanup | refactor/slice-15-quality-closeout | refactor(slice-15): complete bounded workflow stages and remove bridges | ops/default/report/slices/code-quality-remediation/slice-15-closeout.md |
| Slice 16 | Final enforcement, documentation and integrated closure | refactor/slice-16-quality-closure | refactor(slice-16): enforce final contracts and close the remediation set | ops/default/report/slices/code-quality-remediation/slice-16-closure.md |

## Task evidence

### Slice 1: Inventory and enforceable quality policy

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T1.1 | Reconcile every row of inventories A, B and C with its actual declaration and callers. Inventory all 14 large files, 11 duplicate pairs, seven dead symbols, eight error names, 23 unexplained errcheck suppressions, control flags and compatibility paths. Record false positives with evidence; do not equate metrics with defects. | pending | — | not run |
| T1.2 | Propose and implement, only after tool approval, one small standard-library Go AST check alongside the existing signature-style check. Inspect production declarations, grouped and named parameters/results, function literals and interfaces without counting receivers as parameters. Require explicit outcomes for named aliases, variadics, build constraints and generated files; do not silently skip them. | pending | — | not run |
| T1.3 | Capture an exact temporary debt manifest keyed by declaration identity and offending signature. Reject new or worsened violations, including moved or renamed declarations that do not retain explicit identity. Keep existing defect-oriented lint checks unfiltered; never apply a global new-from-merge-base filter. | pending | — | not run |
| T1.4 | Use a maximum of seven parameters initially, then six after Slice 11 and five at closure. Existing violations above the active ceiling may appear only as explicit, shrinking debt. Track all nonconforming result shapes and a separate exact external-interface exception list; do not exempt locally owned interfaces. | pending | — | not run |
| T1.5 | Exercise the checker with compact table-driven fixtures, including same-branch execution on dev, unchanged debt, new declarations, named results, aliases and stale exemptions. Establish the package dependency and naming ledger used by later moves; resolve shared-value ownership before a move can form an import cycle. | pending | — | not run |
| T1.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 2: Shared boundary values and residual core helpers

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T2.1 | For root-level helpers outside the later capability slices, replace unrelated positional outputs with named results and repeated cohesive arguments with operation inputs. Review the config, context, migration, workspace, report and proving detections assigned by the inventory. | pending | — | not run |
| T2.2 | Preserve absence, applied, changed, cached and partial outcomes separately. Use a named result with a found or changed field when absence is normal; use errors.Is-compatible errors where absence is already an error. Preserve all caller branches and error wrapping. | pending | — | not run |
| T2.3 | Resolve the real shared values, path/hash primitives and boundary ports needed by package extraction. Reuse existing owners; introduce a focused owner only when the actual dependency graph requires it. No generic common, util or options package, and no stored context.Context. | pending | — | not run |
| T2.4 | Review oversized proving.go and shared parsers/schema files. Split distinct responsibilities, preserve cohesive parsers or schema literals with a recorded reason, and simplify justified control-flag helpers. Assign every remaining root production file a final owner in the dependency ledger. | pending | — | not run |
| T2.5 | Fix owned dead/test-only helpers, error names and unexplained suppressions encountered here. Preserve meaningful cleanup errors; suppress only an explicitly justified, owned best-effort close. | pending | — | not run |
| T2.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 3: Stable records and positional SQL

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T3.1 | Freeze compatibility fixtures for WorkflowInvocation, WorkflowMergerInput, IndexedTicket and WorkflowWorkStatus before grouping fields. Record all field names, tags, omit rules, timestamps, state values and canonical hashing paths in the model companion. | pending | — | not run |
| T3.2 | Group cohesive identity, Git, timing and outcome fields while preserving existing encoded bytes wherever bytes feed hashes or retained evidence. Keep a dedicated flat serialization shape when embedding changes order or omission. Update construction, scanning and mapping code together. | pending | — | not run |
| T3.3 | Inventory every production SQL bind or scan above 16 fields, not just the three examples. Introduce one row mapping and one explicit ordered column definition per affected table, with nearby values and scan targets. Keep SQL identifiers static and user data bound. | pending | — | not run |
| T3.4 | Start with execution_repository.go, remote_command_record_repository.go and telegram_document_outbox.go. Share only small statement-construction code that is actually identical; retain table-specific updates, filters, conditional writes and transaction ownership. | pending | — | not run |
| T3.5 | Use real SQLite round trips with distinct same-type column values, null/empty distinctions, inserts, updates, omitted/default values and scan-order checks. Reopen old fixtures. Prove rollback and durable owner records remain unchanged; no schema migration is needed merely to rearrange Go fields. | pending | — | not run |
| T3.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 4: Typed commands and central application operations

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T4.1 | Replace root Metadata dependency lookup with one typed command constructor value and method actions. Cover every command constructor, not just the reported 15; remove ignored dependency and output parameters and the service locator only after all production callers move. | pending | — | not run |
| T4.2 | Mechanically split ticket.go and author.go by actual subcommand responsibility. Split maintenance.go and service_work.go where their independent operations justify it. Keep moves, signature changes and semantic extractions in distinct task commits. | pending | — | not run |
| T4.3 | Move reusable ticket loop, review, merge and start-preflight orchestration behind the existing central service in service_ticket files. Use one cohesive ticket-loop value while passing context explicitly. Route existing CLI, TUI, Web and Telegram callers through the same operation where semantics match. | pending | — | not run |
| T4.4 | Remove service.dumpDirectory and the other app-owned test-only symbols after checking production call paths; redirect retained tests to live operations. Resolve newTelegramServeCommand, serviceFromCommand and owned error-name/unparam findings without deleting meaningful regression coverage. | pending | — | not run |
| T4.5 | Finish every signature row assigned to this slice, including command entrypoints, gate/report adapters and ticket-loop results. Preserve CLI flags, output, cancellation, finite attempts and host-owned publication authority. | pending | — | not run |
| T4.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 5: TUI model ownership and Web view sections

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T5.1 | Measure the existing Model Update and View paths with a small benchmark before changing receivers. Make *Model the Bubble Tea model and group navigation, interaction, layout, background work and per-view state by owner. | pending | — | not run |
| T5.2 | Convert internal update helpers into pointer-receiver methods returning tea.Cmd or a named outcome. Retain only externally imposed interface signatures, initially Model.Update. Inspect captured commands, message ordering, cancellation, snapshots and test assumptions for aliasing or concurrent mutation. | pending | — | not run |
| T5.3 | Split model.go and shell.go by view responsibility; express action dispatch and key hints as data where this is clearer. Resolve duplicate pairs D01 and D11 and the TUI plan-action adapter switch while preserving focus, empty states and mode-specific actions. | pending | — | not run |
| T5.4 | Split GuidedUseCaseView by existing page sections and WorkflowWorkStatus by cohesive projection components. Keep application service results adapter-neutral, preserve public JSON via the model contract, and keep the current Web forms and visual language. | pending | — | not run |
| T5.5 | Review webui.go and every assigned TUI/Web signature detection. Compare the measured Update/View behavior after the change and run a focused race selection for changed asynchronous ownership; do not replace all large values with pointers. | pending | — | not run |
| T5.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 6: Gate, notifications and project packages

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T6.1 | Move gate behavior into internal/core/gate with its narrow inputs and results. Resolve its shared-error and path dependencies before moving; update direct callers in the same compiling change. | pending | — | not run |
| T6.2 | Move notification state, delivery ownership and persistence into internal/core/notify. Keep notification idempotency, event ordering, backoff and acknowledgement semantics; do not fold it into Telegram receiver storage. | pending | — | not run |
| T6.3 | Move project identity, registry and configuration behavior into internal/core/project as justified by the dependency ledger. Retain neutral shared primitives at their actual lowest owner, not copied package helpers. | pending | — | not run |
| T6.4 | For each package, move files with git mv in one commit, then apply precise package-qualified names and typed inputs/results in subsequent commits. Remove old entrypoints, re-exports and duplicate ownership after callers move. | pending | — | not run |
| T6.5 | Finish assigned signature, state-constant, error-name, control-flag and suppression work. Add or update required package READMEs and the architecture glossary with code-backed meanings. | pending | — | not run |
| T6.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 7: Telegram transport, receiver and retained delivery

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T7.1 | Move the core Telegram receiver, transport and retained operation/outbox behavior into internal/core/telegram. Introduce a nested outbox owner only if storage and transport separation actually need it; preserve one transactional owner for each existing operation. | pending | — | not run |
| T7.2 | Replace SendDocument, SendAction, presentation senders and notice insert argument lists with named request/row values. Convert all assigned result shapes while preserving missing, stale, duplicate and acknowledgement outcomes. | pending | — | not run |
| T7.3 | Apply the row mappings from Slice 3 to notice, document, action and response queues. Preserve exact original-document attachment bytes, captions, fallback behavior, finite quotas and restart reconciliation. | pending | — | not run |
| T7.4 | Resolve duplicate dispatch pair D06 through a narrow helper with separate handler tables. Verify that removing telegramGuidedResultExport and telegramHelpReply does not remove their live replacements or useful behavioral tests. | pending | — | not run |
| T7.5 | Consolidate error type names, typed receiver/operation/outbox states, unused parameters and justified close handling. Update app adapters directly; do not re-export the moved package through core. | pending | — | not run |
| T7.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 8: Quiz domain and model adapter

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T8.1 | Move quiz state, transitions and domain payloads to internal/core/quiz; move OpenAI-specific request/schema/response behavior to internal/core/quiz/openai. Keep provider-specific IO out of the quiz domain. | pending | — | not run |
| T8.2 | Split quiz_openai_model.go by schema, request and response responsibilities; split quiz_model_payload.go and service_quiz.go where cohesion warrants it. Retain the justified schema-literal complexity exception rather than creating artificial fragments. | pending | — | not run |
| T8.3 | Replace assigned quiz model argument/result tuples with domain inputs and outputs, preserving question identity, answer sufficiency, model errors, token accounting and finite round limits. | pending | — | not run |
| T8.4 | Remove service.newQuizModelRequest if its production path is still unused; update tests to the live request path. Preserve Deep/Fast/manual/guided behavior and existing provider selection. | pending | — | not run |
| T8.5 | Keep cancellation explicit and secrets out of payloads/logs. Finish local signature, flag, state and suppression entries and update the dependency ledger and glossary. | pending | — | not run |
| T8.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 9: Authoring, document generation and guided preparation

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T9.1 | Move authoring session and dump/spec/plan artifact behavior into owned internal/core/authoring boundaries; move plan/catalog behavior into its own capability only where the dependency ledger demonstrates a distinct responsibility. Keep guided inputs and committed-source provenance intact. | pending | — | not run |
| T9.2 | Read the governing spec/plan for both sides of D02, D04, D05, D07, D09 and D10. Extract catalog, generation, guided-preflight, authoring-step and dump-update logic only where both copies implement the same rule; record intentional semantic differences with precise evidence. | pending | — | not run |
| T9.3 | Share the actually common context/model/validation/recording steps of spec, acceptance and plan generation. Keep artifact-specific schemas, approval, source observation, publication and idempotency decisions explicit; do not introduce a generic generation framework. | pending | — | not run |
| T9.4 | Split service_plan.go and authoring helpers by responsibility, finish assigned context/dump/spec/plan/guided result and parameter shapes, and handle the acceptance_feature.go large-file disposition. | pending | — | not run |
| T9.5 | Use ports to the existing workflow owner for guided preparation instead of a second coordinator. Preserve manual authoring, legacy specification formats, Git checkpoints, exact hashes and retry receipts; update existing docs in their native format. | pending | — | not run |
| T9.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 10: Harness contracts and runtime lifecycle

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T10.1 | Use the existing internal/harness packages for adapter-neutral runner contracts and provider/process adapters. Move core-owned admission/binding policy to internal/core/harness when required; do not make the runner library import project workflow policy. | pending | — | not run |
| T10.2 | Convert runtimeArtifactPaths and runtime command/path/result tuples into existing or new precise values. Refactor supervised-process repeated inputs and harness/process exchange signatures with explicit resource ownership. | pending | — | not run |
| T10.3 | Separate executable identity, model/profile selection, frozen capabilities and host-owned execution policy. Preserve serialized binding hashes, finite budgets, output retention and provider-independent error behavior. | pending | — | not run |
| T10.4 | Derive contexts at the caller or use a narrowly owned cancellation handle; do not store operation contexts or lose cleanup on early errors. Review relevant boolean flags and test-only wrappers with real call-path evidence. | pending | — | not run |
| T10.5 | Move and rename in separate commits, finish assigned signatures, update app adapters and package documentation, and remove cross-package compatibility aliases before closing the slice. | pending | — | not run |
| T10.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 11: Ticket, readiness and execution boundaries

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T11.1 | Move ticket models, indexing and repository-owned ticket rules into internal/core/ticket. Isolate execution/readiness logic in a focused existing or necessary capability owner rather than putting unrelated execution orchestration in the ticket package. | pending | — | not run |
| T11.2 | Remove duplicate readiness pair D03 before splitting work_readiness_service.go. Preserve the distinction between user checkout, ticket worktree, preview, authoritative readiness and explicit safe repair. | pending | — | not run |
| T11.3 | Finish assigned execution/work/evidence/remote-command/review helper signatures and operation arguments. Move timePtr to a test helper and remove unused production-only wrappers without losing live-path coverage. | pending | — | not run |
| T11.4 | Use typed states per ticket/execution/review machine and a typed workflow preparation mode. Preserve wire spellings, database values, absence handling and existing transition guards. | pending | — | not run |
| T11.5 | Reconcile each owned five-parameter watch item, control flag and legacy path. Advance the global parameter debt ceiling from seven to six only when above-six debt has been fixed or explicitly assigned to the remaining workflow slices; record exact counts. | pending | — | not run |
| T11.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 12: Workflow state, store and coordinator package

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T12.1 | Freeze an explicit map of all current WorkflowStore methods: durable read/mutation/transaction, coordinator, integration, merge, forge/review or closeout. Retain one persistence/locking authority and identify every transaction and recovery boundary. | pending | — | not run |
| T12.2 | Move the workflow family and its actual callers to internal/core/workflow as a compiling mechanical relocation. Put shared workflow values under internal/core/workflow/state only where stages must share them without importing their coordinator. Never split methods of a nonlocal Go type across packages. | pending | — | not run |
| T12.3 | Separate storage and mutation from workflow policy. Define consumer-owned, cohesive stage ports implemented by the store/coordinator; move logic to stage-owned methods/functions in subsequent slices. Keep lock acquisition, journaling and atomic multi-record writes at their current owner. | pending | — | not run |
| T12.4 | Replace the assigned store/coordinator tuple results and repeated inputs with named results and typed modes/states. Deduplicate receipt writing D08 only after confirming operation receipt semantics match. | pending | — | not run |
| T12.5 | Switch app, authoring, ticket and adapters to the new workflow boundary. Any temporary compile-time type bridge must be enumerated with an owning removal task by Slice 15; no new exported internals solely to bypass package privacy. | pending | — | not run |
| T12.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 13: Candidate validation and merge stages

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T13.1 | Extract candidate preparation/validation into internal/core/workflow/integrate through the ports established in Slice 12. Preserve exact source/target observation, guard checks, validation claims and atomic acceptance. | pending | — | not run |
| T13.2 | Extract merge and repair policy into internal/core/workflow/merge. Convert executeWorkflowMergerDriver, preparation claim/result tuples and repeated claim/prepared/tool/process inputs into cohesive execution values; contexts stay explicit. | pending | — | not run |
| T13.3 | Retain store-owned transactions, claim generations, capacity/budget reservations, process identity and exact validation evidence. Stage code must not import core/workflow or obtain a mutable store through an untyped escape hatch. | pending | — | not run |
| T13.4 | Preserve stale input rejection, already-applied versus absent outcomes, supervised child cleanup, interrupted preparation and lost acknowledgement reconciliation. Finish all assigned signatures, typed modes, flags and error/suppression entries. | pending | — | not run |
| T13.5 | Move each package in its own mechanical task commit, then rename to package-qualified verbs/nouns with tests unchanged. Record concrete import and responsibility reductions rather than only file-count changes. | pending | — | not run |
| T13.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 14: Forge publication and review stages

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T14.1 | Extract provider publication, submission and pull-request interaction into internal/core/workflow/forge. Extract review decisions into internal/core/workflow/review. Keep the provider-independent workflow coordinator separate from these adapters. | pending | — | not run |
| T14.2 | Distinguish ticket-native authority from forge-native publication: native is not a global synonym for forge. Rename only symbols whose actual meaning is proven; retain established domain vocabulary and persisted spellings where it remains precise. | pending | — | not run |
| T14.3 | Preserve exact target/head binding, source identity, permission/protection checks, policy revision, explicit human approval and host-owned publication. An open PR or worker exit cannot become confirmed delivery. | pending | — | not run |
| T14.4 | Convert assigned submission/review tuples and repeated parameters into precise stage values. Preserve idempotent dispatch, uncertain outcome reconciliation, duplicate acknowledgement and transactional receipt ownership. | pending | — | not run |
| T14.5 | Update service/adapter callers and glossary, finish local state/flag/error/suppression work, and verify package dependencies and direct consumer interfaces without exposing store internals. | pending | — | not run |
| T14.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 15: Workflow closeout and coordinator cleanup

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T15.1 | Extract closing-gate and closure behavior into internal/core/workflow/closeout; leave core/workflow responsible for sequencing, dispatch and lifecycle. Preserve confirmation, closing-gate waiver, exact-candidate and terminal outcome distinctions. | pending | — | not run |
| T15.2 | Complete result/input conversions for remaining workflow lifecycle, loose queue, frontier, continuation, amendment, process and capacity paths. Remove workflowClosingInvocationFiles unused context and prepareWorkflowTask/withoutWorkflowPublicationProfile production-only wrappers where still applicable. | pending | — | not run |
| T15.3 | Drain policy methods from the persistence owner into their proper stages without moving atomic state transitions or locks to callers. Inspect the final store API method by method for cohesive persistence ownership; do not close merely because 909 methods moved to another directory. | pending | — | not run |
| T15.4 | Remove all temporary core/workflow type aliases, pass-through facades and migration-only interfaces introduced by the set. Check that core contains only the shared responsibilities recorded in the dependency ledger. | pending | — | not run |
| T15.5 | Finish the naming/compatibility ledger: define frontier, adoption, loose selection, continuation, amendment and closing from governing specs, then make only internal unambiguous renames. Resolve each legacy path and control flag as a supported contract, safe removal or evidence-backed retained case within this cycle. | pending | — | not run |
| T15.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

### Slice 16: Final enforcement, documentation and integrated closure

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T16.1 | Reconcile all 771 baseline inventory entries, all 14 large files, six wide structs, 11 duplicate pairs, seven dead symbols and every minor item against code and merged task evidence. Fix residual accepted defects in this set; metrics or migrated paths are not resolution evidence. | pending | — | not run |
| T16.2 | Set the final maximum to five parameters and two results, with second result error or bool. Enable compatible revive argument/result rules and dupl at the approved threshold 150 only after their debt is removed; keep the AST check for semantic result shapes and exact external-interface exceptions. | pending | — | not run |
| T16.3 | Remove the temporary debt manifest/ratchet and every provisional linter suppression introduced by this set. Resolve stale exception entries and check the complete production tree on dev, not only the branch diff. Retain intentional semantic duplicate and external-interface exceptions only with exact rationale and evidence. | pending | — | not run |
| T16.4 | Publish the final package map and glossary under docs/reference or docs/explanation, update moved API references and package READMEs, and record every source-report resolution with commit and validation evidence. Preserve the source review itself until real fixes are integrated. | pending | — | not run |
| T16.5 | Run the selected exact-candidate integrated gate, including real persistence/recovery, provider/process and representative Web/TUI/Telegram boundaries. Compare the TUI benchmark with baseline. Record command timings and failures without expanding to historical or exhaustive gates mechanically. | pending | — | not run |
| T16.6 | Close all 16 slice tickets, tracker rows and reports only after verified dev integration and the required final gate. Produce one cycle report with before/after counts, compatibility evidence and any justified exclusions. Do not create a second plan, a successor Looper run or a seventeenth slice for known work. | pending | — | not run |
| T16.7 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | pending | — | not run |

## Slice delivery

| Number | State | Ticket | PR | Integrated commit | Report |
| --- | --- | --- | --- | --- | --- |
| 1 | pending | TKT-20261026084821 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-1-policy.md` |
| 2 | pending | TKT-20261026084822 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-2-values.md` |
| 3 | pending | TKT-20261026084823 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-3-storage.md` |
| 4 | pending | TKT-20261026084824 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-4-commands.md` |
| 5 | pending | TKT-20261026084825 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-5-views.md` |
| 6 | pending | TKT-20261026084826 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-6-leaves.md` |
| 7 | pending | TKT-20261026084827 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-7-telegram.md` |
| 8 | pending | TKT-20261026084828 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-8-quiz.md` |
| 9 | pending | TKT-20261026084829 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-9-authoring.md` |
| 10 | pending | TKT-20261026084830 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-10-harness.md` |
| 11 | pending | TKT-20261026084831 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-11-tickets.md` |
| 12 | pending | TKT-20261026084832 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-12-workflow.md` |
| 13 | pending | TKT-20261026084833 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-13-integration.md` |
| 14 | pending | TKT-20261026084834 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-14-publication.md` |
| 15 | pending | TKT-20261026084835 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-15-closeout.md` |
| 16 | pending | TKT-20261026084836 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-16-closure.md` |

## Acceptance ledger

| Contract | Required evidence | State |
| --- | --- | --- |
| R1 | All 121 A rows fixed; 230 B rows reviewed; final production maximum five; no temporary allowances | pending |
| R2 | All 420 C rows fixed or precise proved external-interface exceptions; second-result check active | pending |
| R3 | Owned capability/stage graph, central service, narrow ports, no cycle/monolith/bridge | pending |
| R4 | 14 file dispositions tied to actual cohesive owners | pending |
| R5 | Six type/view dispositions; original wire/hash compatibility; measured TUI change | pending |
| R6 | D01-D11 have semantic comparison and fix or exact justified retention | pending |
| R7 | All seven symbols have live-path/deletion/test-helper evidence | pending |
| R8 | Typed command injection throughout; service locator removed | pending |
| R9 | High-arity bind/scan inventory, owned row mapping, real distinguishable-value insert/update/rollback | pending |
| R10 | Error/dispatch/suppression/unused/state/mode fixes; flag/legacy/performance dispositions | pending |
| Recovery | Original operation identity, authority, budgets, locks, receipts and terminal semantics preserved | pending |
| Integrated gate | Plan-selected commands passed on exact final dev candidate; timings and environment boundary retained | pending |
| Cycle closure | 16 verified merges, all tasks/tickets/reports consistent, cycle report complete | pending |

## Baseline metrics and migration checkpoints

| Signal | Source baseline | After Slice 1 | After Slice 11 | Final |
| --- | --- | --- | --- | --- |
| Above five parameters | 121 | pending | pending | zero unapproved |
| Exactly five parameters | 230 watch items | pending | pending | allowed; cohesion reviewed |
| Above seven parameters | 19 | pending | pending | zero unapproved |
| Outside result policy | 420, including imposed signatures | pending | pending | zero unapproved |
| WorkflowStore methods | 909 | pending | pending | cohesive durable operations only; not a numerical target |
| Temporary debt entries / bridges | not introduced yet | pending | pending | zero |
| TUI Update/View benchmark | unmeasured in this planning task | pending | pending | measured comparison |

Do not treat these baseline counts as fresh gate output. Each exception needs its declaration, actual external interface/contract, reason and retained validation. Explicitly reassociate moved declarations; no allowance grows automatically after a rename.

## Prerequisites and decisions

1. Source review and local code match the stated baseline; the two original review files are preserved byte for byte.
2. The maintainer approved the specification/model, exact 16-slice map and local AST/lint work on 2026-10-26.
3. Implementation, conditional Rebase + Fast-forward merge delegation for the 16 slice PRs, conventional commit/push and startup of one Looper run were authorized on 2026-10-26. Controller progress belongs to its retained checkpoint and this tracker.
4. The plan pins three existing full-gate entrypoints before controller configuration. Slice 1 records timing and the focused package selections; later moves update slice evidence while preserving the accepted contracts.
5. The plan contains the package dependency, temporary-bridge removal and domain-name rules. Any ambiguity is resolved from the existing contract or surfaced as a decision; it is not silently renamed.

## Closure evidence

Final candidate: pending.
Final integrated gate: not run.
Cycle report: planned at `ops/default/report/cycles/code-quality-remediation.md`; not created before implementation.
