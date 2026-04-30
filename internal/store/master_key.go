package store

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

var activeMasterKey []byte

const (
	argon2Time       = uint32(1)
	argon2Memory     = uint32(65536)
	argon2Threads    = uint8(4)
	saltSize         = 32
	masterKeyVersion = 1
)

type masterKeyDocument struct {
	Version int    `json:"version"`
	Key     string `json:"key"`
}

func masterKeyPath() string {
	return filepath.Join(filepath.Dir(DefaultPath()), "master.key")
}

// getMasterKey is called by crypto.go on every encrypt/decrypt.
func getMasterKey() ([]byte, error) {
	if activeMasterKey == nil {
		return nil, errors.New("master key not initialized: is the daemon running?")
	}
	return activeMasterKey, nil
}

// InitMasterKey is called once by daemon.Run() before net.Listen.
// On first run, readPassphrase is called to optionally derive the master key from a user passphrase.
// On subsequent runs, the master key is loaded from disk without any prompting.
func InitMasterKey(readPassphrase func(prompt string) (string, error)) error {
	path := masterKeyPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read master key file: %w", err)
		}
		return createMasterKey(path, readPassphrase)
	}

	var doc masterKeyDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse master key file: %w", err)
	}
	if doc.Version != masterKeyVersion {
		return fmt.Errorf("unsupported master key file version: %d", doc.Version)
	}

	return loadMasterKey(doc)
}

func createMasterKey(path string, readPassphrase func(prompt string) (string, error)) error {
	passphrase, err := readPassphrase("Set a passphrase for your master key (press Enter to generate one randomly): ")
	if err != nil {
		return fmt.Errorf("read passphrase: %w", err)
	}

	var masterKey []byte
	if passphrase == "" {
		masterKey, err = randomBytes(keySize)
		if err != nil {
			return err
		}
	} else {
		salt, err := randomBytes(saltSize)
		if err != nil {
			return err
		}
		// Derive the master key from the passphrase. The salt is discarded after
		// derivation — the derived key itself is what gets stored, so no prompting
		// is needed on subsequent daemon starts.
		masterKey = argon2.IDKey([]byte(passphrase), salt, argon2Time, argon2Memory, argon2Threads, uint32(keySize))
	}

	doc := masterKeyDocument{
		Version: masterKeyVersion,
		Key:     base64.StdEncoding.EncodeToString(masterKey),
	}

	if err := writeMasterKeyFile(path, doc); err != nil {
		return err
	}

	fmt.Printf("master key created at %s\n", path)
	activeMasterKey = masterKey
	return nil
}

// loadMasterKey reads the stored master key from disk. No user interaction.
func loadMasterKey(doc masterKeyDocument) error {
	key, err := base64.StdEncoding.DecodeString(doc.Key)
	if err != nil {
		return fmt.Errorf("decode master key: %w", err)
	}
	if len(key) != keySize {
		return fmt.Errorf("invalid master key size: %d", len(key))
	}
	activeMasterKey = key
	return nil
}

// writeMasterKeyFile writes atomically via temp file + rename to avoid corruption on power failure.
func writeMasterKeyFile(path string, doc masterKeyDocument) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode master key document: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".master.key.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp key file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op if rename succeeded

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("set temp key file permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp key file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp key file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("install master key file: %w", err)
	}
	return nil
}
