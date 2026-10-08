package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

func readWorkflowEnvelope(path string) (WorkflowRecord, error) {
	data, err := readWorkflowFile(path)
	if err != nil {
		return WorkflowRecord{}, err
	}

	r := WorkflowRecord{Metadata: map[string]json.RawMessage{}}
	if err := json.Unmarshal(data, &r.Metadata); err != nil {
		return WorkflowRecord{}, err
	}

	if _, ok := r.Metadata["type"]; !ok {
		return WorkflowRecord{}, ErrNotWorkflow
	}

	if err := json.Unmarshal(data, &r.WorkflowIdentity); err != nil {
		return WorkflowRecord{}, err
	}

	if err := json.Unmarshal(r.Metadata["type"], &r.Kind); err != nil {
		return WorkflowRecord{}, err
	}

	r.SourceHash = workflowHash(data)

	return r, nil
}

func (s *WorkflowStore) readRecord(relative string) (WorkflowRecord, error) {
	path, err := s.safePath(relative)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if strings.HasSuffix(path, ".md") {
		return s.readMarkdown(relative, path)
	}

	r, err := readWorkflowEnvelope(path)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if _, ok := r.Metadata["workflow"]; !ok {
		return WorkflowRecord{}, ErrNotWorkflow
	}

	var schema string
	if err := json.Unmarshal(r.Metadata["schema_version"], &schema); err != nil || schema != "ops-layout/v1" {
		return WorkflowRecord{}, fmt.Errorf("unsupported workflow envelope schema")
	}

	if err := decodeWorkflowDomain(&r, r.Metadata["workflow"]); err != nil {
		return WorkflowRecord{}, err
	}

	if filepath.Base(path) == "metadata.json" {
		r.Body, err = readWorkflowFile(filepath.Join(filepath.Dir(path), "artifact.md"))
		if err != nil {
			return WorkflowRecord{}, err
		}

		r.SourceHash += ":" + workflowHash(r.Body)
	}

	r.Path = relative

	return r, r.Validate()
}

func decodeWorkflowDomain(r *WorkflowRecord, data []byte) error {
	r.Domain = map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &r.Domain); err != nil {
		return err
	}

	switch r.Kind {
	case "ticket":
		r.Task = &WorkflowTask{}

		return json.Unmarshal(data, r.Task)
	case "plan":
		r.Set = &WorkflowSet{}

		return json.Unmarshal(data, r.Set)
	case "receipt":
		r.Receipt = &WorkflowReceipt{}

		return json.Unmarshal(data, r.Receipt)
	}

	return fmt.Errorf("unsupported workflow record kind: %s", r.Kind)
}

func workflowDomainValue(r WorkflowRecord) any {
	if r.Task != nil {
		return *r.Task
	}

	if r.Set != nil {
		return *r.Set
	}

	return *r.Receipt
}

func workflowJSONFields(value any) (map[string]json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

func mergeWorkflowFields(old map[string]json.RawMessage, value any) (map[string]json.RawMessage, error) {
	fields, err := workflowJSONFields(value)
	if err != nil {
		return nil, err
	}

	result := map[string]json.RawMessage{}
	for key, raw := range old {
		result[key] = raw
	}

	typeOf := reflect.TypeOf(value)
	for i := 0; i < typeOf.NumField(); i++ {
		key := strings.Split(typeOf.Field(i).Tag.Get("json"), ",")[0]
		delete(result, key)
	}

	for key, raw := range fields {
		result[key] = raw
	}

	return result, nil
}

func encodeWorkflowRecord(r WorkflowRecord) ([]byte, error) {
	if strings.HasSuffix(r.Path, ".md") {
		return encodeWorkflowMarkdown(r)
	}

	metadata, err := mergeWorkflowFields(r.Metadata, r.WorkflowIdentity)
	if err != nil {
		return nil, err
	}

	domain, err := mergeWorkflowFields(r.Domain, workflowDomainValue(r))
	if err != nil {
		return nil, err
	}

	if err := preserveWorkflowSetFields(domain, r); err != nil {
		return nil, err
	}

	metadata["workflow"], err = json.Marshal(domain)
	if err != nil {
		return nil, err
	}

	metadata["type"] = json.RawMessage(strconv.Quote(r.Kind))
	metadata["schema_version"] = json.RawMessage(`"ops-layout/v1"`)

	data, err := json.MarshalIndent(metadata, "", "  ")

	return append(data, '\n'), err
}

func preserveWorkflowSetFields(domain map[string]json.RawMessage, r WorkflowRecord) error {
	if r.Set == nil {
		return nil
	}

	gateFields := map[string]json.RawMessage{}
	if raw, ok := r.Domain["closing_validation"]; ok {
		if err := json.Unmarshal(raw, &gateFields); err != nil {
			return err
		}
	}

	merged, err := mergeWorkflowFields(gateFields, r.Set.ClosingValidation)
	if err != nil {
		return err
	}

	domain["closing_validation"], err = json.Marshal(merged)
	if err != nil {
		return err
	}

	return preserveWorkflowSliceFields(domain, r)
}

func preserveWorkflowSliceFields(domain map[string]json.RawMessage, r WorkflowRecord) error {
	oldSlices := []map[string]json.RawMessage{}
	if raw, ok := r.Domain["slices"]; ok {
		if err := json.Unmarshal(raw, &oldSlices); err != nil {
			return err
		}
	}

	byID := map[string]map[string]json.RawMessage{}
	for _, fields := range oldSlices {
		byID[string(fields["id"])] = fields
	}

	merged := []map[string]json.RawMessage{}

	for _, descriptor := range r.Set.Slices {
		key := strconv.Quote(descriptor.ID)

		fields, err := mergeWorkflowFields(byID[string(key)], descriptor)
		if err != nil {
			return err
		}

		merged = append(merged, fields)
	}

	data, err := json.Marshal(merged)
	domain["slices"] = data

	return err
}
