package crypto

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestLaravelEncrypterRoundtrip(t *testing.T) {
	appKey := "base64:sx5KF4L3SBkmubxoILg5gPFHlVZ+qFollNm2Yr86nco="
	enc, err := NewLaravelEncrypter(appKey)
	if err != nil {
		t.Fatalf("Failed to create encrypter: %v", err)
	}

	testPayload := []byte("secret_api_key_1234567890")
	encrypted, err := enc.Encrypt(testPayload)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(testPayload) {
		t.Errorf("Decrypted mismatch. Expected %s, got %s", testPayload, decrypted)
	}
}

func TestLaravelEncrypterTamper(t *testing.T) {
	appKey := "base64:sx5KF4L3SBkmubxoILg5gPFHlVZ+qFollNm2Yr86nco="
	enc, err := NewLaravelEncrypter(appKey)
	if err != nil {
		t.Fatalf("Failed to create encrypter: %v", err)
	}

	testPayload := []byte("important_credentials")
	encrypted, err := enc.Encrypt(testPayload)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Tamper with the encrypted string
	tampered := encrypted[:len(encrypted)-5] + "AAAAA"
	_, err = enc.Decrypt(tampered)
	if err == nil {
		t.Errorf("Expected decryption error on tampered payload, got nil")
	}
}

func TestPhpUnserializer(t *testing.T) {
	phpSerialized := []byte(`a:3:{s:8:"username";s:5:"admin";s:8:"password";s:6:"secret";s:7:"enabled";b:1;}`)
	parsed, err := ParsePhpOrJson(phpSerialized)
	if err != nil {
		t.Fatalf("Failed to parse php serialized data: %v", err)
	}

	if parsed["username"] != "admin" {
		t.Errorf("Expected username admin, got %v", parsed["username"])
	}
	if parsed["password"] != "secret" {
		t.Errorf("Expected password secret, got %v", parsed["password"])
	}
	if parsed["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", parsed["enabled"])
	}
}

func TestRealGatewaysDecryptionFromScratchDB(t *testing.T) {
	appKey := "base64:sx5KF4L3SBkmubxoILg5gPFHlVZ+qFollNm2Yr86nco="
	enc, err := NewLaravelEncrypter(appKey)
	if err != nil {
		t.Fatalf("Failed to create encrypter: %v", err)
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch")
	if err != nil {
		t.Skipf("Scratch DB not accessible, skipping live table test: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Scratch DB not running: %v", err)
	}

	rows, err := db.Query("SELECT id, name, driver, credentials FROM gateways WHERE credentials IS NOT NULL")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int64
		var name, driver, credsEnc string
		if err := rows.Scan(&id, &name, &driver, &credsEnc); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		plain, err := enc.Decrypt(credsEnc)
		if err != nil {
			t.Errorf("Gateway ID %d (%s, driver: %s) failed decryption: %v", id, name, driver, err)
			continue
		}

		parsed, err := ParsePhpOrJson(plain)
		if err != nil {
			t.Errorf("Gateway ID %d (%s, driver: %s) failed parsing: %v", id, name, driver, err)
			continue
		}

		if len(parsed) == 0 {
			t.Errorf("Gateway ID %d (%s) decrypted to empty map", id, name)
		}
		count++
	}

	if count == 0 {
		t.Fatalf("No gateways found to decrypt in scratch DB")
	}
}
