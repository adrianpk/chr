package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// ReportKind defines a package boundary type.
type ReportKind string

const (
	// ReportKindExecution defines a package boundary constant.
	ReportKindExecution ReportKind = "execution"
	// ReportKindLoop defines a package boundary constant.
	ReportKindLoop ReportKind = "loop"
)

// ReportListFilter describes a package boundary value.
type ReportListFilter struct {
	Kind        ReportKind
	TicketKey   string
	ExecutionID string
	LoopID      string
	Status      string
}

// ReportListEntry describes a package boundary value.
type ReportListEntry struct {
	Kind         ReportKind
	Path         string
	RelativePath string
	TicketKey    string
	ExecutionID  string
	LoopID       string
	Status       string
	ResultPath   string
}

// ReportLister describes a package boundary value.
type ReportLister struct{}

// NewReportLister constructs the corresponding package service or value.
func NewReportLister() ReportLister {
	return ReportLister{}
}

// ListProjectReports lists the requested data.
func (l ReportLister) ListProjectReports(
	ctx context.Context,
	projectDir string,
	filter ReportListFilter,
) ([]ReportListEntry, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return nil, err
	}

	root, err := layout.ArtifactRoot("run")
	if err != nil {
		return nil, err
	}

	entries, err := walkReportListEntries(ctx, projectDir, root, filter)
	if err != nil {
		if os.IsNotExist(err) {
			return []ReportListEntry{}, nil
		}

		return nil, fmt.Errorf("list project reports: %w", err)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].RelativePath < entries[j].RelativePath
	})

	return entries, nil
}

func walkReportListEntries(ctx context.Context, projectDir, root string, filter ReportListFilter) ([]ReportListEntry, error) {
	entries := []ReportListEntry{}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		lookup, err := reportListEntryFromPath(projectDir, path)
		if err != nil {
			return err
		}

		if !lookup.Found || !reportMatchesFilter(lookup.Entry, filter) {
			return nil
		}

		entries = append(entries, lookup.Entry)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

type reportListLookup struct {
	Entry ReportListEntry
	Found bool
}

func reportListEntryFromPath(projectDir, path string) (reportListLookup, error) {
	filename := filepath.Base(path)

	var kind ReportKind

	switch filename {
	case "execution-report.md":
		kind = ReportKindExecution
	case "loop-report.md":
		kind = ReportKindLoop
	default:
		return reportListLookup{}, nil
	}

	metadata, err := readReportListMetadata(path)
	if err != nil {
		return reportListLookup{}, err
	}

	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return reportListLookup{}, fmt.Errorf("relativize report path: %w", err)
	}

	report := ReportListEntry{
		Kind:         kind,
		Path:         path,
		RelativePath: filepath.ToSlash(relativePath),
		TicketKey:    metadata["ticket"],
		ExecutionID:  metadata["execution"],
		LoopID:       metadata["loop"],
		Status:       metadata["status"],
		ResultPath:   reportListAbsoluteMetadataPath(projectDir, metadata["result"]),
	}

	return reportListLookup{Entry: report, Found: true}, nil
}

func reportListAbsoluteMetadataPath(projectDir, path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "none" {
		return ""
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(projectDir, filepath.FromSlash(path))
}

func readReportListMetadata(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open report metadata: %w", err)
	}

	defer func() {
		//nolint:errcheck // Scanner errors are checked below.
		_ = file.Close()
	}()

	metadata := map[string]string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "- ") {
			continue
		}

		key, value, ok := strings.Cut(strings.TrimPrefix(line, "- "), ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if _, exists := metadata[key]; exists {
			continue
		}

		metadata[key] = strings.TrimSpace(value)
	}

	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("scan report metadata: %w", err)
	}

	return metadata, nil
}

func reportMatchesFilter(report ReportListEntry, filter ReportListFilter) bool {
	return reportKindMatchesFilter(report, filter) &&
		stringFieldMatchesFilter(report.TicketKey, filter.TicketKey) &&
		stringFieldMatchesFilter(report.ExecutionID, filter.ExecutionID) &&
		stringFieldMatchesFilter(report.LoopID, filter.LoopID) &&
		stringFieldMatchesFilter(report.Status, filter.Status)
}

func reportKindMatchesFilter(report ReportListEntry, filter ReportListFilter) bool {
	return filter.Kind == "" || report.Kind == filter.Kind
}

func stringFieldMatchesFilter(value, filter string) bool {
	return filter == "" || value == filter
}
