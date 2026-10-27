package config

import (
	"context"
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// A definitive local preparation rejection preserves its cause and leaves no
// published image or receipt, unlike an uncertain failure after publication.
func TestTelegramHarnessSelectionPreparation(t *testing.T) {
	project := t.TempDir()
	request := harnessSelectionTestRequest("unprepared", contenthash.Sum(nil), "alpha")
	options := harnessSelectionTestOptions(request)
	cause := errors.New("fixture local preparation rejected")
	options.Validate = func(context.Context, Config) error { return cause }

	if _, err := ChangeHarnessSelections(t.Context(), project, options); !errors.Is(err, ErrHarnessSelectionPreparation) || !errors.Is(err, cause) {
		t.Fatal("prepublication rejection lost its identity", err)
	}

	if len(readHarnessSelectionTestConfig(t, project)) != 0 {
		t.Fatal("rejected preparation published configuration")
	}

	if receipt, err := readHarnessSelectionReceipt(project, request, ""); err != nil || receipt.Found {
		t.Fatal("rejected preparation retained a receipt", err)
	}
}
