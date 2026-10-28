package app

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

func guidedInvocationMatches(run workflowcore.Invocation, session authoringcore.Session) bool {
	if run.Admission == nil || run.Admission.Guided == nil || session.Guided.Work.Materialization == nil {
		return false
	}

	proof := run.Admission.Guided

	return run.ProjectID == session.ProjectID && run.Namespace == session.Namespace && run.RequestKey == session.Guided.Work.Invocation.RequestKey && proof.RunID == session.ID && proof.Actor == session.Guided.Actor && proof.MaterializationID == session.Guided.Work.Materialization.ID
}

func (s service) inspectGuidedWork(ctx context.Context, result *GuidedUseCaseInspection) {
	session := result.Session

	work := session.Guided.Work
	if work == nil || work.Invocation == nil {
		return
	}

	s.workflowNamespace = session.Namespace

	run, err := s.ReadWorkflow(ctx, work.Invocation.ID)
	if err != nil || !guidedInvocationMatches(run, *session) {
		result.Problems = append(result.Problems, "The original native work binding needs recovery before execution status can be shown.")

		return
	}

	status, err := s.WorkflowStatus(ctx, run.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		result.Problems = append(result.Problems, "Native work status is unavailable. Inspect its retained invocation before continuing.")

		return
	}

	result.Work = &status
}
