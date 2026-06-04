package chc

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const defaultPreviousAttemptMaxValidationLines = 12

// PreviousAttemptContextOptions describes a package boundary value.
type PreviousAttemptContextOptions struct {
	Execution                  ExecutionRecord
	PreviousAttemptContextPath string
	MaxValidationLines         int
}

// PreviousAttemptContext describes a package boundary value.
type PreviousAttemptContext struct {
	ExecutionID                string
	TicketKey                  string
	Status                     ExecutionStatus
	FailureKind                ExecutionFailureKind
	FailureReason              string
	ValidationCommand          string
	Result                     *ExecutionResult
	ChangedFiles               []ExecutionResultChangedFile
	ValidationPath             string
	ValidationExcerpt          []string
	FailureSignatures          []string
	RepeatedFailureSignatures  []string
	PreviousAttemptContextPath string
	Sources                    []ContextSourceInventoryEntry
}

// LoadPreviousAttemptContext loads the corresponding package data.
func LoadPreviousAttemptContext(opts PreviousAttemptContextOptions) (PreviousAttemptContext, error) {
	err := ValidateExecutionRecord(opts.Execution)
	if err != nil {
		return PreviousAttemptContext{}, err
	}

	validationContent, validationSource, err := loadPreviousAttemptValidation(opts.Execution)
	if err != nil {
		return PreviousAttemptContext{}, err
	}

	result, resultSource, err := loadPreviousAttemptResult(opts.Execution)
	if err != nil {
		return PreviousAttemptContext{}, err
	}

	changedFiles := previousAttemptChangedFiles(result)
	signatures := validationFailureSignatures(validationContent)

	return PreviousAttemptContext{
		ExecutionID:                opts.Execution.ID,
		TicketKey:                  opts.Execution.TicketKey,
		Status:                     opts.Execution.Status,
		FailureKind:                ExecutionFailureKindFromReason(opts.Execution.FailureReason),
		FailureReason:              opts.Execution.FailureReason,
		ValidationCommand:          opts.Execution.Runtime.ValidationCommand,
		Result:                     result,
		ChangedFiles:               changedFiles,
		ValidationPath:             opts.Execution.Artifacts.ValidationPath,
		ValidationExcerpt:          validationExcerpt(validationContent, previousAttemptValidationLimit(opts.MaxValidationLines)),
		FailureSignatures:          signatures,
		RepeatedFailureSignatures:  repeatedValidationFailureSignatures(validationContent),
		PreviousAttemptContextPath: strings.TrimSpace(opts.PreviousAttemptContextPath),
		Sources:                    []ContextSourceInventoryEntry{validationSource, resultSource},
	}, nil
}

// RenderPreviousAttemptContext renders the corresponding package data.
func RenderPreviousAttemptContext(ctx PreviousAttemptContext) string {
	var output strings.Builder

	output.WriteString("# Previous Attempt Context\n\n")
	writeMarkdownField(&output, "Execution ID", ctx.ExecutionID)
	writeMarkdownField(&output, "Ticket Key", ctx.TicketKey)
	writeMarkdownField(&output, "Execution Status", string(ctx.Status))
	writeMarkdownField(&output, "Failure Kind", string(ctx.FailureKind))
	writeMarkdownField(&output, "Failure Reason", ctx.FailureReason)
	writeMarkdownField(&output, "Validation Command", ctx.ValidationCommand)
	writeMarkdownField(&output, "Validation Path", ctx.ValidationPath)
	writePreviousAttemptResult(&output, ctx.Result)
	writePreviousAttemptChangedFiles(&output, ctx.ChangedFiles)
	writeMarkdownList(&output, "Failure Signatures", ctx.FailureSignatures)
	writeMarkdownList(&output, "Repeated Failure Signatures", ctx.RepeatedFailureSignatures)

	output.WriteString("## Validation Excerpt\n\n")
	writeTextBlock(&output, ctx.ValidationExcerpt, "No validation output was captured.")

	output.WriteString("## Conversation History\n\n")
	output.WriteString("Implicit conversation history is not included. Only materialized execution artifacts are used for retry context.\n")

	return output.String()
}

// RenderPreviousAttemptContextLink renders the corresponding package data.
func RenderPreviousAttemptContextLink(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	var output strings.Builder

	output.WriteString("## Previous Attempt\n\n")
	output.WriteString("Read previous attempt evidence from `")
	output.WriteString(path)
	output.WriteString("` before retrying the ticket.\n")

	return output.String()
}

func loadPreviousAttemptValidation(record ExecutionRecord) (string, ContextSourceInventoryEntry, error) {
	source := ContextSourceInventoryEntry{
		ID:          "previous-attempt:validation",
		Kind:        ContextSourceKindPreviousAttempt,
		Class:       ContextSourceClassFresh,
		Path:        record.Artifacts.ValidationPath,
		Description: "Previous attempt validation artifact.",
		Required:    false,
	}

	content, err := os.ReadFile(record.Artifacts.ValidationPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			source.Reason = "file was not found"

			return "", source, nil
		}

		return "", source, fmt.Errorf("read previous validation artifact: %w", err)
	}

	source.Included = true

	return string(content), source, nil
}

func loadPreviousAttemptResult(record ExecutionRecord) (*ExecutionResult, ContextSourceInventoryEntry, error) {
	source := ContextSourceInventoryEntry{
		ID:          "previous-attempt:result",
		Kind:        ContextSourceKindPreviousAttempt,
		Class:       ContextSourceClassFresh,
		Path:        record.Artifacts.ResultPath,
		Description: "Previous attempt result artifact.",
		Required:    false,
	}

	result, err := LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			source.Reason = "file was not found"

			return nil, source, nil
		}

		return nil, source, fmt.Errorf("load previous result artifact: %w", err)
	}

	source.Included = true

	return &result, source, nil
}

func previousAttemptChangedFiles(result *ExecutionResult) []ExecutionResultChangedFile {
	if result == nil {
		return nil
	}

	return append([]ExecutionResultChangedFile(nil), result.ChangedFiles...)
}

func validationExcerpt(content string, limit int) []string {
	lines := nonEmptyLines(content)
	if len(lines) > limit {
		return lines[:limit]
	}

	return lines
}

func validationFailureSignatures(content string) []string {
	lines := nonEmptyLines(content)
	signatures := make([]string, 0, len(lines))
	seen := map[string]struct{}{}

	for _, line := range lines {
		if !looksLikeFailureLine(line) {
			continue
		}

		signature := normalizeFailureSignature(line)

		_, ok := seen[signature]
		if ok {
			continue
		}

		seen[signature] = struct{}{}
		signatures = append(signatures, signature)

		if len(signatures) >= defaultPreviousAttemptMaxValidationLines {
			break
		}
	}

	return signatures
}

func repeatedValidationFailureSignatures(content string) []string {
	counts := map[string]int{}

	for _, line := range nonEmptyLines(content) {
		if !looksLikeFailureLine(line) {
			continue
		}

		signature := normalizeFailureSignature(line)
		counts[signature]++
	}

	repeated := make([]string, 0, len(counts))
	for signature, count := range counts {
		if count > 1 {
			repeated = append(repeated, signature)
		}
	}

	return repeated
}

func nonEmptyLines(content string) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result = append(result, line)
	}

	return result
}

func looksLikeFailureLine(line string) bool {
	lower := strings.ToLower(line)

	for _, marker := range []string{"fail", "error", "panic", "undefined", "expected", "got", "missing"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}

func normalizeFailureSignature(line string) string {
	line = strings.TrimSpace(line)
	if len(line) > defaultContextSearchLineLength {
		line = strings.TrimSpace(line[:defaultContextSearchLineLength]) + "..."
	}

	return line
}

func previousAttemptValidationLimit(value int) int {
	if value > 0 {
		return value
	}

	return defaultPreviousAttemptMaxValidationLines
}

func writePreviousAttemptResult(output *strings.Builder, result *ExecutionResult) {
	output.WriteString("## Result\n\n")

	if result == nil {
		output.WriteString("No result artifact was captured.\n\n")

		return
	}

	output.WriteString("- status: ")
	output.WriteString(string(result.Status))
	output.WriteString("\n")
	output.WriteString("- summary: ")
	output.WriteString(result.Summary)
	output.WriteString("\n\n")
}

func writePreviousAttemptChangedFiles(output *strings.Builder, changedFiles []ExecutionResultChangedFile) {
	output.WriteString("## Changed Files\n\n")

	if len(changedFiles) == 0 {
		output.WriteString("No changed files were reported.\n\n")

		return
	}

	for _, changedFile := range changedFiles {
		output.WriteString("- `")
		output.WriteString(changedFile.Path)
		output.WriteString("`")

		if changedFile.Summary != "" {
			output.WriteString(": ")
			output.WriteString(changedFile.Summary)
		}

		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeTextBlock(output *strings.Builder, lines []string, emptyMessage string) {
	if len(lines) == 0 {
		output.WriteString(emptyMessage)
		output.WriteString("\n")

		return
	}

	output.WriteString("```text\n")

	for _, line := range lines {
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("```\n")
}
