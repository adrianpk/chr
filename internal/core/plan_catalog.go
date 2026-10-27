package core

import "context"

// ListPlanDocuments enumerates concrete canonical plans across namespaces,
// including retained archives, within 4096 inventory entries and 8 MiB per read.
func ListPlanDocuments(ctx context.Context, projectDir string) ([]PlanDocument, error) {
	return listAuthoringDocuments[PlanDocument](ctx, projectDir, "plan", readPlanCatalogDocument)
}

func readPlanCatalogDocument(ctx context.Context, projectDir, reference string) (authoringCatalogDocument[PlanDocument], error) {
	document, err := ReadPlanDocument(ctx, projectDir, reference)
	if err != nil {
		return authoringCatalogDocument[PlanDocument]{}, err
	}

	return authoringCatalogDocument[PlanDocument]{value: document, reference: document.Reference, size: document.size}, nil
}
