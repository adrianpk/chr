package authoring

import (
	"errors"
	"fmt"
	"testing"
)

// Alias lookup follows a bounded cross-namespace chain and rejects dangling,
// cyclic, overlong, raw and scoped destinations without rewriting old bodies.
func TestProjectDumpAliases(t *testing.T) {
	for _, name := range []string{"chain", "cycle", "missing", "raw", "scoped", "too long"} {
		t.Run(name, func(t *testing.T) {
			project, old, target := prepareProjectDumpAlias(t, name)

			resolved, err := ResolveProjectDumpReference(t.Context(), project, old)
			if name == "chain" {
				if err != nil || resolved.Reference != target {
					t.Fatal("cross-namespace alias failed", resolved, err)
				}
			} else if !errors.Is(err, ErrProjectDumpReconciliation) {
				t.Fatal("invalid alias accepted", resolved, err)
			}

			if _, err := ObserveProjectDump(t.Context(), project); name != "chain" && !errors.Is(err, ErrProjectDumpReconciliation) {
				t.Fatal("inventory accepted invalid aliases", err)
			}
		})
	}
}

func prepareProjectDumpAlias(t *testing.T, name string) (string, string, string) {
	t.Helper()

	project := t.TempDir()
	old := projectDumpTestArtifact(t, project, "default", "Old", "Original retained bytes.\n")
	target := projectDumpTestArtifact(t, project, "other", "Current", "Current bytes.\n")
	patchProjectDumpTestMetadata(t, project, old, map[string]any{"canonical_ref": target})

	switch name {
	case "cycle":
		patchProjectDumpTestMetadata(t, project, target, map[string]any{"canonical_ref": old})
	case "missing":
		patchProjectDumpTestMetadata(t, project, old, map[string]any{"canonical_ref": "ops/default/dump/missing/artifact.md"})
	case "raw":
		writeTestFile(t, project+"/"+target, "---\ntype: dump_raw\n---\nRaw bytes.\n")
	case "scoped":
		writeTestFile(t, project+"/"+target, "---\ncase: named-work\n---\nScoped bytes.\n")
	case "too long":
		for i := range projectDumpAliasLimit {
			next := projectDumpTestArtifact(t, project, "other", fmt.Sprintf("Alias %d", i), "Retained bytes.\n")
			patchProjectDumpTestMetadata(t, project, target, map[string]any{"canonical_ref": next})

			target = next
		}
	}

	return project, old, target
}
