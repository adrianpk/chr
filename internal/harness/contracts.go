// Package harness defines the technical boundary for owned agent calls.
package harness

import (
	"context"
	"time"
)

// Byte ceilings apply before accepting a call or a complete result.
const (
	MaxInput            = 4_194_304
	MaxRecord           = 1_048_576
	MaxStdout           = 67_108_864
	MaxStderr           = 2_097_152
	MaxDiagnostic       = 262_144
	MaxWorkerOutput     = 1_048_576
	MaxStructuredOutput = 65_536
)

// Capabilities describes supported operations, not remote authentication.
type Capabilities struct {
	Generation              bool
	Worker                  bool
	ReadOnlyFiles           bool
	CertifiedWholeCallUsage bool
}

// AgentRunRequest contains host-prepared input and caller-owned artifact paths.
type AgentRunRequest struct {
	ID                 string
	Purpose            string
	ProjectDir         string
	RunDir             string
	Prompt             string
	OutputSchema       []byte
	OutputName         string
	Program            string
	Model              string
	Timeout            time.Duration
	OutputLimit        int
	DeferResponseTrace bool
}

// Usage preserves unknown usage as nil; observed usage is not certification.
type Usage struct {
	Input  *int64
	Output *int64
}

// AgentRunResult contains normalized output and bounded sanitized diagnostics.
type AgentRunResult struct {
	ID          string
	Provider    string
	Model       string
	RunDir      string
	PromptPath  string
	OutputPath  string
	StdoutPath  string
	StderrPath  string
	CommandPath string
	Command     []string
	Output      []byte
	Stdout      []byte
	Stderr      []byte
	Usage       Usage
	StartedAt   time.Time
	Latency     time.Duration
}

// AgentRunner performs one independently owned generation call.
type AgentRunner interface {
	RunAgent(context.Context, AgentRunRequest) (AgentRunResult, error)
}

// AgentCommand is an executable invocation; stdin never appears in arguments.
type AgentCommand struct {
	Program string
	Args    []string
	Dir     string
}

// AgentCommandResult retains bounded diagnostics and a complete stdout fallback.
type AgentCommandResult struct {
	Stdout  []byte
	Stderr  []byte
	Latency time.Duration
}

// AgentCommandExecutor owns subprocess startup, cancellation and reaping.
type AgentCommandExecutor interface {
	RunAgentCommand(context.Context, AgentCommand, string, time.Time) (AgentCommandResult, error)
}

// EnvironmentVariable is an explicitly selected worker environment value.
type EnvironmentVariable struct {
	Name  string
	Value string
}

// WorkerRequest carries frozen technical input without workflow decisions.
type WorkerRequest struct {
	InputPath string
	Env       []EnvironmentVariable
}

// WorkerCommand describes an invocation in the host-selected execution lane.
type WorkerCommand struct {
	Program   string
	Args      []string
	Env       []EnvironmentVariable
	InputPath string
}

// Worker constructs commands; the host owns execution and result validation.
type Worker interface {
	WorkerCommand(WorkerRequest) (WorkerCommand, error)
	WorkflowWorker() []string
}
