package webui

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// WorkflowProvider exposes shared workflow operations through Web projections.
type WorkflowProvider interface {
	WorkflowView(context.Context, WorkflowViewRequest) (WorkflowView, error)
}

// WorkflowViewRequest selects a bounded preview or retained status page.
type WorkflowViewRequest struct {
	Extent, Identity, TicketKey, InvocationID string
	ProjectArea, Severity, Priority, Order    string
	Offset                                    int
	Revision                                  uint64
}

// WorkflowView contains read-only previews, progress and visible failures.
type WorkflowView struct {
	Enabled bool
	Failure string
	Preview *WorkflowPreviewView
	Current *WorkflowRunView
	Items   []WorkflowRunView
	NextURL string
}

// WorkflowFieldView is one public workflow fact.
type WorkflowFieldView struct{ Label, Value string }

// WorkflowPreviewView binds the displayed selection to explicit confirmation.
type WorkflowPreviewView struct {
	Binding string
	Fields  []WorkflowFieldView
	Members []WorkflowFieldView
}

// WorkflowRunView exposes progress without private ownership credentials.
type WorkflowRunView struct {
	ID, Revision, Status, ApprovalHash string
	Fields, Members, Evidence          []WorkflowFieldView
	NextURL                            string
	Actions                            []string
	CanApprove                         bool
	Integration                        *WorkflowIntegrationControlView
}

// WorkflowIntegrationControlView carries public displayed guards and choices.
type WorkflowIntegrationControlView struct {
	EntryID, EntryRevision, BudgetRevision, CandidateHash string
	GrantID, DecisionRevision, DecisionHash               string
	GrantKinds                                            []string
	FindingCategories                                     []string
	Questions                                             []WorkflowIntegrationQuestionView
}

// WorkflowIntegrationQuestionView requires an explicit answer to one question.
type WorkflowIntegrationQuestionView struct {
	ID, Issue, Authority, AnswerID string
	Options                        []string
	ReplayOwnership                bool
}

func (h handler) workflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	view := h.workflowReadView(r)

	data := PageData{Shell: h.shellData("dash"), View: "workflows", Title: "Workflows", Workflow: view}
	if view.Enabled && view.Preview == nil && view.Failure == "" {
		data.RefreshURL = "/workflows/fragment?" + r.URL.RawQuery
	}

	if r.URL.Path == "/workflows/fragment" {
		if err := h.renderer.Fragment(w, "workflow_content.html", data); err != nil {
			http.Error(w, "render workflows", http.StatusInternalServerError)
		}

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render workflows", http.StatusInternalServerError)
	}
}

func (h handler) workflowReadView(r *http.Request) WorkflowView {
	view := WorkflowView{Failure: "Workflow controls are unavailable."}

	provider, ok := h.provider.(WorkflowProvider)
	if ok {
		request, err := workflowViewRequest(r)
		if err == nil {
			view, err = provider.WorkflowView(r.Context(), request)
		}

		if err != nil {
			view.Failure = err.Error()
		}
	}

	return view
}

func workflowViewRequest(r *http.Request) (WorkflowViewRequest, error) {
	query := r.URL.Query()
	for key, values := range query {
		if len(values) != 1 || len(values[0]) > 128 {
			return WorkflowViewRequest{}, fmt.Errorf("invalid workflow parameter: %s", key)
		}
	}

	offset, err := workflowQueryNumber(query.Get("offset"), 4096)
	if err != nil {
		return WorkflowViewRequest{}, err
	}

	revision, err := workflowQueryNumber(query.Get("revision"), ^uint64(0))
	if err != nil {
		return WorkflowViewRequest{}, err
	}

	return WorkflowViewRequest{Extent: query.Get("extent"), Identity: query.Get("identity"), TicketKey: query.Get("key"), InvocationID: query.Get("id"), ProjectArea: query.Get("project_area"), Severity: query.Get("severity"), Priority: query.Get("priority"), Order: query.Get("order"), Offset: int(offset), Revision: revision}, nil
}

func workflowQueryNumber(value string, limit uint64) (uint64, error) {
	if value == "" {
		return 0, nil
	}

	number, err := strconv.ParseUint(value, 10, 64)
	if err != nil || number > limit {
		return 0, fmt.Errorf("invalid workflow page or revision")
	}

	return number, nil
}

func (h handler) workflowAction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "workflow form bounds exceeded", http.StatusBadRequest)

		return
	}

	if !strings.HasPrefix(r.PostFormValue("action"), "workflow.") {
		http.Error(w, "unsupported workflow action", http.StatusBadRequest)

		return
	}

	for _, values := range r.PostForm {
		if len(values) != 1 {
			http.Error(w, "duplicate workflow parameter", http.StatusBadRequest)

			return
		}
	}

	h.writeAction(w, r)
}
