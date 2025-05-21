package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const (
	// PATLength specifies the length of the generated PAT string (in bytes).
	// 32 bytes will result in a 64-character hex-encoded string.
	PATLength = 32
)

// generatePAT generates a random, opaque string to be used as a PAT.
// It returns a hex-encoded string of length 2*PATLength.
func generatePAT() (string, error) {
	bytes := make([]byte, PATLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for PAT: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// getEncryptionKey retrieves the AES encryption key from the environment variable.
// The key must be 32 bytes long for AES-256.
func getEncryptionKey() ([]byte, error) {
	keyHex := os.Getenv("PAT_ENCRYPTION_KEY")
	if keyHex == "" {
		return nil, fmt.Errorf("PAT_ENCRYPTION_KEY environment variable not set")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PAT_ENCRYPTION_KEY from hex: %w", err)
	}
	if len(key) != 32 { // AES-256 requires a 32-byte key
		return nil, fmt.Errorf("PAT_ENCRYPTION_KEY must be 32 bytes (64 hex characters) long, got %d bytes", len(key))
	}
	return key, nil
}

// encryptPAT encrypts the PAT using AES-GCM.
// It returns a hex-encoded string of the ciphertext.
func encryptPAT(plainTextPAT string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainTextPAT), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decryptPAT decrypts the PAT using AES-GCM.
// The encryptedPAT is expected to be a hex-encoded string.
func decryptPAT(encryptedPATHex string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	encryptedPAT, err := hex.DecodeString(encryptedPATHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted PAT from hex: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	if len(encryptedPAT) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted PAT is too short")
	}

	nonce, ciphertext := encryptedPAT[:gcm.NonceSize()], encryptedPAT[gcm.NonceSize():]
	plainTextBytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt PAT: %w", err)
	}

	return string(plainTextBytes), nil
}
