package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramHarnessStatus struct {
	Selections      HarnessSelectionStatus
	Retained        []WorkflowWorkStatus
	More            bool
	RetainedProblem string
}

type telegramHarnessService interface {
	InspectTelegramHarnessSelections(context.Context, core.TelegramOperationTarget) (telegramHarnessStatus, error)
}

func (s service) telegramHarnessProject(ctx context.Context, target core.TelegramOperationTarget) (service, error) {
	scoped, err := s.telegramProjectService(ctx, target)
	if err != nil {
		return scoped, err
	}

	project := core.Project{ID: scoped.projectContext.ProjectID, Path: scoped.projectContext.ProjectDir, Enabled: true}

	return scoped, core.ValidateTelegramProjectTarget(project, target)
}

// InspectTelegramHarnessSelections resolves the frozen registry root and calls
// the shared selection owner, keeping private profile resources out of replies.
func (s service) InspectTelegramHarnessSelections(ctx context.Context, target core.TelegramOperationTarget) (telegramHarnessStatus, error) {
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
	status.Retained, status.More, status.RetainedProblem = scoped.telegramHarnessRetained(ctx)

	return status, nil
}

func (s service) telegramHarnessRetained(ctx context.Context) ([]WorkflowWorkStatus, bool, string) {
	page, err := s.WorkflowStatusPage(ctx, 0, 4)
	if err != nil {
		return nil, false, "Retained work could not be loaded. Inspect workflow status locally."
	}

	return page.Items, page.NextOffset >= 0, ""
}
