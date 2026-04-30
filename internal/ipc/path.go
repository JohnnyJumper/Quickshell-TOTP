package ipc

import (
	"os"
	"path/filepath"
)

func DefaultSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = os.TempDir()
	}

	return filepath.Join(runtimeDir, "totp", "totp.sock")
}
