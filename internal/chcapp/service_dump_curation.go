package chcapp

import (
	"context"
	"strings"
)

type DumpCurationRequest struct {
	ProjectDir string
	ProjectID  string
	Title      string
	Content    string
	Scope      DumpScope
}

type DumpCurationResult struct {
	PolishedBody string
	Highlights   []string
	Warnings     []string
}

type DumpCurationProvider interface {
	CurateDump(ctx context.Context, request DumpCurationRequest) (DumpCurationResult, error)
}

func normalizeDumpCurationResult(result DumpCurationResult, fallbackBody string) DumpCurationResult {
	result.PolishedBody = strings.TrimSpace(result.PolishedBody)
	if result.PolishedBody == "" {
		result.PolishedBody = strings.TrimSpace(fallbackBody)
	}

	result.Highlights = normalizeDumpCurationStrings(result.Highlights)
	result.Warnings = normalizeDumpCurationStrings(result.Warnings)

	return result
}

func normalizeDumpCurationStrings(values []string) []string {
	normalized := []string{}
	seen := map[string]bool{}

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		normalized = append(normalized, value)
	}

	return normalized
}
