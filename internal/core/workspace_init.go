package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DirectoryName defines a package boundary constant.
	DirectoryName = ".chr"
	// UserDirectoryName defines a package boundary constant.
	UserDirectoryName = "user"
)

// DefaultProfile defines a package boundary constant.
const DefaultProfile = "standard"

const (
	workspaceProfilesDir = "assets/workspace/profiles"
	rootAssetDir         = "root"
)

// Workspace defines a package boundary interface.
type Workspace interface {
	Init(ctx context.Context, opts InitOptions) (InitResult, error)
}

type workspace struct {
	assetsFS fs.FS
}

// NewWorkspace constructs the corresponding package service or value.
func NewWorkspace(assetsFS fs.FS) Workspace {
	return &workspace{assetsFS: assetsFS}
}

// InitOptions describes a package boundary value.
type InitOptions struct {
	ProjectDir string
	Profile    string
	DryRun     bool
}

// InitResult describes a package boundary value.
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
	"user/agents",
	"user/docs",
	"user/specs/drafts",
	"user/specs/accepted",
	"user/specs/archived",
	"user/tickets/open",
	"user/tickets/ready",
	"user/tickets/in_progress",
	"user/tickets/reviewing",
	"user/tickets/blocked",
	"user/tickets/solved",
	"user/tickets/closed",
	"user/comments/tickets",
	"user/context",
	"user/dumps/raw",
	"user/dumps/curated",
	"user/quizzes/raw",
	"user/quizzes/curated",
	"user/playbooks",
	"user/reports/runs",
	"user/reports/audits",
	"user/indexes",
	"user/runtime/runs",
	"user/runtime/locks",
	"user/runtime/traces",
	"user/notifications",
}

// UserRelativePath executes the corresponding package operation.
func UserRelativePath(elements ...string) string {
	parts := append([]string{DirectoryName, UserDirectoryName}, elements...)

	return filepath.ToSlash(filepath.Join(parts...))
}

// UserPath executes the corresponding package operation.
func UserPath(projectDir string, elements ...string) string {
	return filepath.Join(projectDir, filepath.FromSlash(UserRelativePath(elements...)))
}

// Init executes the corresponding package operation.
func (w *workspace) Init(ctx context.Context, opts InitOptions) (InitResult, error) {
	err := ctx.Err()
	if err != nil {
		return InitResult{}, err
	}

	absProjectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return InitResult{}, err
	}

	result := InitResult{Root: filepath.Join(absProjectDir, DirectoryName)}

	err = initDirectories(ctx, absProjectDir, opts.DryRun, &result)
	if err != nil {
		return InitResult{}, err
	}

	profile := opts.Profile
	if profile == "" {
		profile = DefaultProfile
	}

	err = w.initWorkspaceFiles(ctx, absProjectDir, profile, opts.DryRun, &result)
	if err != nil {
		return InitResult{}, err
	}

	err = w.initRootFiles(ctx, absProjectDir, profile, opts.DryRun, &result)
	if err != nil {
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
		err := ctx.Err()
		if err != nil {
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
		err = ctx.Err()
		if err != nil {
			return err
		}

		rel := filepath.Join(DirectoryName, UserDirectoryName, filepath.FromSlash(file.path))
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
		err = ctx.Err()
		if err != nil {
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

	err = fs.WalkDir(w.assetsFS, profileDir, walk)
	if err != nil {
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

	err = os.MkdirAll(path, 0o755)
	if err != nil {
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

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return false, fmt.Errorf("create parent directory: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}

	return true, nil
}
