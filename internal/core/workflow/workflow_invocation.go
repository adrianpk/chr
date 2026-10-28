package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// StartRequest supplies host-only inputs around the retained admission request.
type StartRequest struct {
	workflowstate.StartInput
	HarnessCommand []string                      `json:"-"`
	ResolveHarness HarnessResolver               `json:"-"`
	Selected       *workflowstate.SelectionGuard `json:"-"`
}

// Member retains admission facts and the latest accepted owned image.
type Member struct {
	Admission          PreviewMember `json:"admission"`
	ExpectedRevision   uint64        `json:"expected_revision"`
	ExpectedSourceHash string        `json:"expected_source_hash"`
}

// InputImage freezes an owning set contract separately from its tasks.
type InputImage struct {
	Path       string `json:"path"`
	Revision   uint64 `json:"revision"`
	SourceHash string `json:"source_hash"`
	UseCase    string `json:"use_case,omitempty"`
}

// Attempt exposes stable ownership facts without private credentials.
type Attempt struct {
	CapacitySlot   int              `json:"capacity_slot,omitempty"`
	ID             string           `json:"id"`
	Purpose        string           `json:"purpose,omitempty"`
	TaskKey        string           `json:"task_key"`
	Generation     uint64           `json:"generation"`
	InputRevision  uint64           `json:"input_revision"`
	ContractHash   string           `json:"contract_hash"`
	LeaseUntil     time.Time        `json:"lease_until"`
	StartedAt      time.Time        `json:"started_at"`
	FinishedAt     time.Time        `json:"finished_at,omitempty"`
	Outcome        string           `json:"outcome"`
	Lane           *Lane            `json:"lane,omitempty"`
	InputRef       string           `json:"input_ref,omitempty"`
	InputHash      string           `json:"input_hash,omitempty"`
	ResultRef      string           `json:"result_ref,omitempty"`
	ResultHash     string           `json:"result_hash,omitempty"`
	OwnerProcess   *ProcessIdentity `json:"owner_process,omitempty"`
	Processes      []Process        `json:"processes"`
	Replaces       string           `json:"replaces"`
	CheckpointRef  string           `json:"checkpoint_ref"`
	CheckpointHash string           `json:"checkpoint_hash"`
}

// Invocation is the public durable checkpoint, not an execution token.
type Invocation struct {
	InvocationAdmission
	InvocationIntegration
	InvocationScope
	InvocationProgress
	Identity
	InvocationInput
	InvocationExecution
	InvocationCandidate
	InvocationDelivery
	InvocationReview
	InvocationPreparation
	Submission *SubmissionBinding `json:"submission_binding,omitempty"`
}

// InvocationAdmission groups the retained admission fields without changing their wire order.
type InvocationAdmission struct {
	HarnessBinding *harnesscore.Binding          `json:"harness_binding,omitempty"`
	Admission      *workflowstate.AdmissionGuard `json:"admission_guard,omitempty"`
}

// InvocationIntegration groups the retained integration fields without changing their wire order.
type InvocationIntegration struct {
	IntegrationBudget  *IntegrationBudgetBinding `json:"integration_budget,omitempty"`
	Integration        *IntegrationBinding       `json:"integration_binding,omitempty"`
	IntegrationEntries []IntegrationBinding      `json:"integration_entries,omitempty"`
}

// InvocationScope groups the retained scope fields without changing their wire order.
type InvocationScope struct {
	DependencyTargets []DependencyTarget           `json:"dependency_targets,omitempty"`
	LooseQueue        *LooseQueue                  `json:"loose_queue,omitempty"`
	Closures          []InvocationClosure          `json:"set_closures,omitempty"`
	ClosingSet        string                       `json:"closing_set,omitempty"`
	ExecutionTarget   *workflowstate.PreviewTarget `json:"execution_target,omitempty"`
}

// InvocationProgress groups the retained progress fields without changing their wire order.
type InvocationProgress struct {
	ConfirmedCursor int                     `json:"confirmed_cursor,omitempty"`
	Deliveries      []UnitDelivery          `json:"integrated_units,omitempty"`
	Continuation    *ContinuationCheckpoint `json:"continuation_checkpoint,omitempty"`
}

// InvocationInput groups the retained input fields without changing their wire order.
type InvocationInput struct {
	Schema       string                        `json:"schema"`
	RequestKey   string                        `json:"request_key"`
	RequestHash  string                        `json:"request_hash"`
	Authority    string                        `json:"authority"`
	Selection    workflowstate.PreviewRequest  `json:"selection"`
	Policy       workflowstate.ExecutionPolicy `json:"policy"`
	SetOrderRule string                        `json:"set_order_rule"`
	Members      []Member                      `json:"members"`
	SetImages    []InputImage                  `json:"set_images"`
	Excluded     map[string]int                `json:"excluded"`
}

// InvocationExecution groups the retained execution fields without changing their wire order.
type InvocationExecution struct {
	Cursor               int       `json:"cursor"`
	Status               string    `json:"status"`
	Outcome              string    `json:"outcome"`
	Attempts             []Attempt `json:"attempts"`
	Blocker              *Blocker  `json:"blocker,omitempty"`
	Lane                 *Lane     `json:"lane,omitempty"`
	StopIntent           string    `json:"stop_intent"`
	ResumeStatus         string    `json:"resume_status"`
	CapacityResumeStatus string    `json:"capacity_resume_status,omitempty"`
}

// InvocationCandidate groups the retained candidate fields without changing their wire order.
type InvocationCandidate struct {
	CandidateRef  string `json:"candidate_ref"`
	CandidateHash string `json:"candidate_hash"`
}

// InvocationDelivery groups the retained delivery fields without changing their wire order.
type InvocationDelivery struct {
	PullRequest       *PullRequestBinding `json:"pull_request_binding,omitempty"`
	DeliveryProofRef  string              `json:"delivery_proof_ref"`
	DeliveryProofHash string              `json:"delivery_proof_hash"`
	DeliveryRef       string              `json:"delivery_ref"`
	DeliveryHash      string              `json:"delivery_hash"`
}

// InvocationReview groups the retained review fields without changing their wire order.
type InvocationReview struct {
	ReviewRestriction workflowstate.ReviewMode `json:"review_restriction,omitempty"`
	Review            *ReviewBinding           `json:"review_binding,omitempty"`
	ReviewCount       int                      `json:"review_count,omitempty"`
	ReviewRepairs     []ReviewRepairBudget     `json:"review_repairs,omitempty"`
}

// InvocationPreparation groups the retained preparation fields without changing their wire order.
type InvocationPreparation struct {
	Preparation      *PreparationBinding     `json:"preparation_binding,omitempty"`
	CandidateRepair  *CandidateRepairBinding `json:"candidate_repair,omitempty"`
	PreparationCount int                     `json:"preparation_count,omitempty"`
}

type workflowInvocationFile struct {
	HarnessCommand []string `json:"harness_command,omitempty"`
	Invocation
	OwnerRoot   string            `json:"owner_root"`
	Owners      map[string]string `json:"owner_token_hashes"`
	RequestKeys []string          `json:"request_keys"`
	sourceHash  string
	fields      map[string]json.RawMessage
}

// StartResult distinguishes an accepted request from an attachment.
type StartResult struct {
	Invocation Invocation
	Attached   bool
}

// StartWorkflow freezes membership once; retries attach without launching work.
func (s *Store) StartWorkflow(ctx context.Context, request StartRequest) (result StartResult, err error) {
	return s.StartWorkflowGuarded(ctx, request, nil)
}

// StartWorkflowGuarded performs the caller's last effect check under exclusion,
// after reconciling an original admission and before creating new native work.
func (s *Store) StartWorkflowGuarded(ctx context.Context, request StartRequest, check AdmissionCheck) (result StartResult, err error) {
	requestedPolicy := request.Policy

	request, err = normalizeWorkflowStart(request)
	if err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		return workspacecore.WithConfigLock(ctx, s.layout.ProjectDir(), func() error {
			policy, resolveErr := s.resolveWorkflowExecutionPolicy(ctx, request.RequestKey, requestedPolicy, request.IncludeLaterTickets)
			if resolveErr != nil {
				return resolveErr
			}

			request.Policy = policy

			request, resolveErr = s.resolveWorkflowHarnessRequest(ctx, request)
			if resolveErr != nil {
				return resolveErr
			}

			if request.Admission == nil {
				if err := s.checkWorkflowSelectionGuard(ctx, request); err != nil {
					return err
				}
			}

			var startErr error

			result, startErr = s.startWorkflow(ctx, request, check)

			return startErr
		})
	})

	return result, err
}

func normalizeWorkflowStart(request StartRequest) (StartRequest, error) {
	if err := validateWorkflowStartGuards(request); err != nil {
		return request, err
	}

	if request.Admission != nil {
		request.Selection.SnapshotHash = request.Admission.SnapshotHash
	}

	if request.ReviewRestriction != "" && request.ReviewRestriction != workflowstate.ReviewHuman {
		return request, fmt.Errorf("an invocation may restrict review to human_review only")
	}

	if !workflowstate.ValidRequestKey(request.RequestKey) {
		return request, fmt.Errorf("workflow request key must contain 1..128 plain bytes")
	}

	request.Selection.Offset, request.Selection.Limit = 0, workflowstate.PageLimit

	selection, err := normalizeWorkflowPreviewRequest(request.Selection)
	if err != nil {
		return request, err
	}

	request.Selection = selection

	request.Policy = normalizeWorkflowExecutionPolicy(request.Policy)

	if !validWorkflowExecutionPolicy(request.Policy) {
		return request, fmt.Errorf("workflow worker policy bounds exceeded")
	}

	if request.Selection.Target.Repository == "" || !workflowCommitID(request.Selection.Target.Commit) {
		return request, fmt.Errorf("workflow execution requires a canonical repository and immutable target commit")
	}

	return request, validateWorkflowPreviewRequest(request.Selection)
}

func validateWorkflowStartGuards(request StartRequest) error {
	if err := validateWorkflowRequestedHarness(request.HarnessBinding, request.HarnessCommand); err != nil {
		return err
	}

	return validateWorkflowAdmissionGuard(request.Admission, request.Selection)
}

func normalizeWorkflowExecutionPolicy(policy workflowstate.ExecutionPolicy) workflowstate.ExecutionPolicy {
	if policy.WorkerTimeoutSeconds == 0 {
		policy.WorkerTimeoutSeconds = 3600
	}

	if policy.MaxWorkerAttempts == 0 {
		policy.MaxWorkerAttempts = 3
	}

	if policy.CheckpointStepLimit == workflowstate.PageLimit {
		policy.CheckpointStepLimit = 0
	}

	return policy
}

func validWorkflowExecutionPolicy(policy workflowstate.ExecutionPolicy) bool {
	return policy.AgentTokenBudget >= 0 && validWorkflowWorkerPolicy(policy) && validWorkflowIntegrationPolicyBounds(policy) && policy.MaxIntegrationAttempts >= 0 && policy.MaxIntegrationAttempts <= 10 && policy.CheckpointStepLimit >= 0 && policy.CheckpointStepLimit <= workflowstate.PageLimit && (policy.ConfigRevision == "" || len(policy.ConfigRevision) == 64)
}

func validWorkflowWorkerPolicy(policy workflowstate.ExecutionPolicy) bool {
	return policy.WorkerTimeoutSeconds >= 1 && policy.WorkerTimeoutSeconds <= 86400 && policy.MaxWorkerAttempts >= 1 && policy.MaxWorkerAttempts <= 128 && policy.ValidationTimeoutSeconds >= 0 && policy.ValidationTimeoutSeconds <= 86400
}

func workflowStartHash(request StartRequest) (string, error) {
	// The resolved policy retains the override's value, so retries and durable
	// reconstruction bind that value rather than a transport-only pointer.
	request.IncludeLaterTickets = nil
	request.RequestKey = ""
	request.Selection.SnapshotHash = ""
	data, err := json.Marshal(request)

	return contenthash.Sum(data), err
}

func (s *Store) startWorkflow(ctx context.Context, request StartRequest, check AdmissionCheck) (StartResult, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return StartResult{}, err
	}

	hash, err := workflowStartHash(request)
	if err != nil {
		return StartResult{}, err
	}

	if previous, found, err := s.findWorkflowInvocation(runs, request.RequestKey, hash); found || err != nil {
		if found && err == nil {
			previous, err = s.attachWorkflowInvocation(previous, request.RequestKey)
		}

		return StartResult{Invocation: previous.Invocation, Attached: found}, err
	}

	return s.startNewWorkflow(ctx, request, hash, runs, check)
}

func (s *Store) startNewWorkflow(ctx context.Context, request StartRequest, hash string, runs []workflowInvocationFile, check AdmissionCheck) (StartResult, error) {
	if err := s.checkWorkflowAdmissionGuard(ctx, request); err != nil {
		return StartResult{}, err
	}

	if check != nil {
		if err := check(ctx); err != nil {
			return StartResult{}, err
		}
	}

	if err := s.checkGuidedWorkflowTarget(ctx, request); err != nil {
		return StartResult{}, err
	}

	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return StartResult{}, err
	}

	if _, err := s.checkWorkflowCoordinator(ctx); err != nil {
		return StartResult{}, err
	}

	preview, err := s.freezeWorkflowSelection(ctx, request.Selection)
	if err != nil {
		return StartResult{}, err
	}

	run, err := s.createWorkflowInvocation(ctx, request, hash, preview, runs, limits)

	return StartResult{Invocation: run.Invocation}, err
}

func (s *Store) freezeWorkflowSelection(ctx context.Context, request workflowstate.PreviewRequest) (Preview, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return Preview{}, err
	}

	all := Preview{}

	for {
		page, err := selection.preview(request)
		if err != nil {
			return all, err
		}

		if err := s.observePrerequisites(ctx, selection, request, &page); err != nil {
			return all, err
		}

		if request.SnapshotHash != "" && request.SnapshotHash != page.SnapshotHash {
			return all, workflowstate.ErrConflict
		}

		if request.Offset == 0 {
			all = page
		} else {
			all.Members = append(all.Members, page.Members...)
		}

		if page.NextOffset < 0 {
			return all, nil
		}

		request.Offset = page.NextOffset
	}
}

func (s *Store) findWorkflowInvocation(runs []workflowInvocationFile, key, hash string) (workflowInvocationFile, bool, error) {
	for _, run := range runs {
		if run.Namespace != s.layout.Namespace() {
			continue
		}

		matchedKey := slices.Contains(run.RequestKeys, key)
		if matchedKey && run.RequestHash != hash {
			return run, false, fmt.Errorf("%w: request key names different frozen input", workflowstate.ErrConflict)
		}

		if matchedKey || run.RequestHash == hash && !workflowInvocationTerminal(run.Status) {
			return run, true, nil
		}
	}

	return workflowInvocationFile{}, false, nil
}

func workflowInvocationTerminal(status string) bool {
	return status == "completed" || status == "cancelled"
}

func (s *Store) createWorkflowInvocation(ctx context.Context, request StartRequest, hash string, preview Preview, runs []workflowInvocationFile, limits workflowCapacityLimits) (run workflowInvocationFile, err error) {
	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return run, err
	}
	defer func() { err = errors.Join(err, s.releaseWorkflowAllocation(guard)) }()

	number, err := s.nextWorkflowInvocationNumber(ctx)
	if err != nil {
		return run, err
	}

	envelope, err := s.materializationEnvelope("run", "Workflow invocation", number, time.Now().UTC())
	if err != nil {
		return run, err
	}

	run = newWorkflowInvocation(s.layout.Root(), envelope.Identity, request, hash, preview)
	if err := s.freezeWorkflowSets(ctx, &run); err != nil {
		return run, err
	}

	if err := s.guardWorkflowAdmission(runs, run); err != nil {
		return run, err
	}

	if err := guardWorkflowReadyQueue(runs, run, limits); err != nil {
		return run, err
	}

	ticketFiles, err := s.workflowTicketAdmissionFiles(&run)
	if err != nil {
		return run, err
	}

	files, err := s.workflowAgentCreationFiles(ctx, envelope, &run)
	if err != nil {
		return run, err
	}

	if err := s.commitAllocatedFiles(append(files, ticketFiles...), guard); err != nil {
		return run, err
	}

	return s.readWorkflowInvocation(run.ID)
}

func newWorkflowInvocation(root string, identity Identity, request StartRequest, hash string, preview Preview) workflowInvocationFile {
	invocation := Invocation{Identity: identity, InvocationInput: InvocationInput{Schema: "chr-invocation/v1", RequestKey: request.RequestKey, RequestHash: hash, Authority: "tickets", Selection: request.Selection, Policy: request.Policy, SetOrderRule: preview.SetOrderRule, Excluded: preview.Excluded}, InvocationExecution: InvocationExecution{Status: "pending"}}

	invocation.HarnessBinding = harnesscore.CloneBinding(request.HarnessBinding)
	if invocation.HarnessBinding != nil {
		invocation.Schema = "chr-invocation/v2"
	}

	invocation.Admission = request.Admission
	invocation.ReviewRestriction = request.ReviewRestriction

	invocation.Selection.SnapshotHash = ""
	for _, member := range preview.Members {
		invocation.Members = append(invocation.Members, Member{Admission: member, ExpectedRevision: member.Revision, ExpectedSourceHash: member.SourceHash})
	}

	if request.Selection.Selector.Extent == "loose" {
		invocation.LooseQueue = newWorkflowLooseQueue(invocation.Members)
	}

	return workflowInvocationFile{Invocation: invocation, HarnessCommand: slices.Clone(request.HarnessCommand), OwnerRoot: root, Owners: map[string]string{}, RequestKeys: []string{request.RequestKey}}
}

func (s *Store) attachWorkflowInvocation(run workflowInvocationFile, key string) (workflowInvocationFile, error) {
	if slices.Contains(run.RequestKeys, key) {
		return run, nil
	}

	if len(run.RequestKeys) >= workflowstate.PageLimit {
		return run, fmt.Errorf("workflow request key attachment bound exceeded")
	}

	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return run, err
	}

	run.RequestKeys = append(run.RequestKeys, key)
	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	change, err := workflowInvocationChange(run)
	if err != nil {
		return run, err
	}

	if err := owner.commitFiles([]workflowstate.FileChange{change}); err != nil {
		return run, err
	}

	return owner.readWorkflowInvocation(run.ID)
}

func workflowInvocationCreationFiles(envelope Record, run workflowInvocationFile) ([]workflowstate.FileChange, error) {
	fields, err := mergeWorkflowFields(envelope.Metadata, envelope.Identity)
	if err != nil {
		return nil, err
	}

	metadata, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return nil, err
	}

	state, err := workflowInvocationChange(run)
	if err != nil {
		return nil, err
	}

	return []workflowstate.FileChange{{Path: envelope.Path, After: append(metadata, '\n')}, {Path: filepath.ToSlash(filepath.Join(filepath.Dir(envelope.Path), "artifact.md")), After: []byte("# Workflow invocation\n")}, state}, nil
}

func workflowInvocationPath(id string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "workflow.json"))
}

func workflowInvocationChange(run workflowInvocationFile) (workflowstate.FileChange, error) {
	if err := validateWorkflowInvocation(run); err != nil {
		return workflowstate.FileChange{}, err
	}

	fields, err := workflowJSONFields(run)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	for key, raw := range run.fields {
		if _, known := fields[key]; !known && !workflowOptionalInvocationField(key) {
			fields[key] = raw
		}
	}

	data, err := json.MarshalIndent(fields, "", "  ")

	return workflowstate.FileChange{Path: workflowInvocationPath(run.ID), Before: run.sourceHash, After: append(data, '\n')}, err
}
