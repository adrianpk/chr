package openai

import quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

type responsesFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict"`
	Schema      map[string]any `json:"schema"`
}

func responseFormat(schema quizmodel.SchemaDefinition) responsesFormat {
	return responsesFormat{Type: "json_schema", Name: schema.Name, Description: schema.Description, Strict: true, Schema: schema.Schema}
}
