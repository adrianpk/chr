package core

import (
	"fmt"
	"math"
	"reflect"
	"slices"
)

func workflowTokenSum(a, b int64) (int64, error) {
	if a < 0 || b < 0 || a > math.MaxInt64-b {
		return 0, workflowIntegrationConflict("token or counter arithmetic overflow")
	}

	return a + b, nil
}

func validateWorkflowIntegrationBudget(b WorkflowIntegrationBudget) error {
	if !validWorkflowIntegrationBudgetHeader(b) {
		return workflowIntegrationConflict("unsupported or invalid budget")
	}

	if err := validateWorkflowIntegrationGrants(b); err != nil {
		return err
	}

	if err := validateWorkflowIntegrationTokens(b); err != nil {
		return err
	}

	if err := validateWorkflowIntegrationUsage(b.Usage); err != nil {
		return err
	}

	return validateWorkflowIntegrationCounters(b)
}

func validWorkflowIntegrationBudgetHeader(b WorkflowIntegrationBudget) bool {
	return b.SchemaVersion == 1 && b.InvocationID != "" && b.Revision > 0 && len(b.UnitCounters) <= WorkflowRecordLimit && len(b.Reservations) <= WorkflowRecordLimit && len(b.Grants) <= WorkflowPageLimit
}

func validateWorkflowIntegrationTokens(b WorkflowIntegrationBudget) error {
	if b.TokenLimit != nil && *b.TokenLimit <= 0 {
		return ErrWorkflowConflict
	}

	settled, reserved, err := workflowIntegrationTokenTotals(b.Reservations)
	if err != nil || settled != b.TokensSettled || reserved != b.TokensReserved {
		return ErrWorkflowConflict
	}

	total, err := workflowTokenSum(settled, reserved)
	if err != nil || b.TokenLimit != nil && total > *b.TokenLimit {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func validateWorkflowIntegrationGrants(b WorkflowIntegrationBudget) error {
	limits := b.AdmissionLimits
	if !validWorkflowIntegrationLimits(limits) {
		return ErrWorkflowConflict
	}

	ids := map[string]bool{}
	for _, grant := range b.Grants {
		if !validWorkflowIntegrationGrant(grant, b.Revision) || ids[grant.ID] {
			return ErrWorkflowConflict
		}

		ids[grant.ID] = true
		if grant.Kind == "tokens" {
			continue
		}

		if err := addWorkflowIntegrationLimit(&limits, grant.Kind, grant.Amount); err != nil {
			return err
		}
	}

	if limits != b.LimitsSnapshot {
		return ErrWorkflowConflict
	}

	return validateWorkflowIntegrationTokenGrants(b)
}

func validWorkflowIntegrationGrant(grant WorkflowIntegrationGrant, revision uint64) bool {
	return validWorkflowRequestKey(grant.ID) && grant.Amount > 0 && grant.PriorRevision < revision && grant.PriorRevision > 0 && !grant.RecordedAt.IsZero() && validWorkflowRequestKey(grant.Actor) && validWorkflowRequestKey(grant.Reason)
}

func validateWorkflowIntegrationTokenGrants(b WorkflowIntegrationBudget) error {
	if b.AdmissionTokenLimit == nil {
		if b.TokenLimit != nil || slices.ContainsFunc(b.Grants, func(g WorkflowIntegrationGrant) bool { return g.Kind == "tokens" }) {
			return ErrWorkflowConflict
		}

		return nil
	}

	limit := *b.AdmissionTokenLimit
	if limit <= 0 {
		return ErrWorkflowConflict
	}

	for _, grant := range b.Grants {
		if grant.Kind != "tokens" {
			continue
		}

		var err error

		limit, err = workflowTokenSum(limit, grant.Amount)
		if err != nil {
			return err
		}
	}

	if b.TokenLimit == nil || *b.TokenLimit != limit {
		return ErrWorkflowConflict
	}

	return nil
}

func addWorkflowIntegrationLimit(limits *WorkflowIntegrationLimits, kind string, amount int64) error {
	var counter *int64

	switch kind {
	case "executor":
		counter = &limits.Worker
	case "reviewer":
		counter = &limits.Review
	case "closing":
		counter = &limits.Closing
	case "merger":
		counter = &limits.Merger
	case "drift":
		counter = &limits.Drift
	case "submission":
		counter = &limits.Submission
	default:
		return ErrWorkflowConflict
	}

	next, err := workflowTokenSum(*counter, amount)
	if err != nil {
		return err
	}

	*counter = next

	return nil
}

func workflowIntegrationTokenTotals(reservations []WorkflowIntegrationReservation) (settled, reserved int64, err error) {
	ids := map[string]bool{}
	for _, r := range reservations {
		if ids[r.ID] {
			return 0, 0, ErrWorkflowConflict
		}

		ids[r.ID] = true
		if err := validateWorkflowIntegrationReservation(r); err != nil {
			return 0, 0, err
		}

		switch r.State {
		case "settled":
			usage, sumErr := workflowTokenSum(r.ReportedInput, r.ReportedOutput)
			if sumErr != nil || usage > r.Allowance.Maximum {
				return 0, 0, ErrWorkflowConflict
			}

			settled, err = workflowTokenSum(settled, usage)
		case "reserved", "launched", "unknown":
			reserved, err = workflowTokenSum(reserved, r.Allowance.Maximum)
		}

		if err != nil {
			return 0, 0, err
		}
	}

	return settled, reserved, nil
}

func validateWorkflowIntegrationReservation(r WorkflowIntegrationReservation) error {
	if !validWorkflowIntegrationReservationOwner(r) {
		return ErrWorkflowConflict
	}

	if r.Kind != "agent_call" {
		return validateWorkflowIntegrationAction(r)
	}

	if !validWorkflowTokenAllowance(r.Allowance) || !slices.Contains([]string{"executor", "merger", "reviewer", "closing"}, r.Role) {
		return ErrWorkflowConflict
	}

	if r.State == "settled" {
		if r.SettlementProof == nil {
			return ErrWorkflowConflict
		}

		return nil
	}

	if r.ReportedInput != 0 || r.ReportedOutput != 0 || r.SettlementProof != nil {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowIntegrationReservationOwner(r WorkflowIntegrationReservation) bool {
	return validWorkflowRequestKey(r.ID) && validWorkflowNativeFingerprint(r.EntryID) && validWorkflowRequestKey(r.AttemptID) && r.Unit.ID != "" && slices.Contains([]string{"reserved", "launched", "unknown", "settled", "released_before_launch"}, r.State)
}

func validWorkflowTokenAllowance(a WorkflowTokenAllowance) bool {
	maximum, err := workflowTokenSum(a.Input, a.Output)

	return err == nil && maximum > 0 && a.Maximum == maximum && a.WholeCall && validWorkflowNativeFingerprint(a.CapabilityRevision) && !a.Deadline.IsZero()
}

func validateWorkflowIntegrationAction(r WorkflowIntegrationReservation) error {
	if !slices.Contains([]string{"drift", "submission"}, r.Kind) || r.Role != "" || r.State != "reserved" || r.Allowance != (WorkflowTokenAllowance{}) {
		return ErrWorkflowConflict
	}

	if r.SettlementProof != nil || r.ReportedInput != 0 || r.ReportedOutput != 0 {
		return ErrWorkflowConflict
	}

	if !validWorkflowIntegrationActionBase(r) {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowIntegrationActionBase(r WorkflowIntegrationReservation) bool {
	if r.Kind == "drift" {
		return workflowCommitID(r.BaseCommit)
	}

	return r.BaseCommit == ""
}

func validateWorkflowIntegrationCounters(b WorkflowIntegrationBudget) error {
	units := map[WorkflowRef]bool{}
	for _, counter := range b.UnitCounters {
		if !validWorkflowIntegrationCounterUnit(counter.Unit) || units[counter.Unit] {
			return ErrWorkflowConflict
		}

		units[counter.Unit] = true
		if err := checkWorkflowIntegrationCounterReservations(counter, b.Reservations); err != nil {
			return err
		}

		if counter.Merger < 0 || counter.Drift < 0 || counter.Submission < 0 {
			return ErrWorkflowConflict
		}
	}

	for _, reservation := range b.Reservations {
		if !units[reservation.Unit] {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func guardWorkflowIntegrationBudget(run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, guard WorkflowIntegrationBudgetGuard) error {
	expected := [4]uint64{run.Revision, budget.Revision, entry.Revision, entry.ClaimGeneration}

	actual := [4]uint64{guard.InvocationRevision, guard.BudgetRevision, guard.EntryRevision, guard.ClaimGeneration}
	if run.ID != guard.InvocationID || entry.ID != guard.EntryID || expected != actual || budget.Revision == math.MaxUint64 || run.Revision == math.MaxUint64 {
		return workflowIntegrationConflict("budget callback has a stale owner or exhausted revision")
	}

	return nil
}

func sameWorkflowIntegrationReservation(a, b WorkflowIntegrationReservation) bool {
	a.State, b.State = "", ""
	a.ReportedInput, a.ReportedOutput, a.SettlementProof = 0, 0, nil
	b.ReportedInput, b.ReportedOutput, b.SettlementProof = 0, 0, nil

	return reflect.DeepEqual(a, b)
}

func checkWorkflowIntegrationUsage(before, now WorkflowIntegrationUsage) error {
	if now.WorkerCount < before.WorkerCount || now.ReviewCount < before.ReviewCount || now.PreparationCount < before.PreparationCount || !workflowUsageIncludes(now.Attempts, before.Attempts) || !workflowUsageIncludes(now.ClosingAttempts, before.ClosingAttempts) {
		return workflowIntegrationConflict("existing consumed counters were removed")
	}

	for _, old := range before.ReviewRepairs {
		index := slices.IndexFunc(now.ReviewRepairs, func(current WorkflowReviewRepairBudget) bool { return current.Unit == old.Unit })
		if index < 0 || now.ReviewRepairs[index].Consumed < old.Consumed {
			return fmt.Errorf("%w: existing review repair consumption decreased", ErrWorkflowConflict)
		}
	}

	return nil
}

func workflowUsageIncludes(now, before []string) bool {
	for _, id := range before {
		if !slices.Contains(now, id) {
			return false
		}
	}

	return true
}

func checkWorkflowIntegrationCounterReservations(counter WorkflowIntegrationCounter, reservations []WorkflowIntegrationReservation) error {
	expected := WorkflowIntegrationCounter{Unit: counter.Unit, ExistingSubmissions: counter.ExistingSubmissions, Submission: counter.ExistingSubmissions}
	if counter.ExistingSubmissions < 0 {
		return ErrWorkflowConflict
	}

	for _, reservation := range reservations {
		if reservation.Unit != counter.Unit {
			continue
		}

		switch reservation.Kind {
		case "drift":
			expected.Drift++
		case "submission":
			expected.Submission++
		case "agent_call":
			if reservation.Role == "merger" {
				expected.Merger++
			}
		}
	}

	if counter != expected {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowIntegrationLimits(limits WorkflowIntegrationLimits) bool {
	return limits.Merger > 0 && limits.Drift > 0 && limits.Submission > 0 && limits.Worker > 0 && limits.Review > 0 && limits.Closing > 0
}

func validateWorkflowIntegrationUsage(usage WorkflowIntegrationUsage) error {
	if usage.WorkerCount < 0 || usage.ReviewCount < 0 || usage.PreparationCount < 0 || len(usage.Attempts) > WorkflowRecordLimit || len(usage.ClosingAttempts) > WorkflowRecordLimit || len(usage.ReviewRepairs) > WorkflowRecordLimit {
		return ErrWorkflowConflict
	}

	for _, repair := range usage.ReviewRepairs {
		if repair.Consumed < 0 {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validWorkflowIntegrationCounterUnit(unit WorkflowRef) bool {
	return unit.ID != "" && slices.Contains([]string{"slice", "ticket"}, unit.Kind)
}
