package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// WorkflowExecutionPolicy freezes finite execution limits at admission.
type WorkflowExecutionPolicy struct {
	WorkerTimeoutSeconds int `json:"worker_timeout_seconds"`
	MaxWorkerAttempts    int `json:"max_worker_attempts"`
}

// WorkflowStartRequest admits an explicit finite selection with a retry key.
type WorkflowStartRequest struct {
	RequestKey string
	Selection  WorkflowPreviewRequest
	Policy     WorkflowExecutionPolicy
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
	WorkflowIdentity
	Schema            string                      `json:"schema"`
	RequestKey        string                      `json:"request_key"`
	RequestHash       string                      `json:"request_hash"`
	Authority         string                      `json:"authority"`
	Selection         WorkflowPreviewRequest      `json:"selection"`
	Policy            WorkflowExecutionPolicy     `json:"policy"`
	SetOrderRule      string                      `json:"set_order_rule"`
	Members           []WorkflowMember            `json:"members"`
	SetImages         []WorkflowInputImage        `json:"set_images"`
	Excluded          map[string]int              `json:"excluded"`
	Cursor            int                         `json:"cursor"`
	Status            string                      `json:"status"`
	Outcome           string                      `json:"outcome"`
	Attempts          []WorkflowAttempt           `json:"attempts"`
	Blocker           *WorkflowBlocker            `json:"blocker,omitempty"`
	Lane              *WorkflowLane               `json:"lane,omitempty"`
	StopIntent        string                      `json:"stop_intent"`
	ResumeStatus      string                      `json:"resume_status"`
	CandidateRef      string                      `json:"candidate_ref"`
	CandidateHash     string                      `json:"candidate_hash"`
	PullRequest       *WorkflowPullRequestBinding `json:"pull_request_binding,omitempty"`
	DeliveryProofRef  string                      `json:"delivery_proof_ref"`
	DeliveryProofHash string                      `json:"delivery_proof_hash"`
	DeliveryRef       string                      `json:"delivery_ref"`
	DeliveryHash      string                      `json:"delivery_hash"`
}

type workflowInvocationFile struct {
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
	request, err = normalizeWorkflowStart(request)
	if err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		var startErr error

		result, startErr = s.startWorkflow(ctx, request)

		return startErr
	})

	return result, err
}

func normalizeWorkflowStart(request WorkflowStartRequest) (WorkflowStartRequest, error) {
	if !validWorkflowRequestKey(request.RequestKey) {
		return request, fmt.Errorf("workflow request key must contain 1..128 plain bytes")
	}

	request.Selection.Offset, request.Selection.Limit = 0, WorkflowPageLimit

	selection, err := normalizeWorkflowPreviewRequest(request.Selection)
	if err != nil {
		return request, err
	}

	request.Selection = selection

	if request.Policy.WorkerTimeoutSeconds == 0 {
		request.Policy.WorkerTimeoutSeconds = 3600
	}

	if request.Policy.MaxWorkerAttempts == 0 {
		request.Policy.MaxWorkerAttempts = 3
	}

	if !validWorkflowExecutionPolicy(request.Policy) {
		return request, fmt.Errorf("workflow worker policy bounds exceeded")
	}

	if request.Selection.Target.Repository == "" || !workflowCommitID(request.Selection.Target.Commit) {
		return request, fmt.Errorf("workflow execution requires a canonical repository and immutable target commit")
	}

	return request, validateWorkflowPreviewRequest(request.Selection)
}

func validWorkflowRequestKey(key string) bool {
	return strings.TrimSpace(key) == key && key != "" && len(key) <= 128 && !strings.ContainsAny(key, "\x00\n\r")
}

func validWorkflowExecutionPolicy(policy WorkflowExecutionPolicy) bool {
	return policy.WorkerTimeoutSeconds >= 1 && policy.WorkerTimeoutSeconds <= 86400 && policy.MaxWorkerAttempts >= 1 && policy.MaxWorkerAttempts <= 128
}

func workflowStartHash(request WorkflowStartRequest) (string, error) {
	request.RequestKey = ""
	request.Selection.SnapshotHash = ""
	data, err := json.Marshal(request)

	return workflowHash(data), err
}

func (s *WorkflowStore) startWorkflow(ctx context.Context, request WorkflowStartRequest) (WorkflowStartResult, error) {
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

	preview, err := s.freezeWorkflowSelection(ctx, request.Selection)
	if err != nil {
		return WorkflowStartResult{}, err
	}

	run, err := s.createWorkflowInvocation(ctx, request, hash, preview, runs)

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
			return all, ErrWorkflowConflict
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
			return run, false, fmt.Errorf("%w: request key names different frozen input", ErrWorkflowConflict)
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

func (s *WorkflowStore) createWorkflowInvocation(ctx context.Context, request WorkflowStartRequest, hash string, preview WorkflowPreview, runs []workflowInvocationFile) (run workflowInvocationFile, err error) {
	guard, err := s.acquireWorkflowAllocation()
	if err != nil {
		return run, err
	}
	defer func() { err = errors.Join(err, s.releaseWorkflowAllocation(guard)) }()

	if len(preview.Members) == 0 {
		return run, fmt.Errorf("selected workflow has no prepared unfinished tasks: %v", preview.Excluded)
	}

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

	files, err := workflowInvocationCreationFiles(envelope, run)
	if err != nil {
		return run, err
	}

	if err := s.commitAllocatedFiles(files, guard); err != nil {
		return run, err
	}

	return s.readWorkflowInvocation(run.ID)
}

func newWorkflowInvocation(root string, identity WorkflowIdentity, request WorkflowStartRequest, hash string, preview WorkflowPreview) workflowInvocationFile {
	invocation := WorkflowInvocation{WorkflowIdentity: identity, Schema: "chr-invocation/v1", RequestKey: request.RequestKey, RequestHash: hash, Authority: "tickets", Selection: request.Selection, Policy: request.Policy, SetOrderRule: preview.SetOrderRule, Excluded: preview.Excluded, Status: "pending"}

	invocation.Selection.SnapshotHash = ""
	for _, member := range preview.Members {
		invocation.Members = append(invocation.Members, WorkflowMember{Admission: member, ExpectedRevision: member.Revision, ExpectedSourceHash: member.SourceHash})
	}

	return workflowInvocationFile{WorkflowInvocation: invocation, OwnerRoot: root, Owners: map[string]string{}, RequestKeys: []string{request.RequestKey}}
}

func (s *WorkflowStore) attachWorkflowInvocation(run workflowInvocationFile, key string) (workflowInvocationFile, error) {
	if slices.Contains(run.RequestKeys, key) {
		return run, nil
	}

	if len(run.RequestKeys) >= WorkflowPageLimit {
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
		if _, known := fields[key]; !known && key != "blocker" {
			fields[key] = raw
		}
	}

	data, err := json.MarshalIndent(fields, "", "  ")

	return workflowFileChange{Path: workflowInvocationPath(run.ID), Before: run.sourceHash, After: append(data, '\n')}, err
}
