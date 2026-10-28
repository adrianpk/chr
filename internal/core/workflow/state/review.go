package state

// ReviewInput binds a decision to host validation and canonical inputs.
type ReviewInput struct {
	Schema          string              `json:"schema"`
	InvocationID    string              `json:"invocation_id"`
	Unit            Ref                 `json:"unit"`
	CandidateID     string              `json:"candidate_id"`
	CandidateRef    string              `json:"candidate_ref"`
	CandidateHash   string              `json:"candidate_hash"`
	Slice           SliceInput          `json:"slice"`
	PullRequest     PullRequestSnapshot `json:"pull_request"`
	TargetCommit    string              `json:"target_commit"`
	Policy          ReviewPolicy        `json:"policy"`
	PreparationRef  string              `json:"preparation_ref,omitempty"`
	PreparationHash string              `json:"preparation_hash,omitempty"`
	PreparedGit     *IntegrationProof   `json:"prepared_git,omitempty"`
	Native          *NativeReviewInput  `json:"native,omitempty"`
}

// ReviewDecision is returned by the distinct supervised review role.
// Its binding hash must name the exact input supplied on stdin.
type ReviewDecision struct {
	BindingHash   string              `json:"binding_hash"`
	Result        AgenticReviewResult `json:"result"`
	RepairTaskKey string              `json:"repair_task_key,omitempty"`
}

// ReviewBinding retains one current decision and immutable evidence.
type ReviewBinding struct {
	Round         int         `json:"round"`
	Unit          Ref         `json:"unit"`
	State         ReviewState `json:"state"`
	InputRef      string      `json:"input_ref"`
	InputHash     string      `json:"input_hash"`
	ResultRef     string      `json:"result_ref"`
	ResultHash    string      `json:"result_hash"`
	DecisionCount int         `json:"decision_count"`
	RepairTaskKey string      `json:"repair_task_key"`
}
