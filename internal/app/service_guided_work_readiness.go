package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) prepareGuidedWorkflowHarness(ctx context.Context, config core.WorkspaceConfig) (core.WorkflowHarnessResolution, error) {
	resolved, err := s.prepareWorkflowHarness(ctx, config)
	if err != nil {
		return core.WorkflowHarnessResolution{}, &core.GuidedPrerequisiteError{Action: "Configure an available worker harness and credentials before retrying admission.", Cause: err}
	}

	return resolved, nil
}

func (s service) checkGuidedWorkPrerequisites(ctx context.Context, selection workflowstate.PreviewRequest) error {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return err
	}

	if err := s.checkGuidedWorker(ctx, config); err != nil {
		return err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return err
	}

	if remote, ok := provider.(core.RemoteReviewAdapter); ok {
		if err := remote.CheckWorkflowRepositoryProvider(ctx, selection.Target.Repository); err != nil {
			return &core.GuidedPrerequisiteError{Action: "Configure the canonical review provider, CLI and credentials before retrying admission.", Cause: err}
		}
	}

	return s.checkGuidedPublisher(config, selection.Target.Repository)
}

func (s service) checkGuidedWorker(ctx context.Context, config core.WorkspaceConfig) error {
	if s.workflowWorker != nil {
		return nil
	}

	selection, err := s.resolveHarnessSelection(ctx, config, harnesscore.Worker, "")
	if err != nil || !selection.View.Readiness.Ready() {
		return &core.GuidedPrerequisiteError{Action: "Configure an available worker harness and credentials before retrying admission.", Cause: err}
	}

	return nil
}

func (s service) checkGuidedPublisher(config core.WorkspaceConfig, repository string) error {
	if config.Work.Integration.PublicationProfile != core.WorkflowPublicationGitAtomicDualRef {
		return nil
	}

	if s.env == nil {
		return &core.GuidedPrerequisiteError{Action: "Configure the private certified publisher for the canonical target before retrying admission."}
	}

	_, err := core.LoadWorkflowNativePublisher(s.env.ConfigDir(), repository, "refs/heads/"+config.Work.BranchPolicy.IntegrationBranch)
	if err != nil {
		return &core.GuidedPrerequisiteError{Action: "Configure the private certified publisher for the canonical target before retrying admission.", Cause: err}
	}

	return nil
}
