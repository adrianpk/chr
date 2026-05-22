package chc

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCaseMetadataServiceSavesAndLoadsCases(t *testing.T) {
	projectDir := t.TempDir()
	service := NewCaseMetadataService(projectDir)

	saved, err := service.Save(CaseMetadata{
		Name:        "Blog Publishing",
		Description: "Validate the blog authoring workflow.",
		Context:     "Long-form case context.",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if saved.Slug != "blog-publishing" {
		t.Fatalf("Slug = %q, want blog-publishing", saved.Slug)
	}

	if !saved.Exists {
		t.Fatal("saved.Exists = false, want true")
	}

	content, err := os.ReadFile(CaseMetadataPath(projectDir, "blog-publishing"))
	if err != nil {
		t.Fatalf("read case metadata file: %v", err)
	}

	assertCaseMetadataFileContains(t, string(content), "schema_version: 1")
	assertCaseMetadataFileContains(t, string(content), `name: "Blog Publishing"`)
	assertCaseMetadataFileContains(t, string(content), `description: "Validate the blog authoring workflow."`)
	assertCaseMetadataFileContains(t, string(content), "Long-form case context.")

	cases, err := service.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}

	if len(cases) != 1 {
		t.Fatalf("len(cases) = %d, want 1", len(cases))
	}

	if cases[0].Name != "Blog Publishing" {
		t.Fatalf("Name = %q, want Blog Publishing", cases[0].Name)
	}

	if cases[0].Context != "Long-form case context." {
		t.Fatalf("Context = %q, want long-form context", cases[0].Context)
	}
}

func assertCaseMetadataFileContains(t *testing.T, content, want string) {
	t.Helper()

	if !strings.Contains(content, want) {
		t.Fatalf("case metadata file missing %q:\n%s", want, content)
	}
}

func TestCaseMetadataServiceRejectsMissingName(t *testing.T) {
	_, err := NewCaseMetadataService(t.TempDir()).Save(CaseMetadata{
		Description: "Missing name.",
	})
	if err == nil {
		t.Fatal("Save returned nil error, want validation failure")
	}
}

func TestCaseMetadataServiceRejectsStaleSave(t *testing.T) {
	projectDir := t.TempDir()
	service := NewCaseMetadataService(projectDir)

	saved, err := service.Save(CaseMetadata{
		Name:        "Example",
		Description: "Example description.",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	err = os.WriteFile(saved.SourcePath, []byte(`---
schema_version: 1
name: "External"
description: "External edit."
---
`), 0o644)
	if err != nil {
		t.Fatalf("write external case metadata edit: %v", err)
	}

	saved.Description = "TUI edit."

	_, err = service.Save(saved)
	if err == nil {
		t.Fatal("Save returned nil error, want stale write failure")
	}

	if !strings.Contains(err.Error(), "resync before saving") {
		t.Fatalf("Save error = %q, want resync guidance", err.Error())
	}
}

func TestCaseMetadataRepositoryUpsertsAndListsCases(t *testing.T) {
	repository := newTestCaseMetadataRepository(t)

	record := IndexedCaseMetadata{
		ProjectID: "project-1",
		Case: CaseMetadata{
			SchemaVersion: CaseMetadataSchemaVersion,
			Slug:          "example",
			Name:          "Example",
			Description:   "Example description.",
			Context:       "Context.",
			SourcePath:    ".chc/cases/example.md",
			SourceHash:    "hash-1",
		},
		IndexedAt: "2026-05-22T10:00:00Z",
	}

	err := repository.UpsertCaseMetadata(context.Background(), record)
	if err != nil {
		t.Fatalf("UpsertCaseMetadata returned error: %v", err)
	}

	record.Case.Description = "Updated description."
	record.Case.SourceHash = "hash-2"
	record.IndexedAt = "2026-05-22T11:00:00Z"

	err = repository.UpsertCaseMetadata(context.Background(), record)
	if err != nil {
		t.Fatalf("second UpsertCaseMetadata returned error: %v", err)
	}

	records, err := repository.ListProjectCaseMetadata(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectCaseMetadata returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	if records[0].Case.Description != "Updated description." {
		t.Fatalf("Description = %q, want updated description", records[0].Case.Description)
	}
}

func TestCaseMetadataRepositoryRejectsInvalidCase(t *testing.T) {
	repository := newTestCaseMetadataRepository(t)

	err := repository.UpsertCaseMetadata(context.Background(), IndexedCaseMetadata{
		ProjectID: "project-1",
		Case: CaseMetadata{
			SchemaVersion: CaseMetadataSchemaVersion,
			Slug:          "missing-name",
			SourcePath:    ".chc/cases/missing-name.md",
			SourceHash:    "hash",
		},
		IndexedAt: "2026-05-22T10:00:00Z",
	})
	if err == nil {
		t.Fatal("UpsertCaseMetadata returned nil error, want validation failure")
	}
}

func TestCaseMetadataLoadAllMissingDirectoryReturnsEmpty(t *testing.T) {
	cases, err := NewCaseMetadataService(t.TempDir()).LoadAll()
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}

	if len(cases) != 0 {
		t.Fatalf("len(cases) = %d, want 0", len(cases))
	}
}

func TestParseCaseMetadataRejectsInvalidFrontmatter(t *testing.T) {
	_, err := ParseCaseMetadata("# Case\n")
	if err == nil {
		t.Fatal("ParseCaseMetadata returned nil error, want failure")
	}
}

func newTestCaseMetadataRepository(t *testing.T) CaseMetadataRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	repository, err := NewCaseMetadataRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewCaseMetadataRepositoryFromDB returned error: %v", err)
	}

	return repository
}

func TestCaseMetadataServiceCreatesUniqueSlug(t *testing.T) {
	projectDir := t.TempDir()
	service := NewCaseMetadataService(projectDir)

	first, err := service.Save(CaseMetadata{Name: "Same Name"})
	if err != nil {
		t.Fatalf("first Save returned error: %v", err)
	}

	second, err := service.Save(CaseMetadata{Name: "Same Name"})
	if err != nil {
		t.Fatalf("second Save returned error: %v", err)
	}

	if filepath.Base(first.SourcePath) != "same-name.md" {
		t.Fatalf("first path = %q, want same-name.md", first.SourcePath)
	}

	if filepath.Base(second.SourcePath) != "same-name-2.md" {
		t.Fatalf("second path = %q, want same-name-2.md", second.SourcePath)
	}
}
