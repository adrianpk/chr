package core

import (
	"fmt"
	"strings"
)

const (
	// TicketResultStatusReviewing defines a package boundary constant.
	TicketResultStatusReviewing = "reviewing"
	// TicketResultStatusBlocked defines a package boundary constant.
	TicketResultStatusBlocked = "blocked"
)

// ResultApplication describes a package boundary value.
type ResultApplication struct {
	ExecutionID   string
	ProjectID     string
	TicketKey     string
	TicketPath    string
	TicketStatus  string
	Execution     ResultApplicationExecution
	Summary       string
	Comment       ResultApplicationComment
	Evidence      ResultApplicationEvidence
	ChangedFiles  []ResultApplicationChangedFile
	BlockedReason ResultApplicationBlockedReason
}

// ResultApplicationExecution describes a package boundary value.
type ResultApplicationExecution struct {
	Status        ExecutionStatus
	FailureKind   ExecutionFailureKind
	FailureReason string
}

// ResultApplicationComment describes a package boundary value.
type ResultApplicationComment struct {
	Author string
	Body   string
}

// ResultApplicationEvidence describes a package boundary value.
type ResultApplicationEvidence struct {
	ValidationCommand string
	ResultPath        string
	LogPath           string
	ValidationPath    string
}

// ResultApplicationChangedFile describes a package boundary value.
type ResultApplicationChangedFile struct {
	Path    string
	Summary string
}

// ResultApplicationBlockedReason describes a package boundary value.
type ResultApplicationBlockedReason struct {
	Kind           ExecutionFailureKind
	Reason         string
	LogPath        string
	ResultPath     string
	ValidationPath string
}

// ResultApplicationOptions describes a package boundary value.
type ResultApplicationOptions struct {
	Execution         ExecutionRecord
	Result            ExecutionResult
	EvidenceReadiness *EvidenceReadinessResult
	ChangedFiles      []ResultApplicationChangedFile
	CommentAuthor     string
}

// NewResultApplication constructs the corresponding package service or value.
func NewResultApplication(opts ResultApplicationOptions) (ResultApplication, error) {
	err := validateResultApplicationInputs(opts)
	if err != nil {
		return ResultApplication{}, err
	}

	application := ResultApplication{
		ExecutionID:  opts.Execution.ID,
		ProjectID:    opts.Execution.ProjectID,
		TicketKey:    opts.Execution.TicketKey,
		TicketPath:   opts.Execution.TicketPath,
		TicketStatus: ticketStatusForResultApplication(opts),
		Execution: ResultApplicationExecution{
			Status:        opts.Execution.Status,
			FailureKind:   ExecutionFailureKindFromReason(opts.Execution.FailureReason),
			FailureReason: opts.Execution.FailureReason,
		},
		Summary:      opts.Result.Summary,
		ChangedFiles: resultApplicationChangedFiles(opts),
		Evidence: ResultApplicationEvidence{
			ValidationCommand: opts.Execution.Runtime.ValidationCommand,
			ResultPath:        opts.Execution.Artifacts.ResultPath,
			LogPath:           opts.Execution.Artifacts.LogPath,
			ValidationPath:    opts.Execution.Artifacts.ValidationPath,
		},
	}

	application.Comment = resultApplicationComment(opts.CommentAuthor, application)
	application.BlockedReason = blockedResultReason(opts, application)

	return application, nil
}

func resultApplicationChangedFiles(opts ResultApplicationOptions) []ResultApplicationChangedFile {
	changedFiles := append([]ResultApplicationChangedFile(nil), opts.ChangedFiles...)

	for _, changedFile := range opts.Result.ChangedFiles {
		changedFiles = append(changedFiles, ResultApplicationChangedFile(changedFile))
	}

	return changedFiles
}

func validateResultApplicationInputs(opts ResultApplicationOptions) error {
	err := ValidateExecutionRecord(opts.Execution)
	if err != nil {
		return err
	}

	err = ValidateExecutionResult(opts.Result)
	if err != nil {
		return err
	}

	if opts.Execution.ID != opts.Result.ExecutionID {
		return fmt.Errorf("result application execution id mismatch: %s", opts.Result.ExecutionID)
	}

	if opts.Execution.TicketKey != opts.Result.TicketKey {
		return fmt.Errorf("result application ticket key mismatch: %s", opts.Result.TicketKey)
	}

	return nil
}

func ticketStatusForResultApplication(opts ResultApplicationOptions) string {
	readinessAllowsReview := opts.EvidenceReadiness == nil || opts.EvidenceReadiness.Allowed
	if opts.Execution.Status == ExecutionStatusSucceeded && opts.Result.Status == ExecutionResultStatusSucceeded && readinessAllowsReview {
		return TicketResultStatusReviewing
	}

	return TicketResultStatusBlocked
}

func resultApplicationComment(author string, application ResultApplication) ResultApplicationComment {
	if author == "" {
		author = "chr"
	}

	return ResultApplicationComment{
		Author: author,
		Body:   RenderResultApplicationCommentBody(application),
	}
}

// RenderResultApplicationCommentBody renders the corresponding package data.
func RenderResultApplicationCommentBody(application ResultApplication) string {
	var body strings.Builder
	body.WriteString("Execution `")
	body.WriteString(application.ExecutionID)
	body.WriteString("` finished with status `")
	body.WriteString(string(application.Execution.Status))
	body.WriteString("`.")

	summary := strings.TrimSpace(application.Summary)
	if summary != "" {
		body.WriteString("\n\nSummary: ")
		body.WriteString(summary)
	}

	if strings.TrimSpace(application.Execution.FailureReason) != "" {
		body.WriteString("\n\nExecution details are available in the runtime artifacts.")
	}

	return body.String()
}

func blockedResultReason(opts ResultApplicationOptions, application ResultApplication) ResultApplicationBlockedReason {
	if application.TicketStatus != TicketResultStatusBlocked {
		return ResultApplicationBlockedReason{}
	}

	evidenceBlocked := opts.Execution.Status == ExecutionStatusSucceeded &&
		opts.EvidenceReadiness != nil &&
		!opts.EvidenceReadiness.Allowed
	if evidenceBlocked {
		return ResultApplicationBlockedReason{
			Kind:           ExecutionFailureValidation,
			Reason:         opts.EvidenceReadiness.BlockerReason(),
			LogPath:        application.Evidence.LogPath,
			ResultPath:     application.Evidence.ResultPath,
			ValidationPath: application.Evidence.ValidationPath,
		}
	}

	return ResultApplicationBlockedReason{
		Kind:           application.Execution.FailureKind,
		Reason:         application.Execution.FailureReason,
		LogPath:        application.Evidence.LogPath,
		ResultPath:     application.Evidence.ResultPath,
		ValidationPath: application.Evidence.ValidationPath,
	}
}
