package app

import (
	"context"
	"log/slog"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (s service) inspectGuidedResult(ctx context.Context, result *GuidedUseCaseInspection) {
	session := result.Session
	if session.Guided.Work == nil || session.Guided.Work.Invocation == nil {
		return
	}

	observed, err := core.ReadGuidedResult(ctx, s.projectContext.ProjectDir, *session)
	if err != nil {
		result.ResultProblem = "Native result evidence is unavailable or inconsistent. Reconcile the original invocation before claiming delivery."
		result.Problems = append(result.Problems, result.ResultProblem)

		return
	}

	result.Result = &observed
}

func (host *guidedWorkHost) reconcileResult(ctx context.Context, operations service, id string) {
	if ctx.Err() != nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	run, err := operations.ReadWorkflow(ctx, id)
	if err != nil || run.Admission == nil || run.Admission.Guided == nil {
		return
	}

	scope := authoringcore.GuidedUseCaseScope{RunID: run.Admission.Guided.RunID, ProjectID: run.ProjectID, Namespace: run.Namespace}
	if _, err := authoringcore.ReconcileGuidedResult(ctx, operations.projectContext.ProjectDir, scope, core.NewGuidedWorkflowObserver(operations.projectContext.ProjectDir)); err != nil {
		slog.WarnContext(ctx, "Guided result evidence needs recovery", "code", "guided_result_blocked")
	}
}
