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
	// WorkReadinessDiagnosticDirtyUserCheckout defines a package boundary constant.
	WorkReadinessDiagnosticDirtyUserCheckout WorkReadinessDiagnosticCode = "dirty_user_checkout"
	// WorkReadinessDiagnosticDirtyTicketWorktree defines a package boundary constant.
	WorkReadinessDiagnosticDirtyTicketWorktree WorkReadinessDiagnosticCode = "dirty_ticket_worktree"
	// WorkReadinessDiagnosticUnsafeTicketWorktree defines a package boundary constant.
	WorkReadinessDiagnosticUnsafeTicketWorktree WorkReadinessDiagnosticCode = "unsafe_ticket_worktree"
	// WorkReadinessDiagnosticOrphanedTicketWorktree defines a package boundary constant.
	WorkReadinessDiagnosticOrphanedTicketWorktree WorkReadinessDiagnosticCode = "orphaned_ticket_worktree"
	// WorkReadinessDiagnosticWorktreeRootMissing defines a package boundary constant.
	WorkReadinessDiagnosticWorktreeRootMissing WorkReadinessDiagnosticCode = "worktree_root_missing"
	// WorkReadinessDiagnosticStaleWorktreeMetadata defines a package boundary constant.
	WorkReadinessDiagnosticStaleWorktreeMetadata WorkReadinessDiagnosticCode = "stale_worktree_metadata"
	// WorkReadinessDiagnosticRuntimeArtifactsVisible defines a package boundary constant.
	WorkReadinessDiagnosticRuntimeArtifactsVisible WorkReadinessDiagnosticCode = "runtime_artifacts_visible"
	// WorkReadinessDiagnosticMissingRemoteRef defines a package boundary constant.
	WorkReadinessDiagnosticMissingRemoteRef WorkReadinessDiagnosticCode = "missing_remote_ref"
	// WorkReadinessDiagnosticRemoteBaseFetchFailed defines a package boundary constant.
	WorkReadinessDiagnosticRemoteBaseFetchFailed WorkReadinessDiagnosticCode = "remote_base_fetch_failed"
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
	// WorkReadinessRepairCreateWorktreeRoot defines a package boundary constant.
	WorkReadinessRepairCreateWorktreeRoot WorkReadinessRepairCode = "create_worktree_root"
	// WorkReadinessRepairPruneWorktreeMetadata defines a package boundary constant.
	WorkReadinessRepairPruneWorktreeMetadata WorkReadinessRepairCode = "prune_worktree_metadata"
	// WorkReadinessRepairRemoveOrphanedWorktree defines a package boundary constant.
	WorkReadinessRepairRemoveOrphanedWorktree WorkReadinessRepairCode = "remove_orphaned_worktree"
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
	ProjectDir     string
	Policy         WorkBranchPolicy
	Workspace      WorkWorkspaceConfig
	Status         WorkReadinessStatus
	Diagnostics    []WorkReadinessDiagnostic
	CurrentBranch  string
	TicketKey      string
	TicketBranch   string
	UserCheckout   WorkReadinessWorkspaceState
	TicketWorktree WorkReadinessWorkspaceState
}

// WorkReadinessWorkspaceState describes one workspace inspected by readiness.
type WorkReadinessWorkspaceState struct {
	Path        string
	State       WorktreeState
	Branch      string
	BaseRef     string
	BaseSHA     string
	StatusLines []string
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
	Path           string
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
