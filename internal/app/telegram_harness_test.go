package app

import (
	"os"
	"strings"
	"testing"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramHarnessFixture(t *testing.T) (*telegramConversationAppFixture, service, string) {
	t.Helper()

	f := newTelegramConversationAppFixture(t)
	operations := telegramOperationConcreteService(t, f.receiver)

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("registered harness project unavailable", err)
	}

	operations.projectContext = commandProjectContext{ProjectID: projects[0].ID, ProjectDir: projects[0].Path, ProjectName: projects[0].Name}
	marker := prepareWebHarnessProfiles(t, operations)

	return f, operations, marker
}

// Real receiver HTTP/SQLite/application reads use the selected registry root,
// show role readiness and public names, and never call a model or launch work.
func TestTelegramHarnessSelectionRead(t *testing.T) {
	f, operations, marker := telegramHarnessFixture(t)
	if reply := f.send(t, "/harness", 0); !strings.Contains(reply.Text, "Select an authorized current project") {
		t.Fatal("inspection bypassed explicit project selection", reply)
	}

	f.send(t, "/project example-project", 0)

	reply := f.send(t, "/harness", 0)
	for _, expected := range []string{"Harness defaults for example-project", "Worker: alpha · Codex · model-a", "Generation: Legacy dispatch", "pi-writing · Pi · example/model-pi", "remote authentication not tested", "New operations only", "worker=PROFILE|-"} {
		if !strings.Contains(reply.Text, expected) {
			t.Fatal("inspection omitted public selection information", expected, reply.Text)
		}
	}

	for _, private := range []string{operations.env.ConfigDir(), "fixture-provider-secret", "credential_ref", "model_config_ref"} {
		if strings.Contains(reply.Text, private) {
			t.Fatal("read reply exposed private setup")
		}
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("inspection launched a driver")
	}

	if len(f.sessions(t)) != 0 {
		t.Fatal("read allocated a conversation")
	}
}

// The deterministic grammar accepts named role assignments and an unambiguous
// clear marker. Unsupported fields and duplicates never become model input.
func TestTelegramHarnessSelectionCommand(t *testing.T) {
	for _, test := range []struct {
		name, text string
		valid      bool
	}{
		{"inspect", "/harness", true},
		{"both roles", "/harness worker=beta generation=pi-writing", true},
		{"clear", "/harness worker=-", true},
		{"private field", "/harness credential_ref=secret", false},
		{"duplicate role", "/harness worker=alpha worker=beta", false},
		{"empty profile", "/harness generation=", false},
		{"too many fields", "/harness worker=alpha generation=beta extra=x", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsedHarness, err := parseTelegramHarness(telegramcore.ReceivedUpdate{Kind: "message", Text: test.text})

			choices, recognized := parsedHarness.Choices, parsedHarness.Recognized
			if !recognized || (err == nil) != test.valid {
				t.Fatal("incorrect deterministic parsing", choices, recognized, err)
			}

			if test.name == "clear" && (choices.Worker == nil || *choices.Worker != "") {
				t.Fatal("clear marker did not preserve an explicit empty choice")
			}
		})
	}
}
