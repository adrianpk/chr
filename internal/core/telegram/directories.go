package telegram

import "forge.adrianpk.com/chr/internal/core/database"

// ConfigDirectory locates private receiver configuration and credentials.
type ConfigDirectory interface {
	ConfigDir() string
}

// StateDirectory locates receiver identity and OS-owned exclusion files.
type StateDirectory interface {
	StateDir() string
}

// StoreDirectories supplies only database and retained receiver state paths.
type StoreDirectories interface {
	database.DataDirectory
	StateDirectory
}
