package core

import projectcore "forge.adrianpk.com/chr/internal/core/project"

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/database"
)

func TestRenderDiagnosticBundleOmitsSecretsAndIncludesDiagnostics(t *testing.T) {
	bundle := DiagnosticBundle{
		ProjectDir: "/tmp/project",
		ProjectID:  "project-1",
		MigrationStatus: database.MigrationStatus{
			DatabasePath:      "/tmp/core.db",
			DatabaseExists:    true,
			SchemaTableExists: true,
			Available:         []database.Migration{{Version: "1", Name: "create_projects"}},
			Applied:           []database.AppliedMigration{{Version: "1", Name: "create_projects"}},
		},
		RegistryDiagnostics: []projectcore.ProjectRegistryDiagnostic{{
			Project: projectcore.Project{
				Name:    "example",
				Path:    "/tmp/project",
				Enabled: true,
			},
			Active: true,
			Status: projectcore.ProjectRegistryDiagnosticOK,
		}},
		RuntimeCleanupPreview: RuntimeArtifactCleanupPreview{
			ArtifactDir: "/tmp/project/.chr/default/runtime/runs",
			Candidates:  []RuntimeArtifactCleanupCandidate{{Path: "/tmp/project/.chr/default/runtime/runs/old"}},
		},
	}

	rendered := RenderDiagnosticBundle(bundle)

	for _, want := range []string{
		"# Diagnostic Bundle",
		"project_id: project-1",
		"database_exists: true",
		"example path=/tmp/project enabled=true active=true status=ok",
		"candidate: /tmp/project/.chr/default/runtime/runs/old",
		"workspace config contents omitted",
		"environment values omitted",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered bundle does not contain %q:\n%s", want, rendered)
		}
	}

	if strings.Contains(rendered, "bot_token") {
		t.Fatalf("rendered bundle contains secret-like token field:\n%s", rendered)
	}
}

func TestWriteDiagnosticBundleWritesMaintenanceFile(t *testing.T) {
	projectDir := t.TempDir()
	bundle := DiagnosticBundle{ProjectDir: projectDir, ProjectID: "project-1"}

	path, err := WriteDiagnosticBundle(context.Background(), projectDir, bundle)
	if err != nil {
		t.Fatalf("WriteDiagnosticBundle returned error: %v", err)
	}

	wantPath := filepath.Join(projectDir, ".chr", "maintenance", "diagnostic-bundle.md")
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read diagnostic bundle: %v", err)
	}

	if !strings.Contains(string(content), "# Diagnostic Bundle") {
		t.Fatalf("content = %q, want diagnostic bundle heading", string(content))
	}
}
