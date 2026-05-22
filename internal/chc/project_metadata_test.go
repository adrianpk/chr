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

func TestProjectMetadataServiceLoadsDefaultFromReadme(t *testing.T) {
	projectDir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(projectDir, "README.md"),
		[]byte("# Example\n\nExample project for testing metadata.\n\n## Later\n\nIgnored."),
		0o644,
	)
	if err != nil {
		t.Fatalf("write README: %v", err)
	}

	metadata, err := NewProjectMetadataService(projectDir).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if metadata.Name != filepath.Base(projectDir) {
		t.Fatalf("Name = %q, want project dir basename", metadata.Name)
	}

	if metadata.Description != "Example project for testing metadata." {
		t.Fatalf("Description = %q, want README paragraph", metadata.Description)
	}

	if metadata.Exists {
		t.Fatal("Exists = true, want false for default metadata")
	}
}

func TestProjectMetadataServiceSavesAndLoadsProjectMarkdown(t *testing.T) {
	projectDir := t.TempDir()
	service := NewProjectMetadataService(projectDir)

	saved, err := service.Save(ProjectMetadata{
		Name:        "Example",
		Description: "Example technical description.",
		Context:     "Long-form context.",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if !saved.Exists {
		t.Fatal("saved.Exists = false, want true")
	}

	content, err := os.ReadFile(ProjectMetadataPath(projectDir))
	if err != nil {
		t.Fatalf("read project metadata file: %v", err)
	}

	for _, want := range []string{
		"schema_version: 1",
		`name: "Example"`,
		`description: "Example technical description."`,
		"Long-form context.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("metadata file missing %q:\n%s", want, string(content))
		}
	}

	loaded, err := service.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Name != "Example" {
		t.Fatalf("Name = %q, want Example", loaded.Name)
	}

	if loaded.Context != "Long-form context." {
		t.Fatalf("Context = %q, want long-form context", loaded.Context)
	}
}

func TestProjectMetadataServiceRejectsStaleSave(t *testing.T) {
	projectDir := t.TempDir()
	service := NewProjectMetadataService(projectDir)

	saved, err := service.Save(ProjectMetadata{
		Name:        "Example",
		Description: "Example technical description.",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	err = os.WriteFile(ProjectMetadataPath(projectDir), []byte(`---
schema_version: 1
name: "External"
description: "External edit."
---
`), 0o644)
	if err != nil {
		t.Fatalf("write external project metadata edit: %v", err)
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

func TestParseProjectMetadataRejectsInvalidFrontmatter(t *testing.T) {
	_, err := ParseProjectMetadata("# Project\n")
	if err == nil {
		t.Fatal("ParseProjectMetadata returned nil error, want failure")
	}
}

func TestProjectMetadataRepositoryUpsertsAndGetsMetadata(t *testing.T) {
	repository := newTestProjectMetadataRepository(t)

	record := IndexedProjectMetadata{
		ProjectID: "project-1",
		Metadata: ProjectMetadata{
			SchemaVersion: ProjectMetadataSchemaVersion,
			Name:          "Example",
			Description:   "Example description.",
			Context:       "Context.",
			SourcePath:    ".chc/project.md",
			SourceHash:    "hash-1",
		},
		IndexedAt: "2026-05-22T10:00:00Z",
	}

	err := repository.UpsertProjectMetadata(context.Background(), record)
	if err != nil {
		t.Fatalf("UpsertProjectMetadata returned error: %v", err)
	}

	got, err := repository.GetProjectMetadata(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("GetProjectMetadata returned error: %v", err)
	}

	if got.Metadata.Name != "Example" {
		t.Fatalf("Name = %q, want Example", got.Metadata.Name)
	}

	record.Metadata.Description = "Updated description."
	record.Metadata.SourceHash = "hash-2"
	record.IndexedAt = "2026-05-22T11:00:00Z"

	err = repository.UpsertProjectMetadata(context.Background(), record)
	if err != nil {
		t.Fatalf("second UpsertProjectMetadata returned error: %v", err)
	}

	got, err = repository.GetProjectMetadata(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("second GetProjectMetadata returned error: %v", err)
	}

	if got.Metadata.Description != "Updated description." {
		t.Fatalf("Description = %q, want updated description", got.Metadata.Description)
	}
}

func TestProjectMetadataRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestProjectMetadataRepository(t)

	_, err := repository.GetProjectMetadata(context.Background(), "missing")
	if !errors.Is(err, ErrProjectMetadataNotFound) {
		t.Fatalf("GetProjectMetadata error = %v, want ErrProjectMetadataNotFound", err)
	}
}

func newTestProjectMetadataRepository(t *testing.T) ProjectMetadataRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repository, err := NewProjectMetadataRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectMetadataRepositoryFromDB returned error: %v", err)
	}

	return repository
}
