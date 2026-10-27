# Code Quality Remediation Delivery

Status: in_progress
Date: 2026-10-26
Delivery set: code-quality-remediation
Planning authorization: maintainer request, 2026-10-26.
Approved: maintainer, 2026-10-26; exact 16-slice map, specification/model, AST checker and lint changes approved.
Implementation authority: maintainer, 2026-10-26; execute all 16 recorded slices through one Looper run.
Merge authority: maintainer, 2026-10-26; Looper may merge the 16 recorded slice PRs into dev with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 8, Quiz domain and model adapter; in progress after verified integration of PR #444.
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b` on dev.

Spec: [Remediation requirements](../spec/code-quality-remediation.md)
Model: [Compatibility model](../spec/code-quality-remediation-model.md)
Tracker: [Tasks and delivery evidence](../tracker/code-quality-remediation.md)
Inventory: [All baseline declarations and owners](code-quality-remediation-inventory.csv)
Review: [Original review](../report/code-quality-review-2026-10-26/report.md)
Detections: [Original complete lists](../report/code-quality-review-2026-10-26/detections.md)

## Outcome and strategy

Resolve all accepted findings in one plan, one tracker, 16 linked slice tickets and one future Looper configuration/run. The package split is part of this same delivery set. There is no separate architecture cycle, chained successor configuration or generic future cleanup phase.

Slice strategy: layered.

Reason: this is structural work across an existing application. Contract/compatibility protection comes first, followed by cohesive state and capability ownership, then workflow stage extraction and final enforcement. Each slice contains its compiling implementation and appropriate tests; there are no intentionally failing test or empty skeleton slices. Closely related package moves may share one slice to stay within 16, but each package is moved in a separate reviewable task commit.

The maintainer approved the complete proposal and requested its conventional commit, push and single-cycle Looper startup on 2026-10-26. This includes the local AST checker and lint changes. The source report remains unchanged until actual remediation evidence is recorded. The 16 slice identities and scope remain fixed.

## Evidence and corrections to the report

1. The baseline matches HEAD. Inventory A contains 121 signatures, B contains 230 legal five-parameter watch items, and C contains 420 result-shape entries. The inventory links every entry to an owning slice; those categories overlap and must not be summed as unique functions.
2. Five parameters is the final allowed maximum. B is a cohesion review list, not 230 automatic violations. Locally owned interfaces must change with their implementations; external interface obligations require exact evidence-backed exceptions.
3. A global issues.new-from-merge-base filter would hide unrelated existing lint failures and becomes especially unsafe when dev is compared with itself. Use the exact temporary debt manifest and keep normal lint output unfiltered. Final enforcement scans the entire production tree.
4. Embedding struct fields may preserve key names but alter JSON order, omission or collisions. Preserve exact bytes on hash-sensitive paths; retain explicit flat serialization where needed.
5. WorkflowStore is the filesystem mutation/locking owner, not a SQL repository. It cannot be decomposed by defining methods of the same Go type in multiple packages. Separate durable operations from stage policy through narrow ports and validate transaction/recovery ownership.
6. Native ticket authority is not a synonym for forge publication; ticket-native-workflow.md distinguishes native records, loose tickets, adoption, continuation, amendments and closing. This set does not approve a blanket vocabulary replacement.
7. scripts/check-go-func-signatures.sh checks added signature line layout, not arity or result semantics. Reuse its make entrypoint integration, but it does not already satisfy R1/R2.
8. The review lists seven representative large files; the scan finds 14. Every one receives a split or an explicit cohesion-based retain decision below. Line count is not itself a permanent lint limit.

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

## Work and acceptance by slice

### Slice 1: Inventory and enforceable quality policy

Findings: R1, R2, R3, R6, R10.
Ticket: [TKT-20261026084821](../ticket/solved/20261026084821-quality-policy.md).

1. **T1.1:** Reconcile every row of inventories A, B and C with its actual declaration and callers. Inventory all 14 large files, 11 duplicate pairs, seven dead symbols, eight error names, 23 unexplained errcheck suppressions, control flags and compatibility paths. Record false positives with evidence; do not equate metrics with defects.
2. **T1.2:** Propose and implement, only after tool approval, one small standard-library Go AST check alongside the existing signature-style check. Inspect production declarations, grouped and named parameters/results, function literals and interfaces without counting receivers as parameters. Require explicit outcomes for named aliases, variadics, build constraints and generated files; do not silently skip them.
3. **T1.3:** Capture an exact temporary debt manifest keyed by declaration identity and offending signature. Reject new or worsened violations, including moved or renamed declarations that do not retain explicit identity. Keep existing defect-oriented lint checks unfiltered; never apply a global new-from-merge-base filter.
4. **T1.4:** Use a maximum of seven parameters initially, then six after Slice 11 and five at closure. Existing violations above the active ceiling may appear only as explicit, shrinking debt. Track all nonconforming result shapes and a separate exact external-interface exception list; do not exempt locally owned interfaces.
5. **T1.5:** Exercise the checker with compact table-driven fixtures, including same-branch execution on dev, unchanged debt, new declarations, named results, aliases and stale exemptions. Establish the package dependency and naming ledger used by later moves; resolve shared-value ownership before a move can form an import cycle.
6. **T1.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: The review inventory is accounted for; the new checks detect regressions without hiding existing lint findings; no implementation approval is inferred from this document.

Focused selection: Checker fixtures; existing CI-script regressions; current production scan; make verify.

### Slice 2: Shared boundary values and residual core helpers

Findings: R1, R2, R3, R4, R10.
Ticket: [TKT-20261026084822](../ticket/solved/20261026084822-quality-values.md).

1. **T2.1:** For root-level helpers outside the later capability slices, replace unrelated positional outputs with named results and repeated cohesive arguments with operation inputs. Review the config, context, migration, workspace, report and proving detections assigned by the inventory.
2. **T2.2:** Preserve absence, applied, changed, cached and partial outcomes separately. Use a named result with a found or changed field when absence is normal; use errors.Is-compatible errors where absence is already an error. Preserve all caller branches and error wrapping.
3. **T2.3:** Resolve the real shared values, path/hash primitives and boundary ports needed by package extraction. Reuse existing owners; introduce a focused owner only when the actual dependency graph requires it. No generic common, util or options package, and no stored context.Context.
4. **T2.4:** Review oversized proving.go and shared parsers/schema files. Split distinct responsibilities, preserve cohesive parsers or schema literals with a recorded reason, and simplify justified control-flag helpers. Assign every remaining root production file a final owner in the dependency ledger.
5. **T2.5:** Fix owned dead/test-only helpers, error names and unexplained suppressions encountered here. Preserve meaningful cleanup errors; suppress only an explicitly justified, owned best-effort close.
6. **T2.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: All assigned signature violations have a compatible replacement; the extraction dependency ledger is acyclic and concrete; every remaining root file has a responsibility rather than an unassigned prefix.

Focused selection: Focused config/context/proving/workspace tests; compile affected reverse dependencies; make verify.

### Slice 3: Stable records and positional SQL

Findings: R1, R2, R5, R9, R10.
Ticket: [TKT-20261026084823](../ticket/solved/20261026084823-quality-storage.md).

1. **T3.1:** Freeze compatibility fixtures for WorkflowInvocation, WorkflowMergerInput, IndexedTicket and WorkflowWorkStatus before grouping fields. Record all field names, tags, omit rules, timestamps, state values and canonical hashing paths in the model companion.
2. **T3.2:** Group cohesive identity, Git, timing and outcome fields while preserving existing encoded bytes wherever bytes feed hashes or retained evidence. Keep a dedicated flat serialization shape when embedding changes order or omission. Update construction, scanning and mapping code together.
3. **T3.3:** Inventory every production SQL bind or scan above 16 fields, not just the three examples. Introduce one row mapping and one explicit ordered column definition per affected table, with nearby values and scan targets. Keep SQL identifiers static and user data bound.
4. **T3.4:** Start with execution_repository.go, remote_command_record_repository.go and telegram_document_outbox.go. Share only small statement-construction code that is actually identical; retain table-specific updates, filters, conditional writes and transaction ownership.
5. **T3.5:** Use real SQLite round trips with distinct same-type column values, null/empty distinctions, inserts, updates, omitted/default values and scan-order checks. Reopen old fixtures. Prove rollback and durable owner records remain unchanged; no schema migration is needed merely to rearrange Go fields.
6. **T3.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Every high-arity SQL mapping is covered by distinguishable round-trip data; old records, wire formats and hash-sensitive byte encodings remain compatible.

Focused selection: Existing execution/remote-command/document-outbox repository tests, extended real-SQLite round trips, JSON/hash compatibility fixtures; make verify.

### Slice 4: Typed commands and central application operations

Findings: R1, R2, R4, R7, R8, R10.
Ticket: [TKT-20261026084824](../ticket/solved/20261026084824-quality-commands.md).

1. **T4.1:** Replace root Metadata dependency lookup with one typed command constructor value and method actions. Cover every command constructor, not just the reported 15; remove ignored dependency and output parameters and the service locator only after all production callers move.
2. **T4.2:** Mechanically split ticket.go and author.go by actual subcommand responsibility. Split maintenance.go and service_work.go where their independent operations justify it. Keep moves, signature changes and semantic extractions in distinct task commits.
3. **T4.3:** Move reusable ticket loop, review, merge and start-preflight orchestration behind the existing central service in service_ticket files. Use one cohesive ticket-loop value while passing context explicitly. Route existing CLI, TUI, Web and Telegram callers through the same operation where semantics match.
4. **T4.4:** Remove service.dumpDirectory and the other app-owned test-only symbols after checking production call paths; redirect retained tests to live operations. Resolve newTelegramServeCommand, serviceFromCommand and owned error-name/unparam findings without deleting meaningful regression coverage.
5. **T4.5:** Finish every signature row assigned to this slice, including command entrypoints, gate/report adapters and ticket-loop results. Preserve CLI flags, output, cancellation, finite attempts and host-owned publication authority.
6. **T4.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Command wiring is compile-time typed; adapters contain parsing and presentation; ticket orchestration has one central operation owner; no metadata service locator remains.

Focused selection: CLI command-construction and ticket loop/start/review/merge regressions with real local Git where affected; make verify.

### Slice 5: TUI model ownership and Web view sections

Findings: R1, R2, R4, R5, R6, R10.
Ticket: [TKT-20261026084825](../ticket/solved/20261026084825-quality-views.md).

1. **T5.1:** Measure the existing Model Update and View paths with a small benchmark before changing receivers. Make *Model the Bubble Tea model and group navigation, interaction, layout, background work and per-view state by owner.
2. **T5.2:** Convert internal update helpers into pointer-receiver methods returning tea.Cmd or a named outcome. Retain only externally imposed interface signatures, initially Model.Update. Inspect captured commands, message ordering, cancellation, snapshots and test assumptions for aliasing or concurrent mutation.
3. **T5.3:** Split model.go and shell.go by view responsibility; express action dispatch and key hints as data where this is clearer. Resolve duplicate pairs D01 and D11 and the TUI plan-action adapter switch while preserving focus, empty states and mode-specific actions.
4. **T5.4:** Split GuidedUseCaseView by existing page sections and WorkflowWorkStatus by cohesive projection components. Keep application service results adapter-neutral, preserve public JSON via the model contract, and keep the current Web forms and visual language.
5. **T5.5:** Review webui.go and every assigned TUI/Web signature detection. Compare the measured Update/View behavior after the change and run a focused race selection for changed asynchronous ownership; do not replace all large values with pointers.
6. **T5.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: The TUI has one pointer model with explicit state owners, equivalent interaction and fewer value copies; Web sections have narrow models and unchanged operation semantics.

Focused selection: Existing TUI model/action tests; focused race on changed async paths; Update/View benchmark comparison; existing WebQuizBrowser and GuidedUseCaseWebBrowser journeys; make verify.

### Slice 6: Gate, notifications and project packages

Findings: R1, R2, R3, R10.
Ticket: [TKT-20261026084826](../ticket/solved/20261026084826-quality-leaves.md).

1. **T6.1:** Move gate behavior into internal/core/gate with its narrow inputs and results. Resolve its shared-error and path dependencies before moving; update direct callers in the same compiling change.
2. **T6.2:** Move notification state, delivery ownership and persistence into internal/core/notify. Keep notification idempotency, event ordering, backoff and acknowledgement semantics; do not fold it into Telegram receiver storage.
3. **T6.3:** Move project identity, registry and configuration behavior into internal/core/project as justified by the dependency ledger. Retain neutral shared primitives at their actual lowest owner, not copied package helpers.
4. **T6.4:** For each package, move files with git mv in one commit, then apply precise package-qualified names and typed inputs/results in subsequent commits. Remove old entrypoints, re-exports and duplicate ownership after callers move.
5. **T6.5:** Finish assigned signature, state-constant, error-name, control-flag and suppression work. Add or update required package READMEs and the architecture glossary with code-backed meanings.
6. **T6.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: The three leaf capabilities have enforced import boundaries and real callers; root core no longer owns their implementations or compatibility facades.

Focused selection: Focused gate, notification delivery/retry and project-registry regressions; package dependency check; make verify.

### Slice 7: Telegram transport, receiver and retained delivery

Findings: R1, R2, R3, R6, R7, R9, R10.
Ticket: [TKT-20261026084827](../ticket/solved/20261026084827-quality-telegram.md).

1. **T7.1:** Move the core Telegram receiver, transport and retained operation/outbox behavior into internal/core/telegram. Introduce a nested outbox owner only if storage and transport separation actually need it; preserve one transactional owner for each existing operation.
2. **T7.2:** Replace SendDocument, SendAction, presentation senders and notice insert argument lists with named request/row values. Convert all assigned result shapes while preserving missing, stale, duplicate and acknowledgement outcomes.
3. **T7.3:** Apply the row mappings from Slice 3 to notice, document, action and response queues. Preserve exact original-document attachment bytes, captions, fallback behavior, finite quotas and restart reconciliation.
4. **T7.4:** Resolve duplicate dispatch pair D06 through a narrow helper with separate handler tables. Verify that removing telegramGuidedResultExport and telegramHelpReply does not remove their live replacements or useful behavioral tests.
5. **T7.5:** Consolidate error type names, typed receiver/operation/outbox states, unused parameters and justified close handling. Update app adapters directly; do not re-export the moved package through core.
6. **T7.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Telegram code has an owned package boundary, cohesive request values and compatible retained delivery; duplicate replies or attachments are not introduced by retries or restarts.

Focused selection: Existing receiver/action/outbox and document multipart fixtures with real local HTTP/SQLite; stale and lost-ack recovery; focused race where state ownership changes; make verify.

### Slice 8: Quiz domain and model adapter

Findings: R1, R2, R3, R4, R7, R10.
Ticket: [TKT-20261026084828](../ticket/in_progress/20261026084828-quality-quiz.md).

1. **T8.1:** Move quiz state, transitions and domain payloads to internal/core/quiz; move OpenAI-specific request/schema/response behavior to internal/core/quiz/openai. Keep provider-specific IO out of the quiz domain.
2. **T8.2:** Split quiz_openai_model.go by schema, request and response responsibilities; split quiz_model_payload.go and service_quiz.go where cohesion warrants it. Retain the justified schema-literal complexity exception rather than creating artificial fragments.
3. **T8.3:** Replace assigned quiz model argument/result tuples with domain inputs and outputs, preserving question identity, answer sufficiency, model errors, token accounting and finite round limits.
4. **T8.4:** Remove service.newQuizModelRequest if its production path is still unused; update tests to the live request path. Preserve Deep/Fast/manual/guided behavior and existing provider selection.
5. **T8.5:** Keep cancellation explicit and secrets out of payloads/logs. Finish local signature, flag, state and suppression entries and update the dependency ledger and glossary.
6. **T8.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Quiz contracts are independent of the OpenAI transport; all existing clarification modes and limits retain their behavior; large provider files are cohesive.

Focused selection: Existing quiz/parser/provider request fixtures and sufficiency tests; shared service generation regressions; no live model account required; make verify.

### Slice 9: Authoring, document generation and guided preparation

Findings: R1, R2, R3, R4, R6, R10.
Ticket: [TKT-20261026084829](../ticket/ready/20261026084829-quality-authoring.md).

1. **T9.1:** Move authoring session and dump/spec/plan artifact behavior into owned internal/core/authoring boundaries; move plan/catalog behavior into its own capability only where the dependency ledger demonstrates a distinct responsibility. Keep guided inputs and committed-source provenance intact.
2. **T9.2:** Read the governing spec/plan for both sides of D02, D04, D05, D07, D09 and D10. Extract catalog, generation, guided-preflight, authoring-step and dump-update logic only where both copies implement the same rule; record intentional semantic differences with precise evidence.
3. **T9.3:** Share the actually common context/model/validation/recording steps of spec, acceptance and plan generation. Keep artifact-specific schemas, approval, source observation, publication and idempotency decisions explicit; do not introduce a generic generation framework.
4. **T9.4:** Split service_plan.go and authoring helpers by responsibility, finish assigned context/dump/spec/plan/guided result and parameter shapes, and handle the acceptance_feature.go large-file disposition.
5. **T9.5:** Use ports to the existing workflow owner for guided preparation instead of a second coordinator. Preserve manual authoring, legacy specification formats, Git checkpoints, exact hashes and retry receipts; update existing docs in their native format.
6. **T9.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Duplicate authoring rules have one owner while genuinely distinct generation contracts remain separate; manual and guided paths produce compatible artifacts and receipts.

Focused selection: Existing generation/source-revision/authoring publication recovery and guided preparation tests with real filesystem/Git; make verify.

### Slice 10: Harness contracts and runtime lifecycle

Findings: R1, R2, R3, R7, R10.
Ticket: [TKT-20261026084830](../ticket/ready/20261026084830-quality-harness.md).

1. **T10.1:** Use the existing internal/harness packages for adapter-neutral runner contracts and provider/process adapters. Move core-owned admission/binding policy to internal/core/harness when required; do not make the runner library import project workflow policy.
2. **T10.2:** Convert runtimeArtifactPaths and runtime command/path/result tuples into existing or new precise values. Refactor supervised-process repeated inputs and harness/process exchange signatures with explicit resource ownership.
3. **T10.3:** Separate executable identity, model/profile selection, frozen capabilities and host-owned execution policy. Preserve serialized binding hashes, finite budgets, output retention and provider-independent error behavior.
4. **T10.4:** Derive contexts at the caller or use a narrowly owned cancellation handle; do not store operation contexts or lose cleanup on early errors. Review relevant boolean flags and test-only wrappers with real call-path evidence.
5. **T10.5:** Move and rename in separate commits, finish assigned signatures, update app adapters and package documentation, and remove cross-package compatibility aliases before closing the slice.
6. **T10.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Runner contracts and project admission policy have clear dependency direction; runtime results are named and process cleanup/cancellation remain correct.

Focused selection: Existing harness runner/codex/pi/process contracts using owned child processes, binding hash fixtures and focused cancellation/race regressions; make verify.

### Slice 11: Ticket, readiness and execution boundaries

Findings: R1, R2, R3, R4, R6, R7, R10.
Ticket: [TKT-20261026084831](../ticket/ready/20261026084831-quality-tickets.md).

1. **T11.1:** Move ticket models, indexing and repository-owned ticket rules into internal/core/ticket. Isolate execution/readiness logic in a focused existing or necessary capability owner rather than putting unrelated execution orchestration in the ticket package.
2. **T11.2:** Remove duplicate readiness pair D03 before splitting work_readiness_service.go. Preserve the distinction between user checkout, ticket worktree, preview, authoritative readiness and explicit safe repair.
3. **T11.3:** Finish assigned execution/work/evidence/remote-command/review helper signatures and operation arguments. Move timePtr to a test helper and remove unused production-only wrappers without losing live-path coverage.
4. **T11.4:** Use typed states per ticket/execution/review machine and a typed workflow preparation mode. Preserve wire spellings, database values, absence handling and existing transition guards.
5. **T11.5:** Reconcile each owned five-parameter watch item, control flag and legacy path. Advance the global parameter debt ceiling from seven to six only when above-six debt has been fixed or explicitly assigned to the remaining workflow slices; record exact counts.
6. **T11.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Ticket/readiness ownership is enforced, all assigned signature entries are closed, and the six-parameter phase cannot hide residual high-arity workflow debt.

Focused selection: Existing indexed-ticket/readiness/doctor/worktree/evidence regressions using real SQLite and temporary Git; make verify.

### Slice 12: Workflow state, store and coordinator package

Findings: R1, R2, R3, R5, R6, R10.
Ticket: [TKT-20261026084832](../ticket/ready/20261026084832-quality-workflow.md).

1. **T12.1:** Freeze an explicit map of all current WorkflowStore methods: durable read/mutation/transaction, coordinator, integration, merge, forge/review or closeout. Retain one persistence/locking authority and identify every transaction and recovery boundary.
2. **T12.2:** Move the workflow family and its actual callers to internal/core/workflow as a compiling mechanical relocation. Put shared workflow values under internal/core/workflow/state only where stages must share them without importing their coordinator. Never split methods of a nonlocal Go type across packages.
3. **T12.3:** Separate storage and mutation from workflow policy. Define consumer-owned, cohesive stage ports implemented by the store/coordinator; move logic to stage-owned methods/functions in subsequent slices. Keep lock acquisition, journaling and atomic multi-record writes at their current owner.
4. **T12.4:** Replace the assigned store/coordinator tuple results and repeated inputs with named results and typed modes/states. Deduplicate receipt writing D08 only after confirming operation receipt semantics match.
5. **T12.5:** Switch app, authoring, ticket and adapters to the new workflow boundary. Any temporary compile-time type bridge must be enumerated with an owning removal task by Slice 15; no new exported internals solely to bypass package privacy.
6. **T12.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Workflow files have a real package owner and compatible state encodings; stage interfaces do not depend on the coordinator; one store owns transactions and recovery.

Focused selection: Existing WorkflowRoundTrip, mutation/receipt/idempotency/materialization and process recovery regressions on owned filesystem/Git; make verify.

### Slice 13: Candidate validation and merge stages

Findings: R1, R2, R3, R10.
Ticket: [TKT-20261026084833](../ticket/ready/20261026084833-quality-integration.md).

1. **T13.1:** Extract candidate preparation/validation into internal/core/workflow/integrate through the ports established in Slice 12. Preserve exact source/target observation, guard checks, validation claims and atomic acceptance.
2. **T13.2:** Extract merge and repair policy into internal/core/workflow/merge. Convert executeWorkflowMergerDriver, preparation claim/result tuples and repeated claim/prepared/tool/process inputs into cohesive execution values; contexts stay explicit.
3. **T13.3:** Retain store-owned transactions, claim generations, capacity/budget reservations, process identity and exact validation evidence. Stage code must not import core/workflow or obtain a mutable store through an untyped escape hatch.
4. **T13.4:** Preserve stale input rejection, already-applied versus absent outcomes, supervised child cleanup, interrupted preparation and lost acknowledgement reconciliation. Finish all assigned signatures, typed modes, flags and error/suppression entries.
5. **T13.5:** Move each package in its own mechanical task commit, then rename to package-qualified verbs/nouns with tests unchanged. Record concrete import and responsibility reductions rather than only file-count changes.
6. **T13.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Validation and merge logic are independently owned stages with narrow ports; the same receipts, budgets and recovery invariants still decide integration.

Focused selection: Existing preparation/admission/merger repair/cancellation and lost-ack fixtures with real Git and child processes; focused race for changed claim ownership; make verify.

### Slice 14: Forge publication and review stages

Findings: R1, R2, R3, R10.
Ticket: [TKT-20261026084834](../ticket/ready/20261026084834-quality-publication.md).

1. **T14.1:** Extract provider publication, submission and pull-request interaction into internal/core/workflow/forge. Extract review decisions into internal/core/workflow/review. Keep the provider-independent workflow coordinator separate from these adapters.
2. **T14.2:** Distinguish ticket-native authority from forge-native publication: native is not a global synonym for forge. Rename only symbols whose actual meaning is proven; retain established domain vocabulary and persisted spellings where it remains precise.
3. **T14.3:** Preserve exact target/head binding, source identity, permission/protection checks, policy revision, explicit human approval and host-owned publication. An open PR or worker exit cannot become confirmed delivery.
4. **T14.4:** Convert assigned submission/review tuples and repeated parameters into precise stage values. Preserve idempotent dispatch, uncertain outcome reconciliation, duplicate acknowledgement and transactional receipt ownership.
5. **T14.5:** Update service/adapter callers and glossary, finish local state/flag/error/suppression work, and verify package dependencies and direct consumer interfaces without exposing store internals.
6. **T14.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Forge transport and review policy are separate from orchestration, without changing approval authority or delivery proof; no blanket native-to-forge rename occurs.

Focused selection: Existing native provider/transport permission and review fixtures through local HTTP/Git; existing review/publication receipt tests; make verify.

### Slice 15: Workflow closeout and coordinator cleanup

Findings: R1, R2, R3, R7, R10.
Ticket: [TKT-20261026084835](../ticket/ready/20261026084835-quality-closeout.md).

1. **T15.1:** Extract closing-gate and closure behavior into internal/core/workflow/closeout; leave core/workflow responsible for sequencing, dispatch and lifecycle. Preserve confirmation, closing-gate waiver, exact-candidate and terminal outcome distinctions.
2. **T15.2:** Complete result/input conversions for remaining workflow lifecycle, loose queue, frontier, continuation, amendment, process and capacity paths. Remove workflowClosingInvocationFiles unused context and prepareWorkflowTask/withoutWorkflowPublicationProfile production-only wrappers where still applicable.
3. **T15.3:** Drain policy methods from the persistence owner into their proper stages without moving atomic state transitions or locks to callers. Inspect the final store API method by method for cohesive persistence ownership; do not close merely because 909 methods moved to another directory.
4. **T15.4:** Remove all temporary core/workflow type aliases, pass-through facades and migration-only interfaces introduced by the set. Check that core contains only the shared responsibilities recorded in the dependency ledger.
5. **T15.5:** Finish the naming/compatibility ledger: define frontier, adoption, loose selection, continuation, amendment and closing from governing specs, then make only internal unambiguous renames. Resolve each legacy path and control flag as a supported contract, safe removal or evidence-backed retained case within this cycle.
6. **T15.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Workflow policy is split by stage, persistence owns only durable operations, migration bridges are gone, and all assigned lifecycle and compatibility items have an evidence-backed disposition.

Focused selection: Existing continuation/closing/process/budget/stop recovery tests, one shared guided parity journey, focused race for lifecycle changes; make verify.

### Slice 16: Final enforcement, documentation and integrated closure

Findings: R1-R10.
Ticket: [TKT-20261026084836](../ticket/ready/20261026084836-quality-closure.md).

1. **T16.1:** Reconcile all 771 baseline inventory entries, all 14 large files, six wide structs, 11 duplicate pairs, seven dead symbols and every minor item against code and merged task evidence. Fix residual accepted defects in this set; metrics or migrated paths are not resolution evidence.
2. **T16.2:** Set the final maximum to five parameters and two results, with second result error or bool. Enable compatible revive argument/result rules and dupl at the approved threshold 150 only after their debt is removed; keep the AST check for semantic result shapes and exact external-interface exceptions.
3. **T16.3:** Remove the temporary debt manifest/ratchet and every provisional linter suppression introduced by this set. Resolve stale exception entries and check the complete production tree on dev, not only the branch diff. Retain intentional semantic duplicate and external-interface exceptions only with exact rationale and evidence.
4. **T16.4:** Publish the final package map and glossary under docs/reference or docs/explanation, update moved API references and package READMEs, and record every source-report resolution with commit and validation evidence. Preserve the source review itself until real fixes are integrated.
5. **T16.5:** Run the selected exact-candidate integrated gate, including real persistence/recovery, provider/process and representative Web/TUI/Telegram boundaries. Compare the TUI benchmark with baseline. Record command timings and failures without expanding to historical or exhaustive gates mechanically.
6. **T16.6:** Close all 16 slice tickets, tracker rows and reports only after verified dev integration and the required final gate. Produce one cycle report with before/after counts, compatibility evidence and any justified exclusions. Do not create a second plan, a successor Looper run or a seventeenth slice for known work.
7. **T16.7:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

Exit: Every accepted finding is resolved or specifically justified as a retained contract/false positive; no temporary debt or bridge remains; the exact integrated candidate passes the selected gate.

Focused selection: Final AST/exception/dependency checks, make verify, representative existing integration selections and benchmark comparison as defined in the plan.

## Finding ownership and completion

| Finding | Owning slices | Required final evidence |
| --- | --- | --- |
| R1 parameters | 1-16 by declaration inventory | Zero unapproved functions above five; exactly-five list reviewed; no temporary high-arity allowance |
| R2 results | 1-16 by declaration inventory | Zero unapproved result shapes; precise external-interface exception list and tests |
| R3 core packages and names | 2, 6-16 | Acyclic package map, domain glossary, no root workflow monolith, no migration facades, consumer-owned stage ports |
| R4 large files | 2, 4, 5, 8, 9, 11, 16 | Disposition of all 14 files, with ownership-based splits or concrete cohesive-parser/schema rationale |
| R5 wide types | 3, 5, 12, 13 | All six reviewed; durable/wire/hash compatibility; measured TUI copy change; narrow view sections |
| R6 duplication | 4-9, 11, 12, 16 | Every D01-D11 fixed or specifically justified from both governing contracts; no broad duplicate exclusions |
| R7 dead code | 2, 4, 7, 8, 11, 15 | Seven named production symbols removed/moved or proved live; useful tests target live code |
| R8 command wiring | 4 | Typed constructor injection for every command; Metadata locator and ignored dependency arguments gone |
| R9 SQL | 3, 7, 11 | All binds/scans above 16 fields reviewed; stable row/column mapping and distinct-value real round trips |
| R10 minor and deferred candidates | Every owner, closed in 16 | Dispatch/key hints, error suffixes, 23 suppressions, unused params, typed states/modes; flag and legacy ledgers; measured performance dispositions |

## Concrete secondary inventory

All 14 oversized files are assigned below. The owner may retain a cohesive parser/schema with a specific explanation; an arbitrary 800-line cap is not acceptance.

| Baseline file | Slice |
| --- | --- |
| internal/core/proving.go | 2 |
| internal/app/ticket.go | 4 |
| internal/app/author.go | 4 |
| internal/app/maintenance.go | 4 |
| internal/app/service_work.go | 4 |
| internal/tui/model.go | 5 |
| internal/tui/shell.go | 5 |
| internal/webui/webui.go | 5 |
| internal/core/quiz_openai_model.go | 8 |
| internal/core/quiz_model_payload.go | 8 |
| internal/app/service_quiz.go | 8 |
| internal/core/acceptance_feature.go | 9 |
| internal/app/service_plan.go | 9 |
| internal/core/work_readiness_service.go | 11 |

| Duplicate | Baseline pair | Slice |
| --- | --- | --- |
| D01 | plans.go / specs.go list views | 5 |
| D02 | plan_catalog / spec_catalog | 9 |
| D03 | readiness repeated preflight blocks | 11 |
| D04 | spec / acceptance generation | 9 |
| D05 | plan / spec generation | 9 |
| D06 | Telegram explanation / work dispatch | 7 |
| D07 | guided admission / preparation | 9 |
| D08 | workflow control receipt / mutation | 12 |
| D09 | authoring contextual quiz / promotion | 9 |
| D10 | dump update branches | 9 |
| D11 | shell ticket column branches | 5 |

R7 ownership: dumpDirectory is Slice 4; newQuizModelRequest Slice 8; telegramGuidedResultExport and telegramHelpReply Slice 7; timePtr Slice 11; prepareWorkflowTask and withoutWorkflowPublicationProfile Slice 15. Verify production reachability again before removal.

R10 ownership: TUI dispatch/key hints Slice 5; errname and errcheck explanations in each owning slice with Slice 16 completeness review; command unused parameters Slice 4; closing unused context Slice 15; typed states/modes Slices 3 and 7-15. Review all 88 reported control parameters and approximately 300 legacy references by owner. They are review candidates, not 388 proven defects. Supported compatibility remains only with an exact contract and removal condition; safe obsolete branches are removed in this cycle. The 4,374 hugeParam warnings and general rangeValCopy heuristics do not justify bulk pointer changes: measure the TUI path and fix any other demonstrated hot-path defect, or close the warning category with the explicit measurement/false-positive boundary.


## Dependency and migration rules

1. Slices execute consecutively from 1 through 16. Each starts from the verified merged dev candidate of its predecessor. The controller does not run independent slices concurrently against the same repository.
2. Capability behavior remains under internal/core; runner adapters remain under internal/harness; the existing central app service stays the operation owner for all adapters.
3. Leaf packages never depend on the workflow coordinator. Workflow stage packages depend on shared values and their narrow consumer-owned ports, not core/workflow. The store/coordinator implements those ports and retains transactions and journals.
4. A move begins with a real unexported-symbol/import dependency check. Resolve shared owners before moving. Do not create a generic shared package or export private internals to force the graph to compile.
5. Use git mv for moves and separate moves, renames and behavior-preserving extraction commits. Update real callers in the same compiling unit. Enumerate any genuinely necessary temporary bridge with its removal task, expire it by Slice 15, and reject permanent re-export facades.
6. Keep wire spelling and existing spec paths stable. Naming changes affect only meanings proven from the governing contract. Broader ambiguous terminology changes require a maintainer decision, not a guessed rename.

## Validation design and cost

Per task, use formatting/compile and the smallest existing behavior selection that protects the changed contract. Per slice, run make verify as this repository requires, plus focused real-boundary checks for affected persistence, Git, HTTP, process or concurrency behavior. Batch cases in the native runner; do not multiply all suites across providers or repeat complete browser journeys for each tuple change.

Final gate for the exact integrated candidate:

1. The approved production AST/result/exception and dependency checks, with no temporary debt; existing strict lint and package README checks remain active through make verify.
2. `make verify`, preserving its strict lint, short Go suite, CI-script regressions and build. Do not replace it with a diff-only lint run.
3. The existing shared-service parity journey below supplies a stable integrated entrypoint while underlying package imports move. Run direct persistence/recovery selections in their owning slices and repeat a combined affected-contract batch in Slice 16 before review; retain the existing execution-repository, WorkflowRoundTrip, WorkflowRecoveryProcess and WorkflowClosingRecovery identities at their new owners. Do not freeze nonexistent future package paths into controller commands.
4. `go test ./internal/app -run '^TestGuidedUseCaseParityNative$' -count=1 -timeout=3m` for one representative cross-adapter journey.
5. `CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^Test(GuidedUseCaseWebBrowser|WebQuizBrowser)$' -count=1 -timeout=3m` for the two existing representative Web interactions affected by model/projection changes. Confirm the available Chromium path at execution, or record an actual environment blocker.
6. One focused receiver/document/restart regression batch in its new package and one harness/process contract batch, using existing local HTTP/SQLite/child-process fixtures. Run -race only on changed shared-state owners; do not repeat the whole tree with -race.
7. The TUI Update/View benchmark comparison and `git diff --check`. A benchmark is evidence for the measured paths, not a whole-application performance claim.

The controller's ordered full-gate commands are fixed now: make verify; the exact app parity command in item 4; the exact browser command in item 5. All three entrypoints exist at the planning baseline. The strengthened quality checks join make verify as approved implementation. Package-specific persistence, recovery, transport, process and race selections belong to slice review, including the Slice 16 combined batch, and do not require nonexistent package paths in the initial configuration. Record relocated test paths in slice evidence without weakening their contracts.

Cost: runtime gates were not rerun for this planning task. Budget each slice for the required make verify gate plus its focused selection; the representative final integration checks are expected to add bounded minutes rather than hours. These are planning estimates, not measured guarantees. Measure during Slice 1, record actual timings, and investigate a material increase rather than raising timeouts or layering historical gates. No exhaustive go test ./..., full-tree race or new live-service certification is added to ordinary slice gates.

## Tickets, ownership and controller handoff

| Number | Ticket | Slice | Initial state |
| --- | --- | --- | --- |
| 1 | [TKT-20261026084821](../ticket/solved/20261026084821-quality-policy.md) | Inventory and enforceable quality policy | ready; predecessor applies |
| 2 | [TKT-20261026084822](../ticket/solved/20261026084822-quality-values.md) | Shared boundary values and residual core helpers | ready; predecessor applies |
| 3 | [TKT-20261026084823](../ticket/solved/20261026084823-quality-storage.md) | Stable records and positional SQL | ready; predecessor applies |
| 4 | [TKT-20261026084824](../ticket/solved/20261026084824-quality-commands.md) | Typed commands and central application operations | ready; predecessor applies |
| 5 | [TKT-20261026084825](../ticket/solved/20261026084825-quality-views.md) | TUI model ownership and Web view sections | ready; predecessor applies |
| 6 | [TKT-20261026084826](../ticket/solved/20261026084826-quality-leaves.md) | Gate, notifications and project packages | ready; predecessor applies |
| 7 | [TKT-20261026084827](../ticket/solved/20261026084827-quality-telegram.md) | Telegram transport, receiver and retained delivery | ready; predecessor applies |
| 8 | [TKT-20261026084828](../ticket/in_progress/20261026084828-quality-quiz.md) | Quiz domain and model adapter | ready; predecessor applies |
| 9 | [TKT-20261026084829](../ticket/ready/20261026084829-quality-authoring.md) | Authoring, document generation and guided preparation | ready; predecessor applies |
| 10 | [TKT-20261026084830](../ticket/ready/20261026084830-quality-harness.md) | Harness contracts and runtime lifecycle | ready; predecessor applies |
| 11 | [TKT-20261026084831](../ticket/ready/20261026084831-quality-tickets.md) | Ticket, readiness and execution boundaries | ready; predecessor applies |
| 12 | [TKT-20261026084832](../ticket/ready/20261026084832-quality-workflow.md) | Workflow state, store and coordinator package | ready; predecessor applies |
| 13 | [TKT-20261026084833](../ticket/ready/20261026084833-quality-integration.md) | Candidate validation and merge stages | ready; predecessor applies |
| 14 | [TKT-20261026084834](../ticket/ready/20261026084834-quality-publication.md) | Forge publication and review stages | ready; predecessor applies |
| 15 | [TKT-20261026084835](../ticket/ready/20261026084835-quality-closeout.md) | Workflow closeout and coordinator cleanup | ready; predecessor applies |
| 16 | [TKT-20261026084836](../ticket/ready/20261026084836-quality-closure.md) | Final enforcement, documentation and integrated closure | ready; predecessor applies |

Each ticket is one slice-sized review unit, linked to this plan and tracker. The tracker owns task execution evidence; Looper owns controller checkpoints and merges only under separately recorded delegation. Tickets are not a second executor and must not be submitted independently to chr native work. On verified integration, mirror that slice's resolved status and commits into its ticket; do not mark a ticket solved merely because code compiled or a PR opened.

Startup authorization is recorded above. Commit and push the approved specification/model, plan, tracker, inventory, tickets and source review to dev, then configure and start one Looper delivery set from this plan/tracker. Keep its configuration outside the repository under Looper's normal owner. Existing unrelated runs retain their ownership. The merge delegation covers only the 16 recorded slice PRs after required checks; existing bounded repair procedures require their own recorded authority.

The initial full-gate configuration names the three existing stable entrypoints above. No new controller gate wrapper or package-discovery framework is required. If the environment lacks a declared executable, report that exact prerequisite before startup; do not substitute an unrelated check or silently shrink acceptance.

## Closure

The cycle closes only after all 16 recorded PRs are verified integrated into dev, all tasks/tickets/reports reflect real evidence, all accepted remediation is resolved, and the selected integrated gate is green for its exact candidate. A supported compatibility contract, imposed external interface, intentionally distinct duplicate or unproved performance warning may be retained only with its specific justification and evidence. Unknown work is not silently relabeled as excluded.

Known fixes must stay in this plan; a failed gate is not permission to start a second cycle or declare success. In-scope integration repairs follow the existing bounded Looper repair mechanism only under recorded authority. Promotion to main and mirror publication remain separate operations.

## Planning validation

The planning consistency check passed for the 21 new artifacts: one specification, one compatibility model, this plan, one tracker, one declaration inventory and 16 tickets. The plan and tracker contain identical consecutive 16-slice maps, with matching branch, PR title, report and task identities. At that initial planning checkpoint all 97 tasks were pending, all tickets were open and both Approved fields were empty. Maintainer approval was subsequently recorded on 2026-10-26 and the 16 tickets were moved to ready with matching references; task execution evidence remains pending.

The inventory contains exactly 121 A rows, 230 B rows and 420 C rows, each assigned to one slice. Local Markdown links, ticket metadata, unique identities, new-file whitespace and unchanged baseline/working-tree state were checked. The original report and detections retain SHA-256 values 7e3004e74fdc91ae752b7483922514de32f643d141890be520ab14bafeb6e33e and 25f8d3e235705bced0c161f1851ee250ceab059b72815e258cee961450c2b36b respectively.

These checks validate the proposal's structure and preservation of its sources. Runtime gates, performance measurements and remediation acceptance remain unexecuted.
