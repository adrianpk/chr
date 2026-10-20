package webui

import (
	"net/http"
	"net/url"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

func (h handler) quizActionResult(w http.ResponseWriter, r *http.Request, request WriteActionRequest, result WriteActionResponse) bool {
	if request.Surface != "quiz" {
		return false
	}

	if result.Status != WriteStatusOK {
		h.renderQuizWriteFailure(w, r, result)

		return true
	}

	if result.RedirectURL == "/quizzes" || !htmx.IsRequest(r) {
		h.redirectAfterWrite(w, r, result.RedirectURL)

		return true
	}

	r.URL.RawQuery = "id=" + url.QueryEscape(request.ID)

	data, err := h.artifactDetailPageData(r, "quizzes")
	if err != nil {
		http.Error(w, "reload quiz; the operation may have committed", http.StatusInternalServerError)

		return true
	}

	data.Write = result

	htmx.Retarget(w, "#quiz-surface")
	htmx.Reswap(w, htmx.SwapOuter)

	if err := h.renderer.Fragment(w, "quiz_answer_content.html", data); err != nil {
		http.Error(w, "render quiz", http.StatusInternalServerError)
	}

	return true
}

func (h handler) renderQuizWriteFailure(w http.ResponseWriter, r *http.Request, result WriteActionResponse) {
	data := PageData{Shell: h.shellData("quizzes"), View: "write_result", Title: "Quiz action", Write: result}
	if htmx.IsRequest(r) {
		if err := h.renderer.Fragment(w, "quiz_action_result.html", data); err != nil {
			http.Error(w, "render quiz error", http.StatusInternalServerError)
		}

		return
	}

	h.renderWriteResult(w, r, result)
}
