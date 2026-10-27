package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// TicketMergeCommandRemote defines a package boundary interface.
type TicketMergeCommandRemote interface {
	TicketMergePreflightRemoteReader
	Merge(context.Context, MergeRequest) (MergeResult, error)
}

// TicketMergeCommandOptions describes a package boundary value.
type TicketMergeCommandOptions struct {
	Ticket        IndexedTicket
	Tickets       []IndexedTicket
	Remote        TicketMergeCommandRemote
	ExpectedBase  string
	MergeStrategy string
	ProjectDir    string
	BranchPolicy  WorkBranchPolicy
	Workspace     workspacecore.ExecutionConfig
	WorktreeGit   WorktreeGit
	DryRun        bool
	MergedAt      time.Time
}

// TicketMergeCommandResult describes a package boundary value.
type TicketMergeCommandResult struct {
	TicketKey       string
	Preflight       TicketMergePreflightResult
	Merge           MergeResult
	Status          string
	MergeState      string
	MergeResult     string
	MergeActor      string
	MergeStrategy   string
	MergeCommit     string
	MergeBlocker    string
	WorktreeCleanup WorktreeCleanupResult
	DryRun          bool
}

// RunTicketMergeCommand executes the corresponding package operation.
func RunTicketMergeCommand(ctx context.Context, opts TicketMergeCommandOptions) (TicketMergeCommandResult, error) {
	if err := ctx.Err(); err != nil {
		return TicketMergeCommandResult{}, err
	}

	preflight, err := RunTicketMergePreflight(ctx, TicketMergePreflightOptions{
		Ticket:        opts.Ticket,
		Tickets:       opts.Tickets,
		Remote:        opts.Remote,
		ExpectedBase:  opts.ExpectedBase,
		MergeStrategy: opts.MergeStrategy,
	})
	if err != nil {
		return TicketMergeCommandResult{}, err
	}

	result := ticketMergeCommandResult(opts.Ticket.Key, preflight, opts.DryRun)
	if opts.DryRun {
		return result, nil
	}

	if !preflight.Allowed {
		return result, applyTicketMergeCommandResult(ctx, opts.Ticket.Path, result, opts.MergedAt)
	}

	merge, err := opts.Remote.Merge(ctx, MergeRequest{URL: preflight.ReviewRef.URL, Strategy: preflight.MergeStrategy})
	if err != nil {
		result.Preflight.Allowed = false
		result.Preflight.Blockers = append(result.Preflight.Blockers, TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockProviderUnreadable,
			Kind:   TicketMergePreflightBlockKindOperational,
			Detail: err.Error(),
		})
		result.MergeState = TicketMergeStateBlocked
		result.MergeResult = TicketMergeResultBlocked
		result.MergeBlocker = err.Error()

		return result, applyTicketMergeCommandResult(ctx, opts.Ticket.Path, result, opts.MergedAt)
	}

	result.Merge = merge
	result.Status = TicketStatusDone
	result.MergeState = TicketMergeStateMerged
	result.MergeResult = TicketMergeResultMerged
	result.MergeCommit = merge.SHA

	err = applyTicketMergeCommandResult(ctx, opts.Ticket.Path, result, opts.MergedAt)
	if err != nil {
		return result, err
	}

	result.WorktreeCleanup, err = ticketMergeWorktreeCleanup(ctx, opts, result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func ticketMergeWorktreeCleanup(ctx context.Context, opts TicketMergeCommandOptions, result TicketMergeCommandResult) (WorktreeCleanupResult, error) {
	if result.MergeState != TicketMergeStateMerged || opts.Workspace.Mode != workspacecore.ExecutionWorktree {
		return WorktreeCleanupResult{}, nil
	}

	return CleanupMergedTicketWorktree(ctx, WorktreeCleanupOptions{
		ProjectDir:   opts.ProjectDir,
		TicketKey:    opts.Ticket.Key,
		Branch:       opts.Ticket.Branch,
		BranchPolicy: opts.BranchPolicy,
		Workspace:    opts.Workspace,
		Git:          opts.WorktreeGit,
	})
}

func ticketMergeCommandResult(ticketKey string, preflight TicketMergePreflightResult, dryRun bool) TicketMergeCommandResult {
	result := TicketMergeCommandResult{
		TicketKey:     ticketKey,
		Preflight:     preflight,
		Status:        TicketStatusReadyForMerge,
		MergeState:    TicketMergeStatePending,
		MergeResult:   "",
		MergeActor:    TicketMergeActorMerger,
		MergeStrategy: preflight.MergeStrategy,
		DryRun:        dryRun,
	}

	if !preflight.Allowed {
		result.MergeState = TicketMergeStateBlocked
		result.MergeResult = TicketMergeResultBlocked
		result.MergeBlocker = ticketMergeCommandBlockerSummary(preflight.Blockers)
	}

	return result
}

func ticketMergeCommandBlockerSummary(blockers []TicketMergePreflightBlocker) string {
	if len(blockers) == 0 {
		return ""
	}

	parts := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		parts = append(parts, blocker.Code)
	}

	return strings.Join(parts, ",")
}

func applyTicketMergeCommandResult(ctx context.Context, path string, result TicketMergeCommandResult, mergedAt time.Time) error {
	content, err := readLegacyTicketFile(path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	now := mergedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	mergedAtValue := opsmetadata.Quote("")
	if result.MergeState == TicketMergeStateMerged {
		mergedAtValue = opsmetadata.Timestamp(now)
	}

	updated := string(content)
	for _, field := range []struct {
		key        string
		value      string
		beforeKeys []string
	}{
		{key: "status", value: opsmetadata.Quote(result.Status), beforeKeys: []string{"priority", "severity", "type"}},
		{key: "merge_state", value: opsmetadata.Quote(result.MergeState), beforeKeys: []string{"merge_result", "merge_actor", "merge_commit", "merge_blocker", "merged_at", "created_at"}},
		{key: "merge_result", value: opsmetadata.Quote(result.MergeResult), beforeKeys: []string{"merge_actor", "merge_commit", "merge_blocker", "merged_at", "created_at"}},
		{key: "merge_actor", value: opsmetadata.Quote(result.MergeActor), beforeKeys: []string{"merge_commit", "merge_blocker", "merged_at", "created_at"}},
		{key: "merge_commit", value: opsmetadata.Quote(result.MergeCommit), beforeKeys: []string{"merge_blocker", "merged_at", "created_at"}},
		{key: "merge_blocker", value: opsmetadata.Quote(result.MergeBlocker), beforeKeys: []string{"merged_at", "created_at"}},
		{key: "merged_at", value: mergedAtValue, beforeKeys: []string{"created_at"}},
		{key: "updated_at", value: opsmetadata.Timestamp(now), beforeKeys: nil},
	} {
		updated, err = setTicketFrontmatterScalarField(updated, field.key, field.value, field.beforeKeys)
		if err != nil {
			return err
		}
	}

	err = writeLegacyTicketFile(ctx, path, content, []byte(updated))
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}
