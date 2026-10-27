package model

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func planOrderJSONSchema() map[string]any {
	task := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"id", "task_order", "title"}, "properties": map[string]any{"id": map[string]any{"type": "string"}, "task_order": map[string]any{"type": "integer", "minimum": 1}, "title": map[string]any{"type": "string"}}}
	slice := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"id", "slice_order", "title", "tasks"}, "properties": map[string]any{"id": map[string]any{"type": "string"}, "slice_order": map[string]any{"type": "integer", "minimum": 1}, "title": map[string]any{"type": "string"}, "tasks": map[string]any{"type": "array", "minItems": 1, "maxItems": workflowstate.PageLimit, "items": task}}}

	return map[string]any{"type": "array", "minItems": 1, "maxItems": workflowstate.PageLimit, "items": slice}
}
