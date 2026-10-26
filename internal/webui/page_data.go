// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/webui/ui"
)

func (h handler) projectPageData(r *http.Request) (PageData, error) {
	project, err := h.provider.ProjectView(r.Context())
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:    h.shellData("project"),
		View:     "project",
		Title:    "Project",
		Revision: project.Revision,
		Project:  project,
	}, nil
}

func (h handler) boardPageData(r *http.Request) (PageData, error) {
	offset := 0
	if value := r.URL.Query().Get("offset"); value != "" {
		var err error
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 || offset > 4096 {
			return PageData{}, fmt.Errorf("invalid Board page offset")
		}
	}
	board, err := h.provider.BoardView(r.Context(), BoardRequest{Offset: offset, Namespace: r.URL.Query().Get("namespace"), CaseSlug: r.URL.Query().Get("case")})
	if err != nil {
		return PageData{}, err
	}

	refresh := "/board/fragment"
	if offset != 0 {
		refresh += "?offset=" + strconv.Itoa(offset)
	}

	return PageData{
		Shell:      h.shellData("board"),
		View:       "board",
		Title:      "Board",
		Revision:   board.Revision,
		RefreshURL: refresh,
		Board:      board,
	}, nil
}

func (h handler) ticketDetailPageData(r *http.Request) (PageData, error) {
	ticket, err := h.provider.TicketDetailView(r.Context(), TicketDetailRequest{
		Key: strings.TrimSpace(r.URL.Query().Get("key")),
		New: r.URL.Query().Get("new") == "1",
	})
	if err != nil {
		return PageData{}, err
	}

	return PageData{
		Shell:    h.shellData("board"),
		View:     "ticket_detail",
		Title:    ticket.Title,
		Revision: ticket.Revision,
		Ticket:   ticket,
	}, nil
}

func ticketDetailRequest(r *http.Request, key string) *http.Request {
	clone := r.Clone(r.Context())
	query := clone.URL.Query()
	query.Set("key", key)
	clone.URL.RawQuery = query.Encode()

	return clone
}

func (h handler) dashPageData(r *http.Request) (PageData, error) {
	dash, err := h.readDashPage(r)
	if err != nil {
		return PageData{}, err
	}

	return h.dashPageProjection(r, dash), nil
}

func (h handler) catalogPageData(r *http.Request, family string) (PageData, error) {
	catalog, err := h.provider.CatalogView(r.Context(), family)
	if err != nil {
		return PageData{}, err
	}

	guided := GuidedUseCaseView{}
	if family == "cases" {
		guided, err = h.readGuidedUseCase(r.Context(), "", r.URL.Query().Get("namespace"))
		if err != nil {
			return PageData{}, err
		}
		workflows := make(map[string]string, len(guided.Runs))
		for _, run := range guided.Runs {
			workflows[run.CaseSlug] = run.URL
		}
		for i := range catalog.Items {
			catalog.Items[i].WorkflowURL = workflows[catalog.Items[i].Scope]
		}
	}

	return PageData{
		Guided:     guided,
		Shell:      h.shellData(family),
		View:       "catalog",
		Title:      catalog.Title,
		Revision:   catalog.Revision,
		RefreshURL: "/" + family + "/fragment",
		Catalog:    catalog,
	}, nil
}

func (h handler) artifactDetailPageData(r *http.Request, family string) (PageData, error) {
	detail, err := h.provider.ArtifactDetailView(r.Context(), ArtifactDetailRequest{
		Family: family,
		ID:     r.URL.Query().Get("id"),
	})
	if err != nil {
		return PageData{}, err
	}

	title := detail.Title
	if detail.Quiz != nil {
		title = "Quizzes"
	}
	if detail.Spec != nil {
		title = "Specs"
	}

	return PageData{
		Shell:      h.shellData(family),
		View:       "artifact_detail",
		Title:      title,
		Revision:   detail.Revision,
		RefreshURL: artifactDetailRefreshURL(family, r.URL.Query().Get("id")),
		Detail:     detail,
	}, nil
}

func artifactDetailRefreshURL(family, id string) string {
	if family == "cases" || family == "dumps" || family == "quizzes" || family == "specs" {
		return ""
	}

	return "/" + family + "/detail/fragment?id=" + url.QueryEscape(id)
}

func (h handler) shellData(active string) ShellData {
	data := h.data
	_, data.HarnessesEnabled = h.provider.(HarnessSelectionProvider)
	if strings.TrimSpace(data.ProjectName) == "" {
		data.ProjectName = data.ProjectID
	}
	data.ProjectName = displayProjectName(data.ProjectName)

	data.Navigation = make([]ui.NavLink, 0, len(h.data.Navigation))

	for _, link := range h.data.Navigation {
		link.Active = navLinkActive(active, link.Href)
		data.Navigation = append(data.Navigation, link)
	}

	return data
}

func displayProjectName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	first := name[0]
	if first >= 'a' && first <= 'z' {
		return string(first-'a'+'A') + name[1:]
	}

	return name
}

func navLinkActive(active, href string) bool {
	switch active {
	case "project":
		return href == "/" || href == "/project"
	case "board":
		return href == "/board"
	case "dash":
		return href == "/dash"
	case "cases", "dumps", "quizzes", "specs", "plans":
		return href == "/"+active
	default:
		return false
	}
}
