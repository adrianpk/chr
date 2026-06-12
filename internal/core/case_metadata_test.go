package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	if !strings.Contains(err.Error(), "refresh before saving") {
		t.Fatalf("Save error = %q, want resync guidance", err.Error())
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
