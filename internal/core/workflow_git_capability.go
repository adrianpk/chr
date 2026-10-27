package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// WorkflowGitPublicationAdvertisement is a read-only observation of the actual
// receive-pack server. It is not a claim that an update was sent or applied.
type WorkflowGitPublicationAdvertisement struct {
	SourceCommit string
	TargetCommit string
	Fingerprint  string
}

// WorkflowGitPublicationInspector observes transport capability without updates.
type WorkflowGitPublicationInspector interface {
	InspectWorkflowGitPublication(context.Context, WorkflowGitPublication) (WorkflowGitPublicationAdvertisement, error)
}

// InspectWorkflowGitPublication reads the actual advertisement and terminates
// with an empty command list. Expected-old checks are repeated during publication.
func (git OSWorkflowGitPublisher) InspectWorkflowGitPublication(ctx context.Context, request WorkflowGitPublication) (observed WorkflowGitPublicationAdvertisement, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := git.validatePublication(ctx, request); err != nil {
		return observed, err
	}

	session, err := git.openReceivePack(ctx, request.Directory)
	if err != nil {
		return observed, err
	}
	defer func() { err = errors.Join(err, session.Close()) }()

	advertisement, err := readWorkflowAdvertisement(session.Advertisement())
	if err != nil {
		return observed, err
	}

	observed.SourceCommit, observed.TargetCommit = advertisement.refs[request.SourceRef], advertisement.refs[request.TargetRef]

	fingerprint, err := workflowGitCapabilityFingerprint(advertisement.capabilities)
	if err != nil {
		return observed, err
	}

	observed.Fingerprint = fingerprint
	checkErr := advertisement.check(request)
	finishErr := finishWorkflowGitInspection(session)

	return observed, errors.Join(checkErr, finishErr)
}

func workflowGitCapabilityFingerprint(capabilities map[string]bool) (string, error) {
	// A receive-pack session-id is a nonce, not a change in transaction policy.
	// Bind the features this protocol actually negotiates, independently of
	// optional display/progress features and per-connection diagnostics.
	format := "sha1"
	if capabilities["object-format=sha256"] {
		format = "sha256"
	}

	data, err := json.Marshal(struct {
		Atomic       bool
		ReportStatus bool
		ObjectFormat string
	}{capabilities["atomic"], capabilities["report-status"], format})
	if err != nil {
		return "", err
	}

	return contenthash.Sum(data), nil
}

func finishWorkflowGitInspection(session workflowReceivePack) error {
	reply, err := session.Send([]byte("0000"))
	if err != nil {
		return err
	}

	count, err := io.Copy(io.Discard, io.LimitReader(reply, workflowPublicationTextLimit+1))
	if err != nil || count > workflowPublicationTextLimit {
		return fmt.Errorf("incomplete or oversized inspection response")
	}

	return session.Finish()
}
