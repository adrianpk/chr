package workflow

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func recordWireFactories() map[string]func() any {
	return map[string]func() any{
		// These record snapshots retain the shapes from cdf69b8822f8e498cbff8914d7d93e68cd78ff15.
		"integration-entry":  func() any { return &IntegrationEntry{} },
		"merger-result":      func() any { return &workflowstate.MergerResult{} },
		"merger-repair-plan": func() any { return &workflowstate.MergerRepairPlan{} },
		"merger-acceptance":  func() any { return &MergerAcceptance{} },
		"guided-session":     func() any { return &authoringcore.Session{} },
		"guided-control":     func() any { return &authoringcore.GuidedControlState{} },
		"workflow-start":     func() any { return &StartRequest{} },
		"invocation":         func() any { return &Invocation{} },
		"merger-input":       func() any { return &workflowstate.MergerInput{} },
		"indexed-ticket":     func() any { return &ticketcore.Indexed{} },
		"invocation-file":    func() any { return &workflowInvocationFile{} },
	}
}

// TestQualityRecordWire preserves exact baseline bytes, omission and promoted-field order.
func TestQualityRecordWire(t *testing.T) {
	for name, factory := range recordWireFactories() {
		for _, mode := range []string{"zero", "empty", "populated"} {
			t.Run(name+"/"+mode, func(t *testing.T) { assertRecordWire(t, name, mode, factory) })
		}
	}
}

func assertRecordWire(t *testing.T, name, mode string, factory func() any) {
	t.Helper()

	data, err := os.ReadFile("testdata/quality-records/" + name + "-" + mode + ".json")
	if err != nil {
		t.Fatal(err)
	}

	record := factory()
	if err := json.Unmarshal(data, record); err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, bytes.TrimSpace(data)) {
		t.Fatalf("baseline bytes changed for %s/%s", name, mode)
	}
}
