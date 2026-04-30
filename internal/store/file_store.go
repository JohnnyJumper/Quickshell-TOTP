package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	path string
}

type accountDocument struct {
	Version  int       `json:"version"`
	Accounts []Account `json:"accounts"`
}

type encryptedDocument struct {
	Version    int    `json:"version"`
	Cipher     string `json:"cipher"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func NewFileStore(path string) FileStore {
	if path == "" {
		path = DefaultPath()
	}

	return FileStore{
		path: path,
	}
}

func (fileStore FileStore) Load() ([]Account, error) {
	bytes, err := os.ReadFile(fileStore.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Account{}, nil
		}

		return nil, fmt.Errorf("read store file: %w", err)
	}

	var encryptedFile encryptedDocument
	if err := json.Unmarshal(bytes, &encryptedFile); err != nil {
		return nil, fmt.Errorf("decode encrypted store document: %w", err)
	}

	plaintext, err := decrypt(encryptedFile)
	if err != nil {
		return nil, err
	}

	var accountFile accountDocument
	if err := json.Unmarshal(plaintext, &accountFile); err != nil {
		return nil, fmt.Errorf("decode account document: %w", err)
	}

	return accountFile.Accounts, nil
}

func (fileStore FileStore) Save(accounts []Account) error {
	accountFile := accountDocument{
		Version:  documentVersion,
		Accounts: accounts,
	}

	plaintext, err := json.MarshalIndent(accountFile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode account document: %w", err)
	}

	encryptedFile, err := encrypt(plaintext)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(encryptedFile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode encrypted store document: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(fileStore.path), 0700); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}

	if err := os.WriteFile(fileStore.path, bytes, 0600); err != nil {
		return fmt.Errorf("write store file: %w", err)
	}

	return nil
}
