package chc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	defaultContextTreeMaxEntries   = 200
	defaultContextSearchMaxHits    = 40
	defaultContextSearchLineLength = 180
)

type TreeSearchContextOptions struct {
	ProjectDir        string
	Ticket            IndexedTicket
	ValidationCommand string
	MaxTreeEntries    int
	MaxSearchHits     int
}

type TreeSearchContext struct {
	ProjectDir string
	Tree       []ContextTreeEntry
	Terms      []string
	Hits       []ContextSearchHit
	Sources    []ContextSourceInventoryEntry
}

type ContextTreeEntry struct {
	Path  string
	IsDir bool
}

type ContextSearchHit struct {
	Term    string
	Path    string
	Line    int
	Excerpt string
}

func LoadTreeSearchContext(opts TreeSearchContextOptions) (TreeSearchContext, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return TreeSearchContext{}, err
	}

	tree, err := captureContextTree(projectDir, contextTreeLimit(opts.MaxTreeEntries))
	if err != nil {
		return TreeSearchContext{}, err
	}

	terms := DeriveContextSearchTerms(opts.Ticket, opts.ValidationCommand)

	hits, err := searchContextTerms(projectDir, terms, contextSearchLimit(opts.MaxSearchHits))
	if err != nil {
		return TreeSearchContext{}, err
	}

	return TreeSearchContext{
		ProjectDir: projectDir,
		Tree:       tree,
		Terms:      terms,
		Hits:       hits,
		Sources:    treeSearchSources(len(tree), len(hits)),
	}, nil
}

func DeriveContextSearchTerms(ticket IndexedTicket, validationCommand string) []string {
	seen := map[string]struct{}{}
	terms := make([]string, 0, 16)

	addSearchTokens(&terms, seen, ticket.Key)
	addSearchTokens(&terms, seen, ticket.Title)
	addSearchTokens(&terms, seen, ticket.UseCase)
	addSearchTokens(&terms, seen, ticket.ProjectArea)
	addSearchTokens(&terms, seen, validationCommand)

	for _, tag := range ticket.Tags {
		addSearchTokens(&terms, seen, tag)
	}

	if len(terms) > 12 {
		return terms[:12]
	}

	return terms
}

func RenderCodeContext(ctx TreeSearchContext) string {
	var output strings.Builder

	output.WriteString("# Code Context\n\n")
	writeMarkdownField(&output, "Project Directory", ctx.ProjectDir)
	writeMarkdownList(&output, "Search Terms", ctx.Terms)

	output.WriteString("## File Tree\n\n")
	writeContextTree(&output, ctx.Tree)

	output.WriteString("## Text Search Hits\n\n")
	writeContextSearchHits(&output, ctx.Hits)

	return output.String()
}

func captureContextTree(projectDir string, maxEntries int) ([]ContextTreeEntry, error) {
	entries := make([]ContextTreeEntry, 0, maxEntries)

	err := filepath.WalkDir(projectDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == projectDir {
			return nil
		}

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return fmt.Errorf("resolve tree path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		if shouldSkipContextPath(relPath, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		entries = append(entries, ContextTreeEntry{
			Path:  relPath,
			IsDir: entry.IsDir(),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("capture context tree: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	if len(entries) > maxEntries {
		return entries[:maxEntries], nil
	}

	return entries, nil
}

func searchContextTerms(projectDir string, terms []string, maxHits int) ([]ContextSearchHit, error) {
	files, err := searchableContextFiles(projectDir)
	if err != nil {
		return nil, err
	}

	hits := make([]ContextSearchHit, 0, maxHits)
	for _, path := range files {
		fileHits, err := searchContextFile(projectDir, path, terms, maxHits-len(hits))
		if err != nil {
			return nil, err
		}

		hits = append(hits, fileHits...)
		if len(hits) >= maxHits {
			return hits, nil
		}
	}

	return hits, nil
}

func searchableContextFiles(projectDir string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(projectDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == projectDir {
			return nil
		}

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return fmt.Errorf("resolve searchable path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		if shouldSkipContextPath(relPath, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if entry.IsDir() || !isSearchableContextFile(relPath) {
			return nil
		}

		files = append(files, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect searchable context files: %w", err)
	}

	sort.Strings(files)

	return files, nil
}

func searchContextFile(projectDir, path string, terms []string, remaining int) ([]ContextSearchHit, error) {
	if remaining <= 0 || len(terms) == 0 {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open searchable context file: %w", err)
	}

	defer func() {
		//nolint:errcheck // Search reads only; scanner errors are checked below.
		_ = file.Close()
	}()

	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return nil, fmt.Errorf("resolve search hit path: %w", err)
	}

	hits := make([]ContextSearchHit, 0, remaining)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		matchedTerm := firstMatchingContextTerm(line, terms)
		if matchedTerm == "" {
			continue
		}

		hits = append(hits, ContextSearchHit{
			Term:    matchedTerm,
			Path:    filepath.ToSlash(relPath),
			Line:    lineNumber,
			Excerpt: truncateSearchExcerpt(strings.TrimSpace(line)),
		})
		if len(hits) >= remaining {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan searchable context file: %w", err)
	}

	return hits, nil
}

func addSearchTokens(terms *[]string, seen map[string]struct{}, value string) {
	for _, token := range strings.FieldsFunc(value, searchTokenSeparator) {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if len(normalized) < 3 {
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		*terms = append(*terms, normalized)
	}
}

func searchTokenSeparator(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

func firstMatchingContextTerm(line string, terms []string) string {
	normalizedLine := strings.ToLower(line)

	for _, term := range terms {
		if strings.Contains(normalizedLine, term) {
			return term
		}
	}

	return ""
}

func shouldSkipContextPath(path string, isDir bool) bool {
	firstPart := path
	if before, _, ok := strings.Cut(path, "/"); ok {
		firstPart = before
	}

	switch firstPart {
	case ".git", "node_modules", "vendor", "tmp", "_workspace":
		return true
	}

	if isDir && path == DirectoryName+"/runtime/runs" {
		return true
	}

	return strings.HasPrefix(path, DirectoryName+"/runtime/runs/")
}

func isSearchableContextFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".js", ".jsx", ".ts", ".tsx", ".md", ".markdown", ".txt", ".yml", ".yaml", ".json", ".toml", ".sh", ".sql":
		return true
	default:
		return false
	}
}

func truncateSearchExcerpt(value string) string {
	if len(value) <= defaultContextSearchLineLength {
		return value
	}

	return strings.TrimSpace(value[:defaultContextSearchLineLength]) + "..."
}

func contextTreeLimit(value int) int {
	if value > 0 {
		return value
	}

	return defaultContextTreeMaxEntries
}

func contextSearchLimit(value int) int {
	if value > 0 {
		return value
	}

	return defaultContextSearchMaxHits
}

func writeContextTree(output *strings.Builder, entries []ContextTreeEntry) {
	if len(entries) == 0 {
		output.WriteString("No file tree entries were captured.\n\n")

		return
	}

	output.WriteString("```text\n")

	for _, entry := range entries {
		output.WriteString(entry.Path)

		if entry.IsDir {
			output.WriteString("/")
		}

		output.WriteString("\n")
	}

	output.WriteString("```\n\n")
}

func writeContextSearchHits(output *strings.Builder, hits []ContextSearchHit) {
	if len(hits) == 0 {
		output.WriteString("No text search hits were captured.\n")

		return
	}

	for _, hit := range hits {
		output.WriteString("- `")
		output.WriteString(hit.Path)
		output.WriteString(":")
		output.WriteString(strconv.Itoa(hit.Line))
		output.WriteString("` ")
		output.WriteString("`")
		output.WriteString(hit.Term)
		output.WriteString("` ")
		output.WriteString(hit.Excerpt)
		output.WriteString("\n")
	}
}

func treeSearchSources(treeEntries, searchHits int) []ContextSourceInventoryEntry {
	return []ContextSourceInventoryEntry{
		{
			ID:          "filesystem-tree",
			Kind:        ContextSourceKindFilesystemTree,
			Class:       ContextSourceClassFresh,
			Description: "Bounded filesystem tree.",
			Required:    false,
			Included:    treeEntries > 0,
		},
		{
			ID:          "text-search",
			Kind:        ContextSourceKindTextSearch,
			Class:       ContextSourceClassFresh,
			Description: "Targeted text search over source, tests, docs, and .chc files.",
			Required:    false,
			Included:    searchHits > 0,
		},
	}
}
