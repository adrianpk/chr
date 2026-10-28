package workflow

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

const workflowCoordinatorPath = "coordinator.json"

// ErrWorkflowCoordinator blocks execution outside the registered local owner.
var ErrWorkflowCoordinator = errors.New("workflow coordinator is not registered for this repository")

type workflowCoordinator struct {
	Schema string `json:"schema"`
	ID     string `json:"id"`
}

type workflowCoordinatorRegistration struct {
	workflowCoordinator
	Owner string `json:"owner"`
}

func workflowCoordinatorOwner(common string) (string, error) {
	machine, err := os.ReadFile("/etc/machine-id")
	if err != nil || len(machine) == 0 {
		return "", errors.Join(ErrWorkflowCoordinator, err)
	}

	info, err := os.Stat(common)
	if err != nil {
		return "", err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", ErrWorkflowCoordinator
	}
	// A move preserves the directory identity. A copied Git directory or another
	// host cannot obtain authority by copying its registration file.
	image := fmt.Appendf(machine, "\x00%d\x00%d", stat.Dev, stat.Ino)

	return contenthash.Sum(image), nil
}

func (s *Store) workflowLocalCoordinator(ctx context.Context, register bool) (workflowCoordinator, error) {
	common, err := projectcore.ResolveRepositoryDirectory(ctx, s.layout.ProjectDir())
	if err != nil {
		return workflowCoordinator{}, err
	}

	path := filepath.Join(common, "chr-workflow-coordinator.json")

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) && register {
		return s.registerWorkflowCoordinator(ctx, common, path)
	}

	if err != nil {
		return workflowCoordinator{}, errors.Join(ErrWorkflowCoordinator, err)
	}

	var registration workflowCoordinatorRegistration
	if err := json.Unmarshal(data, &registration); err != nil {
		return workflowCoordinator{}, errors.Join(ErrWorkflowCoordinator, err)
	}

	owner, err := workflowCoordinatorOwner(common)
	if err != nil || registration.Owner != owner || !validWorkflowCoordinator(registration.workflowCoordinator) {
		return workflowCoordinator{}, errors.Join(ErrWorkflowCoordinator, err)
	}

	return registration.workflowCoordinator, nil
}

func validWorkflowCoordinator(coordinator workflowCoordinator) bool {
	_, err := hex.DecodeString(coordinator.ID)

	return coordinator.Schema == "chr-coordinator/v1" && len(coordinator.ID) == 64 && err == nil
}

func (s *Store) registerWorkflowCoordinator(ctx context.Context, common, path string) (workflowCoordinator, error) {
	markers, err := s.workflowCoordinatorMarkers(ctx)
	if err != nil {
		return workflowCoordinator{}, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return workflowCoordinator{}, err
	}

	if len(markers) != 0 || len(snapshot.Records) != 0 {
		return workflowCoordinator{}, fmt.Errorf("%w: copied or unregistered authority requires coordinator reconciliation", ErrWorkflowCoordinator)
	}

	id, err := newWorkflowClaimToken()
	if err != nil {
		return workflowCoordinator{}, err
	}

	owner, err := workflowCoordinatorOwner(common)
	if err != nil {
		return workflowCoordinator{}, err
	}

	registration := workflowCoordinatorRegistration{workflowCoordinator: workflowCoordinator{Schema: "chr-coordinator/v1", ID: id}, Owner: owner}

	data, err := json.Marshal(registration)
	if err != nil {
		return workflowCoordinator{}, err
	}

	return registration.workflowCoordinator, opsfile.WriteAtomic(path, data)
}

func (s *Store) workflowCoordinatorMarkers(ctx context.Context) ([]workflowCoordinator, error) {
	// Worker-owned lane copies do not redefine the host coordinator. Their
	// committed registration is checked separately before accepting a result.
	layouts, err := s.workflowOperationLayouts(ctx)
	if err != nil {
		return nil, err
	}

	markers := []workflowCoordinator{}

	for _, layout := range layouts {
		marker, err := readWorkflowCoordinator(filepath.Join(layout.Root(), workflowCoordinatorPath))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		markers = append(markers, marker)
	}

	return markers, nil
}

func readWorkflowCoordinator(path string) (workflowCoordinator, error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return workflowCoordinator{}, err
	}

	return decodeWorkflowCoordinator(data)
}

func decodeWorkflowCoordinator(data []byte) (workflowCoordinator, error) {
	var marker workflowCoordinator
	if err := json.Unmarshal(data, &marker); err != nil || !validWorkflowCoordinator(marker) {
		return marker, errors.Join(ErrWorkflowCoordinator, err)
	}

	return marker, nil
}

func (s *Store) checkWorkflowCoordinator(ctx context.Context) (workflowCoordinator, error) {
	coordinator, err := s.workflowLocalCoordinator(ctx, false)
	if err != nil {
		return coordinator, err
	}

	markers, err := s.workflowCoordinatorMarkers(ctx)
	if err != nil {
		return coordinator, err
	}

	if len(markers) == 0 {
		return coordinator, ErrWorkflowCoordinator
	}

	return coordinator, checkWorkflowCoordinatorMarkers(markers, coordinator)
}

func checkWorkflowCoordinatorMarkers(markers []workflowCoordinator, owner workflowCoordinator) error {
	for _, marker := range markers {
		if marker != owner {
			return fmt.Errorf("%w: workflow authority belongs to another coordinator", ErrWorkflowCoordinator)
		}
	}

	return nil
}
