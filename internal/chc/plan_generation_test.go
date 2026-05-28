package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveGeneratedPlanCopiesCanonicalPlanWithArchiveMetadata(t *testing.T) {
	projectDir := t.TempDir()
	scope := QuizScope{ProjectID: "project-1", CaseSlug: "blog"}
	createdAt := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	archivedAt := time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC)

	_, err := WriteGeneratedPlan(context.Background(), projectDir, GeneratedPlan{
		Scope:            scope,
		Title:            "Blog Plan",
		Body:             "Implementation slices.",
		SourceRequestID:  "request-1",
		SourceResponseID: "response-1",
		CreatedAt:        createdAt,
	})
	if err != nil {
		t.Fatalf("WriteGeneratedPlan returned error: %v", err)
	}

	paths, archived, err := ArchiveGeneratedPlan(context.Background(), projectDir, scope, archivedAt)
	if err != nil {
		t.Fatalf("ArchiveGeneratedPlan returned error: %v", err)
	}

	if !archived {
		t.Fatal("archived = false, want true")
	}

	if paths.RelativePath != ".chc/user/plans/archive/blog-plan-20260525-111213000000014.md" {
		t.Fatalf("RelativePath = %q, want deterministic archive path", paths.RelativePath)
	}

	content, err := os.ReadFile(paths.Path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	for _, want := range []string{
		`title: "Blog Plan"`,
		`archived_from: ".chc/user/plans/blog-plan.md"`,
		`archived_at: "2026-05-25T11:12:13.000000014Z"`,
		"# Blog Plan",
		"Implementation slices.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("archived content missing %q in:\n%s", want, string(content))
		}
	}
}

func TestArchiveGeneratedPlanReturnsFalseWhenCanonicalPlanIsMissing(t *testing.T) {
	projectDir := t.TempDir()
	scope := QuizScope{ProjectID: "project-1", CaseSlug: "blog"}
	archivedAt := time.Date(2026, 5, 25, 11, 12, 13, 0, time.UTC)

	_, archived, err := ArchiveGeneratedPlan(context.Background(), projectDir, scope, archivedAt)
	if err != nil {
		t.Fatalf("ArchiveGeneratedPlan returned error: %v", err)
	}

	if archived {
		t.Fatal("archived = true, want false")
	}

	archiveRoot := filepath.Join(projectDir, DirectoryName, PlanDirectory, "archive")
	if _, err := os.Stat(archiveRoot); !os.IsNotExist(err) {
		t.Fatalf("archive root exists after missing archive: %v", err)
	}
}
