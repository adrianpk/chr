package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const diagnosticBundleRelativePath = ".chc/maintenance/diagnostic-bundle.md"

// DiagnosticBundle describes a package boundary value.
type DiagnosticBundle struct {
	ProjectDir             string
	ProjectID              string
	MigrationStatus        MigrationStatus
	RegistryDiagnostics    []ProjectRegistryDiagnostic
	ExecutionArtifactDrift []ExecutionArtifactDriftDiagnostic
	RuntimeCleanupPreview  RuntimeArtifactCleanupPreview
	Reports                []ReportListEntry
}

// DiagnosticBundlePath executes the corresponding package operation.
func DiagnosticBundlePath(projectDir string) string {
	return filepath.Join(projectDir, filepath.FromSlash(diagnosticBundleRelativePath))
}

// RenderDiagnosticBundle renders the corresponding package data.
func RenderDiagnosticBundle(bundle DiagnosticBundle) string {
	var output strings.Builder

	output.WriteString("# Diagnostic Bundle\n\n")
	writeDiagnosticBundleSummary(&output, bundle)
	writeDiagnosticBundleDatabase(&output, bundle.MigrationStatus)
	writeDiagnosticBundleRegistry(&output, bundle.RegistryDiagnostics)
	writeDiagnosticBundleExecutionArtifacts(&output, bundle.ExecutionArtifactDrift)
	writeDiagnosticBundleCleanupPreview(&output, bundle.RuntimeCleanupPreview)
	writeDiagnosticBundleReports(&output, bundle.Reports)
	writeDiagnosticBundleSecretPolicy(&output)

	return output.String()
}

// WriteDiagnosticBundle executes the corresponding package operation.
func WriteDiagnosticBundle(
	ctx context.Context,
	projectDir string,
	bundle DiagnosticBundle,
) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := DiagnosticBundlePath(projectDir)

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return "", fmt.Errorf("create diagnostic bundle directory: %w", err)
	}

	err = os.WriteFile(path, []byte(RenderDiagnosticBundle(bundle)), 0o644)
	if err != nil {
		return "", fmt.Errorf("write diagnostic bundle: %w", err)
	}

	return path, nil
}

func writeDiagnosticBundleSummary(output *strings.Builder, bundle DiagnosticBundle) {
	output.WriteString("## Summary\n\n")
	writeDiagnosticBundleField(output, "project_dir", bundle.ProjectDir)
	writeDiagnosticBundleField(output, "project_id", bundle.ProjectID)
	writeDiagnosticBundleField(output, "database", bundle.MigrationStatus.DatabasePath)
	writeDiagnosticBundleField(output, "registry_projects", fmt.Sprintf("%d", len(bundle.RegistryDiagnostics)))
	writeDiagnosticBundleField(output, "execution_artifacts", fmt.Sprintf("%d", len(bundle.ExecutionArtifactDrift)))
	writeDiagnosticBundleField(output, "cleanup_candidates", fmt.Sprintf("%d", len(bundle.RuntimeCleanupPreview.Candidates)))
	writeDiagnosticBundleField(output, "reports", fmt.Sprintf("%d", len(bundle.Reports)))
	output.WriteString("\n")
}

func writeDiagnosticBundleDatabase(output *strings.Builder, status MigrationStatus) {
	output.WriteString("## Database\n\n")
	writeDiagnosticBundleField(output, "database_exists", fmt.Sprintf("%t", status.DatabaseExists))
	writeDiagnosticBundleField(output, "schema_migrations_exists", fmt.Sprintf("%t", status.SchemaTableExists))
	writeDiagnosticBundleField(output, "available", fmt.Sprintf("%d", len(status.Available)))
	writeDiagnosticBundleField(output, "applied", fmt.Sprintf("%d", len(status.Applied)))
	writeDiagnosticBundleField(output, "pending", fmt.Sprintf("%d", len(status.Pending)))
	writeDiagnosticBundleField(output, "unknown", fmt.Sprintf("%d", len(status.Unknown)))
	output.WriteString("\n")
}

func writeDiagnosticBundleRegistry(output *strings.Builder, diagnostics []ProjectRegistryDiagnostic) {
	output.WriteString("## Project Registry\n\n")

	for _, diagnostic := range diagnostics {
		output.WriteString("- ")
		output.WriteString(diagnostic.Project.Name)
		output.WriteString(" path=")
		output.WriteString(diagnostic.Project.Path)
		output.WriteString(" enabled=")
		fmt.Fprintf(output, "%t", diagnostic.Project.Enabled)
		output.WriteString(" active=")
		fmt.Fprintf(output, "%t", diagnostic.Active)
		output.WriteString(" status=")
		output.WriteString(string(diagnostic.Status))
		output.WriteString("\n")
	}

	if len(diagnostics) == 0 {
		output.WriteString("- none\n")
	}

	output.WriteString("\n")
}

func writeDiagnosticBundleExecutionArtifacts(
	output *strings.Builder,
	diagnostics []ExecutionArtifactDriftDiagnostic,
) {
	output.WriteString("## Execution Artifacts\n\n")

	for _, diagnostic := range diagnostics {
		output.WriteString("- execution=")
		output.WriteString(diagnostic.Execution.ID)
		output.WriteString(" ticket=")
		output.WriteString(diagnostic.Execution.TicketKey)
		output.WriteString(" status=")
		output.WriteString(string(diagnostic.Execution.Status))
		output.WriteString(" artifact_status=")
		output.WriteString(string(diagnostic.Status))
		output.WriteString("\n")
	}

	if len(diagnostics) == 0 {
		output.WriteString("- none\n")
	}

	output.WriteString("\n")
}

func writeDiagnosticBundleCleanupPreview(output *strings.Builder, preview RuntimeArtifactCleanupPreview) {
	output.WriteString("## Runtime Cleanup Preview\n\n")
	writeDiagnosticBundleField(output, "artifact_dir", preview.ArtifactDir)
	writeDiagnosticBundleField(output, "referenced_run_dirs", fmt.Sprintf("%d", len(preview.ReferencedRunDirs)))
	writeDiagnosticBundleField(output, "cleanup_candidates", fmt.Sprintf("%d", len(preview.Candidates)))

	for _, candidate := range preview.Candidates {
		output.WriteString("- candidate: ")
		output.WriteString(candidate.Path)
		output.WriteString("\n")
	}

	if len(preview.Candidates) == 0 {
		output.WriteString("- candidate: none\n")
	}

	output.WriteString("\n")
}

func writeDiagnosticBundleReports(output *strings.Builder, reports []ReportListEntry) {
	output.WriteString("## Reports\n\n")

	for _, report := range reports {
		output.WriteString("- kind=")
		output.WriteString(string(report.Kind))
		output.WriteString(" path=")
		output.WriteString(report.RelativePath)
		output.WriteString(" ticket=")
		output.WriteString(emptyDiagnosticBundleField(report.TicketKey))
		output.WriteString(" execution=")
		output.WriteString(emptyDiagnosticBundleField(report.ExecutionID))
		output.WriteString(" loop=")
		output.WriteString(emptyDiagnosticBundleField(report.LoopID))
		output.WriteString(" status=")
		output.WriteString(emptyDiagnosticBundleField(report.Status))
		output.WriteString("\n")
	}

	if len(reports) == 0 {
		output.WriteString("- none\n")
	}

	output.WriteString("\n")
}

func writeDiagnosticBundleSecretPolicy(output *strings.Builder) {
	output.WriteString("## Secret Policy\n\n")
	output.WriteString("- workspace config contents omitted\n")
	output.WriteString("- environment values omitted\n")
	output.WriteString("- credential files omitted\n")
}

func writeDiagnosticBundleField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyDiagnosticBundleField(value))
	output.WriteString("\n")
}

func emptyDiagnosticBundleField(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}

	return value
}
