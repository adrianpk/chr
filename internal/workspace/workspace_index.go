package workspace

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// WorkspaceEdgeKind defines a package boundary type.
type WorkspaceEdgeKind string

const (
	// WorkspaceEdgeContains defines a package boundary constant.
	WorkspaceEdgeContains WorkspaceEdgeKind = "contains"
	// WorkspaceEdgeReferencesCase defines a package boundary constant.
	WorkspaceEdgeReferencesCase WorkspaceEdgeKind = "references_case"
)

// WorkspaceEdge describes a package boundary value.
type WorkspaceEdge struct {
	From string
	To   string
	Kind WorkspaceEdgeKind
}

// WorkspaceIndex describes a package boundary value.
type WorkspaceIndex struct {
	NodesByID    map[string]WorkspaceNode
	NodesByPath  map[string]string
	Edges        []WorkspaceEdge
	ByKind       map[WorkspaceNodeKind][]string
	ByCaseSlug   map[string][]string
	CasesBySlug  map[string]string
	TicketsByKey map[string]string
	Diagnostics  []WorkspaceGraphDiagnostic
	BuiltAt      time.Time
}

// WorkspaceIndexHealth describes a package boundary value.
type WorkspaceIndexHealth struct {
	NodeCount                  int
	EdgeCount                  int
	DiagnosticCount            int
	UnresolvedReferenceCount   int
	SemanticDigestPendingCount int
	SemanticDigestStaleCount   int
	BuiltAt                    time.Time
}

// WorkspaceIndexChanges describes a package boundary value.
type WorkspaceIndexChanges struct {
	Added   []string
	Changed []string
	Removed []string
}

// NewWorkspaceIndexFromGraph constructs the corresponding package service or value.
func NewWorkspaceIndexFromGraph(graph WorkspaceGraph) WorkspaceIndex {
	index := WorkspaceIndex{
		NodesByID:    map[string]WorkspaceNode{},
		NodesByPath:  map[string]string{},
		ByKind:       map[WorkspaceNodeKind][]string{},
		ByCaseSlug:   map[string][]string{},
		CasesBySlug:  map[string]string{},
		TicketsByKey: map[string]string{},
		Diagnostics:  append([]WorkspaceGraphDiagnostic(nil), graph.Diagnostics...),
		BuiltAt:      time.Now().UTC(),
	}

	var projectID string
	if graph.Project != nil {
		projectID = index.addNode(*graph.Project)
	}

	index.addCaseNodes(graph, projectID)
	index.addProjectWideNodes(graph, projectID)
	index.addCaseDocumentNodes(graph)
	index.addUnresolvedCaseReferenceNodes(graph)
	index.finalize()

	return index
}

func (i *WorkspaceIndex) addCaseNodes(graph WorkspaceGraph, projectID string) {
	for _, node := range graph.Cases {
		nodeID := i.addNode(node)

		i.CasesBySlug[node.Slug] = nodeID
		if projectID != "" {
			i.addEdge(projectID, nodeID, WorkspaceEdgeContains)
		}
	}
}

func (i *WorkspaceIndex) addProjectWideNodes(graph WorkspaceGraph, projectID string) {
	for _, node := range graph.ProjectWide {
		nodeID := i.addNode(node)
		if projectID != "" {
			i.addEdge(projectID, nodeID, WorkspaceEdgeContains)
		}
	}
}

func (i *WorkspaceIndex) addCaseDocumentNodes(graph WorkspaceGraph) {
	caseSlugs := sortedWorkspaceCaseSlugs(graph.CaseDocuments)
	for _, slug := range caseSlugs {
		caseID := i.CasesBySlug[slug]
		for _, node := range graph.CaseDocuments[slug] {
			nodeID := i.addNode(node)
			i.ByCaseSlug[slug] = appendWorkspaceStringIfMissing(i.ByCaseSlug[slug], nodeID)

			if caseID != "" {
				i.addEdge(caseID, nodeID, WorkspaceEdgeContains)
				i.addEdge(nodeID, caseID, WorkspaceEdgeReferencesCase)
			}
		}
	}
}

func (i *WorkspaceIndex) addUnresolvedCaseReferenceNodes(graph WorkspaceGraph) {
	for _, unresolved := range graph.UnresolvedCaseReferences {
		for _, node := range unresolved.References {
			nodeID := i.addNode(node)
			i.ByCaseSlug[unresolved.CaseSlug] = appendWorkspaceStringIfMissing(i.ByCaseSlug[unresolved.CaseSlug], nodeID)
		}
	}
}

// BuildIndex builds the requested workspace artifact.
func (b WorkspaceGraphBuilder) BuildIndex(ctx context.Context) (WorkspaceIndex, error) {
	graph, err := b.Build(ctx)
	if err != nil {
		return WorkspaceIndex{}, err
	}

	return NewWorkspaceIndexFromGraph(graph), nil
}

// Health executes the corresponding package operation.
func (i WorkspaceIndex) Health() WorkspaceIndexHealth {
	return WorkspaceIndexHealth{
		NodeCount:                len(i.NodesByID),
		EdgeCount:                len(i.Edges),
		DiagnosticCount:          len(i.Diagnostics),
		UnresolvedReferenceCount: len(i.unresolvedCaseSlugs()),
		BuiltAt:                  i.BuiltAt,
	}
}

// HealthWithSemanticDigestStates executes the corresponding package operation.
func (i WorkspaceIndex) HealthWithSemanticDigestStates(
	states []WorkspaceSemanticDigestState,
) WorkspaceIndexHealth {
	health := i.Health()

	for _, state := range states {
		switch state.Status {
		case WorkspaceSemanticDigestPending:
			health.SemanticDigestPendingCount++
		case WorkspaceSemanticDigestStale:
			health.SemanticDigestStaleCount++
		}
	}

	return health
}

// ChangesSince executes the corresponding package operation.
func (i WorkspaceIndex) ChangesSince(previous WorkspaceIndex) WorkspaceIndexChanges {
	changes := WorkspaceIndexChanges{}

	for id, node := range i.NodesByID {
		previousNode, ok := previous.NodesByID[id]
		if !ok {
			changes.Added = append(changes.Added, id)

			continue
		}

		if workspaceNodeSourceChanged(previousNode, node) {
			changes.Changed = append(changes.Changed, id)
		}
	}

	for id := range previous.NodesByID {
		if _, ok := i.NodesByID[id]; !ok {
			changes.Removed = append(changes.Removed, id)
		}
	}

	changes.sort()

	return changes
}

func (i *WorkspaceIndex) addNode(node WorkspaceNode) string {
	id := WorkspaceNodeID(node)
	if id == "" {
		return ""
	}

	i.NodesByID[id] = node
	i.ByKind[node.Kind] = appendWorkspaceStringIfMissing(i.ByKind[node.Kind], id)

	for _, path := range workspaceNodeIndexPaths(node) {
		i.NodesByPath[path] = id
	}

	if node.Kind == WorkspaceNodeTicket && node.Slug != "" {
		i.TicketsByKey[node.Slug] = id
	}

	return id
}

func (i *WorkspaceIndex) addEdge(from, to string, kind WorkspaceEdgeKind) {
	if from == "" || to == "" {
		return
	}

	i.Edges = append(i.Edges, WorkspaceEdge{
		From: from,
		To:   to,
		Kind: kind,
	})
}

func (i *WorkspaceIndex) finalize() {
	for kind := range i.ByKind {
		sort.Strings(i.ByKind[kind])
	}

	for slug := range i.ByCaseSlug {
		sort.Strings(i.ByCaseSlug[slug])
	}

	sort.SliceStable(i.Edges, func(left, right int) bool {
		if i.Edges[left].From == i.Edges[right].From {
			if i.Edges[left].To == i.Edges[right].To {
				return i.Edges[left].Kind < i.Edges[right].Kind
			}

			return i.Edges[left].To < i.Edges[right].To
		}

		return i.Edges[left].From < i.Edges[right].From
	})
}

// WorkspaceNodeID executes the corresponding package operation.
func WorkspaceNodeID(node WorkspaceNode) string {
	switch {
	case node.RelativePath != "":
		return string(node.Kind) + ":" + filepath.ToSlash(node.RelativePath)
	case node.Path != "":
		return string(node.Kind) + ":" + filepath.ToSlash(node.Path)
	case node.Slug != "":
		return string(node.Kind) + ":" + node.Slug
	default:
		return ""
	}
}

func workspaceNodeIndexPaths(node WorkspaceNode) []string {
	paths := []string{}
	if strings.TrimSpace(node.RelativePath) != "" {
		paths = append(paths, filepath.ToSlash(node.RelativePath))
	}

	if strings.TrimSpace(node.Path) != "" {
		paths = append(paths, filepath.ToSlash(node.Path))
	}

	return paths
}

func sortedWorkspaceCaseSlugs(documents map[string][]WorkspaceNode) []string {
	slugs := make([]string, 0, len(documents))
	for slug := range documents {
		slugs = append(slugs, slug)
	}

	sort.Strings(slugs)

	return slugs
}

func appendWorkspaceStringIfMissing(values []string, value string) []string {
	if value == "" {
		return values
	}

	for _, current := range values {
		if current == value {
			return values
		}
	}

	return append(values, value)
}

func (i WorkspaceIndex) unresolvedCaseSlugs() []string {
	slugs := []string{}

	for slug := range i.ByCaseSlug {
		if _, ok := i.CasesBySlug[slug]; !ok {
			slugs = append(slugs, slug)
		}
	}

	sort.Strings(slugs)

	return slugs
}

func workspaceNodeSourceChanged(previous, current WorkspaceNode) bool {
	if previous.SourceHash != current.SourceHash {
		return true
	}

	if !previous.ModifiedAt.Equal(current.ModifiedAt) {
		return true
	}

	return previous.Size != current.Size
}

func (c *WorkspaceIndexChanges) sort() {
	sort.Strings(c.Added)
	sort.Strings(c.Changed)
	sort.Strings(c.Removed)
}
