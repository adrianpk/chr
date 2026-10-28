package app

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s service) prepareGuidedWorkflowHarness(ctx context.Context, config workspaceconfig.Config) (workflowcore.HarnessResolution, error) {
	resolved, err := s.prepareWorkflowHarness(ctx, config)
	if err != nil {
		return workflowcore.HarnessResolution{}, &authoringcore.GuidedPrerequisiteError{Action: "Configure an available worker harness and credentials before retrying admission.", Cause: err}
	}

	return resolved, nil
}

func (s service) checkGuidedWorkPrerequisites(ctx context.Context, selection workflowstate.PreviewRequest) error {
	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
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

	if remote, ok := provider.(workflowcore.RemoteReviewAdapter); ok {
		if err := remote.CheckWorkflowRepositoryProvider(ctx, selection.Target.Repository); err != nil {
			return &authoringcore.GuidedPrerequisiteError{Action: "Configure the canonical review provider, CLI and credentials before retrying admission.", Cause: err}
		}
	}

	return s.checkGuidedPublisher(config, selection.Target.Repository)
}

func (s service) checkGuidedWorker(ctx context.Context, config workspaceconfig.Config) error {
	if s.workflowWorker != nil {
		return nil
	}

	selection, err := s.resolveHarnessSelection(ctx, config, harnesscore.Worker, "")
	if err != nil || !selection.View.Readiness.Ready() {
		return &authoringcore.GuidedPrerequisiteError{Action: "Configure an available worker harness and credentials before retrying admission.", Cause: err}
	}

	return nil
}

func (s service) checkGuidedPublisher(config workspaceconfig.Config, repository string) error {
	if config.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return nil
	}

	if s.env == nil {
		return &authoringcore.GuidedPrerequisiteError{Action: "Configure the private certified publisher for the canonical target before retrying admission."}
	}

	_, err := workflowcore.LoadWorkflowNativePublisher(s.env.ConfigDir(), repository, "refs/heads/"+config.Work.BranchPolicy.IntegrationBranch)
	if err != nil {
		return &authoringcore.GuidedPrerequisiteError{Action: "Configure the private certified publisher for the canonical target before retrying admission.", Cause: err}
	}

	return nil
}
