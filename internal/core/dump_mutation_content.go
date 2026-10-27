package core

import (
	"context"
	"errors"
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func defaultDumpCapture(request DumpMutationRequest) DumpCapturePreparer {
	return func(_ context.Context, destination, raw string) (DumpCapturePreparation, error) {
		scope := ""
		if request.CaseSlug != "" {
			scope = "case: " + opsmetadata.Quote(request.CaseSlug) + "\n"
		}

		prepared := DumpCapturePreparation{
			Content:    "---\ntitle: " + opsmetadata.Quote(request.Title) + "\ntype: dump\nstatus: curated\nsource: " + opsmetadata.Quote(raw) + "\n" + scope + "---\n\n" + request.Content + "\n",
			RawContent: "---\ntitle: " + opsmetadata.Quote(request.Title+" Raw Capture") + "\ntype: dump_raw\nstatus: raw\ncurated: " + opsmetadata.Quote(destination) + "\n" + scope + "---\n\n" + request.Content + "\n",
		}

		return prepared, nil
	}
}

func validateDumpPreparation(binding dumpMutationBinding, prepared DumpCapturePreparation) error {
	if len(prepared.Content) > workflowstate.FileLimit || len(prepared.RawContent) > workflowstate.FileLimit || strings.TrimSpace(prepared.Content) == "" || strings.TrimSpace(prepared.RawContent) == "" {
		return errors.New("prepared dump components must be nonempty and at most 1 MiB each")
	}

	for _, content := range []string{prepared.Content, prepared.RawContent} {
		if err := validateDumpPreparedScope(binding.Request, content); err != nil {
			return err
		}
	}

	return validateDumpPreparedIdentity(binding, prepared)
}

func validateDumpPreparedScope(request DumpMutationRequest, content string) error {
	fields, err := projectDumpBodyFields([]byte(content))
	if err != nil {
		return err
	}

	scope, err := projectDumpScope(fields["case"], fields["use_case"], fields["case_slug"])
	if err != nil || scope != request.CaseSlug {
		return errors.New("prepared dump scope differs from the capture scope")
	}

	return nil
}

func validateDumpPreparedIdentity(binding dumpMutationBinding, prepared DumpCapturePreparation) error {
	fields, err := projectDumpBodyFields([]byte(prepared.RawContent))
	if err != nil || fields["type"] != "dump_raw" || fields["curated"] != binding.Destination {
		return errors.New("raw dump component requires dump_raw identity")
	}

	fields, err = projectDumpBodyFields([]byte(prepared.Content))
	if err != nil || fields["type"] != "dump" || fields["source"] != binding.RawReference {
		return errors.New("curated dump component must retain its raw capture reference")
	}

	return nil
}

func dumpMutationBody(request DumpMutationRequest, target dumpMutationTarget, prepared *DumpCapturePreparation, before []byte) ([]byte, error) {
	if request.Action == "save" {
		return dumpSavedBody(before, []byte(request.Content))
	}

	contribution := []byte(request.Content)

	if prepared != nil {
		if target.Artifact.Reference == "" {
			return []byte(prepared.Content), nil
		}

		body, err := dumpEditableBody([]byte(prepared.Content))
		if err != nil {
			return nil, err
		}

		contribution = body
	}

	return append(append(append([]byte(nil), before...), []byte("\n\n")...), append(contribution, '\n')...), nil
}

func dumpSavedBody(before, content []byte) ([]byte, error) {
	body, err := dumpEditableBody(content)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(string(before), "---\n") {
		return body, nil
	}

	front, _, err := splitWorkflowMarkdown(before)
	if err != nil {
		return nil, err
	}

	return []byte("---\n" + front + "\n---\n" + string(body)), nil
}

func dumpEditableBody(content []byte) ([]byte, error) {
	if !strings.HasPrefix(string(content), "---\n") {
		return content, nil
	}

	_, body, err := splitWorkflowMarkdown(content)

	return body, err
}
