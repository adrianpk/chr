package webui

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// Frozen pre-grouping bytes protect field order, nil/empty values and nested
// action/request identities while the presentation view gains section owners.
func TestGuidedViewWire(t *testing.T) {
	for _, mode := range []string{"zero", "empty", "populated"} {
		t.Run(mode, func(t *testing.T) {
			data, err := os.ReadFile("testdata/quality-guided/view-" + mode + ".json")
			if err != nil {
				t.Fatal(err)
			}

			var view GuidedUseCaseView
			if err := json.Unmarshal(data, &view); err != nil {
				t.Fatal(err)
			}

			encoded, err := json.Marshal(view)
			if err != nil || !bytes.Equal(encoded, bytes.TrimSpace(data)) {
				t.Fatalf("guided wire changed: %v\n%s", err, encoded)
			}
		})
	}

	data, err := os.ReadFile("testdata/quality-guided/fields.json")
	if err != nil {
		t.Fatal(err)
	}

	var expected []guidedWireField
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}

	if actual := guidedWireFields(reflect.TypeFor[GuidedUseCaseView]()); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("guided fields changed: %v", actual)
	}
}

type guidedWireField struct{ Name, Type, Tag string }

func guidedWireFields(typ reflect.Type) []guidedWireField {
	var fields []guidedWireField

	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Anonymous {
			fields = append(fields, guidedWireFields(field.Type)...)
		} else {
			fields = append(fields, guidedWireField{Name: field.Name, Type: field.Type.String(), Tag: string(field.Tag)})
		}
	}

	return fields
}
