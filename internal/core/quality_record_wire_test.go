package core

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func recordWireFactories() map[string]func() any {
	return map[string]func() any{
		"invocation":      func() any { return &WorkflowInvocation{} },
		"merger-input":    func() any { return &WorkflowMergerInput{} },
		"indexed-ticket":  func() any { return &IndexedTicket{} },
		"invocation-file": func() any { return &workflowInvocationFile{} },
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
