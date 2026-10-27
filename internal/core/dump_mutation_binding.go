package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const dumpMutationBindingLimit = 8 * workflowstate.WorkflowFileLimit

type dumpMutationBinding struct {
	Schema, Digest, Checksum  string
	Request                   DumpMutationRequest
	Destination, RawReference string
	Prepared                  *DumpCapturePreparation
	Attempts                  int
}

func (r *fileAuthoringSessionRepository) prepareDumpMutation(ctx context.Context, request DumpMutationRequest, digest string, prepare DumpCapturePreparer) (dumpMutationBinding, error) {
	binding, found, err := r.readDumpMutationBinding(request, digest)
	if err != nil {
		return binding, err
	}

	target, err := r.dumpMutationTarget(ctx, request)
	if err != nil {
		return binding, err
	}

	if found && (binding.Destination != target.Destination || binding.RawReference != target.RawReference) {
		return binding, workflowstate.ErrConflict
	}

	if !found {
		binding = dumpMutationBinding{Schema: "chr-project-dump-request/v1", Digest: digest, Request: request, Destination: target.Destination, RawReference: target.RawReference}
		if err := r.reserveDumpMutation(binding); err != nil {
			return binding, err
		}
	}

	if request.Action != "capture" || binding.Prepared != nil {
		return binding, nil
	}

	return r.prepareDumpCapture(ctx, binding, prepare)
}

func (r *fileAuthoringSessionRepository) readDumpMutationBinding(request DumpMutationRequest, digest string) (dumpMutationBinding, bool, error) {
	var binding dumpMutationBinding

	data, err := r.readBoundedAuthoringDocument(filepath.ToSlash(authoringRequestRef(request.OperationID)), dumpMutationBindingLimit)
	if errors.Is(err, os.ErrNotExist) {
		return binding, false, nil
	}

	if err != nil {
		return binding, false, err
	}

	if err := decodeAuthoringState(data, &binding); err != nil {
		return binding, false, err
	}

	return binding, true, r.validateDumpMutationBinding(binding, digest)
}

func (r *fileAuthoringSessionRepository) validateDumpMutationBinding(binding dumpMutationBinding, digest string) error {
	checksum := binding.Checksum
	binding.Checksum = ""

	encoded, err := json.Marshal(binding)
	if err != nil {
		return err
	}

	actualDigest, err := dumpMutationDigest(r.projectDir, binding.Request)
	if err != nil {
		return err
	}

	if binding.Schema != "chr-project-dump-request/v1" || digest != binding.Digest || digest != actualDigest || checksum != contenthash.Sum(encoded) || binding.Attempts < 0 || binding.Attempts > 5 {
		return workflowstate.ErrConflict
	}

	return nil
}

func (r *fileAuthoringSessionRepository) reserveDumpMutation(binding dumpMutationBinding) error {
	if err := r.projectDumpOperationAvailable(binding.Request.OperationID); err != nil {
		return err
	}

	if err := r.authoringRequestCapacity(); err != nil {
		return err
	}

	return r.saveDumpMutationBinding(binding)
}

func (r *fileAuthoringSessionRepository) saveDumpMutationBinding(binding dumpMutationBinding) error {
	binding.Checksum = ""

	encoded, err := json.Marshal(binding)
	if err != nil {
		return err
	}

	binding.Checksum = contenthash.Sum(encoded)

	encoded, err = json.Marshal(binding)
	if err != nil {
		return err
	}

	path, err := r.authoringPath(authoringRequestRef(binding.Request.OperationID))
	if err != nil {
		return err
	}

	return atomicBoundedWorkflowFile(path, encoded, dumpMutationBindingLimit)
}

func (r *fileAuthoringSessionRepository) prepareDumpCapture(ctx context.Context, binding dumpMutationBinding, prepare DumpCapturePreparer) (dumpMutationBinding, error) {
	if binding.Attempts >= 5 {
		return binding, ErrAuthoringProviderExhausted
	}

	binding.Attempts++
	if err := r.saveDumpMutationBinding(binding); err != nil {
		return binding, err
	}

	if prepare == nil {
		prepare = defaultDumpCapture(binding.Request)
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	prepared, err := prepare(ctx, binding.Destination, binding.RawReference)
	if err != nil {
		return binding, err
	}

	if err := validateDumpPreparation(binding, prepared); err != nil {
		return binding, err
	}

	binding.Prepared = &prepared

	return binding, r.saveDumpMutationBinding(binding)
}
