package store

import (
	"os"
	"path/filepath"
)

type Store interface {
	Load() ([]Account, error)
	Save(accounts []Account) error
}

func DefaultPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), "totp", "accounts.enc")
		}

		dataHome = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(dataHome, "totp", "accounts.enc")
}
