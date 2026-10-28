package state

import (
	"time"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

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

// ExecutionSelection preserves admission identity while refreshing execution base.
func (run Invocation) ExecutionSelection() PreviewRequest {
	selection := run.Selection
	if run.ExecutionTarget != nil {
		selection.Target = *run.ExecutionTarget
	}

	return selection
}

// InvocationAdmission groups the retained admission fields without changing their wire order.
type InvocationAdmission struct {
	HarnessBinding *harnesscore.Binding `json:"harness_binding,omitempty"`
	Admission      *AdmissionGuard      `json:"admission_guard,omitempty"`
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

// InvocationExecution groups the retained execution fields without changing their wire order.
type InvocationExecution struct {
	Cursor               int              `json:"cursor"`
	Status               InvocationStatus `json:"status"`
	Outcome              string           `json:"outcome"`
	Attempts             []Attempt        `json:"attempts"`
	Blocker              *Blocker         `json:"blocker,omitempty"`
	Lane                 *Lane            `json:"lane,omitempty"`
	StopIntent           string           `json:"stop_intent"`
	ResumeStatus         InvocationStatus `json:"resume_status"`
	CapacityResumeStatus InvocationStatus `json:"capacity_resume_status,omitempty"`
}

// InvocationInput groups the retained input fields without changing their wire order.
type InvocationInput struct {
	Schema       string          `json:"schema"`
	RequestKey   string          `json:"request_key"`
	RequestHash  string          `json:"request_hash"`
	Authority    string          `json:"authority"`
	Selection    PreviewRequest  `json:"selection"`
	Policy       ExecutionPolicy `json:"policy"`
	SetOrderRule string          `json:"set_order_rule"`
	Members      []Member        `json:"members"`
	SetImages    []InputImage    `json:"set_images"`
	Excluded     map[string]int  `json:"excluded"`
}

// InvocationIntegration groups the retained integration fields without changing their wire order.
type InvocationIntegration struct {
	IntegrationBudget  *IntegrationBudgetBinding `json:"integration_budget,omitempty"`
	Integration        *IntegrationBinding       `json:"integration_binding,omitempty"`
	IntegrationEntries []IntegrationBinding      `json:"integration_entries,omitempty"`
}

// InvocationPreparation groups the retained preparation fields without changing their wire order.
type InvocationPreparation struct {
	Preparation      *PreparationBinding     `json:"preparation_binding,omitempty"`
	CandidateRepair  *CandidateRepairBinding `json:"candidate_repair,omitempty"`
	PreparationCount int                     `json:"preparation_count,omitempty"`
}

// InvocationProgress groups the retained progress fields without changing their wire order.
type InvocationProgress struct {
	ConfirmedCursor int                     `json:"confirmed_cursor,omitempty"`
	Deliveries      []UnitDelivery          `json:"integrated_units,omitempty"`
	Continuation    *ContinuationCheckpoint `json:"continuation_checkpoint,omitempty"`
}

// InvocationReview groups the retained review fields without changing their wire order.
type InvocationReview struct {
	ReviewRestriction ReviewMode           `json:"review_restriction,omitempty"`
	Review            *ReviewBinding       `json:"review_binding,omitempty"`
	ReviewCount       int                  `json:"review_count,omitempty"`
	ReviewRepairs     []ReviewRepairBudget `json:"review_repairs,omitempty"`
}

// InvocationScope groups the retained scope fields without changing their wire order.
type InvocationScope struct {
	DependencyTargets []DependencyTarget  `json:"dependency_targets,omitempty"`
	LooseQueue        *LooseQueue         `json:"loose_queue,omitempty"`
	Closures          []InvocationClosure `json:"set_closures,omitempty"`
	ClosingSet        string              `json:"closing_set,omitempty"`
	ExecutionTarget   *PreviewTarget      `json:"execution_target,omitempty"`
}
