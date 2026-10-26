package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestScanShapes protects grouped names, receiver exclusion and nested contracts.
func TestScanShapes(t *testing.T) {
	cases := []struct {
		name       string
		source     string
		parameters int
		results    int
	}{
		{"grouped names", "func f(a, b, c int) (x, y string) { panic(0) }", 3, 2},
		{"variadic", "func f(a int, rest ...string) error { return nil }", 2, 1},
		{"receiver", "type T int; func (t *T) f(a int) {}", 1, 0},
		{"interface", "type T interface { f(a, b int) (int, string, error) }", 2, 3},
		{"function type", "type f func(a, b int) (int, bool)", 2, 2},
		{"literal", "var f = func(a, b int) bool { return false }", 2, 1},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			scan := scanFixture(t, map[string]string{"source.go": "package fixture\n" + item.source})
			if len(scan.Declarations) != 1 {
				t.Fatalf("declarations = %v", scan.Declarations)
			}

			value := scan.Declarations[0]
			if value.Parameters != item.parameters || len(value.Results) != item.results {
				t.Fatalf("signature counts = %+v", value)
			}
		})
	}
}

func scanFixture(t *testing.T, files map[string]string) sourceScan {
	t.Helper()

	root := t.TempDir()
	if output, err := exec.Command("git", "init", "--initial-branch=dev", root).CombinedOutput(); err != nil {
		t.Fatalf("initialize fixture: %s: %v", output, err)
	}

	for name, source := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	scan, err := scanSources(root)
	if err != nil {
		t.Fatal(err)
	}

	return scan
}
