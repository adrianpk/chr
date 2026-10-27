package gate

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GateGitEvidenceService describes a package boundary value.
type GateGitEvidenceService struct{}

// GateGitEvidenceRequest describes a package boundary value.
type GateGitEvidenceRequest struct {
	ProjectDir string
	BaseRef    string
	HeadRef    string
}

// GateGitEvidence describes a package boundary value.
type GateGitEvidence struct {
	Base      GateGitRef
	Head      GateGitRef
	MergeBase string
	Commits   []GateGitCommit
	Files     []GateGitChangedFile
	DiffStat  string
}

// GateGitRef describes a package boundary value.
type GateGitRef struct {
	Ref string
	SHA string
}

// GateGitCommit describes a package boundary value.
type GateGitCommit struct {
	SHA     string
	Subject string
}

// GateGitChangedFile describes a package boundary value.
type GateGitChangedFile struct {
	Status string
	Path   string
	Old    string
}

type gateGitEvidenceRequest struct {
	projectDir string
	baseRef    string
	headRef    string
	rangeRef   string
}

// NewGateGitEvidenceService constructs the corresponding package service or value.
func NewGateGitEvidenceService() GateGitEvidenceService {
	return GateGitEvidenceService{}
}

// Collect executes the corresponding package operation.
func (s GateGitEvidenceService) Collect(
	ctx context.Context,
	request GateGitEvidenceRequest,
) (GateGitEvidence, error) {
	normalized, err := normalizeGateGitEvidenceRequest(request)
	if err != nil {
		return GateGitEvidence{}, err
	}

	base, head, err := collectGateGitRefs(ctx, normalized)
	if err != nil {
		return GateGitEvidence{}, err
	}

	mergeBase, commits, files, diffStat, err := collectGateGitRange(ctx, normalized)
	if err != nil {
		return GateGitEvidence{}, err
	}

	return GateGitEvidence{
		Base:      base,
		Head:      head,
		MergeBase: mergeBase,
		Commits:   commits,
		Files:     files,
		DiffStat:  diffStat,
	}, nil
}

func normalizeGateGitEvidenceRequest(request GateGitEvidenceRequest) (gateGitEvidenceRequest, error) {
	normalized := gateGitEvidenceRequest{
		projectDir: strings.TrimSpace(request.ProjectDir),
		baseRef:    strings.TrimSpace(request.BaseRef),
		headRef:    strings.TrimSpace(request.HeadRef),
	}

	if normalized.projectDir == "" {
		return gateGitEvidenceRequest{}, fmt.Errorf("project directory is required")
	}

	if normalized.baseRef == "" {
		return gateGitEvidenceRequest{}, fmt.Errorf("base ref is required")
	}

	if normalized.headRef == "" {
		return gateGitEvidenceRequest{}, fmt.Errorf("head ref is required")
	}

	normalized.rangeRef = normalized.baseRef + ".." + normalized.headRef

	return normalized, nil
}

func collectGateGitRefs(
	ctx context.Context,
	request gateGitEvidenceRequest,
) (GateGitRef, GateGitRef, error) {
	baseSHA, err := gitOutput(ctx, request.projectDir, "rev-parse", "--verify", request.baseRef)
	if err != nil {
		return GateGitRef{}, GateGitRef{}, fmt.Errorf("resolve base ref %q: %w", request.baseRef, err)
	}

	headSHA, err := gitOutput(ctx, request.projectDir, "rev-parse", "--verify", request.headRef)
	if err != nil {
		return GateGitRef{}, GateGitRef{}, fmt.Errorf("resolve head ref %q: %w", request.headRef, err)
	}

	return GateGitRef{Ref: request.baseRef, SHA: baseSHA}, GateGitRef{Ref: request.headRef, SHA: headSHA}, nil
}

func collectGateGitRange(
	ctx context.Context,
	request gateGitEvidenceRequest,
) (string, []GateGitCommit, []GateGitChangedFile, string, error) {
	mergeBase, err := gitOutput(ctx, request.projectDir, "merge-base", request.baseRef, request.headRef)
	if err != nil {
		return "", nil, nil, "", fmt.Errorf("resolve merge base for %q and %q: %w", request.baseRef, request.headRef, err)
	}

	commitsOutput, err := gitOutput(ctx, request.projectDir, "log", "--format=%H%x09%s", request.rangeRef)
	if err != nil {
		return "", nil, nil, "", fmt.Errorf("collect commits for %q: %w", request.rangeRef, err)
	}

	filesOutput, err := gitOutput(ctx, request.projectDir, "diff", "--name-status", request.rangeRef)
	if err != nil {
		return "", nil, nil, "", fmt.Errorf("collect changed files for %q: %w", request.rangeRef, err)
	}

	diffStat, err := gitOutput(ctx, request.projectDir, "diff", "--shortstat", request.rangeRef)
	if err != nil {
		return "", nil, nil, "", fmt.Errorf("collect diff stat for %q: %w", request.rangeRef, err)
	}

	return mergeBase, parseGateGitCommits(commitsOutput), parseGateGitChangedFiles(filesOutput), diffStat, nil
}

func parseGateGitCommits(output string) []GateGitCommit {
	commits := []GateGitCommit{}

	for _, line := range gateNonEmptyLines(output) {
		sha, subject, ok := strings.Cut(line, "\t")
		if !ok {
			commits = append(commits, GateGitCommit{SHA: line})

			continue
		}

		commits = append(commits, GateGitCommit{SHA: sha, Subject: subject})
	}

	return commits
}

func parseGateGitChangedFiles(output string) []GateGitChangedFile {
	files := []GateGitChangedFile{}

	for _, line := range gateNonEmptyLines(output) {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		status := strings.TrimSpace(parts[0])
		if strings.HasPrefix(status, "R") && len(parts) >= 3 {
			files = append(files, GateGitChangedFile{
				Status: status,
				Old:    filepath.ToSlash(parts[1]),
				Path:   filepath.ToSlash(parts[2]),
			})

			continue
		}

		files = append(files, GateGitChangedFile{
			Status: status,
			Path:   filepath.ToSlash(parts[1]),
		})
	}

	return files
}

func gateNonEmptyLines(output string) []string {
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
