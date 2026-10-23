package webui

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

// GuidedUseCaseStep names a user task in the shared case workflow.
type GuidedUseCaseStep struct {
	Label   string
	Current bool
}

func (h handler) guidedPageData(view GuidedUseCaseView) PageData {
	view.StateToken = guidedStateToken(view)

	query := url.Values{"namespace": {view.Namespace}}
	if view.CaseSlug != "" {
		query.Set("case", view.CaseSlug)
	} else {
		query.Set("id", view.RunID)
	}

	view.StateURL = "/cases/workflow/state?" + query.Encode()

	view.StageLabel, view.Hint = guidedStageCopy(view)
	for _, step := range []struct{ stage, label string }{{"name", "Name"}, {"dump", "Dump"}, {"quiz", "Quiz"}, {"specification", "Spec"}, {"preparation", "Plan"}, {"execution", "Development"}} {
		view.Steps = append(view.Steps, GuidedUseCaseStep{Label: step.label, Current: step.stage == view.Stage})
	}

	data := PageData{Shell: h.shellData("cases"), View: "guided", Title: "New workflow", Guided: view}
	if view.Name != "" {
		data.Title = view.Name
	}

	if view.Question != "" && !view.Suspended && !view.Processing && view.PendingAction == "" {
		data.Detail = ArtifactDetailView{Title: view.Name, Quiz: &QuizAnswerView{Question: view.Question, QuestionID: view.QuestionID, SessionID: view.RunID, DraftContext: view.Namespace, CanAnswer: true, Open: true, Started: true}}
	}

	return data
}

func guidedStageCopy(view GuidedUseCaseView) (string, string) {
	if view.Status == "completed" {
		return "Completed", "The results are ready to review."
	}

	if view.Status == "cancelled" {
		return "Cancelled", "This use case has stopped. Its saved work remains available."
	}

	if view.Suspended {
		return "Paused", "Your progress is saved. Resume when you are ready."
	}

	switch view.Stage {
	case "name":
		return "Name your use case", "What do you want to build or change?"
	case "dump":
		return "Dump", "Write everything you know about this use case. You can add more before continuing."
	case "quiz":
		return guidedQuizCopy(view)
	case "specification":
		return "Preparing the specification", "Your context and answers define the agreed change."
	case "preparation":
		return "Preparing the plan", "The agreed specification is being broken into tickets."
	default:
		return "Development", "Follow the work and review its results here."
	}
}

// The observation excludes fresh form tokens. Reading it never advances work.
func guidedStateToken(view GuidedUseCaseView) string {
	work := ""
	if view.Work != nil {
		work = fmt.Sprintf("%+v", *view.Work)
	}

	snapshot := fmt.Sprintf("%q", []string{view.Observation, view.Stage, view.Status, view.QuestionID, view.PendingAction, fmt.Sprint(view.Processing), work})

	return fmt.Sprintf("%x", sha256.Sum256([]byte(snapshot)))
}

func (h handler) guidedUseCaseState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	for key, values := range r.URL.Query() {
		limit := 128
		if key == "case" {
			limit = 1024
		}

		if len(values) != 1 || len(values[0]) > limit {
			http.Error(w, "invalid workflow parameter", http.StatusBadRequest)

			return
		}
	}

	view, err := h.readCaseWorkflow(r)
	if err != nil {
		http.Error(w, "Unable to refresh this workflow.", http.StatusBadRequest)

		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(struct {
		State string `json:"state"`
	}{guidedStateToken(view)}); err != nil {
		slog.DebugContext(r.Context(), "Workflow observation response interrupted")
	}
}

// Only a functional reference crosses screens; the owner resolves retained inputs.
func (h handler) readCaseWorkflow(r *http.Request) (GuidedUseCaseView, error) {
	query := r.URL.Query()
	if query.Get("case") == "" {
		return h.readGuidedUseCase(r.Context(), query.Get("id"), query.Get("namespace"))
	}

	if query.Get("id") != "" {
		return GuidedUseCaseView{}, errors.New("choose one use case")
	}

	provider, ok := h.provider.(CaseWorkflowProvider)
	if !ok {
		return GuidedUseCaseView{}, errors.New("use-case lookup unavailable")
	}

	return provider.CaseWorkflowView(r.Context(), query.Get("case"), query.Get("namespace"))
}

func guidedQuizCopy(view GuidedUseCaseView) (string, string) {
	if view.Question != "" && !view.Processing && view.PendingAction == "" {
		return "Quiz", "Answer each question to clarify the use case."
	}

	return "Preparing the quiz", "The next question will appear here automatically."
}
