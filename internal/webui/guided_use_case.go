package webui

import (
	"context"
	"net/http"
)

// GuidedUseCaseProvider projects a retained journey through shared commands.
type GuidedUseCaseProvider interface {
	GuidedUseCaseView(context.Context, string, string) (GuidedUseCaseView, error)
}

// GuidedUseCaseLink names a resumable journey without exposing its internal identity.
type GuidedUseCaseLink struct{ Name, URL string }

// GuidedUseCaseView presents current capture, setup and an inspectable committed input.
type GuidedUseCaseView struct {
	Result                                                  *GuidedUseCaseResult
	Ideas                                                   []GuidedUseCaseIdea
	IdeaRequestID, IdeaRevision                             string
	CanRetainIdea                                           bool
	Controls                                                []GuidedUseCaseControl
	Suspended                                               bool
	ControlStatus                                           string
	Work                                                    *GuidedUseCaseWork
	Enabled                                                 bool
	RunID, Namespace, Name, Revision, RequestID             string
	Stage, Status, Capture, Commit, CommittedDump           string
	PendingAction, PendingContent                           string
	Question, QuestionID, InputDigest, RetryAction, Blocker string
	Artifacts                                               []GuidedUseCaseArtifact
	Policy, Problems                                        []string
	Runs                                                    []GuidedUseCaseLink
}

// GuidedUseCaseControl binds a lifecycle action to both displayed owner revisions.
type GuidedUseCaseControl struct {
	Action, Label, RequestID, Revision, NativeRevision string
	Content, QuestionID, InputDigest                   string
}

// GuidedUseCaseWork projects actual native progress without driving execution.
type GuidedUseCaseWork struct {
	Publication, IntegrationPhase                                string
	MaxWorkerAttempts, WorkerTimeoutSeconds                      int
	Status, Action, PullRequestURL                               string
	Total, Validated, Confirmed, Integrated, Remaining, Attempts int
}

// GuidedUseCaseIdea presents an attributable request pending separate review.
type GuidedUseCaseIdea struct{ Content, Actor, CreatedAt string }

// GuidedUseCaseResult presents only verified owner evidence and code locations.
type GuidedUseCaseResult struct {
	Complete     bool
	Summary      string
	Deliveries   []GuidedUseCaseDelivery
	Validation   []GuidedUseCaseValidation
	ClosingCount int
}

// GuidedUseCaseDelivery identifies confirmed code and its retained receipt.
type GuidedUseCaseDelivery struct{ Repository, Branch, Commit, PullRequest, Method, Reference, Hash string }

// GuidedUseCaseValidation projects host-observed validation without worker output.
type GuidedUseCaseValidation struct {
	Label, Outcome, Reference, Waiver string
	Commands                          []GuidedUseCaseValidationCommand
}

// GuidedUseCaseValidationCommand retains the actual argv display and exit code.
type GuidedUseCaseValidationCommand struct {
	Command  string
	ExitCode int
}

// GuidedUseCaseArtifact exposes one complete committed authoring document.
type GuidedUseCaseArtifact struct{ Stage, Reference, Commit, Content string }

func (h handler) guidedUseCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	query := r.URL.Query()
	for key, values := range query {
		if len(values) != 1 || len(values[0]) > 128 {
			http.Error(w, "invalid guided parameter: "+key, http.StatusBadRequest)

			return
		}
	}

	view, err := h.readGuidedUseCase(r.Context(), query.Get("id"), query.Get("namespace"))
	if err != nil {
		http.Error(w, "load guided use case: "+err.Error(), http.StatusBadRequest)

		return
	}

	data := PageData{Shell: h.shellData("guided"), View: "guided", Title: "Guided use cases", Guided: view}
	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render guided use case", http.StatusInternalServerError)
	}
}

func (h handler) guidedUseCaseAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 72<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "guided form bounds exceeded", http.StatusBadRequest)

		return
	}

	for _, values := range r.PostForm {
		if len(values) != 1 {
			http.Error(w, "duplicate guided parameter", http.StatusBadRequest)

			return
		}
	}

	if r.PostFormValue("surface") != "guided" || (r.PostFormValue("action") != "guided.start" && r.PostFormValue("action") != "guided.input") {
		http.Error(w, "unsupported guided action", http.StatusBadRequest)

		return
	}

	h.writeAction(w, r)
}

func (h handler) readGuidedUseCase(ctx context.Context, id, namespace string) (GuidedUseCaseView, error) {
	provider, ok := h.provider.(GuidedUseCaseProvider)
	if !ok {
		return GuidedUseCaseView{}, nil
	}

	return provider.GuidedUseCaseView(ctx, id, namespace)
}
