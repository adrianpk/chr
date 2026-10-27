# Slice 9: Authoring, document generation and guided preparation

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-9-quality-authoring`
PR: pending

## Purpose

Give authoring sessions, generated documents and guided preparation cohesive
owners while preserving artifact-specific approval, source and retry contracts.

## Delivered Behavior

Core/authoring/acceptance owns feature artifacts, static Gherkin checks, canonical
feature/round-note paths, rendering, answer retention and the existing writes.
Its application consumers import it directly. Artifact writing, clarification
notes and static checks have separate files. Quiz owns the original ScopeID path
rule; shared model payloads retain their existing generation owner.

Acceptance WriteResult preserves static rejection findings with no published
path, zero results on the original other error branches, and paths/checks after
success. Real rejection tests prove that failed checks create no ops envelope.
All 60 assigned tuple occurrences and three parameter-count occurrences are fixed;
26 five-parameter declarations retain their documented cohesive contracts.
The 89 assigned signature occurrences have exact dispositions. Debt decreases
from 301 to 237 with the same ceiling of seven and one external-interface
exception. Three required harness/parser counterparts from later owners changed
with their current Slice 9 consumers; no successor task is marked delivered.

Shared operational layout/envelopes, metadata scalars and bounded file/lock
primitives have explicit ops owners. The retained FileChange value belongs to
workflow/state; project-relative spelling and adjacent replacement belong to
project. Existing journal, transaction, permission, reclamation, release and
recovery decisions stay with their operation owners. Workflow/state also owns
nine pure control/evidence values retained by guided state. Native policy,
authority, receipt validation and store recovery remain in their existing owners.

Retained admission/selection values belong to workflow/state. Retained harness
identity belongs to core/harness and uses only the adapter-neutral capability
value. Ticket filtering and ordering stay native-owned. StartInput retains the
original serializable start fields; the native request embeds it and supplies
command, resolver and transient selection separately. Guided state stores only
StartInput, and app supplies its existing resolver at native admission. Frozen
guided-session, workflow-start and invocation bytes remain unchanged. The one-way
state-to-harness dependency has positive and cycle-rejection checks.

Core/authoring owns the complete retained session graph: 31 session, artifact,
provenance and guided types, their enum constants and six Valid methods. The
original session/reference and guided validators move with their values. Native
admission retains its narrower intent predicate; it does not substitute full
validation. Field order/tags, enum spelling, bounds, guard order and diagnostics
stay unchanged. Root aliases and forwarding validation APIs are absent.

Workspace owns the existing default-user spelling and lexical UserPath and
UserRelativePath operations. The original variadic append/join/slash conversion
sequence remains unchanged; initialization, authoring and native consumers use
the shared owner directly. Permission, layout, journal and recovery stay with
their existing operation owners.

GuidedWorkflowObserver supplies authoring-owned stop, original-admission and
verified-result observation ports. The existing native owner implements them
with its original bound reads, request lookup, proof inspection, locks, guards
and partial error outcomes. Stop acknowledgement remains distinct from native
quiescence; normal admission absence remains distinct from read failure. Result
projection reuses GuidedResultReference once, with the original native update
time. The existing authoring lock, guard/transition order and journal writes
remain in place. Complete repository ownership and preparation ports are pending.

D02 shares bounded concrete catalog traversal with typed native readers. D04
shares live payload/context preparation and D05 shares latest scoped index
selection. Final model request assembly serves spec, acceptance, observed plan
and legacy tracking paths. D07
shares scoped retained-step entry. D09 shares artifact metadata/directory/body
writes. D10 shares CLI revision/setup/output mechanics. Each retains its distinct
artifact rules, action/collision diagnostics, provenance, authority and receipts.

## Implementation Notes

Mechanical relocation, package naming/result conversion and the acceptance file
split are separate commits. The plan service separates API, generation, model
requests, publication, tracking and views. The move ledger retains 226 declaration
preimages with current names/files and extraction stages. The 60 tuple
dispositions, three parameter fixes, 26 cohesion dispositions and six duplicate
comparisons link exact source commits and validation. Original guided-session,
control and workflow-start wire images were frozen before the nine-value move.
Mechanical relocation and
package-qualified naming have separate commits, with no root aliases. The later
admission-value move retains 21 declaration/constant-spec preimages; StartInput
retains its original complete host-request declaration before field extraction.
All 47 keyed native request constructors use the shared value explicitly. The
session graph contributes 44 declaration preimages, validation contributes 12,
workspace paths three and native observation transitions three. The declaration
ledger now retains 226 relocation/extraction/transition preimages.

The initial scan of 101 assigned root authoring files identified 78 external
shared declarations. This is dependency evidence, not completion of the session
and repository move. Resolve remaining workflow/harness/ticket/context edges at
their real shared owner or consumer-owned port before moving those files.

Each duplicate comparison reads both governing contracts. Concrete spec and plan
catalogs keep their different image accounting and native readers. Guided
preparation and admission keep separate action, policy, host and receipt rules.
Contextual quiz and legacy quiz-to-draft keep separate content/provenance and
collision errors. Dump save and append keep separate typed mutations. Existing
assertions remain unchanged for each duplicate extraction. Plan lookup call sites
later adopt named outcomes while preserving their original assertions.

## Contracts Added or Changed

Plan results name stage output, request/response provenance, source/target
observations and expected absence. Guided results name committed input, runtime,
checkpoint and request/digest observations; original partial error values remain.
Dump and assisted-dump results name body/metadata/history images, mutation
results, retained bindings, session images and receipt absence. Partial failures,
image order and journal ownership remain unchanged. None of these result values
changes persisted schemas or hash inputs. Context results also name normalized
paths, document/source inventory and prior attempt artifacts. Plan lookup results
keep presence on errors, preventing matching failed reads from reallocating work.
CLI summaries, direct scoped model callbacks, native harness resolution,
frontmatter images, revision/order bytes, archive presence, spec preparation and
authoring/case receipts now have explicit shapes. Optional legacy frontmatter and
required native envelopes keep distinct presence rules and errors. Authoring
receipt errors keep decoded partial results; case receipt errors remain zero.
Continuation changes/events remain in the original journal with unchanged raw
capture omission and provenance. These changes preserve original hashes and
zero/empty outcomes without adding persisted fields.

Public acceptance names use Artifact, Paths, RoundNotes, Layout, Check and Write.
The spec remains authoritative and features remain supporting evidence. Original
static checks precede writing; round notes keep retained answers and prompts.

Operational metadata keeps quoting, list splitting/omission, optional/zero times,
UTC RFC3339, permissive fallback and diagnostics. File IO keeps default/explicit
bounds, modes, error text, temporary names, sync/cleanup and advisory lock retry.
FileChange retains its original JSON tags and field order. Lexical project-path
containment remains distinct from symlink permission checks and synced writes.

Catalog traversal keeps namespace and numeric-prefix selection, directory error
handling, 4096 inventory entries, aggregate 8 MiB, reference order, nil/empty/error
outcomes and final pending-journal inspection. Listing never adopts records.
Primitive APIs pass context explicitly. The guided helper preserves the original
bounded service context and namespace assignment at retained-step entry.

## Files of Interest

- `internal/core/authoring/`: session/guided value and validation files, native
  observation contracts and acceptance artifact/round/check files.
- `internal/core/guided_workflow_observer.go` and `workspace/paths.go`.
- `internal/core/ops/`, `ops/file/`, `ops/metadata/`, project path/file primitives
  and `internal/core/workflow/state/file_change.go`.
- `internal/app/service_document_model.go`, `service_document_view.go`, plan
  responsibility files and guided input/observation result types.
- `internal/core/authoring_catalog.go`, `authoring_artifact_step.go`,
  `internal/app/service_guided_work_step.go` and `dump_cli_update.go`.
- [Move evidence](../../../plan/code-quality-remediation-slice-9-moves.csv),
  [signature transitions](../../../plan/code-quality-remediation-slice-9-signatures.csv),
  [recorded dispositions](../../../plan/code-quality-remediation-slice-9-dispositions.csv),
  [duplicate comparisons](../../../plan/code-quality-remediation-slice-9-duplicates.csv),
  [current owners](../../../plan/code-quality-remediation-core-owners.csv) and
  [dependency ledger](../../../plan/code-quality-remediation-boundaries.md).

## Validation

The completed task checkpoints passed these focused real-boundary commands:

- `go test ./internal/core/ops/file ./internal/core/workflow/state ./scripts/goquality ./internal/core ./internal/app -run '^Test(LeafDependencies|Authoring|DumpMutation|SpecMutation|QuizAction|ProjectDump|WorkflowRoundTrip|WorkflowMaterialization|WorkflowRecoveryProcess)' -count=1 -timeout=3m`
- `go test ./internal/core/ops/metadata ./scripts/goquality ./internal/core ./internal/app -run '^Test(LeafDependencies|Ticket(Parser|Renderer)|Spec|Plan|DumpMutation|Authoring|ProjectMetadata|ProjectDump)' -count=1 -timeout=3m`
- `go test ./internal/core/authoring/acceptance ./scripts/goquality ./internal/core ./internal/app -run '^Test(Acceptance|ValidateAcceptance|CheckAcceptance|WriteAcceptance|MergeAcceptance|DecodeQuizModelFeature|LeafDependencies|Service(Acceptance|Spec|Plan)|QuizArtifact)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestGuidedUseCase(Preparation|Admission|RecoveryAdmission|ControlAdmission|TelegramScope)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(Authoring|PlanContextualQuiz|ValidateAuthoringQuiz|PlanQuizToDraft|WriteContextualQuiz|WriteQuizToDraft|AuthorContextual)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(DumpCLIWriteBinding|DumpNewAndShowUseServiceCommands|ServiceAppendAndOpenDump|ServiceDumpWriteGuards|ServiceDumpProjectionFailure|ServiceDumpCapture|ServiceProjectDump)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(SpecCatalog|PlanCatalog|SpecMutation|PlanPublication|ServiceSpec|ServicePlan)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(ServiceGenerate(Spec|AcceptanceFeature)|ServiceListSpecs|SpecDocumentCatalog|SpecGenerationScopes)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(Service(List(Specs|Plans)|Generate(Spec|Plan|AcceptanceFeature)|RegeneratePlan)|SpecDocumentCatalog|SpecGenerationScopes|PlanCatalog|PlanPublication)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(Service(GeneratePlan|RegeneratePlan|BuildsPlan|MaterializePlanTracking|CreatePlanTrackingTickets|ListPlans|Plan|Workflow)|Plan(Readiness|SpecApproval|GenerationDiscovery|Publication)|GuidedUseCasePreparation|Web.*(Plan|Case))' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(Service(Generate(Spec|AcceptanceFeature|Plan)|RegeneratePlan|BuildsPlan|MaterializePlanTracking|CreatePlanTrackingTickets)|SpecGenerationScopes|Plan(Readiness|SpecApproval|GenerationDiscovery|Publication)|GuidedUseCasePreparation)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestHarnessGenerationBinding(Authoring|Plan)$' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestGuidedUseCase(Quiz|SharedQuiz|Spec|Preparation|Recovery)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(DumpMutation|ProjectDump|Authoring.*Dump|AuthorAssistedDump|AssistedDump|Service(Dump|AppendAndOpenDump))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(DumpMutation|ProjectDump|Authoring.*Dump|AuthorAssistedDump|Assisted.*Dump|Service(Dump|AppendAndOpenDump))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/workflow/state ./internal/app ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|GuidedUseCase(Control|Result)|Workflow(Control|Integration(Grant|Decision|Budget)|Closing)|AgenticReview)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|WorkflowControlCAS|GuidedUseCase(ControlCAS|ControlUncertain|ControlAdmissionAuthority|ResultCorrupt)|WorkflowIntegrationGrant)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(Load(AgentProtocol|ProjectDocument|PreviousAttempt)Context|Render(AgentContext|ProtocolsContext|SourceInventory|PreviousAttempt)|WorkspaceContext|ComposeContext|MaterializeContext|ContextReadiness)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(PlanPublication|PlanOrderPublication|PlanDelivery|ServicePlan|GuidedUseCasePreparation|WebPlanDelivery)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(Author(ContextualQuiz|Assisted)|HarnessGenerationBindingAuthoring|Service(ListSpecs|ListPlans|GenerateSpec|GeneratePlan))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(QualityRecordWire|HarnessBinding|GuidedUseCase(Admission|Preparation))' -count=1 -timeout=3m`
- `go test ./internal/core/ops/metadata ./internal/core/workflow/state ./scripts/goquality ./internal/core ./internal/app -run '^Test(LeafDependencies|QualityRecordWire|Spec(Document|Mutation|Generation)|Plan(Catalog|Publication|Revision)|DumpMutation|ProjectDump|Workflow(RoundTrip|Adoption|Loose)|Ticket(Parser|Document|Board|Edit))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(ArchiveGeneratedPlan|Plan(Order|Publication|Delivery|Revision)|Spec(Mutation|Generation)|GuidedUseCasePreparation|Workflow(PlanOrder|Tracker|Materialization))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/telegram ./internal/app -run '^Test(Authoring|PlanAuthoringArtifact|Telegram(OperationCase|Authoring)|CaseMetadata|ServiceCase)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/telegram ./internal/app -run '^Test(Authoring|PlanAuthoringArtifact|LoadAssistedAuthoring|RenderAssistedAuthoring|CurateAssisted|AskNextAssisted|AnswerAssisted|TelegramAuthoring|AuthorAssisted|GuidedUseCase(Capture|Quiz|SharedQuiz))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(LoadGoGraphContext|PlanContextualQuizGeneration|WriteContextualQuiz|PlanQuizToDraft|WriteQuizToDraft|AuthorContextual|AuthorAssistedDump)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(TicketLoop|WorkflowLoose|WorkflowSelection|ServiceWork)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/harness ./internal/core/workflow/state ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|HarnessBinding|HarnessSelection|GuidedUseCase(Admission|Preparation|Capture)|Workflow(Config|Admission|ControlCAS|Selection|Loose)|TicketLoop|LegacyWorkObservation)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/harness ./internal/core/workflow/state ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|GuidedUseCase(Admission|Preparation|RecoveryAdmission|ControlAdmission|ControlCAS|Capture)|HarnessBinding|Workflow(Admission|ConfigAdmission|Selection|Start|Request))' -count=1 -timeout=3m`
- `go test ./scripts/goquality -run '^TestLeafDependencies$' -count=1`
- `go test ./internal/core ./internal/core/authoring ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|AuthoringSession|ValidateAuthoringSession|RenderAndParseAuthoringSession|GuidedUseCase(Capture|Admission|Control|Preparation|Result)|AuthorAssistedDump)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/authoring ./internal/app ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies|ValidateAuthoringSession|AuthoringSession|RenderAndParseAuthoringSession|SynthesizeAssisted|GuidedUseCase(Admission|Capture|Control)|GuidedAdmission)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/workspace ./internal/app -run '^Test(Workspace|User(Path|RelativePath)|CaseMetadata|ProjectMetadata|AuthoringSession|ProjectDump|LegacyWorkObservation)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/authoring ./internal/app -run '^Test(QualityRecordWire|GuidedUseCase(Admission|Control|Recovery|Result)|GuidedAdmission)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/core/authoring ./internal/app -run '^Test(QualityRecordWire|GuidedUseCaseResult)' -count=1 -timeout=3m`
- `make check-go-style` — passed with 237 exact debts and one external exception.
- `make check-package-readmes` — passed at the package extraction checkpoints.
- Pinned v2.14 strict lint passed for changed core/app, acceptance/quiz and checker
  owners at their completed checkpoints; the latest core/application command was
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/core/authoring ./internal/app`.
  The session move also passed strict TUI/checker lint and workspace paths passed
  strict workspace/TUI lint at their completed checkpoints.
- `git diff --check` — passed at task commits.

Initial compile/style diagnostics for moved allocator privacy, exact type debt
qualification, package naming, signature layout, direct-caller adaptation and
whitespace were corrected
without suppressions or relaxed checks. Original harness and session name prefixes were
retained in mechanical moves and corrected in separate naming commits. The
session relocation qualified nine enum map keys without changing their mapping.
Native result projection reuses the retained reference value, removing redundant
copy loops without a suppression. The
expanded dependency case table moved into a fixture helper to keep its test
within the existing function bound, preserving every case and assertion.
Final clean-head controller validation
has not run; this slice is not PR-ready.

## Risks and Follow-ups

Complete the session/repository move, workflow ports and remaining approved
Slice 9 tasks. Then prepare the canonical PR, commit/push its
references and record final ordered validation on its exact clean HEAD.
