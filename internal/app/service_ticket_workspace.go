package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type ticketStartWorktreePreparer interface {
	Prepare(context.Context, executioncore.WorktreePlanOptions) (executioncore.WorktreePrepareResult, error)
}

var newTicketStartWorktreePreparer = func() ticketStartWorktreePreparer {
	service := executioncore.NewWorktreeService()

	return service
}

func prepareTicketStartWorkspace(ctx context.Context, projectContext commandProjectContext, ticket ticketcore.Indexed, config workspaceconfig.Config, opts ticketStartPreflightOptions) (ticketStartWorkspace, error) {
	workspace := ticketStartWorkspace{
		Ticket:     ticket,
		TicketPath: ticket.Path,
	}
	if config.Work.Workspace.Mode != workspacecore.ExecutionWorktree {
		return workspace, nil
	}

	selection, err := core.SelectTicketBranch(ticket)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	planOptions := executioncore.WorktreePlanOptions{
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

	return prepareTicketStartWorktree(ctx, ticket, selection, planOptions)
}

func previewTicketStartWorktree(ticket ticketcore.Indexed, planOptions executioncore.WorktreePlanOptions) (ticketStartWorkspace, error) {
	plan, err := executioncore.NewWorktreePlan(planOptions)
	if err != nil {
		return ticketStartWorkspace{}, err
	}

	return ticketStartWorkspaceFromPlan(ticket, plan)
}

func prepareTicketStartWorktree(ctx context.Context, ticket ticketcore.Indexed, selection core.TicketBranchSelection, planOptions executioncore.WorktreePlanOptions) (ticketStartWorkspace, error) {
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

func ticketStartWorkspaceFromPlan(ticket ticketcore.Indexed, plan executioncore.WorktreePlan) (ticketStartWorkspace, error) {
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
	err := workflowcore.NewTicketReviewMetadataUpdater().UpdateReviewMetadata(ctx, workflowcore.UpdateTicketReviewMetadataOptions{
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

func writeTicketStartWorktreeMetadata(plan executioncore.WorktreePlan) error {
	path, err := executioncore.WorktreeMetadataPath(plan)
	if err != nil {
		return err
	}

	return executioncore.WriteWorktreeMetadata(path, executioncore.NewWorktreeMetadata(plan, time.Now().UTC()))
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
