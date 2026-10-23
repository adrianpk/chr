package webui

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// GuidedUseCasePath selects the existing product screen for the retained step.
func GuidedUseCasePath(stage string) string {
	switch stage {
	case "dump":
		return "/dumps"
	case "quiz":
		return "/quizzes"
	case "specification":
		return "/specs"
	case "preparation":
		return "/plans"
	case "execution", "results":
		return "/board"
	default:
		return "/cases"
	}
}

func (h handler) guidedPageData(r *http.Request, view GuidedUseCaseView) PageData {
	view.StateToken = guidedStateToken(view)

	query := url.Values{"namespace": {view.Namespace}}
	if view.CaseSlug != "" {
		query.Set("case", view.CaseSlug)
	} else {
		query.Set("id", view.RunID)
	}

	view.StateURL = "/cases/workflow/state?" + query.Encode()
	view.CurrentURL = GuidedUseCasePath(view.Stage) + "?" + query.Encode()
	view.Surface = guidedSurface(r.URL.Path, view.Stage)
	view.Starting = view.Surface == "cases" && (view.Stage == "name" || view.RunID == "") && !view.Suspended
	view.Following = "/"+view.Surface == GuidedUseCasePath(view.Stage)
	view.ReadOnly = !view.Following
	guidedPageActions(&view)

	data := PageData{Shell: h.shellData(view.Surface), View: "guided", Title: view.Name, Guided: view}
	if view.Starting {
		data.Title = "Cases"
	}

	guidedNavigation(&data)

	data.Detail = ArtifactDetailView{Family: view.Surface, Title: view.Name}
	data.Detail.ID = view.QuizID
	data.Detail.Quiz = guidedQuizView(view)

	data.Detail.Content = guidedDocument(view)

	return data
}

func guidedSurface(path, stage string) string {
	switch path {
	case "/cases", "/dumps", "/quizzes", "/specs", "/plans", "/board":
		return strings.TrimPrefix(path, "/")
	default:
		return strings.TrimPrefix(GuidedUseCasePath(stage), "/")
	}
}

func guidedNavigation(data *PageData) {
	if data.Guided.CaseSlug == "" {
		return
	}

	query := url.Values{"case": {data.Guided.CaseSlug}, "namespace": {data.Guided.Namespace}}
	for i, link := range data.Shell.Navigation {
		switch link.Href {
		case "/dumps", "/quizzes", "/specs", "/plans", "/board":
			data.Shell.Navigation[i].Href += "?" + query.Encode()
		}
	}
}

func guidedDocument(view GuidedUseCaseView) string {
	kinds := map[string]string{"dumps": "dump", "quizzes": "quiz", "specs": "specification", "plans": "plan"}
	for _, artifact := range view.Artifacts {
		if artifact.Stage == kinds[view.Surface] {
			return artifact.Content
		}
	}

	if view.Surface == "dumps" {
		return view.CommittedDump
	}

	return ""
}

func guidedMessage(view GuidedUseCaseView) string {
	if view.Status == "completed" {
		return ""
	}

	if view.Status == "cancelled" {
		return "Cancelled."
	}

	if view.Suspended {
		if view.Status == "paused" {
			return "Paused."
		}

		return "Stopping…"
	}

	if !view.Processing {
		return ""
	}

	switch view.Stage {
	case "quiz":
		return "Preparing the next question…"
	case "specification":
		return "Preparing the spec…"
	case "preparation":
		return "Preparing the plan…"
	default:
		return "Working…"
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

	query := url.Values{"case": {view.CaseSlug}, "namespace": {view.Namespace}}
	if view.CaseSlug == "" {
		query.Del("case")
		query.Set("id", view.RunID)
	}

	if err := json.NewEncoder(w).Encode(struct {
		State string `json:"state"`
		URL   string `json:"url"`
	}{guidedStateToken(view), GuidedUseCasePath(view.Stage) + "?" + query.Encode()}); err != nil {
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

func guidedQuizView(view GuidedUseCaseView) *QuizAnswerView {
	if view.Surface == "quizzes" && view.Following && !view.Suspended && !view.Processing && view.PendingAction == "" {
		return view.Quiz
	}

	return nil
}

// Input forms already own their actions. Lifecycle controls belong to the current
// automatic step; a pause has no useful effect while authoring awaits input.
func guidedVisibleControls(view GuidedUseCaseView) []GuidedUseCaseControl {
	if !guidedControlSurface(view) {
		return nil
	}

	controls := []GuidedUseCaseControl{}

	for _, control := range view.Controls {
		if control.Action == "pause" && control.Label == "Pause" && !guidedCanPause(view) {
			continue
		}

		control.Primary = control.Action == "resume" || strings.HasPrefix(control.Label, "Retry ")
		controls = append(controls, control)
	}

	sort.SliceStable(controls, func(i, j int) bool { return guidedControlBefore(controls[i], controls[j]) })

	return controls
}

func guidedCanPause(view GuidedUseCaseView) bool {
	if view.Stage != "execution" || view.Work == nil {
		return false
	}

	switch view.Work.Status {
	case "pending", "running", "waiting_review", "waiting_integration":
		return true
	default:
		return false
	}
}

func guidedPageActions(view *GuidedUseCaseView) {
	view.Message = guidedMessage(*view)
	view.Controls = guidedVisibleControls(*view)
	view.ShowControls = len(view.Controls) > 0
	view.ShowReview = view.Following && view.Work != nil && view.Work.Status == "waiting_review" && view.Work.PullRequestURL != ""
	view.ShowRetry = view.Following && !view.Suspended && !view.Processing && (view.PendingAction != "" || view.RetryAction != "")
}

func guidedWaitingForInput(view GuidedUseCaseView) bool {
	if view.Suspended || view.Processing {
		return false
	}

	return view.Surface == "dumps" || view.Surface == "quizzes" && view.Quiz != nil && view.Quiz.CanAnswer
}

func guidedControlBefore(left, right GuidedUseCaseControl) bool {
	if left.Primary != right.Primary {
		return !left.Primary
	}

	return left.Action == "cancel" && right.Action != "cancel"
}

func guidedControlSurface(view GuidedUseCaseView) bool {
	if view.Starting || view.ReadOnly || guidedWaitingForInput(view) {
		return false
	}

	// Document generation holds the authoring mutation until its result is
	// retained. It cannot currently acknowledge an immediate lifecycle stop.
	return !view.Processing || view.Suspended || view.Work != nil
}
