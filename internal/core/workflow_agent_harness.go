package core

import (
	"context"
	"fmt"
	"slices"
)

const (
	workflowAgentCallProtocol = "chr-agent-call/v1"
	workflowTokenAccounting   = "input-output-tokens/v1"
)

// WorkflowAgentHarness supplies a host-certified whole-call driver and finite
// per-call input/output grants. The invocation owns the total ceiling.
type WorkflowAgentHarness struct {
	Command      []string
	Certificate  WorkflowEvidenceRef
	InputTokens  int64
	OutputTokens int64
}

// WorkflowAgentHarnessCertificate attests enforcement over the entire call,
// including internal tool turns. Cached input is included in Input; Output
// includes reasoning. Output-only meters and incompatible units cannot qualify.
type WorkflowAgentHarnessCertificate struct {
	SchemaVersion      int      `json:"schema_version"`
	Protocol           string   `json:"protocol"`
	Command            []string `json:"command"`
	ExecutableHash     string   `json:"executable_hash"`
	CapabilityRevision string   `json:"capability_revision"`
	Roles              []string `json:"roles"`
	Accounting         string   `json:"accounting"`
	EnforcesWholeCall  bool     `json:"enforces_whole_call"`
	ReportsWholeCall   bool     `json:"reports_whole_call"`
}

// WorkflowAgentCall seals a reservation and certified driver into task input.
type WorkflowAgentCall struct {
	ReservationID string                 `json:"reservation_id"`
	Role          string                 `json:"role"`
	HarnessRef    WorkflowEvidenceRef    `json:"harness_ref"`
	Allowance     WorkflowTokenAllowance `json:"allowance"`
}

// WorkflowAgentResponse is emitted by the trusted driver, separately from its
// model's Result. Missing usage remains unknown; zero usage must be explicit.
type WorkflowAgentResponse struct {
	Protocol           string `json:"protocol"`
	ReservationID      string `json:"reservation_id"`
	AttemptID          string `json:"attempt_id"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
	Input              *int64 `json:"input"`
	Output             *int64 `json:"output"`
	Result             string `json:"result"`
}

func (s *WorkflowStore) checkWorkflowAgentHarness(ctx context.Context, id, role string, harness WorkflowAgentHarness) (WorkflowAgentHarnessCertificate, error) {
	var certificate WorkflowAgentHarnessCertificate
	if err := ctx.Err(); err != nil {
		return certificate, err
	}

	if err := s.readWorkflowNativeEvidence(id, harness.Certificate, &certificate); err != nil {
		return certificate, err
	}

	if err := checkWorkflowAgentCertificate(certificate, role, harness.Command); err != nil {
		return certificate, err
	}

	return certificate, checkWorkflowMergerExecutable(harness.Command, certificate.ExecutableHash)
}

func checkWorkflowAgentCertificate(c WorkflowAgentHarnessCertificate, role string, command []string) error {
	if c.SchemaVersion != 1 || c.Protocol != workflowAgentCallProtocol || !slices.Equal(c.Command, command) || !validWorkflowNativeFingerprint(c.CapabilityRevision) || !slices.Contains(c.Roles, role) {
		return fmt.Errorf("%w: certified role driver required", ErrWorkflowIntegrationBudget)
	}

	if !c.EnforcesWholeCall || !c.ReportsWholeCall || c.Accounting != workflowTokenAccounting {
		return fmt.Errorf("%w: compatible whole-call enforcement and reporting required", ErrWorkflowIntegrationBudget)
	}

	return nil
}
