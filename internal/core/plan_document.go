package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanDocument is one concrete plan body, independent of latest-in-scope
// generation lookup and internal tracker projections.
type PlanDocument struct {
	Reference, Revision, Title, Body string
	OriginReference                  string
	size                             int
}

// ReadPlanDocument observes the exact canonical plan without recovery or writes.
func ReadPlanDocument(ctx context.Context, projectDir, reference string) (PlanDocument, error) {
	if err := validatePlanDocumentReference(reference); err != nil {
		return PlanDocument{}, errors.Join(os.ErrNotExist, err)
	}

	target, err := observePlanTarget(ctx, projectDir, reference)
	if err != nil {
		return PlanDocument{}, err
	}

	if target.images[0].After == nil {
		return PlanDocument{}, os.ErrNotExist
	}

	parsedMarkdown, err := opsmetadata.SplitMarkdown(target.images[0].After)
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

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

	origin, err := planDocumentOrigin(target.metadata)
	if err != nil {
		return PlanDocument{}, err
	}

	return PlanDocument{Reference: reference, Revision: target.revision, Title: title, Body: string(body), OriginReference: origin, size: len(target.images[0].After) + len(target.images[1].After)}, nil
}

func planDocumentOrigin(metadata map[string]any) (string, error) {
	if metadata["workflow"] == nil {
		return "", nil
	}

	data, err := json.Marshal(metadata["workflow"])
	if err != nil {
		return "", err
	}

	var set workflowstate.WorkflowSet
	if err := json.Unmarshal(data, &set); err != nil {
		return "", err
	}

	if set.PlanRevision == nil {
		return "", nil
	}

	namespace, ok := metadata["namespace"].(string)
	if !ok || !opsnamespace.Valid(namespace) {
		return "", errors.New("invalid plan namespace")
	}

	return filepath.ToSlash(filepath.Join(opscore.DirectoryName, namespace, set.PlanRevision.Reference)), nil
}

func validatePlanDocumentReference(reference string) error {
	if filepath.ToSlash(filepath.Clean(reference)) != reference || !filepath.IsLocal(reference) {
		return errors.New("invalid plan reference")
	}

	parts := strings.Split(reference, "/")
	if len(parts) != 5 || parts[0] != opscore.DirectoryName || !opsnamespace.Valid(parts[1]) || parts[2] != "plan" || !opscore.HasArtifactNumberPrefix(parts[3]) || parts[4] != opscore.ArtifactBodyFile() {
		return errors.New("plan reference must identify a canonical plan artifact")
	}

	return nil
}
