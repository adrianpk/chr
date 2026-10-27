package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowExecutionPolicy freezes finite execution limits at admission.
type WorkflowExecutionPolicy struct {
	AgentTokenBudget         int64  `json:"agent_token_budget,omitempty"`
	ConfigRevision           string `json:"config_revision,omitempty"`
	ValidationTimeoutSeconds int    `json:"validation_timeout_seconds,omitempty"`
	IncludeLaterTickets      bool   `json:"include_later_tickets,omitempty"`
	WorkerTimeoutSeconds     int    `json:"worker_timeout_seconds"`
	MaxWorkerAttempts        int    `json:"max_worker_attempts"`
	MaxIntegrationAttempts   int    `json:"max_integration_attempts,omitempty"`
	MaxMergerAttempts        int    `json:"max_merger_attempts,omitempty"`
	MaxDriftRevisions        int    `json:"max_drift_revisions,omitempty"`
	CheckpointStepLimit      int    `json:"checkpoint_step_limit,omitempty"`
}

// WorkflowStartRequest admits an explicit finite selection with a retry key.
type WorkflowStartRequest struct {
	HarnessBinding      *HarnessBinding         `json:"harness_binding,omitempty"`
	HarnessCommand      []string                `json:"-"`
	ResolveHarness      WorkflowHarnessResolver `json:"-"`
	Admission           *WorkflowAdmissionGuard `json:"admission_guard,omitempty"`
	IncludeLaterTickets *bool                   `json:"include_later_tickets_override,omitempty"`
	Selected            *WorkflowSelectionGuard `json:"-"`
	RequestKey          string
	Selection           WorkflowPreviewRequest
	Policy              WorkflowExecutionPolicy
	ReviewRestriction   ReviewWorkflowMode `json:"review_restriction,omitempty"`
}

// WorkflowMember retains admission facts and the latest accepted owned image.
type WorkflowMember struct {
	Admission          WorkflowPreviewMember `json:"admission"`
	ExpectedRevision   uint64                `json:"expected_revision"`
	ExpectedSourceHash string                `json:"expected_source_hash"`
}

// WorkflowInputImage freezes an owning set contract separately from its tasks.
type WorkflowInputImage struct {
	Path       string `json:"path"`
	Revision   uint64 `json:"revision"`
	SourceHash string `json:"source_hash"`
	UseCase    string `json:"use_case,omitempty"`
}

// WorkflowAttempt exposes stable ownership facts without private credentials.
type WorkflowAttempt struct {
	CapacitySlot   int                      `json:"capacity_slot,omitempty"`
	ID             string                   `json:"id"`
	Purpose        string                   `json:"purpose,omitempty"`
	TaskKey        string                   `json:"task_key"`
	Generation     uint64                   `json:"generation"`
	InputRevision  uint64                   `json:"input_revision"`
	ContractHash   string                   `json:"contract_hash"`
	LeaseUntil     time.Time                `json:"lease_until"`
	StartedAt      time.Time                `json:"started_at"`
	FinishedAt     time.Time                `json:"finished_at,omitempty"`
	Outcome        string                   `json:"outcome"`
	Lane           *WorkflowLane            `json:"lane,omitempty"`
	InputRef       string                   `json:"input_ref,omitempty"`
	InputHash      string                   `json:"input_hash,omitempty"`
	ResultRef      string                   `json:"result_ref,omitempty"`
	ResultHash     string                   `json:"result_hash,omitempty"`
	OwnerProcess   *WorkflowProcessIdentity `json:"owner_process,omitempty"`
	Processes      []WorkflowProcess        `json:"processes"`
	Replaces       string                   `json:"replaces"`
	CheckpointRef  string                   `json:"checkpoint_ref"`
	CheckpointHash string                   `json:"checkpoint_hash"`
}

// WorkflowInvocation is the public durable checkpoint, not an execution token.
type WorkflowInvocation struct {
	WorkflowInvocationAdmission
	WorkflowInvocationIntegration
	WorkflowInvocationScope
	WorkflowInvocationProgress
	WorkflowIdentity
	WorkflowInvocationInput
	WorkflowInvocationExecution
	WorkflowInvocationCandidate
	WorkflowInvocationDelivery
	WorkflowInvocationReview
	WorkflowInvocationPreparation
	Submission *WorkflowSubmissionBinding `json:"submission_binding,omitempty"`
}

// WorkflowInvocationAdmission groups the retained admission fields without changing their wire order.
type WorkflowInvocationAdmission struct {
	HarnessBinding *HarnessBinding         `json:"harness_binding,omitempty"`
	Admission      *WorkflowAdmissionGuard `json:"admission_guard,omitempty"`
}

// WorkflowInvocationIntegration groups the retained integration fields without changing their wire order.
type WorkflowInvocationIntegration struct {
	IntegrationBudget  *WorkflowIntegrationBudgetBinding `json:"integration_budget,omitempty"`
	Integration        *WorkflowIntegrationBinding       `json:"integration_binding,omitempty"`
	IntegrationEntries []WorkflowIntegrationBinding      `json:"integration_entries,omitempty"`
}

// WorkflowInvocationScope groups the retained scope fields without changing their wire order.
type WorkflowInvocationScope struct {
	DependencyTargets []WorkflowDependencyTarget  `json:"dependency_targets,omitempty"`
	LooseQueue        *WorkflowLooseQueue         `json:"loose_queue,omitempty"`
	Closures          []WorkflowInvocationClosure `json:"set_closures,omitempty"`
	ClosingSet        string                      `json:"closing_set,omitempty"`
	ExecutionTarget   *WorkflowPreviewTarget      `json:"execution_target,omitempty"`
}

// WorkflowInvocationProgress groups the retained progress fields without changing their wire order.
type WorkflowInvocationProgress struct {
	ConfirmedCursor int                             `json:"confirmed_cursor,omitempty"`
	Deliveries      []WorkflowUnitDelivery          `json:"integrated_units,omitempty"`
	Continuation    *WorkflowContinuationCheckpoint `json:"continuation_checkpoint,omitempty"`
}

// WorkflowInvocationInput groups the retained input fields without changing their wire order.
type WorkflowInvocationInput struct {
	Schema       string                  `json:"schema"`
	RequestKey   string                  `json:"request_key"`
	RequestHash  string                  `json:"request_hash"`
	Authority    string                  `json:"authority"`
	Selection    WorkflowPreviewRequest  `json:"selection"`
	Policy       WorkflowExecutionPolicy `json:"policy"`
	SetOrderRule string                  `json:"set_order_rule"`
	Members      []WorkflowMember        `json:"members"`
	SetImages    []WorkflowInputImage    `json:"set_images"`
	Excluded     map[string]int          `json:"excluded"`
}

// WorkflowInvocationExecution groups the retained execution fields without changing their wire order.
type WorkflowInvocationExecution struct {
	Cursor               int               `json:"cursor"`
	Status               string            `json:"status"`
	Outcome              string            `json:"outcome"`
	Attempts             []WorkflowAttempt `json:"attempts"`
	Blocker              *WorkflowBlocker  `json:"blocker,omitempty"`
	Lane                 *WorkflowLane     `json:"lane,omitempty"`
	StopIntent           string            `json:"stop_intent"`
	ResumeStatus         string            `json:"resume_status"`
	CapacityResumeStatus string            `json:"capacity_resume_status,omitempty"`
}

// WorkflowInvocationCandidate groups the retained candidate fields without changing their wire order.
type WorkflowInvocationCandidate struct {
	CandidateRef  string `json:"candidate_ref"`
	CandidateHash string `json:"candidate_hash"`
}

// WorkflowInvocationDelivery groups the retained delivery fields without changing their wire order.
type WorkflowInvocationDelivery struct {
	PullRequest       *WorkflowPullRequestBinding `json:"pull_request_binding,omitempty"`
	DeliveryProofRef  string                      `json:"delivery_proof_ref"`
	DeliveryProofHash string                      `json:"delivery_proof_hash"`
	DeliveryRef       string                      `json:"delivery_ref"`
	DeliveryHash      string                      `json:"delivery_hash"`
}

// WorkflowInvocationReview groups the retained review fields without changing their wire order.
type WorkflowInvocationReview struct {
	ReviewRestriction ReviewWorkflowMode           `json:"review_restriction,omitempty"`
	Review            *WorkflowReviewBinding       `json:"review_binding,omitempty"`
	ReviewCount       int                          `json:"review_count,omitempty"`
	ReviewRepairs     []WorkflowReviewRepairBudget `json:"review_repairs,omitempty"`
}

// WorkflowInvocationPreparation groups the retained preparation fields without changing their wire order.
type WorkflowInvocationPreparation struct {
	Preparation      *WorkflowPreparationBinding     `json:"preparation_binding,omitempty"`
	CandidateRepair  *WorkflowCandidateRepairBinding `json:"candidate_repair,omitempty"`
	PreparationCount int                             `json:"preparation_count,omitempty"`
}

type workflowInvocationFile struct {
	HarnessCommand []string `json:"harness_command,omitempty"`
	WorkflowInvocation
	OwnerRoot   string            `json:"owner_root"`
	Owners      map[string]string `json:"owner_token_hashes"`
	RequestKeys []string          `json:"request_keys"`
	sourceHash  string
	fields      map[string]json.RawMessage
}

// WorkflowStartResult distinguishes an accepted request from an attachment.
type WorkflowStartResult struct {
	Invocation WorkflowInvocation
	Attached   bool
}

// StartWorkflow freezes membership once; retries attach without launching work.
func (s *WorkflowStore) StartWorkflow(ctx context.Context, request WorkflowStartRequest) (result WorkflowStartResult, err error) {
	return s.StartWorkflowGuarded(ctx, request, nil)
}

// StartWorkflowGuarded performs the caller's last effect check under exclusion,
// after reconciling an original admission and before creating new native work.
func (s *WorkflowStore) StartWorkflowGuarded(ctx context.Context, request WorkflowStartRequest, check WorkflowAdmissionCheck) (result WorkflowStartResult, err error) {
	requestedPolicy := request.Policy

	request, err = normalizeWorkflowStart(request)
	if err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		return withWorkspaceConfigLock(ctx, s.layout.ProjectDir(), func() error {
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

func normalizeWorkflowStart(request WorkflowStartRequest) (WorkflowStartRequest, error) {
	if err := validateWorkflowStartGuards(request); err != nil {
		return request, err
	}

	if request.Admission != nil {
		request.Selection.SnapshotHash = request.Admission.SnapshotHash
	}

	if request.ReviewRestriction != "" && request.ReviewRestriction != ReviewWorkflowModeHumanReview {
		return request, fmt.Errorf("an invocation may restrict review to human_review only")
	}

	if !workflowstate.ValidRequestKey(request.RequestKey) {
		return request, fmt.Errorf("workflow request key must contain 1..128 plain bytes")
	}

	request.Selection.Offset, request.Selection.Limit = 0, workflowstate.WorkflowPageLimit

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

func validateWorkflowStartGuards(request WorkflowStartRequest) error {
	if err := validateWorkflowRequestedHarness(request.HarnessBinding, request.HarnessCommand); err != nil {
		return err
	}

	return validateWorkflowAdmissionGuard(request.Admission, request.Selection)
}

func normalizeWorkflowExecutionPolicy(policy WorkflowExecutionPolicy) WorkflowExecutionPolicy {
	if policy.WorkerTimeoutSeconds == 0 {
		policy.WorkerTimeoutSeconds = 3600
	}

	if policy.MaxWorkerAttempts == 0 {
		policy.MaxWorkerAttempts = 3
	}

	if policy.CheckpointStepLimit == workflowstate.WorkflowPageLimit {
		policy.CheckpointStepLimit = 0
	}

	return policy
}

func validWorkflowExecutionPolicy(policy WorkflowExecutionPolicy) bool {
	return policy.AgentTokenBudget >= 0 && validWorkflowWorkerPolicy(policy) && validWorkflowIntegrationPolicyBounds(policy) && policy.MaxIntegrationAttempts >= 0 && policy.MaxIntegrationAttempts <= 10 && policy.CheckpointStepLimit >= 0 && policy.CheckpointStepLimit <= workflowstate.WorkflowPageLimit && (policy.ConfigRevision == "" || len(policy.ConfigRevision) == 64)
}

func validWorkflowWorkerPolicy(policy WorkflowExecutionPolicy) bool {
	return policy.WorkerTimeoutSeconds >= 1 && policy.WorkerTimeoutSeconds <= 86400 && policy.MaxWorkerAttempts >= 1 && policy.MaxWorkerAttempts <= 128 && policy.ValidationTimeoutSeconds >= 0 && policy.ValidationTimeoutSeconds <= 86400
}

func workflowStartHash(request WorkflowStartRequest) (string, error) {
	// The resolved policy retains the override's value, so retries and durable
	// reconstruction bind that value rather than a transport-only pointer.
	request.IncludeLaterTickets = nil
	request.RequestKey = ""
	request.Selection.SnapshotHash = ""
	data, err := json.Marshal(request)

	return contenthash.Sum(data), err
}

func (s *WorkflowStore) startWorkflow(ctx context.Context, request WorkflowStartRequest, check WorkflowAdmissionCheck) (WorkflowStartResult, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return WorkflowStartResult{}, err
	}

	hash, err := workflowStartHash(request)
	if err != nil {
		return WorkflowStartResult{}, err
	}

	if previous, found, err := s.findWorkflowInvocation(runs, request.RequestKey, hash); found || err != nil {
		if found && err == nil {
			previous, err = s.attachWorkflowInvocation(previous, request.RequestKey)
		}

		return WorkflowStartResult{Invocation: previous.WorkflowInvocation, Attached: found}, err
	}

	return s.startNewWorkflow(ctx, request, hash, runs, check)
}

func (s *WorkflowStore) startNewWorkflow(ctx context.Context, request WorkflowStartRequest, hash string, runs []workflowInvocationFile, check WorkflowAdmissionCheck) (WorkflowStartResult, error) {
	if err := s.checkWorkflowAdmissionGuard(ctx, request); err != nil {
		return WorkflowStartResult{}, err
	}

	if check != nil {
		if err := check(ctx); err != nil {
			return WorkflowStartResult{}, err
		}
	}

	if err := s.checkGuidedWorkflowTarget(ctx, request); err != nil {
		return WorkflowStartResult{}, err
	}

	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return WorkflowStartResult{}, err
	}

	if _, err := s.checkWorkflowCoordinator(ctx); err != nil {
		return WorkflowStartResult{}, err
	}

	preview, err := s.freezeWorkflowSelection(ctx, request.Selection)
	if err != nil {
		return WorkflowStartResult{}, err
	}

	run, err := s.createWorkflowInvocation(ctx, request, hash, preview, runs, limits)

	return WorkflowStartResult{Invocation: run.WorkflowInvocation}, err
}

func (s *WorkflowStore) freezeWorkflowSelection(ctx context.Context, request WorkflowPreviewRequest) (WorkflowPreview, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return WorkflowPreview{}, err
	}

	all := WorkflowPreview{}

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

func (s *WorkflowStore) findWorkflowInvocation(runs []workflowInvocationFile, key, hash string) (workflowInvocationFile, bool, error) {
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

func (s *WorkflowStore) createWorkflowInvocation(ctx context.Context, request WorkflowStartRequest, hash string, preview WorkflowPreview, runs []workflowInvocationFile, limits workflowCapacityLimits) (run workflowInvocationFile, err error) {
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

	run = newWorkflowInvocation(s.layout.Root(), envelope.WorkflowIdentity, request, hash, preview)
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

func newWorkflowInvocation(root string, identity WorkflowIdentity, request WorkflowStartRequest, hash string, preview WorkflowPreview) workflowInvocationFile {
	invocation := WorkflowInvocation{WorkflowIdentity: identity, WorkflowInvocationInput: WorkflowInvocationInput{Schema: "chr-invocation/v1", RequestKey: request.RequestKey, RequestHash: hash, Authority: "tickets", Selection: request.Selection, Policy: request.Policy, SetOrderRule: preview.SetOrderRule, Excluded: preview.Excluded}, WorkflowInvocationExecution: WorkflowInvocationExecution{Status: "pending"}}

	invocation.HarnessBinding = cloneHarnessBinding(request.HarnessBinding)
	if invocation.HarnessBinding != nil {
		invocation.Schema = "chr-invocation/v2"
	}

	invocation.Admission = request.Admission
	invocation.ReviewRestriction = request.ReviewRestriction

	invocation.Selection.SnapshotHash = ""
	for _, member := range preview.Members {
		invocation.Members = append(invocation.Members, WorkflowMember{Admission: member, ExpectedRevision: member.Revision, ExpectedSourceHash: member.SourceHash})
	}

	if request.Selection.Selector.Extent == "loose" {
		invocation.LooseQueue = newWorkflowLooseQueue(invocation.Members)
	}

	return workflowInvocationFile{WorkflowInvocation: invocation, HarnessCommand: slices.Clone(request.HarnessCommand), OwnerRoot: root, Owners: map[string]string{}, RequestKeys: []string{request.RequestKey}}
}

func (s *WorkflowStore) attachWorkflowInvocation(run workflowInvocationFile, key string) (workflowInvocationFile, error) {
	if slices.Contains(run.RequestKeys, key) {
		return run, nil
	}

	if len(run.RequestKeys) >= workflowstate.WorkflowPageLimit {
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

	if err := owner.commitFiles([]workflowFileChange{change}); err != nil {
		return run, err
	}

	return owner.readWorkflowInvocation(run.ID)
}

func workflowInvocationCreationFiles(envelope WorkflowRecord, run workflowInvocationFile) ([]workflowFileChange, error) {
	fields, err := mergeWorkflowFields(envelope.Metadata, envelope.WorkflowIdentity)
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

	return []workflowFileChange{{Path: envelope.Path, After: append(metadata, '\n')}, {Path: filepath.ToSlash(filepath.Join(filepath.Dir(envelope.Path), "artifact.md")), After: []byte("# Workflow invocation\n")}, state}, nil
}

func workflowInvocationPath(id string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "workflow.json"))
}

func workflowInvocationChange(run workflowInvocationFile) (workflowFileChange, error) {
	if err := validateWorkflowInvocation(run); err != nil {
		return workflowFileChange{}, err
	}

	fields, err := workflowJSONFields(run)
	if err != nil {
		return workflowFileChange{}, err
	}

	for key, raw := range run.fields {
		if _, known := fields[key]; !known && !workflowOptionalInvocationField(key) {
			fields[key] = raw
		}
	}

	data, err := json.MarshalIndent(fields, "", "  ")

	return workflowFileChange{Path: workflowInvocationPath(run.ID), Before: run.sourceHash, After: append(data, '\n')}, err
}
