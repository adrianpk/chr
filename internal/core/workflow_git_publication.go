package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WorkflowGitPublication supplies immutable object identities and full refs to
// the trusted transport. It does not confer forge or receipt authority.
type WorkflowGitPublication struct {
	Repository              string
	SourceRepository        string
	Directory               string
	SourceRef               string
	TargetRef               string
	SourceRewriteAuthorized bool
	OwnershipRevision       string
	CandidateRevision       string
	Git                     WorkflowIntegrationProof
}

// WorkflowGitPublicationResult distinguishes transport application from delivery.
// Unknown means commands may have reached the server and must not be retried.
type WorkflowGitPublicationResult struct {
	Schema                 string `json:"schema"`
	State                  string `json:"state"`
	CommandsSent           bool   `json:"commands_sent"`
	OriginalSourceCommit   string `json:"original_source_commit"`
	AdvertisedSourceCommit string `json:"advertised_source_commit"`
	AdvertisedTargetCommit string `json:"advertised_target_commit"`
	PublishedSourceCommit  string `json:"published_source_commit,omitempty"`
	ResultCommit           string `json:"result_commit"`
	CandidateRevision      string `json:"candidate_revision"`
}

// WorkflowGitPublisher owns the advertisement and atomic two-ref exchange.
type WorkflowGitPublisher interface {
	PublishWorkflowGit(context.Context, WorkflowGitPublication) (WorkflowGitPublicationResult, error)
}

// OSWorkflowGitPublisher uses an explicit host-owned repository/transport mapping.
type OSWorkflowGitPublisher struct {
	repository string
	transport  string
	sshConfig  string
	httpClient *http.Client
}

// NewWorkflowGitPublisher stores trusted transport configuration without IO.
func NewWorkflowGitPublisher(repository, transport string) OSWorkflowGitPublisher {
	return OSWorkflowGitPublisher{repository: repository, transport: transport}
}

// WithSSHConfig selects a host-owned SSH configuration file; credentials remain
// in SSH's agent, keys and configuration rather than publication evidence.
func (git OSWorkflowGitPublisher) WithSSHConfig(path string) OSWorkflowGitPublisher {
	git.sshConfig = path

	return git
}

// WithHTTPClient selects host-owned TLS/proxy configuration. Publication always
// disables redirects, including redirects allowed by the supplied client.
func (git OSWorkflowGitPublisher) WithHTTPClient(client *http.Client) OSWorkflowGitPublisher {
	git.httpClient = client

	return git
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
	result = WorkflowGitPublicationResult{Schema: "chr-git-publication-result/v1", State: "not_sent", OriginalSourceCommit: request.Git.SourceCommit, ResultCommit: request.Git.ResultCommit, CandidateRevision: request.CandidateRevision}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := git.validatePublication(ctx, request); err != nil {
		return result, err
	}

	pack, err := workflowPublicationPack(ctx, request)
	if err != nil {
		return result, err
	}

	session, err := git.openReceivePack(ctx, request.Directory)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, session.Close()) }()

	advertisement, err := readWorkflowAdvertisement(session.Advertisement())
	if err != nil {
		return result, err
	}

	result.AdvertisedSourceCommit, result.AdvertisedTargetCommit = advertisement.refs[request.SourceRef], advertisement.refs[request.TargetRef]
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

	return git.validatePublicationProof(ctx, request)
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

	result.PublishedSourceCommit = request.Git.ResultCommit

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
