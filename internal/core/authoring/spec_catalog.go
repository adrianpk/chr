package authoring

import "context"

// ListSpecDocuments enumerates actual canonical documents across namespaces,
// with at most 4096 inventory entries and 8 MiB of retained images per read.
func ListSpecDocuments(ctx context.Context, projectDir string) ([]SpecDocument, error) {
	return listAuthoringDocuments[SpecDocument](ctx, projectDir, "spec", readSpecCatalogDocument)
}

func readSpecCatalogDocument(ctx context.Context, projectDir, reference string) (authoringCatalogDocument[SpecDocument], error) {
	document, err := ReadSpecDocument(ctx, projectDir, reference)
	if err != nil {
		return authoringCatalogDocument[SpecDocument]{}, err
	}

	return authoringCatalogDocument[SpecDocument]{value: document, reference: document.Reference, size: len(document.images[0].After) + len(document.images[1].After)}, nil
}
