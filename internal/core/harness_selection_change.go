package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ErrHarnessSelectionPreparation identifies a rejected candidate before any
// selection image or receipt is published; transport failures remain distinct.
var ErrHarnessSelectionPreparation = errors.New("harness selection preparation failed")

// HarnessSelectionAudit records the last accepted change without private resources.
type HarnessSelectionAudit struct {
	ProjectRef    string `json:"project_ref"`
	ChangedBy     string `json:"changed_by,omitempty"`
	ChangedAt     string `json:"changed_at,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	RequestDigest string `json:"request_digest,omitempty"`
}

// HarnessSelectionSnapshot binds both role references to one full configuration image.
type HarnessSelectionSnapshot struct {
	HarnessSelectionAudit
	Worker     string `json:"worker,omitempty"`
	Generation string `json:"generation,omitempty"`
	Revision   string `json:"revision"`
}

// HarnessSelectionChangeRequest omits unchanged roles; an explicit empty reference
// clears a role. Transport adapters translate their null choices to empty references.
type HarnessSelectionChangeRequest struct {
	Binding          workflowstate.MutationBinding `json:"binding"`
	ProjectRef       string                        `json:"project_ref"`
	ExpectedRevision string                        `json:"expected_revision"`
	Worker           *string                       `json:"worker,omitempty"`
	Generation       *string                       `json:"generation,omitempty"`
}

// HarnessSelectionChangeOptions supplies trusted authority and local preparation.
// Neither callback is serialized or supplied by interpreted model output.
type HarnessSelectionChangeOptions struct {
	Request   HarnessSelectionChangeRequest
	Authorize func(context.Context) error
	Validate  func(context.Context, WorkspaceConfig) error
}

// HarnessSelections projects one coherent workspace image and its last change.
func (c WorkspaceConfig) HarnessSelections(project string) (HarnessSelectionSnapshot, error) {
	if c.HarnessAudit.ProjectRef != "" && c.HarnessAudit.ProjectRef != project {
		return HarnessSelectionSnapshot{}, fmt.Errorf("%w: harness selections belong to another project", workflowstate.ErrConflict)
	}

	audit := c.HarnessAudit
	audit.ProjectRef = project

	return HarnessSelectionSnapshot{HarnessSelectionAudit: audit, Worker: c.Harnesses.Worker, Generation: c.Harnesses.Generation, Revision: c.Revision()}, nil
}

// ChangeHarnessSelections publishes an authorized complete image under the
// configuration owner's cross-process lock. Replays return the original snapshot.
func ChangeHarnessSelections(ctx context.Context, project string, opts HarnessSelectionChangeOptions) (result HarnessSelectionSnapshot, err error) {
	if opts.Authorize == nil || opts.Validate == nil {
		return result, errors.New("harness selection requires authority and preparation")
	}

	digest, err := harnessSelectionRequestDigest(opts.Request)
	if err != nil {
		return result, err
	}

	if err := opts.Authorize(ctx); err != nil {
		return result, err
	}

	err = workspacecore.WithWorkspaceConfigLock(ctx, project, func() error {
		if err := opts.Authorize(ctx); err != nil {
			return err
		}

		if err := requireNoWorkflowUpgrade(project); err != nil {
			return err
		}

		if err := recoverHarnessSelectionChange(project, opts.Request, digest); err != nil {
			return err
		}

		receipt, err := readHarnessSelectionReceipt(project, opts.Request, digest)
		if receipt.Found || err != nil {
			result = receipt.Snapshot

			return err
		}

		result, err = publishHarnessSelectionChange(ctx, project, opts, digest)

		return err
	})

	return result, err
}

func harnessSelectionRequestDigest(r HarnessSelectionChangeRequest) (string, error) {
	if r.Binding.Validate() != nil || !IsValidOperatorID(r.Binding.Operator) {
		return "", errors.New("harness selection requires an authorized actor and preparation")
	}

	if (workflowstate.MutationBinding{RequestID: r.ProjectRef, Operator: r.Binding.Operator}).Validate() != nil || !harnesscore.SelectionDigest(r.ExpectedRevision) {
		return "", errors.New("harness selection requires an exact project and configuration revision")
	}

	if err := validateHarnessSelectionChoices(r); err != nil {
		return "", err
	}

	data, err := json.Marshal(r)

	return contenthash.Sum(data), err
}

func validateHarnessSelectionChoices(r HarnessSelectionChangeRequest) error {
	if r.Worker == nil && r.Generation == nil {
		return errors.New("at least one explicit role change is required")
	}

	for _, value := range []*string{r.Worker, r.Generation} {
		if value != nil && *value != "" && !harnesscore.ValidProfileName(*value) {
			return errors.New("harness selection requires a named profile or clear")
		}
	}

	return nil
}

func publishHarnessSelectionChange(ctx context.Context, project string, opts HarnessSelectionChangeOptions, digest string) (HarnessSelectionSnapshot, error) {
	journal, err := prepareHarnessSelectionChange(ctx, project, opts, digest)
	if err != nil {
		return HarnessSelectionSnapshot{}, err
	}

	if err := checkHarnessSelectionReceiptCapacity(project); err != nil {
		return journal.Result, err
	}

	if err := ctx.Err(); err != nil {
		return journal.Result, err
	}

	if err := writeHarnessSelectionJournal(project, journal); err != nil {
		return journal.Result, err
	}

	return journal.Result, finishHarnessSelectionChange(project, journal)
}

func prepareHarnessSelectionChange(ctx context.Context, project string, opts HarnessSelectionChangeOptions, digest string) (harnessSelectionJournal, error) {
	var journal harnessSelectionJournal

	before, err := readHarnessSelectionConfig(project)
	if err != nil {
		return journal, err
	}

	if contenthash.Sum(before) != opts.Request.ExpectedRevision {
		return journal, workflowstate.ErrConflict
	}

	config, err := parseWorkflowConfig(before)
	if err != nil {
		return journal, err
	}

	if _, err := config.HarnessSelections(opts.Request.ProjectRef); err != nil {
		return journal, err
	}

	after := harnessSelectionChangeImage(before, opts.Request, digest, time.Now().UTC().Format(time.RFC3339Nano))

	candidate, err := parseWorkflowConfig(after)
	if err != nil {
		return journal, err
	}

	if err := opts.Validate(ctx, candidate); err != nil {
		return journal, fmt.Errorf("%w: %w", ErrHarnessSelectionPreparation, err)
	}

	result, err := candidate.HarnessSelections(opts.Request.ProjectRef)
	journal = harnessSelectionJournal{Schema: "chr-harness-selection-change/v1", Request: opts.Request, Result: result, Before: contenthash.Sum(before)}

	return journal, err
}

func harnessSelectionChangeImage(before []byte, request HarnessSelectionChangeRequest, digest, changedAt string) []byte {
	after := before

	for _, field := range []struct {
		key   string
		value *string
	}{{"worker", request.Worker}, {"generation", request.Generation}} {
		if field.value == nil {
			continue
		}

		value := *field.value
		switch value {
		case "":
			value = "null"
		case "null":
			value = strconv.Quote(value)
		}

		after = setWorkflowUpgradeValue(after, "harnesses", field.key, value)
	}

	for _, field := range []struct{ key, value string }{
		{"project_ref", request.ProjectRef},
		{"changed_by", request.Binding.Operator},
		{"changed_at", changedAt},
		{"request_id", request.Binding.RequestID},
		{"request_digest", digest},
	} {
		after = setWorkflowUpgradeValue(after, "harnesses", field.key, field.value)
	}

	return after
}
