package workflow

import (
	"context"
	"encoding/json"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowSetAmendmentReceipt struct {
	Schema           string                               `json:"schema"`
	SetID            string                               `json:"set_id"`
	PreviousRevision uint64                               `json:"previous_revision"`
	PreviousHash     string                               `json:"previous_hash"`
	Plan             generationcore.GeneratedWorkflowPlan `json:"approved_plan"`
	ReopenedClosure  string                               `json:"reopened_closure,omitempty"`
	ClosureHash      string                               `json:"closure_hash,omitempty"`
	RecordedAt       time.Time                            `json:"recorded_at"`
}

func (s *Store) finishWorkflowAmendment(ctx context.Context, builder *workflowAmendmentBuild, request SetAmendmentRequest, body []byte) error {
	set := &builder.result.Set

	receipt := workflowSetAmendmentReceipt{Schema: "chr-set-amendment/v1", SetID: set.ID, PreviousRevision: set.Revision, PreviousHash: set.SourceHash, Plan: request.Plan, RecordedAt: time.Now().UTC()}
	if request.Reopen {
		receipt.ReopenedClosure = set.Set.ClosureRef

		var err error

		receipt.ClosureHash, err = s.workflowEvidenceHash(set.Set.ClosureRef)
		if err != nil {
			return err
		}

		set.Set.ClosureHistory = append(set.Set.ClosureHistory, set.Set.ClosureRef)
		set.Set.ClosureRef, set.Set.Status = "", "ready"
	}

	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	builder.result.ReceiptRef = builder.root + "/amendment.json"
	set.Set.Amendments = append(set.Set.Amendments, workflowstate.WorkflowAmendmentRef{Revision: set.Revision + 1, Ref: builder.result.ReceiptRef, Hash: contenthash.Sum(data)})
	set.Set.PlanRef, set.Set.PlanHash = builder.root+"/plan.md", request.Plan.PlanHash
	set.Set.ClosingValidation = request.Plan.ClosingValidation

	orderFiles, err := s.materializedOrderChanges(&builder.result, request.Plan)
	if err != nil {
		return err
	}

	builder.files = append(builder.files, orderFiles...)
	set.Revision++
	set.UpdatedAt = receipt.RecordedAt
	builder.files = append(builder.files, workflowstate.FileChange{Path: builder.result.ReceiptRef, After: data}, workflowstate.FileChange{Path: set.Set.PlanRef, After: body})

	return s.workflowAmendmentRecordFiles(ctx, builder)
}

func (s *Store) workflowAmendmentRecordFiles(ctx context.Context, builder *workflowAmendmentBuild) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	for _, task := range builder.result.Tasks {
		if task.SourceHash != "" {
			continue
		}

		snapshot.Records = append(snapshot.Records, task)

		files, err := s.encodeChanges(task)
		if err != nil {
			return err
		}

		builder.files = append(builder.files, files...)
	}

	if err := replaceWorkflowSnapshot(&snapshot, builder.result.Set); err != nil {
		return err
	}

	files, err := s.encodeChanges(builder.result.Set)
	if err != nil {
		return err
	}

	builder.result.Set.SourceHash = workflowClaimRecordHash(builder.result.Set, files)
	builder.files = append(builder.files, files...)

	return nil
}

func (s *Store) workflowAmendmentInvocationFiles(ctx context.Context, runs []workflowInvocationFile, old, next workflowstate.Record, files []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
	for _, run := range runs {
		if run.Namespace != old.Namespace || workflowstate.InvocationTerminal(run.Status) || run.OwnerRoot != s.layout.Root() {
			continue
		}

		changed := false

		for i, image := range run.SetImages {
			if image.Path == old.Path {
				run.SetImages[i].Revision, run.SetImages[i].SourceHash = next.Revision, next.SourceHash
				changed = true
			}
		}

		if !changed {
			continue
		}

		run.Revision++
		run.UpdatedAt = time.Now().UTC()

		change, err := workflowInvocationChange(run)
		if err != nil {
			return nil, err
		}

		files = append(files, change)
	}

	return files, ctx.Err()
}
