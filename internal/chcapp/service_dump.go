package chcapp

import (
	"sort"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type DumpScope struct {
	CaseSlug string
}

type DumpView struct {
	NodeID       string
	Title        string
	Status       string
	Scope        DumpScope
	RelativePath string
	UpdatedAt    time.Time
}

type DumpListResult struct {
	Message string
	Dumps   []DumpView
}

type DumpCreateRequest struct {
	Title   string
	Content string
	Scope   DumpScope
}

type DumpAppendRequest struct {
	RelativePath string
	Content      string
}

type DumpSaveRequest struct {
	RelativePath string
	Content      string
}

type DumpOpenRequest struct {
	RelativePath string
}

type DumpOperationResult struct {
	Message string
	Dumps   []DumpView
}

func (s service) ListDumps() (DumpListResult, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return DumpListResult{}, err
	}

	return DumpListResult{
		Message: "Dumps loaded.",
		Dumps:   dumpViewsFromIndex(index),
	}, nil
}

func (s service) dumpDirectory() string {
	return chc.DumpDirectory(s.projectContext.ProjectDir)
}

func dumpViewsFromIndex(index chc.WorkspaceIndex) []DumpView {
	nodeIDs := append([]string(nil), index.ByKind[chc.WorkspaceNodeDump]...)
	sort.Strings(nodeIDs)

	dumps := make([]DumpView, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		node, ok := index.NodesByID[nodeID]
		if !ok {
			continue
		}

		dumps = append(dumps, dumpViewFromNode(nodeID, node))
	}

	return dumps
}

func dumpViewFromNode(nodeID string, node chc.WorkspaceNode) DumpView {
	return DumpView{
		NodeID:       nodeID,
		Title:        node.Title,
		Status:       node.Status,
		Scope:        DumpScope{CaseSlug: node.CaseSlug},
		RelativePath: node.RelativePath,
		UpdatedAt:    node.ModifiedAt,
	}
}
