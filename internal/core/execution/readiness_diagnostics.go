package execution

import (
	"fmt"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func addTicketWorktreeDiagnostic(report WorkReadinessReport, inspection WorktreeInspection) WorkReadinessReport {
	switch inspection.State {
	case WorktreeStateDirty:
		report.Diagnostics = append(report.Diagnostics, dirtyTicketWorktreeDiagnostic(inspection))
	case WorktreeStateMismatched, WorktreeStateBranchInUse:
		report.Diagnostics = append(report.Diagnostics, unsafeTicketWorktreeDiagnostic(inspection))
	case WorktreeStateOrphaned:
		report.Diagnostics = append(report.Diagnostics, orphanedTicketWorktreeDiagnostic(inspection))
	case WorktreeStateMissing:
		report = checkStaleWorktreeMetadata(report, inspection)
	}

	return report
}

func policyInvalidDiagnostic(err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticPolicyInvalid,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   err.Error(),
	}
}

func gitUnavailableDiagnostic(err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticGitUnavailable,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("git state is unavailable: %v", err),
	}
}

func gitStateInProgressDiagnostic(state string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticGitStateInProgress,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("git operation in progress: %s", state),
	}
}

func wrongBranchDiagnostic(expectedBranch, currentBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticWrongBranch,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("expected branch %s but current branch is %s", expectedBranch, currentBranch),
		CurrentBranch:  currentBranch,
		ExpectedBranch: expectedBranch,
	}
}

func missingBranchDiagnostic(expectedBranch, currentBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticMissingBranch,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("expected branch %s does not exist", expectedBranch),
		CurrentBranch:  currentBranch,
		ExpectedBranch: expectedBranch,
	}
}

func runtimeArtifactsVisibleDiagnostic(mode workspacecore.ExecutionMode, count int) WorkReadinessDiagnostic {
	severity := WorkReadinessSeverityBlocking
	if mode == workspacecore.ExecutionWorktree {
		severity = WorkReadinessSeverityInfo
	}

	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticRuntimeArtifactsVisible,
		Severity: severity,
		Detail:   fmt.Sprintf("chr runtime artifacts are visible to git: %d", count),
	}
}

func dirtyWorktreeDiagnostic(count int) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyWorktree,
		Severity: WorkReadinessSeverityBlocking,
		Detail:   fmt.Sprintf("worktree has user-visible changes: %d", count),
	}
}

func userCheckoutDirtyDiagnostic(mode workspacecore.ExecutionMode, count int) WorkReadinessDiagnostic {
	if mode != workspacecore.ExecutionWorktree {
		return dirtyWorktreeDiagnostic(count)
	}

	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyUserCheckout,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("user checkout has visible changes: %d", count),
	}
}

func dirtyTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticDirtyTicketWorktree,
		Severity: WorkReadinessSeverityBlocking,
		Detail: fmt.Sprintf(
			"ticket worktree %s has visible changes: %d",
			inspection.Plan.WorktreePath,
			len(inspection.StatusLines),
		),
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
	}
}

func unsafeTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	detail := fmt.Sprintf("ticket worktree %s is %s", inspection.Plan.WorktreePath, inspection.State)
	if len(inspection.Diagnostics) > 0 {
		detail = inspection.Diagnostics[0].Detail
	}

	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticUnsafeTicketWorktree,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         detail,
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
	}
}

func orphanedTicketWorktreeDiagnostic(inspection WorktreeInspection) WorkReadinessDiagnostic {
	detail := fmt.Sprintf("ticket worktree %s is orphaned", inspection.Plan.WorktreePath)
	if len(inspection.Diagnostics) > 0 {
		detail = inspection.Diagnostics[0].Detail
	}

	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticOrphanedTicketWorktree,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         detail,
		ExpectedBranch: inspection.Plan.Branch,
		RemoteRef:      inspection.Plan.BaseRef,
		Path:           inspection.Plan.WorktreePath,
	}
}

func worktreeRootMissingDiagnostic(path string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticWorktreeRootMissing,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("worktree root %s is missing", path),
		Path:     path,
	}
}

func staleWorktreeMetadataDiagnostic(path string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:     WorkReadinessDiagnosticStaleWorktreeMetadata,
		Severity: WorkReadinessSeverityInfo,
		Detail:   fmt.Sprintf("worktree metadata %s points to a missing worktree", path),
		Path:     path,
	}
}

func missingRemoteRefDiagnostic(remote, remoteRef, expectedBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticMissingRemoteRef,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("remote tracking ref %s is missing", remoteRef),
		ExpectedBranch: expectedBranch,
		Remote:         remote,
		RemoteRef:      remoteRef,
	}
}

func branchBehindRemoteDiagnostic(remote, remoteRef, expectedBranch string) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:           WorkReadinessDiagnosticBranchBehindRemote,
		Severity:       WorkReadinessSeverityBlocking,
		Detail:         fmt.Sprintf("branch %s does not contain %s", expectedBranch, remoteRef),
		ExpectedBranch: expectedBranch,
		Remote:         remote,
		RemoteRef:      remoteRef,
	}
}

func remoteBaseFetchFailedDiagnostic(remote, remoteRef string, err error) WorkReadinessDiagnostic {
	return WorkReadinessDiagnostic{
		Code:      WorkReadinessDiagnosticRemoteBaseFetchFailed,
		Severity:  WorkReadinessSeverityBlocking,
		Detail:    fmt.Sprintf("fetch %s failed before resolving %s: %v", remote, remoteRef, err),
		Remote:    remote,
		RemoteRef: remoteRef,
	}
}
