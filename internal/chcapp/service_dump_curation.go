package chcapp

import (
	"context"
	"regexp"
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

type fallbackDumpCurationProvider struct {
	Reason string
}

func (p fallbackDumpCurationProvider) CurateDump(
	_ context.Context,
	request DumpCurationRequest,
) (DumpCurationResult, error) {
	body := redactDumpCurationSecrets(strings.TrimSpace(request.Content))
	if body == "" {
		body = "No dump content provided."
	}

	highlights := deterministicDumpHighlights(body)
	warnings := []string{}

	if strings.TrimSpace(p.Reason) != "" {
		warnings = append(warnings, strings.TrimSpace(p.Reason))
	}

	return normalizeDumpCurationResult(DumpCurationResult{
		PolishedBody: body,
		Highlights:   highlights,
		Warnings:     warnings,
	}, body), nil
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

func deterministicDumpHighlights(body string) []string {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	highlights := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))

		if line == "" {
			continue
		}

		highlights = append(highlights, line)
		if len(highlights) == 5 {
			break
		}
	}

	if len(highlights) == 0 {
		return []string{"No recoverable source points were available."}
	}

	return highlights
}

var dumpCurationSecretAssignmentPattern = regexp.MustCompile(`(?i)\b(api[_-]?(?:key|token)|token|password|secret|credential)=([^\s,"}]+)`)

func redactDumpCurationSecrets(value string) string {
	return dumpCurationSecretAssignmentPattern.ReplaceAllString(value, "$1=[REDACTED]")
}
