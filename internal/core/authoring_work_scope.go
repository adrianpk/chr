package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// AuthoringIntent separates a requested explanation, preparation and development.
type AuthoringIntent string

// Only an explicit development decision can authorize a prepared scope.
const (
	AuthoringIntentExplain AuthoringIntent = "explain"
	AuthoringIntentPrepare AuthoringIntent = "prepare"
	AuthoringIntentDevelop AuthoringIntent = "develop"
)

// AuthoringWorkScope binds a proposed native description to canonical inputs.
// It does not authorize admission, alter policy or replace a workflow preview.
type AuthoringWorkScope struct {
	Schema          string                  `json:"schema"`
	Binding         WorkflowMutationBinding `json:"binding"`
	Intent          AuthoringIntent         `json:"intent"`
	Instruction     string                  `json:"instruction"`
	ProjectID       string                  `json:"project_id"`
	Namespace       string                  `json:"namespace"`
	SessionID       string                  `json:"session_id"`
	SessionRef      string                  `json:"session_ref"`
	SessionRevision int64                   `json:"session_revision"`
	SourceHash      string                  `json:"source_hash"`
	DescriptionRef  string                  `json:"description_ref"`
	DescriptionHash string                  `json:"description_hash"`
	ContractHash    string                  `json:"contract_hash"`
	RecordHash      string                  `json:"record_hash"`
	Hash            string                  `json:"scope_hash"`
}

// AuthoringWorkAuthorization retains a host-validated explicit decision for the
// exact displayed scope. Provider output cannot supply this value.
type AuthoringWorkAuthorization struct {
	Binding   WorkflowMutationBinding `json:"binding"`
	ScopeHash string                  `json:"scope_hash"`
}

// SealAuthoringWorkScope computes the immutable scope binding after validation.
func SealAuthoringWorkScope(scope AuthoringWorkScope) (AuthoringWorkScope, error) {
	scope.Schema, scope.Hash = "chr-authoring-work/v1", ""
	if err := validateAuthoringWorkScope(scope); err != nil {
		return scope, err
	}

	encoded, err := json.Marshal(scope)
	scope.Hash = workflowHash(encoded)

	return scope, err
}

// ValidateAuthoringWorkScope checks the original scope and its complete digest.
func ValidateAuthoringWorkScope(scope AuthoringWorkScope) error {
	expected, err := SealAuthoringWorkScope(scope)
	if err != nil {
		return err
	}

	if scope.Schema != expected.Schema || scope.Hash != expected.Hash {
		return ErrWorkflowConflict
	}

	return nil
}

// AuthorizeAuthoringWork binds an explicit caller decision to the observed scope.
// A changed description or source requires a new decision for the changed hash.
func AuthorizeAuthoringWork(scope AuthoringWorkScope, decision WorkflowMutationBinding, observed string) (AuthoringWorkAuthorization, error) {
	if err := ValidateAuthoringWorkScope(scope); err != nil {
		return AuthoringWorkAuthorization{}, err
	}

	if err := validateWorkflowMutationBinding(decision); err != nil {
		return AuthoringWorkAuthorization{}, err
	}

	if scope.Intent != AuthoringIntentDevelop || decision.Operator != scope.Binding.Operator || observed != scope.Hash {
		return AuthoringWorkAuthorization{}, ErrWorkflowConflict
	}

	return AuthoringWorkAuthorization{Binding: decision, ScopeHash: observed}, nil
}

// ValidateAuthoringWorkAuthorization rejects widening a preparation or reusing
// a decision for another operator or changed canonical requirements.
func ValidateAuthoringWorkAuthorization(scope AuthoringWorkScope, authority *AuthoringWorkAuthorization) error {
	if err := ValidateAuthoringWorkScope(scope); err != nil {
		return err
	}

	if scope.Intent == AuthoringIntentPrepare {
		if authority != nil {
			return ErrWorkflowConflict
		}

		return nil
	}

	if authority == nil {
		return ErrAuthoringAuthorityRequired
	}

	_, err := AuthorizeAuthoringWork(scope, authority.Binding, authority.ScopeHash)

	return err
}

// ErrAuthoringAuthorityRequired requests one explicit decision for displayed work.
var ErrAuthoringAuthorityRequired = errors.New("the prepared scope requires explicit development authority")

func validateAuthoringWorkScope(scope AuthoringWorkScope) error {
	if err := validateWorkflowMutationBinding(scope.Binding); err != nil {
		return err
	}

	if scope.Intent != AuthoringIntentPrepare && scope.Intent != AuthoringIntentDevelop {
		return errors.New("only preparation or development can describe native work")
	}

	if err := validateAuthoringWorkBounds(scope); err != nil {
		return err
	}

	if err := validateAuthoringWorkPaths(scope); err != nil {
		return err
	}

	for _, digest := range []string{scope.SourceHash, scope.DescriptionHash, scope.ContractHash, scope.RecordHash} {
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != 32 {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func validateAuthoringWorkBounds(scope AuthoringWorkScope) error {
	if strings.TrimSpace(scope.Instruction) == "" || len(scope.Instruction) > 8192 || !utf8.ValidString(scope.Instruction) || scope.ProjectID == "" || len(scope.ProjectID) > 128 || !IsValidOpsNamespace(scope.Namespace) || scope.SessionRevision < 1 {
		return errors.New("invalid bounded authoring work scope")
	}

	return validateAuthoringSessionID(scope.SessionID)
}

func validateAuthoringWorkPaths(scope AuthoringWorkScope) error {
	if scope.SessionRef != filepath.ToSlash(filepath.Join(".chr", "authoring", "sessions", scope.SessionID, "session.md")) || scope.DescriptionRef == "" || len(scope.DescriptionRef) > 1024 {
		return ErrWorkflowConflict
	}

	if !filepath.IsLocal(scope.DescriptionRef) || filepath.Clean(scope.DescriptionRef) != scope.DescriptionRef || (!validAuthoringArtifactRef(scope.Namespace, AuthoringArtifactRef{Path: scope.DescriptionRef, Kind: AuthoringArtifactKindSpec}) && !validAuthoringArtifactRef(scope.Namespace, AuthoringArtifactRef{Path: scope.DescriptionRef, Kind: AuthoringArtifactKindDump})) {
		return ErrWorkflowConflict
	}

	return nil
}
