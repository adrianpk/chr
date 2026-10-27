package app

import (
	"context"
	"time"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type telegramHarnessStatus struct {
	Selections      HarnessSelectionStatus
	Retained        []WorkflowWorkStatus
	More            bool
	RetainedProblem string
}

type telegramHarnessService interface {
	InspectTelegramHarnessSelections(context.Context, telegramcore.OperationTarget) (telegramHarnessStatus, error)
	ChangeTelegramHarnessSelections(context.Context, telegramcore.OperationTarget, workspaceconfig.HarnessSelectionChangeRequest, func(context.Context) error) (harnesscore.SelectionSnapshot, error)
}

func (s service) telegramHarnessProject(ctx context.Context, target telegramcore.OperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := projectcore.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}

	return scoped, telegramcore.ValidateProjectTarget(project, target)
}

// InspectTelegramHarnessSelections resolves the frozen registry root and calls
// the shared selection owner, keeping private profile resources out of replies.
func (s service) InspectTelegramHarnessSelections(ctx context.Context, target telegramcore.OperationTarget) (telegramHarnessStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	scoped, err := s.telegramHarnessProject(ctx, target)
	if err != nil {
		return telegramHarnessStatus{}, err
	}

	selections, err := scoped.InspectHarnessSelections(ctx, "spec_generation")
	if err != nil {
		return telegramHarnessStatus{}, err
	}

	status := telegramHarnessStatus{Selections: selections}
	retainedWork := scoped.telegramHarnessRetained(ctx)
	status.Retained, status.More, status.RetainedProblem = retainedWork.Items, retainedWork.More, retainedWork.Problem

	return status, nil
}

func (s service) telegramHarnessRetained(ctx context.Context) telegramHarnessRetention {
	page, err := s.WorkflowStatusPage(ctx, 0, 4)
	if err != nil {
		return telegramHarnessRetention{Items: nil, More: false, Problem: "Retained work could not be loaded. Inspect workflow status locally."}
	}

	return telegramHarnessRetention{Items: page.Items, More: page.NextOffset >= 0, Problem: ""}
}

// ChangeTelegramHarnessSelections resolves the registry root and delegates the
// confirmed envelope to the shared atomic configuration owner.
func (s service) ChangeTelegramHarnessSelections(ctx context.Context, target telegramcore.OperationTarget, change workspaceconfig.HarnessSelectionChangeRequest, authorize func(context.Context) error) (harnesscore.SelectionSnapshot, error) {
	scoped, err := s.telegramHarnessProject(ctx, target)
	if err != nil {
		return harnesscore.SelectionSnapshot{}, err
	}

	return scoped.ChangeHarnessSelections(ctx, change, authorize)
}
