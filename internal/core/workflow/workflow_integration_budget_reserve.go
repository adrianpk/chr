package workflow

import (
	"context"
	"reflect"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) reserveWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request IntegrationReserve) (IntegrationBudget, error) {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, request.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return budget, err
	}

	index := slices.IndexFunc(budget.Reservations, func(old IntegrationReservation) bool { return old.ID == request.Reservation.ID })
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

func checkWorkflowIntegrationReservationRetry(old, next IntegrationReservation) error {
	if !sameWorkflowIntegrationReservation(old, next) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowIntegrationReservationAdmission(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, budget IntegrationBudget, request IntegrationReserve) error {
	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
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

func checkWorkflowIntegrationCallDeadline(policy workflowstate.ExecutionPolicy, budget IntegrationBudget, reservation IntegrationReservation) error {
	if reservation.Kind != "agent_call" {
		return nil
	}

	deadline := reservation.Allowance.Deadline
	if reservation.Allowance == (workflowstate.TokenAllowance{}) {
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

func workflowIntegrationReservationStopped(run workflowInvocationFile, entry IntegrationEntry) bool {
	return run.StopIntent != "" || entry.StopIntent != "" || workflowIntegrationTerminal(entry.State) || entry.State == "paused" || entry.State == "blocked"
}

func workflowIntegrationReservationMatches(r IntegrationReservation, entry IntegrationEntry) bool {
	return r.EntryID == entry.ID && r.Unit == entry.Unit && r.ClaimGeneration == entry.ClaimGeneration && r.State == "reserved"
}

func chargeWorkflowIntegrationCounter(budget *IntegrationBudget, r IntegrationReservation) error {
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

	index := slices.IndexFunc(budget.UnitCounters, func(c IntegrationCounter) bool { return c.Unit == r.Unit })
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

func checkWorkflowIntegrationCounterCharge(budget IntegrationBudget, r IntegrationReservation, kind string) error {
	if kind == "drift" && workflowIntegrationBaseCharged(budget.Reservations, r) {
		return workflowstate.IntegrationConflict("replacement base already charged under another identity")
	}

	total := int64(0)

	for _, counter := range budget.UnitCounters {
		var err error

		total, err = workflowstate.TokenSum(total, workflowIntegrationCounterValue(counter, kind))
		if err != nil {
			return err
		}
	}

	if total >= workflowIntegrationLimitValue(budget.LimitsSnapshot, kind) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func workflowIntegrationCounterValue(counter IntegrationCounter, kind string) int64 {
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

func workflowIntegrationLimitValue(limits IntegrationLimits, kind string) int64 {
	return workflowIntegrationCounterValue(IntegrationCounter{Merger: limits.Merger, Drift: limits.Drift, Submission: limits.Submission}, kind)
}

// IntegrationGrantRequest requires explicit operator identity and reason.
// The budget owner supplies the prior revision and recording timestamp.
type IntegrationGrantRequest struct {
	IntegrationBudgetGuard
	Grant workflowstate.IntegrationGrant
}

// GrantWorkflowIntegration extends a ceiling without changing consumed history.
func (s *Store) GrantWorkflowIntegration(ctx context.Context, request IntegrationGrantRequest) (budget IntegrationBudget, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		budget, readErr = invocationLocation.Owner.grantWorkflowIntegration(ctx, invocationLocation.Run, request, nil)

		return readErr
	})

	return budget, err
}

func (s *Store) grantWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request IntegrationGrantRequest, control *workflowControlCommit) (IntegrationBudget, error) {
	budgetObservation, err := s.workflowGrantBudget(ctx, run, request.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return budget, err
	}

	grant := request.Grant
	if found, err := workflowIntegrationGrantRetry(budget, grant); found || err != nil {
		return budget, s.completeWorkflowControlRetry(ctx, run, control, err)
	}

	if err := guardWorkflowGrant(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
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

func applyWorkflowIntegrationGrant(budget *IntegrationBudget, grant workflowstate.IntegrationGrant) error {
	if grant.Kind != "tokens" {
		return addWorkflowIntegrationLimit(&budget.LimitsSnapshot, grant.Kind, grant.Amount)
	}

	if budget.TokenLimit == nil {
		return workflowstate.ErrConflict
	}

	limit, err := workflowstate.TokenSum(*budget.TokenLimit, grant.Amount)
	if err != nil {
		return err
	}

	budget.TokenLimit = &limit

	return nil
}

func checkWorkflowIntegrationRoleBudget(run workflowInvocationFile, budget IntegrationBudget, reservation IntegrationReservation) error {
	if reservation.Kind != "agent_call" || reservation.Role == "merger" {
		return nil
	}

	if workflowAgentAttemptReserved(budget, reservation) {
		return workflowstate.ErrConflict
	}

	roleUsage := workflowIntegrationRoleUsage(budget, reservation.Role)

	consumed, limit := roleUsage.Consumed, roleUsage.Limit

	consumed += workflowAgentUnrecordedCalls(run, budget, reservation.Role)

	if consumed > limit || consumed == limit && !workflowAgentUsageRecorded(run, reservation) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func workflowIntegrationRoleUsage(budget IntegrationBudget, role string) workflowRoleUsage {
	switch role {
	case "executor":
		return workflowRoleUsage{Consumed: int64(budget.Usage.WorkerCount), Limit: budget.LimitsSnapshot.Worker}
	case "reviewer":
		return workflowRoleUsage{Consumed: int64(budget.Usage.ReviewCount), Limit: budget.LimitsSnapshot.Review}
	case "closing":
		return workflowRoleUsage{Consumed: int64(len(budget.Usage.ClosingAttempts)), Limit: budget.LimitsSnapshot.Closing}
	default:
		return workflowRoleUsage{Consumed: 0, Limit: 0}
	}
}

func workflowIntegrationBaseCharged(reservations []IntegrationReservation, reservation IntegrationReservation) bool {
	return slices.ContainsFunc(reservations, func(old IntegrationReservation) bool {
		return old.Kind == "drift" && old.Unit == reservation.Unit && old.BaseCommit == reservation.BaseCommit
	})
}

func workflowAgentUsageRecorded(run workflowInvocationFile, r IntegrationReservation) bool {
	return slices.ContainsFunc(run.Attempts, func(a Attempt) bool {
		role := "executor"
		if a.Purpose == workflowSliceReviewPurpose {
			role = "reviewer"
		}

		return (a.Purpose == "" || a.Purpose == workflowSliceReviewPurpose) && role == r.Role && a.ID == r.AttemptID && a.Generation == r.ClaimGeneration
	})
}

func workflowAgentAttemptReserved(budget IntegrationBudget, r IntegrationReservation) bool {
	return slices.ContainsFunc(budget.Reservations, func(old IntegrationReservation) bool {
		return old.Kind == "agent_call" && old.Role == r.Role && old.AttemptID == r.AttemptID && old.ClaimGeneration == r.ClaimGeneration
	})
}

func workflowAgentUnrecordedCalls(run workflowInvocationFile, budget IntegrationBudget, role string) int64 {
	var count int64

	for _, old := range budget.Reservations {
		if old.Kind == "agent_call" && old.Role == role && !workflowAgentUsageRecorded(run, old) {
			count++
		}
	}

	return count
}

func workflowIntegrationGrantRetry(budget IntegrationBudget, grant workflowstate.IntegrationGrant) (bool, error) {
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
