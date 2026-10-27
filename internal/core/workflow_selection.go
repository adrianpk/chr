package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowSelector names an exact extent. Only loose work omits Identity.
type WorkflowSelector struct {
	Extent   string
	Identity string
}

// WorkflowPreviewRequest bounds a read-only selection of authoritative records.
type WorkflowPreviewRequest struct {
	Selector         WorkflowSelector
	Offset           int
	Limit            int
	SnapshotHash     string
	Target           WorkflowPreviewTarget
	CumulativeBranch string
	CumulativeCommit string
	Filters          TicketLoopFilters
	Statuses         []string
	Order            TicketLoopOrder
	MinPriority      string
	MinSeverity      string
}

// WorkflowPreviewMember freezes the identity, revision and contract of one task.
type WorkflowPreviewMember struct {
	Key          string
	Path         string
	Revision     uint64
	SourceHash   string
	ContractHash string
	SetID        string
	SliceID      string
	TaskOrder    int
	Status       string
	Phase        string
	Readiness    string
	Blocker      *WorkflowBlocker
}

// WorkflowPreview returns one bounded page and counts for the whole extent.
// Exclusion reasons are mutually exclusive; excluded work is never admitted.
type WorkflowPreview struct {
	Selector     WorkflowSelector
	Authority    string
	Members      []WorkflowPreviewMember
	Total        int
	Excluded     map[string]int
	SnapshotHash string
	NextOffset   int
	SetOrderRule string
	admitted     map[WorkflowRef]bool
}

type workflowSelection struct {
	snapshot       WorkflowSnapshot
	tickets        []IndexedTicket
	records        map[string]WorkflowRecord
	sets           map[string]WorkflowRecord
	slices         map[string]WorkflowSlice
	evidenceHashes map[string]string
}

// Preview resolves exact scope under the same exclusion used by domain writes.
// It does not claim work, create invocations, refresh SQL or edit a tracker.
func (s *WorkflowStore) Preview(ctx context.Context, request WorkflowPreviewRequest) (preview WorkflowPreview, err error) {
	request, err = normalizeWorkflowPreviewRequest(request)
	if err != nil {
		return WorkflowPreview{}, err
	}

	if err := validateWorkflowPreviewRequest(request); err != nil {
		return WorkflowPreview{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		selection, readErr := s.readWorkflowSelection(ctx)
		if readErr != nil {
			return readErr
		}

		preview, readErr = selection.preview(request)
		if readErr != nil {
			return readErr
		}

		return s.observePrerequisites(ctx, selection, request, &preview)
	})

	if err == nil && request.SnapshotHash != "" && request.SnapshotHash != preview.SnapshotHash {
		return WorkflowPreview{}, fmt.Errorf("%w: restart workflow preview pagination", workflowstate.ErrConflict)
	}

	return preview, err
}

func validateWorkflowPreviewRequest(request WorkflowPreviewRequest) error {
	if request.Limit < 1 || request.Limit > WorkflowPageLimit || request.Offset < 0 || request.Offset > WorkflowRecordLimit {
		return fmt.Errorf("workflow preview page bounds exceeded")
	}

	selector := request.Selector
	if strings.TrimSpace(selector.Identity) != selector.Identity || len(selector.Identity) > 128 {
		return fmt.Errorf("invalid workflow selector identity")
	}

	if err := validateWorkflowPreviewContext(request); err != nil {
		return err
	}

	return validateWorkflowSelector(selector)
}

func validateWorkflowSelector(selector WorkflowSelector) error {
	switch selector.Extent {
	case "ticket", "slice", "slice_set", "use_case":
		if selector.Identity == "" {
			return fmt.Errorf("%s identity is required", selector.Extent)
		}
	case "loose":
		if selector.Identity != "" {
			return fmt.Errorf("loose extent cannot name an identity")
		}
	default:
		return fmt.Errorf("unsupported workflow extent: %s", selector.Extent)
	}

	return nil
}

func (selection workflowSelection) preview(request WorkflowPreviewRequest) (WorkflowPreview, error) {
	if err := selection.resolveSelector(request.Selector); err != nil {
		return WorkflowPreview{}, err
	}

	preview := WorkflowPreview{Selector: request.Selector, Authority: "tickets", Excluded: map[string]int{}, NextOffset: -1}

	requestHash, err := workflowPreviewRequestHash(request)
	if err != nil {
		return WorkflowPreview{}, err
	}

	preview.SnapshotHash = contenthash.Sum([]byte(selection.hash() + requestHash))

	members := []WorkflowPreviewMember{}

	for _, ticket := range selection.tickets {
		record, prepared := selection.records[ticket.Key]
		if !selection.matches(request.Selector, ticket, record) {
			continue
		}

		reason := workflowPreviewExclusion(request, ticket, record, prepared)

		if reason != "" {
			preview.Excluded[reason]++

			continue
		}

		members = append(members, workflowPreviewMember(record))
	}

	var orderErr error

	preview.SetOrderRule, orderErr = selection.resolveOrder(members, request)
	if orderErr != nil {
		return WorkflowPreview{}, orderErr
	}

	preview.admitted = workflowAdmittedRefs(members, request.Selector)
	preview.Total = len(members)
	start := min(request.Offset, len(members))
	end := min(start+request.Limit, len(members))

	preview.Members = members[start:end]
	if end < len(members) {
		preview.NextOffset = end
	}

	return preview, nil
}

func (selection workflowSelection) resolveSelector(selector WorkflowSelector) error {
	found := selector.Extent == "loose"
	switch selector.Extent {
	case "ticket":
		for _, ticket := range selection.tickets {
			if ticket.Key == selector.Identity {
				found = true
			}
		}
	case "slice":
		_, found = selection.slices[selector.Identity]
	case "slice_set":
		_, found = selection.sets[selector.Identity]
	case "use_case":
		for _, set := range selection.sets {
			if set.Set.UseCase == selector.Identity {
				found = true
			}
		}
	}

	if !found {
		return fmt.Errorf("workflow %s identity not found: %s", selector.Extent, selector.Identity)
	}

	return nil
}

func (selection workflowSelection) matches(selector WorkflowSelector, ticket IndexedTicket, record WorkflowRecord) bool {
	switch selector.Extent {
	case "ticket":
		return ticket.Key == selector.Identity
	case "slice":
		return record.Task != nil && record.Task.SliceID == selector.Identity
	case "slice_set":
		return record.Task != nil && record.Task.SetID == selector.Identity
	case "use_case":
		if record.Task != nil && record.Task.SetID != "" {
			return selection.sets[record.Task.SetID].Set.UseCase == selector.Identity
		}

		return ticket.UseCase == selector.Identity
	default:
		return true
	}
}

func workflowSelectionExclusion(selector WorkflowSelector, record WorkflowRecord, prepared bool) string {
	if !prepared {
		return "unprepared"
	}

	if selector.Extent == "use_case" && record.Task.SetID == "" {
		return "unsliced"
	}

	if selector.Extent == "loose" && record.Task.SetID != "" {
		return "slice_owned"
	}

	if record.Task.Phase == "integrated" {
		return "integrated"
	}

	return ""
}

func workflowPreviewMember(record WorkflowRecord) WorkflowPreviewMember {
	task := record.Task

	return WorkflowPreviewMember{Key: task.Key, Path: record.Path, Revision: record.Revision, SourceHash: record.SourceHash, ContractHash: task.ContractHash, SetID: task.SetID, SliceID: task.SliceID, TaskOrder: task.Order, Status: task.Status, Phase: task.Phase}
}

func workflowAdmittedRefs(members []WorkflowPreviewMember, selector WorkflowSelector) map[WorkflowRef]bool {
	admitted := map[WorkflowRef]bool{}
	for _, member := range members {
		admitted[WorkflowRef{Kind: "ticket", ID: member.Key}] = true
		if selector.Extent != "ticket" && member.SliceID != "" {
			admitted[WorkflowRef{Kind: "slice", ID: member.SliceID}] = true
		}
	}

	return admitted
}

func workflowPreviewRequestHash(request WorkflowPreviewRequest) (string, error) {
	request.Offset, request.Limit, request.SnapshotHash = 0, 0, ""

	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	return contenthash.Sum(encoded), nil
}

func validateWorkflowPreviewContext(request WorkflowPreviewRequest) error {
	if len(request.Target.Repository) > 2048 || strings.ContainsAny(request.Target.Repository, "\x00\n\r") {
		return fmt.Errorf("invalid workflow target repository")
	}

	for _, commit := range []string{request.Target.Commit, request.CumulativeCommit} {
		if commit != "" && !workflowCommitID(commit) {
			return fmt.Errorf("workflow preview requires full immutable commit IDs")
		}
	}

	branch := request.CumulativeBranch
	if branch != "" && (len(branch) > 256 || !strings.HasPrefix(branch, "refs/heads/") || strings.ContainsAny(branch, "\x00\n\r ~^:?*[\\")) {
		return fmt.Errorf("invalid cumulative branch ref")
	}

	return nil
}

func workflowPreviewExclusion(request WorkflowPreviewRequest, ticket IndexedTicket, record WorkflowRecord, prepared bool) string {
	reason := workflowSelectionExclusion(request.Selector, record, prepared)
	if reason == "" && request.Selector.Extent == "loose" {
		return workflowLooseExclusion(request, ticket)
	}

	return reason
}

func (selection workflowSelection) resolveOrder(members []WorkflowPreviewMember, request WorkflowPreviewRequest) (string, error) {
	if request.Selector.Extent == "loose" {
		selection.orderLooseMembers(members, request.Order)

		return string(request.Order), nil
	}

	return selection.orderMembers(members, request.Selector)
}
