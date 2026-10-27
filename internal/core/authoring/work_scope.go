package authoring

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkScope binds a proposed native description to canonical inputs.
// It does not authorize admission, alter policy or replace a workflow preview.
type WorkScope struct {
	Schema          string                        `json:"schema"`
	Binding         workflowstate.MutationBinding `json:"binding"`
	Intent          authoringstate.Intent         `json:"intent"`
	Instruction     string                        `json:"instruction"`
	ProjectID       string                        `json:"project_id"`
	Namespace       string                        `json:"namespace"`
	SessionID       string                        `json:"session_id"`
	SessionRef      string                        `json:"session_ref"`
	SessionRevision int64                         `json:"session_revision"`
	SourceHash      string                        `json:"source_hash"`
	DescriptionRef  string                        `json:"description_ref"`
	DescriptionHash string                        `json:"description_hash"`
	ContractHash    string                        `json:"contract_hash"`
	RecordHash      string                        `json:"record_hash"`
	Hash            string                        `json:"scope_hash"`
}

// WorkAuthorization retains a host-validated explicit decision for the
// exact displayed scope. Provider output cannot supply this value.
type WorkAuthorization struct {
	Binding   workflowstate.MutationBinding `json:"binding"`
	ScopeHash string                        `json:"scope_hash"`
}

// SealAuthoringWorkScope computes the immutable scope binding after validation.
func SealAuthoringWorkScope(scope WorkScope) (WorkScope, error) {
	scope.Schema, scope.Hash = "chr-authoring-work/v1", ""
	if err := validateAuthoringWorkScope(scope); err != nil {
		return scope, err
	}

	encoded, err := json.Marshal(scope)
	scope.Hash = contenthash.Sum(encoded)

	return scope, err
}

// ValidateWorkScope checks the original scope and its complete digest.
func ValidateWorkScope(scope WorkScope) error {
	expected, err := SealAuthoringWorkScope(scope)
	if err != nil {
		return err
	}

	if scope.Schema != expected.Schema || scope.Hash != expected.Hash {
		return workflowstate.ErrConflict
	}

	return nil
}

// AuthorizeAuthoringWork binds an explicit caller decision to the observed scope.
// A changed description or source requires a new decision for the changed hash.
func AuthorizeAuthoringWork(scope WorkScope, decision workflowstate.MutationBinding, observed string) (WorkAuthorization, error) {
	if err := ValidateWorkScope(scope); err != nil {
		return WorkAuthorization{}, err
	}

	if err := decision.Validate(); err != nil {
		return WorkAuthorization{}, err
	}

	if scope.Intent != authoringstate.IntentDevelop || decision.Operator != scope.Binding.Operator || observed != scope.Hash {
		return WorkAuthorization{}, workflowstate.ErrConflict
	}

	return WorkAuthorization{Binding: decision, ScopeHash: observed}, nil
}

// ValidateWorkAuthorization rejects widening a preparation or reusing
// a decision for another operator or changed canonical requirements.
func ValidateWorkAuthorization(scope WorkScope, authority *WorkAuthorization) error {
	if err := ValidateWorkScope(scope); err != nil {
		return err
	}

	if scope.Intent == authoringstate.IntentPrepare {
		if authority != nil {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if authority == nil {
		return ErrAuthorityRequired
	}

	_, err := AuthorizeAuthoringWork(scope, authority.Binding, authority.ScopeHash)

	return err
}

// ErrAuthorityRequired requests one explicit decision for displayed work.
var ErrAuthorityRequired = errors.New("the prepared scope requires explicit development authority")

func validateAuthoringWorkScope(scope WorkScope) error {
	if err := scope.Binding.Validate(); err != nil {
		return err
	}

	if scope.Intent != authoringstate.IntentPrepare && scope.Intent != authoringstate.IntentDevelop {
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
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validateAuthoringWorkBounds(scope WorkScope) error {
	if strings.TrimSpace(scope.Instruction) == "" || len(scope.Instruction) > 8192 || !utf8.ValidString(scope.Instruction) || scope.ProjectID == "" || len(scope.ProjectID) > 128 || !opsnamespace.Valid(scope.Namespace) || scope.SessionRevision < 1 {
		return errors.New("invalid bounded authoring work scope")
	}

	return authoringstate.ValidateSessionID(scope.SessionID)
}

func validateAuthoringWorkPaths(scope WorkScope) error {
	if scope.SessionRef != filepath.ToSlash(filepath.Join(".chr", "authoring", "sessions", scope.SessionID, "session.md")) || scope.DescriptionRef == "" || len(scope.DescriptionRef) > 1024 {
		return workflowstate.ErrConflict
	}

	if !filepath.IsLocal(scope.DescriptionRef) || filepath.Clean(scope.DescriptionRef) != scope.DescriptionRef || (!validAuthoringArtifactRef(scope.Namespace, ArtifactRef{Path: scope.DescriptionRef, Kind: ArtifactKindSpec}) && ValidateProjectDumpReference(scope.DescriptionRef) != nil) {
		return workflowstate.ErrConflict
	}

	return nil
}
