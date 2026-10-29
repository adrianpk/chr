package app

import (
	"testing"

	"golang.org/x/sys/unix"
)

// This named acceptance family selects the existing owned production boundaries
// once, including all five runtime uncertainties and the complete wizard path.
func TestTUIWebParity(t *testing.T) {
	if testing.Short() {
		t.Skip("owned parity acceptance is selected explicitly")
	}

	for _, test := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"navigation and legacy access", TestTUINavigationTerminal},
		{"case and dump forms", TestTUIFormsTerminal},
		{"dump conflict and reload", TestTUIFormsStaleTerminal},
		{"complete spec catalog", TestTUISpecCatalog},
		{"spec form and conflict", TestTUISpecTerminal},
		{"quiz controls and history", TestTUIQuizTerminal},
		{"plan and native tickets", TestTUIPlanTerminal},
		{"ticket fields and discussion", TestTUITicketTerminal},
		{"retained workflow controls", TestTUIWorkflowTerminal},
		{"palette availability", TestTUIParityCapabilities},
		{"wizard capture replay", TestTUIWizardDumpReplay},
		{"wizard cancellation", TestTUIWizardCancelTerminal},
		{"wizard polling", TestTUIWizardPollingTerminal},
		{"wizard verified delivery", tuiParityWizardTerminal},
	} {
		t.Run(test.name, test.run)
	}
}

func resizeTUIParityTerminal(t *testing.T, terminal *guidedTerminal, width, height uint16) {
	t.Helper()

	if err := unix.IoctlSetWinsize(int(terminal.master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Col: width, Row: height}); err != nil {
		t.Fatal(err)
	}
}
