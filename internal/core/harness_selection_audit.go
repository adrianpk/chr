package core

import (
	"errors"
	"time"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func applyHarnessSelectionAudit(config *WorkspaceConfig, values map[string]string) error {
	audit := &config.HarnessAudit
	fields := []struct {
		key    string
		target *string
	}{
		{"project_ref", &audit.ProjectRef}, {"changed_by", &audit.ChangedBy}, {"changed_at", &audit.ChangedAt}, {"request_id", &audit.RequestID}, {"request_digest", &audit.RequestDigest},
	}
	count := 0

	for _, field := range fields {
		if value, ok := values["harnesses."+field.key]; ok {
			*field.target = value
			count++
		}
	}

	if count == 0 {
		return nil
	}

	if count != len(fields) {
		return errors.New("harness selection change metadata must be complete")
	}

	return validateHarnessSelectionAudit(*audit)
}

func validateHarnessSelectionAudit(audit harnesscore.HarnessSelectionAudit) error {
	if (workflowstate.MutationBinding{RequestID: audit.ProjectRef, Operator: audit.ChangedBy}).Validate() != nil || !workspacecore.ValidOperatorID(audit.ChangedBy) {
		return errors.New("invalid harness selection project or actor")
	}

	if (workflowstate.MutationBinding{RequestID: audit.RequestID, Operator: audit.ChangedBy}).Validate() != nil || !harnesscore.SelectionDigest(audit.RequestDigest) {
		return errors.New("invalid harness selection request metadata")
	}

	when, err := time.Parse(time.RFC3339Nano, audit.ChangedAt)
	if err != nil || when.Location() != time.UTC {
		return errors.New("harness selection timestamp must be UTC")
	}

	return nil
}
