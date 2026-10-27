package host

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// RunMetadata records host facts independently of semantic approval.
// ObservedModel stays absent unless a transport supplies verified observation.
type RunMetadata struct {
	Schema            string                        `json:"schema"`
	RequestID         string                        `json:"request_id"`
	Purpose           string                        `json:"purpose"`
	OperationID       string                        `json:"operation_id"`
	Selection         harnesscore.GenerationBinding `json:"selection"`
	SelectionRevision string                        `json:"selection_revision"`
	ObservedModel     string                        `json:"observed_model,omitempty"`
	InputRef          string                        `json:"input_ref"`
	SchemaRef         string                        `json:"schema_ref,omitempty"`
	OutputRef         string                        `json:"output_ref,omitempty"`
	OutputHash        string                        `json:"output_hash,omitempty"`
	ProcessOutcome    string                        `json:"process_outcome"`
	OutputValidation  string                        `json:"output_validation"`
	Usage             *RunUsage                     `json:"usage,omitempty"`
	UsageStatus       string                        `json:"usage_status"`
	StartedAt         time.Time                     `json:"started_at"`
	FinishedAt        *time.Time                    `json:"finished_at,omitempty"`
	DiagnosticsRef    string                        `json:"diagnostics_ref,omitempty"`
}

// RunUsage retains unknown counts as absent, never numeric zero.
type RunUsage struct {
	Input              *int64 `json:"input,omitempty"`
	Output             *int64 `json:"output,omitempty"`
	CacheRead          *int64 `json:"cache_read,omitempty"`
	CacheWrite         *int64 `json:"cache_write,omitempty"`
	AccountingRevision string `json:"accounting_revision"`
}

func startGenerationRun(ctx context.Context, request harness.AgentRunRequest) (*RunMetadata, error) {
	generation, bound := GenerationSelection(ctx)
	if !bound {
		return nil, nil
	}

	if !authoringstate.ValidOperationIdentity(request.ID) || !authoringstate.ValidOperationIdentity(request.Purpose) {
		return nil, errors.New("generation call requires bounded request and purpose identities")
	}

	record := &RunMetadata{
		Schema: "chr-harness-run/v1", RequestID: request.ID, Purpose: request.Purpose,
		OperationID: generation.ID, Selection: generation.Binding, SelectionRevision: generation.Binding.Selection.SelectionRevision,
		InputRef:       "sha256:" + contenthash.Sum([]byte(request.Prompt)),
		ProcessOutcome: "started", OutputValidation: "unknown", UsageStatus: "unknown", StartedAt: time.Now().UTC(),
	}
	if len(request.OutputSchema) != 0 {
		record.SchemaRef = "sha256:" + contenthash.Sum(request.OutputSchema)
	}

	if err := checkRunIdentity(request, *record); err != nil {
		return nil, err
	}

	return record, writeGenerationFile(runMetadataPath(request), record)
}

func checkRunIdentity(request harness.AgentRunRequest, record RunMetadata) error {
	data, err := readGenerationFile(runMetadataPath(request))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var retained RunMetadata
	if err := decodeGenerationFile(data, &retained); err != nil {
		return err
	}

	if !sameRunCall(retained, record) {
		return errors.New("generation call identity conflicts with its retained artifact")
	}

	return harnesscore.ValidateGenerationBinding(retained.Selection)
}

func sameRunCall(left, right RunMetadata) bool {
	return left.Schema == right.Schema && left.RequestID == right.RequestID && left.Purpose == right.Purpose && left.OperationID == right.OperationID &&
		left.Selection.BindingHash == right.Selection.BindingHash
}

func finishGenerationRun(request harness.AgentRunRequest, result harness.AgentRunResult, record *RunMetadata, runErr error) error {
	if record == nil {
		return runErr
	}

	finished := time.Now().UTC()
	record.FinishedAt = &finished
	record.ProcessOutcome = runProcessOutcome(runErr)
	usage := runMeasuredUsage(result.Usage)
	record.Usage, record.UsageStatus = usage.Usage, usage.Status
	record.ObservedModel = result.ObservedModel

	if runErr == nil {
		runErr = retainNormalizedOutput(request, result.Output, record)
		if runErr != nil {
			record.ProcessOutcome = "protocol_error"
		}
	}

	if len(result.Stderr) != 0 {
		// The adapter owns bounded sanitized diagnostics in the existing bundle.
		record.DiagnosticsRef = "stderr.log"
	}

	return errors.Join(runErr, writeGenerationFile(runMetadataPath(request), record))
}

func retainNormalizedOutput(request harness.AgentRunRequest, output []byte, record *RunMetadata) error {
	limit := harness.MaxStructuredOutput
	if request.OutputLimit > 0 && request.OutputLimit < limit {
		limit = request.OutputLimit
	}

	if len(output) > limit {
		return process.ErrLimit
	}

	name := request.OutputName
	if name == "" {
		name = "output.md"
	}

	if err := opsfile.WriteAtomicBounded(filepath.Join(request.RunDir, name), output, limit); err != nil {
		return err
	}

	record.OutputRef, record.OutputHash = name, contenthash.Sum(output)

	return nil
}

func runProcessOutcome(err error) string {
	switch {
	case err == nil:
		return "completed"
	case errors.Is(err, context.DeadlineExceeded):
		return "timed_out"
	case errors.Is(err, harness.ErrAborted):
		return "cancelled"
	case errors.Is(err, harness.ErrProtocol), errors.Is(err, process.ErrLimit):
		return "protocol_error"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "failed"
	}
}

type measuredRunUsage struct {
	Usage  *RunUsage
	Status string
}

func runMeasuredUsage(usage harness.Usage) measuredRunUsage {
	if (usage.Input != nil && *usage.Input < 0) || (usage.Output != nil && *usage.Output < 0) || (usage.Input == nil && usage.Output == nil) {
		return measuredRunUsage{Status: "unknown"}
	}

	measured := &RunUsage{Input: usage.Input, Output: usage.Output, AccountingRevision: "chr-harness-observed/v1"}
	measured.CacheRead = optionalMeasuredCounter(usage.CacheRead)
	measured.CacheWrite = optionalMeasuredCounter(usage.CacheWrite)

	if usage.Input == nil || usage.Output == nil {
		return measuredRunUsage{Usage: measured, Status: "partial"}
	}

	return measuredRunUsage{Usage: measured, Status: "known"}
}

func optionalMeasuredCounter(value *int64) *int64 {
	if value != nil && *value >= 0 {
		return value
	}

	return nil
}

func runMetadataPath(request harness.AgentRunRequest) string {
	runDir := request.RunDir
	if runDir == "" {
		runDir = workspacecore.UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	return filepath.Join(runDir, "harness-run.json")
}

// ValidateGenerationOutput records the caller's parser result. Process
// success never implies schema validity or approval of an authoring artifact.
func ValidateGenerationOutput(ctx context.Context, request harness.AgentRunRequest, validationErr error) error {
	request.ID = GenerationRequestID(ctx, request.ID)

	generation, bound := GenerationSelection(ctx)
	if !bound {
		return validationErr
	}

	data, err := readGenerationFile(runMetadataPath(request))
	if err != nil {
		return errors.Join(validationErr, err)
	}

	var record RunMetadata
	if err := decodeGenerationFile(data, &record); err != nil {
		return errors.Join(validationErr, err)
	}

	if err := validateRunForOperation(record, request, generation.ID, generation.Binding); err != nil {
		return errors.Join(validationErr, err)
	}

	record.OutputValidation = "valid"
	if validationErr != nil {
		record.OutputValidation = "invalid"
	}

	return errors.Join(validationErr, writeGenerationFile(runMetadataPath(request), record))
}

func validateRunForOperation(record RunMetadata, request harness.AgentRunRequest, id string, binding harnesscore.GenerationBinding) error {
	if record.Schema != "chr-harness-run/v1" || record.RequestID != request.ID || record.Purpose != request.Purpose || record.OperationID != id || record.Selection.BindingHash != binding.BindingHash || record.ProcessOutcome != "completed" {
		return errors.New("invalid generation validation evidence")
	}

	return harnesscore.ValidateGenerationBinding(record.Selection)
}
