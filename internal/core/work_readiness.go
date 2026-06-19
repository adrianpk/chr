package core

// WorkReadinessStatus defines a package boundary type.
type WorkReadinessStatus string

const (
	// WorkReadinessStatusReady defines a package boundary constant.
	WorkReadinessStatusReady WorkReadinessStatus = "ready"
	// WorkReadinessStatusBlocked defines a package boundary constant.
	WorkReadinessStatusBlocked WorkReadinessStatus = "blocked"

	// WorkReadinessSeverityInfo defines a package boundary constant.
	WorkReadinessSeverityInfo WorkReadinessSeverity = "info"
	// WorkReadinessSeverityWarning defines a package boundary constant.
	WorkReadinessSeverityWarning WorkReadinessSeverity = "warning"
	// WorkReadinessSeverityBlocking defines a package boundary constant.
	WorkReadinessSeverityBlocking WorkReadinessSeverity = "blocking"

	// WorkReadinessDiagnosticWrongBranch defines a package boundary constant.
	WorkReadinessDiagnosticWrongBranch WorkReadinessDiagnosticCode = "wrong_branch"
	// WorkReadinessDiagnosticMissingBranch defines a package boundary constant.
	WorkReadinessDiagnosticMissingBranch WorkReadinessDiagnosticCode = "missing_branch"
	// WorkReadinessDiagnosticDirtyWorktree defines a package boundary constant.
	WorkReadinessDiagnosticDirtyWorktree WorkReadinessDiagnosticCode = "dirty_worktree"
	// WorkReadinessDiagnosticRuntimeArtifactsVisible defines a package boundary constant.
	WorkReadinessDiagnosticRuntimeArtifactsVisible WorkReadinessDiagnosticCode = "runtime_artifacts_visible"
	// WorkReadinessDiagnosticMissingRemoteRef defines a package boundary constant.
	WorkReadinessDiagnosticMissingRemoteRef WorkReadinessDiagnosticCode = "missing_remote_ref"
	// WorkReadinessDiagnosticBranchBehindRemote defines a package boundary constant.
	WorkReadinessDiagnosticBranchBehindRemote WorkReadinessDiagnosticCode = "branch_behind_remote"
	// WorkReadinessDiagnosticGitStateInProgress defines a package boundary constant.
	WorkReadinessDiagnosticGitStateInProgress WorkReadinessDiagnosticCode = "git_state_in_progress"
	// WorkReadinessDiagnosticGitUnavailable defines a package boundary constant.
	WorkReadinessDiagnosticGitUnavailable WorkReadinessDiagnosticCode = "git_unavailable"
	// WorkReadinessDiagnosticPolicyInvalid defines a package boundary constant.
	WorkReadinessDiagnosticPolicyInvalid WorkReadinessDiagnosticCode = "policy_invalid"

	// WorkReadinessRepairIgnoreRuntimeArtifacts defines a package boundary constant.
	WorkReadinessRepairIgnoreRuntimeArtifacts WorkReadinessRepairCode = "ignore_runtime_artifacts"
	// WorkReadinessRepairFetchRemote defines a package boundary constant.
	WorkReadinessRepairFetchRemote WorkReadinessRepairCode = "fetch_remote"
	// WorkReadinessRepairFastForwardBranch defines a package boundary constant.
	WorkReadinessRepairFastForwardBranch WorkReadinessRepairCode = "fast_forward_branch"
	// WorkReadinessRepairSwitchBranch defines a package boundary constant.
	WorkReadinessRepairSwitchBranch WorkReadinessRepairCode = "switch_branch"
	// WorkReadinessRepairCreateBranch defines a package boundary constant.
	WorkReadinessRepairCreateBranch WorkReadinessRepairCode = "create_branch"
	// WorkReadinessRepairManual defines a package boundary constant.
	WorkReadinessRepairManual WorkReadinessRepairCode = "manual"
)

// WorkReadinessSeverity defines a package boundary type.
type WorkReadinessSeverity string

// WorkReadinessDiagnosticCode defines a package boundary type.
type WorkReadinessDiagnosticCode string

// WorkReadinessRepairCode defines a package boundary type.
type WorkReadinessRepairCode string

// WorkReadinessReport describes a package boundary value.
type WorkReadinessReport struct {
	ProjectDir    string
	Policy        WorkBranchPolicy
	Workspace     WorkWorkspaceConfig
	Status        WorkReadinessStatus
	Diagnostics   []WorkReadinessDiagnostic
	CurrentBranch string
	TicketKey     string
	TicketBranch  string
}

// WorkReadinessDiagnostic describes a package boundary value.
type WorkReadinessDiagnostic struct {
	Code           WorkReadinessDiagnosticCode
	Severity       WorkReadinessSeverity
	Detail         string
	CurrentBranch  string
	ExpectedBranch string
	Remote         string
	RemoteRef      string
	Repair         WorkReadinessRepair
}

// WorkReadinessRepair describes a package boundary value.
type WorkReadinessRepair struct {
	Available bool
	Code      WorkReadinessRepairCode
	Safe      bool
	Detail    string
}

// WorkReadinessRepairResult describes a package boundary value.
type WorkReadinessRepairResult struct {
	Before  WorkReadinessReport
	After   WorkReadinessReport
	Applied []WorkReadinessAppliedRepair
}

// WorkReadinessAppliedRepair describes a package boundary value.
type WorkReadinessAppliedRepair struct {
	Code   WorkReadinessRepairCode
	Detail string
}
