package tui

import (
	"errors"
	"testing"
)

// Editor completion preserves the observed source revision, rotates changed
// intent, retains input on failure and rejects an older draft's delayed reply.
func TestDumpExternalBinding(t *testing.T) {
	model := NewModel(Options{})
	model.startDumpEdit("ops/default/dump/one/artifact.md", "UNSENT", "original-revision")

	original := model.dumpEdit.OperationID
	binding := model.beginOperation("dump:external")
	model.completeDumpDraftEditor(dumpDraftEditorFinishedMsg{Binding: binding, Result: DumpActionResult{Content: "UNSENT plus external"}})

	if model.dumpEdit.Input.Value() != "UNSENT plus external" || model.dumpEdit.Revision != "original-revision" || model.dumpEdit.OperationID == original {
		t.Fatal("editor replaced its source guard or lost text")
	}

	binding = model.beginOperation("dump:external")
	model.completeDumpDraftEditor(dumpDraftEditorFinishedMsg{Binding: binding, Err: errors.New("editor failed")})

	if model.dumpEdit.Input.Value() != "UNSENT plus external" {
		t.Fatal("editor failure discarded input")
	}

	binding = model.beginOperation("dump:external")
	model.startDumpEdit("ops/default/dump/two/artifact.md", "NEWER_DRAFT", "newer-revision")
	model.completeDumpDraftEditor(dumpDraftEditorFinishedMsg{Binding: binding, Result: DumpActionResult{Content: "obsolete output"}})

	if model.dumpEdit.Input.Value() != "NEWER_DRAFT" {
		t.Fatal("old editor overwrote a newer interaction")
	}
}

// The append form freezes the displayed key/revision before text entry; moving
// catalog selection cannot redirect that already captured write.
func TestDumpAppendTarget(t *testing.T) {
	executor := &fakeDumpActionExecutor{}
	model := NewModel(Options{DumpExecutor: executor, Dumps: []DumpView{
		{Title: "First", RelativePath: "ops/default/dump/one/artifact.md", Revision: "one"},
		{Title: "Second", RelativePath: "ops/default/dump/two/artifact.md", Revision: "two"},
	}})
	model.selectView(ViewDumps)
	model.selectNextDump()
	model.planDumpAction(DumpActionAppend)
	model.textForm.Fields[0].Input.SetValue("FREE_TEXT_APPEND")

	model.selectedDump = 0
	model.submitDumpForm()

	request := executor.request
	if request.Path != "ops/default/dump/two/artifact.md" || request.Revision != "two" || request.Content != "FREE_TEXT_APPEND" {
		t.Fatal("append was rebound", request)
	}
}
