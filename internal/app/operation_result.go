package app

import (
	"errors"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func operationResult(summary string, err error) operation.Result {
	if err == nil {
		return operation.Result{Status: operation.OK, Summary: summary}
	}

	if errors.Is(err, workflowstate.ErrConflict) {
		return operation.Result{Status: operation.Stale, Summary: "The source changed after it was loaded.", Validation: []string{err.Error()}, Reload: true}
	}

	if errors.Is(err, authoringcore.ErrProjectDumpReconciliation) {
		return operation.Result{Status: operation.Blocked, Summary: "The retained operation needs reconciliation.", BlockedReason: err.Error(), Retry: true}
	}

	if errors.Is(err, workflowstate.ErrOwned) || errors.Is(err, workflowstate.ErrObservationRequired) || errors.Is(err, authoringcore.ErrAuthorityRequired) {
		return operation.Result{Status: operation.Blocked, Summary: "Operation blocked.", BlockedReason: err.Error(), Reload: true}
	}

	return operation.Result{Status: operation.ValidationFailed, Summary: "Write action failed validation.", Validation: []string{err.Error()}}
}
