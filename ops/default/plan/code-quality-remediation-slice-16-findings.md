# Final source reconciliation

Status: source assessment complete; PR-head and integrated acceptance pending
Task: T16.1 through T16.5
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`
Assessed implementation: `50d562611bfb5d8e725ff62da389fd9c2e372818`

This is resolution evidence for the approved source review, not a replacement
plan or a claim that integration or the final gate has occurred. The original
review and detections remain frozen until controller-owned closure. The owning
slice reports record implementation commits and exact executed checks. Regression
names in inventories identify protection; they do not claim another execution.

## Signatures, files and ownership

The [final signature map](code-quality-remediation-slice-16-signatures.csv)
accounts for all 771 original occurrences: 121 A, 230 B and 420 C. It records
570 conforming named-contract dispositions, 200 cohesive five-input cases and
one exact Bubble Tea interface exception. Each B row has its own semantic
assessment. The current checker inspects 9,411 production signature declarations
including interfaces, function types and literals, with maximum five and one
external exception. This broader scope is not compared as a function-count
delta with the review's 8,286 function declarations.

The [secondary map](code-quality-remediation-slice-16-secondary.csv) preserves
all 454 original path/line/evidence rows. The source-frozen candidates contain
130 boolean-signature references and 291 compatibility references, rather than
the review prose's 88 controls and approximate 300 references. Each has a current
declaration hash, meaning/contract, regression protection and retirement
condition. Existing 217 workflow assessments retain their individual C15
contracts. Two unreachable TUI wrappers are explicitly removed. Supported
legacy readers, formats and commands retain versioned retirement conditions;
their presence is not unresolved migration debt. Ordinary data, recognition,
presence and focus booleans are not automatically replaced by mode enums.

All 14 large-file rows record the actual operation split and current owners,
with the corresponding move/contract ledger and per-capability regression.
All original files are smaller or removed. A cohesive strict schema literal
remains intentional; line count is an inspection signal rather than proof of
semantic extraction. The [core map](code-quality-remediation-core-owners.csv)
accounts for the remaining root files. The workflow maps reconcile all 908
source Store methods; 777 retained methods enforce actual persistence/authority
guards. Stage policy, host lifecycle and coordinator operations have separate
owners. All 22 consumer ports are finite operations, without Store or opaque
escape hatches. There is no root workflow forwarding facade or type alias.

## Six wide values

| Original value and width | Implemented ownership | Exact compatibility evidence |
| --- | --- | --- |
| `tui.Model`, 85 | One pointer model with 17 named substates; commands capture immutable IO inputs before leaving the loop. | [85-field map](code-quality-remediation-slice-5-model-fields.csv), `TestModelOwnership`, final benchmark below. |
| `WorkflowInvocation`, 48 | `workflow/state.Invocation` embeds identity, admission, integration, scope, progress, input, execution, candidate, delivery, review and preparation values. | [Record fields](code-quality-remediation-record-fields.csv), frozen `TestQualityRecordWire` and `TestWorkflowRoundTrip`. |
| `GuidedUseCaseView`, 47 | Guided page/question/ideas/lifecycle/run/step/diagnostic sections; partials receive their own section. | [47-field map](code-quality-remediation-slice-5-guided-fields.csv), [partial map](code-quality-remediation-slice-5-web-sections.csv), `TestGuidedViewWire`. |
| `IndexedTicket`, 40 | `ticket.Indexed` has identity, attention, summary, classification, source, review, merge, times and projection owners. | Record fields, frozen wire images, SQL binding and index round-trip tests. |
| `WorkflowMergerInput`, 39 | `workflow/state.MergerInput` separates operator, identity, Git, context, scope, validation and limits. | Record fields and exact frozen wire images. |
| `WorkflowWorkStatus`, 37 | App projections embed bindings, lifecycle, identity, selection, outcome, scope, counts, control, delivery and page sections. | Record fields, frozen status JSON and the existing guided view integration. |

These groupings preserve flat field names, omission rules, retained enum values,
hash inputs and SQL order; fewer top-level fields alone would not prove that.

## Eleven duplicate pairs

| Pair | Shared rule and retained boundary | Evidence and implementation |
| --- | --- | --- |
| D01 | Spec/Plan catalog, detail and focus rendering; per-kind actions/status remain distinct. | [Slice 5 presentation](code-quality-remediation-slice-5-presentation.csv), 32 frozen complete frames; merged PR #442. |
| D02 | Bounded document catalog traversal; typed readers, archives, source validation and byte accounting remain distinct. | [Slice 9 duplicate ledger](code-quality-remediation-slice-9-duplicates.csv), `678d9e6abaf4785c88569e2ab8cb5427cce83fe4`. |
| D03 | Readiness normalization; repair retains the Git mutation interface and explicit safe-repair gates. | [Slice 11 duplicate ledger](code-quality-remediation-slice-11-duplicates.csv), `93847b39ae8f156545a540cf252289a92aecd736`. |
| D04 | Model context/request assembly; spec and acceptance discovery, schemas, readiness, artifacts and approval remain distinct. | Slice 9 ledger, `bba7d53e5f341edc8fdb3dce82d1d6eb30879341`. |
| D05 | Scoped latest-node selection and expected absence; document readers, lifecycle and publication stay distinct. | Slice 9 ledger, `712e812ac13123a4bcc58dedc41d450a3b514af2`. |
| D06 | Telegram retained lookup/parsing route; actual authority, model, document and native handlers remain distinct. | [Slice 7 cleanup](code-quality-remediation-slice-7-cleanup.csv), `telegramCommandRoute`; merged PR #444. |
| D07 | Guided scoped resolution and retained authoring-step entry; preparation publication and native admission keep their own guards. | Slice 9 ledger, `ae88f240639d21b9ebdad6801abddaa48deacc26`. |
| D08 | Contained receipt read, expected absence and decode are shared by `readWorkflowOperationReceipt`. Mutation/control checksums, identities, schemas and locked writers remain distinct. One generic receipt writer would conflate those contracts. | [Slice 12 report](../report/slices/code-quality-remediation/slice-12-workflow.md), `4ffe06de32d98e02ab388fa0613fd9f0b61e2933`; ten receipt presence/corruption/cross-schema regressions. |
| D09 | Authoring envelope/directory/body write inside existing exclusion; collision checks, provenance and legacy promotion semantics stay distinct. | Slice 9 ledger, `e2d01335ce70ba2d708346624fac50e0375bee1f`. |
| D10 | Dump CLI service/revision/printing mechanics; save and append keep typed requests and separate replacement/extension contracts. | Slice 9 ledger, `129fb1b9d256ce21b37dc3b50af498fd44ecea57`. |
| D11 | The actual source ranges are Spec/Plan panel layouts, sharing bounds/focus/maximization. The earlier ticket-column label was incorrect. | Slice 5 presentation ledger, complete frozen frames; merged PR #442. |

The existing linear spec/acceptance generation suppressions retain logging and
failure attribution through distinct guarded publications. They predate this set
and are not new migration exemptions. Permanent production dupl uses threshold
150; intentional semantic differences retain their specific contracts above.
The other five current complexity suppressions cover the linear comments
grammar, two centralized palette dispatch/source-selection boundaries, one model
constructor and one strict schema literal. Each retains an in-source rationale;
none bypasses the production signature or dependency policy.

## Seven dead symbols and minor findings

| Original symbol | Final disposition and evidence |
| --- | --- |
| `service.dumpDirectory` | Removed in Slice 4; actual dump service operations and command binding remain. |
| `service.newQuizModelRequest` | Removed in Slice 8; tests and both production consumers use `newQuizModelRequestWithEntries`. |
| `telegramGuidedResultExport` | Forwarding wrapper removed in Slice 7; the live guarded `guidedResultExport` remains. |
| `telegramHelpReply` | Forwarding wrapper removed in Slice 7; actual `renderTelegramHelp` remains. |
| `timePtr` | Exists only in `ticket/index_test.go`; no production declaration. |
| `WorkflowStore.prepareWorkflowTask` | Exists only in `workflow_execution_prepare_fixtures_test.go`; delegates to the real explicit automatic-mode path. |
| `withoutWorkflowPublicationProfile` | Removed in Slice 15 with its obsolete test-only production wrapper. |

All seven production declarations are absent in the current parsed source.
The [134 constructor map](code-quality-remediation-command-constructors.csv)
records typed factory injection and removal of dependency metadata lookup and
unused forwarding parameters (Slice 4, PR #441). The
[SQL row map](code-quality-remediation-sql-rows.csv) and Slice 3 report (PR #440)
cover fourteen wide binding/scan boundaries across eight tables, explicit column/bind/scan order,
schema, NULL/time/bool semantics and actual SQLite round-trip fixtures.

The secondary map resolves all eight error names. The last three were corrected
in `436f6742f7f224ac8c25bfde7cc61bb6ab328332`; constructors/callers changed without
aliases, while raw literals, unwrap/classification behavior and sentinel values
remain identical. Nineteen affected process/classification regressions passed.
The [cleanup map](code-quality-remediation-slice-16-cleanup.csv) reconciles the
reported 23 bare errcheck sites: eleven production, twelve test fixtures. Each
current ignored close has a resource-specific reason; authoritative write,
permission, broker and process errors remain observed. Current owned roots have
no bare errcheck or `nolint:unused` directives. The permanent checker has no
phase, ratchet or temporary debt file. Its only exact external exception is
bound to installed Bubble Tea v1.3.10 and compile-time interface satisfaction.

Typed workflow progress/resume, process and receipt states preserve stored
spellings (Slice 12); Telegram modes retain exact HTTP/output contracts (Slice 7).
Slice 5 shares ordered dispatch/key hints and presentation; Slice 15 removes
unused closing projection context and provisional forwarding. The final
[capability map](../../../docs/explanation/capability-boundaries.md) and
[glossary](../../../docs/reference/workflow-glossary.md) locate the real owners.
No broad terminology or schema migration is inferred from this refactor.

## Measured performance boundary

`go test ./internal/tui -run '^$' -bench '^BenchmarkModel(Update|View)$' -benchmem -benchtime=200ms -count=3`
passed in 3.689 seconds on the assessed implementation above. The
[final fifteen samples](code-quality-remediation-slice-16-benchmarks.csv)
preserve every observation and the actual measured revision. The unchanged
[Slice 5 baseline](code-quality-remediation-slice-5-benchmarks.csv) supplies the
same five cases and three samples per phase.

| Path | Baseline median ns/op | Final median ns/op | Baseline B/op | Final B/op |
| --- | ---: | ---: | ---: | ---: |
| Update idle | 51,532 | 18.61 | 229,376 | 0 |
| Update focus | 88,149 | 879.9 | 345,162 | 1,096 |
| Update resize | 41,288 | 15.27 | 172,032 | 0 |
| View plain | 225,533 | 118,033 | 435,290 | 435,292 |
| View shell | 692,490 | 452,316 | 551,454 | 550,216 |

The model remains 49,832 bytes; idle/resize allocate zero times and focus nine
times. View allocation counts remain 703/1,404. Latency and small byte variation
are host-sensitive observations. This establishes the measured model-copying
remedy, not a whole-application speed claim. The 4,374 heuristic hugeParam
warnings and general rangeValCopy signals do not establish other hot-path
defects; no bulk pointer conversion is justified by those warning counts.
