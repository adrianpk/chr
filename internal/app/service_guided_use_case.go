package app

import (
	"context"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// GuidedUseCaseInspection contains retained state and prospective local readiness.
type GuidedUseCaseInspection struct {
	Quiz          *quizcore.Observation
	Ideas         []core.GuidedLaterIdea
	Result        *core.WorkflowResultInspection
	ResultProblem string
	Work          *WorkflowWorkStatus
	ProjectName   string
	Namespace     string
	Session       *core.AuthoringSession
	Runs          []core.AuthoringSession
	Policy        []string
	Problems      []string
	CommittedDump string
	Question      string
	Artifacts     []GuidedUseCaseArtifact
}

func (s service) guidedUseCaseScope(id, namespace string) (core.GuidedUseCaseScope, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace})
	if err != nil {
		return core.GuidedUseCaseScope{}, err
	}

	return core.GuidedUseCaseScope{RunID: id, ProjectID: s.projectContext.ProjectID, Namespace: layout.Namespace()}, nil
}

// StartGuidedUseCase retains explicit progression intent in the selected scope.
func (s service) StartGuidedUseCase(ctx context.Context, id, namespace, actor string) (core.AuthoringSession, error) {
	scope, err := s.guidedUseCaseScope(id, namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	return core.StartGuidedUseCase(ctx, s.projectContext.ProjectDir, core.GuidedUseCaseStart{GuidedUseCaseScope: scope, Actor: actor, ProgressionIntent: true})
}

// ChangeGuidedUseCase applies current input and advances through available authoring stages.
func (s service) ChangeGuidedUseCase(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope
	if command.Action == "idea" {
		return core.RetainGuidedIdea(ctx, s.projectContext.ProjectDir, command)
	}

	if guidedControlAction(command.Action) {
		return s.changeGuidedControl(ctx, command)
	}

	if host, ok := s.guidedWorkHost.(guidedProgressActivator); ok {
		return s.changeGuidedOnHost(ctx, command, host)
	}

	session, err := s.applyGuidedCommand(ctx, command)
	if err != nil {
		return session, err
	}

	return s.advanceGuidedUseCase(ctx, session)
}

// InspectGuidedUseCase reads state, immutable input and local setup without dispatch.
func (s service) InspectGuidedUseCase(ctx context.Context, id, namespace string) (GuidedUseCaseInspection, error) {
	scope, err := s.guidedUseCaseScope(id, namespace)
	if err != nil {
		return GuidedUseCaseInspection{}, err
	}

	result := GuidedUseCaseInspection{Namespace: scope.Namespace, ProjectName: s.projectContext.ProjectName}
	s.inspectGuidedReadiness(ctx, &result)

	if id == "" {
		result.Runs, err = core.ListGuidedUseCases(ctx, s.projectContext.ProjectDir, scope.ProjectID, scope.Namespace)

		return result, err
	}

	session, err := core.ReadGuidedUseCase(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return result, err
	}

	result.Session = &session
	if err := s.inspectGuidedArtifacts(ctx, &result); err != nil {
		return result, err
	}

	s.inspectGuidedQuestion(ctx, &result)
	s.inspectGuidedWork(ctx, &result)
	s.inspectGuidedResult(ctx, &result)

	result.Ideas, err = core.ReadGuidedIdeas(ctx, s.projectContext.ProjectDir, session)

	return result, err
}

func (s service) inspectGuidedReadiness(ctx context.Context, result *GuidedUseCaseInspection) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		result.Problems = append(result.Problems, err.Error())

		return
	}

	result.Policy = []string{
		"Execution target: " + config.Work.BranchPolicy.IntegrationBranch,
		"Review policy: " + string(config.Review.Mode),
		"Publication policy: " + config.Work.Integration.PublicationProfile,
		fmt.Sprintf("Limits: %d worker attempts, %d integration attempts, %d seconds per worker", config.Work.MaxWorkerAttempts, config.Work.MaxIntegrationAttempts, config.Work.WorkerTimeoutSeconds),
	}
	if config.Workflow.Version != 1 {
		result.Problems = append(result.Problems, "Native workflow execution is not configured.")
	}

	if err := core.CheckGuidedCheckpointReadiness(ctx, s.projectContext.ProjectDir); err != nil {
		result.Problems = append(result.Problems, err.Error())
	}

	for _, role := range []core.HarnessRole{core.HarnessGeneration, core.HarnessWorker} {
		selection, err := s.ResolveHarnessSelection(ctx, role, string(quizmodel.ActionNextQuestion))
		if err != nil {
			result.Problems = append(result.Problems, string(role)+" setup: "+err.Error())

			continue
		}

		if !selection.Readiness.Ready() {
			result.Problems = append(result.Problems, string(role)+" setup is incomplete.")
		}
	}
}

// GuidedUseCaseArtifact is a complete original document read from its retained commit.
type GuidedUseCaseArtifact struct{ Stage, Reference, Commit, Content string }

func (s service) inspectGuidedArtifacts(ctx context.Context, result *GuidedUseCaseInspection) error {
	for _, checkpoint := range result.Session.Guided.Checkpoints {
		for _, artifact := range checkpoint.Artifacts {
			if !strings.HasSuffix(artifact.Path, "/artifact.md") {
				continue
			}

			content, err := core.ReadGuidedCommittedArtifact(ctx, s.projectContext.ProjectDir, checkpoint, artifact.Path)
			if err != nil {
				return err
			}

			if checkpoint.Stage == "dump" && result.CommittedDump == "" {
				result.CommittedDump = string(content)

				continue
			}

			result.Artifacts = append(result.Artifacts, GuidedUseCaseArtifact{Stage: checkpoint.Stage, Reference: artifact.Path, Commit: checkpoint.Commit, Content: string(content)})
		}
	}

	return nil
}

func (s service) inspectGuidedQuestion(ctx context.Context, result *GuidedUseCaseInspection) {
	session := result.Session

	state := session.Guided
	if state.Stage != "quiz" || state.PendingOperation != nil || state.Quiz == nil {
		return
	}

	observed, err := s.quizRepository().Observe(ctx, quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: state.CaseSlug})
	if err != nil {
		result.Problems = append(result.Problems, "Clarification could not be read. Correct its stored inputs before retrying.")

		return
	}

	if observed.State.SessionID != state.Quiz.SessionID || observed.Revision != state.Quiz.Revision {
		result.Problems = append(result.Problems, "Clarification changed outside this run. Its original binding requires recovery.")

		return
	}

	result.Quiz = &observed

	question := observed.State.ActiveQuestion
	if question != nil && question.ID == state.Quiz.QuestionID {
		result.Question = question.Prompt
	}
}
