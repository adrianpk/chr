package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestNativeFanoutCrashProcess exits the owned test host during the actual
// member/set/queue transaction, leaving its existing write-ahead journal intact.
func TestNativeFanoutCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_NATIVE_FANOUT_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	index, err := strconv.Atoi(os.Getenv("CHR_NATIVE_FANOUT_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowFinalizationJournal(t, store) {
			os.Exit(23)
		}

		return nil
	}
	_, err = store.FinalizeWorkflowSlice(t.Context(), os.Getenv("CHR_NATIVE_FANOUT_ID"), nil, nil)
	t.Fatal("owned fanout writer did not exit", err)
}

// TestNativeFanoutCrash recovers real host death at each member, the owning set,
// queue integration and invocation cursor. Replay preserves the one receipt and
// completed members while leaving the later slice and set closure unchanged.
func TestNativeFanoutCrash(t *testing.T) {
	// Artifact bodies are unchanged and require no replacement. These indexes
	// cover each task metadata image, set metadata, queue and invocation.
	for _, index := range []int{0, 2, 4, 6, 8, 9} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))
			before := nativeFanoutSnapshot(t, fixture.f.store)

			data, err := os.ReadFile(filepath.Join(fixture.f.store.layout.Root(), receipt.Path))
			if err != nil {
				t.Fatal(err)
			}

			crashNativeFanout(t, fixture, index)

			fixture.f.store = NewWorkflowStore(fixture.f.store.layout, fixture.f.store.projectID)
			if _, err := fixture.f.store.FinalizeWorkflowSlice(t.Context(), fixture.f.entry.InvocationID, nil, nil); err != nil {
				t.Fatal(err)
			}

			checkNativeFinalization(t, fixture, receipt, data)
			checkNativeFanoutLater(t, fixture, before)
		})
	}
}

func crashNativeFanout(t *testing.T, fixture nativeCanonicalFixture, index int) {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNativeFanoutCrashProcess$")

	command.Env = append(os.Environ(), "CHR_NATIVE_FANOUT_DIR="+fixture.f.store.layout.ProjectDir(), "CHR_NATIVE_FANOUT_ID="+fixture.f.entry.InvocationID, "CHR_NATIVE_FANOUT_INDEX="+strconv.Itoa(index))
	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("owned fanout crash: %v: %s", err, output)
	}
}

func nativeFanoutSnapshot(t *testing.T, store *WorkflowStore) WorkflowSnapshot {
	t.Helper()

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	return snapshot
}

func checkNativeFanoutLater(t *testing.T, fixture nativeCanonicalFixture, before WorkflowSnapshot) {
	t.Helper()

	unit := fixture.f.candidate.Native.Unit
	for _, record := range before.Records {
		if record.Task != nil && record.Task.SliceID != unit.ID {
			after, err := fixture.f.store.Read(t.Context(), record.Path)
			if err != nil || !reflect.DeepEqual(record, after) {
				t.Fatal("fanout changed later work", err)
			}
		}

		if record.Set != nil {
			checkNativeFanoutSet(t, fixture.f.store, record, unit.ID)
		}
	}
}

func checkNativeFanoutSet(t *testing.T, store *WorkflowStore, before WorkflowRecord, unit string) {
	t.Helper()

	after, err := store.Read(t.Context(), before.Path)
	if err != nil || after.Set.Status != before.Set.Status || after.Set.Closing != nil {
		t.Fatal("fanout prematurely closed its owning set", err)
	}

	for i, descriptor := range before.Set.Slices {
		if descriptor.ID != unit && !reflect.DeepEqual(descriptor, after.Set.Slices[i]) {
			t.Fatal("fanout changed a later slice")
		}
	}
}
