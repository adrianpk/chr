package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMarkdownFields(front string) (map[string]json.RawMessage, error) {
	fields := map[string]json.RawMessage{}

	for _, line := range strings.Split(front, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		value = strings.TrimSpace(value)

		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate workflow frontmatter key: %s", key)
		}

		if json.Valid([]byte(value)) {
			fields[key] = json.RawMessage(value)
		} else {
			fields[key] = json.RawMessage(strconv.Quote(opsmetadata.Unquote(value)))
		}
	}

	return fields, nil
}

func (s *Store) readMarkdown(relative, path string) (Record, error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return Record{}, err
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(data)
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		return Record{}, err
	}

	if !strings.Contains("\n"+front, "\nworkflow_schema:") {
		return Record{}, workflowstate.ErrNotWorkflow
	}

	fields, err := workflowMarkdownFields(front)
	if err != nil {
		return Record{}, err
	}

	if _, ok := fields["workflow_schema"]; !ok {
		return Record{}, workflowstate.ErrNotWorkflow
	}

	r := Record{Kind: "ticket", Path: relative, Body: body, Metadata: map[string]json.RawMessage{"frontmatter": json.RawMessage(strconv.Quote(front))}, SourceHash: contenthash.Sum(data)}

	encoded, err := json.Marshal(fields)
	if err != nil {
		return Record{}, err
	}

	if err := json.Unmarshal(encoded, &r.Identity); err != nil {
		return Record{}, err
	}

	if err := decodeWorkflowDomain(&r, encoded); err != nil {
		return Record{}, err
	}

	if r.ID == "" {
		r.ID = r.Task.Key
	}

	return r, r.Validate()
}

func encodeWorkflowMarkdown(r Record) ([]byte, error) {
	var front string
	if raw, ok := r.Metadata["frontmatter"]; ok {
		if err := json.Unmarshal(raw, &front); err != nil {
			return nil, err
		}
	}

	fields, err := workflowJSONFields(workflowDomainValue(r))
	if err != nil {
		return nil, err
	}

	identity, err := workflowJSONFields(r.Identity)
	if err != nil {
		return nil, err
	}

	for key, value := range identity {
		fields[key] = value
	}

	return []byte("---\n" + patchWorkflowFrontmatter(front, fields, r) + "\n---\n" + string(r.Body)), nil
}

func patchWorkflowFrontmatter(front string, fields map[string]json.RawMessage, r Record) string {
	owned := map[string]bool{}

	for _, value := range []any{r.Identity, workflowDomainValue(r)} {
		typeOf := reflect.TypeOf(value)
		for i := 0; i < typeOf.NumField(); i++ {
			owned[strings.Split(typeOf.Field(i).Tag.Get("json"), ",")[0]] = true
		}
	}

	lines := []string{}

	for _, line := range strings.Split(front, "\n") {
		key, _, ok := strings.Cut(line, ":")
		if ok && owned[key] {
			if raw, exists := fields[key]; exists {
				lines = append(lines, key+": "+string(raw))
				delete(fields, key)
			}

			continue
		}

		lines = append(lines, line)
	}

	keys := []string{}
	for key := range fields {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		lines = append(lines, key+": "+string(fields[key]))
	}

	return strings.Join(lines, "\n")
}

// Adopt explicitly binds an existing ticket while preserving its original location.
func (s *Store) Adopt(ctx context.Context, path string, task Task, observedHash string) (adopted Record, err error) {
	err = s.withLock(ctx, func() error {
		record, readErr := s.prepareAdoption(path, task, observedHash)
		if readErr != nil {
			return readErr
		}

		adopted, readErr = s.writeRecord(ctx, record, 0)

		return readErr
	})

	return adopted, err
}

func (s *Store) prepareAdoption(relative string, task Task, observedHash string) (Record, error) {
	path, err := s.safePath(relative)
	if err != nil {
		return Record{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return Record{}, err
	}

	if contenthash.Sum(data) != observedHash {
		return Record{}, workflowstate.ErrConflict
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(data)
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		return Record{}, err
	}

	fields, err := workflowMarkdownFields(front)
	if err != nil {
		return Record{}, err
	}

	var key, created string
	if err := json.Unmarshal(fields["key"], &key); err != nil || key != task.Key {
		return Record{}, fmt.Errorf("adoption must preserve ticket key")
	}

	if err := json.Unmarshal(fields["created_at"], &created); err != nil {
		return Record{}, err
	}

	instant, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return Record{}, err
	}

	return Record{Identity: workflowstate.Identity{ID: key, ProjectID: s.projectID, Namespace: s.layout.Namespace(), Revision: 1, CreatedAt: instant, UpdatedAt: time.Now().UTC()}, Kind: "ticket", Path: relative, Body: body, Task: &task, Metadata: map[string]json.RawMessage{"frontmatter": json.RawMessage(strconv.Quote(front))}, SourceHash: observedHash}, nil
}
