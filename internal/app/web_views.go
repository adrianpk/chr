//nolint:wsl_v5 // View assembly keeps adjacent parsing and projection steps together.
package app

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

type webViewProvider struct {
	env            core.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	config         resolvedWebServeConfig
	serviceFactory func(context.Context) service
}

func (p webViewProvider) service(ctx context.Context) service {
	operations := newService(ctx, p.env, p.projectContext, p.assetsFS)
	if p.serviceFactory != nil {
		operations = p.serviceFactory(ctx)
	}
	if p.config.guidedHost != nil {
		operations.guidedWorkHost = p.config.guidedHost
	}

	return operations
}

func (p webViewProvider) ProjectView(context.Context) (webui.ProjectView, error) {
	view := webui.ProjectView{
		ProjectID:  p.projectContext.ProjectID,
		ProjectDir: p.projectContext.ProjectDir,
	}
	view.Revision = webStableRevision(view)

	return view, nil
}

func (p webViewProvider) BoardView(ctx context.Context, request webui.BoardRequest) (webui.BoardView, error) {
	operations := p.service(ctx)
	if request.Namespace != "" {
		operations.workflowNamespace = request.Namespace
	}
	board, err := operations.readTicketBoardPage(request.Offset, request.CaseSlug)
	if err != nil {
		return webui.BoardView{}, err
	}

	view := webBoardView(board)
	view.LoadedCount = board.LoadedCount
	if board.NextOffset >= 0 {
		view.NextURL = "/board?offset=" + strconv.Itoa(board.NextOffset)
	}
	if request.Offset > 0 {
		view.PreviousURL = "/board?offset=" + strconv.Itoa(max(0, request.Offset-workflowstate.PageLimit))
	}
	view.Revision = webStableRevision(struct {
		Source string
		Offset int
	}{board.Revision, request.Offset})

	return view, nil
}

func (p webViewProvider) TicketDetailView(ctx context.Context, request webui.TicketDetailRequest) (webui.TicketDetailView, error) {
	if request.New {
		return p.completeTicketEditor(webui.TicketDetailView{New: true, Fields: map[string]string{"status": string(ticketcore.StatusBacklog), "priority": "normal", "severity": "unclassified", "type": "task"}})
	}

	document, err := p.service(ctx).ReadTicketDocument(request.Key)
	if err != nil {
		return webui.TicketDetailView{}, err
	}

	ticket := document.Ticket
	view := webui.TicketDetailView{
		NativeWorkflow:  ticket.NativeWorkflow,
		WorkflowEnabled: p.workflowEnabled(),
		Key:             ticket.Key,
		Title:           ticket.Title,
		Priority:        ticket.Priority,
		Status:          string(ticket.Status),
		Type:            ticket.Type,
		ProjectArea:     ticket.ProjectArea,
		UseCase:         ticket.UseCase,
		Slice:           ticket.Slice,
		RelativePath:    ticket.RelativePath,
		Synopsis:        document.Fields["synopsis"],
		Description:     document.Description,
		Fields:          document.Fields,
		ReadOnly:        document.ReadOnly,
		ReadOnlyReason:  document.ReadOnlyReason,
		Phase:           document.Phase,
		Comments:        parseWebTicketComments(document.Discussion),
	}
	view.Revision = document.Revision

	return p.completeTicketEditor(view)
}

//nolint:cyclop,gocognit // The linear parser mirrors the ticket comments section grammar.
func parseWebTicketComments(content string) []webui.TicketCommentView {
	var comments []webui.TicketCommentView
	var current *webui.TicketCommentView
	inComments := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "# Comments" {
			inComments = true

			continue
		}

		if !inComments {
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			break
		}

		if strings.HasPrefix(trimmed, "## ") {
			if current != nil {
				comments = append(comments, trimWebTicketComment(*current))
			}
			current = webTicketCommentHeader(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))

			continue
		}

		if current == nil {
			continue
		}

		if current.Body != "" || trimmed != "" {
			current.Body += line + "\n"
		}
	}

	if current != nil {
		comments = append(comments, trimWebTicketComment(*current))
	}

	return comments
}

func trimWebTicketComment(comment webui.TicketCommentView) webui.TicketCommentView {
	comment.Author = strings.TrimSpace(comment.Author)
	comment.CreatedAt = strings.TrimSpace(comment.CreatedAt)
	comment.Body = strings.TrimSpace(comment.Body)

	return comment
}

func webTicketCommentHeader(header string) *webui.TicketCommentView {
	header = strings.TrimSpace(header)

	parts := strings.Fields(header)
	if len(parts) < 2 {
		return &webui.TicketCommentView{Author: header}
	}

	if createdAt, err := time.Parse(time.RFC3339, parts[0]); err == nil {
		return &webui.TicketCommentView{
			Author:    strings.Join(parts[1:], " "),
			CreatedAt: createdAt.Format("02 Jan 15:04"),
		}
	}

	return &webui.TicketCommentView{Author: header}
}

func webBoardView(board ticketcore.Board) webui.BoardView {
	columns := make([]webui.BoardColumnView, 0, len(board.Columns))
	for _, column := range board.Columns {
		columns = append(columns, webui.BoardColumnView{
			ID:      column.ID,
			Title:   column.Title,
			Count:   column.Count,
			Tickets: webTicketViews(column.Tickets),
		})
	}

	attention := make([]webui.BoardAttentionView, 0, len(board.Attention))
	for _, item := range board.Attention {
		attention = append(attention, webui.BoardAttentionView{
			Ticket:  webTicketView(item.Ticket),
			Reasons: append([]string(nil), item.Reasons...),
		})
	}

	return webui.BoardView{
		ProjectID:       board.ProjectID,
		TotalCount:      board.TotalCount,
		NormalCount:     board.NormalCount,
		AttentionCount:  board.AttentionCount,
		Columns:         columns,
		Attention:       attention,
		NormalColumnIDs: append([]string(nil), board.NormalColumnIDs...),
	}
}

func webTicketViews(tickets []ticketcore.Indexed) []webui.BoardTicketView {
	views := make([]webui.BoardTicketView, 0, len(tickets))
	for _, ticket := range tickets {
		views = append(views, webTicketView(ticket))
	}

	return views
}

func webTicketView(ticket ticketcore.Indexed) webui.BoardTicketView {
	return webui.BoardTicketView{
		Revision:          ticket.ContentHash,
		Assignee:          ticket.Assignee,
		Key:               ticket.Key,
		Title:             ticket.Title,
		Synopsis:          ticket.Synopsis,
		Status:            string(ticket.Status),
		BoardColumn:       string(ticket.BoardColumn),
		Priority:          ticket.Priority,
		Severity:          ticket.Severity,
		Type:              ticket.Type,
		UseCase:           ticket.UseCase,
		ProjectArea:       ticket.ProjectArea,
		Slice:             ticket.Slice,
		Branch:            ticket.Branch,
		ReviewState:       string(ticket.ReviewState),
		ReviewRound:       ticket.ReviewRound,
		ReviewResult:      string(ticket.ReviewResult),
		ReviewCommentsURL: ticket.ReviewCommentsURL,
		MergeState:        string(ticket.MergeState),
		MergeResult:       string(ticket.MergeResult),
		MergeBlocker:      ticket.MergeBlocker,
		RelativePath:      ticket.RelativePath,
		SourceKind:        ticket.SourceKind,
		SourceRef:         ticket.SourceRef,
	}
}
