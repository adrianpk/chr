package app

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadRecords observes native records and use-case metadata through their existing owners.
func (s service) ReadRecords(ctx context.Context, namespace string, query RecordQuery) (RecordPage, error) {
	if _, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace}); err != nil {
		return RecordPage{}, err
	}

	if query.Offset < 0 || query.Offset > workflowstate.RecordLimit || query.Limit < 0 || query.Limit > 50 {
		return RecordPage{}, fmt.Errorf("invalid record page bounds")
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	records, err := s.recordViews(ctx, namespace, query)
	if err != nil {
		return RecordPage{}, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].ParentID+"/"+records[i].Identity < records[j].ParentID+"/"+records[j].Identity
	})

	start := min(query.Offset, len(records))
	end := min(start+query.Limit, len(records))

	return RecordPage{Records: records[start:end], NextOffset: end, HasMore: end < len(records)}, nil
}

func (s service) recordViews(ctx context.Context, namespace string, query RecordQuery) ([]RecordView, error) {
	if query.Identity != "" && !validRecordIdentity(query.Identity) {
		return nil, fmt.Errorf("invalid record identity")
	}

	if query.Kind == RecordUseCase {
		return s.caseRecordViews(ctx, query)
	}

	store, err := s.scopedRecordStore(namespace)
	if err != nil {
		return nil, err
	}

	snapshot, err := store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}

	return nativeRecordViews(snapshot, query)
}

func (s service) scopedRecordStore(namespace string) (*workflowcore.Store, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace})
	if err != nil {
		return nil, err
	}

	return workflowcore.NewWorkflowStore(layout, s.projectContext.ProjectID), nil
}

func nativeRecordViews(snapshot workflowstate.Snapshot, query RecordQuery) ([]RecordView, error) {
	if !nativeRecordKind(query.Kind) {
		return nil, ErrRecordUnsupported
	}

	var result []RecordView

	for _, record := range snapshot.Records {
		if query.Kind == RecordTicket && record.Task != nil {
			result = append(result, nativeRecordView(record, RecordTicket))
		}

		if query.Kind == RecordSliceSet && record.Set != nil {
			result = append(result, nativeRecordView(record, RecordSliceSet))
		}

		if query.Kind == RecordSlice && record.Set != nil {
			result = append(result, sliceRecordViews(record)...)
		}
	}

	return filterRecordViews(result, query)
}

func nativeRecordView(record workflowstate.Record, kind RecordKind) RecordView {
	view := RecordView{Kind: kind, Identity: record.ID, Revision: record.Revision, SourceHash: record.SourceHash, Reference: record.Path, Body: string(record.Body)}
	if err := json.Unmarshal(record.Metadata["title"], &view.Title); err != nil {
		view.Title = record.ID
	}

	if record.Task != nil {
		view.Identity, view.Phase = record.Task.Key, record.Task.Phase
	}

	if record.Set != nil {
		view.Phase = record.Set.Status
	}

	return view
}

func sliceRecordViews(record workflowstate.Record) []RecordView {
	var views []RecordView
	for _, slice := range record.Set.Slices {
		views = append(views, RecordView{Kind: RecordSlice, Identity: slice.ID, ParentID: record.ID, Title: slice.Title, Phase: slice.Phase, Revision: record.Revision, SourceHash: record.SourceHash, Reference: record.Path + "#" + slice.ID})
	}

	return views
}

func filterRecordViews(views []RecordView, query RecordQuery) ([]RecordView, error) {
	var result []RecordView

	for _, view := range views {
		if query.Identity != "" && view.Identity != query.Identity {
			continue
		}

		if query.ParentID != "" && view.ParentID != query.ParentID {
			continue
		}

		if query.Identity == "" {
			view.Body = ""
		}

		result = append(result, view)
	}

	if query.Identity != "" && len(result) != 1 {
		return nil, fmt.Errorf("record is absent or ambiguous; select its exact parent and identity")
	}

	return result, nil
}

func (s service) caseRecordViews(ctx context.Context, query RecordQuery) ([]RecordView, error) {
	cases, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).LoadBounded(ctx, workflowstate.RecordLimit)
	if err != nil {
		return nil, err
	}

	if len(cases) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("case record capacity exceeded")
	}

	var views []RecordView

	for _, metadata := range cases {
		view, err := caseRecordView(s.projectContext.ProjectDir, metadata)
		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	return filterRecordViews(views, query)
}

func caseRecordView(projectDir string, metadata authoringcore.CaseMetadata) (RecordView, error) {
	relative, err := filepath.Rel(projectDir, metadata.SourcePath)
	if err != nil {
		return RecordView{}, err
	}

	return RecordView{Kind: RecordUseCase, Identity: metadata.Slug, Title: metadata.Name, SourceHash: metadata.SourceHash, Reference: filepath.ToSlash(relative), Body: metadata.Description + "\n" + metadata.Context}, nil
}

func validRecordIdentity(id string) bool {
	return len(id) > 0 && len(id) <= 1024 && filepath.IsLocal(id) && filepath.Base(id) == id && !strings.ContainsAny(id, "\x00\r\n")
}

func nativeRecordKind(kind RecordKind) bool {
	return kind == RecordTicket || kind == RecordSliceSet || kind == RecordSlice
}
