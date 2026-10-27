package quiz

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ScopeID retains the canonical project or single-segment case artifact identity.
func ScopeID(scope Scope) (string, error) {
	caseSlug := strings.TrimSpace(scope.CaseSlug)
	if caseSlug == "" {
		return "project", nil
	}

	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(caseSlug)))
	if clean == "." || clean == ".." || strings.Contains(clean, "/") {
		return "", fmt.Errorf("quiz scope case slug must be a single path segment")
	}

	return clean, nil
}
