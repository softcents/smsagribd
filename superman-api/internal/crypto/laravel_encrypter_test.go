package crypto

import (
	"encoding/base64"
	"testing"
)

func TestLaravelEncrypter_RoundTrip(t *testing.T) {
	// 32-byte key
	appKey := "base64:" + base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))

	enc, err := NewLaravelEncrypter(appKey)
	if err != nil {
		t.Fatalf("Failed to create encrypter: %v", err)
	}

	creds := map[string]any{
		"username": "superman_gw_user",
		"password": "superman_secret_password",
		"api_key":  "gw_api_key_12345",
		"priority": float64(1),
	}

	encrypted, err := enc.EncryptArray(creds)
	if err != nil {
		t.Fatalf("EncryptArray failed: %v", err)
	}

	if encrypted == "" {
		t.Fatal("Encrypted string is empty")
	}

	// Decrypt back
	decrypted, err := enc.DecryptArray(encrypted)
	if err != nil {
		t.Fatalf("DecryptArray failed: %v", err)
	}

	if decrypted["username"] != "superman_gw_user" {
		t.Errorf("Expected username 'superman_gw_user', got '%v'", decrypted["username"])
	}
	if decrypted["password"] != "superman_secret_password" {
		t.Errorf("Expected password 'superman_secret_password', got '%v'", decrypted["password"])
	}
	if decrypted["api_key"] != "gw_api_key_12345" {
		t.Errorf("Expected api_key 'gw_api_key_12345', got '%v'", decrypted["api_key"])
	}
}

func TestLaravelEncrypter_TamperDetection(t *testing.T) {
	appKey := "base64:" + base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	enc, err := NewLaravelEncrypter(appKey)
	if err != nil {
		t.Fatalf("Failed to create encrypter: %v", err)
	}

	creds := map[string]any{"key": "value"}
	encrypted, err := enc.EncryptArray(creds)
	if err != nil {
		t.Fatalf("EncryptArray failed: %v", err)
	}

	// Tamper with the encrypted string
	tampered := encrypted[:len(encrypted)-5] + "AAAAA"
	_, err = enc.DecryptArray(tampered)
	if err == nil {
		t.Error("Expected error when decrypting tampered payload, got nil")
	}
}
