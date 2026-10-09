package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type workflowUpgradeJournal struct {
	Schema      int                   `json:"schema"`
	ProjectID   string                `json:"project_id"`
	Namespace   string                `json:"namespace"`
	RequestHash string                `json:"request_hash"`
	Prior       WorkflowConfiguration `json:"prior"`
	Current     WorkflowConfiguration `json:"current"`
	Checksum    string                `json:"checksum"`
}

func (s *WorkflowStore) upgradeWorkflowConfiguration(ctx context.Context, opts WorkflowUpgradeOptions) (WorkflowUpgradeResult, error) {
	hash, err := workflowUpgradeRequestHash(opts)
	if err != nil {
		return WorkflowUpgradeResult{}, err
	}

	journal, data, err := s.prepareWorkflowUpgrade(ctx, opts, hash)
	if err != nil {
		return WorkflowUpgradeResult{}, err
	}

	result := WorkflowUpgradeResult{Snapshot: filepath.ToSlash(filepath.Join(".chr", "default", "workflow-upgrades", hash+".json")), Prior: journal.Prior, Current: journal.Current}

	if opts.Adoption != nil {
		adopted, err := s.adoptPlanWorkflow(ctx, *opts.Adoption)
		if err != nil {
			return result, err
		}

		result.Adopted = &adopted
	}

	if err := s.writeWorkflowUpgradeSnapshot(journal); err != nil {
		return result, err
	}

	if err := s.finishWorkflowUpgrade(journal, data); err != nil {
		return result, err
	}

	return result, nil
}

func workflowUpgradeRequestHash(opts WorkflowUpgradeOptions) (string, error) {
	data, err := json.Marshal(opts)
	if err == nil && len(data) > WorkflowFileLimit {
		return "", fmt.Errorf("workflow upgrade request exceeds 1 MiB")
	}

	return workflowHash(data), err
}

func (s *WorkflowStore) prepareWorkflowUpgrade(ctx context.Context, opts WorkflowUpgradeOptions, hash string) (workflowUpgradeJournal, []byte, error) {
	data, err := s.readWorkflowUpgradeFile("config.yml")
	if errors.Is(err, os.ErrNotExist) {
		data, err = nil, nil
	}

	if err != nil {
		return workflowUpgradeJournal{}, nil, err
	}

	if journal, found, err := s.retainedWorkflowUpgrade(hash); found || err != nil {
		image, imageErr := workflowUpgradeJournalImage(journal, data, opts)

		return journal, image, errors.Join(err, imageErr)
	}

	config, err := parseWorkflowConfig(data)
	if err != nil {
		return workflowUpgradeJournal{}, nil, err
	}

	if config.Workflow.Version != 0 || workflowHash(data) != opts.ExpectedRevision {
		return workflowUpgradeJournal{}, nil, fmt.Errorf("workflow upgrade requires the unchanged legacy configuration")
	}

	if err := s.preflightWorkflowUpgradeAdoption(ctx, opts); err != nil {
		return workflowUpgradeJournal{}, nil, err
	}

	return s.newWorkflowUpgradeJournal(data, config, opts, hash)
}

func (s *WorkflowStore) newWorkflowUpgradeJournal(data []byte, config WorkspaceConfig, opts WorkflowUpgradeOptions, hash string) (workflowUpgradeJournal, []byte, error) {
	journal := workflowUpgradeJournal{Schema: 1, ProjectID: s.projectID, Namespace: s.layout.Namespace(), RequestHash: hash, Prior: effectiveWorkflowConfiguration(config, "")}

	image, err := workflowUpgradeConfig(data, journal.Prior, opts)
	if err != nil {
		return journal, nil, err
	}

	current, err := parseWorkflowConfig(image)
	if err != nil {
		return journal, nil, err
	}

	journal.Current = effectiveWorkflowConfiguration(current, "")

	journal.Checksum, err = workflowUpgradeJournalHash(journal)
	if err != nil {
		return journal, nil, err
	}

	return journal, image, s.writeWorkflowUpgradeFile(workflowUpgradePending, journal)
}

func (s *WorkflowStore) preflightWorkflowUpgradeAdoption(ctx context.Context, opts WorkflowUpgradeOptions) error {
	if opts.Adoption == nil {
		return nil
	}

	result, err := s.preparePlanAdoption(ctx, *opts.Adoption)
	if err != nil {
		return err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	snapshot.Records = append(snapshot.Records, result.Set)
	snapshot.Records = append(snapshot.Records, result.Tasks...)

	return snapshot.Validate()
}

func workflowUpgradeJournalImage(journal workflowUpgradeJournal, data []byte, opts WorkflowUpgradeOptions) ([]byte, error) {
	observed := workflowHash(data)
	if observed == journal.Current.Revision {
		return data, checkWorkflowUpgradePolicy(data, journal.Current)
	}

	if observed != journal.Prior.Revision || journal.Prior.Revision != opts.ExpectedRevision {
		return nil, fmt.Errorf("workflow upgrade configuration changed; reconcile the pending upgrade")
	}

	if err := checkWorkflowUpgradePolicy(data, journal.Prior); err != nil {
		return nil, err
	}

	image, err := workflowUpgradeConfig(data, journal.Prior, opts)
	if err == nil && workflowHash(image) != journal.Current.Revision {
		return nil, ErrWorkflowConflict
	}

	return image, err
}

func (s *WorkflowStore) retainedWorkflowUpgrade(hash string) (workflowUpgradeJournal, bool, error) {
	data, err := s.readWorkflowUpgradeFile(workflowUpgradePending)
	if errors.Is(err, os.ErrNotExist) {
		data, err = s.readWorkflowUpgradeFile("workflow-upgrades/" + hash + ".json")
	}

	if errors.Is(err, os.ErrNotExist) {
		return workflowUpgradeJournal{}, false, nil
	}

	journal := workflowUpgradeJournal{}
	if err != nil {
		return journal, false, err
	}

	if err := json.Unmarshal(data, &journal); err != nil {
		return journal, true, err
	}

	return journal, true, s.validateWorkflowUpgradeJournal(journal, hash)
}

func (s *WorkflowStore) validateWorkflowUpgradeJournal(journal workflowUpgradeJournal, requestHash string) error {
	hash, err := workflowUpgradeJournalHash(journal)
	if err != nil {
		return err
	}

	if journal.Schema != 1 || journal.ProjectID != s.projectID || journal.Namespace != s.layout.Namespace() {
		return fmt.Errorf("foreign or malformed workflow upgrade journal")
	}

	if journal.RequestHash != requestHash || hash != journal.Checksum {
		return fmt.Errorf("workflow upgrade journal checksum or request changed")
	}

	return validateWorkflowUpgradePolicyPair(journal)
}

func validateWorkflowUpgradePolicyPair(journal workflowUpgradeJournal) error {
	if len(journal.Prior.Values) != 13 || len(journal.Current.Values) != 13 || journal.Prior.Version != 0 || journal.Current.Version != 1 {
		return fmt.Errorf("invalid workflow upgrade policy snapshot")
	}

	return nil
}

func workflowUpgradeJournalHash(journal workflowUpgradeJournal) (string, error) {
	journal.Checksum = ""
	data, err := json.Marshal(journal)

	return workflowHash(data), err
}

func checkWorkflowUpgradePolicy(data []byte, expected WorkflowConfiguration) error {
	config, err := parseWorkflowConfig(data)
	if err != nil {
		return err
	}

	actual, err := json.Marshal(effectiveWorkflowConfiguration(config, ""))
	if err != nil {
		return err
	}

	image, err := json.Marshal(expected)
	if err == nil && workflowHash(actual) != workflowHash(image) {
		return ErrWorkflowConflict
	}

	return err
}

func (s *WorkflowStore) writeWorkflowUpgradeSnapshot(journal workflowUpgradeJournal) error {
	relative := "workflow-upgrades/" + journal.RequestHash + ".json"

	data, err := s.readWorkflowUpgradeFile(relative)
	if errors.Is(err, os.ErrNotExist) {
		return s.writeWorkflowUpgradeFile(relative, journal)
	}

	if err != nil {
		return err
	}

	image, err := json.Marshal(journal)
	if err == nil && workflowHash(image) != workflowHash(data) {
		return ErrWorkflowConflict
	}

	return err
}

func (s *WorkflowStore) finishWorkflowUpgrade(journal workflowUpgradeJournal, image []byte) error {
	path, err := workflowUpgradePath(s.layout.ProjectDir(), "config.yml")
	if err != nil {
		return err
	}

	current, err := s.readWorkflowUpgradeFile("config.yml")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if workflowHash(current) != journal.Prior.Revision && workflowHash(current) != journal.Current.Revision {
		return ErrWorkflowConflict
	}

	if err := atomicWorkflowFile(path, image); err != nil {
		return err
	}

	if err := s.workflowUpgradeAfterReplace(2); err != nil {
		return err
	}

	return s.removeWorkflowUpgradePending()
}

func (s *WorkflowStore) removeWorkflowUpgradePending() error {
	pending, err := workflowUpgradePath(s.layout.ProjectDir(), workflowUpgradePending)
	if err != nil {
		return err
	}

	if err := os.Remove(pending); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return syncWorkflowDirectory(filepath.Dir(pending))
}

func (s *WorkflowStore) readWorkflowUpgradeFile(relative string) ([]byte, error) {
	path, err := workflowUpgradePath(s.layout.ProjectDir(), relative)
	if err != nil {
		return nil, err
	}

	return readWorkflowFile(path)
}

func (s *WorkflowStore) writeWorkflowUpgradeFile(relative string, value workflowUpgradeJournal) error {
	path, err := workflowUpgradePath(s.layout.ProjectDir(), relative)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := atomicWorkflowFile(path, data); err != nil {
		return err
	}

	stage := 0
	if relative != workflowUpgradePending {
		stage = 1
	}

	return s.workflowUpgradeAfterReplace(stage)
}

func (s *WorkflowStore) workflowUpgradeAfterReplace(stage int) error {
	if s.afterReplace != nil {
		return s.afterReplace(stage)
	}

	return nil
}
