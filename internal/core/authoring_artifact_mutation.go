package core

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileAuthoringSessionRepository) authoringArtifactChanges(binding authoringRequestBinding, snapshot AuthoringSnapshot, ref AuthoringArtifactRef, body string) ([]workflowFileChange, error) {
	if !validAuthoringArtifactRef(binding.Request.Namespace, ref) {
		return nil, workflowstate.ErrConflict
	}

	observation := authoringArtifactObservation(snapshot, ref.Path)

	replacing := binding.Request.Action == "refine" || binding.Request.Action == "ask"
	if observation != nil && !replacing {
		return nil, workflowstate.ErrConflict
	}

	if observation == nil && binding.Request.Action == "refine" {
		return nil, workflowstate.ErrConflict
	}

	metadata, err := r.authoringArtifactMetadata(binding, snapshot, ref, observation)
	if err != nil {
		return nil, err
	}

	changes := []workflowFileChange{{Path: filepath.FromSlash(ref.Path), After: []byte(body)}, {Path: filepath.Join(filepath.Dir(ref.Path), "metadata.json"), After: metadata}}
	if observation != nil {
		changes[0].Before, changes[1].Before = observation.SourceHash, observation.MetadataHash
	}

	return changes, nil
}

func (r *fileAuthoringSessionRepository) authoringArtifactMetadata(binding authoringRequestBinding, snapshot AuthoringSnapshot, ref AuthoringArtifactRef, observed *AuthoringArtifactObservation) ([]byte, error) {
	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: r.projectDir, Namespace: binding.Request.Namespace})
	if err != nil {
		return nil, err
	}

	metadata := opscore.NewOpsArtifactMetadata(opscore.CreateOpsArtifactOptions{Layout: layout, Type: authoringArtifactType(ref.Kind), Title: snapshot.Session.Topic, Producer: "chr", ExternalIDs: []string{binding.Request.SessionID}}, ref.ID, time.Now().UTC())
	if observed != nil {
		metadata, err = r.observedAuthoringMetadata(ref, *observed)
		if err != nil {
			return nil, err
		}
	}

	metadata["authoring_session"] = binding.Request.SessionID
	metadata["source_request_id"] = binding.Request.Binding.RequestID
	metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339)
	encoded, err := json.MarshalIndent(metadata, "", "  ")

	return append(encoded, '\n'), err
}

func advanceAuthoringMetadata(metadata map[string]any) error {
	revision, ok := metadata["revision"].(float64)
	if !ok || revision < 1 || revision > 4096 {
		return errors.New("invalid authoring artifact revision")
	}

	metadata["revision"] = int64(revision) + 1

	return nil
}

func authoringArtifactObservation(snapshot AuthoringSnapshot, path string) *AuthoringArtifactObservation {
	for index := range snapshot.Artifacts {
		if snapshot.Artifacts[index].Ref.Path == path {
			return &snapshot.Artifacts[index]
		}
	}

	return nil
}

func (r *fileAuthoringSessionRepository) observedAuthoringMetadata(ref AuthoringArtifactRef, observed AuthoringArtifactObservation) (map[string]any, error) {
	encoded, err := r.authoringRead(filepath.Join(filepath.Dir(ref.Path), "metadata.json"))
	if err != nil {
		return nil, err
	}

	if contenthash.Sum(encoded) != observed.MetadataHash {
		return nil, workflowstate.ErrConflict
	}

	var metadata map[string]any
	if err := json.Unmarshal(encoded, &metadata); err != nil {
		return nil, err
	}

	if err := advanceAuthoringMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
