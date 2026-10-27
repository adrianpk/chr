package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// PrepareGuidedWork generates and materializes exactly the committed approved scope.
func (s service) PrepareGuidedWork(ctx context.Context, command authoringcore.GuidedUseCaseCommand) (authoringcore.Session, error) {
	if command.Action != "prepare" {
		return authoringcore.Session{}, errors.New("unsupported guided preparation command")
	}

	return s.applyGuidedWorkStep(ctx, command, service.prepareGuidedWork)
}

func (s service) prepareGuidedWork(ctx context.Context, session *authoringcore.Session) error {
	if session.Guided.SpecApproval == nil {
		return workflowstate.ErrConflict
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return err
	}

	if config.Workflow.Version != 1 {
		return &authoringcore.GuidedPrerequisiteError{Action: "Configure native workflow version 1 before retrying work preparation."}
	}

	if err := s.reserveGuidedPlan(ctx, session); err != nil {
		return err
	}

	operation := session.Guided.PendingOperation
	if operation.PreparedPlan == nil {
		prepared, err := s.generateGuidedPlan(ctx, *session)
		if err != nil {
			return err
		}

		operation.PreparedPlan = &prepared

		if err := authoringcore.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
			return err
		}
	}

	if err := s.publishGuidedPlan(ctx, session); err != nil {
		return err
	}

	return s.materializeGuidedWork(ctx, session)
}

func (s service) reserveGuidedPlan(ctx context.Context, session *authoringcore.Session) error {
	if session.Guided.Work != nil {
		return nil
	}

	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}

	paths, err := authoringcore.GeneratedPlanLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	if _, err := os.Stat(paths.Path); !errors.Is(err, os.ErrNotExist) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	revision, err := authoringcore.PlanTargetRevision(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	session.Guided.Work = &authoringcore.GuidedWorkState{PublicationOperation: generationcore.PlanPublicationOperation{ID: "guided-plan-" + session.Guided.PendingOperation.Digest, Scope: scope, ExpectedRevision: revision, TicketDelivery: true}}

	return authoringcore.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session)
}

func (s service) generateGuidedPlan(ctx context.Context, session authoringcore.Session) (authoringcore.GuidedPreparedPlan, error) {
	bound, err := s.bindGeneration(ctx, session.Guided.Work.PublicationOperation.ID, string(quizmodel.ActionGeneratePlan))
	ctx = bound.context(ctx)

	if err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	observed, err := bound.guidedPlanSources(ctx, session)
	if err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	sources := observed.Sources

	if _, err := bound.assessPlanReadiness(ctx, sources); err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	generated, err := bound.generatePlanWithConfiguredBackend(ctx, sources.Scope, observed.State, sources)
	if err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	plan := generated.Plan
	plan.SourceRequestID, plan.SourceResponseID = strings.Join(generated.RequestIDs, ","), strings.Join(generated.ResponseIDs, ",")
	plan.Order = authoringcore.ScopePlanDeliveryOrder(plan.Order, session.Guided.Work.PublicationOperation)

	plan.RequireOrderMap = true
	if err := authoringcore.ValidateGuidedPlannedOrder(plan.Order); err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	if err := generationcore.ValidateGeneratedPlan(plan); err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	encoded, err := json.Marshal(plan)
	if err != nil {
		return authoringcore.GuidedPreparedPlan{}, err
	}

	if len(encoded) > 64<<10 {
		return authoringcore.GuidedPreparedPlan{}, errors.New("guided generated plan exceeds 64 KiB")
	}

	return authoringcore.GuidedPreparedPlan{Plan: plan, TargetRevision: session.Guided.Work.PublicationOperation.ExpectedRevision, InputDigest: sources.Hash}, nil
}

func (s service) publishGuidedPlan(ctx context.Context, session *authoringcore.Session) error {
	if session.Guided.Work.Publication != nil {
		return nil
	}

	result, err := authoringcore.PublishGuidedPlan(ctx, s.projectContext.ProjectDir, authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace})
	if err != nil {
		return err
	}

	if err := s.guidedPreparationEffect("plan"); err != nil {
		return err
	}

	session.Guided.Work.Publication = &result

	return authoringcore.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session)
}

func (s service) materializeGuidedWork(ctx context.Context, session *authoringcore.Session) error {
	operation := session.Guided.PendingOperation
	if operation.PreparedWorkflow == nil {
		input, err := s.generateGuidedContracts(ctx, *session)
		if err != nil {
			return &authoringcore.GuidedPrerequisiteError{Action: "Correct native contract generation or provider setup, then retry the retained preparation.", Cause: err}
		}

		operation.PreparedWorkflow = &input

		if err := authoringcore.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
			return err
		}
	}

	if session.Guided.Work.Materialization == nil {
		if err := s.retainGuidedMaterialization(ctx, session); err != nil {
			return err
		}
	}

	return authoringcore.SealGuidedPreparation(ctx, s.projectContext.ProjectDir, session)
}

func (s service) retainGuidedMaterialization(ctx context.Context, session *authoringcore.Session) error {
	result, err := s.MaterializeWorkflowPlan(ctx, *session.Guided.PendingOperation.PreparedWorkflow)
	if err != nil {
		return err
	}

	if err := s.guidedPreparationEffect("materialization"); err != nil {
		return err
	}

	reference, err := core.GuidedMaterializationBinding(result, *session.Guided.PendingOperation.PreparedWorkflow)
	if err != nil {
		return err
	}

	session.Guided.Work.Materialization = &reference
	if err := authoringcore.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
		return err
	}

	return nil
}

func (s service) guidedPreparationEffect(phase string) error {
	if s.afterGuidedPreparation != nil {
		return s.afterGuidedPreparation(phase)
	}

	return nil
}
