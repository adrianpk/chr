package chc

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DirectoryName = ".chc"

const DefaultProfile = "standard"

const (
	workspaceProfilesDir = "assets/workspace/profiles"
	rootAssetDir         = "root"
)

type Workspace interface {
	Init(ctx context.Context, opts InitOptions) (InitResult, error)
}

type workspace struct {
	assetsFS fs.FS
}

func NewWorkspace(assetsFS fs.FS) Workspace {
	return &workspace{assetsFS: assetsFS}
}

type InitOptions struct {
	ProjectDir string
	Profile    string
	DryRun     bool
}

type InitResult struct {
	Root    string
	Created []string
	Skipped []string
}

type entry struct {
	path    string
	content string
}

var directories = []string{
	"agents",
	"docs",
	"specs/drafts",
	"specs/accepted",
	"specs/archived",
	"tickets/open",
	"tickets/ready",
	"tickets/in_progress",
	"tickets/reviewing",
	"tickets/blocked",
	"tickets/solved",
	"tickets/closed",
	"comments/tickets",
	"context",
	"dumps/raw",
	"dumps/curated",
	"quizzes/raw",
	"quizzes/curated",
	"playbooks",
	"reports/runs",
	"reports/audits",
	"indexes",
	"runtime/runs",
	"runtime/locks",
	"runtime/traces",
	"notifications",
}

func (w *workspace) Init(ctx context.Context, opts InitOptions) (InitResult, error) {
	if err := ctx.Err(); err != nil {
		return InitResult{}, err
	}

	absProjectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return InitResult{}, err
	}

	result := InitResult{Root: filepath.Join(absProjectDir, DirectoryName)}

	if err := initDirectories(ctx, absProjectDir, opts.DryRun, &result); err != nil {
		return InitResult{}, err
	}

	profile := opts.Profile
	if profile == "" {
		profile = DefaultProfile
	}

	if err := w.initWorkspaceFiles(ctx, absProjectDir, profile, opts.DryRun, &result); err != nil {
		return InitResult{}, err
	}

	if err := w.initRootFiles(ctx, absProjectDir, profile, opts.DryRun, &result); err != nil {
		return InitResult{}, err
	}

	return result, nil
}

func resolveProjectDir(projectDir string) (string, error) {
	if projectDir == "" {
		projectDir = "."
	}

	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}

	info, err := os.Stat(absProjectDir)
	if err != nil {
		return "", fmt.Errorf("inspect project directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("project directory is not a directory: %s", absProjectDir)
	}

	return absProjectDir, nil
}

func initDirectories(ctx context.Context, projectDir string, dryRun bool, result *InitResult) error {
	for _, dir := range directories {
		if err := ctx.Err(); err != nil {
			return err
		}

		rel := filepath.Join(DirectoryName, filepath.FromSlash(dir))
		target := filepath.Join(projectDir, rel)

		created, err := ensureDir(target, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, rel, created)
	}

	return nil
}

func (w *workspace) initWorkspaceFiles(ctx context.Context, projectDir, profile string, dryRun bool, result *InitResult) error {
	files, err := w.profileFiles(profile, workspaceFile)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		rel := filepath.Join(DirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(projectDir, rel)

		created, err := createFile(target, file.content, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, rel, created)
	}

	return nil
}

func (w *workspace) initRootFiles(ctx context.Context, projectDir, profile string, dryRun bool, result *InitResult) error {
	files, err := w.profileFiles(profile, rootFile)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		target := filepath.Join(projectDir, filepath.FromSlash(file.path))

		created, err := createFile(target, file.content, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, file.path, created)
	}

	return nil
}

type profileFileKind int

const (
	workspaceFile profileFileKind = iota
	rootFile
)

func (w *workspace) profileFiles(profile string, kind profileFileKind) ([]entry, error) {
	profileDir, err := profileDir(profile)
	if err != nil {
		return nil, err
	}

	files := []entry{}
	walk := w.profileFileWalker(profileDir, kind, &files)

	if err := fs.WalkDir(w.assetsFS, profileDir, walk); err != nil {
		return nil, fmt.Errorf("read embedded workspace profile %q: %w", profile, err)
	}

	return files, nil
}

func (w *workspace) profileFileWalker(profileDir string, kind profileFileKind, files *[]entry) fs.WalkDirFunc {
	return func(path string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if dirEntry.IsDir() {
			return nil
		}

		file, include, err := w.profileFile(profileDir, path, kind)
		if err != nil {
			return err
		}

		if include {
			*files = append(*files, file)
		}

		return nil
	}
}

func (w *workspace) profileFile(profileDir, path string, kind profileFileKind) (entry, bool, error) {
	rel, err := filepath.Rel(profileDir, path)
	if err != nil {
		return entry{}, false, fmt.Errorf("resolve embedded workspace profile path %s: %w", path, err)
	}

	if !includeProfileFile(rel, kind) {
		return entry{}, false, nil
	}

	content, err := fs.ReadFile(w.assetsFS, path)
	if err != nil {
		return entry{}, false, fmt.Errorf("read embedded workspace profile file %s: %w", path, err)
	}

	outputPath, err := outputProfilePath(rel, kind)
	if err != nil {
		return entry{}, false, fmt.Errorf("resolve embedded root profile path %s: %w", path, err)
	}

	return entry{path: outputPath, content: string(content)}, true, nil
}

func outputProfilePath(path string, kind profileFileKind) (string, error) {
	if kind != rootFile {
		return path, nil
	}

	return filepath.Rel(rootAssetDir, path)
}

func includeProfileFile(path string, kind profileFileKind) bool {
	slashPath := filepath.ToSlash(path)
	isRootFile := slashPath == rootAssetDir || strings.HasPrefix(slashPath, rootAssetDir+"/")

	if kind == rootFile {
		return isRootFile
	}

	return !isRootFile
}

func profileDir(profile string) (string, error) {
	if profile == "" {
		profile = DefaultProfile
	}

	if profile != "standard" && profile != "personal" && profile != "minimal" {
		return "", fmt.Errorf("unknown workspace profile %q", profile)
	}

	return filepath.ToSlash(filepath.Join(workspaceProfilesDir, profile)), nil
}

func recordPath(result *InitResult, rel string, created bool) {
	path := filepath.ToSlash(rel)
	if created {
		result.Created = append(result.Created, path)

		return
	}

	result.Skipped = append(result.Skipped, path)
}

func ensureDir(path string, dryRun bool) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("path exists and is not a directory: %s", path)
		}

		return false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect directory: %w", err)
	}

	if dryRun {
		return true, nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, fmt.Errorf("create directory: %w", err)
	}

	return true, nil
}

func createFile(path, content string, dryRun bool) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("path exists and is not a file: %s", path)
		}

		return false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect file: %w", err)
	}

	if dryRun {
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create parent directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}

	return true, nil
}
