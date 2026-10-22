package app

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

// AdmitGuidedWork binds original progression intent to one exact native invocation.
func (s service) AdmitGuidedWork(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	if command.Action != "admit" {
		return core.AuthoringSession{}, errors.New("unsupported guided admission command")
	}

	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *core.AuthoringSession) error {
		bounded := s
		bounded.ctx, bounded.workflowNamespace = ctx, session.Namespace

		return bounded.admitGuidedWork(ctx, session)
	})
}

func (s service) admitGuidedWork(ctx context.Context, session *core.AuthoringSession) error {
	if s.guidedWorkHost == nil {
		return &core.GuidedPrerequisiteError{Action: "Start a managed Web host, then retry this retained operation."}
	}

	if err := s.guidedWorkHost.ready(); err != nil {
		return err
	}

	work := session.Guided.Work
	if work == nil || work.Materialization == nil {
		return core.ErrWorkflowConflict
	}

	if work.Admission == nil {
		if err := s.reserveGuidedAdmission(ctx, session); err != nil {
			return err
		}
	}

	if err := s.ensureGuidedInvocation(ctx, session); err != nil {
		return err
	}

	if err := s.guidedWorkHost.activate(s, work.Invocation.ID); err != nil {
		return err
	}

	if err := s.guidedAdmissionEffect("start"); err != nil {
		return err
	}

	session.Guided.Stage, session.Guided.Status = "execution", "running"

	return nil
}

func (s service) ensureGuidedInvocation(ctx context.Context, session *core.AuthoringSession) error {
	if session.Guided.Work.Invocation != nil {
		return nil
	}

	return s.retainGuidedInvocation(ctx, session)
}

func (s service) reserveGuidedAdmission(ctx context.Context, session *core.AuthoringSession) error {
	checkpoint, err := core.GuidedPreparationCheckpoint(*session)
	if err != nil {
		return err
	}

	digest, err := core.GuidedCheckpointDigest(checkpoint)
	if err != nil {
		return err
	}

	work := session.Guided.Work

	preview, err := s.PreviewSelectedWork(ctx, core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "slice_set", Identity: work.Materialization.SetID}}, "")
	if err != nil {
		return &core.GuidedPrerequisiteError{Action: "Correct native workflow authority and canonical target setup, then retry the retained admission.", Cause: err}
	}

	if err := core.ValidateGuidedWorkPreview(*work.Materialization, preview.View.Preview); err != nil {
		return &core.GuidedPrerequisiteError{Action: "Restore the sealed native task membership and contracts before retrying admission.", Cause: err}
	}

	policy := preview.View.ExecutionPolicy
	policy.IncludeLaterTickets = false
	includeLater := false
	proof := &core.GuidedWorkflowAdmission{RunID: session.ID, Actor: session.Guided.Actor, Policy: core.GuidedWorkAdmissionPolicy, OperationDigest: session.Guided.PendingOperation.Digest, CheckpointDigest: digest, MaterializationID: work.Materialization.ID}
	work.Admission = &core.WorkflowStartRequest{RequestKey: "guided-work-" + session.Guided.PendingOperation.Digest, Selection: preview.Request, IncludeLaterTickets: &includeLater, Admission: &core.WorkflowAdmissionGuard{Guided: proof, SnapshotHash: preview.Request.SnapshotHash, Policy: preview.View.Policy, ExecutionPolicy: policy}}

	if err := core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
		return err
	}

	return s.guidedAdmissionEffect("preview")
}

func (s service) retainGuidedInvocation(ctx context.Context, session *core.AuthoringSession) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	request := *session.Guided.Work.Admission
	request.ResolveHarness = s.prepareGuidedWorkflowHarness

	started, err := store.StartWorkflowGuarded(ctx, request, s.boundAdmissionCheck(request.Selection, func(ctx context.Context) error {
		return s.checkGuidedWorkPrerequisites(ctx, request.Selection)
	}))
	if err != nil {
		var prerequisite *core.GuidedPrerequisiteError
		if errors.As(err, &prerequisite) {
			return err
		}

		return &core.GuidedPrerequisiteError{Action: "Reconcile the retained admission with native policy, target and worker setup before retrying.", Cause: err}
	}

	if err := s.guidedAdmissionEffect("admission"); err != nil {
		return err
	}

	run := started.Invocation
	session.Guided.Work.Invocation = &core.GuidedInvocationReference{ID: run.ID, RequestKey: run.RequestKey, Revision: run.Revision}

	return core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session)
}

func (s service) guidedAdmissionEffect(phase string) error {
	if s.afterGuidedAdmission != nil {
		return s.afterGuidedAdmission(phase)
	}

	return nil
}
