package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) prepareLegacyWorkflowPlan(opts WorkflowPlanAdoptionOptions) (WorkflowRecord, error) {
	if err := validateLegacyWorkflowPlanPath(opts); err != nil {
		return WorkflowRecord{}, err
	}

	path, err := s.safePath(opts.PlanPath)
	if err != nil {
		return WorkflowRecord{}, err
	}

	record, err := readWorkflowEnvelope(path)
	if err != nil {
		return record, err
	}

	if !s.legacyWorkflowPlanScope(record) {
		return record, fmt.Errorf("foreign legacy plan identity")
	}

	if _, adopted := record.Metadata["workflow"]; adopted {
		return record, workflowstate.ErrConflict
	}

	body, err := s.readApprovedPlan(opts.Plan)
	if err != nil {
		return record, err
	}

	record.SourceHash += ":" + contenthash.Sum(body)
	if record.SourceHash != opts.SourceHash {
		return record, workflowstate.ErrConflict
	}

	record.ProjectID, record.Path, record.Body = s.projectID, opts.PlanPath, body
	record.Revision++
	record.UpdatedAt = time.Now().UTC()

	if err := validateLegacyWorkflowPlanState(record.Metadata, body, opts.Plan.UseCase); err != nil {
		return record, err
	}

	return record, nil
}

func (s *WorkflowStore) prepareLegacyWorkflowTasks(plan generationcore.GeneratedWorkflowPlan, selections []WorkflowAdoptionTask) ([]WorkflowRecord, []string, error) {
	byID := map[string]WorkflowAdoptionTask{}

	paths := map[string]bool{}
	for _, selection := range selections {
		if byID[selection.TaskID].Path != "" || paths[selection.Path] || selection.SourceHash == "" {
			return nil, nil, fmt.Errorf("ambiguous legacy task selection")
		}

		byID[selection.TaskID], paths[selection.Path] = selection, true
	}

	records := []WorkflowRecord{}
	keys := []string{}

	for _, slice := range plan.Slices {
		for _, task := range slice.Tasks {
			selection, ok := byID[task.ID]
			if !ok {
				return nil, nil, fmt.Errorf("missing selected legacy task: %s", task.ID)
			}

			record, err := s.prepareLegacyWorkflowTicket(selection, task.Body, plan.UseCase)
			if err != nil {
				return nil, nil, err
			}

			records = append(records, record)
			keys = append(keys, record.Task.Key)
		}
	}

	if len(records) != len(selections) {
		return nil, nil, fmt.Errorf("legacy selection contains tasks outside the approved plan")
	}

	return records, keys, nil
}

func (s *WorkflowStore) prepareLegacyWorkflowTicket(selection WorkflowAdoptionTask, body, useCase string) (WorkflowRecord, error) {
	if !strings.HasPrefix(selection.Path, "ticket/") {
		return WorkflowRecord{}, fmt.Errorf("selected ticket must belong to the namespace")
	}

	path, err := s.safePath(selection.Path)
	if err != nil {
		return WorkflowRecord{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return WorkflowRecord{}, err
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(data)
	front, original := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		return WorkflowRecord{}, err
	}

	fields, err := workflowMarkdownFields(front)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if err := s.checkLegacyWorkflowTicket(fields, original, body, useCase); err != nil {
		return WorkflowRecord{}, err
	}

	var key, status string
	if err := json.Unmarshal(fields["key"], &key); err != nil {
		return WorkflowRecord{}, err
	}

	if err := json.Unmarshal(fields["status"], &status); err != nil {
		return WorkflowRecord{}, err
	}

	task := WorkflowTask{Schema: WorkflowSchema, Key: key, Status: status, Phase: "pending", ContractHash: "adoption-pending"}

	return s.prepareAdoption(selection.Path, task, selection.SourceHash)
}

func (s *WorkflowStore) checkLegacyWorkflowTicket(fields map[string]json.RawMessage, body []byte, approvedBody, useCase string) error {
	for _, field := range []string{"workflow_schema", "active_attempt_id", "claim_generation"} {
		if _, exists := fields[field]; exists {
			return fmt.Errorf("selected ticket already has workflow ownership or requires reconciliation")
		}
	}

	if err := s.checkLegacyWorkflowScope(fields, useCase); err != nil {
		return err
	}

	var status string
	if err := json.Unmarshal(fields["status"], &status); err != nil {
		return err
	}

	if status != TicketStatusReadyForDev && status != TicketStatusBacklog {
		return fmt.Errorf("legacy task status requires explicit evidence reconciliation: %s", status)
	}

	if string(body) != approvedBody {
		return fmt.Errorf("legacy body differs from the approved task input")
	}

	return nil
}

func (s *WorkflowStore) bindLegacyWorkflowTasks(templates, selected []WorkflowRecord) ([]WorkflowRecord, error) {
	for i := range templates {
		record := selected[i]
		task := *templates[i].Task
		task.Status = record.Task.Status
		record.Task = &task
		ref := filepath.ToSlash(filepath.Join("plan", task.SetID, "extension", "chr", "task-"+strconv.Itoa(i+1)+".json"))

		var front string
		if err := json.Unmarshal(record.Metadata["frontmatter"], &front); err != nil {
			return nil, err
		}

		record.Metadata["frontmatter"] = json.RawMessage(strconv.Quote(front + "\nexecution_contract_ref: " + strconv.Quote(ref)))
		templates[i] = record
	}

	return templates, nil
}

func (s *WorkflowStore) checkLegacyWorkflowKeys(ctx context.Context, selected []WorkflowRecord) error {
	owners := map[string]string{}
	for _, record := range selected {
		if owners[record.Task.Key] != "" {
			return fmt.Errorf("ambiguous selected legacy key")
		}

		owners[record.Task.Key] = record.Path
	}

	count := 0

	return filepath.WalkDir(filepath.Join(s.layout.Root(), "ticket"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		count++
		if count > workflowstate.RecordLimit {
			return fmt.Errorf("legacy adoption scan bound exceeded")
		}

		return s.checkLegacyWorkflowKey(path, owners)
	})
}

func (s *WorkflowStore) checkLegacyWorkflowKey(path string, owners map[string]string) error {
	data, err := opsfile.Read(path)
	if err != nil {
		return err
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(data)
	front := parsedMarkdown.Front

	if errors.Is(err, workflowstate.ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	fields, err := workflowMarkdownFields(front)
	if err != nil {
		return err
	}

	var key string
	if raw, ok := fields["key"]; ok {
		if err := json.Unmarshal(raw, &key); err != nil {
			return err
		}
	}

	relative, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	if expected := owners[key]; expected != "" && expected != filepath.ToSlash(relative) {
		return fmt.Errorf("ambiguous legacy ticket representation: %s", key)
	}

	return nil
}

func validateLegacyWorkflowPlanPath(opts WorkflowPlanAdoptionOptions) error {
	if !strings.HasPrefix(opts.PlanPath, "plan/") || filepath.Base(opts.PlanPath) != "metadata.json" || opts.Plan.PlanRef != workflowAdoptionPlanBody(opts.PlanPath) {
		return fmt.Errorf("adoption requires the selected plan's canonical envelope and original body")
	}

	return nil
}

func (s *WorkflowStore) checkLegacyWorkflowScope(fields map[string]json.RawMessage, useCase string) error {
	for field, want := range map[string]string{"project_id": s.projectID, "namespace": s.layout.Namespace()} {
		var observed string
		if raw, ok := fields[field]; ok {
			if err := json.Unmarshal(raw, &observed); err != nil || observed != want {
				return fmt.Errorf("foreign legacy ticket scope")
			}
		}
	}

	return validateLegacyWorkflowUseCase(fields, useCase)
}

func (s *WorkflowStore) legacyWorkflowPlanScope(record WorkflowRecord) bool {
	return record.Kind == "plan" && record.Namespace == s.layout.Namespace() && (record.ProjectID == "" || record.ProjectID == s.projectID)
}

func validateLegacyWorkflowUseCase(fields map[string]json.RawMessage, expected string) error {
	for _, field := range []string{"use_case", "case_slug"} {
		var observed string
		if raw, ok := fields[field]; ok {
			if err := json.Unmarshal(raw, &observed); err != nil {
				return err
			}

			if observed != "" && observed != expected {
				return fmt.Errorf("foreign legacy use-case association")
			}
		}
	}

	return nil
}

func validateLegacyWorkflowPlanState(metadata map[string]json.RawMessage, body []byte, useCase string) error {
	var status string
	if err := json.Unmarshal(metadata["status"], &status); err != nil {
		return err
	}

	switch status {
	case "running", "closing", "completed", "done", "delivered":
		return fmt.Errorf("legacy plan status requires explicit evidence reconciliation")
	}

	if err := validateLegacyWorkflowUseCase(metadata, useCase); err != nil {
		return err
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(body)
	front := parsedMarkdown.Front

	if errors.Is(err, workflowstate.ErrNotWorkflow) {
		return nil
	}

	if err != nil {
		return err
	}

	fields, err := workflowMarkdownFields(front)
	if err != nil {
		return err
	}

	return validateLegacyWorkflowUseCase(fields, useCase)
}
