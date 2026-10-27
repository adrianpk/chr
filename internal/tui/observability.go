package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	daemoncore "forge.adrianpk.com/chr/internal/core/harness/daemon"
)

// ObservabilityView summarizes execution, loop, and runtime observability.
type ObservabilityView struct {
	Authority                core.WorkStateAuthority
	WorkflowPage             WorkflowPageView
	DaemonStatus             string
	DaemonPID                int
	DaemonUpdatedAt          string
	DaemonHeartbeat          string
	DaemonFailure            string
	LatestExecution          ExecutionView
	LatestLoop               LoopRunView
	LatestEvent              AgentEventView
	RecentEvents             []AgentEventView
	RuntimeImage             string
	RuntimeDockerfile        string
	RuntimeArtifactDir       string
	RuntimeValidationCommand string
	RuntimeChecks            []RuntimeCheckView
}

// AgentEventView is the terminal projection of one agent-work event.
type AgentEventView struct {
	Time            string
	Level           string
	Phase           string
	TicketKey       string
	Message         string
	LoopID          string
	ExecutionID     string
	TicketPath      string
	RuntimeLog      string
	ValidationLog   string
	ExecutionReport string
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
func NewObservabilityView(projectDir string, records ObservabilityRecords) ObservabilityView {
	return ObservabilityView{
		DaemonStatus:             string(records.Daemon.Status),
		DaemonPID:                records.Daemon.PID,
		DaemonUpdatedAt:          renderTime(records.Daemon.UpdatedAt),
		DaemonHeartbeat:          renderOptionalTime(records.Daemon.HeartbeatAt),
		DaemonFailure:            records.Daemon.FailureReason,
		LatestExecution:          newExecutionView(projectDir, latestExecution(records.Executions)),
		LatestLoop:               newLoopRunView(latestLoopRun(records.Loops)),
		LatestEvent:              latestAgentEventView(projectDir, records.Events),
		RecentEvents:             newAgentEventViews(projectDir, records.Events),
		RuntimeImage:             records.Runtime.Plan.Image,
		RuntimeDockerfile:        relativeRuntimePath(projectDir, records.Runtime.Plan.DockerfilePath),
		RuntimeArtifactDir:       relativeRuntimePath(projectDir, records.Runtime.Plan.ArtifactDir),
		RuntimeValidationCommand: records.Runtime.Plan.ValidationCommand,
		RuntimeChecks:            newRuntimeCheckViews(projectDir, records.Runtime.Checks),
	}
}

func newAgentEventViews(projectDir string, events []core.AgentEvent) []AgentEventView {
	views := make([]AgentEventView, 0, len(events))
	for _, event := range events {
		views = append(views, AgentEventView{
			Time:          renderAgentEventTime(event.Timestamp),
			Level:         string(event.Level),
			Phase:         string(event.Phase),
			TicketKey:     event.TicketKey,
			Message:       event.Message,
			LoopID:        event.LoopID,
			ExecutionID:   event.ExecutionID,
			TicketPath:    relativeRuntimePath(projectDir, event.Refs["ticket_path"]),
			RuntimeLog:    relativeRuntimePath(projectDir, event.Refs["runtime_log_path"]),
			ValidationLog: relativeRuntimePath(projectDir, event.Refs["validation_log_path"]),
			ExecutionReport: relativeRuntimePath(
				projectDir,
				event.Refs["execution_report_path"],
			),
		})
	}

	return views
}

func latestAgentEventView(projectDir string, events []core.AgentEvent) AgentEventView {
	var latest core.AgentEvent

	for _, event := range events {
		if latest.Timestamp.IsZero() || event.Timestamp.After(latest.Timestamp) {
			latest = event
		}
	}

	if latest.Timestamp.IsZero() {
		return AgentEventView{}
	}

	views := newAgentEventViews(projectDir, []core.AgentEvent{latest})
	if len(views) == 0 {
		return AgentEventView{}
	}

	return views[0]
}

func renderAgentEventTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Local().Format("15:04:05")
}

func (m *Model) renderAgent() string {
	var output strings.Builder

	output.WriteString("Agent\n")
	fmt.Fprintf(&output, "  Daemon: %s\n", m.observability.DaemonStatus)
	fmt.Fprintf(&output, "  PID: %s\n", renderPID(m.observability.DaemonPID))
	fmt.Fprintf(&output, "  Updated: %s\n", m.observability.DaemonUpdatedAt)
	fmt.Fprintf(&output, "  Heartbeat: %s\n", emptyAsUnknown(m.observability.DaemonHeartbeat))
	fmt.Fprintf(&output, "  Failure: %s\n\n", emptyAsNone(m.observability.DaemonFailure))
	output.WriteString("Latest Loop\n")
	output.WriteString(renderLoopRunDetail(m.observability.LatestLoop))
	output.WriteString("\n\n")
	output.WriteString("Latest Execution\n")
	output.WriteString(renderExecutionDetail(m.observability.LatestExecution))

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) renderRuntime() string {
	var output strings.Builder

	output.WriteString("Runtime\n")
	fmt.Fprintf(&output, "  Image: %s\n", m.observability.RuntimeImage)
	fmt.Fprintf(&output, "  Dockerfile: %s\n", m.observability.RuntimeDockerfile)
	fmt.Fprintf(&output, "  Artifact Dir: %s\n", m.observability.RuntimeArtifactDir)
	fmt.Fprintf(&output, "  Validation: %s\n\n", m.observability.RuntimeValidationCommand)
	output.WriteString("Checks\n")

	if len(m.observability.RuntimeChecks) == 0 {
		output.WriteString("  No runtime checks were loaded.")

		return output.String()
	}

	for _, check := range m.observability.RuntimeChecks {
		fmt.Fprintf(&output, "  %-12s %-4s %s", check.Name, check.Status, check.Path)

		if check.Message != "" && check.Message != "ok" {
			fmt.Fprintf(&output, " (%s)", check.Message)
		}

		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) renderWorkspaceIndexHealth() string {
	health := m.workspaceIndex.HealthWithSemanticDigestStates(m.semanticDigestStates)

	var output strings.Builder

	output.WriteString("Workspace Index\n")
	fmt.Fprintf(&output, "  Nodes: %d\n", health.NodeCount)
	fmt.Fprintf(&output, "  Edges: %d\n", health.EdgeCount)
	fmt.Fprintf(&output, "  Diagnostics: %d\n", health.DiagnosticCount)
	fmt.Fprintf(&output, "  Unresolved refs: %d\n", health.UnresolvedReferenceCount)
	fmt.Fprintf(&output, "  Semantic pending: %d\n", health.SemanticDigestPendingCount)
	fmt.Fprintf(&output, "  Semantic stale: %d", health.SemanticDigestStaleCount)

	return output.String()
}

func renderLoopRunDetail(loop LoopRunView) string {
	if loop.ID == "" {
		return "  No loop runs recorded."
	}

	var output strings.Builder

	fmt.Fprintf(&output, "  ID: %s\n", loop.ID)
	fmt.Fprintf(&output, "  Status: %s\n", loop.Status)
	fmt.Fprintf(&output, "  Updated: %s\n", loop.UpdatedAt)
	fmt.Fprintf(&output, "  Counts: %d eligible / %d processed / %d blocked / %d failed\n",
		loop.EligibleCount,
		loop.ProcessedCount,
		loop.BlockedCount,
		loop.FailedCount)
	fmt.Fprintf(&output, "  Failure: %s", emptyAsNone(loop.FailureReason))

	return output.String()
}

func renderExecutionDetail(execution ExecutionView) string {
	if execution.ID == "" {
		return "  No executions recorded."
	}

	var output strings.Builder

	fmt.Fprintf(&output, "  ID: %s\n", execution.ID)
	fmt.Fprintf(&output, "  Ticket: %s\n", execution.TicketKey)
	fmt.Fprintf(&output, "  Status: %s\n", execution.Status)
	fmt.Fprintf(&output, "  Updated: %s\n", execution.UpdatedAt)
	fmt.Fprintf(&output, "  Failure: %s\n", emptyAsNone(execution.FailureReason))
	fmt.Fprintf(&output, "  Run Dir: %s\n", execution.RunDir)
	fmt.Fprintf(&output, "  Result: %s\n", execution.ResultPath)
	fmt.Fprintf(&output, "  Log: %s\n", execution.LogPath)
	fmt.Fprintf(&output, "  Validation: %s", execution.ValidationPath)

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

// ObservabilityRecords contains the loaded daemon, execution, loop and runtime observations.
type ObservabilityRecords struct {
	Daemon     daemoncore.DaemonState
	Executions []core.ExecutionRecord
	Loops      []core.LoopRunRecord
	Runtime    core.RuntimeValidationResult
	Events     []core.AgentEvent
}
