package app

import "io"

// IO binds the application's process streams at assembly time.
type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// commandFactory owns typed command dependencies before any action executes.
type commandFactory struct {
	io   IO
	deps appDependencies
}
