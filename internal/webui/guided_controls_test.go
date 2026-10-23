package webui

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// TestCaseWorkflowActions keeps lifecycle actions with the state they control,
// excludes redundant authoring pauses and leaves the shared input forms intact.
func TestCaseWorkflowActions(t *testing.T) {
	for _, test := range []struct {
		name, path string
		view       GuidedUseCaseView
		want       []string
	}{
		{"dump input", "/dumps", GuidedUseCaseView{Stage: "dump"}, nil},
		{"quiz input", "/quizzes", GuidedUseCaseView{Stage: "quiz", Quiz: &QuizAnswerView{Open: true, Started: true, CanAnswer: true}}, nil},
		{"quiz processing", "/quizzes", GuidedUseCaseView{Stage: "quiz", Processing: true}, nil},
		{"spec processing", "/specs", GuidedUseCaseView{Stage: "specification", Processing: true}, nil},
		{"plan recovery", "/plans", GuidedUseCaseView{Stage: "preparation", PendingAction: "admit", Status: "blocked"}, []string{"cancel", "admit"}},
		{"prior document", "/dumps", GuidedUseCaseView{Stage: "preparation", PendingAction: "admit", Status: "blocked"}, nil},
		{"paused authoring", "/quizzes", GuidedUseCaseView{Stage: "quiz", Suspended: true, Status: "paused"}, []string{"cancel", "resume"}},
		{"execution", "/board", GuidedUseCaseView{Stage: "execution", Work: &GuidedUseCaseWork{Status: "running"}}, []string{"pause", "cancel"}},
		{"blocked execution", "/board", GuidedUseCaseView{Stage: "execution", Work: &GuidedUseCaseWork{Status: "blocked"}}, []string{"cancel"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			view := test.view
			view.RunID, view.Name = "retained-workflow", "Checkout"

			view.Controls = []GuidedUseCaseControl{{Action: "pause", Label: "Pause"}, {Action: "cancel", Label: "Cancel"}}
			if view.Suspended {
				view.Controls[0] = GuidedUseCaseControl{Action: "resume", Label: "Resume"}
			}

			handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: view}})
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path+"?case=checkout", nil))

			if response.Code != http.StatusOK {
				t.Fatal(response.Code, response.Body.String())
			}

			assertCaseWorkflowActions(t, response.Body.String(), test.want)
		})
	}
}

func assertCaseWorkflowActions(t *testing.T, body string, expected []string) {
	t.Helper()

	panels := regexp.MustCompile(`(?s)<section class="workflow-progress"[^>]*>.*?</section>`).FindAllString(body, -1)

	var actions []string

	for _, panel := range panels {
		footers := regexp.MustCompile(`(?s)<footer class="ticket-editor-footer">.*?</footer>`).FindAllString(panel, -1)
		for _, footer := range footers {
			for _, field := range regexp.MustCompile(`name="step" value="([^"]+)"`).FindAllStringSubmatch(footer, -1) {
				actions = append(actions, field[1])
			}
		}
	}

	slices.Sort(actions)

	want := slices.Clone(expected)
	slices.Sort(want)

	if !slices.Equal(actions, want) {
		t.Fatalf("wrong contextual actions: got %v; want %v", actions, want)
	}

	if strings.Contains(body, `class="workflow-controls"`) {
		t.Fatal("standalone lifecycle toolbar remains")
	}

	if slices.Contains(expected, "cancel") && !strings.Contains(body, "data-workflow-cancel") {
		t.Fatal("workflow cancellation lacks its explicit confirmation")
	}
}
