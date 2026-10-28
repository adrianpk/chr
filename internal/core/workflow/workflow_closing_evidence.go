package workflow

import (
	"context"
	"fmt"
	"reflect"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClosingEvidence struct{ store *Store }

func (e workflowClosingEvidence) Read(relative string) ([]byte, error) {
	path, err := e.store.safePath(relative)
	if err != nil {
		return nil, err
	}

	return opsfile.Read(path)
}

func (s *Store) checkWorkflowClosingInputs(ctx context.Context, claim *workflowClosingClaim) error {
	if err := s.checkWorkflowClosingPublications(ctx); err != nil {
		return err
	}

	if err := s.checkWorkflowCoordinatorTarget(ctx, claim.input.Target); err != nil {
		return err
	}

	set, err := s.readRecord(claim.set.Path)
	if err != nil {
		return err
	}

	if set.Revision != claim.set.Revision || set.SourceHash != claim.set.SourceHash {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowClosingMembers(ctx, claim, set); err != nil {
		return err
	}

	refs := claim.input.EvidenceHashes
	if err := s.checkWorkflowPrerequisiteEvidence(refs); err != nil {
		return err
	}

	data, err := (workflowClosingEvidence{store: s}).Read(claim.state.Input.Ref)
	if err != nil || contenthash.Sum(data) != claim.state.Input.Hash {
		return errorsOrConflict(err)
	}

	return nil
}

func (s *Store) checkWorkflowClosingMembers(ctx context.Context, claim *workflowClosingClaim, set workflowstate.Record) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	members, err := workflowcloseout.Members(snapshot.Records, set.ID)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(members, claim.input.Members) {
		return fmt.Errorf("%w: actual closing members changed", workflowstate.ErrConflict)
	}

	return workflowClosingProofs(s, FrontierPorts{}).CheckReceipts(ctx, set)
}

func (s *Store) readWorkflowClosedSet(set workflowstate.Record) (SetClosureResult, error) {
	if set.Set.Closing == nil || len(set.Set.Closing.Attempts) == 0 {
		return SetClosureResult{}, workflowstate.ErrConflict
	}

	state, err := workflowcloseout.ReadAttempt(set.ID, set.Set.Closing, len(set.Set.Closing.Attempts)-1, workflowClosingEvidence{store: s})
	if err != nil || state.Outcome != "completed" || state.Result.Ref != set.Set.ClosureRef {
		return SetClosureResult{}, errorsOrConflict(err)
	}

	receipt, err := workflowcloseout.ReadReceipt(state, workflowClosingEvidence{store: s})
	if err != nil {
		return SetClosureResult{}, err
	}

	return SetClosureResult{Set: set, Receipt: receipt, ReceiptRef: state.Result.Ref, Outcome: state.Outcome}, nil
}

func (e workflowClosingEvidence) Check(hashes map[string]string) error {
	return e.store.checkWorkflowPrerequisiteEvidence(hashes)
}
