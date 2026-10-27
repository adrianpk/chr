// Package project owns registered identity, active selection and registry diagnostics.
package project

import "time"

// Project describes a package boundary value.
type Project struct {
	ID            string
	Name          string
	Path          string
	Enabled       bool
	RegisteredAt  time.Time
	LastIndexedAt *time.Time
}
