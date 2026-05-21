package chc

import "fmt"

const (
	TicketResultStatusReviewing = "reviewing"
	TicketResultStatusBlocked   = "blocked"
)

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

type ResultApplicationExecution struct {
	Status        ExecutionStatus
	FailureKind   ExecutionFailureKind
	FailureReason string
}

type ResultApplicationComment struct {
	Author string
	Body   string
}

type ResultApplicationEvidence struct {
	ValidationCommand string
	ResultPath        string
	LogPath           string
	ValidationPath    string
}

type ResultApplicationChangedFile struct {
	Path    string
	Summary string
}

type ResultApplicationBlockedReason struct {
	Kind           ExecutionFailureKind
	Reason         string
	LogPath        string
	ResultPath     string
	ValidationPath string
}

type ResultApplicationOptions struct {
	Execution     ExecutionRecord
	Result        ExecutionResult
	ChangedFiles  []ResultApplicationChangedFile
	CommentAuthor string
}

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
		TicketStatus: ticketStatusForResultApplication(opts.Execution, opts.Result),
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
	application.BlockedReason = resultApplicationBlockedReason(application)

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

func ticketStatusForResultApplication(record ExecutionRecord, result ExecutionResult) string {
	if record.Status == ExecutionStatusSucceeded && result.Status == ExecutionResultStatusSucceeded {
		return TicketResultStatusReviewing
	}

	return TicketResultStatusBlocked
}

func resultApplicationComment(author string, application ResultApplication) ResultApplicationComment {
	if author == "" {
		author = "chc"
	}

	return ResultApplicationComment{
		Author: author,
		Body:   RenderResultApplicationCommentBody(application),
	}
}

func RenderResultApplicationCommentBody(application ResultApplication) string {
	body := fmt.Sprintf(
		"Execution %s finished with status %s.\n\nSummary: %s\n\nResult: %s\nLog: %s\nValidation: %s",
		application.ExecutionID,
		application.Execution.Status,
		application.Summary,
		application.Evidence.ResultPath,
		application.Evidence.LogPath,
		application.Evidence.ValidationPath,
	)

	if application.Execution.FailureReason != "" {
		body += fmt.Sprintf("\n\nFailure: %s", application.Execution.FailureReason)
	}

	return body
}

func resultApplicationBlockedReason(application ResultApplication) ResultApplicationBlockedReason {
	if application.TicketStatus != TicketResultStatusBlocked {
		return ResultApplicationBlockedReason{}
	}

	return ResultApplicationBlockedReason{
		Kind:           application.Execution.FailureKind,
		Reason:         application.Execution.FailureReason,
		LogPath:        application.Evidence.LogPath,
		ResultPath:     application.Evidence.ResultPath,
		ValidationPath: application.Evidence.ValidationPath,
	}
}
