package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublishGuidedPlan composes the canonical publication owner using only retained run inputs.
func PublishGuidedPlan(ctx context.Context, root string, scope GuidedUseCaseScope) (result PlanPublicationResult, err error) {
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: root}

		session, readErr := r.getGuardedSession(ctx, scope.RunID)
		if readErr != nil {
			return readErr
		}

		if err := ValidateGuidedScope(session, scope); err != nil {
			return err
		}

		plan, err := guidedPublicationPlan(session)
		if err != nil {
			return err
		}

		previous, lookupErr := LookupPlanPublication(ctx, root, *plan.Operation)
		if lookupErr != nil || previous.Found {
			result = previous.Result

			return lookupErr
		}

		if _, err := WriteGeneratedPlan(ctx, root, plan); err != nil {
			return err
		}

		committed, lookupErr := LookupPlanPublication(ctx, root, *plan.Operation)

		result, err = committed.Result, lookupErr
		if err == nil && !committed.Found {
			err = workflowstate.ErrConflict
		}

		return err
	})

	return result, err
}

func guidedPublicationPlan(session Session) (generationcore.GeneratedPlan, error) {
	state := session.Guided
	if err := guardGuidedPreparationState(state); err != nil {
		return generationcore.GeneratedPlan{}, err
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

	return plan, generationcore.ValidateGeneratedPlan(plan)
}

// SealGuidedPreparation freezes the initial plan, bound tracker and native contracts.
func SealGuidedPreparation(ctx context.Context, root string, session *Session) error {
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

func guidedPreparationPaths(session Session) ([]string, error) {
	work := session.Guided.Work
	if work.Publication == nil {
		return nil, workflowstate.ErrConflict
	}

	materialized := work.Materialization
	binding := materialized.PlanRevision

	root := filepath.ToSlash(filepath.Join(opscore.DirectoryName, session.Namespace))
	if validatePlanDocumentReference(work.Publication.Reference) != nil || !strings.HasPrefix(work.Publication.Reference, root+"/") {
		return nil, workflowstate.ErrConflict
	}

	references := []string{work.Publication.Reference[len(root)+1:], filepath.Join(filepath.Dir(work.Publication.Reference[len(root)+1:]), "metadata.json"), binding.BodyRef, binding.OrderRef, binding.TrackerRef, binding.CurrentTrackerRef, filepath.Join(filepath.Dir(binding.CurrentTrackerRef), "metadata.json"), materialized.ReceiptRef, materialized.SetPath, filepath.Join(filepath.Dir(materialized.SetPath), "artifact.md")}
	for _, file := range []string{"planned-order.json", "planned-tracker.md", "binding.json", "planned-binding.json", "projection.json"} {
		references = append(references, filepath.Join(filepath.Dir(binding.OrderRef), file))
	}

	trackerMeta, err := workflowstate.TrackerMetadataPath(materialized.SetID)
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

func workSliceContracts(session Session) []string {
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
func ValidateGuidedContracts(input, generated generationcore.GeneratedWorkflowPlan) error {
	if err := ValidateGuidedPlannedOrder(generationcore.PlannedWorkflowOrder(generated)); err != nil {
		return err
	}

	if err := generationcore.ValidateGeneratedWorkflowPlan(generated); err != nil {
		return err
	}

	if err := guardGuidedContractDependencies(generated); err != nil {
		return err
	}

	if !reflect.DeepEqual(generationcore.PlannedWorkflowOrder(input), generationcore.PlannedWorkflowOrder(generated)) {
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

func guardGuidedPlanScope(session Session) error {
	state := session.Guided
	operation := state.Work.PublicationOperation
	prepared := state.PendingOperation.PreparedPlan

	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: state.CaseSlug}
	if prepared.Plan.Scope != scope || operation.Scope != scope || operation.ExpectedRevision != prepared.TargetRevision || operation.Replace || !operation.TicketDelivery {
		return workflowstate.ErrConflict
	}

	return nil
}

func guidedPlanSourceBindings(checkpoint GuidedStageCheckpoint) []generationcore.PlanSourceBinding {
	var inputs []generationcore.PlanSourceBinding

	for _, source := range append(append([]GuidedCommittedArtifact{}, checkpoint.Sources...), checkpoint.Artifacts...) {
		commit := source.Commit
		if commit == "" {
			commit = checkpoint.Commit
		}

		inputs = append(inputs, generationcore.PlanSourceBinding{Kind: "committed", Reference: source.Path, Revision: commit + ":" + source.Hash})
	}

	return inputs
}

// ValidateGuidedPlannedOrder keeps the complete execution selection within one bounded native page.
func ValidateGuidedPlannedOrder(order []generationcore.PlanOrderSlice) error {
	if err := generationcore.ValidatePlannedOrder(order); err != nil {
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

func guardGuidedContractDependencies(plan generationcore.GeneratedWorkflowPlan) error {
	allowed := map[workflowstate.Ref]bool{}
	for _, slice := range plan.Slices {
		allowed[workflowstate.Ref{Kind: "slice", ID: slice.ID}] = true
		for _, task := range slice.Tasks {
			allowed[workflowstate.Ref{Kind: "ticket", ID: task.ID}] = true
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

func guardGuidedDependencies(refs []workflowstate.Ref, allowed map[workflowstate.Ref]bool) error {
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
