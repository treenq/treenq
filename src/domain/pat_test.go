package domain

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestGeneratePAT(t *testing.T) {
	pat1, err := generatePAT()
	if err != nil {
		t.Fatalf("generatePAT() failed: %v", err)
	}

	// Test length
	expectedLength := PATLength * 2 // Hex encoding doubles the length
	if len(pat1) != expectedLength {
		t.Errorf("generatePAT() returned string of length %d, want %d", len(pat1), expectedLength)
	}

	// Test hex encoding
	_, err = hex.DecodeString(pat1)
	if err != nil {
		t.Errorf("generatePAT() did not return a valid hex-encoded string: %v", err)
	}

	// Test randomness (subsequent call should be different)
	pat2, err := generatePAT()
	if err != nil {
		t.Fatalf("generatePAT() failed on second call: %v", err)
	}
	if pat1 == pat2 {
		t.Errorf("generatePAT() returned the same string on subsequent calls, indicating lack of randomness")
	}
}

func TestEncryptDecryptPAT(t *testing.T) {
	// Valid key for testing (32 bytes / 64 hex characters)
	testKey := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	originalKeyEnv := os.Getenv("PAT_ENCRYPTION_KEY")
	os.Setenv("PAT_ENCRYPTION_KEY", testKey)
	defer os.Setenv("PAT_ENCRYPTION_KEY", originalKeyEnv) // Restore original env var

	pat, err := generatePAT()
	if err != nil {
		t.Fatalf("generatePAT() failed: %v", err)
	}

	encrypted, err := encryptPAT(pat)
	if err != nil {
		t.Fatalf("encryptPAT() failed: %v", err)
	}

	if pat == encrypted {
		t.Errorf("Encrypted PAT is the same as plain text PAT")
	}

	decrypted, err := decryptPAT(encrypted)
	if err != nil {
		t.Fatalf("decryptPAT() failed: %v", err)
	}

	if decrypted != pat {
		t.Errorf("Decrypted PAT does not match original. Got %q, want %q", decrypted, pat)
	}

	// Test with tampered ciphertext
	// Ensure tamperedEncrypted is actually different from encrypted.
	// If encrypted is all '0's at the end, and we replace with '0000', it might not change.
	// So, if it ends up being the same, try '1111'.
	tamperedSuffix := "0000"
	if len(encrypted) >= 4 && strings.HasSuffix(encrypted, tamperedSuffix) {
		tamperedSuffix = "1111"
	}
	tamperedEncrypted := encrypted[:len(encrypted)-len(tamperedSuffix)] + tamperedSuffix

	if encrypted == tamperedEncrypted && len(encrypted) >= 4 {
		// This case should ideally not be hit if PATs are random enough and suffix logic is good.
		t.Logf("Warning: Tampered ciphertext ended up being the same as original. Encrypted: %s", encrypted)
	}


	_, err = decryptPAT(tamperedEncrypted)
	if err == nil {
		t.Errorf("Expected an error when decrypting tampered data, but got nil. Tampered: %s, Original: %s", tamperedEncrypted, encrypted)
	} else {
		t.Logf("Correctly failed to decrypt tampered data: %v", err) // Log the error for info
	}

	// Test decryption with a wrong key
	wrongTestKey := "1111111111111111111111111111111111111111111111111111111111111111" // Different 32-byte key
	os.Setenv("PAT_ENCRYPTION_KEY", wrongTestKey)

	_, err = decryptPAT(encrypted) // Try to decrypt original ciphertext with wrong key
	if err == nil {
		t.Errorf("Expected an error when decrypting with a wrong key, but got nil")
	} else {
		t.Logf("Correctly failed to decrypt with wrong key: %v", err)
	}
}

func TestGetEncryptionKey(t *testing.T) {
	originalKeyEnv := os.Getenv("PAT_ENCRYPTION_KEY")
	defer func() {
		// Restore original env var, handling case where it might have been unset
		if originalKeyEnv == "" {
			os.Unsetenv("PAT_ENCRYPTION_KEY")
		} else {
			os.Setenv("PAT_ENCRYPTION_KEY", originalKeyEnv)
		}
	}()

	// Test case 1: Key not set
	os.Unsetenv("PAT_ENCRYPTION_KEY")
	_, err := getEncryptionKey()
	if err == nil {
		t.Errorf("Expected error when PAT_ENCRYPTION_KEY is not set, got nil")
	} else if !strings.Contains(err.Error(), "PAT_ENCRYPTION_KEY environment variable not set") {
		t.Errorf("Error message for unset key does not contain expected text. Got: %v", err)
	}


	// Test case 2: Invalid hex
	os.Setenv("PAT_ENCRYPTION_KEY", "this-is-definitely-not-hex-string-and-is-long-enough")
	_, err = getEncryptionKey()
	if err == nil {
		t.Errorf("Expected error for invalid hex PAT_ENCRYPTION_KEY, got nil")
	} else if !strings.Contains(err.Error(), "failed to decode PAT_ENCRYPTION_KEY from hex") {
		t.Errorf("Error message for invalid hex does not contain expected text. Got: %v", err)
	}

	// Test case 3: Incorrect length (e.g., 30 hex chars -> 15 bytes)
	os.Setenv("PAT_ENCRYPTION_KEY", "000102030405060708090a0b0c0d0e") // 15 bytes (30 hex chars)
	_, err = getEncryptionKey()
	if err == nil {
		t.Errorf("Expected error for incorrect length PAT_ENCRYPTION_KEY, got nil")
	} else if !strings.Contains(err.Error(), "must be 32 bytes") {
		t.Errorf("Error message for incorrect length does not contain expected text. Got: %v", err)
	}

	// Test case 4: Incorrect length (e.g., 34 hex chars -> 17 bytes)
	os.Setenv("PAT_ENCRYPTION_KEY", "000102030405060708090a0b0c0d0e0f1011") // 17 bytes (34 hex chars)
	_, err = getEncryptionKey()
	if err == nil {
		t.Errorf("Expected error for incorrect length PAT_ENCRYPTION_KEY (17 bytes), got nil")
	} else if !strings.Contains(err.Error(), "must be 32 bytes") {
		t.Errorf("Error message for incorrect length (17 bytes) does not contain expected text. Got: %v", err)
	}


	// Test case 5: Correct key
	validTestKey := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f" // 32 bytes
	os.Setenv("PAT_ENCRYPTION_KEY", validTestKey)
	keyBytes, err := getEncryptionKey()
	if err != nil {
		t.Errorf("Unexpected error for valid key: %v", err)
	}
	if len(keyBytes) != 32 {
		t.Errorf("Expected key length 32 bytes, got %d", len(keyBytes))
	}
	expectedBytes, _ := hex.DecodeString(validTestKey)
	if !strings.EqualFold(hex.EncodeToString(keyBytes), hex.EncodeToString(expectedBytes)) {
		t.Errorf("Returned key bytes do not match expected bytes. Got %x, want %x", keyBytes, expectedBytes)
	}
}
