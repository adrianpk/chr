package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// The continuation builds dump images, never a nested publication. Its existing
// session/transcript/receipt journal owns the entire resulting change set.
func (r *fileAuthoringSessionRepository) prepareAuthoringDump(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, session AuthoringSession, body string) (AuthoringArtifactRef, []workflowFileChange, error) {
	observed, err := r.observeAuthoringDump(ctx, snapshot)
	if err != nil {
		return AuthoringArtifactRef{}, nil, err
	}

	if snapshot.dumpRevision != observed.Revision {
		return AuthoringArtifactRef{}, nil, ErrWorkflowConflict
	}

	target, err := r.authoringDumpTarget(session, observed)
	if err != nil {
		return AuthoringArtifactRef{}, nil, err
	}

	mutation, err := authoringDumpBinding(binding, snapshot, session, target, body)
	if err != nil {
		return AuthoringArtifactRef{}, nil, err
	}

	changes, _, err := r.dumpMutationChanges(mutation, target)
	if err != nil {
		return AuthoringArtifactRef{}, nil, err
	}

	// Earlier turns already retained this session's raw input. Never rewrite it.
	if authoringArtifactObservation(snapshot, target.RawReference) != nil {
		changes = withoutAuthoringRawChanges(changes, target.RawReference)
	}

	return authoringDumpRef(target.Destination), changes, nil
}

func (r *fileAuthoringSessionRepository) authoringDumpTarget(session AuthoringSession, observed ProjectDumpObservation) (dumpMutationTarget, error) {
	target := dumpMutationTarget{Project: observed.Project, Destination: observed.Destination}
	if len(session.Inputs) == 0 || session.Inputs[0].Kind != AuthoringArtifactKindDump || !validAuthoringArtifactRef(session.Namespace, session.Inputs[0]) {
		return target, ErrWorkflowConflict
	}

	if target.Destination == "" {
		allocated, err := r.newDumpMutationTarget(DefaultOpsNamespace, session.Topic)
		if err != nil {
			return target, err
		}

		target.Destination = allocated.Destination
	} else {
		target.Artifact = observed.Candidates[0]
	}

	target.RawReference = session.Inputs[0].Path

	return target, nil
}

func (r *fileAuthoringSessionRepository) prepareAuthoringDumpSnapshot(ctx context.Context, snapshot *AuthoringSnapshot, request AuthoringContinuationRequest) error {
	if request.Target != AuthoringArtifactKindDump || request.Action == "answer" {
		return nil
	}

	observed, err := r.observeAuthoringDump(ctx, *snapshot)
	if err != nil {
		return err
	}

	if request.Action == "refine" {
		output := authoringArtifactObservation(*snapshot, observed.Destination)
		if output == nil || output.Ref.Role != AuthoringArtifactRoleOutput {
			return ErrWorkflowConflict
		}
	}

	snapshot.dumpRevision = observed.Revision
	if observed.Destination == "" || authoringArtifactObservation(*snapshot, observed.Destination) != nil {
		return nil
	}

	_, err = r.readAuthoringDumpImage(snapshot, authoringDumpRef(observed.Destination), observed.Destination, nil)

	return err
}

func (r *fileAuthoringSessionRepository) observeAuthoringDump(ctx context.Context, snapshot AuthoringSnapshot) (ProjectDumpObservation, error) {
	observed, err := observeProjectDump(ctx, r.projectDir)
	if err != nil {
		return observed, err
	}

	// A question-only creation may have retained its own raw input before a
	// dump was generated. Only that exact observed input may bootstrap it.
	if observed.ReconciliationReason != "" && len(observed.Candidates) == 0 && observed.Project.ProjectDumpRef == "" && authoringOwnsRawComponents(snapshot, observed.Components) {
		observed.ReconciliationReason = ""
	}

	if observed.ReconciliationReason != "" {
		return observed, fmt.Errorf("%w: %s", ErrProjectDumpReconciliation, observed.ReconciliationReason)
	}

	return observed, nil
}

func authoringOwnsRawComponents(snapshot AuthoringSnapshot, components []ProjectDumpArtifact) bool {
	if len(components) == 0 {
		return false
	}

	for _, component := range components {
		observed := authoringArtifactObservation(snapshot, component.Reference)
		if observed == nil || observed.Ref.Role != AuthoringArtifactRoleInput || observed.SourceHash != component.SourceHash || observed.MetadataHash != component.MetadataHash {
			return false
		}
	}

	return true
}

func authoringDumpPreparation(snapshot AuthoringSnapshot, session AuthoringSession, request AuthoringContinuationRequest, target dumpMutationTarget, body string) (*DumpCapturePreparation, error) {
	if err := validateDumpPreparedScope(DumpMutationRequest{}, body); err != nil {
		return nil, err
	}

	contribution, err := dumpEditableBody([]byte(body))
	if err != nil {
		return nil, err
	}

	raw := renderAuthoringCapture(session, request.Capture, request.OriginalInstruction)

	if observed := authoringArtifactObservation(snapshot, target.RawReference); observed != nil {
		fields, err := projectDumpBodyFields([]byte(observed.Body))
		if err != nil || fields["type"] != "dump_raw" {
			return nil, ErrWorkflowConflict
		}

		raw = observed.Body
	} else {
		raw = strings.Replace(raw, "---\n", "---\ncurated: "+yamlString(target.Destination)+"\n", 1)
	}

	return &DumpCapturePreparation{
		Content:    "---\ntitle: " + yamlString(session.Topic) + "\ntype: dump\nstatus: curated\nsource: " + yamlString(target.RawReference) + "\n---\n\n" + string(contribution),
		RawContent: raw,
	}, nil
}

func withoutAuthoringRawChanges(changes []workflowFileChange, raw string) []workflowFileChange {
	retained := changes[:0]
	for _, change := range changes {
		if change.Path != filepath.FromSlash(raw) && change.Path != filepath.Join(filepath.Dir(raw), "metadata.json") {
			retained = append(retained, change)
		}
	}

	return retained
}

func authoringDumpRef(reference string) AuthoringArtifactRef {
	return AuthoringArtifactRef{ID: strings.Split(reference, "/")[3], Kind: AuthoringArtifactKindDump, Role: AuthoringArtifactRoleOutput, Path: reference, Description: "Shared project-wide dump."}
}

func authoringDumpBinding(binding authoringRequestBinding, snapshot AuthoringSnapshot, session AuthoringSession, target dumpMutationTarget, body string) (dumpMutationBinding, error) {
	request := DumpMutationRequest{OperationID: binding.Request.Binding.RequestID, Action: "capture", Title: session.Topic, Content: binding.Request.Capture}

	mutation := dumpMutationBinding{Request: request, Destination: target.Destination, RawReference: target.RawReference}
	if request.Content == "" {
		mutation.Request.Content = body
	}

	if binding.Request.Action == "refine" {
		if authoringArtifactObservation(snapshot, target.Destination) == nil {
			return mutation, ErrWorkflowConflict
		}

		mutation.Request.Action, mutation.Request.Content = "save", body

		return mutation, nil
	}

	prepared, err := authoringDumpPreparation(snapshot, session, binding.Request, target, body)
	mutation.Prepared = prepared

	return mutation, err
}
