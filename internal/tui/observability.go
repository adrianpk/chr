package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

// ObservabilityView summarizes execution, loop, and runtime observability.
type ObservabilityView struct {
	DaemonStatus             string
	DaemonPID                int
	DaemonUpdatedAt          string
	DaemonHeartbeat          string
	DaemonFailure            string
	LatestExecution          ExecutionView
	LatestLoop               LoopRunView
	RecentEvents             []AgentEventView
	RuntimeImage             string
	RuntimeDockerfile        string
	RuntimeArtifactDir       string
	RuntimeValidationCommand string
	RuntimeChecks            []RuntimeCheckView
}

// AgentEventView is the terminal projection of one agent-work event.
type AgentEventView struct {
	Time          string
	Level         string
	TicketKey     string
	Message       string
	LoopID        string
	ExecutionID   string
	RuntimeLog    string
	ValidationLog string
}

// ExecutionView is the terminal projection of one execution record.
type ExecutionView struct {
	ID             string
	TicketKey      string
	Status         string
	UpdatedAt      string
	FailureReason  string
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

// LoopRunView is the terminal projection of one agent loop run.
type LoopRunView struct {
	ID             string
	Status         string
	UpdatedAt      string
	EligibleCount  int
	ProcessedCount int
	BlockedCount   int
	FailedCount    int
	FailureReason  string
}

// RuntimeCheckView is the terminal projection of one runtime readiness check.
type RuntimeCheckView struct {
	Name    string
	Status  string
	Path    string
	Message string
}

// NewObservabilityView selects the latest observability records for the TUI.
func NewObservabilityView(
	projectDir string,
	daemon core.DaemonState,
	executions []core.ExecutionRecord,
	loops []core.LoopRunRecord,
	runtime core.RuntimeValidationResult,
	events []core.AgentEvent,
) ObservabilityView {
	return ObservabilityView{
		DaemonStatus:             string(daemon.Status),
		DaemonPID:                daemon.PID,
		DaemonUpdatedAt:          renderTime(daemon.UpdatedAt),
		DaemonHeartbeat:          renderOptionalTime(daemon.HeartbeatAt),
		DaemonFailure:            daemon.FailureReason,
		LatestExecution:          newExecutionView(projectDir, latestExecution(executions)),
		LatestLoop:               newLoopRunView(latestLoopRun(loops)),
		RecentEvents:             newAgentEventViews(projectDir, events),
		RuntimeImage:             runtime.Plan.Image,
		RuntimeDockerfile:        relativeRuntimePath(projectDir, runtime.Plan.DockerfilePath),
		RuntimeArtifactDir:       relativeRuntimePath(projectDir, runtime.Plan.ArtifactDir),
		RuntimeValidationCommand: runtime.Plan.ValidationCommand,
		RuntimeChecks:            newRuntimeCheckViews(projectDir, runtime.Checks),
	}
}

func newAgentEventViews(projectDir string, events []core.AgentEvent) []AgentEventView {
	views := make([]AgentEventView, 0, len(events))
	for _, event := range events {
		views = append(views, AgentEventView{
			Time:          renderAgentEventTime(event.Timestamp),
			Level:         string(event.Level),
			TicketKey:     event.TicketKey,
			Message:       event.Message,
			LoopID:        event.LoopID,
			ExecutionID:   event.ExecutionID,
			RuntimeLog:    relativeRuntimePath(projectDir, event.Refs["runtime_log_path"]),
			ValidationLog: relativeRuntimePath(projectDir, event.Refs["validation_log_path"]),
		})
	}

	return views
}

func renderAgentEventTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Local().Format("15:04:05")
}

func (m Model) renderAgent() string {
	var output strings.Builder

	output.WriteString("Agent\n")
	output.WriteString(fmt.Sprintf("  Daemon: %s\n", m.observability.DaemonStatus))
	output.WriteString(fmt.Sprintf("  PID: %s\n", renderPID(m.observability.DaemonPID)))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", m.observability.DaemonUpdatedAt))
	output.WriteString(fmt.Sprintf("  Heartbeat: %s\n", emptyAsUnknown(m.observability.DaemonHeartbeat)))
	output.WriteString(fmt.Sprintf("  Failure: %s\n\n", emptyAsNone(m.observability.DaemonFailure)))
	output.WriteString("Latest Loop\n")
	output.WriteString(renderLoopRunDetail(m.observability.LatestLoop))
	output.WriteString("\n\n")
	output.WriteString("Latest Execution\n")
	output.WriteString(renderExecutionDetail(m.observability.LatestExecution))

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) renderRuntime() string {
	var output strings.Builder

	output.WriteString("Runtime\n")
	output.WriteString(fmt.Sprintf("  Image: %s\n", m.observability.RuntimeImage))
	output.WriteString(fmt.Sprintf("  Dockerfile: %s\n", m.observability.RuntimeDockerfile))
	output.WriteString(fmt.Sprintf("  Artifact Dir: %s\n", m.observability.RuntimeArtifactDir))
	output.WriteString(fmt.Sprintf("  Validation: %s\n\n", m.observability.RuntimeValidationCommand))
	output.WriteString("Checks\n")

	if len(m.observability.RuntimeChecks) == 0 {
		output.WriteString("  No runtime checks were loaded.")

		return output.String()
	}

	for _, check := range m.observability.RuntimeChecks {
		output.WriteString(fmt.Sprintf("  %-12s %-4s %s", check.Name, check.Status, check.Path))

		if check.Message != "" && check.Message != "ok" {
			output.WriteString(fmt.Sprintf(" (%s)", check.Message))
		}

		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) renderWorkspaceIndexHealth() string {
	health := m.workspaceIndex.HealthWithSemanticDigestStates(m.semanticDigestStates)

	var output strings.Builder

	output.WriteString("Workspace Index\n")
	output.WriteString(fmt.Sprintf("  Nodes: %d\n", health.NodeCount))
	output.WriteString(fmt.Sprintf("  Edges: %d\n", health.EdgeCount))
	output.WriteString(fmt.Sprintf("  Diagnostics: %d\n", health.DiagnosticCount))
	output.WriteString(fmt.Sprintf("  Unresolved refs: %d\n", health.UnresolvedReferenceCount))
	output.WriteString(fmt.Sprintf("  Semantic pending: %d\n", health.SemanticDigestPendingCount))
	output.WriteString(fmt.Sprintf("  Semantic stale: %d", health.SemanticDigestStaleCount))

	return output.String()
}

func renderLoopRunDetail(loop LoopRunView) string {
	if loop.ID == "" {
		return "  No loop runs recorded."
	}

	var output strings.Builder

	output.WriteString(fmt.Sprintf("  ID: %s\n", loop.ID))
	output.WriteString(fmt.Sprintf("  Status: %s\n", loop.Status))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", loop.UpdatedAt))
	output.WriteString(fmt.Sprintf("  Counts: %d eligible / %d processed / %d blocked / %d failed\n",
		loop.EligibleCount,
		loop.ProcessedCount,
		loop.BlockedCount,
		loop.FailedCount,
	))
	output.WriteString(fmt.Sprintf("  Failure: %s", emptyAsNone(loop.FailureReason)))

	return output.String()
}

func renderExecutionDetail(execution ExecutionView) string {
	if execution.ID == "" {
		return "  No executions recorded."
	}

	var output strings.Builder

	output.WriteString(fmt.Sprintf("  ID: %s\n", execution.ID))
	output.WriteString(fmt.Sprintf("  Ticket: %s\n", execution.TicketKey))
	output.WriteString(fmt.Sprintf("  Status: %s\n", execution.Status))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", execution.UpdatedAt))
	output.WriteString(fmt.Sprintf("  Failure: %s\n", emptyAsNone(execution.FailureReason)))
	output.WriteString(fmt.Sprintf("  Run Dir: %s\n", execution.RunDir))
	output.WriteString(fmt.Sprintf("  Result: %s\n", execution.ResultPath))
	output.WriteString(fmt.Sprintf("  Log: %s\n", execution.LogPath))
	output.WriteString(fmt.Sprintf("  Validation: %s", execution.ValidationPath))

	return output.String()
}

func latestExecution(executions []core.ExecutionRecord) core.ExecutionRecord {
	var latest core.ExecutionRecord
	for _, execution := range executions {
		if latest.ID == "" || execution.UpdatedAt.After(latest.UpdatedAt) {
			latest = execution
		}
	}

	return latest
}

func latestLoopRun(loops []core.LoopRunRecord) core.LoopRunRecord {
	var latest core.LoopRunRecord
	for _, loop := range loops {
		if latest.ID == "" || loop.UpdatedAt.After(latest.UpdatedAt) {
			latest = loop
		}
	}

	return latest
}

func newExecutionView(projectDir string, execution core.ExecutionRecord) ExecutionView {
	if execution.ID == "" {
		return ExecutionView{}
	}

	return ExecutionView{
		ID:             execution.ID,
		TicketKey:      execution.TicketKey,
		Status:         string(execution.Status),
		UpdatedAt:      renderTime(execution.UpdatedAt),
		FailureReason:  execution.FailureReason,
		RunDir:         relativeRuntimePath(projectDir, execution.Artifacts.RunDir),
		ResultPath:     relativeRuntimePath(projectDir, execution.Artifacts.ResultPath),
		LogPath:        relativeRuntimePath(projectDir, execution.Artifacts.LogPath),
		ValidationPath: relativeRuntimePath(projectDir, execution.Artifacts.ValidationPath),
	}
}

func newLoopRunView(loop core.LoopRunRecord) LoopRunView {
	if loop.ID == "" {
		return LoopRunView{}
	}

	return LoopRunView{
		ID:             loop.ID,
		Status:         string(loop.Status),
		UpdatedAt:      renderTime(loop.UpdatedAt),
		EligibleCount:  loop.EligibleCount,
		ProcessedCount: loop.ProcessedCount,
		BlockedCount:   loop.BlockedCount,
		FailedCount:    loop.FailedCount,
		FailureReason:  loop.FailureReason,
	}
}

func newRuntimeCheckViews(projectDir string, checks []core.RuntimeValidationCheck) []RuntimeCheckView {
	views := make([]RuntimeCheckView, 0, len(checks))
	for _, check := range checks {
		views = append(views, RuntimeCheckView{
			Name:    check.Name,
			Status:  renderRuntimeCheckStatus(check.OK),
			Path:    relativeRuntimePath(projectDir, check.Path),
			Message: check.Message,
		})
	}

	return views
}

func renderRuntimeCheckStatus(ok bool) string {
	if ok {
		return "ok"
	}

	return "fail"
}

func renderPID(pid int) string {
	if pid <= 0 {
		return "none"
	}

	return fmt.Sprintf("%d", pid)
}

func renderOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return renderTime(*value)
}

func relativeRuntimePath(projectDir, path string) string {
	if path == "" {
		return ""
	}

	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relativePath)
}

func emptyAsNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}

	return value
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}

	return value
}
