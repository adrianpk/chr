package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// WorkflowGitPublication supplies immutable object identities and full refs to
// the trusted transport. It does not confer forge or receipt authority.
type WorkflowGitPublication struct {
	Repository string
	Directory  string
	SourceRef  string
	TargetRef  string
	Git        WorkflowIntegrationProof
}

// WorkflowGitPublicationResult distinguishes transport application from delivery.
// Unknown means commands may have reached the server and must not be retried.
type WorkflowGitPublicationResult struct {
	State        string `json:"state"`
	CommandsSent bool   `json:"commands_sent"`
	SourceCommit string `json:"source_commit"`
	TargetCommit string `json:"target_commit"`
	ResultCommit string `json:"result_commit"`
}

// WorkflowGitPublisher owns the advertisement and atomic two-ref exchange.
type WorkflowGitPublisher interface {
	PublishWorkflowGit(context.Context, WorkflowGitPublication) (WorkflowGitPublicationResult, error)
}

// OSWorkflowGitPublisher uses an explicit host-owned repository/transport mapping.
type OSWorkflowGitPublisher struct {
	repository string
	transport  string
}

// NewWorkflowGitPublisher stores trusted transport configuration without IO.
func NewWorkflowGitPublisher(repository, transport string) OSWorkflowGitPublisher {
	return OSWorkflowGitPublisher{repository: repository, transport: transport}
}

type workflowReceivePack interface {
	Advertisement() io.Reader
	Send([]byte) (io.Reader, error)
	Finish() error
	Close() error
}

// PublishWorkflowGit sends both old/new commands only after checking the actual
// receive-pack advertisement. No CLI push or non-atomic fallback is used.
func (git OSWorkflowGitPublisher) PublishWorkflowGit(ctx context.Context, request WorkflowGitPublication) (result WorkflowGitPublicationResult, err error) {
	result = WorkflowGitPublicationResult{State: "not_sent", ResultCommit: request.Git.ResultCommit}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := git.validatePublication(ctx, request); err != nil {
		return result, err
	}

	pack, err := workflowPublicationPack(ctx, request)
	if err != nil {
		return result, err
	}

	session, err := git.openReceivePack(ctx)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, session.Close()) }()

	advertisement, err := readWorkflowAdvertisement(session.Advertisement())
	if err != nil {
		return result, err
	}

	result.SourceCommit, result.TargetCommit = advertisement.refs[request.SourceRef], advertisement.refs[request.TargetRef]
	if err := advertisement.check(request); err != nil {
		return result, err
	}

	return exchangeWorkflowPublication(session, request, pack, result)
}

func (git OSWorkflowGitPublisher) validatePublication(ctx context.Context, request WorkflowGitPublication) error {
	repository, err := workflowCanonicalRepository(git.repository)
	if err != nil || repository != request.Repository || request.Directory == "" {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowPublicationRefs(request); err != nil {
		return err
	}

	reader := workflowIntegrationReader{dir: request.Directory}
	if _, err := reader.output(ctx, "merge-base", "--is-ancestor", request.Git.BaseCommit, request.Git.ResultCommit); err != nil {
		return fmt.Errorf("%w: prepared target is not fast-forward", ErrWorkflowConflict)
	}

	return nil
}

func validateWorkflowPublicationRefs(request WorkflowGitPublication) error {
	for _, ref := range []string{request.SourceRef, request.TargetRef} {
		if !strings.HasPrefix(ref, "refs/heads/") || !validWorkflowPRBranch(strings.TrimPrefix(ref, "refs/heads/")) {
			return ErrWorkflowConflict
		}
	}

	proof := request.Git
	if request.SourceRef == request.TargetRef || proof.SourceCommit == proof.ResultCommit || proof.BaseCommit == proof.ResultCommit {
		return ErrWorkflowObservationRequired
	}

	for _, id := range []string{proof.SourceCommit, proof.BaseCommit, proof.ResultCommit} {
		if !workflowCommitID(id) || len(id) != len(proof.SourceCommit) {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func exchangeWorkflowPublication(session workflowReceivePack, request WorkflowGitPublication, pack []byte, result WorkflowGitPublicationResult) (WorkflowGitPublicationResult, error) {
	commands := workflowPublicationCommands(request)
	commands = append(commands, pack...)
	result.CommandsSent, result.State = true, "unknown"

	report, err := session.Send(commands)
	if err != nil {
		return result, fmt.Errorf("%w: publication exchange failed", ErrWorkflowObservationRequired)
	}

	state, err := readWorkflowPublicationReport(report, request)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrWorkflowObservationRequired, err)
	}

	if err := session.Finish(); err != nil {
		return result, fmt.Errorf("%w: publication transport did not finish", ErrWorkflowObservationRequired)
	}

	result.State = state
	if state == "rejected" {
		return result, ErrWorkflowConflict
	}

	return result, nil
}

func workflowPublicationCommands(request WorkflowGitPublication) []byte {
	proof := request.Git

	capabilities := "report-status atomic"
	if len(proof.ResultCommit) == 64 {
		capabilities += " object-format=sha256"
	}

	first := fmt.Sprintf("%s %s %s\x00%s\n", proof.SourceCommit, proof.ResultCommit, request.SourceRef, capabilities)
	second := fmt.Sprintf("%s %s %s\n", proof.BaseCommit, proof.ResultCommit, request.TargetRef)
	commands := append(workflowPublicationPacket(first), workflowPublicationPacket(second)...)

	return append(commands, []byte("0000")...)
}
