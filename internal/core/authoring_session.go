package core

import (
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// ValidateAuthoringSession validates the corresponding package boundary value.
func ValidateAuthoringSession(session authoringcore.Session) error {
	if session.ID == "" {
		return fmt.Errorf("authoring session id is required")
	}

	if session.ProjectID == "" {
		return fmt.Errorf("authoring session project id is required")
	}

	if !session.Status.Valid() {
		return fmt.Errorf("authoring session status is invalid: %s", session.Status)
	}

	if !session.Mode.Valid() {
		return fmt.Errorf("authoring session mode is invalid: %s", session.Mode)
	}

	if !session.Target.Valid() {
		return fmt.Errorf("authoring session target is invalid: %s", session.Target)
	}

	if session.Topic == "" {
		return fmt.Errorf("authoring session topic is required")
	}

	err := validateAuthoringArtifactRefs("input", session.Inputs, authoringcore.ArtifactRoleInput)
	if err != nil {
		return err
	}

	err = validateAuthoringArtifactRefs("output", session.Outputs, authoringcore.ArtifactRoleOutput)
	if err != nil {
		return err
	}

	err = validateAuthoringProvenanceLinks(session.Provenance)
	if err != nil {
		return err
	}

	return validateAuthoringSessionTimes(session)
}

func validateAuthoringArtifactRefs(name string, refs []authoringcore.ArtifactRef, expectedRole authoringcore.ArtifactRole) error {
	for index, ref := range refs {
		err := validateAuthoringArtifactRef(ref)
		if err != nil {
			return fmt.Errorf("authoring %s artifact %d: %w", name, index+1, err)
		}

		if ref.Role != expectedRole {
			return fmt.Errorf("authoring %s artifact %d role must be %s", name, index+1, expectedRole)
		}
	}

	return nil
}

func validateAuthoringArtifactRef(ref authoringcore.ArtifactRef) error {
	if ref.ID == "" {
		return fmt.Errorf("artifact id is required")
	}

	if !ref.Kind.Valid() {
		return fmt.Errorf("artifact kind is invalid: %s", ref.Kind)
	}

	if !ref.Role.Valid() {
		return fmt.Errorf("artifact role is invalid: %s", ref.Role)
	}

	if ref.Path == "" {
		return fmt.Errorf("artifact path is required")
	}

	return nil
}

func validateAuthoringProvenanceLinks(links []authoringcore.ProvenanceLink) error {
	for index, link := range links {
		if !link.Relation.Valid() {
			return fmt.Errorf("authoring provenance link %d relation is invalid: %s", index+1, link.Relation)
		}

		err := validateAuthoringArtifactRef(link.Source)
		if err != nil {
			return fmt.Errorf("authoring provenance link %d source: %w", index+1, err)
		}

		err = validateAuthoringArtifactRef(link.Target)
		if err != nil {
			return fmt.Errorf("authoring provenance link %d target: %w", index+1, err)
		}

		if link.RecordedAt.IsZero() {
			return fmt.Errorf("authoring provenance link %d recorded time is required", index+1)
		}
	}

	return nil
}

func validateAuthoringSessionTimes(session authoringcore.Session) error {
	if err := validateManagedAuthoringRevision(session); err != nil {
		return err
	}

	if session.CreatedAt.IsZero() {
		return fmt.Errorf("authoring session created time is required")
	}

	if session.UpdatedAt.IsZero() {
		return fmt.Errorf("authoring session updated time is required")
	}

	if session.StartedAt != nil && session.FinishedAt != nil && session.FinishedAt.Before(*session.StartedAt) {
		return fmt.Errorf("authoring session finished before it started")
	}

	return nil
}

func validateManagedAuthoringRevision(session authoringcore.Session) error {
	if err := ValidateGuidedUseCase(session); err != nil {
		return err
	}

	if session.Revision < 0 || (session.Revision > 0 && !opsnamespace.Valid(session.Namespace)) {
		return fmt.Errorf("invalid managed authoring revision or namespace")
	}

	return nil
}
