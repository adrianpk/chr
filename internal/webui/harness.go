package webui

import (
	"context"
	"net/http"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// HarnessSelectionProvider supplies readable project choices and bindings.
type HarnessSelectionProvider interface {
	HarnessSelectionView(context.Context) (HarnessSelectionView, error)
}

// HarnessSelectionView separates defaults for new operations from retained work.
type HarnessSelectionView struct {
	ProjectID, Revision, RequestID string
	Worker, Generation             HarnessRoleView
	Retained                       []HarnessBindingView
	More                           bool
	Problem, RetainedProblem       string
}

// HarnessRoleView contains one role's current default and named choices.
type HarnessRoleView struct {
	Current, Setup, Problem string
	Choices                 []TicketOptionView
}

// HarnessBindingView contains a retained invocation's public worker identity.
type HarnessBindingView struct{ ID, Status, Label string }

func (h handler) harnesses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	provider, ok := h.provider.(HarnessSelectionProvider)
	if !ok {
		http.NotFound(w, r)

		return
	}

	view, err := provider.HarnessSelectionView(r.Context())
	if err != nil {
		view.Problem = "Harness choices could not be loaded. Check private profile setup and reload."
	}

	h.renderHarnesses(w, r, view)
}

func (h handler) renderHarnesses(w http.ResponseWriter, r *http.Request, view HarnessSelectionView) {
	data := PageData{Shell: h.shellData("project"), View: "harnesses", Title: "Harness defaults", Harnesses: view}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if htmx.IsRequest(r) {
		if err := h.renderer.Fragment(w, "harness_modal.html", data); err != nil {
			http.Error(w, "render harness defaults", http.StatusInternalServerError)
		}

		return
	}

	if err := h.renderer.Page(w, "shell.html", data); err != nil {
		http.Error(w, "render harness defaults", http.StatusInternalServerError)
	}
}
