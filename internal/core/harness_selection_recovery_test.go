package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Recovery covers each durable boundary: retained intent, renamed config, and
// recorded receipt. It completes the original request without rerunning setup.
func TestHarnessSelectionRecovery(t *testing.T) {
	for _, stage := range []string{"intent", "config", "receipt"} {
		t.Run(stage, func(t *testing.T) { checkHarnessSelectionRecovery(t, stage) })
	}
}

func checkHarnessSelectionRecovery(t *testing.T, stage string) {
	t.Helper()

	project := t.TempDir()
	request := harnessSelectionTestRequest("interrupted", contenthash.Sum(nil), "alpha")

	journal, image := retainedHarnessSelectionTestChange(t, request)
	if err := writeHarnessSelectionJournal(project, journal); err != nil {
		t.Fatal(err)
	}

	if stage != "intent" {
		writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), string(image))
	}

	if stage == "receipt" {
		if err := writeHarnessSelectionFile(project, harnessSelectionReceiptPath(request.Binding.RequestID), journal); err != nil {
			t.Fatal(err)
		}
	}

	options := harnessSelectionTestOptions(request)
	options.Validate = func(_ context.Context, _ WorkspaceConfig) error { t.Fatal("recovery repeated preparation"); return nil }

	result, err := ChangeHarnessSelections(t.Context(), project, options)
	if err != nil || result != journal.Result || string(readHarnessSelectionTestConfig(t, project)) != string(image) {
		t.Fatal("recovery changed the accepted request", result, err)
	}

	if _, err := os.Stat(workspacecore.UserPath(project, harnessSelectionPending)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("intent remained after recovery", err)
	}
}

// A reused identity cannot replace an interrupted change's choices or trigger
// its publication; the original authorized payload remains recoverable.
func TestHarnessSelectionRecoveryIdentity(t *testing.T) {
	project := t.TempDir()
	request := harnessSelectionTestRequest("interrupted", contenthash.Sum(nil), "alpha")

	journal, _ := retainedHarnessSelectionTestChange(t, request)
	if err := writeHarnessSelectionJournal(project, journal); err != nil {
		t.Fatal(err)
	}

	changed := request

	changed.Generation = harnessSelectionTestChoice("beta")
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(changed)); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("interrupted identity changed payload", err)
	}

	if len(readHarnessSelectionTestConfig(t, project)) != 0 {
		t.Fatal("conflicting replay published configuration")
	}

	accepted, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || accepted != journal.Result {
		t.Fatal("original payload could not recover", accepted, err)
	}
}

// A third configuration image cannot be overwritten during recovery, and a
// corrupt intent cannot fabricate a successful receipt.
func TestHarnessSelectionRecoveryConflict(t *testing.T) {
	for _, change := range []string{"foreign edit", "corrupt checksum"} {
		t.Run(change, func(t *testing.T) {
			project := t.TempDir()
			request := harnessSelectionTestRequest("interrupted", contenthash.Sum(nil), "alpha")

			journal, _ := retainedHarnessSelectionTestChange(t, request)
			if err := writeHarnessSelectionJournal(project, journal); err != nil {
				t.Fatal(err)
			}

			original := "unrelated: manual-edit\n"

			if change == "corrupt checksum" {
				writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, harnessSelectionPending), "{}")
			}

			writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), original)

			if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request)); err == nil {
				t.Fatal("conflicting recovery accepted")
			}

			if string(readHarnessSelectionTestConfig(t, project)) != original {
				t.Fatal("recovery overwrote foreign configuration")
			}
		})
	}
}

// Accepted identities are never silently evicted. Exhaustion preserves the old
// configuration while already accepted request IDs remain replayable.
func TestHarnessSelectionReceiptCapacity(t *testing.T) {
	project := t.TempDir()
	request := harnessSelectionTestRequest("first", contenthash.Sum(nil), "alpha")

	result, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil {
		t.Fatal(err)
	}

	dir := workspacecore.UserPath(project, "harness-selection-receipts")
	for i := 1; i < harnessSelectionReceiptCapacity; i++ {
		writeHarnessSelectionTestFile(t, filepath.Join(dir, "reserved-"+strconv.Itoa(i)), "reserved")
	}

	next := harnessSelectionTestRequest("next", result.Revision, "beta")
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(next)); err == nil {
		t.Fatal("receipt capacity was exceeded")
	}

	if contenthash.Sum(readHarnessSelectionTestConfig(t, project)) != result.Revision {
		t.Fatal("capacity rejection changed configuration")
	}

	replay, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || replay != result {
		t.Fatal("full capacity hid the accepted response", replay, err)
	}
}

// A retained selection intent prevents an upgrade from changing the same image;
// replay finishes that intent before the original explicit upgrade can proceed.
func TestHarnessSelectionUpgradeFence(t *testing.T) {
	store, options := workflowUpgradeFixture(t, "")
	request := harnessSelectionTestRequest("before-upgrade", options.ExpectedRevision, "alpha")
	request.ProjectRef = store.projectID

	journal, _ := retainedHarnessSelectionTestChange(t, request)
	if err := writeHarnessSelectionJournal(store.layout.ProjectDir(), journal); err != nil {
		t.Fatal(err)
	}

	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), options); err == nil {
		t.Fatal("upgrade overwrote a pending selection intent")
	}

	accepted, err := ChangeHarnessSelections(t.Context(), store.layout.ProjectDir(), harnessSelectionTestOptions(request))
	if err != nil {
		t.Fatal(err)
	}

	options.ExpectedRevision = accepted.Revision

	upgraded, err := store.UpgradeWorkflowConfiguration(t.Context(), options)
	if err != nil || upgraded.Current.Version != 1 {
		t.Fatal("upgrade did not resume after selection recovery", err)
	}

	config, err := LoadWorkspaceConfig(store.layout.ProjectDir())
	if err != nil || config.Harnesses.Worker != "alpha" || config.HarnessAudit.RequestID != request.Binding.RequestID {
		t.Fatal("upgrade lost selections or their audit", err)
	}
}

func retainedHarnessSelectionTestChange(t *testing.T, request HarnessSelectionChangeRequest) (harnessSelectionJournal, []byte) {
	t.Helper()

	digest, err := harnessSelectionRequestDigest(request)
	if err != nil {
		t.Fatal(err)
	}

	image := harnessSelectionChangeImage(nil, request, digest, "2026-10-19T12:00:00Z")

	config, err := parseWorkflowConfig(image)
	if err != nil {
		t.Fatal(err)
	}

	result, err := config.HarnessSelections(request.ProjectRef)
	if err != nil {
		t.Fatal(err)
	}

	journal := harnessSelectionJournal{Schema: "chr-harness-selection-change/v1", Request: request, Result: result, Before: request.ExpectedRevision}

	journal.Checksum, err = harnessSelectionJournalHash(journal)
	if err != nil {
		t.Fatal(err)
	}

	return journal, image
}
