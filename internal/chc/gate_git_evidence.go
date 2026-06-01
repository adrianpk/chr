package chc

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type GateGitEvidenceService struct{}

type GateGitEvidenceRequest struct {
	ProjectDir string
	BaseRef    string
	HeadRef    string
}

type GateGitEvidence struct {
	Base      GateGitRef
	Head      GateGitRef
	MergeBase string
	Commits   []GateGitCommit
	Files     []GateGitChangedFile
	DiffStat  string
}

type GateGitRef struct {
	Ref string
	SHA string
}

type GateGitCommit struct {
	SHA     string
	Subject string
}

type GateGitChangedFile struct {
	Status string
	Path   string
	Old    string
}

func NewGateGitEvidenceService() GateGitEvidenceService {
	return GateGitEvidenceService{}
}

func (s GateGitEvidenceService) Collect(
	ctx context.Context,
	request GateGitEvidenceRequest,
) (GateGitEvidence, error) {
	projectDir := strings.TrimSpace(request.ProjectDir)
	if projectDir == "" {
		return GateGitEvidence{}, fmt.Errorf("project directory is required")
	}

	baseRef := strings.TrimSpace(request.BaseRef)
	if baseRef == "" {
		return GateGitEvidence{}, fmt.Errorf("base ref is required")
	}

	headRef := strings.TrimSpace(request.HeadRef)
	if headRef == "" {
		return GateGitEvidence{}, fmt.Errorf("head ref is required")
	}

	baseSHA, err := gitOutput(ctx, projectDir, "rev-parse", "--verify", baseRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("resolve base ref %q: %w", baseRef, err)
	}

	headSHA, err := gitOutput(ctx, projectDir, "rev-parse", "--verify", headRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("resolve head ref %q: %w", headRef, err)
	}

	mergeBase, err := gitOutput(ctx, projectDir, "merge-base", baseRef, headRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("resolve merge base for %q and %q: %w", baseRef, headRef, err)
	}

	commitsOutput, err := gitOutput(ctx, projectDir, "log", "--format=%H%x09%s", baseRef+".."+headRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("collect commits for %q..%q: %w", baseRef, headRef, err)
	}

	filesOutput, err := gitOutput(ctx, projectDir, "diff", "--name-status", baseRef+".."+headRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("collect changed files for %q..%q: %w", baseRef, headRef, err)
	}

	diffStat, err := gitOutput(ctx, projectDir, "diff", "--shortstat", baseRef+".."+headRef)
	if err != nil {
		return GateGitEvidence{}, fmt.Errorf("collect diff stat for %q..%q: %w", baseRef, headRef, err)
	}

	return GateGitEvidence{
		Base: GateGitRef{
			Ref: baseRef,
			SHA: baseSHA,
		},
		Head: GateGitRef{
			Ref: headRef,
			SHA: headSHA,
		},
		MergeBase: mergeBase,
		Commits:   parseGateGitCommits(commitsOutput),
		Files:     parseGateGitChangedFiles(filesOutput),
		DiffStat:  diffStat,
	}, nil
}

func gitOutput(ctx context.Context, projectDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", projectDir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", strings.Join(cmd.Args, " "), strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
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
