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

const maxParameters = 5

type allowanceIndex struct {
	values   map[string]allowance
	failures []string
}

func invalidResults(results []string) bool {
	return len(results) > 2 || len(results) == 2 && results[1] != "error" && results[1] != "bool"
}

func violates(value declaration) bool {
	return value.Parameters > maxParameters || invalidResults(value.Results)
}

func checkPolicy(scan sourceScan, exceptions []allowance) []string {
	exempt := indexAllowances(exceptions)
	failures := exempt.failures

	for _, value := range scan.Declarations {
		failures = append(failures, checkDeclaration(value, exempt.values)...)
	}

	for id := range exempt.values {
		failures = append(failures, "stale external exception: "+id)
	}

	sort.Strings(failures)

	return failures
}

func indexAllowances(values []allowance) allowanceIndex {
	indexed := make(map[string]allowance, len(values))

	var failures []string

	for _, value := range values {
		if value.ID == "" || value.Signature == "" || value.Reason == "" || value.Contract == "" {
			failures = append(failures, "incomplete external exception: "+value.ID)
		}

		if _, exists := indexed[value.ID]; exists {
			failures = append(failures, "duplicate external exception: "+value.ID)
		}

		indexed[value.ID] = value
	}

	return allowanceIndex{values: indexed, failures: failures}
}

func checkDeclaration(value declaration, exempt map[string]allowance) []string {
	exception, hasException := exempt[value.ID]
	delete(exempt, value.ID)

	if hasException {
		if exception.Signature != value.Signature || !violates(value) {
			return []string{"changed or unnecessary external exception: " + value.ID}
		}

		return nil
	}

	if violates(value) {
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
