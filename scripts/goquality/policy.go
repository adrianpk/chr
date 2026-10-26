package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type allowance struct {
	ID        string `json:"id"`
	Signature string `json:"signature"`
	Reason    string `json:"reason"`
	Contract  string `json:"contract,omitempty"`
}

type policy struct {
	MaxParameters int         `json:"max_parameters"`
	Debt          []allowance `json:"debt"`
}

type allowanceIndex struct {
	values   map[string]allowance
	failures []string
}

func invalidResults(results []string) bool {
	return len(results) > 2 || len(results) == 2 && results[1] != "error" && results[1] != "bool"
}

func violates(value declaration, maximum int) bool {
	return value.Parameters > maximum || invalidResults(value.Results)
}

func checkPolicy(scan sourceScan, rules policy, exceptions []allowance) []string {
	var failures []string
	if rules.MaxParameters < 5 || rules.MaxParameters > 7 {
		failures = append(failures, "parameter ceiling must be five, six or seven")
	}

	debt := indexAllowances(rules.Debt, false)
	exempt := indexAllowances(exceptions, true)

	failures = append(failures, debt.failures...)

	failures = append(failures, exempt.failures...)
	for _, value := range scan.Declarations {
		failures = append(failures, checkDeclaration(value, rules.MaxParameters, debt.values, exempt.values)...)
	}

	for id := range debt.values {
		failures = append(failures, "stale debt: "+id)
	}

	for id := range exempt.values {
		failures = append(failures, "stale external exception: "+id)
	}

	sort.Strings(failures)

	return failures
}

func indexAllowances(values []allowance, external bool) allowanceIndex {
	indexed := make(map[string]allowance, len(values))

	var failures []string

	for _, value := range values {
		if value.ID == "" || value.Signature == "" || value.Reason == "" || external && value.Contract == "" {
			failures = append(failures, "incomplete allowance: "+value.ID)
		}

		if _, exists := indexed[value.ID]; exists {
			failures = append(failures, "duplicate allowance: "+value.ID)
		}

		indexed[value.ID] = value
	}

	return allowanceIndex{values: indexed, failures: failures}
}

func checkDeclaration(value declaration, maximum int, debt, exempt map[string]allowance) []string {
	old, hasDebt := debt[value.ID]
	exception, hasException := exempt[value.ID]
	delete(debt, value.ID)
	delete(exempt, value.ID)

	if hasDebt && hasException {
		return []string{"debt and external exception overlap: " + value.ID}
	}

	if hasException {
		if exception.Signature != value.Signature || !violates(value, 5) {
			return []string{"changed or unnecessary external exception: " + value.ID}
		}

		return nil
	}

	if hasDebt {
		if old.Signature != value.Signature || !violates(value, 5) {
			return []string{"changed or resolved debt must be reconciled explicitly: " + value.ID}
		}

		return nil
	}

	if violates(value, maximum) {
		return []string{fmt.Sprintf("%s: unapproved signature %s (%d parameters, results %v)", value.ID, value.Signature, value.Parameters, value.Results)}
	}

	return nil
}

func readJSON(path string, destination any) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			fmt.Fprintln(os.Stderr, closeErr)
		}
	}()

	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()

	return decoder.Decode(destination)
}
