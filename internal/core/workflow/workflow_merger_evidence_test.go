package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeMergerEvidence checks real retained files and containment, including
// symlink escapes and corruption with unchanged references after restart.
func TestNativeMergerEvidence(t *testing.T) {
	for _, mode := range []string{"valid", "corrupt", "symlink", "oversized", "foreign manifest"} {
		t.Run(mode, func(t *testing.T) {
			store, input, now := nativeMergerInput(t)
			path := filepath.Join(store.layout.Root(), input.TargetChangeRefs[0].Path)
			mutateNativeMergerEvidence(t, mode, path, store, &input)

			sealed, err := SealWorkflowMergerInput(input, now)
			if err != nil {
				t.Fatal(err)
			}

			peer := NewWorkflowStore(store.layout, store.projectID)

			err = peer.CheckWorkflowMergerEvidence(t.Context(), sealed, now)
			if (err == nil) != (mode == "valid") {
				t.Fatal("evidence integrity or containment was bypassed", err)
			}
		})
	}
}

func mutateNativeMergerEvidence(t *testing.T, mode, path string, store *Store, input *MergerInput) {
	t.Helper()

	switch mode {
	case "corrupt", "oversized":
		data := []byte("Changed target requirements")
		if mode == "oversized" {
			data = []byte(strings.Repeat("x", workflowstate.FileLimit+1))
		}

		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	case "symlink":
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		outside := filepath.Join(t.TempDir(), "target.json")
		if err := os.WriteFile(outside, data, 0o600); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(outside, path); err != nil {
			t.Fatal(err)
		}
	case "foreign manifest":
		manifest := NativeSourceManifest{Schema: "chr-native-source/v1", Repository: "https://forge.example.test/testing/other", SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: []string{input.Source}}
		input.SourceManifestRef = retainNativeRecordEvidence(t, store, input.InvocationID, "foreign-source", manifest)
	}
}

// TestNativeMergerCheckpoint binds the exact replay step and ordered prior
// mappings to the original manifest and separately retained checkpoint image.
func TestNativeMergerCheckpoint(t *testing.T) {
	for _, mode := range []string{"valid", "wrong step", "wrong parent", "wrong predecessor", "changed evidence", "merge checkpoint"} {
		t.Run(mode, func(t *testing.T) {
			store, input, now := nativeMergerInput(t)
			nativeMergerCheckpoint(t, store, &input)
			mutateNativeMergerCheckpoint(mode, &input)

			sealed, err := SealWorkflowMergerInput(input, now)
			if err == nil {
				err = store.CheckWorkflowMergerEvidence(t.Context(), sealed, now)
			}

			if (err == nil) != (mode == "valid") {
				t.Fatal("replay checkpoint lost its exact position", err)
			}
		})
	}
}

func nativeMergerCheckpoint(t *testing.T, store *Store, input *MergerInput) {
	t.Helper()

	input.Method = "rebase"
	previous := strings.Repeat("3", 40)
	prepared := strings.Repeat("4", 40)
	manifest := NativeSourceManifest{Schema: "chr-native-source/v1", Repository: input.Repository, SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: []string{previous, input.Source}}
	input.SourceManifestRef = retainNativeRecordEvidence(t, store, input.InvocationID, "replay-source", manifest)
	checkpoint := MergerCheckpoint{ReplayIndex: 1, OriginalCommit: input.Source, PreparedParent: prepared, Predecessors: []IntegrationReplay{{Source: previous, Commit: prepared, EmptyDecision: "already_applied"}}}
	checkpoint.Evidence = retainNativeRecordEvidence(t, store, input.InvocationID, "checkpoint", checkpoint)
	input.HistoryCheckpoint = &checkpoint
}

func mutateNativeMergerCheckpoint(mode string, input *MergerInput) {
	switch mode {
	case "wrong step":
		input.HistoryCheckpoint.OriginalCommit = strings.Repeat("5", 40)
	case "wrong parent":
		input.HistoryCheckpoint.PreparedParent = input.Base
	case "wrong predecessor":
		input.HistoryCheckpoint.Predecessors[0].Source = strings.Repeat("6", 40)
	case "changed evidence":
		input.HistoryCheckpoint.Predecessors[0].EmptyDecision = "nonempty"
	case "merge checkpoint":
		input.Method = "merge"
	}
}

// TestNativeMergerCancellation preserves context cancellation before reading
// host evidence; a cancelled check cannot act as admission or acceptance.
func TestNativeMergerCancellation(t *testing.T) {
	store, input, now := nativeMergerInput(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := store.CheckWorkflowMergerEvidence(ctx, input, now); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled host check read role evidence", err)
	}

	// Deadline expiry is an input error, independent of an otherwise live host.
	if err := store.CheckWorkflowMergerEvidence(t.Context(), input, input.Deadline); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("expired evidence check passed", err)
	}
}
