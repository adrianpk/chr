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
Twenty assigned tuple occurrences are fixed; debt decreases from 301 to 281, with
the same ceiling of seven and one external-interface exception. Other assigned
signature occurrences still require their recorded dispositions.

Shared operational layout/envelopes, metadata scalars and bounded file/lock
primitives have explicit ops owners. The retained FileChange value belongs to
workflow/state; project-relative spelling and adjacent replacement belong to
project. Existing journal, transaction, permission, reclamation, release and
recovery decisions stay with their operation owners.

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
requests, publication, tracking and views. The move ledger retains 130 declaration
preimages with current names/files and extraction stages. Twenty tuple
dispositions, five cohesion reviews and six duplicate comparisons link exact
source commits and validation.

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
changes persisted schemas or hash inputs.

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

- `internal/core/authoring/acceptance/`: artifact.go, round.go and check.go.
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
- `make check-go-style` — passed with 281 exact debts and one external exception.
- `make check-package-readmes` — passed at the package extraction checkpoints.
- Pinned v2.14 strict lint passed for changed core/app, acceptance/quiz and checker
  owners at their completed checkpoints; the latest core/application command was
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`.
- `git diff --check` — passed at task commits.

Initial compile/style diagnostics for moved allocator privacy, exact type debt
qualification, package naming, signature layout, direct-caller adaptation and
whitespace were corrected
without suppressions or relaxed checks. Final clean-head controller validation
has not run; this slice is not PR-ready.

## Risks and Follow-ups

Complete the session/repository move and workflow ports, remaining assigned
signatures and the other approved Slice 9 tasks. Then prepare the canonical PR, commit/push its
references and record final ordered validation on its exact clean HEAD.
