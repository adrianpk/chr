package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// GeneratedWorkflowPlan is explicit approved materialization input. Its request
// identity survives retries; the store allocates the set envelope and ticket keys.
type GeneratedWorkflowPlan struct {
	PlanRevision      *PlanRevisionBinding     `json:"plan_revision,omitempty"`
	MaterializationID string                   `json:"materialization_id"`
	Title             string                   `json:"title"`
	PlanRef           string                   `json:"plan_ref"`
	PlanHash          string                   `json:"plan_hash"`
	UseCase           string                   `json:"use_case,omitempty"`
	UseCaseOrder      int                      `json:"use_case_order,omitempty"`
	ClosingValidation WorkflowValidation       `json:"closing_validation"`
	Slices            []GeneratedWorkflowSlice `json:"slices"`
}

// GeneratedWorkflowSlice has stable identity and set-local order, independent of
// its display label. Its tasks form one integration unit.
type GeneratedWorkflowSlice struct {
	ID         string                  `json:"id"`
	Label      string                  `json:"label"`
	Order      int                     `json:"slice_order"`
	Title      string                  `json:"title"`
	Contract   string                  `json:"contract"`
	Validation WorkflowValidation      `json:"validation"`
	DependsOn  []WorkflowRef           `json:"depends_on,omitempty"`
	Tasks      []GeneratedWorkflowTask `json:"tasks"`
}

// GeneratedWorkflowTask is one task contract, not a slice ticket. Ticket
// dependencies may name input task IDs; materialization resolves them to keys.
type GeneratedWorkflowTask struct {
	ID         string             `json:"id"`
	Label      string             `json:"label"`
	Order      int                `json:"task_order"`
	Title      string             `json:"title"`
	Body       string             `json:"body"`
	Outcome    string             `json:"outcome"`
	Scope      string             `json:"scope"`
	Acceptance []string           `json:"acceptance"`
	Validation WorkflowValidation `json:"validation"`
	DependsOn  []WorkflowRef      `json:"depends_on,omitempty"`
	Priority   string             `json:"priority"`
	Severity   string             `json:"severity"`
	Type       string             `json:"type"`
	Tags       []string           `json:"tags,omitempty"`
}

var workflowInputIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// ValidateGeneratedWorkflowPlan rejects partial or ambiguous hierarchy before IO.
func ValidateGeneratedWorkflowPlan(plan GeneratedWorkflowPlan) error {
	if err := validateWorkflowPlanIdentity(plan); err != nil {
		return err
	}

	if err := validateWorkflowGate(plan.ClosingValidation); err != nil {
		return err
	}

	if len(plan.Slices) == 0 || len(plan.Slices) > WorkflowPageLimit {
		return fmt.Errorf("workflow slice bound exceeded or empty plan")
	}

	if err := validateWorkflowPlanHierarchy(plan.Slices); err != nil {
		return err
	}

	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}

	if len(data) > WorkflowFileLimit {
		return fmt.Errorf("workflow materialization input exceeds 1 MiB")
	}

	return nil
}

func validateWorkflowPlanIdentity(plan GeneratedWorkflowPlan) error {
	if !workflowInputIDPattern.MatchString(plan.MaterializationID) || strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("workflow materialization identity and title are required")
	}

	if !filepath.IsLocal(plan.PlanRef) || filepath.Clean(plan.PlanRef) != plan.PlanRef || !strings.HasPrefix(plan.PlanRef, "plan/") || plan.PlanHash == "" {
		return fmt.Errorf("workflow plan reference and observed hash are required")
	}

	if plan.UseCaseOrder < 0 || (plan.UseCaseOrder > 0 && plan.UseCase == "") {
		return fmt.Errorf("workflow use-case order requires its stable association")
	}

	return nil
}

func validateWorkflowPlanHierarchy(slices []GeneratedWorkflowSlice) error {
	ids, orders, tasks := map[string]bool{}, map[int]bool{}, map[string]bool{}
	for _, slice := range slices {
		if !workflowInputIDPattern.MatchString(slice.ID) || slice.ID == slice.Label || ids[slice.ID] || slice.Order < 1 || orders[slice.Order] {
			return fmt.Errorf("invalid or duplicate stable slice identity/order: %s", slice.ID)
		}

		ids[slice.ID], orders[slice.Order] = true, true
		if err := validateWorkflowPlanSlice(slice, tasks); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowPlanSlice(slice GeneratedWorkflowSlice, tasks map[string]bool) error {
	if err := validateWorkflowPlanSliceContract(slice); err != nil {
		return err
	}

	if err := validateWorkflowGate(slice.Validation); err != nil {
		return err
	}

	orders := map[int]bool{}
	for _, task := range slice.Tasks {
		if !workflowInputIDPattern.MatchString(task.ID) || task.ID == task.Label || tasks[task.ID] || task.Order < 1 || orders[task.Order] {
			return fmt.Errorf("invalid or duplicate stable task identity/order: %s", task.ID)
		}

		tasks[task.ID], orders[task.Order] = true, true
		if err := validateWorkflowPlanTask(task); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowPlanSliceContract(slice GeneratedWorkflowSlice) error {
	if strings.TrimSpace(slice.Title) == "" || strings.TrimSpace(slice.Contract) == "" || len(slice.Tasks) == 0 || len(slice.Tasks) > WorkflowPageLimit {
		return fmt.Errorf("slice %s requires a contract and bounded nonempty tasks", slice.ID)
	}

	return nil
}

func validateWorkflowPlanTask(task GeneratedWorkflowTask) error {
	for _, value := range []string{task.Title, task.Body, task.Outcome, task.Scope} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("task %s requires title, body, outcome and scope", task.ID)
		}
	}

	if len(task.Acceptance) == 0 || len(task.Acceptance) > WorkflowPageLimit {
		return fmt.Errorf("task %s requires bounded acceptance", task.ID)
	}

	for _, criterion := range task.Acceptance {
		if strings.TrimSpace(criterion) == "" {
			return fmt.Errorf("empty task acceptance criterion")
		}
	}

	return validateWorkflowGate(task.Validation)
}

// WorkflowContentHash binds an approved source image without deriving identity.
func WorkflowContentHash(data []byte) string { return contenthash.Sum(data) }
