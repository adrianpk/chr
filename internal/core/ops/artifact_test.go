package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateOpsArtifactAllocatesGlobalIDAndEnvelope(t *testing.T) {
	projectDir := t.TempDir()

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(layout.Root(), "plan", "00001-existing"), 0o755); err != nil {
		t.Fatal(err)
	}

	artifact, err := CreateOpsArtifact(CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       "Parser validation",
		Producer:    "chr",
		ExternalIDs: []string{"exec-1"},
		Now:         func() time.Time { return time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("CreateOpsArtifact returned error: %v", err)
	}

	if artifact.ID != "00002-parser-validation" {
		t.Fatalf("ID = %q", artifact.ID)
	}

	assertOpsArtifactEnvelopeExists(t, artifact.Dir)

	var metadata struct {
		ExternalIDs []string `json:"external_ids"`
	}
	readOpsArtifactMetadata(t, artifact.Dir, &metadata)

	if len(metadata.ExternalIDs) != 1 || metadata.ExternalIDs[0] != "exec-1" {
		t.Fatalf("external IDs = %#v", metadata.ExternalIDs)
	}
}

func TestCreateOpsArtifactRejectsExistingAllocationLock(t *testing.T) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	lockPath := filepath.Join(layout.Root(), "lock", "id.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(lockPath, []byte("busy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: layout, Type: "run", Title: "Parser", Producer: "chr"}); err == nil {
		t.Fatal("CreateOpsArtifact returned nil error")
	}
}

func TestRefreshOpsArtifactMetadataIndexesSharedFiles(t *testing.T) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: layout, Type: "run", Title: "Parser", Producer: "chr"})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(artifact.Dir, "files", "context.md"), []byte("context\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(artifact.Dir, "evidence", "result.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(artifact.Dir, "extension", "chr", "state.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RefreshOpsArtifactMetadata(artifact.Dir, "succeeded", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	var metadata struct {
		Files  []string `json:"files"`
		Status string   `json:"status"`
	}
	readOpsArtifactMetadata(t, artifact.Dir, &metadata)

	if metadata.Status != "succeeded" {
		t.Fatalf("status = %q", metadata.Status)
	}

	if got, want := metadata.Files, []string{"artifact.md", "evidence/result.json", "files/context.md"}; !equalStringSlices(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func assertOpsArtifactEnvelopeExists(t *testing.T, dir string) {
	t.Helper()

	for _, path := range []string{"artifact.md", "metadata.json", "files", "evidence", filepath.Join("extension", "chr")} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func readOpsArtifactMetadata(t *testing.T, dir string, target any) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		t.Fatal(err)
	}
}

func equalStringSlices(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}

	return true
}
