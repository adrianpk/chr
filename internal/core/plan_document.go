package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PlanDocument is one concrete plan body, independent of latest-in-scope
// generation lookup and internal tracker projections.
type PlanDocument struct {
	Reference, Revision, Title, Body string
	size                             int
}

// ReadPlanDocument observes the exact canonical plan without recovery or writes.
func ReadPlanDocument(ctx context.Context, projectDir, reference string) (PlanDocument, error) {
	if err := validatePlanDocumentReference(reference); err != nil {
		return PlanDocument{}, err
	}

	target, err := observePlanTarget(ctx, projectDir, reference)
	if err != nil {
		return PlanDocument{}, err
	}

	if target.images[0].After == nil {
		return PlanDocument{}, os.ErrNotExist
	}

	front, body, err := splitSpecDocument(target.images[0].After)
	if err != nil {
		return PlanDocument{}, err
	}

	fields, err := projectDocumentFields(front)
	if err != nil {
		return PlanDocument{}, err
	}

	title := fields["title"]
	if title == "" {
		value, ok := target.metadata["title"].(string)
		if !ok {
			return PlanDocument{}, errors.New("invalid plan title metadata")
		}

		title = value
	}

	if strings.TrimSpace(title) == "" {
		return PlanDocument{}, errors.New("plan title is required")
	}

	return PlanDocument{Reference: reference, Revision: target.revision, Title: title, Body: string(body), size: len(target.images[0].After) + len(target.images[1].After)}, nil
}

func validatePlanDocumentReference(reference string) error {
	if filepath.ToSlash(filepath.Clean(reference)) != reference || !filepath.IsLocal(reference) {
		return errors.New("invalid plan reference")
	}

	parts := strings.Split(reference, "/")
	if len(parts) != 5 || parts[0] != OpsDirectoryName || !IsValidOpsNamespace(parts[1]) || parts[2] != "plan" || !opsArtifactIDPattern.MatchString(parts[3]) || parts[4] != OpsArtifactBodyFile() {
		return errors.New("plan reference must identify a canonical plan artifact")
	}

	return nil
}
