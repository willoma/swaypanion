package socket

import (
	"os"
	"path/filepath"
)

const (
	EndOfCommand = 0x05
	ArgSeparator = 0x1f
)

func Path() string {
	socketdir := os.Getenv("XDG_RUNTIME_DIR")

	if socketdir == "" {
		socketdir = os.TempDir()
	}

	return filepath.Join(socketdir, "swaypanion.sock")
}
