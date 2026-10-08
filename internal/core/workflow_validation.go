package core

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
)

// Validate rejects partial or inconsistent workflow records before persistence.
func (r WorkflowRecord) Validate() error {
	if err := r.validateIdentity(); err != nil {
		return err
	}

	domains := 0

	for _, present := range []bool{r.Task != nil, r.Set != nil, r.Receipt != nil} {
		if present {
			domains++
		}
	}

	if domains != 1 {
		return fmt.Errorf("workflow record requires exactly one domain")
	}

	if err := validateWorkflowRecordBounds(r); err != nil {
		return err
	}

	return r.validateDomain()
}

func (r WorkflowRecord) validateIdentity() error {
	if r.ID == "" || r.ProjectID == "" || !IsValidOpsNamespace(r.Namespace) || r.Revision == 0 {
		return fmt.Errorf("workflow identity, scope and positive revision are required")
	}

	if r.CreatedAt.IsZero() || r.UpdatedAt.Before(r.CreatedAt) {
		return fmt.Errorf("workflow creation and update times are invalid")
	}

	return nil
}

func validateWorkflowRecordBounds(r WorkflowRecord) error {
	if len(r.Body) > WorkflowFileLimit || r.Revision > math.MaxInt64 {
		return fmt.Errorf("workflow record size or revision bound exceeded")
	}

	data, err := json.Marshal(workflowDomainValue(r))
	if err != nil {
		return err
	}

	if len(data) > WorkflowFileLimit {
		return fmt.Errorf("workflow record exceeds 1 MiB")
	}

	return nil
}

func (r WorkflowRecord) validateDomain() error {
	switch {
	case r.Kind == "ticket" && r.Task != nil:
		return validateWorkflowTask(*r.Task)
	case r.Kind == "plan" && r.Set != nil:
		return validateWorkflowSet(*r.Set)
	case r.Kind == "receipt" && r.Receipt != nil:
		return validateWorkflowReceipt(*r.Receipt)
	}

	return fmt.Errorf("workflow record requires exactly one matching domain: %s", r.Kind)
}

func validateWorkflowTask(t WorkflowTask) error {
	if t.Schema != WorkflowSchema || t.Key == "" || t.ContractHash == "" {
		return fmt.Errorf("unsupported workflow schema or missing ticket key/contract")
	}

	if (t.SetID == "") != (t.SliceID == "") || (t.SliceID == "") != (t.Order == 0) || t.Order < 0 {
		return fmt.Errorf("workflow task association must be all absent or all present")
	}

	if err := validateWorkflowDependencies(t.DependsOn); err != nil {
		return err
	}

	return validateWorkflowTaskPhase(t)
}

func validateWorkflowTaskPhase(t WorkflowTask) error {
	allowed := map[string][]string{
		"pending":       {TicketStatusBacklog, TicketStatusReadyForDev},
		"running":       {TicketStatusInProgress},
		"validated":     {TicketStatusInProgress},
		"waiting_slice": {TicketStatusInProgress, TicketStatusReadyForReview, TicketStatusReadyForMerge},
		"blocked":       {TicketStatusBlockedLegacy},
		"integrated":    {TicketStatusDone},
	}
	if !slices.Contains(allowed[t.Phase], t.Status) {
		return fmt.Errorf("invalid workflow task phase/status: %s/%s", t.Phase, t.Status)
	}

	return validateWorkflowTaskEvidence(t)
}

func validateWorkflowTaskEvidence(t WorkflowTask) error {
	if t.Phase == "running" && (t.AttemptID == "" || t.ClaimGeneration == 0) {
		return fmt.Errorf("running task requires fenced ownership")
	}

	if slices.Contains([]string{"validated", "waiting_slice"}, t.Phase) && len(t.Evidence) == 0 {
		return fmt.Errorf("validated task requires evidence")
	}

	if t.Phase == "integrated" && t.DeliveryRef == "" {
		return fmt.Errorf("integrated task requires receipt")
	}

	return validateWorkflowTaskDelivery(t)
}

func validateWorkflowSet(s WorkflowSet) error {
	if err := validateWorkflowGate(s.ClosingValidation); err != nil {
		return err
	}

	if err := validateWorkflowSetContract(s); err != nil {
		return err
	}

	return validateWorkflowSlices(s.Slices)
}

func validateWorkflowSetContract(s WorkflowSet) error {
	if s.Schema != WorkflowSchema || s.PlanRef == "" || s.PlanHash == "" || s.UseCaseOrder < 0 {
		return fmt.Errorf("unsupported workflow set schema or missing plan contract")
	}

	if s.UseCaseOrder > 0 && s.UseCase == "" {
		return fmt.Errorf("workflow set order requires a use case")
	}

	if !slices.Contains([]string{"proposed", "ready", "running", "closing", "completed", "blocked"}, s.Status) {
		return fmt.Errorf("invalid workflow set status: %s", s.Status)
	}

	if s.Status == "completed" && s.ClosureRef == "" {
		return fmt.Errorf("completed workflow set requires closure evidence")
	}

	return nil
}

func validateWorkflowSlices(descriptors []WorkflowSlice) error {
	ids, orders := map[string]bool{}, map[int]bool{}
	for _, slice := range descriptors {
		if ids[slice.ID] || orders[slice.Order] {
			return fmt.Errorf("duplicate workflow slice identity/order")
		}

		ids[slice.ID], orders[slice.Order] = true, true

		if err := validateWorkflowSlice(slice); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowGate(g WorkflowValidation) error {
	if !g.Required {
		if strings.TrimSpace(g.WaiverReason) == "" {
			return fmt.Errorf("waived workflow gate requires a reason")
		}

		return nil
	}

	if len(g.Commands) == 0 || len(g.Commands) > 16 || g.TimeoutSeconds < 1 || g.TimeoutSeconds > 86400 {
		return fmt.Errorf("workflow gate command/deadline bounds exceeded")
	}

	seen := map[string]bool{}

	for _, argv := range g.Commands {
		key := strings.Join(argv, "\x00")
		if err := validateWorkflowCommand(argv); err != nil {
			return err
		}

		if seen[key] {
			return fmt.Errorf("duplicate workflow validation command")
		}

		seen[key] = true
	}

	return nil
}

func validateWorkflowCommand(argv []string) error {
	if len(argv) == 0 || len(argv) > 64 || argv[0] == "" || len(strings.Join(argv, " ")) > 4096 || strings.Contains(strings.Join(argv, ""), "\x00") {
		return fmt.Errorf("invalid workflow validation command")
	}

	return nil
}

func validateWorkflowSlice(s WorkflowSlice) error {
	if s.ID == "" || s.Order < 1 || s.ContractRef == "" || s.ContractHash == "" || s.TaskInputHash == "" {
		return fmt.Errorf("workflow slice identity/order/contract are required")
	}

	if !slices.Contains([]string{"pending", "developing", "validating", "reviewing", "ready_for_merge", "merging", "awaiting_observation", "integrated", "blocked"}, s.Phase) {
		return fmt.Errorf("invalid workflow slice phase: %s", s.Phase)
	}

	if s.Phase == "integrated" && s.DeliveryRef == "" {
		return fmt.Errorf("integrated workflow slice requires receipt")
	}

	return validateWorkflowDependencies(s.DependsOn)
}

func validateWorkflowDependencies(refs []WorkflowRef) error {
	if len(refs) > WorkflowPageLimit {
		return fmt.Errorf("workflow dependency bound exceeded")
	}

	seen := map[WorkflowRef]bool{}
	for _, ref := range refs {
		if ref.ID == "" || !slices.Contains([]string{"ticket", "slice"}, ref.Kind) || seen[ref] {
			return fmt.Errorf("invalid or duplicate workflow dependency")
		}

		seen[ref] = true
	}

	return nil
}

func validateWorkflowReceipt(r WorkflowReceipt) error {
	if r.Schema != WorkflowSchema || r.Unit.ID == "" || !slices.Contains([]string{"ticket", "slice"}, r.Unit.Kind) {
		return fmt.Errorf("unsupported workflow receipt schema or unit")
	}

	for _, value := range []string{r.CandidateID, r.SubmissionID, r.Provider, r.Repository, r.PullRequest, r.SourceCommit, r.BaseCommit, r.PreparedTree, r.HistoryProof, r.ResultCommit, r.VerifiedTree, r.ObservedTarget, r.VerificationRef} {
		if value == "" {
			return fmt.Errorf("workflow receipt integration proof is incomplete")
		}
	}

	if !slices.Contains([]string{"merge", "rebase", "squash"}, r.Method) || r.ConfirmedAt.IsZero() || len(r.ValidationRefs) == 0 || r.PreparedTree != r.VerifiedTree {
		return fmt.Errorf("workflow receipt method/confirmation/tree proof is invalid")
	}

	return nil
}
