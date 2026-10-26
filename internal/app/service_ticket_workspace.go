package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type ticketStartWorktreePreparer interface {
	Prepare(context.Context, core.WorktreePlanOptions) (core.WorktreePrepareResult, error)
}

var newTicketStartWorktreePreparer = func() ticketStartWorktreePreparer {
	service := core.NewWorktreeService()

	return service
}

func prepareTicketStartWorkspace(ctx context.Context, projectContext commandProjectContext, ticket core.IndexedTicket, config core.WorkspaceConfig, opts ticketStartPreflightOptions) (ticketStartWorkspace, error) {
	workspace := ticketStartWorkspace{
		Ticket:     ticket,
		TicketPath: ticket.Path,
	}
	if config.Work.Workspace.Mode != core.WorkWorkspaceModeWorktree {
		return workspace, nil
	}

	selection, err := core.SelectTicketBranch(ticket)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	planOptions := core.WorktreePlanOptions{
		ProjectDir:   projectContext.ProjectDir,
		TicketKey:    ticket.Key,
		Branch:       selection.Branch,
		BranchPolicy: config.Work.BranchPolicy,
		Workspace:    config.Work.Workspace,
		WorktreePath: workspace.ExecutionWorkspaceDir,
	}

	if !opts.PrepareWorktree {
		return previewTicketStartWorktree(ticket, planOptions)
	}

	return prepareTicketStartWorktree(ctx, projectContext.ProjectDir, ticket, config, selection, planOptions)
}

func previewTicketStartWorktree(ticket core.IndexedTicket, planOptions core.WorktreePlanOptions) (ticketStartWorkspace, error) {
	plan, err := core.NewWorktreePlan(planOptions)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	return ticketStartWorkspaceFromPlan(ticket, plan)
}

func prepareTicketStartWorktree(ctx context.Context, projectDir string, ticket core.IndexedTicket, config core.WorkspaceConfig, selection core.TicketBranchSelection, planOptions core.WorktreePlanOptions) (ticketStartWorkspace, error) {
	result, err := newTicketStartWorktreePreparer().Prepare(ctx, planOptions)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace, err := ticketStartWorkspaceFromPlan(ticket, result.Plan)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace.PreparedWorktree = true

	if selection.Derived {
		workspace, err = writeDerivedTicketBranch(ctx, workspace, selection.Branch)
		if err != nil {
			return ticketStartWorkspace{}, err
		}
	}

	if err := writeTicketStartWorktreeMetadata(result.Plan); err != nil {
		return ticketStartWorkspace{}, err
	}

	return workspace, nil
}

func ticketStartWorkspaceFromPlan(ticket core.IndexedTicket, plan core.WorktreePlan) (ticketStartWorkspace, error) {
	ticketPath, err := equivalentProjectPath(plan.ProjectDir, plan.WorktreePath, ticket.Path)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	return ticketStartWorkspace{
		Ticket:                ticket,
		TicketPath:            ticketPath,
		TicketBranch:          plan.Branch,
		ExecutionWorkspaceDir: plan.WorktreePath,
		WorkspaceBaseRef:      plan.BaseRef,
		WorkspaceBaseSHA:      plan.BaseSHA,
	}, nil
}

func writeDerivedTicketBranch(ctx context.Context, workspace ticketStartWorkspace, branch string) (ticketStartWorkspace, error) {
	err := core.NewTicketReviewMetadataUpdater().UpdateReviewMetadata(ctx, core.UpdateTicketReviewMetadataOptions{
		Path:      workspace.TicketPath,
		Branch:    branch,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	workspace.Ticket.Branch = branch

	return workspace, nil
}

func writeTicketStartWorktreeMetadata(plan core.WorktreePlan) error {
	path, err := core.WorktreeMetadataPath(plan)
	if err != nil {
		return err
	}

	return core.WriteWorktreeMetadata(path, core.NewWorktreeMetadata(plan, time.Now().UTC()))
}

func equivalentProjectPath(sourceRoot, targetRoot, path string) (string, error) {
	targetRelative, targetErr := filepath.Rel(targetRoot, path)
	if targetErr == nil && !strings.HasPrefix(targetRelative, "..") && !filepath.IsAbs(targetRelative) {
		return path, nil
	}

	relative, err := filepath.Rel(sourceRoot, path)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path %q is not under source root %q", path, sourceRoot)
	}

	return filepath.Join(targetRoot, relative), nil
}
