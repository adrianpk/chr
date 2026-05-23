package chcapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceFilesDoNotImportTUIAdapter(t *testing.T) {
	files, err := filepath.Glob("service*.go")
	if err != nil {
		t.Fatalf("glob service files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no service files found")
	}

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		if strings.Contains(string(content), "internal/chctui") {
			t.Fatalf("%s imports internal/chctui; service files must stay adapter-neutral", path)
		}
	}
}
