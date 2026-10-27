package app

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

var telegramDocumentNamePattern = regexp.MustCompile(`[^a-z0-9]+`)

func telegramDocumentName(kind, topic string) string {
	slug := strings.Trim(telegramDocumentNamePattern.ReplaceAllString(strings.ToLower(topic), "-"), "-")
	if len(slug) > 80 {
		slug = strings.Trim(slug[:80], "-")
	}

	if slug == "" {
		return kind + ".md"
	}

	return kind + "-" + slug + ".md"
}

func telegramDocumentForTarget(target telegramcore.OperationTarget, source telegramDocumentSource) tgmessage.Document {
	return tgmessage.Document{ProjectID: target.ProjectID, Namespace: target.Namespace, ProjectBindingDigest: target.ProjectBindingDigest, SourceRef: source.Reference, SourceHash: source.Hash, Filename: source.Filename, MediaType: source.MediaType, Payload: source.Payload}
}

func (r telegramReceiver) authoringDocument(ctx context.Context, request telegramcore.OperationRequest, result core.AuthoringContinuationResult) (*tgmessage.DocumentResponse, error) {
	if result.Question != nil {
		return nil, nil
	}

	owner, err := telegramRetainedAuthoringInput(request)
	if err != nil {
		return nil, err
	}

	if owner.Target != core.AuthoringArtifactKindDump && owner.Target != core.AuthoringArtifactKindSpec && owner.Target != core.AuthoringArtifactKindPlan {
		return nil, nil
	}

	operations, ok := r.operations.(telegramAuthoringService)
	if !ok {
		return nil, ErrRecordUnsupported
	}

	id, err := telegramAuthoringSessionID(result.SessionRef)
	if err != nil {
		return nil, err
	}

	snapshot, err := operations.ObserveTelegramAuthoring(ctx, request.Target, id)
	if err != nil {
		return nil, err
	}

	if snapshot.SourceHash != result.SourceHash {
		return nil, workflowstate.ErrConflict
	}

	return telegramPrimaryAuthoringDocument(request.Target, result, snapshot, owner.Target)
}

func telegramPrimaryAuthoringDocument(target telegramcore.OperationTarget, result core.AuthoringContinuationResult, snapshot core.AuthoringSnapshot, kind core.AuthoringArtifactKind) (*tgmessage.DocumentResponse, error) {
	for index := len(snapshot.Artifacts) - 1; index >= 0; index-- {
		artifact := snapshot.Artifacts[index]
		if artifact.Ref.Kind != kind || artifact.Body == "" {
			continue
		}

		payload := []byte(artifact.Body)

		headline := string(kind) + " ready for " + result.Session.Topic + "."
		switch kind {
		case core.AuthoringArtifactKindSpec:
			headline = "Draft specification ready for " + result.Session.Topic + "."
		case core.AuthoringArtifactKindDump:
			headline = "Your request and details are saved as a dump for " + result.Session.Topic + "."
		}

		return &tgmessage.DocumentResponse{Message: tgmessage.Blocks{Family: "21", InitialText: headline, ContentKind: "markdown", Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete document attached."}, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: artifact.Ref.Path, Hash: artifact.SourceHash, Filename: telegramDocumentName(string(kind), result.Session.Topic), MediaType: "text/markdown", Payload: payload})}, nil
	}

	return nil, workflowstate.ErrConflict
}

func telegramExplanationDocument(target telegramcore.OperationTarget, requestID string, result AuthoringExplanation, source []byte) tgmessage.DocumentResponse {
	text := result.Text
	if len(result.SourceRefs) > 0 {
		text += "\n\nCanonical sources:\n" + strings.Join(result.SourceRefs, "\n")
	}

	payload := []byte(text)

	return tgmessage.DocumentResponse{Message: tgmessage.Blocks{Family: "24", InitialText: "Explanation ready.", ContentKind: "markdown", Content: tgmessage.DocumentExcerpt([]byte(result.Text)), SecondaryText: "Complete explanation attached."}, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: "exports/explanation-" + requestID + ".md", Hash: recordDigest(source), Filename: "explanation.md", MediaType: "text/markdown", Payload: payload})}
}

func telegramRecordDocument(target telegramcore.OperationTarget, command telegramRecordCommand, page RecordPage) (tgmessage.DocumentResponse, error) {
	payload, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		return tgmessage.DocumentResponse{}, err
	}

	var lines []string

	for _, record := range page.Records {
		line := record.Title
		if record.Phase != "" {
			line += " — " + record.Phase
		}

		lines = append(lines, line)
	}

	secondary := "Complete records-page JSON export attached."
	if page.HasMore {
		secondary += fmt.Sprintf(" Next page: use the same list command with {\"offset\":%d}.", page.NextOffset)
	}

	filename := fmt.Sprintf("records-%s-page-%d.json", command.Kind, command.Query.Offset)

	return tgmessage.DocumentResponse{Message: tgmessage.Blocks{Family: "28", InitialText: "Records in " + target.Namespace, ContentKind: "list", Content: strings.Join(lines, "\n"), SecondaryText: secondary}, Document: telegramDocumentForTarget(target, telegramDocumentSource{Reference: "exports/" + filename, Hash: recordDigest(payload), Filename: filename, MediaType: "application/json", Payload: payload})}, nil
}

type telegramDocumentSource struct {
	Reference, Hash, Filename, MediaType string
	Payload                              []byte
}
