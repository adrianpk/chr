# Quality dependency and naming ledger

Owner: T1.5; refined by T2.3/T2.4 before package moves.
Baseline: `3a35bdf3445120cad3c47874d59712ef47227531`.
Plan: [Approved delivery](code-quality-remediation.md).

## Observed production graph

`go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/... ./cmd/...`
confirms these module-local edges at the Slice 6 implementation checkpoint. Standard-library and external dependencies
are omitted here; tests do not define production ownership.

| Consumer | Existing dependencies |
| --- | --- |
| cmd/chr | app |
| app | core, core/database, core/gate, core/notify, core/ops/namespace, core/project, core/telegram/message, core/workflow/state, harness, harness/codex, harness/pi, tui, webui, webui/ui |
| core | core/configvalue, core/contenthash, core/database, core/gate, core/notify, core/ops/namespace, core/project, core/telegram/message, core/workflow/state, harness, harness/process |
| core/configvalue | none |
| core/contenthash | none |
| core/database | none |
| core/gate | core/configvalue |
| core/notify | core/contenthash, core/database, core/ops/namespace, core/project, core/telegram/message, core/workflow/state |
| core/ops/namespace | none |
| core/project | core/database |
| core/telegram/message | core/contenthash, core/ops/namespace |
| core/workflow/state | core/contenthash |
| harness | none |
| harness/codex | harness, harness/process |
| harness/pi | harness, harness/process |
| harness/process | harness |
| tui | core, core/workflow/state |
| webui | webui/htmx, webui/render, webui/ui |
| webui/htmx | none |
| webui/render | webui/htmx, webui/ui |
| webui/ui | webui/htmx |

All names above are relative to internal/ except cmd/chr. The graph compiles
without cycles. The tooling command scripts/goquality depends only on the
standard library and has no application import edge.

## Shared values and move prerequisites

| Responsibility | Owner and allowed direction | Required before moving |
| --- | --- | --- |
| Adapter-neutral runner results and process execution | Existing internal/harness and harness/process; never import core workflow policy | Keep provider identity and project admission distinct in Slice 10 |
| Workspace configuration, user environment and neutral path/hash operations | Existing core declarations; retain only genuinely shared primitives at the core root | T2.3/T2.4 isolate dependencies before any leaf imports root core; root must not import that leaf or coordinator |
| Project identity and registry | core/project after Slice 6 | Move behavior with its config consumer ports; no import of a workflow stage |
| Ticket identity, indexing and status | core/ticket after Slice 11 | Separate readiness/execution from ticket values; preserve persisted spelling |
| Workflow invocation, claim and retained outcome values | core/workflow/state after Slice 12 | State contains values, not storage, dispatch, transport or provider clients |
| Durable filesystem operations, locks and journals | core/workflow persistence owner after Slice 12 | Store owns atomic operations over state; never expose a mutable untyped store or split a transaction across stage callers |
| Candidate preparation and merge policy | core/workflow/integrate and core/workflow/merge | Consumer-owned narrow ports use state values; store/coordinator supplies implementations |
| Provider publication and review | core/workflow/forge and core/workflow/review | Depend on state and provider contracts, not core/workflow coordinator |
| Closing decisions and receipts | core/workflow/closeout | Keep receipt/terminal distinctions and durable ownership |
| Sequencing and lifecycle | core/workflow | Imports stages and implements their ports; stages never import this coordinator |
| Reusable application operations | Existing central internal/app service | Adapter wrappers map input/results; service files stay free of TUI/Web types |

No move is authorized by a name alone. Each task first checks actual symbol
and import dependencies, resolves private helper ownership and updates callers
in the same compiling unit. The approved owner handles unresolved shared-value
cohesion before moving; exporting private internals or adding a generic common,
util or options package is not an acceptable cycle repair.

## Shared declarations resolved in Slice 2

| Declaration owner | Actual consumers and dependency decision |
| --- | --- |
| core/contenthash/hash.go | Slice 6 gives the existing workflowHash/contentHash SHA-256 implementation one neutral owner. Workflow, retained messages, guided checkpoints, authoring and configuration call Sum directly with the same bytes and encoding. No root forwarding facade remains. |
| core/project_paths.go | resolveProjectPath and pathIsInsideProject; runtime, context and workspace consumers retain absolute/relative containment behavior. This owner depends only on filepath and strings. |
| core/project_files.go | Existing writeFileAtomically; project/artifact consumers retain their own permission, locking and transaction decisions. This helper keeps its existing temporary-file behavior and is not substituted for guarded workflow writes. |
| core/config.go, config_workflow.go and capability config files | WorkspaceConfig and its configuration values stay in the existing configuration owner. Configuration parsing must not import the future workflow coordinator. Shared workflow state may be a dependency only when it has no dependency back to configuration or stages. |
| core/user_environment.go | Existing UserEnvironment boundary is reused; operation contexts remain explicit. No new global service, generic options object or stored context is introduced. |
| core/ops_layout.go and ops_named_artifact.go | Existing operational layout and new NamedOpsArtifactPath own absolute/project-relative body paths. Authoring callers keep their distinct validation, approval and publication operations. |

These are concrete declarations, not a request to export unrelated private
helpers. Before a consumer moves, its package receives only the values and
consumer-owned operations it actually needs. A stage cannot import its
coordinator to reach these helpers. At the Slice 2 checkpoint no package edge changed. The actual Slice 6 imports
above and enforced dependency contracts now govern the extracted owners.

## Domain names

| Term | Governing meaning and constraint |
| --- | --- |
| ticket-native | Canonical ticket-owned authority in spec/stable/ticket-native-workflow.md; not a synonym for provider publication |
| forge | Provider publication and PR interaction; never confirmation of delivery merely from an open PR |
| adoption | The existing ticket-native adoption contract; retain original operation identity and authority |
| frontier / loose selection | Existing eligible ticket selection rules; define the exact selected record before renaming helpers |
| continuation / amendment | Existing retained-work transitions; preserve revision, budget and request identity |
| closing | Existing closing gate and receipt-backed terminal decision; validation is not integration |

The original specifications retain their paths and wire terms. Package names
use plain nouns; package qualification replaces redundant prefixes only after
the mechanical move. The state and package maps above are dependency
constraints, not permission to rename an ambiguous domain concept.

## Root production-file dispositions

[The exact file ledger](code-quality-remediation-core-owners.csv) enumerates all
704 current non-test files under internal/core, their declared symbols,
existing signature owners, final responsibility and migration/retention task.
It provides a finite per-file handoff, including exact declarations. Multiple owners
on a workflow file require separating store-owned atomic operations from stage
policy in T12.1-T15.4; they do not allow moving locking into independent callers.
Signature conversion and package-move ownership can differ and are recorded
separately. This ledger is not evidence that the later moves are complete.

The previous proving.go is split into proving_objective.go,
proving_sandbox.go, proving_validation.go and proving_loop.go. Shared parsing
and exit-status interpretation live in frontmatter.go and process_result.go.
The moved function bodies and public contracts remain unchanged. Config parsing
stays cohesive in its existing owner; the shared frontmatter parser is small
and contains no approval or publication policy. Other reported large files and
schema literals keep their recorded Slice 8/9/11 owners.

[Slice 2 dispositions](code-quality-remediation-slice-2-dispositions.csv)
accounts for every assigned A/B/C row. All 15 distinct violating declarations
have named replacements; all eight legal five-parameter watch items have
specific cohesion reasons. Flags that represent parser outcome, required-source
diagnostics, paging or dry-run keep their existing meanings and finite bounds.
The initialization operation groups its own creation inputs; no generic bag
or unused root-file creation flag is introduced.

## Temporary bridges

None introduced in Slice 1. Any necessary bridge must identify its declaration,
consumer, owning removal task and validation before introduction. All bridges
expire by T15.4; permanent aliases or pass-through facades do not satisfy the
package split. Baseline debt reassociation requires explicit identity and
signature changes in the manifest, plus the original inventory reference.

## Command and ticket operation ownership in Slice 4

The typed commandFactory owns process IO and embedded assets for all 134
constructors. Commands parse flags/arguments and render outputs. The existing
central service owns shared ticket preflight, selection, execution, application,
legacy loop, review and merge operations in service_ticket files. Those stages
use explicit operation contexts and domain/core values rather than TUI/Web
presentation types. Manual, agentic and trusted review retain distinct governing
approval rules, and native/legacy authority guards remain at admission.

The mechanical move ledger preserves 292 complete declarations before semantic
extraction. The operation ledger identifies subsequent ownership changes; the
69-row final disposition ledger records every assigned signature occurrence.
No package edge, exported migration facade or generic service layer is added.
The existing daemonWorkflowRunner owns iteration resources, and adapter-specific
Web dispatch results remain in the Web adapter. Gate report metadata and
notification event identity are cohesive values at their current owners, ready
for their recorded Slice 6 capability moves.

## Slice 6 leaf prerequisites

Gate classification consumes loaded execution report identities, changed paths and
load diagnostics through its own report-source port. ReportLister implements the
storage mapping. Application composition owns ops artifact allocation/metadata and
supplies the explicit gate configuration file path. Gate retains validation, Git
evidence, classification and report encoding; it will not import root core.

Configuration literal interpretation has real gate and workspace consumers. The
existing string, boolean and non-empty assignment functions move unchanged into
core/configvalue, with one owner and no policy or IO. Root configuration keys
remain with their current owner. Gate's gitOutput has no free-function callers
outside gate: workflowProofReader.gitOutput is a distinct method, so the gate
implementation remains local and no shared Git package is introduced.

The project_dump prefix does not establish project-registry ownership. Its actual
receivers are fileAuthoringSessionRepository and its writes use authoring locks,
journals and receipts. These boundaries stay atomic during the project extraction;
T6.3 will record the exact dependency-backed owner rather than copy storage code.

### Gate boundary delivered in T6.1

core/gate imports only core/configvalue and the standard library. Root report
storage implements the gate-owned loaded-report port; app composition supplies
paths, artifact allocation and metadata refresh. There are no core aliases or
forwarding entrypoints for gate behavior. The mechanical move ledger proves all
125 production/test declaration texts unchanged before package-qualified renames.

The gate-owned parser, Git refs, Git range and execution-evidence collection now
return named values. Parser presence remains distinct from section/list state;
Git failures retain their original error context. setConfigValue consumes the
parsed line, and report constructors retain five coherent report inputs. Report
status and Woodpecker configuration outcomes have typed constants with the same
serialized spellings. Preview and overwrite flags retain the supported init
contract: DryRun takes precedence, and Write replaces only an existing owned file.

### Shared database dependency for T6.2/T6.3

Notification and project repositories use the same core.db schema, migration
assets, connection settings, timestamp representation and cancellation handling.
Those existing operations now have one core/database owner. Each repository
still owns its SQL, bound row values, transaction, quota reservation and commit.
Read-only registry/receiver inspection opens an existing database without migration.
No repository reaches another capability's private database field.

Database directory inputs use only DataDir; the root UserEnvironment remains the
host configuration boundary. The schema-status fixture explicitly constructs an
old database instead of calling a private migration helper. Existing migration,
real-SQLite record, notification quota/retention and registry regressions pass.

### Project identity boundary delivered in T6.3

core/project owns registered identities, the registry, active selection and registry
diagnostics. It imports core/database and no root core, adapter or coordinator.
DataDir and ConfigDir are separate consumer-owned directory capabilities. App
composition supplies the workspace directory name. Root host/workspace configuration
still owns the broader environment and configuration policies used by many domains.

Registry and ActiveStore are the actual implementations, with direct callers and
no re-export layer. Registry diagnostic status spellings, SQLite columns/order,
active-project file bytes and atomic replacement remain unchanged. The 39 moved
declarations have exact mechanical hashes before package-qualified name changes.
The indexed-ticket row reader now owns its one-method scanner port, consistent
with the existing notification and execution row ports.

RootBindingDigest hashes project ID, registered absolute root and resolved root,
separated by NUL bytes. Both registered and resolved paths remain in the digest.
Telegram conversation mapping retains its original safe diagnostic type and text
when the shared identity operation reports an unavailable root.

The project_dump files and project_metadata service are authoring behavior, not
registry behavior. Their actual fileAuthoringSessionRepository receivers use
shared authoring mutation locks, durable receipts and project-dump reconciliation.
Their final package owner is core/authoring in Slice 9; moving those operations
into the identity package would split that existing transactional owner. All
baseline signature rows assigned to Slice 6 remain assigned here and must be
resolved by T6.5; all are now fixed in the Slice 6 dispositions ledger. project_userspace remains shared workspace/operator configuration
at the core root. No helper implementation is copied to make the graph compile.

### Shared notification value dependencies

core/contenthash owns the existing SHA-256/hex content digest used by source and
retained records. The two identical root digest implementations now use that one
owner. core/ops/namespace owns the existing namespace spelling rule; layout and
transport still own selection and authorization. core/workflow/state owns the
single existing ErrConflict identity shared by conditional durable writes. It
contains no store or coordinator behavior. These dependencies let notification
storage preserve its errors and hashes without importing the root package.

Focused original record-byte, old SQLite, notification snapshot/quota/retention,
operational layout and workflow-mutation checks pass after these ownership changes.

### Telegram payload prerequisite for T6.2

core/telegram/message owns bounded blocks/entities, complete document values and
retained-presentation validation shared by notification and receiver consumers.
It imports only content hashes, namespace spelling and the standard library.
The 27 original declaration token streams are unchanged after explicit identifier
mapping; literals and tags are part of that proof. Frozen storage/presentation
checks and original Unicode/format fixtures pass. The acknowledgement's two
transport fields retain their types and meanings.

The receiver still maps its own persisted delivery state, while the pure decoder
checks hash, revision, rendering, text, entities and document binding once. Root
notification HTTP behavior and receiver mutation ownership remain with their
recorded transport owner. This prerequisite moves payload contracts only; no
receiver state, delivery claim, retry allowance or storage transaction is moved
into notification persistence.

### Notification storage move prerequisites

Notification configuration, document row mapping and delivery validation now live
beside notification state. The cross-table document quota predicate is one static
schema guard under core/database; each admission still executes its own insert
and quota reservation inside its original transaction. Root HTTP sending calls the
notification-owned validation contract and shares only the pure message codec.

Integration fixtures keep their owned SQL connection and call the public repository
factory. The retention fixture now triggers purge through real admission without
adding a retained image; all original twenty/four/twenty-two/twenty-four image,
receipt and hash assertions remain. Snapshot, changed-target, uncertainty, quota,
retention, old-database, rollback and application notification checks pass.

### Notification boundary delivered in T6.2

core/notify owns events/configuration, audit persistence, frozen Telegram lookup,
admission, conditional reserve/completion and bounded retained-document cleanup.
The mechanical commit preserves all 76 complete declarations before name and
signature changes. Direct callers use Repository, Event, Config and Telegram
operation values; no root re-export or migration facade remains.

TelegramLookup preserves delivery data and Frozen separately from errors, including
SQL absence, untouched legacy rows, malformed manifests and partial read results.
TelegramCompletion groups the event/target binding, transport acknowledgement,
safe failure code/uncertainty and completion time; context stays an operation
argument. Acknowledgement still clears failure/uncertainty, conditional updates
keep their original guards, and completion cannot grant another dispatch attempt.
Both locally owned interface signatures change with the implementation.

Typed event/status/source/severity constants retain their stored spellings. Read
row cleanup remains explicitly best effort after iteration errors are checked;
transaction rollback/commit failures remain observable. Receiver HTTP sending and
its storage keep their own owner, with notification validation over shared payload
values. Core/notify imports no root core, adapter or workflow coordinator.


### Complete Slice 6 contract and dependency dispositions

[All 19 assigned occurrences](code-quality-remediation-slice-6-dispositions.csv)
link original IDs to actual declarations, signatures, implementation commits and
focused validation. Sixteen are fixed; NewReport, NewFailedReport and the guarded
document lookup retain their cohesive five-input contracts. The six authoring
result conversions retain byte counts, body/metadata separation, partial change
results, normal receipt absence and parse-found outcomes. Receipt callbacks assign
the outer result/error under the existing authoring mutation lock; no receipt or
transaction ownership moves to project identity.

The existing production AST check now enforces every extracted owner in
scripts/goquality/boundaries.go. Imports are checked in generated files, every
build variant and descendants; root core, adapters and workflow coordination are
forbidden. Named imports, aliases and dot/blank imports have the same direction.
Test-only integration fixtures retain access to public consumer boundaries.
Standard-library facilities and existing external database drivers remain valid.

The glossary uses gate for Git/configured-validation evidence and report encoding;
project for registered identity/selection; notification for retained event and
conditional delivery ownership; message for immutable rendering contracts; and
state.ErrConflict for the single existing conditional-write conflict identity.
Notification pending/skipped/delivered/failed, event source/severity and registry
diagnostic spellings are unchanged. Gate report status and Woodpecker action
use named types with their original values. No public wire/status rename occurs.

ReserveTelegram's plain flag selects the original dispatch or the one definite
format fallback, guarded by original attempt/fallback predicates; it is not an
unbounded retry switch. Frozen, Found, Present, Uncertain and rendering/enablement
flags retain independent domain meanings. Existing read-only-close suppressions
remain documented next to checked iteration/scanner errors; owned error sentinels
have Err names and preserve their identities. No broad suppression or facade is
added. The root-only bindDump distinction remains part of authoring authority.

Package READMEs state the direct capabilities and dependencies. The finite file
ledger records current declarations, while baseline signature ownership remains
unchanged. The schema assets, original source review and detections are unchanged.
Global temporary debt is 410 with ceiling seven and the same single external
exception; no assigned Slice 6 declaration remains in that manifest.


## Slice 7 Telegram move prerequisites

The actual 70 root Telegram production files share workflow binding validation,
receipt reference derivation, authoring session identity, model error identities,
operational namespace resolution and directory access with other capabilities.
Those edges are resolved before the receiver/transport move. A root import or a
copy of private validators is not a permitted cycle repair.

workflow/state now owns MutationBinding, its original validation, ValidRequestKey
and MutationReceiptPath. The existing JSON field order/tags, validation errors,
request spelling and receipt hash/path are unchanged. Store safe-path checks,
receipt IO, locking and atomic updates remain with the workflow mutation owner.
Both receiver and workflow consumers use these actual value contracts directly.
The state import bound admits only contenthash for the canonical receipt digest;
it still rejects the coordinator, persistence and transport.


Telegram configuration/token reads consume only ConfigDir. OS exclusion and
model claims consume only StateDir. Receiver construction consumes DataDir and
StateDir; the live store retains only the state-directory capability needed for
retry exclusion. Read-only receiver/diagnostic inspection consumes the existing
database.DataDirectory and creates no lock/configuration ownership. The broader
UserEnvironment remains host composition and implements these narrow ports.
Original database paths, file permissions, lock names, descriptor lifetimes,
model slot limits and retry budgets are unchanged.


project.ResolveDirectory is the original shared directory validator, with unchanged
absolute resolution, filesystem checks and error text. Root workspace/layout and
Telegram callers use that same implementation. namespace.Resolve owns the original
explicit/OPS_USER/default choice, trimming and spelling errors; namespace.Default
and UserEnv have one owner, with no root re-export. Filesystem layout still owns
operational artifact paths. Both Telegram conversation allocation and target
resolution retain directory validation immediately before namespace choice, in
their original order after the project binding digest. No transaction boundary,
registry root digest, namespace precedence or environment setting changes.


### Telegram dispatch pair D06

The explanation and work responders now share only retained lookup, command
recognition, malformed-input diagnostics and current authority/project population.
Each constructs its own telegramCommandRoute with its original operation ID,
parser, help and resume callback. Its typed command remains captured locally;
the shared input pointer populates that same command before dispatch. Explanation
preparation/model/document behavior and native work preview/admission keep their
separate handlers, guards and confirmation contracts. Context is passed into each
operation; no context is stored in a reusable route.

The definition-only telegramHelpReply wrapper and telegramGuidedResultExport alias
are removed. Help tests still exercise the live renderer with real metadata;
the guided result guard uses the actual shared guidedResultExport type. All
existing assertions remain, including full result, original document and unsafe
export guards. Focused explanation/work/help and guided result-guard checks pass.


### Remaining receiver move edges at the Slice 7 checkpoint

The current declaration scan still finds the broad AuthoringSession input and
validateGuidedUseCase at guided association; shared authoring session identity,
intent and workspace directory spelling; the two structured-model error
identities; and generic remote-command record mapping. Resolve those edges at
their actual shared value or composition owner before moving receiver files.
Do not copy validators, import root core from Telegram, expose store internals or
leave pass-through facades. Preserve domain validation before binding a narrow
routing projection and preserve errors.Is identity for model capacity/output.
Legacy remote-update decoding and generic command mapping have separate owners;
verify real callers before retaining, relocating or removing either path.

The mechanical receiver move and all 156 assigned signature dispositions remain
pending. The focused task evidence does not substitute for the final clean-head
make verify and controller-recorded PR validation.
