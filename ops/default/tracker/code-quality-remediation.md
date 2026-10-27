# Code Quality Remediation Tracker

Status: in_progress
Date: 2026-10-26
Delivery set: code-quality-remediation
Approved: maintainer, 2026-10-26; exact 16-slice map, specification/model, AST checker and lint changes approved.
Implementation authority: maintainer, 2026-10-26; execute all 16 recorded slices through one Looper run.
Merge authority: maintainer, 2026-10-26; Looper may merge the 16 recorded slice PRs into dev with Rebase + Fast-forward after their required checks pass.
Active slice: Slice 11, Ticket, readiness and execution boundaries; activated after verified integration of PR #447.
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`

Plan: [Single-cycle plan](../plan/code-quality-remediation.md)
Spec: [Remediation requirements](../spec/code-quality-remediation.md)
Model: [Compatibility model](../spec/code-quality-remediation-model.md)
Inventory: [Baseline rows and owners](../plan/code-quality-remediation-inventory.csv)
Source: [Quality review](../report/code-quality-review-2026-10-26/report.md)

## Authority and state

The maintainer approved implementation, conventional commit/push and startup of one Looper run on 2026-10-26. Merge delegation covers the exact 16 slice PRs into dev after their required checks. Slices 1 through 10 are delivered through verified PRs #438, #439, #440, #441, #442, #443, #444, #445, #446 and #447. This tracker owns task/commit/validation evidence; slice tickets mirror verified delivery state. Existing completed/blocked sets and unrelated tickets are not modified.

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
| T1.1 | Reconcile every row of inventories A, B and C with its actual declaration and callers. Inventory all 14 large files, 11 duplicate pairs, seven dead symbols, eight error names, 23 unexplained errcheck suppressions, control flags and compatibility paths. Record false positives with evidence; do not equate metrics with defects. | delivered | bd6112b180b1445b139a2c43db34da2144075364 | 771 exact declaration/location matches; source call candidates recorded |
| T1.2 | Propose and implement, only after tool approval, one small standard-library Go AST check alongside the existing signature-style check. Inspect production declarations, grouped and named parameters/results, function literals and interfaces without counting receivers as parameters. Require explicit outcomes for named aliases, variadics, build constraints and generated files; do not silently skip them. | delivered | 4d3a12f563035688fc6c10635bec1674cd826e08 | go test ./scripts/goquality; focused strict lint |
| T1.3 | Capture an exact temporary debt manifest keyed by declaration identity and offending signature. Reject new or worsened violations, including moved or renamed declarations that do not retain explicit identity. Keep existing defect-oriented lint checks unfiltered; never apply a global new-from-merge-base filter. | delivered | 192edb3fe1ad5b107e058c30b2ca74f46f948380 | go run ./scripts/goquality |
| T1.4 | Use a maximum of seven parameters initially, then six after Slice 11 and five at closure. Existing violations above the active ceiling may appear only as explicit, shrinking debt. Track all nonconforming result shapes and a separate exact external-interface exception list; do not exempt locally owned interfaces. | delivered | 8e787a9961db4ee9eb51b4984066dd9374ca291d | make check-go-style; ceiling 7; 532 exact debt entries; one external exception |
| T1.5 | Exercise the checker with compact table-driven fixtures, including same-branch execution on dev, unchanged debt, new declarations, named results, aliases and stale exemptions. Establish the package dependency and naming ledger used by later moves; resolve shared-value ownership before a move can form an import cycle. | delivered | 7cf213f4ed4e10290feb4efdcfb8e477b291f46c | go test ./scripts/goquality; focused strict lint; production scan |
| T1.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 72c769c47790e2f502309e2646f7769b97419bba | make verify; git diff --check passed on 0d609b0c590c7e23a184513393fa64ab8d394b67; PR #438 integrated |

### Slice 2: Shared boundary values and residual core helpers

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T2.1 | For root-level helpers outside the later capability slices, replace unrelated positional outputs with named results and repeated cohesive arguments with operation inputs. Review the config, context, migration, workspace, report and proving detections assigned by the inventory. | delivered | e7b109ce268f9ceace9c981d64b09274fc883104 | focused config/init/proving/context/guided checks; production policy scan |
| T2.2 | Preserve absence, applied, changed, cached and partial outcomes separately. Use a named result with a found or changed field when absence is normal; use errors.Is-compatible errors where absence is already an error. Preserve all caller branches and error wrapping. | delivered | d326ac2c8e3b3bc5199d6e0227c7d26d3daeeb96 | focused config/admission/guided/cache/report/init regressions; policy scan |
| T2.3 | Resolve the real shared values, path/hash primitives and boundary ports needed by package extraction. Reuse existing owners; introduce a focused owner only when the actual dependency graph requires it. No generic common, util or options package, and no stored context.Context. | delivered | c7c885e14a02e7ba48314844e8f2ff9d7233dd1b | focused config/workspace/guided/proving regressions; unchanged function bodies and import direction |
| T2.4 | Review oversized proving.go and shared parsers/schema files. Split distinct responsibilities, preserve cohesive parsers or schema literals with a recorded reason, and simplify justified control-flag helpers. Assign every remaining root production file a final owner in the dependency ledger. | delivered | 21256878c7fe9f02c7f03eb7948c0901085e83b6 | 49 exact declaration bodies/types preserved; focused proving/workspace/config regressions; 678 file owners and 24 row dispositions |
| T2.5 | Fix owned dead/test-only helpers, error names and unexplained suppressions encountered here. Preserve meaningful cleanup errors; suppress only an explicitly justified, owned best-effort close. | delivered | 38f215efce75a1470284c38fc5e590506891fefb | TestProvingLogFailure failed before the fix and passed afterward; existing proving regressions passed |
| T2.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 4275d727615ce66861ebf4a114f494e9d6252369 | make verify, focused core regressions and git diff --check passed on de3691d3b0155f850006f242ee2d99062a995ce0; PR #439 integrated |

### Slice 3: Stable records and positional SQL

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T3.1 | Freeze compatibility fixtures for WorkflowInvocation, WorkflowMergerInput, IndexedTicket and WorkflowWorkStatus before grouping fields. Record all field names, tags, omit rules, timestamps, state values and canonical hashing paths in the model companion. | delivered | ba53ac7259f8004345715f90c9482ea5544cbe36 | TestQualityRecordWire on unmodified records; zero, empty and populated byte fixtures plus field manifests |
| T3.2 | Group cohesive identity, Git, timing and outcome fields while preserving existing encoded bytes wherever bytes feed hashes or retained evidence. Keep a dedicated flat serialization shape when embedding changes order or omission. Update construction, scanning and mapping code together. | delivered | d2c7676bca1271270ba2ec175b618264f872a50f | frozen byte fixtures and focused merger/invocation/index/status regressions |
| T3.3 | Inventory every production SQL bind or scan above 16 fields, not just the three examples. Introduce one row mapping and one explicit ordered column definition per affected table, with nearby values and scan targets. Keep SQL identifiers static and user data bound. | delivered | 8e261ff51c808bcde5ea4f12708721dc226d0e95; 7f472692cf536292337d61a495ae7bed8fb704ff | eight table mappings and ordered static column definitions; legacy and focused real-SQLite regressions passed |
| T3.4 | Start with execution_repository.go, remote_command_record_repository.go and telegram_document_outbox.go. Share only small statement-construction code that is actually identical; retain table-specific updates, filters, conditional writes and transaction ownership. | delivered | 7f472692cf536292337d61a495ae7bed8fb704ff | focused execution/remote/notification/index/conversation/document regressions; strict core lint; policy scan: 517 unchanged debt entries |
| T3.5 | Use real SQLite round trips with distinct same-type column values, null/empty distinctions, inserts, updates, omitted/default values and scan-order checks. Reopen old fixtures. Prove rollback and durable owner records remain unchanged; no schema migration is needed merely to rearrange Go fields. | delivered | 8e261ff51c808bcde5ea4f12708721dc226d0e95; 346721b45aa01e8e7fe72fe41b7f9fbb7c2473da | TestQualityStorage: legacy reopening, full stored-column comparisons, NULL/default transitions, distinct peer IDs and hashes, and injected transaction rollback passed; strict core lint passed |
| T3.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 073af8e920058308fe15839e7f832d43792fb4a0 | make verify, focused record/storage regressions and git diff --check passed on fec20d73396156cf8f389fb11e68de8ca0cadcf1; PR #440 integrated |

### Slice 4: Typed commands and central application operations

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T4.1 | Replace root Metadata dependency lookup with one typed command constructor value and method actions. Cover every command constructor, not just the reported 15; remove ignored dependency and output parameters and the service locator only after all production callers move. | delivered | c6da9e52f81ab6a2260dfbaa3e374c6dce3e6bfa | 134 constructors; 74 forwarding actions removed; root/ticket/author/Telegram setup regressions and metadata-independent asset binding passed; strict app/entrypoint lint passed; 513 debt entries |
| T4.2 | Mechanically split ticket.go and author.go by actual subcommand responsibility. Split maintenance.go and service_work.go where their independent operations justify it. Keep moves, signature changes and semantic extractions in distinct task commits. | delivered | 398dbf67a8aa92ac0bd45a9c6aac313fec6b3d88 | 292 exact declarations preserved across 29 focused files; largest file 617 lines; focused real command/workflow regressions and strict app lint passed; 22 debt identities explicitly moved |
| T4.3 | Move reusable ticket loop, review, merge and start-preflight orchestration behind the existing central service in service_ticket files. Use one cohesive ticket-loop value while passing context explicitly. Route existing CLI, TUI, Web and Telegram callers through the same operation where semantics match. | delivered | 1f0c7c7c2f8237bef5653c975623a55d4b77b36a | focused real ticket/start/loop/review/merge and agent-work regressions passed; strict app lint passed; 505 debt entries; explicit operation inputs and partial loop outcomes |
| T4.4 | Remove service.dumpDirectory and the other app-owned test-only symbols after checking production call paths; redirect retained tests to live operations. Resolve newTelegramServeCommand, serviceFromCommand and owned error-name/unparam findings without deleting meaningful regression coverage. | delivered | f40d62d4775628abf58d527d93e11e8fae557a7e | real root, metadata, report, ticket, dump and work command checks passed; strict app lint passed; 503 exact debt entries |
| T4.5 | Finish every signature row assigned to this slice, including command entrypoints, gate/report adapters and ticket-loop results. Preserve CLI flags, output, cancellation, finite attempts and host-owned publication authority. | delivered | 42db97f8af41b95f61120415afdc16931891b47c | 69 assigned rows: 47 fixed and 22 cohesive five-input contracts; 472 exact debt entries; real CLI/Git/SQLite/HTTP/provider batch passed in 92.5s; strict app lint passed |
| T4.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 03f576b9f7e35fa8f2f6c43f617bfd0c64c4b20b  make verify, focused real command/work regressions and git diff --check passed on 6b287a79bbe8b545a4d36633d24ceccea01469d0; PR #441 integrated |

### Slice 5: TUI model ownership and Web view sections

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T5.1 | Measure the existing Model Update and View paths with a small benchmark before changing receivers. Make *Model the Bubble Tea model and group navigation, interaction, layout, background work and per-view state by owner. | delivered | c016b6c2a5a02fb2095938272ced3f3e8b007f7c; 6d66d759799af7e94ccec3ed90f3b7ac6c8afd2e | baseline Update/View benchmarks; 85 fields grouped by 17 owners; pointer receiver/constructor contract; all TUI tests and signature policy passed; queued refresh captures its executor/request |
| T5.2 | Convert internal update helpers into pointer-receiver methods returning tea.Cmd or a named outcome. Retain only externally imposed interface signatures, initially Model.Update. Inspect captured commands, message ordering, cancellation, snapshots and test assumptions for aliasing or concurrent mutation. | delivered | 14d8c3aa5a606b3965b81be419c8d0d4becf41b7 | 37 internal result violations removed; all TUI tests and strict lint passed; no command closure reads Model; 435 exact debt entries |
| T5.3 | Split model.go and shell.go by view responsibility; express action dispatch and key hints as data where this is clearer. Resolve duplicate pairs D01 and D11 and the TUI plan-action adapter switch while preserving focus, empty states and mode-specific actions. | delivered | ef54c01fec174cf1421414fd93aaac191a942484; 264941b4cff3623e4cffd77f26695031669c3048 | 193 declarations mechanically preserved; D01/D11 share catalog/detail/panel rendering; 32 frozen complete frames unchanged; ordered dispatch/hints; plan adapter and real guided TUI/plan checks passed; strict app/TUI lint passed |
| T5.4 | Split GuidedUseCaseView by existing page sections and WorkflowWorkStatus by cohesive projection components. Keep application service results adapter-neutral, preserve public JSON via the model contract, and keep the current Web forms and visual language. | delivered | 73cf1944ab4829db08d8e0cf7d791a4adbe32be0 | all 47 guided fields/tags/order and zero/empty/populated bytes preserved; seven narrow template inputs; existing adapter-neutral workflow status components retain frozen bytes; HTTP guided/control checks and strict app/Web lint passed |
| T5.5 | Review webui.go and every assigned TUI/Web signature detection. Compare the measured Update/View behavior after the change and run a focused race selection for changed asynchronous ownership; do not replace all large values with pointers. | delivered | 28fbcdadde20bea738693e0e32ba10257000a625; 739cdf1b3e1b6734a68d4ba7a0c891f030074f9c; cf3c0f537406aac6b379609009f68f2c3a77b220 | 50 rows: 47 fixed, two cohesive five-input renderers, one exact interface; 425 debt entries; 53 exact Web declarations split into seven files; focused checks/strict lint passed; benchmark comparison recorded: idle/resize 0 B/op, focus 1,096 B/op; View allocation counts unchanged; focused race/browser checks run on the final PR head through the controller |
| T5.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 80077d2cf393f121ffe107a7c9ca35c943faef4c | PR #442; focused checks and benchmark comparison passed; ordered make verify, TUI/Web/app, ownership race, Chromium and whitespace checks run through the controller on the clean PR-reference HEAD; six ordered controller checks passed on 5a84d1af3d6a855ac1ce1e3f767080db3ef101c3; PR #442 integrated |

### Slice 6: Gate, notifications and project packages

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T6.1 | Move gate behavior into internal/core/gate with its narrow inputs and results. Resolve its shared-error and path dependencies before moving; update direct callers in the same compiling change. | delivered | cacbf861e1d3; aabf8c55c1bd; ca51bf101665 | 125 exact moved declarations; focused real Git/child validation and app gate regressions passed; four named result conversions, no root facade; strict core/app/gate/configvalue lint passed; final slice checks pending |
| T6.2 | Move notification state, delivery ownership and persistence into internal/core/notify. Keep notification idempotency, event ordering, backoff and acknowledgement semantics; do not fold it into Telegram receiver storage. | delivered | b4f60d9e2612; ee3144d5d28c; 101ff15a3c7b; 241227d6066f; 439667cfd815 | 76 exact notification move proofs and 27 shared-payload token proofs; real SQLite/HTTP snapshot, quota, retention, uncertainty/reopen and conditional-delivery checks passed; strict lint passed |
| T6.3 | Move project identity, registry and configuration behavior into internal/core/project as justified by the dependency ledger. Retain neutral shared primitives at their actual lowest owner, not copied package helpers. | delivered | aafd48f494c9; bb30375e56de; f4722f65a22e; 31c756cc3659 | 39 exact project move proofs; explicit directory inputs and root-binding digest; focused real SQLite/registry/active selection/Telegram and app maintenance checks passed; strict core/app/project/database lint and signature policy passed |
| T6.4 | For each package, move files with git mv in one commit, then apply precise package-qualified names and typed inputs/results in subsequent commits. Remove old entrypoints, re-exports and duplicate ownership after callers move. | delivered | aabf8c55c1bd; ca51bf101665; f4722f65a22e; 31c756cc3659; 241227d6066f; 439667cfd815; 69d0358d9912 | 240 exact mechanical declarations, followed by separate qualified-name and named operation/result changes; direct callers and no root facades |
| T6.5 | Finish assigned signature, state-constant, error-name, control-flag and suppression work. Add or update required package READMEs and the architecture glossary with code-backed meanings. | delivered | 69d0358d9912; 112fd73e0153 | All 19 assigned rows: 16 fixed, three cohesive five-input contracts; 410 debts and unchanged external exception; named authoring results preserve outer receipt binding; enforced import bounds and adversarial AST fixtures passed |
| T6.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | bd553e812057d7d0d943c385ee99787c41cbdd59 | Controller make verify, focused non-short core/app regressions, focused notification/authoring race checks and git diff --check passed on c7d7040fde4385a64b20322da1bc5874a38c2deb; PR #443 integrated |

### Slice 7: Telegram transport, receiver and retained delivery

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T7.1 | Move the core Telegram receiver, transport and retained operation/outbox behavior into internal/core/telegram. Introduce a nested outbox owner only if storage and transport separation actually need it; preserve one transactional owner for each existing operation. | delivered | 734cafffcefb; 071a0e1dd62a; 7a845d3f83df; 450b92e4cf5d; 4fe90efe678e; 128bd7279de8; 71a0527b449c | 71 production files and 51 owner tests moved; 453 identical declarations; direct app/domain consumers, shared pure values and enforced import bounds; original transactions and full guided domain guard retained |
| T7.2 | Replace SendDocument, SendAction, presentation senders and notice insert argument lists with named request/row values. Convert all assigned result shapes while preserving missing, stale, duplicate and acknowledgement outcomes. | delivered | 6122bca87167; 12afe643af87; cb4affbe32cf; 430ed7a6973c | All 156 assigned occurrences conform: 37 A, 52 B watch items, 67 C; named operation inputs/results preserve recognition, missing/stale/duplicate and acknowledgement distinctions; 307 remaining debts belong to other owners |
| T7.3 | Apply the row mappings from Slice 3 to notice, document, action and response queues. Preserve exact original-document attachment bytes, captions, fallback behavior, finite quotas and restart reconciliation. | delivered | a11b75fa23c1; b9e70d85c053; cb61567da591 | Typed notice/action/response/document/delivery rows; 174 unchanged SQL expressions; distinct-column SQLite and swap mutation proof; original frozen bytes, caption fallback, quotas, uncertainty and restart regressions pass |
| T7.4 | Resolve duplicate dispatch pair D06 through a narrow helper with separate handler tables. Verify that removing telegramGuidedResultExport and telegramHelpReply does not remove their live replacements or useful behavioral tests. | delivered | a055ba79f991 | D06 intake has one owner with distinct explanation/work route records; both test-only wrappers removed and live renderer/result guard tests retained; focused real work/explanation/help and guided-result guard checks plus pinned app lint pass |
| T7.5 | Consolidate error type names, typed receiver/operation/outbox states, unused parameters and justified close handling. Update app adapters directly; do not re-export the moved package through core. | delivered | 71a0527b449c; cb61567da591 | Six transport state machines retain stored strings; explicit Error names, observable SQL cursor cleanup and one justified HTTP body-close suppression; unused serve inputs were resolved by Slice 4 |
| T7.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 00546d27e408af797280a10377717603d9261b64 | Controller make verify, focused non-short Telegram/core/app regressions, focused concurrency/recovery race checks and git diff --check passed on 92f91847ffdc374c53b8044e9435b7af6d1cb722; PR #444 integrated |

### Slice 8: Quiz domain and model adapter

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T8.1 | Move quiz state, transitions and domain payloads to internal/core/quiz; move OpenAI-specific request/schema/response behavior to internal/core/quiz/openai. Keep provider-specific IO out of the quiz domain. | delivered | 2bd2c3156e3c; 3e12b10d2219; 83c1efb4677f; 4b8b1869f972; 0a751d414f31; be31a8b01b75 | Direct domain/model/OpenAI owners, shared generated values and private journal images; 404 extraction entries; enforced imports and focused real-boundary checks pass |
| T8.2 | Split quiz_openai_model.go by schema, request and response responsibilities; split quiz_model_payload.go and service_quiz.go where cohesion warrants it. Retain the justified schema-literal complexity exception rather than creating artificial fragments. | delivered | 83c1efb4677f; ecf4baf844ac; 4b8b1869f972; be31a8b01b75 | Cohesive schema/instructions/decoding/context/payload and API request/response/trace files; service split; justified schema literal exception retained; strict lint and focused checks pass |
| T8.3 | Replace assigned quiz model argument/result tuples with domain inputs and outputs, preserving question identity, answer sufficiency, model errors, token accounting and finite round limits. | delivered | d8782c80998f; ecf4baf844ac; be31a8b01b75 | All eight dispositions: six fixed, two cohesive legal five-input operations; named raw/bound output, trace observation, receipt lookup, metadata and parent/trunk results; 301 debts belong to other slices |
| T8.4 | Remove service.newQuizModelRequest if its production path is still unused; update tests to the live request path. Preserve Deep/Fast/manual/guided behavior and existing provider selection. | delivered | 4b8b1869f972 | Production-unused newQuizModelRequest removed; four readiness tests use real persisted entries with newQuizModelRequestWithEntries; focused manual, Deep/Fast and guided service checks pass |
| T8.5 | Keep cancellation explicit and secrets out of payloads/logs. Finish local signature, flag, state and suppression entries and update the dependency ledger and glossary. | delivered | 3e12b10d2219; 0a751d414f31; be31a8b01b75 | Explicit host/cancellation and actual provider identity; fixed OpenAI wire flags; original states, body limits and cleanup semantics; HTTP authorization/trace secrecy and owner policy checks pass |
| T8.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 300a3c956d8e | PR #445 merged at 2c649d98d3297914ac7c118cf3aecd46fd8276d1; controller ledger: make verify, focused non-short model/service/wire tests, focused recovery/capacity race and whitespace all pass on that exact HEAD |

### Slice 9: Authoring, document generation and guided preparation

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T9.1 | Move authoring session and dump/spec/plan artifact behavior into owned internal/core/authoring boundaries; move plan/catalog behavior into its own capability only where the dependency ledger demonstrates a distinct responsibility. Keep guided inputs and committed-source provenance intact. | delivered |e63688b43071; a2e76eec801e; 2f9e9717c027; 0e1d949a3fca; f7df89655484; dbdb1ac183be; ea96e86c861a; d8b7a9921cb5; eac5ca5dd8e1; 7d8acfadb0ff; 2527d57f8339; 8490f7aeb303; 4e8ed9c2792b; b263991b7f52; 769e2047686c; 23b7fe8d1a9b; 5c533bfdb8cf; 64376e9f69c1; ce0f625d1ca1; 59c5e103482c; ff99e09f4347; bb954ba8a84b; 80dd1afed160; 00497b20b93d; a654e9300f3b; 3278be43b287; ffcce459864c; 2af8e9523cbb; fad092e249e3; 23b2d04b1e7f | Complete original session/artifact owner moved: 95 source and 53 test files; native context/harness adapters and six runtime/execution context files retain actual responsibilities; 1243 preimages; real lifecycle/wire/publication checks, import bounds and strict lint pass |
| T9.2 | Read the governing spec/plan for both sides of D02, D04, D05, D07, D09 and D10. Extract catalog, generation, guided-preflight, authoring-step and dump-update logic only where both copies implement the same rule; record intentional semantic differences with precise evidence. | delivered | ae88f240639d; e2d01335ce70; 129fb1b9d256; 678d9e6abaf4; bba7d53e5f34; 712e812ac131 | All six governing-contract comparisons recorded: bounded catalog, model request, latest scoped selection, retained-step, artifact write and CLI revision/output rules shared; distinct authority/publication/provenance/receipts retained; existing regressions and strict lint pass |
| T9.3 | Share the actually common context/model/validation/recording steps of spec, acceptance and plan generation. Keep artifact-specific schemas, approval, source observation, publication and idempotency decisions explicit; do not introduce a generic generation framework. | delivered |bba7d53e5f34; 701bbd85d763; ffcce459864c | Concrete payload/model steps shared; workspace context is injected at the original observation point, request construction is pure, native harness behavior retained; existing real context/provider/trace tests and strict lint pass |
| T9.4 | Split service_plan.go and authoring helpers by responsibility, finish assigned context/dump/spec/plan/guided result and parameter shapes, and handle the acceptance_feature.go large-file disposition. | delivered |1d108d03c5c2; d577def590fa; 571c431a25d8; d647834c037d; 65353580359a; 34d663404459; 0adb83936123; 115897594fc5; 969eba2252d7; 051c19eac8fb; bf04bf96c5d6; 7d8acfadb0ff; 1671d3ea3af1; 89898ddd5bd8; fcd5f288a097; 3278be43b287; ffcce459864c; 2af8e9523cbb; fad092e249e3; 23b2d04b1e7f | Plan/acceptance splits and package naming complete; all 89 current signatures reconciled: 60 tuple fixes, three parameter fixes (two from Slice 4), 26 cohesive five-input contracts; 237 exact debts, ceiling seven, one external exception; focused checks and strict lint pass |
| T9.5 | Use ports to the existing workflow owner for guided preparation instead of a second coordinator. Preserve manual authoring, legacy specification formats, Git checkpoints, exact hashes and retry receipts; update existing docs in their native format. | delivered |63a843aa48a6; bb954ba8a84b; a654e9300f3b; 3278be43b287; ffcce459864c; 2af8e9523cbb; fad092e249e3; 23b2d04b1e7f | Original native observation, guarded same-claim session reads and exact artifact codecs compose the existing native owner; native admission/preview/materialization/result proofs, plan resume and credentials/signing stay native; real guided lifecycle and wire checks pass |
| T9.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | 429c357e99a4 | Canonical PR #446; immutable report introduction e63688b43071c51e17917a0f85a02ebf2de88f01; controller ledger and PR Validation own the ordered exact clean-head make verify, focused non-short checks and whitespace; verified integration at 0ce066b1599b55f6efc3f151bb960a74b75d5e91 |

### Slice 10: Harness contracts and runtime lifecycle

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T10.1 | Use the existing internal/harness packages for adapter-neutral runner contracts and provider/process adapters. Move core-owned admission/binding policy to internal/core/harness when required; do not make the runner library import project workflow policy. | delivered | 7d00c7770dbc; a261db1cb5b6; 06bcd16df8b3; bf9849e8335b; 79319ed7b15d; 50fb5406fa2f; ec1d09dfac3b; 92e386c0eedb; ea0987b16bcf; eb952aac2190; 1358e6e3; 1e9ba276; afb9bb71; eba7672f; bde6294f; 3e1b123e; 9984aa3f; 93333783; ff6c5c6d; 465beb55; dd82e0a5 | Direct identity/host/runtime/daemon/trace owners and concrete native composition; 327 verified declaration/constant preimages; dependency, real lifecycle, config, wire and adapter checks pass. |
| T10.2 | Convert runtimeArtifactPaths and runtime command/path/result tuples into existing or new precise values. Refactor supervised-process repeated inputs and harness/process exchange signatures with explicit resource ownership. | delivered | e6dd03717ce3; de7dcac38eb7; da9df7fdb102; b1656b55c5d7; 6d0163ab2096; fc42f7dff782; 3e1b123e; 9984aa3f; afb9bb71 | All 38 occurrences reconciled: 31 fixed, seven cohesive contracts retained; two additional resolver contracts fixed; 207 exact debts; named runtime/process/Pi values and caller-context checks pass. |
| T10.3 | Separate executable identity, model/profile selection, frozen capabilities and host-owned execution policy. Preserve serialized binding hashes, finite budgets, output retention and provider-independent error behavior. | delivered | 7d00c7770dbc; de7dcac38eb7; 06bcd16df8b3; bf9849e8335b; b1656b55c5d7; 6d0163ab2096; 79319ed7b15d; 50fb5406fa2f; ec1d09dfac3b; 92e386c0eedb; ea0987b16bcf; eb952aac2190; fc42f7dff782; afb9bb71; eba7672f; bde6294f; 3e1b123e; 9984aa3f; 93333783; ff6c5c6d; 465beb55; dd82e0a5 | Identity, profiles, capabilities and frozen launches are separate from concrete host policy; exact wire/hash/SQL/default/path contracts and independent native config fences pass. |
| T10.4 | Derive contexts at the caller or use a narrowly owned cancellation handle; do not store operation contexts or lose cleanup on early errors. Review relevant boolean flags and test-only wrappers with real call-path evidence. | delivered | e6dd03717ce3; fc42f7dff782; 93333783; ff6c5c6d; d48ea454 | Caller-owned contexts/cancellation, early cleanup and existing capacity lifetime pass real race checks; binary flags and removed trace facade have concrete call-path dispositions. |
| T10.5 | Move and rename in separate commits, finish assigned signatures, update app adapters and package documentation, and remove cross-package compatibility aliases before closing the slice. | delivered | 7d00c7770dbc; a261db1cb5b6; e6dd03717ce3; de7dcac38eb7; 06bcd16df8b3; bf9849e8335b; da9df7fdb102; b1656b55c5d7; 6d0163ab2096; 79319ed7b15d; 50fb5406fa2f; ec1d09dfac3b; 92e386c0eedb; ea0987b16bcf; eb952aac2190; fc42f7dff782; 1358e6e3; 1e9ba276; eba7672f; bde6294f; 3e1b123e; 9984aa3f; 93333783; ff6c5c6d; 465beb55; dd82e0a5; d48ea454 | Moves/naming are separate task commits; direct consumers, owner/boundary ledgers and package docs reconcile; no migration aliases; focused checks and pinned strict lint pass. |
| T10.6 | Validate this slice, update its ticket/tracker, produce its report, and verify integration into dev. | delivered | ce13c3111491; add303a00a38 | Canonical PR #447; immutable report introduction a95068b39dbfc2288406df06a4f562b44994b3af; controller exact ordered make verify, focused integration/race and whitespace checks passed; integration verified at add303a00a38df26dd23b25dbf5dc5dc59cf0e9a |

### Slice 11: Ticket, readiness and execution boundaries

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T11.1 | Move ticket models, indexing and repository-owned ticket rules into internal/core/ticket. Isolate execution/readiness logic in a focused existing or necessary capability owner rather than putting unrelated execution orchestration in the ticket package. | in_progress | 27695c85c175; f6dc2e3bb7d1; f231fa850074; 07fc450731e1; bff89207ea2b; f48f661726f4; 68e8ac7c46eb; 85087e758af6; 643ced936086; 08e9ae2aee5c | Ticket records/rendering/parsing/validation/status have direct owners; native projection/resource composition preserves shared SQLite and source/cache authority; 494 declaration/test preimages; 24 component tests placed with ticket; ticket SQL/discovery/reconciliation and key reservation have direct owners; focused real-file/SQLite/native/app/TUI/wire/dependency tests and pinned lint pass. Readiness/repair/worktree ownership and guards are implemented. Remaining execution record/runner and parser/validation test ownership pending. |
| T11.2 | Remove duplicate readiness pair D03 before splitting work_readiness_service.go. Preserve the distinction between user checkout, ticket worktree, preview, authoritative readiness and explicit safe repair. | in_progress | 93847b39ae8f | D03 normalization removed before service splitting; unchanged public check/repair/worktree/selection regressions, signature policy and strict lint pass. Service/owner split is implemented in core/execution with direct native consumers and 268 additional preimages; final slice validation remains pending. |
| T11.3 | Finish assigned execution/work/evidence/remote-command/review helper signatures and operation arguments. Move timePtr to a test helper and remove unused production-only wrappers without losing live-path coverage. | in_progress | f231fa850074; 9ec776761252; 68e8ac7c46eb; a7c9ebe7ed39 | Assigned metadata-time tuple and six-input indexed-time helper fixed; timePtr is test-only. Real parser/SQLite/wire regressions and strict lint pass. Reconciliation now uses a named parse outcome with an optional blocker; the retained-row regression passes. The ticket number/presence reader is also named; native key observation stays separate from the injected file-writer allocator. Named readiness/repair/status/configuration results close 13 more assigned tuples; 190 exact debts remain at ceiling seven. Other assigned signatures/wrappers pending. |
| T11.4 | Use typed states per ticket/execution/review machine and a typed workflow preparation mode. Preserve wire spellings, database values, absence handling and existing transition guards. | pending | — | not run |
| T11.5 | Reconcile each owned five-parameter watch item, control flag and legacy path. Advance the global parameter debt ceiling from seven to six only when above-six debt has been fixed or explicitly assigned to the remaining workflow slices; record exact counts. | in_progress | f231fa850074; 9ec776761252; 68e8ac7c46eb | Partial signature disposition ledger records 17 fixed assigned entries. Ceiling remains seven with 190 exact debts and one external exception; remaining signature/watch/flag/legacy dispositions and the six-parameter phase are pending. |
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
| 1 | delivered | TKT-20261026084821 | #438 | 0d609b0c590c7e23a184513393fa64ab8d394b67 | `ops/default/report/slices/code-quality-remediation/slice-1-policy.md` |
| 2 | delivered | TKT-20261026084822 | #439 | de3691d3b0155f850006f242ee2d99062a995ce0 | `ops/default/report/slices/code-quality-remediation/slice-2-values.md` |
| 3 | delivered | TKT-20261026084823 | #440 | fec20d73396156cf8f389fb11e68de8ca0cadcf1 | `ops/default/report/slices/code-quality-remediation/slice-3-storage.md` |
| 4 | delivered | TKT-20261026084824 | #441 | 6b287a79bbe8b545a4d36633d24ceccea01469d0 | `ops/default/report/slices/code-quality-remediation/slice-4-commands.md` |
| 5 | delivered | TKT-20261026084825 | #442 | 5a84d1af3d6a855ac1ce1e3f767080db3ef101c3 | `ops/default/report/slices/code-quality-remediation/slice-5-views.md` |
| 6 | delivered | TKT-20261026084826 | #443 | c7d7040fde4385a64b20322da1bc5874a38c2deb | `ops/default/report/slices/code-quality-remediation/slice-6-leaves.md` |
| 7 | delivered | TKT-20261026084827 | #444 | 92f91847ffdc374c53b8044e9435b7af6d1cb722 | `ops/default/report/slices/code-quality-remediation/slice-7-telegram.md` |
| 8 | delivered | TKT-20261026084828 | #445 | 2c649d98d3297914ac7c118cf3aecd46fd8276d1 | `ops/default/report/slices/code-quality-remediation/slice-8-quiz.md` |
| 9 | delivered | TKT-20261026084829 | #446 | 0ce066b1599b55f6efc3f151bb960a74b75d5e91 | `ops/default/report/slices/code-quality-remediation/slice-9-authoring.md` |
| 10 | delivered | TKT-20261026084830 | #447 | add303a00a38df26dd23b25dbf5dc5dc59cf0e9a | `ops/default/report/slices/code-quality-remediation/slice-10-harness.md` |
| 11 | in_progress | TKT-20261026084831 | pending | — | `ops/default/report/slices/code-quality-remediation/slice-11-tickets.md` |
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
