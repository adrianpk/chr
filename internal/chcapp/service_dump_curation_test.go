package chcapp

import (
	"strings"
	"testing"
)

func TestFallbackDumpCurationProviderPreservesBodyAndHighlights(t *testing.T) {
	provider := fallbackDumpCurationProvider{Reason: "llm_unavailable"}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{
		Title:   "Rough Notes",
		Content: "Use API_TOKEN=abc123\n\n- Keep the dump useful.\n- Preserve intent.",
	})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if strings.Contains(result.PolishedBody, "abc123") {
		t.Fatalf("PolishedBody contains secret: %q", result.PolishedBody)
	}

	if !strings.Contains(result.PolishedBody, "API_TOKEN=[REDACTED]") {
		t.Fatalf("PolishedBody = %q, want redacted token", result.PolishedBody)
	}

	if len(result.Highlights) != 3 {
		t.Fatalf("Highlights = %#v, want three recoverable lines", result.Highlights)
	}

	if result.Warnings[0] != "llm_unavailable" {
		t.Fatalf("Warnings = %#v, want llm_unavailable", result.Warnings)
	}
}

func TestFallbackDumpCurationProviderHandlesEmptyContent(t *testing.T) {
	provider := fallbackDumpCurationProvider{}

	result, err := provider.CurateDump(t.Context(), DumpCurationRequest{Title: "Empty"})
	if err != nil {
		t.Fatalf("CurateDump returned error: %v", err)
	}

	if result.PolishedBody != "No dump content provided." {
		t.Fatalf("PolishedBody = %q", result.PolishedBody)
	}

	if len(result.Highlights) != 1 {
		t.Fatalf("Highlights = %#v, want one fallback highlight", result.Highlights)
	}
}
