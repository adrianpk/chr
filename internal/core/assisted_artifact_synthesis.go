package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// AssistedArtifactSynthesisOptions describes a package boundary value.
type AssistedArtifactSynthesisOptions struct {
	ProjectDir     string
	OperationID    string
	Session        authoringcore.Session
	CuratedSummary AssistedCuratedAnswerSummary
	SourceRef      authoringcore.ArtifactRef
	TargetKind     authoringcore.ArtifactKind
	TargetName     string
	AllowReplace   bool
	Repository     AuthoringSessionRepository
	Now            func() time.Time
}

// AssistedArtifactSynthesisResult describes a package boundary value.
type AssistedArtifactSynthesisResult struct {
	Session authoringcore.Session
	Layout  AuthoringArtifactLayout
	Content string
	// Mutation retains committed dump hashes independently of later edits.
	Mutation *ProjectDumpMutationResult
}

// SynthesizeAssistedArtifact executes the corresponding package operation.
func synthesizeAssistedArtifact(ctx context.Context, opts AssistedArtifactSynthesisOptions) (AssistedArtifactSynthesisResult, error) {
	err := validateAssistedArtifactSynthesisOptions(opts)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	if opts.TargetKind == authoringcore.ArtifactKindDump {
		return synthesizeAssistedDump(ctx, opts)
	}

	layout, err := PlanAuthoringArtifactLayout(AuthoringArtifactLayoutOptions{
		ProjectDir:   opts.ProjectDir,
		SessionID:    opts.Session.ID,
		Kind:         opts.TargetKind,
		Role:         authoringcore.ArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Name:         opts.TargetName,
		Description:  "Synthesized assisted authoring artifact.",
		AllowReplace: opts.AllowReplace,
	})
	if err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	if layout.Collision.Exists && !layout.Collision.Allowed {
		return AssistedArtifactSynthesisResult{}, fmt.Errorf("assisted artifact already exists: %s", layout.FullPath)
	}

	content := RenderAssistedSynthesizedArtifact(opts.Session, opts.CuratedSummary, opts.SourceRef, layout.Ref)

	if err := EnsureAuthoringArtifactMetadata(opts.ProjectDir, layout.Ref); err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	err = os.MkdirAll(layout.Dir, 0o755)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, fmt.Errorf("create assisted artifact directory: %w", err)
	}

	err = os.WriteFile(layout.FullPath, []byte(content), 0o644)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, fmt.Errorf("write assisted artifact: %w", err)
	}

	session, err := updateAssistedArtifactSynthesisSession(ctx, opts, layout.Ref)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	return AssistedArtifactSynthesisResult{
		Session: session,
		Layout:  layout,
		Content: content,
	}, nil
}

// RenderAssistedSynthesizedArtifact renders the corresponding package data.
func RenderAssistedSynthesizedArtifact(
	session authoringcore.Session,
	summary AssistedCuratedAnswerSummary,
	sourceRef authoringcore.ArtifactRef,
	targetRef authoringcore.ArtifactRef,
) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("session_id: " + opsmetadata.Quote(session.ID) + "\n")
	output.WriteString("source_artifact: " + opsmetadata.Quote(sourceRef.Path) + "\n")
	output.WriteString("target_artifact: " + opsmetadata.Quote(targetRef.Path) + "\n")
	output.WriteString("target_kind: " + opsmetadata.Quote(string(targetRef.Kind)) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# ")
	output.WriteString(session.Topic)
	output.WriteString("\n\n")
	output.WriteString("## Summary\n\n")
	output.WriteString(emptyMarkdownValue(summary.ModelSummary, "No curated summary was available."))
	output.WriteString("\n\n")
	writeSynthesizedAnswers(&output, summary.Answers)
	writeSynthesizedDecisions(&output, summary.Decisions)
	writeSynthesizedOpenQuestions(&output, summary.OpenQuestions)
	output.WriteString("## Provenance\n\n")
	output.WriteString("- source: ")
	output.WriteString(sourceRef.Path)
	output.WriteString("\n")
	output.WriteString("- session: ")
	output.WriteString(session.ID)
	output.WriteString("\n")

	return output.String()
}

func validateAssistedArtifactSynthesisOptions(opts AssistedArtifactSynthesisOptions) error {
	err := ValidateAuthoringSession(opts.Session)
	if err != nil {
		return err
	}

	if strings.TrimSpace(opts.ProjectDir) == "" {
		return fmt.Errorf("assisted artifact synthesis project dir is required")
	}

	if opts.CuratedSummary.SessionID != opts.Session.ID {
		return fmt.Errorf("assisted artifact synthesis summary session id does not match session")
	}

	if err := ValidateAuthoringArtifactRef(opts.SourceRef); err != nil {
		return err
	}

	if !opts.TargetKind.Valid() {
		return fmt.Errorf("assisted artifact synthesis target kind is invalid: %s", opts.TargetKind)
	}

	if opts.TargetKind == authoringcore.ArtifactKindSession || opts.TargetKind == authoringcore.ArtifactKindTranscript {
		return fmt.Errorf("assisted artifact synthesis target kind is session-owned: %s", opts.TargetKind)
	}

	if opts.Repository == nil {
		return fmt.Errorf("assisted artifact synthesis repository is required")
	}

	return nil
}

func updateAssistedArtifactSynthesisSession(
	ctx context.Context,
	opts AssistedArtifactSynthesisOptions,
	targetRef authoringcore.ArtifactRef,
) (authoringcore.Session, error) {
	now := assistedArtifactSynthesisNow(opts.Now)
	session := opts.Session

	session.Outputs = append(session.Outputs, targetRef)
	session.Provenance = append(session.Provenance, authoringcore.ProvenanceLink{
		Relation:   authoringcore.ProvenanceRelationGenerated,
		Source:     opts.SourceRef,
		Target:     targetRef,
		Decision:   "Synthesized assisted authoring artifact from curated summary.",
		RecordedAt: now,
	})
	session.UpdatedAt = now

	_, err := opts.Repository.UpdateSession(ctx, session)
	if err != nil {
		return authoringcore.Session{}, err
	}

	return session, nil
}

func writeSynthesizedAnswers(output *strings.Builder, answers []AssistedCuratedAnswer) {
	output.WriteString("## Captured Answers\n\n")

	if len(answers) == 0 {
		output.WriteString("No captured answers were available.\n\n")

		return
	}

	for _, answer := range answers {
		output.WriteString("- ")
		output.WriteString(answer.Question)
		output.WriteString(": ")
		output.WriteString(answer.Answer)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeSynthesizedDecisions(output *strings.Builder, decisions []string) {
	output.WriteString("## Decisions\n\n")

	if len(decisions) == 0 {
		output.WriteString("No explicit decisions were available.\n\n")

		return
	}

	for _, decision := range decisions {
		output.WriteString("- ")
		output.WriteString(decision)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeSynthesizedOpenQuestions(output *strings.Builder, questions []AssistedCuratedOpenQuestion) {
	output.WriteString("## Open Questions\n\n")

	if len(questions) == 0 {
		output.WriteString("No open questions remain.\n\n")

		return
	}

	for _, question := range questions {
		output.WriteString("- ")
		output.WriteString(question.Question)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func assistedArtifactSynthesisNow(now func() time.Time) time.Time {
	if now == nil {
		return time.Now().UTC()
	}

	return now().UTC()
}
