package app

import (
	"context"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// AdmitGuidedWork binds original progression intent to one exact native invocation.
func (s service) AdmitGuidedWork(ctx context.Context, command authoringcore.GuidedUseCaseCommand) (authoringcore.Session, error) {
	if command.Action != "admit" {
		return authoringcore.Session{}, errors.New("unsupported guided admission command")
	}

	return s.applyGuidedWorkStep(ctx, command, service.admitGuidedWork)
}

func (s service) admitGuidedWork(ctx context.Context, session *authoringcore.Session) error {
	if s.guidedWorkHost == nil {
		return &core.GuidedPrerequisiteError{Action: "Start a managed workflow host, then retry this retained operation."}
	}

	if err := s.guidedWorkHost.ready(); err != nil {
		return err
	}

	work := session.Guided.Work
	if work == nil || work.Materialization == nil {
		return workflowstate.ErrConflict
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

func (s service) ensureGuidedInvocation(ctx context.Context, session *authoringcore.Session) error {
	if session.Guided.Work.Invocation != nil {
		return nil
	}

	return s.retainGuidedInvocation(ctx, session)
}

func (s service) reserveGuidedAdmission(ctx context.Context, session *authoringcore.Session) error {
	checkpoint, err := core.GuidedPreparationCheckpoint(*session)
	if err != nil {
		return err
	}

	digest, err := core.GuidedCheckpointDigest(checkpoint)
	if err != nil {
		return err
	}

	work := session.Guided.Work

	preview, err := s.PreviewSelectedWork(ctx, workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: work.Materialization.SetID}}, "")
	if err != nil {
		return &core.GuidedPrerequisiteError{Action: "Correct native workflow authority and canonical target setup, then retry the retained admission.", Cause: err}
	}

	if err := core.ValidateGuidedWorkPreview(*work.Materialization, preview.View.Preview); err != nil {
		return &core.GuidedPrerequisiteError{Action: "Restore the sealed native task membership and contracts before retrying admission.", Cause: err}
	}

	policy := preview.View.ExecutionPolicy
	policy.IncludeLaterTickets = false
	includeLater := false
	proof := &workflowstate.GuidedAdmission{RunID: session.ID, Actor: session.Guided.Actor, Policy: core.GuidedWorkAdmissionPolicy, OperationDigest: session.Guided.PendingOperation.Digest, CheckpointDigest: digest, MaterializationID: work.Materialization.ID}
	work.Admission = &workflowstate.StartInput{
		RequestKey:          "guided-work-" + session.Guided.PendingOperation.Digest,
		Selection:           preview.Request,
		IncludeLaterTickets: &includeLater,
		Admission:           &workflowstate.AdmissionGuard{Guided: proof, SnapshotHash: preview.Request.SnapshotHash, Policy: preview.View.Policy, ExecutionPolicy: policy},
	}

	if err := core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
		return err
	}

	return s.guidedAdmissionEffect("preview")
}

func (s service) retainGuidedInvocation(ctx context.Context, session *authoringcore.Session) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	request := core.WorkflowStartRequest{StartInput: *session.Guided.Work.Admission, ResolveHarness: s.prepareGuidedWorkflowHarness}

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
	session.Guided.Work.Invocation = &authoringcore.GuidedInvocationReference{ID: run.ID, RequestKey: run.RequestKey, Revision: run.Revision}

	return core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session)
}

func (s service) guidedAdmissionEffect(phase string) error {
	if s.afterGuidedAdmission != nil {
		return s.afterGuidedAdmission(phase)
	}

	return nil
}
