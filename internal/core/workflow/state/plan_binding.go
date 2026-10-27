package state

// WorkflowSlice is a set-owned descriptor; tickets alone own task membership.
type WorkflowSlice struct {
	ID            string `json:"id"`
	Order         int    `json:"slice_order"`
	Title         string `json:"title"`
	ContractRef   string `json:"contract_ref"`
	ContractHash  string `json:"contract_hash"`
	TaskInputHash string `json:"task_input_hash"`
	DependsOn     []Ref  `json:"depends_on,omitempty"`
	Phase         string `json:"phase"`
	Branch        string `json:"branch,omitempty"`
	CandidateRef  string `json:"candidate_ref"`
	PullRequest   string `json:"pull_request,omitempty"`
	DeliveryRef   string `json:"delivery_ref,omitempty"`
}

// WorkflowSet owns ordered slices and its closing contract.
type WorkflowSet struct {
	PlanRevision      *PlanRevisionBinding    `json:"plan_revision,omitempty"`
	Closing           *WorkflowClosingBinding `json:"closing,omitempty"`
	Amendments        []WorkflowAmendmentRef  `json:"amendments,omitempty"`
	Schema            string                  `json:"workflow_schema"`
	UseCase           string                  `json:"use_case,omitempty"`
	UseCaseOrder      int                     `json:"use_case_order,omitempty"`
	Status            string                  `json:"status"`
	PlanRef           string                  `json:"plan_ref"`
	PlanHash          string                  `json:"plan_hash"`
	Slices            []WorkflowSlice         `json:"slices"`
	ClosingValidation Validation              `json:"closing_validation"`
	ClosureRef        string                  `json:"closure_ref,omitempty"`
	ClosureHistory    []string                `json:"closure_history,omitempty"`
}

// WorkflowClosingBinding retains bounded attempt history across explicit reopening.
// Each set invocation keeps its own retry budget and immutable evidence.
type WorkflowClosingBinding struct {
	Attempts []ClosingEvidence `json:"attempts"`
}

// WorkflowAmendmentRef binds an immutable approved amendment to its owning set.
type WorkflowAmendmentRef struct {
	Revision uint64 `json:"revision"`
	Ref      string `json:"ref"`
	Hash     string `json:"hash"`
}

// PlanRevisionBinding identifies retained source and companion images. Hashing
// the plan body independently prevents circular plan/tracker dependencies.
type PlanRevisionBinding struct {
	PlannedOrderHash   string `json:"planned_order_hash"`
	PlannedTrackerHash string `json:"planned_tracker_hash"`
	Reference          string `json:"plan_ref"`
	Revision           uint64 `json:"plan_revision"`
	BodyHash           string `json:"plan_body_hash"`
	BodyRef            string `json:"body_ref"`
	OrderRef           string `json:"order_ref"`
	OrderHash          string `json:"order_hash"`
	TrackerRef         string `json:"tracker_ref"`
	TrackerHash        string `json:"tracker_hash"`
	CurrentTrackerRef  string `json:"current_tracker_ref"`
}
