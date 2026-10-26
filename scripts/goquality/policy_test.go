package main

import (
	"strings"
	"testing"
)

// TestDebtRatchet protects exact identities and rejects stale or changed allowances.
func TestDebtRatchet(t *testing.T) {
	value := declaration{ID: "internal/source.go::f", Signature: "func() (string, int)", Results: []string{"string", "int"}}
	old := allowance{ID: value.ID, Signature: value.Signature, Reason: "owned baseline debt"}
	external := old
	external.Contract = "external.Interface.f"

	cases := []struct {
		name       string
		values     []declaration
		debt       []allowance
		exceptions []allowance
		want       string
	}{
		{"new violation", []declaration{value}, nil, nil, "unapproved signature"},
		{"unchanged debt", []declaration{value}, []allowance{old}, nil, ""},
		{"moved declaration", []declaration{changedDeclaration(value, "moved", value.Signature)}, []allowance{old}, nil, "stale debt"},
		{"changed signature", []declaration{changedDeclaration(value, value.ID, "func() (int, int)")}, []allowance{old}, nil, "changed or resolved debt"},
		{"removed declaration", nil, []allowance{old}, nil, "stale debt"},
		{"resolved debt", []declaration{{ID: value.ID, Signature: value.Signature}}, []allowance{old}, nil, "changed or resolved debt"},
		{"external contract", []declaration{value}, nil, []allowance{external}, ""},
		{"missing contract", []declaration{value}, nil, []allowance{old}, "incomplete allowance"},
		{"stale exception", nil, nil, []allowance{external}, "stale external exception"},
		{"changed exception", []declaration{changedDeclaration(value, value.ID, "changed")}, nil, []allowance{external}, "changed or unnecessary"},
		{"overlapping lists", []declaration{value}, []allowance{old}, []allowance{external}, "overlap"},
		{"duplicate debt", []declaration{value}, []allowance{old, old}, nil, "duplicate allowance"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			failures := checkPolicy(sourceScan{Declarations: item.values}, policy{MaxParameters: 7, Debt: item.debt}, item.exceptions)
			assertPolicyOutcome(t, failures, item.want)
		})
	}
}

func changedDeclaration(value declaration, id, signature string) declaration {
	value.ID, value.Signature = id, signature

	return value
}

func assertPolicyOutcome(t *testing.T, failures []string, want string) {
	t.Helper()

	if want == "" && len(failures) != 0 || want != "" && !strings.Contains(strings.Join(failures, "\n"), want) {
		t.Fatalf("failures = %v, want %q", failures, want)
	}
}

// TestPolicyCeilings protects all three phases without granting new result debt.
func TestPolicyCeilings(t *testing.T) {
	cases := []struct {
		name       string
		maximum    int
		parameters int
		results    []string
		want       string
	}{
		{"initial boundary", 7, 7, nil, ""},
		{"initial overflow", 7, 8, nil, "unapproved"},
		{"middle boundary", 6, 6, nil, ""},
		{"middle overflow", 6, 7, nil, "unapproved"},
		{"final boundary", 5, 5, nil, ""},
		{"final overflow", 5, 6, nil, "unapproved"},
		{"error result", 7, 1, []string{"Value", "error"}, ""},
		{"bool result", 7, 1, []string{"Value", "bool"}, ""},
		{"unrelated pair", 7, 1, []string{"Value", "string"}, "unapproved"},
		{"triple result", 7, 1, []string{"Value", "bool", "error"}, "unapproved"},
		{"invalid ceiling", 8, 1, nil, "ceiling"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := declaration{ID: "f", Signature: "fixture", Parameters: item.parameters, Results: item.results}
			failures := checkPolicy(sourceScan{Declarations: []declaration{value}}, policy{MaxParameters: item.maximum}, nil)
			assertPolicyOutcome(t, failures, item.want)
		})
	}
}
