package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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
	Root      string
	Created   []string
	Skipped   []string
	Preserved []string
}

type entry struct {
	path    string
	content string
}

var directories = []string{
	workspacecore.UserDirectoryName + "/agents",
	workspacecore.UserDirectoryName + "/docs",
	workspacecore.UserDirectoryName + "/specs/drafts",
	workspacecore.UserDirectoryName + "/specs/accepted",
	workspacecore.UserDirectoryName + "/specs/archived",
	workspacecore.UserDirectoryName + "/tickets/open",
	workspacecore.UserDirectoryName + "/tickets/ready",
	workspacecore.UserDirectoryName + "/tickets/in_progress",
	workspacecore.UserDirectoryName + "/tickets/reviewing",
	workspacecore.UserDirectoryName + "/tickets/blocked",
	workspacecore.UserDirectoryName + "/tickets/solved",
	workspacecore.UserDirectoryName + "/tickets/closed",
	workspacecore.UserDirectoryName + "/comments/tickets",
	workspacecore.UserDirectoryName + "/context",
	workspacecore.UserDirectoryName + "/dumps/raw",
	workspacecore.UserDirectoryName + "/dumps/curated",
	workspacecore.UserDirectoryName + "/quizzes/raw",
	workspacecore.UserDirectoryName + "/quizzes/curated",
	workspacecore.UserDirectoryName + "/playbooks",
	workspacecore.UserDirectoryName + "/reports/runs",
	workspacecore.UserDirectoryName + "/reports/audits",
	workspacecore.UserDirectoryName + "/indexes",
	workspacecore.UserDirectoryName + "/runtime/locks",
	workspacecore.UserDirectoryName + "/runtime/traces",
	workspacecore.UserDirectoryName + "/notifications",
}

// Init executes the corresponding package operation.
func (w *workspace) Init(ctx context.Context, opts InitOptions) (InitResult, error) {
	err := ctx.Err()
	if err != nil {
		return InitResult{}, err
	}

	absProjectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return InitResult{}, err
	}

	result := InitResult{Root: filepath.Join(absProjectDir, workspacecore.DirectoryName)}

	_, err = os.Stat(result.Root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return InitResult{}, fmt.Errorf("inspect workspace: %w", err)
	}

	newProject := errors.Is(err, os.ErrNotExist)

	err = initDirectories(ctx, absProjectDir, opts.DryRun, &result)
	if err != nil {
		return InitResult{}, err
	}

	profile := opts.Profile
	if profile == "" {
		profile = DefaultProfile
	}

	err = w.initWorkspaceFiles(ctx, workspaceFileInitialization{ProjectDir: absProjectDir, Profile: profile, NewProject: newProject, DryRun: opts.DryRun}, &result)
	if err != nil {
		return InitResult{}, err
	}

	err = w.initRootFiles(ctx, absProjectDir, profile, opts.DryRun, &result)
	if err != nil {
		return InitResult{}, err
	}

	return result, nil
}

func initDirectories(ctx context.Context, projectDir string, dryRun bool, result *InitResult) error {
	for _, dir := range directories {
		err := ctx.Err()
		if err != nil {
			return err
		}

		rel := filepath.Join(workspacecore.DirectoryName, filepath.FromSlash(dir))
		target := filepath.Join(projectDir, rel)

		created, err := ensureDir(target, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, rel, created)
	}

	return nil
}

type workspaceFileInitialization struct {
	ProjectDir string
	Profile    string
	NewProject bool
	DryRun     bool
}

func (w *workspace) initWorkspaceFiles(ctx context.Context, input workspaceFileInitialization, result *InitResult) error {
	files, err := w.profileFiles(input.Profile, workspaceFile)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = ctx.Err()
		if err != nil {
			return err
		}

		rel := filepath.Join(workspacecore.DirectoryName, workspacecore.UserDirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(input.ProjectDir, rel)
		// Absence is an existing policy choice, not permission to enable version one.
		if file.path == "config.yml" && !input.NewProject {
			_, err = os.Stat(target)
			if errors.Is(err, os.ErrNotExist) {
				result.Preserved = append(result.Preserved, filepath.ToSlash(rel))

				continue
			}
		}

		created, err := createFile(target, file.content, input.DryRun)
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

		result, err := w.profileFile(profileDir, path, kind)
		if err != nil {
			return err
		}

		if result.Included {
			*files = append(*files, result.File)
		}

		return nil
	}
}

type profileFileResult struct {
	File     entry
	Included bool
}

func (w *workspace) profileFile(profileDir, path string, kind profileFileKind) (profileFileResult, error) {
	rel, err := filepath.Rel(profileDir, path)
	if err != nil {
		return profileFileResult{}, fmt.Errorf("resolve embedded workspace profile path %s: %w", path, err)
	}

	if !includeProfileFile(rel, kind) {
		return profileFileResult{}, nil
	}

	content, err := fs.ReadFile(w.assetsFS, path)
	if err != nil {
		return profileFileResult{}, fmt.Errorf("read embedded workspace profile file %s: %w", path, err)
	}

	outputPath, err := outputProfilePath(rel, kind)
	if err != nil {
		return profileFileResult{}, fmt.Errorf("resolve embedded root profile path %s: %w", path, err)
	}

	return profileFileResult{File: entry{path: outputPath, content: string(content)}, Included: true}, nil
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

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}

	_, writeErr := file.WriteString(content)

	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return false, fmt.Errorf("write created file: %w", err)
	}

	return true, nil
}
