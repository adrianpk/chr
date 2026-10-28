package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MergerHarness identifies a trusted model driver, not an unrestricted
// worker CLI. Its retained certificate must establish that model tool calls use
// this broker exclusively; the driver itself is part of the host trust boundary.
type MergerHarness struct {
	Command     []string
	Certificate workflowstate.EvidenceRef
}

// MergerHarnessCertificate binds an externally certified driver image
// and arguments to its tool and accounting contracts. Agent output cannot supply
// this certificate or enable built-in filesystem, shell, network or MCP tools.
type MergerHarnessCertificate struct {
	Accounting           string   `json:"accounting,omitempty"`
	SchemaVersion        int      `json:"schema_version"`
	Protocol             string   `json:"protocol"`
	Command              []string `json:"command"`
	ExecutableHash       string   `json:"executable_hash"`
	CapabilityRevision   string   `json:"capability_revision"`
	BuiltinToolsDisabled bool     `json:"builtin_tools_disabled"`
	EnforcesWholeCall    bool     `json:"enforces_whole_call"`
	ReportsWholeCall     bool     `json:"reports_whole_call"`
}

// MergerRequest carries host-owned admission evidence. Runtime identity,
// capacity, remaining attempts and deadline are supplied again by the host.
type MergerRequest struct {
	IntegrationBudgetGuard
	Input workflowstate.MergerInput
}

// MergerAttempt retains a proposal independently of candidate approval.
type MergerAttempt struct {
	Acceptance    *MergerAcceptance             `json:"acceptance,omitempty"`
	SchemaVersion int                           `json:"schema_version"`
	Input         workflowstate.MergerInput     `json:"input"`
	HarnessRef    workflowstate.EvidenceRef     `json:"harness_ref"`
	Harness       workflowstate.CommandResult   `json:"harness"`
	Proposal      *workflowstate.MergerResult   `json:"proposal"`
	Checks        []workflowstate.CommandResult `json:"checks"`
	TrustedUsage  *workflowstate.MergerUsage    `json:"trusted_usage,omitempty"`
	Failure       string                        `json:"failure,omitempty"`
}

func (s *Store) checkWorkflowMergerHarness(ctx context.Context, input workflowstate.MergerInput, harness MergerHarness) (MergerHarnessCertificate, error) {
	var certificate MergerHarnessCertificate
	if err := ctx.Err(); err != nil {
		return certificate, err
	}

	if err := s.readWorkflowNativeEvidence(input.InvocationID, harness.Certificate, &certificate); err != nil {
		return certificate, err
	}

	if err := checkWorkflowMergerCertificate(certificate, harness.Command); err != nil {
		return certificate, err
	}

	if err := checkWorkflowMergerExecutable(harness.Command, certificate.ExecutableHash); err != nil {
		return certificate, err
	}

	if err := checkWorkflowMergerTokenCapability(input, certificate); err != nil {
		return certificate, err
	}

	return certificate, nil
}

func checkWorkflowMergerExecutable(command []string, hash string) error {
	if !workflowmerge.ValidCommand(command) || !filepath.IsAbs(command[0]) || !workflowstate.ValidFingerprint(hash) {
		return workflowstate.ErrConflict
	}

	path, err := filepath.EvalSymlinks(command[0])
	if err != nil || path != command[0] {
		return workflowstate.ErrConflict
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }() //nolint:errcheck

	if err := checkWorkflowMergerExecutableMode(file); err != nil {
		return err
	}

	actual, err := workflowMergerFileHash(file)
	if err != nil || actual != hash {
		return fmt.Errorf("%w: certified merger driver changed", workflowstate.ErrConflict)
	}

	return nil
}

func encodeWorkflowMergerAttempt(attempt MergerAttempt) ([]byte, error) {
	data, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, workflowstate.ErrConflict
	}

	return append(data, '\n'), nil
}

func checkWorkflowMergerCertificate(certificate MergerHarnessCertificate, command []string) error {
	if certificate.SchemaVersion != 1 || certificate.Protocol != workflowmerge.BrokerProtocol || !certificate.BuiltinToolsDisabled || !slices.Equal(certificate.Command, command) || !workflowstate.ValidFingerprint(certificate.CapabilityRevision) {
		return fmt.Errorf("%w: certified broker-only merger driver required", workflowstate.ErrConflict)
	}

	return nil
}

func checkWorkflowMergerTokenCapability(input workflowstate.MergerInput, certificate MergerHarnessCertificate) error {
	if allowance := input.TokenAllowance; allowance != nil && (!certificate.EnforcesWholeCall || !certificate.ReportsWholeCall || certificate.Accounting != workflowTokenAccounting || allowance.CapabilityRevision != certificate.CapabilityRevision) {
		return fmt.Errorf("%w: whole-call token enforcement required", ErrWorkflowIntegrationBudget)
	}

	return nil
}

func checkWorkflowMergerExecutableMode(file *os.File) error {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return workflowstate.ErrConflict
	}

	return nil
}
