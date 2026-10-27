package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// RoundNotes describes persisted feature-round clarification notes.
type RoundNotes struct {
	Scope            quizcore.Scope
	Status           string
	RejectionReason  string
	Prompts          []generationcore.AcceptanceFeaturePrompt
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// RoundPaths describes the canonical feature-round note paths.
type RoundPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// RoundLayout executes the corresponding package operation.
func RoundLayout(projectDir string, scope quizcore.Scope) (RoundPaths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return RoundPaths{}, err
	}

	scopeID, err := quizcore.ScopeID(scope)
	if err != nil {
		return RoundPaths{}, err
	}

	body, err := opscore.NamedArtifactBody(projectDir, "spec", scopeID+"-acceptance-feature-round")
	if err != nil {
		return RoundPaths{}, err
	}

	return RoundPaths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

// WriteRoundNotes executes the corresponding package operation.
func WriteRoundNotes(ctx context.Context, projectDir string, notes RoundNotes) (RoundPaths, error) {
	err := ctx.Err()
	if err != nil {
		return RoundPaths{}, err
	}

	err = ValidateRoundNotes(notes)
	if err != nil {
		return RoundPaths{}, err
	}

	paths, err := RoundLayout(projectDir, notes.Scope)
	if err != nil {
		return RoundPaths{}, err
	}

	scopeID, err := quizcore.ScopeID(notes.Scope)
	if err != nil {
		return RoundPaths{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(projectDir, "spec", scopeID+"-acceptance-feature-round"); err != nil {
		return RoundPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return RoundPaths{}, fmt.Errorf("create acceptance feature round directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderRoundNotes(paths, notes)), 0o644)
	if err != nil {
		return RoundPaths{}, fmt.Errorf("write acceptance feature round notes: %w", err)
	}

	return paths, nil
}

// RenderRoundNotes renders the corresponding package data.
func RenderRoundNotes(paths RoundPaths, notes RoundNotes) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("title: " + opsmetadata.Quote("Acceptance Feature Round") + "\n")
	output.WriteString("status: " + opsmetadata.Quote(notes.Status) + "\n")
	output.WriteString("type: " + opsmetadata.Quote("acceptance_feature_round") + "\n")
	output.WriteString("case: " + opsmetadata.Quote(notes.Scope.CaseSlug) + "\n")
	output.WriteString("source_request: " + opsmetadata.Quote(notes.SourceRequestID) + "\n")
	output.WriteString("source_response: " + opsmetadata.Quote(notes.SourceResponseID) + "\n")
	output.WriteString("updated_at: " + opsmetadata.Quote(notes.CreatedAt.UTC().Format(time.RFC3339)) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Acceptance Feature Round\n\n")
	output.WriteString("## Missing Information\n\n")
	output.WriteString(strings.TrimSpace(notes.RejectionReason) + "\n\n")
	output.WriteString("## Prompts\n\n")

	for _, prompt := range notes.Prompts {
		output.WriteString("### " + prompt.ID + "\n\n")
		output.WriteString("- Question: " + prompt.Question + "\n")
		output.WriteString("- Default: " + prompt.DefaultAnswer + "\n")

		if prompt.Reason != "" {
			output.WriteString("- Reason: " + prompt.Reason + "\n")
		}

		output.WriteString("- Answer:")

		if prompt.Answer != "" {
			output.WriteString(" " + prompt.Answer)
		}

		output.WriteString("\n\n")
	}

	output.WriteString("## Source\n\n")
	output.WriteString("- feature round: " + paths.RelativePath + "\n")

	return output.String()
}

// MergeRoundAnswers copies existing Markdown answers into matching new prompts.
func MergeRoundAnswers(notes RoundNotes, existingContent string) RoundNotes {
	answers := roundAnswers(existingContent)
	if len(answers) == 0 {
		return notes
	}

	merged := notes

	merged.Prompts = make([]generationcore.AcceptanceFeaturePrompt, len(notes.Prompts))
	for index, prompt := range notes.Prompts {
		merged.Prompts[index] = prompt
		if strings.TrimSpace(prompt.Answer) != "" {
			continue
		}

		answer := answers[prompt.ID]
		if answer != "" {
			merged.Prompts[index].Answer = answer
		}
	}

	return merged
}

func roundAnswers(content string) map[string]string {
	answers := map[string]string{}
	lines := strings.Split(content, "\n")
	currentID := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			currentID = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))

			continue
		}

		if currentID == "" || !strings.HasPrefix(trimmed, "- Answer:") {
			continue
		}

		answer := strings.TrimSpace(strings.TrimPrefix(trimmed, "- Answer:"))
		if answer != "" {
			answers[currentID] = answer
		}
	}

	return answers
}

// ValidateRoundNotes validates the corresponding package boundary value.
func ValidateRoundNotes(notes RoundNotes) error {
	err := quizcore.ValidateScope(notes.Scope)
	if err != nil {
		return err
	}

	if strings.TrimSpace(notes.Status) != generationcore.AcceptanceFeatureStatusNeedsMoreInfo {
		return fmt.Errorf("acceptance feature round status must be needs_more_info")
	}

	if strings.TrimSpace(notes.RejectionReason) == "" {
		return fmt.Errorf("acceptance feature round rejection reason is required")
	}

	if len(generationcore.NormalizeAcceptanceFeaturePrompts(notes.Prompts)) == 0 {
		return fmt.Errorf("acceptance feature round prompts are required")
	}

	if strings.TrimSpace(notes.SourceRequestID) == "" {
		return fmt.Errorf("acceptance feature round source request id is required")
	}

	if strings.TrimSpace(notes.SourceResponseID) == "" {
		return fmt.Errorf("acceptance feature round source response id is required")
	}

	if notes.CreatedAt.IsZero() {
		return fmt.Errorf("acceptance feature round created time is required")
	}

	return nil
}
