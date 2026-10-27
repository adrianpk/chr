package authoring

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Pointer shape is validated without consulting a cache or accepting normalized
// traversal, absolute paths, another artifact kind or an invalid namespace.
func TestProjectDumpReference(t *testing.T) {
	for _, test := range []struct {
		name, reference string
		valid           bool
	}{
		{"absent", "", true},
		{"default", "ops/default/dump/00001-context/artifact.md", true},
		{"other namespace", "ops/team-a/dump/00001-context/artifact.md", true},
		{"traversal", "ops/default/dump/../context/artifact.md", false},
		{"absolute", "/ops/default/dump/context/artifact.md", false},
		{"wrong kind", "ops/default/spec/context/artifact.md", false},
		{"wrong body", "ops/default/dump/context/metadata.json", false},
		{"bad namespace", "ops/../dump/context/artifact.md", false},
		{"padded namespace", "ops/ team /dump/context/artifact.md", false},
		{"escaped separator", "ops/default/dump/context\\outside/artifact.md", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			metadata := ProjectMetadata{Name: "Example", Description: "Project context.", ProjectDumpRef: test.reference}

			_, err := ParseProjectMetadata(RenderProjectMetadata(metadata))
			if (err == nil) != test.valid {
				t.Fatal("invalid pointer acceptance", test.reference, err)
			}
		})
	}
}

// Invalid envelope identities, unsafe paths, malformed frontmatter and oversized
// sources fail explicitly instead of making room for an accidental new dump.
func TestProjectDumpSources(t *testing.T) {
	for _, defect := range []string{"metadata identity", "symlink body", "symlink namespace", "oversized body", "incomplete frontmatter", "duplicate scope", "timestamp"} {
		t.Run(defect, func(t *testing.T) {
			project := t.TempDir()
			reference := projectDumpTestArtifact(t, project, "default", "General", "General context.\n")
			damageProjectDumpSource(t, project, reference, defect)

			observed, err := ObserveProjectDump(t.Context(), project)
			if defect == "timestamp" {
				if err != nil || len(observed.Candidates) != 1 || observed.Candidates[0].TimestampIssue == "" {
					t.Fatal("missing timestamp evidence was hidden", observed, err)
				}

				return
			}

			if err == nil {
				t.Fatal("invalid source was accepted", defect, observed)
			}
		})
	}
}

func damageProjectDumpSource(t *testing.T, project, reference, defect string) {
	t.Helper()

	body := filepath.Join(project, filepath.FromSlash(reference))

	switch defect {
	case "metadata identity", "timestamp":
		path := filepath.Join(filepath.Dir(body), "metadata.json")

		changes := map[string][2]string{
			"metadata identity": {"\"namespace\": \"default\"", "\"namespace\": \"other\""},
			"timestamp":         {"2026-01-02T03:04:05Z", "invalid timestamp"},
		}
		values := changes[defect]
		replaceProjectDumpTestMetadata(t, path, values[0], values[1])
	case "symlink body":
		outside := filepath.Join(t.TempDir(), "outside.md")
		writeTestFile(t, outside, "Outside the project.\n")

		if err := os.Remove(body); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(outside, body); err != nil {
			t.Fatal(err)
		}
	case "symlink namespace":
		if err := os.Symlink(filepath.Join(project, "ops", "default"), filepath.Join(project, "ops", "linked")); err != nil {
			t.Fatal(err)
		}
	case "oversized body":
		writeTestFile(t, body, strings.Repeat("x", workflowstate.FileLimit+1))
	case "incomplete frontmatter":
		writeTestFile(t, body, "---\ntype: dump\nMissing closing delimiter.\n")
	case "duplicate scope":
		writeTestFile(t, body, "---\ncase: first\ncase: second\n---\nConflicting scope.\n")
	}
}

func replaceProjectDumpTestMetadata(t *testing.T, path, old, value string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, path, strings.Replace(string(data), old, value, 1))
}

// Two updates based on the same source must have one winner. Descriptive edits
// keep the canonical pointer, while rebinding it requires its explicit owner.
func TestProjectMetadataDumpOwnership(t *testing.T) {
	project := t.TempDir()
	reference := projectDumpTestArtifact(t, project, "default", "General", "General context.\n")
	metadataService := NewProjectMetadataService(project)

	_, err := metadataService.Save(ProjectMetadata{Name: "Example", Description: "Context.", ProjectDumpRef: reference})
	if !errors.Is(err, ErrProjectDumpReconciliation) {
		t.Fatal("unobserved pointer binding was allowed", err)
	}

	observed, err := ObserveProjectDump(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	saved, err := metadataService.AdoptProjectDump(t.Context(), observed.Revision)
	if err != nil {
		t.Fatal(err)
	}

	checkProjectMetadataConcurrent(t, metadataService, saved)

	current, err := metadataService.Load()
	if err != nil || current.ProjectDumpRef != reference {
		t.Fatal("descriptive edit erased canonical pointer", current, err)
	}

	current.ProjectDumpRef = ""
	if _, err := metadataService.SaveContext(t.Context(), current); err != nil {
		t.Fatal(err)
	}

	current, err = metadataService.Load()
	if err != nil || current.ProjectDumpRef != reference {
		t.Fatal("omitted pointer was cleared", current, err)
	}
}

func checkProjectMetadataConcurrent(t *testing.T, metadataService ProjectMetadataService, saved ProjectMetadata) {
	t.Helper()

	var group sync.WaitGroup

	start := make(chan struct{})
	results := make(chan error, 2)

	for _, name := range []string{"First edit", "Second edit"} {
		group.Go(func() {
			<-start

			update := saved
			update.Name = name

			_, err := metadataService.SaveContext(t.Context(), update)
			results <- err
		})
	}

	close(start)
	group.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Fatal("metadata source comparison was outside exclusion", successes)
	}
}

// Cancellation cannot start adoption or a project metadata write.
func TestProjectDumpCancellation(t *testing.T) {
	project := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := ObserveProjectDump(ctx, project); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled observation continued", err)
	}

	if _, err := NewProjectMetadataService(project).AdoptProjectDump(ctx, "observed"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled adoption continued", err)
	}

	if _, err := NewProjectMetadataService(project).SaveContext(ctx, ProjectMetadata{Name: "Example", Description: "Context."}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled metadata write continued", err)
	}

	if _, err := os.Stat(ProjectMetadataPath(project)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled operation wrote metadata", err)
	}
}

// Byte-identical documents in separate projects cannot reuse an adoption binding.
func TestProjectDumpBindingScope(t *testing.T) {
	projects := []string{t.TempDir(), t.TempDir()}
	for _, project := range projects {
		projectDumpTestArtifact(t, project, "default", "General", "The same copied context.\n")
		writeTestFile(t, ProjectMetadataPath(project), RenderProjectMetadata(ProjectMetadata{Name: "Example", Description: "Same copied metadata."}))
	}

	observed, err := ObserveProjectDump(t.Context(), projects[0])
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewProjectMetadataService(projects[1]).AdoptProjectDump(t.Context(), observed.Revision)
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("observation crossed project roots", err)
	}
}
