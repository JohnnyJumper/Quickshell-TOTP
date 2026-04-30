package store

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	documentVersion = 1

	cipherName = "xchacha20poly1305"

	keySize = chacha20poly1305.KeySize
)

func encrypt(plaintext []byte) (encryptedDocument, error) {
	key, err := getMasterKey()
	if err != nil {
		return encryptedDocument{}, err
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return encryptedDocument{}, fmt.Errorf("create cipher: %w", err)
	}

	nonce, err := randomBytes(aead.NonceSize())
	if err != nil {
		return encryptedDocument{}, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return encryptedDocument{
		Version:    documentVersion,
		Cipher:     cipherName,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func decrypt(document encryptedDocument) ([]byte, error) {
	if document.Cipher != cipherName {
		return nil, fmt.Errorf("unsupported cipher: %s", document.Cipher)
	}

	nonce, err := base64.StdEncoding.DecodeString(document.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(document.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	key, err := getMasterKey()
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt store: %w", err)
	}

	return plaintext, nil
}

func randomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)

	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}

	return bytes, nil
}
