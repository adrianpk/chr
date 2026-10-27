package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedPreparedPlan retains validated generation before revision-bound publication.
type GuidedPreparedPlan struct {
	Plan           GeneratedPlan `json:"plan"`
	TargetRevision string        `json:"target_revision"`
	InputDigest    string        `json:"input_digest"`
}

// GuidedWorkState links the canonical plan, materialization and native invocation owners.
type GuidedWorkState struct {
	Admission            *WorkflowStartRequest           `json:"admission,omitempty"`
	Invocation           *GuidedInvocationReference      `json:"invocation,omitempty"`
	PublicationOperation PlanPublicationOperation        `json:"publication_operation"`
	Publication          *PlanPublicationResult          `json:"publication,omitempty"`
	Materialization      *GuidedMaterializationReference `json:"materialization,omitempty"`
}

// GuidedWorkMember identifies one exact initial native task contract.
type GuidedWorkMember struct {
	Key, Path, ContractHash, SourceHash string
	Revision                            uint64
}

// GuidedMaterializationReference links the original receipt and exact native membership.
type GuidedMaterializationReference struct {
	ID, ReceiptRef, SetID, SetPath string
	PlanRevision                   PlanRevisionBinding
	Members                        []GuidedWorkMember
}

// PublishGuidedPlan composes the canonical publication owner using only retained run inputs.
func PublishGuidedPlan(ctx context.Context, root string, scope GuidedUseCaseScope) (result PlanPublicationResult, err error) {
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: root}

		session, readErr := r.getGuardedSession(ctx, scope.RunID)
		if readErr != nil {
			return readErr
		}

		if err := guardGuidedScope(session, scope); err != nil {
			return err
		}

		plan, err := guidedPublicationPlan(session)
		if err != nil {
			return err
		}

		previous, found, lookupErr := LookupPlanPublication(ctx, root, *plan.Operation)
		if lookupErr != nil || found {
			result = previous

			return lookupErr
		}

		if _, err := WriteGeneratedPlan(ctx, root, plan); err != nil {
			return err
		}

		result, found, err = LookupPlanPublication(ctx, root, *plan.Operation)
		if err == nil && !found {
			err = workflowstate.ErrConflict
		}

		return err
	})

	return result, err
}

func guidedPublicationPlan(session AuthoringSession) (GeneratedPlan, error) {
	state := session.Guided
	if err := guardGuidedPreparationState(state); err != nil {
		return GeneratedPlan{}, err
	}

	prepared := state.PendingOperation.PreparedPlan
	plan := prepared.Plan
	operation := state.Work.PublicationOperation

	if err := guardGuidedPlanScope(session); err != nil {
		return plan, err
	}

	checkpoint, err := guidedStageCheckpoint(session, "specification")
	if err != nil {
		return plan, err
	}

	digest, err := GuidedCheckpointDigest(checkpoint)
	if err != nil || digest != prepared.InputDigest {
		return plan, errors.Join(workflowstate.ErrConflict, err)
	}

	plan.Operation, plan.CreateOnly, plan.RequireOrderMap = &operation, true, true
	plan.ExpectedRevision, plan.SourceObservationHash = prepared.TargetRevision, digest

	plan.SourceInputs = guidedPlanSourceBindings(checkpoint)
	if err := ValidateGuidedPlannedOrder(plan.Order); err != nil {
		return plan, err
	}

	return plan, ValidateGeneratedPlan(plan)
}

// GuidedMaterializationBinding retains owner identities without copying native task state.
func GuidedMaterializationBinding(result WorkflowMaterializationResult, input GeneratedWorkflowPlan) (GuidedMaterializationReference, error) {
	if result.Set.Set == nil || result.Set.Set.PlanRevision == nil || result.Set.Set.UseCase != input.UseCase || len(result.Tasks) == 0 || len(result.Tasks) > workflowstate.PageLimit {
		return GuidedMaterializationReference{}, workflowstate.ErrConflict
	}

	ref := GuidedMaterializationReference{ID: input.MaterializationID, ReceiptRef: result.ReceiptRef, SetID: result.Set.ID, SetPath: result.Set.Path, PlanRevision: *result.Set.Set.PlanRevision}
	for _, task := range result.Tasks {
		if task.Task == nil || task.Task.SetID != ref.SetID {
			return ref, workflowstate.ErrConflict
		}

		ref.Members = append(ref.Members, GuidedWorkMember{Key: task.Task.Key, Path: task.Path, ContractHash: task.Task.ContractHash, SourceHash: task.SourceHash, Revision: task.Revision})
	}

	return ref, nil
}

// SealGuidedPreparation freezes the initial plan, bound tracker and native contracts.
func SealGuidedPreparation(ctx context.Context, root string, session *AuthoringSession) error {
	if session.Guided.Work == nil || session.Guided.Work.Materialization == nil || session.Guided.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	source, err := guidedStageCheckpoint(*session, "specification")
	if err != nil {
		return err
	}

	var paths []string
	if session.Guided.PendingCheckpoint == nil {
		paths, err = guidedPreparationPaths(*session)
		if err != nil {
			return err
		}
	}

	if err := sealGuidedStage(ctx, root, session, guidedStageSeal{Stage: "preparation", Paths: paths, Source: source}); err != nil {
		return err
	}

	session.Guided.Status, session.Guided.Blocker = "blocked", "Prepared work is sealed. Continue the retained run to admit and start it."

	return nil
}

func guidedPreparationPaths(session AuthoringSession) ([]string, error) {
	work := session.Guided.Work
	if work.Publication == nil {
		return nil, workflowstate.ErrConflict
	}

	materialized := work.Materialization
	binding := materialized.PlanRevision

	root := filepath.ToSlash(filepath.Join(OpsDirectoryName, session.Namespace))
	if validatePlanDocumentReference(work.Publication.Reference) != nil || !strings.HasPrefix(work.Publication.Reference, root+"/") {
		return nil, workflowstate.ErrConflict
	}

	references := []string{work.Publication.Reference[len(root)+1:], filepath.Join(filepath.Dir(work.Publication.Reference[len(root)+1:]), "metadata.json"), binding.BodyRef, binding.OrderRef, binding.TrackerRef, binding.CurrentTrackerRef, filepath.Join(filepath.Dir(binding.CurrentTrackerRef), "metadata.json"), materialized.ReceiptRef, materialized.SetPath, filepath.Join(filepath.Dir(materialized.SetPath), "artifact.md")}
	for _, file := range []string{"planned-order.json", "planned-tracker.md", "binding.json", "planned-binding.json", "projection.json"} {
		references = append(references, filepath.Join(filepath.Dir(binding.OrderRef), file))
	}

	trackerMeta, err := workflowTrackerMetadataPath(materialized.SetID)
	if err != nil {
		return nil, err
	}

	references = append(references, "tracker/"+materialized.SetID+".md", trackerMeta)
	for i, member := range materialized.Members {
		references = append(references, member.Path, filepath.Join(filepath.Dir(member.Path), "artifact.md"), filepath.Join("plan", materialized.SetID, "extension", "chr", fmt.Sprintf("task-%d.json", i+1)))
	}

	references = append(references, workSliceContracts(session)...)

	paths := make([]string, 0, len(references))
	for _, reference := range references {
		paths = append(paths, filepath.ToSlash(filepath.Join(root, reference)))
	}

	return paths, nil
}

func workSliceContracts(session AuthoringSession) []string {
	input := session.Guided.PendingOperation.PreparedWorkflow
	if input == nil {
		return nil
	}

	var refs []string
	for _, slice := range input.Slices {
		refs = append(refs, filepath.ToSlash(filepath.Join("plan", session.Guided.Work.Materialization.SetID, "extension", "chr", "slice-"+slice.ID+".json")))
	}

	return refs
}

// ValidateGuidedContracts rejects provider changes to owner-controlled identities and order.
func ValidateGuidedContracts(input, generated GeneratedWorkflowPlan) error {
	if err := ValidateGuidedPlannedOrder(plannedWorkflowOrder(generated)); err != nil {
		return err
	}

	if err := ValidateGeneratedWorkflowPlan(generated); err != nil {
		return err
	}

	if err := guardGuidedContractDependencies(generated); err != nil {
		return err
	}

	if !reflect.DeepEqual(plannedWorkflowOrder(input), plannedWorkflowOrder(generated)) {
		return workflowstate.ErrConflict
	}

	expected := input

	expected.Slices, expected.ClosingValidation = generated.Slices, generated.ClosingValidation
	if !reflect.DeepEqual(expected, generated) {
		return workflowstate.ErrConflict
	}

	encoded, err := json.Marshal(generated)
	if len(encoded) > 128<<10 {
		return errors.New("guided native contracts exceed 128 KiB")
	}

	return err
}

func guardGuidedPlanScope(session AuthoringSession) error {
	state := session.Guided
	operation := state.Work.PublicationOperation
	prepared := state.PendingOperation.PreparedPlan

	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: state.CaseSlug}
	if prepared.Plan.Scope != scope || operation.Scope != scope || operation.ExpectedRevision != prepared.TargetRevision || operation.Replace || !operation.TicketDelivery {
		return workflowstate.ErrConflict
	}

	return nil
}

func guidedPlanSourceBindings(checkpoint GuidedStageCheckpoint) []PlanSourceBinding {
	var inputs []PlanSourceBinding

	for _, source := range append(append([]GuidedCommittedArtifact{}, checkpoint.Sources...), checkpoint.Artifacts...) {
		commit := source.Commit
		if commit == "" {
			commit = checkpoint.Commit
		}

		inputs = append(inputs, PlanSourceBinding{Kind: "committed", Reference: source.Path, Revision: commit + ":" + source.Hash})
	}

	return inputs
}

// ValidateGuidedPlannedOrder keeps the complete execution selection within one bounded native page.
func ValidateGuidedPlannedOrder(order []PlanOrderSlice) error {
	if err := ValidatePlannedOrder(order); err != nil {
		return err
	}

	total := 0
	for _, slice := range order {
		total += len(slice.Tasks)
	}

	if total > workflowstate.PageLimit {
		return errors.New("guided work requires at most 128 tasks")
	}

	return nil
}

func guardGuidedContractDependencies(plan GeneratedWorkflowPlan) error {
	allowed := map[WorkflowRef]bool{}
	for _, slice := range plan.Slices {
		allowed[WorkflowRef{Kind: "slice", ID: slice.ID}] = true
		for _, task := range slice.Tasks {
			allowed[WorkflowRef{Kind: "ticket", ID: task.ID}] = true
		}
	}

	for _, slice := range plan.Slices {
		if err := guardGuidedDependencies(slice.DependsOn, allowed); err != nil {
			return err
		}

		for _, task := range slice.Tasks {
			if err := guardGuidedDependencies(task.DependsOn, allowed); err != nil {
				return err
			}
		}
	}

	return nil
}

func guardGuidedDependencies(refs []WorkflowRef, allowed map[WorkflowRef]bool) error {
	for _, ref := range refs {
		if !allowed[ref] {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func guardGuidedPreparationState(state *GuidedUseCaseState) error {
	if state == nil {
		return workflowstate.ErrConflict
	}

	if state.Stage != "preparation" || state.SpecApproval == nil || state.Work == nil || state.PendingOperation == nil || state.PendingOperation.PreparedPlan == nil {
		return workflowstate.ErrConflict
	}

	return nil
}
