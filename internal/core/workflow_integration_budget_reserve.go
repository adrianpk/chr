package core

import (
	"context"
	"reflect"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) reserveWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationReserve) (WorkflowIntegrationBudget, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return budget, err
	}

	index := slices.IndexFunc(budget.Reservations, func(old WorkflowIntegrationReservation) bool { return old.ID == request.Reservation.ID })
	if index >= 0 {
		return budget, checkWorkflowIntegrationReservationRetry(budget.Reservations[index], request.Reservation)
	}

	if err := s.checkWorkflowIntegrationReservationAdmission(ctx, run, entry, budget, request); err != nil {
		return budget, err
	}

	if err := checkWorkflowIntegrationRoleBudget(run, budget, request.Reservation); err != nil {
		return budget, err
	}

	if err := chargeWorkflowIntegrationCounter(&budget, request.Reservation); err != nil {
		return budget, err
	}

	budget.Reservations = append(budget.Reservations, request.Reservation)

	return budget, s.commitWorkflowIntegrationBudget(ctx, &run, &budget)
}

func checkWorkflowIntegrationReservationRetry(old, next WorkflowIntegrationReservation) error {
	if !sameWorkflowIntegrationReservation(old, next) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationReservationAdmission(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, request WorkflowIntegrationReserve) error {
	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return err
	}

	if workflowIntegrationReservationStopped(run, entry) {
		return ErrWorkflowStop
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	r := request.Reservation
	if !workflowIntegrationReservationMatches(r, entry) || len(budget.Reservations) >= workflowstate.RecordLimit {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowIntegrationReservation(r); err != nil {
		return err
	}

	return checkWorkflowIntegrationCallDeadline(run.Policy, budget, r)
}

func checkWorkflowIntegrationCallDeadline(policy WorkflowExecutionPolicy, budget WorkflowIntegrationBudget, reservation WorkflowIntegrationReservation) error {
	if reservation.Kind != "agent_call" {
		return nil
	}

	deadline := reservation.Allowance.Deadline
	if reservation.Allowance == (WorkflowTokenAllowance{}) {
		deadline = reservation.RuntimeDeadline

		if budget.TokenLimit != nil {
			return ErrWorkflowIntegrationBudget
		}
	}

	now := time.Now().UTC()
	if !deadline.After(now) || deadline.After(now.Add(time.Duration(policy.WorkerTimeoutSeconds)*time.Second)) {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowIntegrationReservationStopped(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return run.StopIntent != "" || entry.StopIntent != "" || workflowIntegrationTerminal(entry.State) || entry.State == "paused" || entry.State == "blocked"
}

func workflowIntegrationReservationMatches(r WorkflowIntegrationReservation, entry WorkflowIntegrationEntry) bool {
	return r.EntryID == entry.ID && r.Unit == entry.Unit && r.ClaimGeneration == entry.ClaimGeneration && r.State == "reserved"
}

func chargeWorkflowIntegrationCounter(budget *WorkflowIntegrationBudget, r WorkflowIntegrationReservation) error {
	kind := r.Kind
	if kind == "agent_call" {
		kind = r.Role
	}

	if !slices.Contains([]string{"merger", "drift", "submission"}, kind) {
		return nil
	}

	if err := checkWorkflowIntegrationCounterCharge(*budget, r, kind); err != nil {
		return err
	}

	index := slices.IndexFunc(budget.UnitCounters, func(c WorkflowIntegrationCounter) bool { return c.Unit == r.Unit })
	if index < 0 {
		return workflowstate.ErrConflict
	}

	counter := &budget.UnitCounters[index]

	switch kind {
	case "merger":
		counter.Merger++
	case "drift":
		counter.Drift++
	case "submission":
		counter.Submission++
	}

	return nil
}

func checkWorkflowIntegrationCounterCharge(budget WorkflowIntegrationBudget, r WorkflowIntegrationReservation, kind string) error {
	if kind == "drift" && workflowIntegrationBaseCharged(budget.Reservations, r) {
		return workflowIntegrationConflict("replacement base already charged under another identity")
	}

	total := int64(0)

	for _, counter := range budget.UnitCounters {
		var err error

		total, err = workflowTokenSum(total, workflowIntegrationCounterValue(counter, kind))
		if err != nil {
			return err
		}
	}

	if total >= workflowIntegrationLimitValue(budget.LimitsSnapshot, kind) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func workflowIntegrationCounterValue(counter WorkflowIntegrationCounter, kind string) int64 {
	switch kind {
	case "merger":
		return counter.Merger
	case "drift":
		return counter.Drift
	case "submission":
		return counter.Submission
	default:
		return 0
	}
}

func workflowIntegrationLimitValue(limits WorkflowIntegrationLimits, kind string) int64 {
	return workflowIntegrationCounterValue(WorkflowIntegrationCounter{Merger: limits.Merger, Drift: limits.Drift, Submission: limits.Submission}, kind)
}

// WorkflowIntegrationGrantRequest requires explicit operator identity and reason.
// The budget owner supplies the prior revision and recording timestamp.
type WorkflowIntegrationGrantRequest struct {
	WorkflowIntegrationBudgetGuard
	Grant workflowstate.WorkflowIntegrationGrant
}

// GrantWorkflowIntegration extends a ceiling without changing consumed history.
func (s *WorkflowStore) GrantWorkflowIntegration(ctx context.Context, request WorkflowIntegrationGrantRequest) (budget WorkflowIntegrationBudget, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		budget, readErr = owner.grantWorkflowIntegration(ctx, run, request, nil)

		return readErr
	})

	return budget, err
}

func (s *WorkflowStore) grantWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationGrantRequest, control *workflowControlCommit) (WorkflowIntegrationBudget, error) {
	entry, budget, err := s.workflowGrantBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return budget, err
	}

	grant := request.Grant
	if found, err := workflowIntegrationGrantRetry(budget, grant); found || err != nil {
		return budget, s.completeWorkflowControlRetry(ctx, run, control, err)
	}

	if err := guardWorkflowGrant(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return budget, err
	}

	if err := s.checkWorkflowGrantInputs(ctx, run); err != nil {
		return budget, err
	}

	grant.PriorRevision, grant.RecordedAt = budget.Revision, time.Now().UTC()
	if !validWorkflowIntegrationGrant(grant, budget.Revision+1) || len(budget.Grants) >= workflowstate.PageLimit {
		return budget, workflowstate.ErrConflict
	}

	if err := applyWorkflowIntegrationGrant(&budget, grant); err != nil {
		return budget, err
	}

	if err := boundedWorkflowIntegrationGrant(budget.LimitsSnapshot); err != nil {
		return budget, err
	}

	budget.Grants = append(budget.Grants, grant)

	return budget, s.commitWorkflowControlBudget(ctx, &run, &budget, control)
}

func applyWorkflowIntegrationGrant(budget *WorkflowIntegrationBudget, grant workflowstate.WorkflowIntegrationGrant) error {
	if grant.Kind != "tokens" {
		return addWorkflowIntegrationLimit(&budget.LimitsSnapshot, grant.Kind, grant.Amount)
	}

	if budget.TokenLimit == nil {
		return workflowstate.ErrConflict
	}

	limit, err := workflowTokenSum(*budget.TokenLimit, grant.Amount)
	if err != nil {
		return err
	}

	budget.TokenLimit = &limit

	return nil
}

func checkWorkflowIntegrationRoleBudget(run workflowInvocationFile, budget WorkflowIntegrationBudget, reservation WorkflowIntegrationReservation) error {
	if reservation.Kind != "agent_call" || reservation.Role == "merger" {
		return nil
	}

	if workflowAgentAttemptReserved(budget, reservation) {
		return workflowstate.ErrConflict
	}

	consumed, limit := workflowIntegrationRoleUsage(budget, reservation.Role)
	consumed += workflowAgentUnrecordedCalls(run, budget, reservation.Role)

	if consumed > limit || consumed == limit && !workflowAgentUsageRecorded(run, reservation) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func workflowIntegrationRoleUsage(budget WorkflowIntegrationBudget, role string) (int64, int64) {
	switch role {
	case "executor":
		return int64(budget.Usage.WorkerCount), budget.LimitsSnapshot.Worker
	case "reviewer":
		return int64(budget.Usage.ReviewCount), budget.LimitsSnapshot.Review
	case "closing":
		return int64(len(budget.Usage.ClosingAttempts)), budget.LimitsSnapshot.Closing
	default:
		return 0, 0
	}
}

func workflowIntegrationBaseCharged(reservations []WorkflowIntegrationReservation, reservation WorkflowIntegrationReservation) bool {
	return slices.ContainsFunc(reservations, func(old WorkflowIntegrationReservation) bool {
		return old.Kind == "drift" && old.Unit == reservation.Unit && old.BaseCommit == reservation.BaseCommit
	})
}

func workflowAgentUsageRecorded(run workflowInvocationFile, r WorkflowIntegrationReservation) bool {
	return slices.ContainsFunc(run.Attempts, func(a WorkflowAttempt) bool {
		role := "executor"
		if a.Purpose == workflowSliceReviewPurpose {
			role = "reviewer"
		}

		return (a.Purpose == "" || a.Purpose == workflowSliceReviewPurpose) && role == r.Role && a.ID == r.AttemptID && a.Generation == r.ClaimGeneration
	})
}

func workflowAgentAttemptReserved(budget WorkflowIntegrationBudget, r WorkflowIntegrationReservation) bool {
	return slices.ContainsFunc(budget.Reservations, func(old WorkflowIntegrationReservation) bool {
		return old.Kind == "agent_call" && old.Role == r.Role && old.AttemptID == r.AttemptID && old.ClaimGeneration == r.ClaimGeneration
	})
}

func workflowAgentUnrecordedCalls(run workflowInvocationFile, budget WorkflowIntegrationBudget, role string) int64 {
	var count int64

	for _, old := range budget.Reservations {
		if old.Kind == "agent_call" && old.Role == role && !workflowAgentUsageRecorded(run, old) {
			count++
		}
	}

	return count
}

func workflowIntegrationGrantRetry(budget WorkflowIntegrationBudget, grant workflowstate.WorkflowIntegrationGrant) (bool, error) {
	for _, old := range budget.Grants {
		if old.ID != grant.ID {
			continue
		}

		grant.PriorRevision, grant.RecordedAt = old.PriorRevision, old.RecordedAt
		if !reflect.DeepEqual(old, grant) {
			return false, workflowstate.ErrConflict
		}

		return true, nil
	}

	return false, nil
}
