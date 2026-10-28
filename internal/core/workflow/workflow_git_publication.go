package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GitPublication supplies immutable object identities and full refs to
// the trusted transport. It does not confer forge or receipt authority.
type GitPublication struct {
	Repository              string
	SourceRepository        string
	Directory               string
	SourceRef               string
	TargetRef               string
	SourceRewriteAuthorized bool
	OwnershipRevision       string
	CandidateRevision       string
	Git                     workflowstate.IntegrationProof
	// BeforeSend retains the actual advertisement and exact command intent.
	// A failed callback aborts before any update command is sent.
	BeforeSend func(context.Context, workflowstate.GitPublicationAdvertisement) error `json:"-"`
}

// GitPublisher owns the advertisement and atomic two-ref exchange.
type GitPublisher interface {
	PublishWorkflowGit(context.Context, GitPublication) (workflowstate.GitPublicationResult, error)
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
func (git OSWorkflowGitPublisher) PublishWorkflowGit(ctx context.Context, request GitPublication) (result workflowstate.GitPublicationResult, err error) {
	result = workflowstate.GitPublicationResult{Schema: "chr-git-publication-result/v1", State: workflowstate.GitPublicationNotSent, OriginalSourceCommit: request.Git.SourceCommit, ResultCommit: request.Git.ResultCommit, CandidateRevision: request.CandidateRevision}

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

	if err := retainWorkflowPublicationAdvertisement(ctx, request, advertisement, result); err != nil {
		return result, err
	}

	return exchangeWorkflowPublication(session, request, pack, result)
}

func retainWorkflowPublicationAdvertisement(ctx context.Context, request GitPublication, advertisement workflowAdvertisement, result workflowstate.GitPublicationResult) error {
	if request.BeforeSend == nil {
		return nil
	}

	fingerprint, err := workflowGitCapabilityFingerprint(advertisement.capabilities)
	if err != nil {
		return err
	}

	return request.BeforeSend(ctx, workflowstate.GitPublicationAdvertisement{SourceCommit: result.AdvertisedSourceCommit, TargetCommit: result.AdvertisedTargetCommit, Fingerprint: fingerprint})
}

func (git OSWorkflowGitPublisher) validatePublication(ctx context.Context, request GitPublication) error {
	repository, err := workflowstate.CanonicalRepository(git.repository)
	if err != nil || repository != request.Repository || request.Directory == "" {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowPublicationRefs(request); err != nil {
		return err
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: request.Directory}}
	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", request.Git.BaseCommit, request.Git.ResultCommit); err != nil {
		return fmt.Errorf("%w: prepared target is not fast-forward", workflowstate.ErrConflict)
	}

	return git.validatePublicationProof(ctx, request)
}

func validateWorkflowPublicationRefs(request GitPublication) error {
	for _, ref := range []string{request.SourceRef, request.TargetRef} {
		if !strings.HasPrefix(ref, "refs/heads/") || !workflowstate.ValidBranch(strings.TrimPrefix(ref, "refs/heads/")) {
			return workflowstate.ErrConflict
		}
	}

	proof := request.Git
	if request.SourceRef == request.TargetRef || proof.SourceCommit == proof.ResultCommit || proof.BaseCommit == proof.ResultCommit {
		return workflowstate.ErrObservationRequired
	}

	for _, id := range []string{proof.SourceCommit, proof.BaseCommit, proof.ResultCommit} {
		if !workflowstate.ValidCommitID(id) || len(id) != len(proof.SourceCommit) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func exchangeWorkflowPublication(session workflowReceivePack, request GitPublication, pack []byte, result workflowstate.GitPublicationResult) (workflowstate.GitPublicationResult, error) {
	commands := workflowPublicationCommands(request)
	commands = append(commands, pack...)
	result.CommandsSent, result.State = true, workflowstate.GitPublicationUnknown

	report, err := session.Send(commands)
	if err != nil {
		return result, fmt.Errorf("%w: publication exchange failed", workflowstate.ErrObservationRequired)
	}

	state, err := readWorkflowPublicationReport(report, request)
	if err != nil {
		return result, fmt.Errorf("%w: %w", workflowstate.ErrObservationRequired, err)
	}

	if err := session.Finish(); err != nil {
		return result, fmt.Errorf("%w: publication transport did not finish", workflowstate.ErrObservationRequired)
	}

	result.State = state
	if state == workflowstate.GitPublicationRejected {
		return result, workflowstate.ErrConflict
	}

	result.PublishedSourceCommit = request.Git.ResultCommit

	return result, nil
}

func workflowPublicationCommands(request GitPublication) []byte {
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
