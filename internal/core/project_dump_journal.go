package core

import (
	"encoding/hex"
	"path/filepath"
	"strings"
)

const (
	projectDumpJournalSchema = "chr-authoring-journal/v2"
	projectDumpJournalBytes  = 24 * WorkflowFileLimit
	projectDumpSourceLimit   = 32
)

func authoringJournalLimit(schema string) int {
	if schema == projectDumpJournalSchema {
		return projectDumpJournalBytes
	}

	return WorkflowFileLimit
}

func validAuthoringJournalBounds(journal authoringMutationJournal, size int) bool {
	maximum := 12
	if journal.Schema == projectDumpJournalSchema {
		maximum = 128
	} else if journal.Schema != "chr-authoring-journal/v1" {
		return false
	}

	return len(journal.Changes) >= 2 && len(journal.Changes) <= maximum && size <= authoringJournalLimit(journal.Schema)
}

func validAuthoringJournalVersionPath(schema, reference string) bool {
	if validAuthoringJournalPath(reference) {
		return true
	}

	if schema != projectDumpJournalSchema {
		return false
	}

	if reference == filepath.Join(DirectoryName, UserDirectoryName, "project.md") {
		return true
	}

	return validProjectDumpHistoryPath(reference) || validQuizRuntimeJournalPath(reference) || validPlanRevisionJournalPath(reference) || validPlanTrackerJournalPath(reference)
}

func validQuizRuntimeJournalPath(reference string) bool {
	parts := strings.Split(filepath.ToSlash(reference), "/")
	if len(parts) != 5 || parts[0] != DirectoryName || !IsValidOpsNamespace(parts[1]) || parts[2] != "tmp" || parts[3] != "quiz-runtime" {
		return false
	}

	return strings.HasSuffix(parts[4], ".json") && validateAuthoringSessionID(strings.TrimSuffix(parts[4], ".json")) == nil
}

func validProjectDumpHistoryPath(reference string) bool {
	parts := strings.Split(filepath.ToSlash(reference), "/")
	if len(parts) != 8 || parts[4] != "files" || parts[5] != "history" || len(parts[6]) != 64 {
		return false
	}

	if parts[7] != "artifact.md" && parts[7] != "metadata.json" {
		return false
	}

	_, err := hex.DecodeString(parts[6])

	return err == nil && ValidateProjectDumpReference(strings.Join(parts[:4], "/")+"/artifact.md") == nil
}
