package chc

import "time"

type Project struct {
	ID            string
	Name          string
	Path          string
	Enabled       bool
	RegisteredAt  time.Time
	LastIndexedAt *time.Time
}
