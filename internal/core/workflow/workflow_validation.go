package workflow

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Validate rejects partial or inconsistent workflow records before persistence.
func (r Record) Validate() error {
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

func (r Record) validateIdentity() error {
	if r.ID == "" || r.ProjectID == "" || !opsnamespace.Valid(r.Namespace) || r.Revision == 0 {
		return fmt.Errorf("workflow identity, scope and positive revision are required")
	}

	if r.CreatedAt.IsZero() || r.UpdatedAt.Before(r.CreatedAt) {
		return fmt.Errorf("workflow creation and update times are invalid")
	}

	return nil
}

func validateWorkflowRecordBounds(r Record) error {
	if len(r.Body) > workflowstate.FileLimit || r.Revision > math.MaxInt64 {
		return fmt.Errorf("workflow record size or revision bound exceeded")
	}

	data, err := json.Marshal(workflowDomainValue(r))
	if err != nil {
		return err
	}

	if len(data) > workflowstate.FileLimit {
		return fmt.Errorf("workflow record exceeds 1 MiB")
	}

	return nil
}

func (r Record) validateDomain() error {
	switch {
	case r.Kind == "ticket" && r.Task != nil:
		return validateWorkflowTask(*r.Task)
	case r.Kind == "plan" && r.Set != nil:
		if err := validateWorkflowPlanRevision(r.Namespace, *r.Set); err != nil {
			return err
		}

		return validateWorkflowSet(*r.Set)
	case r.Kind == "receipt" && r.Receipt != nil:
		return validateWorkflowReceipt(*r.Receipt)
	}

	return fmt.Errorf("workflow record requires exactly one matching domain: %s", r.Kind)
}

func validateWorkflowTask(t Task) error {
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

func validateWorkflowTaskPhase(t Task) error {
	allowed := map[string][]string{
		"pending":       {string(ticketcore.StatusBacklog), string(ticketcore.StatusReadyForDev)},
		"running":       {string(ticketcore.StatusInProgress)},
		"validated":     {string(ticketcore.StatusInProgress)},
		"waiting_slice": {string(ticketcore.StatusInProgress), string(ticketcore.StatusReadyForReview), string(ticketcore.StatusReadyForMerge)},
		"blocked":       {string(ticketcore.StatusBlockedLegacy)},
		"integrated":    {string(ticketcore.StatusDone)},
	}
	if !slices.Contains(allowed[t.Phase], t.Status) {
		return fmt.Errorf("invalid workflow task phase/status: %s/%s", t.Phase, t.Status)
	}

	return validateWorkflowTaskEvidence(t)
}

func validateWorkflowTaskEvidence(t Task) error {
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

func validateWorkflowSet(s workflowstate.WorkflowSet) error {
	if err := workflowcloseout.ValidateWorkflowClosingBinding(s); err != nil {
		return err
	}

	if err := validateWorkflowClosureRefs(s.ClosureRef, s.ClosureHistory); err != nil {
		return err
	}

	if err := validateWorkflowAmendmentRefs(s.Amendments); err != nil {
		return err
	}

	if err := workflowstate.ValidateGate(s.ClosingValidation); err != nil {
		return err
	}

	if err := validateWorkflowSetContract(s); err != nil {
		return err
	}

	return validateWorkflowSlices(s.Slices)
}

func validateWorkflowSetContract(s workflowstate.WorkflowSet) error {
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

func validateWorkflowSlices(descriptors []workflowstate.WorkflowSlice) error {
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

func validateWorkflowSlice(s workflowstate.WorkflowSlice) error {
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

func validateWorkflowDependencies(refs []workflowstate.Ref) error {
	if len(refs) > workflowstate.PageLimit {
		return fmt.Errorf("workflow dependency bound exceeded")
	}

	seen := map[workflowstate.Ref]bool{}
	for _, ref := range refs {
		if ref.ID == "" || !slices.Contains([]string{"ticket", "slice"}, ref.Kind) || seen[ref] {
			return fmt.Errorf("invalid or duplicate workflow dependency")
		}

		seen[ref] = true
	}

	return nil
}

func validateWorkflowReceipt(r Receipt) error {
	if err := validateWorkflowReceiptSchema(r); err != nil {
		return err
	}

	if r.Unit.ID == "" || !slices.Contains([]string{"ticket", "slice"}, r.Unit.Kind) {
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
