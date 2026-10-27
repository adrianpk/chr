package gate

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitEvidenceService describes a package boundary value.
type GitEvidenceService struct{}

// GitEvidenceRequest describes a package boundary value.
type GitEvidenceRequest struct {
	ProjectDir string
	BaseRef    string
	HeadRef    string
}

// GitEvidence describes a package boundary value.
type GitEvidence struct {
	Base      GitRef
	Head      GitRef
	MergeBase string
	Commits   []GitCommit
	Files     []GitChangedFile
	DiffStat  string
}

// GitRef describes a package boundary value.
type GitRef struct {
	Ref string
	SHA string
}

// GitCommit describes a package boundary value.
type GitCommit struct {
	SHA     string
	Subject string
}

// GitChangedFile describes a package boundary value.
type GitChangedFile struct {
	Status string
	Path   string
	Old    string
}

type gitEvidenceRequest struct {
	projectDir string
	baseRef    string
	headRef    string
	rangeRef   string
}

// NewGitEvidenceService constructs the corresponding package service or value.
func NewGitEvidenceService() GitEvidenceService {
	return GitEvidenceService{}
}

// Collect executes the corresponding package operation.
func (s GitEvidenceService) Collect(ctx context.Context, request GitEvidenceRequest) (GitEvidence, error) {
	normalized, err := normalizeGitEvidenceRequest(request)
	if err != nil {
		return GitEvidence{}, err
	}

	refs, err := collectGitRefs(ctx, normalized)
	base, head := refs.Base, refs.Head

	if err != nil {
		return GitEvidence{}, err
	}

	gitRangeResult, err := collectGitRange(ctx, normalized)
	mergeBase, commits, files, diffStat := gitRangeResult.MergeBase, gitRangeResult.Commits, gitRangeResult.Files, gitRangeResult.DiffStat

	if err != nil {
		return GitEvidence{}, err
	}

	return GitEvidence{
		Base:      base,
		Head:      head,
		MergeBase: mergeBase,
		Commits:   commits,
		Files:     files,
		DiffStat:  diffStat,
	}, nil
}

func normalizeGitEvidenceRequest(request GitEvidenceRequest) (gitEvidenceRequest, error) {
	normalized := gitEvidenceRequest{
		projectDir: strings.TrimSpace(request.ProjectDir),
		baseRef:    strings.TrimSpace(request.BaseRef),
		headRef:    strings.TrimSpace(request.HeadRef),
	}

	if normalized.projectDir == "" {
		return gitEvidenceRequest{}, fmt.Errorf("project directory is required")
	}

	if normalized.baseRef == "" {
		return gitEvidenceRequest{}, fmt.Errorf("base ref is required")
	}

	if normalized.headRef == "" {
		return gitEvidenceRequest{}, fmt.Errorf("head ref is required")
	}

	normalized.rangeRef = normalized.baseRef + ".." + normalized.headRef

	return normalized, nil
}

func collectGitRefs(ctx context.Context, request gitEvidenceRequest) (gitRefs, error) {
	baseSHA, err := gitOutput(ctx, request.projectDir, "rev-parse", "--verify", request.baseRef)
	if err != nil {
		return gitRefs{Base: GitRef{}, Head: GitRef{}}, fmt.Errorf("resolve base ref %q: %w", request.baseRef, err)
	}

	headSHA, err := gitOutput(ctx, request.projectDir, "rev-parse", "--verify", request.headRef)
	if err != nil {
		return gitRefs{Base: GitRef{}, Head: GitRef{}}, fmt.Errorf("resolve head ref %q: %w", request.headRef, err)
	}

	return gitRefs{Base: GitRef{Ref: request.baseRef, SHA: baseSHA}, Head: GitRef{Ref: request.headRef, SHA: headSHA}}, nil
}

func collectGitRange(ctx context.Context, request gitEvidenceRequest) (gitRange, error) {
	mergeBase, err := gitOutput(ctx, request.projectDir, "merge-base", request.baseRef, request.headRef)
	if err != nil {
		return gitRange{MergeBase: "", Commits: nil, Files: nil, DiffStat: ""}, fmt.Errorf("resolve merge base for %q and %q: %w", request.baseRef, request.headRef, err)
	}

	commitsOutput, err := gitOutput(ctx, request.projectDir, "log", "--format=%H%x09%s", request.rangeRef)
	if err != nil {
		return gitRange{MergeBase: "", Commits: nil, Files: nil, DiffStat: ""}, fmt.Errorf("collect commits for %q: %w", request.rangeRef, err)
	}

	filesOutput, err := gitOutput(ctx, request.projectDir, "diff", "--name-status", request.rangeRef)
	if err != nil {
		return gitRange{MergeBase: "", Commits: nil, Files: nil, DiffStat: ""}, fmt.Errorf("collect changed files for %q: %w", request.rangeRef, err)
	}

	diffStat, err := gitOutput(ctx, request.projectDir, "diff", "--shortstat", request.rangeRef)
	if err != nil {
		return gitRange{MergeBase: "", Commits: nil, Files: nil, DiffStat: ""}, fmt.Errorf("collect diff stat for %q: %w", request.rangeRef, err)
	}

	return gitRange{MergeBase: mergeBase, Commits: parseGitCommits(commitsOutput), Files: parseGitChangedFiles(filesOutput), DiffStat: diffStat}, nil
}

func parseGitCommits(output string) []GitCommit {
	commits := []GitCommit{}

	for _, line := range nonEmptyLines(output) {
		sha, subject, ok := strings.Cut(line, "\t")
		if !ok {
			commits = append(commits, GitCommit{SHA: line})

			continue
		}

		commits = append(commits, GitCommit{SHA: sha, Subject: subject})
	}

	return commits
}

func parseGitChangedFiles(output string) []GitChangedFile {
	files := []GitChangedFile{}

	for _, line := range nonEmptyLines(output) {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		status := strings.TrimSpace(parts[0])
		if strings.HasPrefix(status, "R") && len(parts) >= 3 {
			files = append(files, GitChangedFile{
				Status: status,
				Old:    filepath.ToSlash(parts[1]),
				Path:   filepath.ToSlash(parts[2]),
			})

			continue
		}

		files = append(files, GitChangedFile{
			Status: status,
			Path:   filepath.ToSlash(parts[1]),
		})
	}

	return files
}

func nonEmptyLines(output string) []string {
	lines := []string{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func gitOutput(ctx context.Context, projectDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", projectDir}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", strings.Join(cmd.Args, " "), strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
}

type (
	gitRefs  struct{ Base, Head GitRef }
	gitRange struct {
		MergeBase string
		Commits   []GitCommit
		Files     []GitChangedFile
		DiffStat  string
	}
)
