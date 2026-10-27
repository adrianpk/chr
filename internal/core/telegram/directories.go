package telegram

import "forge.adrianpk.com/chr/internal/core/database"

// TelegramConfigDirectory locates private receiver configuration and credentials.
type TelegramConfigDirectory interface {
	ConfigDir() string
}

// TelegramStateDirectory locates receiver identity and OS-owned exclusion files.
type TelegramStateDirectory interface {
	StateDir() string
}

// TelegramStoreDirectories supplies only database and retained receiver state paths.
type TelegramStoreDirectories interface {
	database.DataDirectory
	TelegramStateDirectory
}
