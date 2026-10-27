package execution

import (
	"path/filepath"
	"strings"
)

// ReportPath executes the corresponding package operation.
func ReportPath(record Record) string {
	if strings.TrimSpace(record.Artifacts.RunDir) == "" {
		return ""
	}

	if strings.TrimSpace(record.Artifacts.ArtifactID) != "" {
		return filepath.Join(record.Artifacts.RunDir, "evidence", "execution-report.md")
	}

	return filepath.Join(record.Artifacts.RunDir, "execution-report.md")
}
