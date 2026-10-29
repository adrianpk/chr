package app

import (
	"errors"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A real editor process changes only its owned private draft. Success, process
// failure and oversized output all clean the session without canonical writes.
func TestServiceDumpEditor(t *testing.T) {
	for _, mode := range []string{"success", "failed", "oversized"} {
		t.Run(mode, func(t *testing.T) {
			checkServiceDumpEditor(t, mode)
		})
	}
}

func checkDumpEditorDisposition(t *testing.T, mode, content string, err error) {
	t.Helper()

	if mode == "success" {
		if err != nil || content != "UNSENT_ORIGINAL\nEXTERNAL_EDIT\n" {
			t.Fatal(content, err)
		}
	} else if err == nil || content != "" {
		t.Fatal("invalid editor output was accepted", mode, err)
	}
}

func checkServiceDumpEditor(t *testing.T, mode string) {
	t.Helper()

	operations := newTestService(t, newTestWorkspace(t))
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	editor := &fakeDumpEditorLauncher{command: DumpEditorCommand{Program: "sh", Args: []string{"-c", "true"}}}
	operations.dumpEditor = editor

	session, err := operations.OpenDumpDraftEditor("UNSENT_ORIGINAL")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupErr := errors.New("fixture cleanup")

		_, finishErr := session.Finish(cleanupErr)
		if !errors.Is(finishErr, cleanupErr) {
			t.Error("editor cleanup lost its disposition", finishErr)
		}
	})

	runErr := error(nil)

	switch mode {
	case "success":
		command := exec.CommandContext(t.Context(), "sh", "-c", `printf '\nEXTERNAL_EDIT\n' >> "$1"`, "editor", editor.path)
		runErr = command.Run()
	case "failed":
		runErr = errors.New("editor failed")
	case "oversized":
		if err = os.WriteFile(editor.path, []byte(strings.Repeat("x", (1<<20)+1)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	content, err := session.Finish(runErr)
	checkDumpEditorDisposition(t, mode, content, err)

	if _, err = os.Stat(filepath.Dir(editor.path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("draft directory leaked", err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("external draft editing wrote canonical files")
	}
}
