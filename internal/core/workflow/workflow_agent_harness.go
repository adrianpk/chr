package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const (
	workflowAgentCallProtocol = "chr-agent-call/v1"
	workflowTokenAccounting   = "input-output-tokens/v1"
)

// AgentHarness supplies a host-certified whole-call driver and finite
// per-call input/output grants. The invocation owns the total ceiling.
type AgentHarness struct {
	Command      []string
	Certificate  workflowstate.EvidenceRef
	InputTokens  int64
	OutputTokens int64
}

// AgentHarnessCertificate attests enforcement over the entire call,
// including internal tool turns. Cached input is included in Input; Output
// includes reasoning. Output-only meters and incompatible units cannot qualify.
type AgentHarnessCertificate struct {
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

// AgentCall seals a reservation and certified driver into task input.
type AgentCall struct {
	ReservationID string                       `json:"reservation_id"`
	Role          string                       `json:"role"`
	HarnessRef    workflowstate.EvidenceRef    `json:"harness_ref"`
	Allowance     workflowstate.TokenAllowance `json:"allowance"`
}

// AgentResponse is emitted by the trusted driver, separately from its
// model's Result. Missing usage remains unknown; zero usage must be explicit.
type AgentResponse struct {
	Protocol           string `json:"protocol"`
	ReservationID      string `json:"reservation_id"`
	AttemptID          string `json:"attempt_id"`
	CapabilityRevision string `json:"capability_revision"`
	WholeCall          bool   `json:"whole_call"`
	Input              *int64 `json:"input"`
	Output             *int64 `json:"output"`
	Result             string `json:"result"`
}

func (s *Store) checkWorkflowAgentHarness(ctx context.Context, id, role string, harness AgentHarness) (AgentHarnessCertificate, error) {
	var certificate AgentHarnessCertificate
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

func checkWorkflowAgentCertificate(c AgentHarnessCertificate, role string, command []string) error {
	if c.SchemaVersion != 1 || c.Protocol != workflowAgentCallProtocol || !slices.Equal(c.Command, command) || !validWorkflowNativeFingerprint(c.CapabilityRevision) || !slices.Contains(c.Roles, role) {
		return fmt.Errorf("%w: certified role driver required", ErrWorkflowIntegrationBudget)
	}

	if !c.EnforcesWholeCall || !c.ReportsWholeCall || c.Accounting != workflowTokenAccounting {
		return fmt.Errorf("%w: compatible whole-call enforcement and reporting required", ErrWorkflowIntegrationBudget)
	}

	return nil
}
