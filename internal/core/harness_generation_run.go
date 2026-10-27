package core

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

// HarnessRunMetadata records host facts independently of semantic approval.
// ObservedModel stays absent unless a transport supplies verified observation.
type HarnessRunMetadata struct {
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
	Usage             *HarnessRunUsage              `json:"usage,omitempty"`
	UsageStatus       string                        `json:"usage_status"`
	StartedAt         time.Time                     `json:"started_at"`
	FinishedAt        *time.Time                    `json:"finished_at,omitempty"`
	DiagnosticsRef    string                        `json:"diagnostics_ref,omitempty"`
}

// HarnessRunUsage retains unknown counts as absent, never numeric zero.
type HarnessRunUsage struct {
	Input              *int64 `json:"input,omitempty"`
	Output             *int64 `json:"output,omitempty"`
	CacheRead          *int64 `json:"cache_read,omitempty"`
	CacheWrite         *int64 `json:"cache_write,omitempty"`
	AccountingRevision string `json:"accounting_revision"`
}

func startHarnessGenerationRun(ctx context.Context, request harness.AgentRunRequest) (*HarnessRunMetadata, error) {
	id, binding, bound := HarnessGenerationSelection(ctx)
	if !bound {
		return nil, nil
	}

	if !authoringstate.ValidOperationIdentity(request.ID) || !authoringstate.ValidOperationIdentity(request.Purpose) {
		return nil, errors.New("generation call requires bounded request and purpose identities")
	}

	record := &HarnessRunMetadata{
		Schema: "chr-harness-run/v1", RequestID: request.ID, Purpose: request.Purpose,
		OperationID: id, Selection: binding, SelectionRevision: binding.Selection.SelectionRevision,
		InputRef:       "sha256:" + contenthash.Sum([]byte(request.Prompt)),
		ProcessOutcome: "started", OutputValidation: "unknown", UsageStatus: "unknown", StartedAt: time.Now().UTC(),
	}
	if len(request.OutputSchema) != 0 {
		record.SchemaRef = "sha256:" + contenthash.Sum(request.OutputSchema)
	}

	if err := checkHarnessRunIdentity(request, *record); err != nil {
		return nil, err
	}

	return record, writeHarnessGenerationFile(harnessRunMetadataPath(request), record)
}

func checkHarnessRunIdentity(request harness.AgentRunRequest, record HarnessRunMetadata) error {
	data, err := readHarnessGenerationFile(harnessRunMetadataPath(request))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var retained HarnessRunMetadata
	if err := decodeHarnessGenerationFile(data, &retained); err != nil {
		return err
	}

	if !sameHarnessRunCall(retained, record) {
		return errors.New("generation call identity conflicts with its retained artifact")
	}

	return harnesscore.ValidateGenerationBinding(retained.Selection)
}

func sameHarnessRunCall(left, right HarnessRunMetadata) bool {
	return left.Schema == right.Schema && left.RequestID == right.RequestID && left.Purpose == right.Purpose && left.OperationID == right.OperationID &&
		left.Selection.BindingHash == right.Selection.BindingHash
}

func finishHarnessGenerationRun(request harness.AgentRunRequest, result harness.AgentRunResult, record *HarnessRunMetadata, runErr error) error {
	if record == nil {
		return runErr
	}

	finished := time.Now().UTC()
	record.FinishedAt = &finished
	record.ProcessOutcome = harnessRunProcessOutcome(runErr)
	record.Usage, record.UsageStatus = harnessRunMeasuredUsage(result.Usage)
	record.ObservedModel = result.ObservedModel

	if runErr == nil {
		runErr = retainNormalizedHarnessOutput(request, result.Output, record)
		if runErr != nil {
			record.ProcessOutcome = "protocol_error"
		}
	}

	if len(result.Stderr) != 0 {
		// The adapter owns bounded sanitized diagnostics in the existing bundle.
		record.DiagnosticsRef = "stderr.log"
	}

	return errors.Join(runErr, writeHarnessGenerationFile(harnessRunMetadataPath(request), record))
}

func retainNormalizedHarnessOutput(request harness.AgentRunRequest, output []byte, record *HarnessRunMetadata) error {
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

func harnessRunProcessOutcome(err error) string {
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

func harnessRunMeasuredUsage(usage harness.Usage) (*HarnessRunUsage, string) {
	if (usage.Input != nil && *usage.Input < 0) || (usage.Output != nil && *usage.Output < 0) || (usage.Input == nil && usage.Output == nil) {
		return nil, "unknown"
	}

	measured := &HarnessRunUsage{Input: usage.Input, Output: usage.Output, AccountingRevision: "chr-harness-observed/v1"}
	measured.CacheRead = optionalMeasuredCounter(usage.CacheRead)
	measured.CacheWrite = optionalMeasuredCounter(usage.CacheWrite)

	if usage.Input == nil || usage.Output == nil {
		return measured, "partial"
	}

	return measured, "known"
}

func optionalMeasuredCounter(value *int64) *int64 {
	if value != nil && *value >= 0 {
		return value
	}

	return nil
}

func harnessRunMetadataPath(request harness.AgentRunRequest) string {
	runDir := request.RunDir
	if runDir == "" {
		runDir = workspacecore.UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	return filepath.Join(runDir, "harness-run.json")
}

// ValidateHarnessGenerationOutput records the caller's parser result. Process
// success never implies schema validity or approval of an authoring artifact.
func ValidateHarnessGenerationOutput(ctx context.Context, request harness.AgentRunRequest, validationErr error) error {
	request.ID = HarnessGenerationRequestID(ctx, request.ID)

	id, binding, bound := HarnessGenerationSelection(ctx)
	if !bound {
		return validationErr
	}

	data, err := readHarnessGenerationFile(harnessRunMetadataPath(request))
	if err != nil {
		return errors.Join(validationErr, err)
	}

	var record HarnessRunMetadata
	if err := decodeHarnessGenerationFile(data, &record); err != nil {
		return errors.Join(validationErr, err)
	}

	if err := validateHarnessRunForOperation(record, request, id, binding); err != nil {
		return errors.Join(validationErr, err)
	}

	record.OutputValidation = "valid"
	if validationErr != nil {
		record.OutputValidation = "invalid"
	}

	return errors.Join(validationErr, writeHarnessGenerationFile(harnessRunMetadataPath(request), record))
}

func validateHarnessRunForOperation(record HarnessRunMetadata, request harness.AgentRunRequest, id string, binding harnesscore.GenerationBinding) error {
	if record.Schema != "chr-harness-run/v1" || record.RequestID != request.ID || record.Purpose != request.Purpose || record.OperationID != id || record.Selection.BindingHash != binding.BindingHash || record.ProcessOutcome != "completed" {
		return errors.New("invalid generation validation evidence")
	}

	return harnesscore.ValidateGenerationBinding(record.Selection)
}
