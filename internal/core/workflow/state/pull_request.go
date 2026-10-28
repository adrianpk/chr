package state

// PullRequestSnapshot contains provider facts, never worker completion claims.
type PullRequestSnapshot struct {
	URL            string `json:"url"`
	Repository     string `json:"repository"`
	HeadRepository string `json:"head_repository"`
	HeadBranch     string `json:"head_branch"`
	BaseBranch     string `json:"base_branch"`
	HeadCommit     string `json:"head_commit"`
	BaseCommit     string `json:"base_commit"`
	State          string `json:"state"`
	Merged         bool   `json:"merged"`
	ResultCommit   string `json:"result_commit"`
}
