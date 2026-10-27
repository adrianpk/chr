package main

import (
	"strings"
	"testing"
)

type leafDependencyCase struct {
	name, path, source string
	blocked            bool
}

func leafDependencyCases() []leafDependencyCase {
	return []leafDependencyCase{
		{"authoring coordinator", "internal/core/authoring/session_value.go", "package authoring\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"authoring namespace", "internal/core/authoring/session_check.go", "package authoring\nimport _ \"" + coreImportPath + "ops/namespace\"\n", false},
		{"namespace authoring cycle", "internal/core/ops/namespace/namespace.go", "package namespace\nimport _ \"" + coreImportPath + "authoring\"\n", true},
		{"authoring retained inputs", "internal/core/authoring/guided_value.go", "package authoring\nimport _ \"" + coreImportPath + "workflow/state\"\n", false},
		{"quiz authoring cycle", "internal/core/quiz/runtime.go", "package quiz\nimport _ \"" + coreImportPath + "authoring\"\n", true},
		{"acceptance coordinator", "internal/core/authoring/acceptance/feature.go", "package acceptance\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"acceptance payload", "internal/core/authoring/acceptance/feature.go", "package acceptance\nimport _ \"" + coreImportPath + "authoring/generation\"\n", false},
		{"metadata coordinator", "internal/core/ops/metadata/scalar.go", "package metadata\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"file coordinator", "internal/core/ops/file/io.go", "package file\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"file bounds", "internal/core/ops/file/io.go", "package file\nimport _ \"" + coreImportPath + "workflow/state\"\n", false},
		{"artifact coordinator", "internal/core/ops/artifact.go", "package ops\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"artifact project", "internal/core/ops/layout.go", "package ops\nimport _ \"" + coreImportPath + "project\"\n", false},
		{"quiz domain coordinator", "internal/core/quiz/runtime.go", "package quiz\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"quiz domain bounds", "internal/core/quiz/curation.go", "package quiz\nimport _ \"" + coreImportPath + "workflow/state\"\n", false},
		{"generation coordinator", "internal/core/authoring/generation/value.go", "package generation\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"generation scope", "internal/core/authoring/generation/value.go", "package generation\nimport _ \"" + coreImportPath + "quiz\"\n", false},
		{"model coordinator", "internal/core/quiz/model/response.go", "package model\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"model contract", "internal/core/quiz/model/response.go", "package model\nimport _ \"" + coreImportPath + "authoring/generation\"\n", false},
		{"model provider", "internal/core/quiz/model/response.go", "package model\nimport _ \"" + coreImportPath + "quiz/openai\"\n", true},
		{"OpenAI coordinator", "internal/core/quiz/openai/adapter.go", "package openai\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"OpenAI contract", "internal/core/quiz/openai/adapter.go", "package openai\nimport _ \"" + coreImportPath + "quiz/model\"\n", false},
		{"receipt digest", "internal/core/workflow/state/mutation.go", "package state\nimport _ \"" + coreImportPath + "contenthash\"\n", false},
		{"native frontmatter", "internal/core/workflow/state/markdown.go", "package state\nimport _ \"" + coreImportPath + "ops/metadata\"\n", false},
		{"metadata state cycle", "internal/core/ops/metadata/markdown.go", "package metadata\nimport _ \"" + coreImportPath + "workflow/state\"\n", true},
		{"harness identity capability", "internal/core/harness/binding.go", "package harness\nimport _ \"" + moduleImportPath + "/internal/harness\"\n", false},
		{"harness identity coordinator", "internal/core/harness/binding.go", "package harness\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"retained harness binding", "internal/core/workflow/state/admission.go", "package state\nimport _ \"" + coreImportPath + "harness\"\n", false},
		{"harness state cycle", "internal/core/harness/binding.go", "package harness\nimport _ \"" + coreImportPath + "workflow/state\"\n", true},
		{"state coordinator", "internal/core/workflow/state/mutation.go", "package state\nimport _ \"" + coreImportPath + "workflow\"\n", true},
		{"shared values", "internal/core/notify/store.go", "package notify\nimport db \"" + coreImportPath + "database\"\n", false},
		{"shared model", "internal/core/model/value.go", "package value\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"shared authoring/state", "internal/core/authoring/state/value.go", "package value\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"shared workspace", "internal/core/workspace/value.go", "package value\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"root", "internal/core/notify/store.go", "package notify\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"coordinator", "internal/core/gate/policy.go", "package gate\nimport w \"" + coreImportPath + "workflow\"\n", true},
		{"adapter", "internal/core/project/registry.go", "package project\nimport . \"" + moduleImportPath + "/internal/app\"\n", true},
		{"descendant", "internal/core/notify/delivery/store.go", "package delivery\nimport _ \"" + moduleImportPath + "/internal/app\"\n", true},
		{"generated", "internal/core/gate/generated.go", "// Code generated by fixture. DO NOT EDIT.\npackage gate\nimport _ \"" + coreImportPath + "notify\"\n", true},
		{"build variant", "internal/core/notify/variant.go", "//go:build fixture_only\n\npackage notify\nimport _ \"" + coreImportPath + "workflow\"\n", true},
		{"receiver coordinator", "internal/core/telegram/receiver.go", "package telegram\nimport _ \"" + moduleImportPath + "/internal/core\"\n", true},
		{"receiver host", "internal/core/telegram/delivery/store.go", "package telegram\nimport _ \"" + coreImportPath + "workflow\"\n", true},
		{"receiver transport", "internal/core/telegram/receiver.go", "package telegram\nimport _ \"" + coreImportPath + "notify\"\n", false},
		{"pure payload", "internal/core/telegram/message/message.go", "package message\nimport _ \"" + coreImportPath + "notify\"\n", true},
		{"test fixture", "internal/core/notify/store_test.go", "package notify_test\nimport _ \"" + moduleImportPath + "/internal/core\"\n", false},
		{"external", "internal/core/database/driver.go", "package database\nimport _ \"modernc.org/sqlite\"\n", false},
		{"assembly", "internal/app/app.go", "package app\nimport _ \"" + moduleImportPath + "/internal/core\"\n", false},
		{"similar sibling", "internal/core/notifyx/store.go", "package notifyx\nimport _ \"" + moduleImportPath + "/internal/app\"\n", false},
	}
}

// Production syntax checks include aliases, generated files, build variants and
// descendants, while preserving external dependencies and test-only fixtures.
func TestLeafDependencies(t *testing.T) {
	for _, test := range leafDependencyCases() {
		t.Run(test.name, func(t *testing.T) {
			scan := scanFixture(t, map[string]string{test.path: test.source})

			failures := checkLeafDependencies(scan.Imports)
			if (len(failures) != 0) != test.blocked {
				t.Fatalf("dependency failures = %v, blocked = %v", failures, test.blocked)
			}

			if test.blocked && (len(failures) != 1 || !strings.Contains(failures[0], test.path+":")) {
				t.Fatalf("missing source evidence: %v", failures)
			}
		})
	}
}
