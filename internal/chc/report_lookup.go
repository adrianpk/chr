package chc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type ReportLookupRequest struct {
	ReportID    string
	Kind        ReportKind
	TicketKey   string
	ExecutionID string
	LoopID      string
	Status      string
}

type ReportLookupResult struct {
	Entry ReportListEntry
	Body  string
}

type ReportLookupService struct {
	lister ReportLister
}

func NewReportLookupService() ReportLookupService {
	return ReportLookupService{
		lister: NewReportLister(),
	}
}

func (s ReportLookupService) LookupReport(
	ctx context.Context,
	projectDir string,
	request ReportLookupRequest,
) (ReportLookupResult, error) {
	filter := ReportListFilter{
		Kind:        request.Kind,
		TicketKey:   request.TicketKey,
		ExecutionID: request.ExecutionID,
		LoopID:      request.LoopID,
		Status:      request.Status,
	}

	reports, err := s.lister.ListProjectReports(ctx, projectDir, filter)
	if err != nil {
		return ReportLookupResult{}, err
	}

	matches := filterReportsByID(projectDir, reports, request.ReportID)
	if len(matches) == 0 {
		return ReportLookupResult{}, fmt.Errorf("report not found")
	}

	if len(matches) > 1 {
		return ReportLookupResult{}, fmt.Errorf("report lookup is ambiguous: %d reports match", len(matches))
	}

	body, err := os.ReadFile(matches[0].Path)
	if err != nil {
		return ReportLookupResult{}, fmt.Errorf("read report: %w", err)
	}

	return ReportLookupResult{
		Entry: matches[0],
		Body:  string(body),
	}, nil
}

func filterReportsByID(projectDir string, reports []ReportListEntry, reportID string) []ReportListEntry {
	if reportID == "" {
		return reports
	}

	matches := []ReportListEntry{}
	for _, report := range reports {
		if reportMatchesID(projectDir, report, reportID) {
			matches = append(matches, report)
		}
	}

	return matches
}

func reportMatchesID(projectDir string, report ReportListEntry, reportID string) bool {
	cleanID := filepath.ToSlash(filepath.Clean(reportID))
	cleanAbsoluteID := filepath.ToSlash(filepath.Clean(filepath.Join(projectDir, reportID)))

	switch cleanID {
	case report.RelativePath, report.TicketKey, report.ExecutionID, report.LoopID:
		return true
	}

	return cleanAbsoluteID == filepath.ToSlash(filepath.Clean(report.Path))
}
