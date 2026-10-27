package authoring

import (
	"encoding/json"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// The provider response boundary keeps structured identities and order. Ticket
// allocation, duplicate identities and unexpected nested fields are rejected.
func TestPlanOrderModelPayload(t *testing.T) {
	cases := []struct {
		name, order string
		valid       bool
	}{
		{"structured order", `[{"id":"slice-1","slice_order":1,"title":"Foundation","tasks":[{"id":"task-t1.1","task_order":1,"title":"Implement foundation"}]}]`, true},
		{"missing tasks", `[{"id":"slice-1","slice_order":1,"title":"Foundation","tasks":[]}]`, false},
		{"allocated ticket", `[{"id":"slice-1","slice_order":1,"title":"Foundation","tasks":[{"id":"task-t1.1","task_order":1,"title":"Implement foundation","ticket_key":"CHR-1"}]}]`, false},
		{"mutable progress", `[{"id":"slice-1","slice_order":1,"title":"Foundation","tasks":[{"id":"task-t1.1","task_order":1,"title":"Implement foundation","status":"done"}]}]`, false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			content := []byte(`{"title":"A plan","body":"Bounded work","status":"planned","order":` + item.order + `}`)

			payload, err := quizmodel.DecodePlanPayload(content)
			if (err == nil) != item.valid {
				t.Fatal(payload, err)
			}

			if !item.valid {
				return
			}

			request := quizmodel.Request{Action: quizmodel.ActionGeneratePlan, Scope: quizcore.Scope{ProjectID: "example"}}

			response, err := quizmodel.ResponseFromOutput(request, content, "openai")
			if err != nil || len(response.Plan.Order) != 1 || response.Plan.Order[0].Tasks[0].ID != "task-t1.1" {
				t.Fatal(response, err)
			}
		})
	}
}

// Captured body-only v1 output remains readable, while new provider schemas
// and output contracts require structured order for generation.
func TestPlanOrderModelCompatibility(t *testing.T) {
	payload, err := quizmodel.DecodePlanPayload([]byte(`{"title":"Legacy","body":"Retained historical body","status":"planned"}`))
	if err != nil || payload.Order != nil {
		t.Fatal(payload, err)
	}

	encoded, err := json.Marshal(quizmodel.OutputSchema(quizmodel.Request{Action: quizmodel.ActionGeneratePlan}).Schema)
	if err != nil {
		t.Fatal(err)
	}

	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(encoded, &schema); err != nil {
		t.Fatal(err)
	}

	if len(schema.Required) != 4 || schema.Required[3] != "order" {
		t.Fatal("provider schema omitted map", schema)
	}

	contract := quizmodel.NewOutputContract(quizmodel.ActionGeneratePlan, quizcore.ModeDeep)
	if len(contract.RequiredFields) != 4 || contract.RequiredFields[3] != "order" {
		t.Fatal("shared generation contract omitted map", contract)
	}
}
