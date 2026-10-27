package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type workflowUpgradeJournal struct {
	Schema      int                                   `json:"schema"`
	ProjectID   string                                `json:"project_id"`
	Namespace   string                                `json:"namespace"`
	RequestHash string                                `json:"request_hash"`
	Prior       workspaceconfig.WorkflowConfiguration `json:"prior"`
	Current     workspaceconfig.WorkflowConfiguration `json:"current"`
	Checksum    string                                `json:"checksum"`
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
	if err == nil && len(data) > workflowstate.FileLimit {
		return "", fmt.Errorf("workflow upgrade request exceeds 1 MiB")
	}

	return contenthash.Sum(data), err
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

	config, err := workspaceconfig.Parse(data)
	if err != nil {
		return workflowUpgradeJournal{}, nil, err
	}

	if config.Workflow.Version != 0 || contenthash.Sum(data) != opts.ExpectedRevision {
		return workflowUpgradeJournal{}, nil, fmt.Errorf("workflow upgrade requires the unchanged legacy configuration")
	}

	if err := s.preflightWorkflowUpgradeAdoption(ctx, opts); err != nil {
		return workflowUpgradeJournal{}, nil, err
	}

	return s.newWorkflowUpgradeJournal(data, config, opts, hash)
}

func (s *WorkflowStore) newWorkflowUpgradeJournal(data []byte, config workspaceconfig.Config, opts WorkflowUpgradeOptions, hash string) (workflowUpgradeJournal, []byte, error) {
	journal := workflowUpgradeJournal{Schema: 1, ProjectID: s.projectID, Namespace: s.layout.Namespace(), RequestHash: hash, Prior: workspaceconfig.EffectiveWorkflowConfiguration(config, "")}

	image, err := workflowUpgradeConfig(data, journal.Prior, opts)
	if err != nil {
		return journal, nil, err
	}

	current, err := workspaceconfig.Parse(image)
	if err != nil {
		return journal, nil, err
	}

	journal.Current = workspaceconfig.EffectiveWorkflowConfiguration(current, "")

	journal.Checksum, err = workflowUpgradeJournalHash(journal)
	if err != nil {
		return journal, nil, err
	}

	return journal, image, s.writeWorkflowUpgradeFile(workspaceconfig.WorkflowUpgradePending, journal)
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
	observed := contenthash.Sum(data)
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
	if err == nil && contenthash.Sum(image) != journal.Current.Revision {
		return nil, workflowstate.ErrConflict
	}

	return image, err
}

func (s *WorkflowStore) retainedWorkflowUpgrade(hash string) (workflowUpgradeJournal, bool, error) {
	data, err := s.readWorkflowUpgradeFile(workspaceconfig.WorkflowUpgradePending)
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
	if journal.Prior.Version != 0 || journal.Current.Version != 1 || !validWorkflowUpgradeConfiguration(journal.Prior) || !validWorkflowUpgradeConfiguration(journal.Current) {
		return fmt.Errorf("invalid workflow upgrade policy snapshot")
	}

	return nil
}

func validWorkflowUpgradeConfiguration(config workspaceconfig.WorkflowConfiguration) bool {
	if len(config.Values) < 13 || len(config.Values) > 15 {
		return false
	}

	allowed := map[string]bool{}
	for _, value := range workspaceconfig.EffectiveWorkflowConfiguration(workspaceconfig.Default(), "").Values {
		allowed[value.Key] = true
	}

	seen := map[string]bool{}
	for _, value := range config.Values {
		if seen[value.Key] || !allowed[value.Key] {
			return false
		}

		seen[value.Key] = true
	}

	for key := range allowed {
		if workflowUpgradeOptionalKey(key) {
			continue
		}

		if !seen[key] {
			return false
		}
	}

	return true
}

func workflowUpgradeOptionalKey(key string) bool {
	return key == "work.integration.publication_profile" || key == "work.integration.max_prepared_per_target"
}

func workflowUpgradeJournalHash(journal workflowUpgradeJournal) (string, error) {
	journal.Checksum = ""
	data, err := json.Marshal(journal)

	return contenthash.Sum(data), err
}

func checkWorkflowUpgradePolicy(data []byte, expected workspaceconfig.WorkflowConfiguration) error {
	config, err := workspaceconfig.Parse(data)
	if err != nil {
		return err
	}

	configuration := workspaceconfig.EffectiveWorkflowConfiguration(config, "")

	configuration, err = workflowUpgradeCompatibleConfiguration(config, configuration, expected)
	if err != nil {
		return err
	}

	actual, err := json.Marshal(configuration)
	if err != nil {
		return err
	}

	image, err := json.Marshal(expected)
	if err == nil && contenthash.Sum(actual) != contenthash.Sum(image) {
		return workflowstate.ErrConflict
	}

	return err
}

func withoutWorkflowPublicationProfile(values []workspaceconfig.WorkflowConfigurationValue) []workspaceconfig.WorkflowConfigurationValue {
	return withoutWorkflowConfigurationKey(values, "work.integration.publication_profile")
}

func withoutWorkflowConfigurationKey(values []workspaceconfig.WorkflowConfigurationValue, key string) []workspaceconfig.WorkflowConfigurationValue {
	result := make([]workspaceconfig.WorkflowConfigurationValue, 0, len(values))
	for _, value := range values {
		if value.Key != key {
			result = append(result, value)
		}
	}

	return result
}

func workflowUpgradeCompatibleConfiguration(config workspaceconfig.Config, configuration, expected workspaceconfig.WorkflowConfiguration) (workspaceconfig.WorkflowConfiguration, error) {
	defaults := workspaceconfig.Default().Work.Integration

	optional := []struct {
		key       string
		unchanged bool
	}{
		{"work.integration.publication_profile", config.Work.Integration.PublicationProfile == defaults.PublicationProfile},
		{"work.integration.max_prepared_per_target", config.Work.Integration.MaxPreparedPerTarget == defaults.MaxPreparedPerTarget},
	}
	for _, field := range optional {
		if slices.ContainsFunc(expected.Values, func(value workspaceconfig.WorkflowConfigurationValue) bool { return value.Key == field.key }) {
			continue
		}
		// An older immutable snapshot lacks these options. Only their established
		// defaults may be projected away; omission cannot authorize another policy.
		if !field.unchanged {
			return configuration, workflowstate.ErrConflict
		}

		configuration.Values = withoutWorkflowConfigurationKey(configuration.Values, field.key)
	}

	return configuration, nil
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
	if err == nil && contenthash.Sum(image) != contenthash.Sum(data) {
		return workflowstate.ErrConflict
	}

	return err
}

func (s *WorkflowStore) finishWorkflowUpgrade(journal workflowUpgradeJournal, image []byte) error {
	path, err := workspacecore.ConfigPath(s.layout.ProjectDir(), "config.yml")
	if err != nil {
		return err
	}

	current, err := s.readWorkflowUpgradeFile("config.yml")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if contenthash.Sum(current) != journal.Prior.Revision && contenthash.Sum(current) != journal.Current.Revision {
		return workflowstate.ErrConflict
	}

	if err := opsfile.WriteAtomic(path, image); err != nil {
		return err
	}

	if err := s.workflowUpgradeAfterReplace(2); err != nil {
		return err
	}

	return s.removeWorkflowUpgradePending()
}

func (s *WorkflowStore) removeWorkflowUpgradePending() error {
	pending, err := workspacecore.ConfigPath(s.layout.ProjectDir(), workspaceconfig.WorkflowUpgradePending)
	if err != nil {
		return err
	}

	if err := os.Remove(pending); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return opsfile.SyncDirectory(filepath.Dir(pending))
}

func (s *WorkflowStore) readWorkflowUpgradeFile(relative string) ([]byte, error) {
	path, err := workspacecore.ConfigPath(s.layout.ProjectDir(), relative)
	if err != nil {
		return nil, err
	}

	return opsfile.Read(path)
}

func (s *WorkflowStore) writeWorkflowUpgradeFile(relative string, value workflowUpgradeJournal) error {
	path, err := workspacecore.ConfigPath(s.layout.ProjectDir(), relative)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := opsfile.WriteAtomic(path, data); err != nil {
		return err
	}

	stage := 0
	if relative != workspaceconfig.WorkflowUpgradePending {
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
