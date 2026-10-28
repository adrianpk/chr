package core

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"strings"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Only a revision-guarded write can advance an inactive loose member's expected
// image. Raw edits and execution/description changes still require reconciliation.
func (s *WorkflowStore) workflowLooseWriteChanges(ctx context.Context, next WorkflowRecord, expected uint64, files []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
	if expected == 0 || next.Task == nil {
		return files, nil
	}

	old, err := s.readRecord(next.Path)
	if err != nil {
		return nil, err
	}

	if !workflowLooseDisplayWrite(old, next) {
		return files, nil
	}

	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return nil, err
	}

	next.SourceHash = workflowClaimRecordHash(next, files)
	for _, run := range runs {
		changed, err := s.workflowLooseWriteMember(&run, old, next)
		if err != nil {
			return nil, err
		}

		if !changed {
			continue
		}

		run.Revision++
		run.UpdatedAt = next.UpdatedAt

		change, err := workflowInvocationChange(run)
		if err != nil {
			return nil, err
		}

		files = append(files, change)
	}

	return files, nil
}

func (s *WorkflowStore) workflowLooseWriteMember(run *workflowInvocationFile, old, next WorkflowRecord) (bool, error) {
	if run.OwnerRoot != s.layout.Root() || run.Namespace != old.Namespace || run.Selection.Selector.Extent != "loose" || workflowInvocationTerminal(run.Status) {
		return false, nil
	}

	for i := run.Cursor; i < len(run.Members); i++ {
		member := &run.Members[i]
		if member.Admission.Key != old.Task.Key || run.Lane != nil && run.Lane.Unit.ID == old.Task.Key {
			continue
		}

		if !workflowMemberImageMatches(old, *member, run.WorkflowIdentity) {
			return false, workflowstate.ErrConflict
		}

		member.ExpectedRevision, member.ExpectedSourceHash = next.Revision, next.SourceHash

		return true, nil
	}

	return false, nil
}

func workflowLooseDisplayWrite(old, next WorkflowRecord) bool {
	if old.Task == nil || next.Task == nil || old.Task.SetID != "" || !slices.Contains([]string{"pending", "blocked"}, old.Task.Phase) {
		return false
	}

	before, err := workflowLooseExecutionImage(old)
	if err != nil {
		return false
	}

	after, err := workflowLooseExecutionImage(next)

	return err == nil && reflect.DeepEqual(before, after)
}

func workflowLooseExecutionImage(record WorkflowRecord) (map[string]json.RawMessage, error) {
	if strings.HasSuffix(record.Path, ".md") {
		data, err := encodeWorkflowMarkdown(record)
		if err != nil {
			return nil, err
		}

		body, err := workflowLooseExecutionBody(data)
		if err != nil {
			return nil, err
		}

		return workflowJSONFields(struct{ Image string }{body})
	}

	data, err := encodeWorkflowRecord(record)
	if err != nil {
		return nil, err
	}

	image := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &image); err != nil {
		return nil, err
	}

	delete(image, "revision")
	delete(image, "updated_at")

	fields, err := mergeWorkflowFields(record.Domain, *record.Task)
	if err != nil {
		return nil, err
	}

	delete(fields, "status")

	image["workflow"], err = json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	body, err := workflowLooseExecutionBody(record.Body)
	if err != nil {
		return nil, err
	}

	image["description_image"], err = json.Marshal(body)

	return image, err
}

func workflowLooseExecutionBody(body []byte) (string, error) {
	content := string(body)
	if !strings.HasPrefix(content, "---\n") {
		return content, nil
	}

	lines, err := ticketmetadata.FrontmatterLines(content)
	if err != nil {
		return "", err
	}

	filtered := []string{}
	skipping := false

	for _, line := range lines {
		key, _, scalar := strings.Cut(line, ":")
		if scalar && strings.TrimSpace(key) == key && key != "" {
			skipping = slices.Contains([]string{"status", "priority", "severity", "project_area", "type", "tags", "updated_at", "revision"}, key)
		}

		if !skipping {
			filtered = append(filtered, line)
		}
	}

	return ticketmetadata.ReplaceFrontmatterLines(content, filtered)
}
