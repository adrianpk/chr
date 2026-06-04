package chc

import (
	"path/filepath"
	"sort"
	"strings"
)

// WorkspaceCaseSuggestion describes a package boundary value.
type WorkspaceCaseSuggestion struct {
	Slug        string
	Name        string
	Description string
	Path        string
	NodeID      string
	Score       int
}

// WorkspaceRelatedDocumentsResult describes a package boundary value.
type WorkspaceRelatedDocumentsResult struct {
	Node     WorkspaceNode
	Related  []WorkspaceNode
	Resolved bool
}

// WorkspaceIndexSearchResult describes a package boundary value.
type WorkspaceIndexSearchResult struct {
	Node                  WorkspaceNode
	NodeID                string
	Score                 int
	MatchedSemanticDigest bool
	SemanticDigestStatus  WorkspaceSemanticDigestStatus
}

// ListCases lists the requested data.
func (i WorkspaceIndex) ListCases() []WorkspaceCaseSuggestion {
	cases := make([]WorkspaceCaseSuggestion, 0, len(i.CasesBySlug))

	slugs := make([]string, 0, len(i.CasesBySlug))
	for slug := range i.CasesBySlug {
		slugs = append(slugs, slug)
	}

	sort.Strings(slugs)

	for _, slug := range slugs {
		nodeID := i.CasesBySlug[slug]
		node := i.NodesByID[nodeID]
		cases = append(cases, workspaceCaseSuggestion(nodeID, node, 0))
	}

	return cases
}

// SuggestCases executes the corresponding package operation.
func (i WorkspaceIndex) SuggestCases(query string) []WorkspaceCaseSuggestion {
	query = normalizeWorkspaceQuery(query)
	if query == "" {
		return i.ListCases()
	}

	suggestions := []WorkspaceCaseSuggestion{}

	for _, suggestion := range i.ListCases() {
		score := workspaceCaseSuggestionScore(suggestion, query)
		if score == 0 {
			continue
		}

		suggestion.Score = score
		suggestions = append(suggestions, suggestion)
	}

	sort.SliceStable(suggestions, func(left, right int) bool {
		if suggestions[left].Score == suggestions[right].Score {
			return suggestions[left].Slug < suggestions[right].Slug
		}

		return suggestions[left].Score > suggestions[right].Score
	})

	return suggestions
}

// ResolvePath executes the corresponding package operation.
func (i WorkspaceIndex) ResolvePath(path string) (WorkspaceNode, bool) {
	nodeID, ok := i.NodesByPath[filepath.ToSlash(path)]
	if !ok {
		return WorkspaceNode{}, false
	}

	node, ok := i.NodesByID[nodeID]

	return node, ok
}

// RelatedDocuments executes the corresponding package operation.
func (i WorkspaceIndex) RelatedDocuments(nodeID string) WorkspaceRelatedDocumentsResult {
	node, ok := i.NodesByID[nodeID]
	if !ok {
		return WorkspaceRelatedDocumentsResult{}
	}

	result := WorkspaceRelatedDocumentsResult{
		Node:     node,
		Resolved: true,
	}

	caseSlug := node.CaseSlug
	if node.Kind == WorkspaceNodeCase {
		caseSlug = node.Slug
	}

	if caseSlug == "" {
		return result
	}

	relatedIDs := append([]string(nil), i.ByCaseSlug[caseSlug]...)
	sort.Strings(relatedIDs)

	for _, relatedID := range relatedIDs {
		if relatedID == nodeID {
			continue
		}

		related, ok := i.NodesByID[relatedID]
		if ok {
			result.Related = append(result.Related, related)
		}
	}

	return result
}

// Search executes the corresponding package operation.
func (i WorkspaceIndex) Search(query string, cache WorkspaceSemanticDigestCache) []WorkspaceIndexSearchResult {
	return i.SearchWithSemanticDigestStates(query, cache, nil)
}

// SearchWithSemanticDigestStates executes the corresponding package operation.
func (i WorkspaceIndex) SearchWithSemanticDigestStates(
	query string,
	cache WorkspaceSemanticDigestCache,
	states []WorkspaceSemanticDigestState,
) []WorkspaceIndexSearchResult {
	query = normalizeWorkspaceQuery(query)
	if query == "" {
		return nil
	}

	statusByNodeID := workspaceSemanticDigestStatusByNodeID(states)
	results := []WorkspaceIndexSearchResult{}

	for _, nodeID := range sortedWorkspaceIndexNodeIDs(i) {
		node := i.NodesByID[nodeID]
		score := workspaceNodeSearchScore(node, query)
		digest, digestOK := cache.Entries[nodeID]

		digestMatched := digestOK && strings.Contains(normalizeWorkspaceQuery(digest.Digest), query)
		if digestMatched {
			score += 20
		}

		if score == 0 {
			continue
		}

		result := WorkspaceIndexSearchResult{
			Node:   node,
			NodeID: nodeID,
			Score:  score,
		}
		if digestOK {
			result.SemanticDigestStatus = workspaceSemanticDigestSearchStatus(statusByNodeID, nodeID)
		}

		if digestMatched {
			result.MatchedSemanticDigest = true
		}

		results = append(results, result)
	}

	sort.SliceStable(results, func(left, right int) bool {
		if results[left].Score == results[right].Score {
			return results[left].NodeID < results[right].NodeID
		}

		return results[left].Score > results[right].Score
	})

	return results
}

func workspaceSemanticDigestStatusByNodeID(
	states []WorkspaceSemanticDigestState,
) map[string]WorkspaceSemanticDigestStatus {
	statusByNodeID := map[string]WorkspaceSemanticDigestStatus{}
	for _, state := range states {
		statusByNodeID[state.NodeID] = state.Status
	}

	return statusByNodeID
}

func workspaceSemanticDigestSearchStatus(
	statusByNodeID map[string]WorkspaceSemanticDigestStatus,
	nodeID string,
) WorkspaceSemanticDigestStatus {
	if status, ok := statusByNodeID[nodeID]; ok {
		return status
	}

	return WorkspaceSemanticDigestReady
}

func workspaceCaseSuggestion(nodeID string, node WorkspaceNode, score int) WorkspaceCaseSuggestion {
	return WorkspaceCaseSuggestion{
		Slug:        node.Slug,
		Name:        firstNonEmpty(node.Name, node.Title),
		Description: node.Description,
		Path:        node.RelativePath,
		NodeID:      nodeID,
		Score:       score,
	}
}

func workspaceCaseSuggestionScore(suggestion WorkspaceCaseSuggestion, query string) int {
	score := 0
	score += workspaceTextScore(suggestion.Slug, query, 100, 60)
	score += workspaceTextScore(suggestion.Name, query, 80, 40)
	score += workspaceTextScore(suggestion.Description, query, 30, 15)

	return score
}

func workspaceNodeSearchScore(node WorkspaceNode, query string) int {
	score := 0
	score += workspaceTextScore(node.Slug, query, 90, 45)
	score += workspaceTextScore(firstNonEmpty(node.Title, node.Name), query, 80, 40)
	score += workspaceTextScore(node.Description, query, 30, 15)
	score += workspaceTextScore(node.Status, query, 20, 10)
	score += workspaceTextScore(node.CaseSlug, query, 25, 10)
	score += workspaceTextScore(node.RelativePath, query, 25, 10)

	return score
}

func workspaceTextScore(value, query string, exact, contains int) int {
	value = normalizeWorkspaceQuery(value)
	if value == "" {
		return 0
	}

	if value == query {
		return exact
	}

	if strings.Contains(value, query) {
		return contains
	}

	return 0
}

func normalizeWorkspaceQuery(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
