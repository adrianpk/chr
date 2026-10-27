# Quality dependency and naming ledger

Owner: T1.5; refined by T2.3/T2.4 before package moves.
Baseline: `3a35bdf3445120cad3c47874d59712ef47227531`.
Plan: [Approved delivery](code-quality-remediation.md).

## Observed production graph

`go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/... ./cmd/...`
confirms these module-local edges. Standard-library and external dependencies
are omitted here; tests do not define production ownership.

| Consumer | Existing dependencies |
| --- | --- |
| cmd/chr | app |
| app | core, harness, harness/codex, harness/pi, tui, webui, webui/ui |
| core | harness, harness/process |
| harness | none |
| harness/codex | harness, harness/process |
| harness/pi | harness, harness/process |
| harness/process | harness |
| tui | core |
| webui | webui/htmx, webui/render, webui/ui |
| webui/render | webui/htmx, webui/ui |
| webui/ui | webui/htmx |
| webui/htmx | none |

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
| core/content_hash.go | Existing workflowHash SHA-256 implementation; workflow, guided checkpoints, authoring and configuration use the same bytes and encoding. It has no storage/coordinator dependency. The existing name and calls remain intact; there is no forwarding facade. |
| core/project_paths.go | resolveProjectPath and pathIsInsideProject; runtime, context and workspace consumers retain absolute/relative containment behavior. This owner depends only on filepath and strings. |
| core/project_files.go | Existing writeFileAtomically; project/artifact consumers retain their own permission, locking and transaction decisions. This helper keeps its existing temporary-file behavior and is not substituted for guarded workflow writes. |
| core/config.go, config_workflow.go and capability config files | WorkspaceConfig and its configuration values stay in the existing configuration owner. Configuration parsing must not import the future workflow coordinator. Shared workflow state may be a dependency only when it has no dependency back to configuration or stages. |
| core/user_environment.go | Existing UserEnvironment boundary is reused; operation contexts remain explicit. No new global service, generic options object or stored context is introduced. |
| core/ops_layout.go and ops_named_artifact.go | Existing operational layout and new NamedOpsArtifactPath own absolute/project-relative body paths. Authoring callers keep their distinct validation, approval and publication operations. |

These are concrete declarations, not a request to export unrelated private
helpers. Before a consumer moves, its package receives only the values and
consumer-owned operations it actually needs. A stage cannot import its
coordinator to reach these helpers. Current core has no new import edge and
the extracted functions are unchanged, so no new package cycle is introduced.

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
678 current non-test files directly under internal/core, their declared symbols,
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
resolved by T6.5. project_userspace remains shared workspace/operator configuration
at the core root. No helper implementation is copied to make the graph compile.
