package authoring

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Lost acknowledgements at every replacement recover body, components, pointer
// and receipt together. Reopening never invokes preparation or appends twice.
func TestDumpMutationRecovery(t *testing.T) {
	for _, row := range []struct {
		action string
		count  int
	}{{"first", 6}, {"capture", 7}, {"append", 5}, {"save", 5}} {
		for fault := range row.count {
			t.Run(fmt.Sprintf("%s %d", row.action, fault), func(t *testing.T) { checkDumpMutationFault(t, row.action, fault) })
		}
	}
}

func prepareDumpMutationFault(t *testing.T, action string) (string, DumpMutationRequest) {
	t.Helper()

	project := t.TempDir()

	request := dumpCaptureRequest("fault", "Final material.")
	if action == "first" {
		return project, request
	}

	initial, err := MutateDump(t.Context(), project, dumpCaptureRequest("initial", "Initial material."), nil)
	if err != nil {
		t.Fatal(err)
	}

	if action != "capture" {
		observed, err := ReadDumpDocument(t.Context(), project, initial.Destination)
		if err != nil {
			t.Fatal(err)
		}

		request.Action, request.Reference, request.ExpectedRevision = action, initial.Destination, observed.Revision
	}

	return project, request
}

func checkDumpMutationFault(t *testing.T, action string, fault int) {
	t.Helper()

	project, request := prepareDumpMutationFault(t, action)
	calls := 0
	prepare := func(ctx context.Context, destination, raw string) (DumpCapturePreparation, error) {
		calls++

		return defaultDumpCapture(request)(ctx, destination, raw)
	}
	repo := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == fault {
			return errors.New("lost acknowledgement")
		}

		return nil
	}}

	first, err := repo.mutateDump(t.Context(), request, prepare)
	if err == nil {
		t.Fatal("fault not exercised")
	}

	if _, err := ObserveProjectDump(t.Context(), project); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("partial capture was readable", err)
	}

	beforeReplay := calls

	replay, err := MutateDump(t.Context(), project, request, prepare)
	if err != nil || replay != first || calls != beforeReplay {
		t.Fatal("retry repeated preparation or lost result", replay, first, calls, err)
	}

	observed, err := ReadDumpDocument(t.Context(), project, replay.Destination)
	if err != nil || strings.Count(observed.Content, "Final material.") != 1 {
		t.Fatal("retry duplicated or lost content", observed, err)
	}
}

// A prepared image survives cancellation before publication. Retry uses that
// image, and a changed input cannot replace its retained request arguments.
func TestDumpMutationPreparedRetry(t *testing.T) {
	project := t.TempDir()
	request := dumpCaptureRequest("prepared", "Retained material.")
	ctx, cancel := context.WithCancel(t.Context())
	calls := 0

	prepare := func(ctx context.Context, destination, raw string) (DumpCapturePreparation, error) {
		calls++

		cancel()

		return defaultDumpCapture(request)(ctx, destination, raw)
	}
	if _, err := MutateDump(ctx, project, request, prepare); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled operation published content", err)
	}

	if _, err := os.Stat(ProjectMetadataPath(project)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled operation published pointer", err)
	}

	changed := request

	changed.Content = "Changed payload."
	if _, err := MutateDump(t.Context(), project, changed, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("reserved operation accepted changed payload", err)
	}

	if _, err := MutateDump(t.Context(), project, request, prepare); err != nil || calls != 1 {
		t.Fatal("retry regenerated prepared content", calls, err)
	}
}

// Preparation failures retain their bounded attempt budget across reopening.
func TestDumpMutationPreparationBudget(t *testing.T) {
	project := t.TempDir()
	request := dumpCaptureRequest("budget", "Original material.")
	calls := 0

	prepare := func(context.Context, string, string) (DumpCapturePreparation, error) {
		calls++

		return DumpCapturePreparation{}, errors.New("fixture preparation failure")
	}
	for range 6 {
		if _, err := MutateDump(t.Context(), project, request, prepare); err == nil {
			t.Fatal("failed preparation accepted")
		}
	}

	if calls != 5 {
		t.Fatal("preparation budget was reset", calls)
	}

	if _, err := MutateDump(t.Context(), project, request, prepare); !errors.Is(err, ErrAuthoringProviderExhausted) {
		t.Fatal("exhaustion not retained", err)
	}
}

// A new capture must not replace a divergent pointer during recovery.
func TestDumpMutationPointerConflict(t *testing.T) {
	project := t.TempDir()
	request := dumpCaptureRequest("pointer", "Original material.")

	repo := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == 4 {
			return errors.New("lost acknowledgement")
		}

		return nil
	}}
	if _, err := repo.mutateDump(t.Context(), request, nil); err == nil {
		t.Fatal("fault not exercised")
	}

	path := ProjectMetadataPath(project)
	writeTestFile(t, path, "Independent project metadata.\n")

	if _, err := MutateDump(t.Context(), project, request, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("recovery replaced divergent project metadata", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "Independent project metadata.\n" {
		t.Fatal("independent bytes lost", err)
	}

	if _, err := os.Stat(filepath.Join(project, authoringRuntimeRef("pending.json"))); err != nil {
		t.Fatal("pending journal discarded", err)
	}
}
