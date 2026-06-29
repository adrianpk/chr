package app

import (
	"context"
	"io/fs"
	"path/filepath"

	"forge.adrianpk.com/chc/internal/core"
	"forge.adrianpk.com/chc/internal/tui"
	"forge.adrianpk.com/chc/internal/webui"
)

type webViewProvider struct {
	env            core.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	config         resolvedWebServeConfig
}

func (p webViewProvider) ProjectView(context.Context) (webui.ProjectView, error) {
	return webui.ProjectView{
		ProjectID:      p.projectContext.ProjectID,
		ProjectDir:     p.projectContext.ProjectDir,
		URL:            p.config.URL,
		Host:           p.config.Web.Host,
		Port:           p.config.Web.Port,
		RequireToken:   string(p.config.Web.RequireToken),
		RefreshSeconds: p.config.Web.RefreshIntervalSeconds,
		Paths: []webui.ProjectPathView{
			{Label: "Workspace", Path: filepath.Join(p.projectContext.ProjectDir, core.DirectoryName)},
			{Label: "User workspace", Path: filepath.Join(p.projectContext.ProjectDir, core.DirectoryName, core.UserDirectoryName)},
			{Label: "Tickets", Path: filepath.Join(p.projectContext.ProjectDir, core.DirectoryName, core.UserDirectoryName, "tickets")},
			{Label: "Runtime", Path: filepath.Join(p.projectContext.ProjectDir, core.DirectoryName, core.UserDirectoryName, "runtime")},
		},
	}, nil
}

func (p webViewProvider) BoardView(ctx context.Context) (webui.BoardView, error) {
	board, err := newService(ctx, p.env, p.projectContext, p.assetsFS).Board()
	if err != nil {
		return webui.BoardView{}, err
	}

	return webBoardView(board), nil
}

func (p webViewProvider) DashView(ctx context.Context) (webui.DashView, error) {
	operations := newService(ctx, p.env, p.projectContext, p.assetsFS)

	board, err := operations.Board()
	if err != nil {
		return webui.DashView{}, err
	}

	readiness, err := loadTUIWorkReadiness(ctx, p.projectContext)
	if err != nil {
		return webui.DashView{}, err
	}

	observability, err := loadTUIObservability(ctx, p.env, p.projectContext, p.assetsFS)
	if err != nil {
		return webui.DashView{}, err
	}

	queue := readyWebQueue(board.Tickets)

	return webui.DashView{
		Readiness:       webWorkReadinessView(readiness),
		Queue:           webui.AgentQueueView{EligibleCount: len(queue), Tickets: queue},
		Current:         webCurrentWorkView(observability.LatestEvent),
		RecentEvents:    webAgentEventViews(observability.RecentEvents, 8),
		LatestLoop:      webLoopRunView(observability.LatestLoop),
		LatestExecution: webExecutionView(observability.LatestExecution),
	}, nil
}

func webBoardView(board core.Board) webui.BoardView {
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

func webTicketViews(tickets []core.IndexedTicket) []webui.BoardTicketView {
	views := make([]webui.BoardTicketView, 0, len(tickets))
	for _, ticket := range tickets {
		views = append(views, webTicketView(ticket))
	}

	return views
}

func webTicketView(ticket core.IndexedTicket) webui.BoardTicketView {
	return webui.BoardTicketView{
		Key:               ticket.Key,
		Title:             ticket.Title,
		Synopsis:          ticket.Synopsis,
		Status:            ticket.Status,
		BoardColumn:       ticket.BoardColumn,
		Priority:          ticket.Priority,
		Severity:          ticket.Severity,
		Type:              ticket.Type,
		UseCase:           ticket.UseCase,
		ProjectArea:       ticket.ProjectArea,
		Slice:             ticket.Slice,
		Branch:            ticket.Branch,
		ReviewState:       ticket.ReviewState,
		ReviewRound:       ticket.ReviewRound,
		ReviewResult:      ticket.ReviewResult,
		ReviewCommentsURL: ticket.ReviewCommentsURL,
		MergeState:        ticket.MergeState,
		MergeResult:       ticket.MergeResult,
		MergeBlocker:      ticket.MergeBlocker,
		RelativePath:      ticket.RelativePath,
		SourceKind:        ticket.SourceKind,
		SourceRef:         ticket.SourceRef,
	}
}

func readyWebQueue(tickets []core.IndexedTicket) []webui.BoardTicketView {
	queue := []webui.BoardTicketView{}
	for _, ticket := range tickets {
		if ticket.Status == core.TicketStatusReadyForDev {
			queue = append(queue, webTicketView(ticket))
		}
	}

	return queue
}

func webWorkReadinessView(view tui.WorkReadinessView) webui.WorkReadinessView {
	diagnostics := make([]webui.WorkReadinessDiagnosticView, 0, len(view.Diagnostics))
	for _, diagnostic := range view.Diagnostics {
		diagnostics = append(diagnostics, webui.WorkReadinessDiagnosticView{
			Code:            diagnostic.Code,
			Severity:        diagnostic.Severity,
			Detail:          diagnostic.Detail,
			RepairAvailable: diagnostic.RepairAvailable,
			RepairSafe:      diagnostic.RepairSafe,
			RepairDetail:    diagnostic.RepairDetail,
		})
	}

	return webui.WorkReadinessView{
		Status:            view.Status,
		WorkspaceMode:     view.WorkspaceMode,
		IntegrationBranch: view.IntegrationBranch,
		CanonicalRemote:   view.CanonicalRemote,
		CurrentBranch:     view.CurrentBranch,
		Diagnostics:       diagnostics,
	}
}

func webCurrentWorkView(event tui.AgentEventView) webui.CurrentWorkView {
	return webui.CurrentWorkView{
		Phase:       event.Phase,
		Message:     event.Message,
		TicketKey:   event.TicketKey,
		ExecutionID: event.ExecutionID,
		LoopID:      event.LoopID,
	}
}

func webAgentEventViews(events []tui.AgentEventView, limit int) []webui.AgentEventView {
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	views := make([]webui.AgentEventView, 0, len(events))
	for _, event := range events {
		views = append(views, webui.AgentEventView{
			Time:        event.Time,
			Level:       event.Level,
			Phase:       event.Phase,
			TicketKey:   event.TicketKey,
			Message:     event.Message,
			LoopID:      event.LoopID,
			ExecutionID: event.ExecutionID,
		})
	}

	return views
}

func webLoopRunView(loop tui.LoopRunView) webui.LoopRunView {
	return webui.LoopRunView{
		ID:             loop.ID,
		Status:         loop.Status,
		UpdatedAt:      loop.UpdatedAt,
		EligibleCount:  loop.EligibleCount,
		ProcessedCount: loop.ProcessedCount,
		BlockedCount:   loop.BlockedCount,
		FailedCount:    loop.FailedCount,
		FailureReason:  loop.FailureReason,
	}
}

func webExecutionView(execution tui.ExecutionView) webui.ExecutionView {
	return webui.ExecutionView{
		ID:            execution.ID,
		TicketKey:     execution.TicketKey,
		Status:        execution.Status,
		UpdatedAt:     execution.UpdatedAt,
		FailureReason: execution.FailureReason,
	}
}
