package main

import (
	"strings"
	"testing"
)

// TestExternalExceptions protects exact external identities and rejects stale,
// changed, duplicate and unnecessary allowances without a migration bypass.
func TestExternalExceptions(t *testing.T) {
	value := declaration{ID: "internal/source.go::f", Signature: "func() (string, int)", Results: []string{"string", "int"}}
	external := allowance{ID: value.ID, Signature: value.Signature, Reason: "external result contract", Contract: "external.Interface.f"}
	incomplete := external
	incomplete.Contract = ""

	cases := []struct {
		name       string
		values     []declaration
		exceptions []allowance
		want       string
	}{
		{"new violation", []declaration{value}, nil, "unapproved signature"},
		{"external contract", []declaration{value}, []allowance{external}, ""},
		{"moved declaration", []declaration{changedDeclaration(value, "moved", value.Signature)}, []allowance{external}, "stale external exception"},
		{"changed signature", []declaration{changedDeclaration(value, value.ID, "func() (int, int)")}, []allowance{external}, "changed or unnecessary"},
		{"removed declaration", nil, []allowance{external}, "stale external exception"},
		{"resolved signature", []declaration{{ID: value.ID, Signature: value.Signature}}, []allowance{external}, "changed or unnecessary"},
		{"missing contract", []declaration{value}, []allowance{incomplete}, "incomplete external exception"},
		{"duplicate exception", []declaration{value}, []allowance{external, external}, "duplicate external exception"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			failures := checkPolicy(sourceScan{Declarations: item.values}, item.exceptions)
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

// TestSignaturePolicy enforces the permanent five-input/result boundary for
// ordinary production declarations; result semantics remain stricter than arity.
func TestSignaturePolicy(t *testing.T) {
	cases := []struct {
		name       string
		parameters int
		results    []string
		want       string
	}{
		{"five inputs", 5, nil, ""},
		{"six inputs", 6, nil, "unapproved"},
		{"error result", 1, []string{"Value", "error"}, ""},
		{"bool result", 1, []string{"Value", "bool"}, ""},
		{"single result", 1, []string{"Value"}, ""},
		{"unrelated pair", 1, []string{"Value", "string"}, "unapproved"},
		{"triple result", 1, []string{"Value", "bool", "error"}, "unapproved"},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := declaration{ID: "f", Signature: "fixture", Parameters: item.parameters, Results: item.results}
			failures := checkPolicy(sourceScan{Declarations: []declaration{value}}, nil)
			assertPolicyOutcome(t, failures, item.want)
		})
	}
}
