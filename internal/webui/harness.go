package webui

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// HarnessSelectionAction publishes exact project-scoped role choices.
const HarnessSelectionAction = "harnesses.change"

// HarnessKeepProfile omits a role change rather than clearing its selection.
const HarnessKeepProfile = ":keep"

// HarnessSelectionProvider supplies readable project choices and bindings.
type HarnessSelectionProvider interface {
	HarnessSelectionView(context.Context) (HarnessSelectionView, error)
}

// HarnessSelectionView separates defaults for new operations from retained work.
type HarnessSelectionView struct {
	ProjectID, Revision, RequestID string
	Worker, Generation             HarnessRoleView
	Retained                       []HarnessBindingView
	More, Modal                    bool
	Problem, RetainedProblem       string
	FormRevision, Status, Message  string
}

// HarnessRoleView contains one role's current default and named choices.
type HarnessRoleView struct {
	Current, Setup, Problem string
	Selected                string
	Choices                 []TicketOptionView
}

// HarnessBindingView contains a retained invocation's public worker identity.
type HarnessBindingView struct{ ID, Status, Label string }

func (h handler) harnesses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	provider, ok := h.provider.(HarnessSelectionProvider)
	if !ok {
		http.NotFound(w, r)

		return
	}

	if r.Method == http.MethodPost {
		h.harnessChange(w, r, provider)

		return
	}

	view, err := provider.HarnessSelectionView(r.Context())
	if err != nil {
		view.Problem = "Harness choices could not be loaded. Check private profile setup and reload."
	}

	h.renderHarnesses(w, r, view)
}

func (h handler) harnessChange(w http.ResponseWriter, r *http.Request, provider HarnessSelectionProvider) {
	request, err := harnessChangeRequest(w, r)
	if err != nil {
		http.Error(w, "invalid harness selection form", http.StatusBadRequest)

		return
	}

	result, err := h.provider.WriteAction(r.Context(), request)
	if err != nil {
		result = WriteActionResponse{Status: WriteStatusBlocked, Summary: "The selection change could not be acknowledged. Retry the same choices or reload current defaults."}
	}

	view, err := provider.HarnessSelectionView(r.Context())
	if err != nil {
		view.Problem = "Current defaults could not be loaded. Reload to inspect the saved state."
	}

	view.Status, view.Message = result.Status, result.Summary
	if result.BlockedReason != "" {
		view.Message += " " + result.BlockedReason
	}

	if len(result.Validation) > 0 {
		view.Message += " " + strings.Join(result.Validation, " ")
	}

	if result.Status != WriteStatusOK {
		view.FormRevision, view.RequestID = request.Revision, request.Fields["request_id"]
		retainHarnessChoice(&view.Worker, request.Fields["worker"])
		retainHarnessChoice(&view.Generation, request.Fields["generation"])
	}

	h.renderHarnesses(w, r, view)
}

func harnessChangeRequest(w http.ResponseWriter, r *http.Request) (WriteActionRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		return WriteActionRequest{}, err
	}

	allowed := map[string]bool{"action": true, "surface": true, "confirm": true, "revision": true, "project_ref": true, "request_id": true, "worker": true, "generation": true}
	if len(r.PostForm) != len(allowed) || r.URL.RawQuery != "" {
		return WriteActionRequest{}, errors.New("exact harness form fields are required")
	}

	for key, values := range r.PostForm {
		if !allowed[key] || len(values) != 1 {
			return WriteActionRequest{}, errors.New("duplicate or unsupported harness field")
		}
	}

	request := writeActionRequest(r)
	if request.Action != HarnessSelectionAction || request.Surface != "harnesses" {
		return WriteActionRequest{}, errors.New("unsupported harness action")
	}

	return request, nil
}

func retainHarnessChoice(role *HarnessRoleView, selected string) {
	role.Selected = selected
	if selected == HarnessKeepProfile || selected == "" {
		return
	}

	for _, choice := range role.Choices {
		if choice.Value == selected {
			return
		}
	}

	role.Choices = append(role.Choices, TicketOptionView{Value: selected, Label: "Requested profile: " + selected + " · unavailable"})
}

func (h handler) renderHarnesses(w http.ResponseWriter, r *http.Request, view HarnessSelectionView) {
	view.Modal = htmx.IsRequest(r)
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
