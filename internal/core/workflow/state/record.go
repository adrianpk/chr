package state

import (
	"time"
)

// Identity is a scoped durable identity independent of a checkout path.
type Identity struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Namespace string    `json:"namespace"`
	Revision  uint64    `json:"revision"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
