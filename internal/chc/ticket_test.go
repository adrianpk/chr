package chc

import (
	"strings"
	"testing"
	"time"
)

func TestTicketRendererRendersFrontmatterAndBody(t *testing.T) {
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 45, 0, 0, time.UTC)
	renderer := NewTicketRenderer()

	content, err := renderer.RenderTicket(Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Priority:    "high",
		Severity:    "critical",
		Type:        "feature",
		Tags:        []string{"uxui", "onboarding"},
		UseCase:     "onboarding",
		ProjectArea: "frontend",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	want := `---
key: "CHC-1"
title: "Do something"
status: "open"
priority: "high"
severity: "critical"
type: "feature"
tags: ["uxui", "onboarding"]
use_case: "onboarding"
project_area: "frontend"
created_at: "2026-05-18T12:30:00Z"
updated_at: "2026-05-18T13:45:00Z"
---

# Description

# Checklist

# Acceptance

# Comments

# Blocker

# Validation
`

	if content != want {
		t.Fatalf("content =\n%s\nwant =\n%s", content, want)
	}
}

func TestTicketRendererKeepsOptionalFieldsEmpty(t *testing.T) {
	renderer := NewTicketRenderer()

	content, err := renderer.RenderTicket(Ticket{Key: "CHC-1", Title: "Do something"})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	for _, want := range []string{
		`priority: ""`,
		`severity: ""`,
		`type: ""`,
		`tags: []`,
		`use_case: ""`,
		`project_area: ""`,
		`created_at: ""`,
		`updated_at: ""`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}
