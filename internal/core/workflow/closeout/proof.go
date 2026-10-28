package closeout

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// DeliveryProof retains the decoded receipt proof and its verified evidence hash.
type DeliveryProof struct {
	Proof            workflowstate.SliceDeliveryProof
	VerificationHash string
}

// Slice binds one integrated descriptor to its actual members and fresh proof.
func (p Proofs) Slice(ctx context.Context, input workflowstate.ClosingInput, descriptor workflowstate.WorkflowSlice) (workflowstate.ClosingSlice, error) {
	if descriptor.Phase != "integrated" || descriptor.DeliveryRef == "" {
		return workflowstate.ClosingSlice{}, fmt.Errorf("%w: all actual slices must integrate", workflowstate.ErrConflict)
	}

	receipt, err := p.Source.Record(descriptor.DeliveryRef)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	observed, err := p.delivery(ctx, receipt)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	proof := observed.Proof

	if err := p.members(ctx, descriptor, receipt, proof); err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	fresh, err := p.observe(ctx, input, receipt, proof)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	refs := append([]string{descriptor.ContractRef, receipt.Path, receipt.Receipt.VerificationRef}, receipt.Receipt.ValidationRefs...)

	refs = append(refs, nativeEvidence(receipt, proof)...)
	if err := p.hashes(input.EvidenceHashes, refs); err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	return workflowstate.ClosingSlice{ID: descriptor.ID, Order: descriptor.Order, Receipt: workflowstate.ClosingEvidence{Ref: receipt.Path, Hash: receipt.SourceHash}, Verification: workflowstate.ClosingEvidence{Ref: receipt.Receipt.VerificationRef, Hash: observed.VerificationHash}, Proof: fresh}, nil
}

func (p Proofs) delivery(ctx context.Context, receipt workflowstate.Record) (DeliveryProof, error) {
	if receipt.Receipt == nil || receipt.ProjectID != p.ProjectID || receipt.Namespace != p.Namespace {
		return DeliveryProof{}, workflowstate.ErrConflict
	}

	if receipt.Receipt.Native != nil {
		return p.Source.NativeDelivery(ctx, receipt)
	}

	data, err := p.Source.Read(receipt.Receipt.VerificationRef)
	if err != nil {
		return DeliveryProof{}, err
	}

	var proof workflowstate.SliceDeliveryProof
	if err := json.Unmarshal(data, &proof); err != nil {
		return DeliveryProof{Proof: proof}, err
	}

	if proof.Schema != "chr-slice-delivery-proof/v1" || !workflowstate.ReceiptMatchesProof(receipt, proof, receipt.Receipt.VerificationRef) {
		return DeliveryProof{Proof: proof}, workflowstate.ErrConflict
	}

	return DeliveryProof{Proof: proof, VerificationHash: contenthash.Sum(data)}, nil
}

func (p Proofs) hashes(hashes map[string]string, refs []string) error {
	for _, ref := range refs {
		if len(hashes) >= workflowstate.RecordLimit {
			return fmt.Errorf("closing evidence observation bound exceeded")
		}

		hash, err := p.Source.Hash(ref)
		if err != nil || hash == "missing" {
			return proofError(err)
		}

		hashes[ref] = hash
	}

	return nil
}

func nativeEvidence(receipt workflowstate.Record, proof workflowstate.SliceDeliveryProof) []string {
	native := receipt.Receipt.Native
	if native == nil {
		return nil
	}

	refs := []string{proof.CandidateRef, native.CandidateRef.Path, native.TransactionProofRef.Path}
	if native.GitApplicationRef != nil {
		refs = append(refs, native.GitApplicationRef.Path)
	}

	for _, ref := range append(native.EvidenceRefs(), native.ObservationRefs...) {
		refs = append(refs, ref.Path)
	}

	return refs
}
