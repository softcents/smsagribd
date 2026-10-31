package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"

	"smswhiteinfo/customer-api/internal/config"
)

func getAuthTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db
}

func TestAuthSecurity_DisposableEmailAndDuplicatePhone(t *testing.T) {
	db := getAuthTestDB(t)
	defer db.Close()
	ctx := context.Background()

	cfg := config.Config{
		JWTSecret:        []byte("01234567890123456789012345678901"),
		TokenTTL:         time.Hour,
		RememberTokenTTL: 24 * time.Hour,
	}
	server := New(cfg, db)

	app := fiber.New()
	app.Post("/register", server.register)

	// Scenario 1: Attempt registration with a known disposable email domain
	disposableBody, _ := json.Marshal(map[string]string{
		"name":     "Disposable Tester",
		"email":    "attacker@mailinator.com",
		"phone":    "+8801712345678",
		"password": "Password123!",
	})
	req1 := httptest.NewRequest("POST", "/register", bytes.NewReader(disposableBody))
	req1.Header.Set("Content-Type", "application/json")
	testCfg := fiber.TestConfig{Timeout: 15 * time.Second}
	resp1, err := app.Test(req1, testCfg)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	if resp1.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for disposable email domain, got %d", resp1.StatusCode)
	}

	var res1 map[string]any
	_ = json.NewDecoder(resp1.Body).Decode(&res1)
	errorsMap, _ := res1["errors"].(map[string]any)
	emailErrors, _ := errorsMap["email"].([]any)
	if len(emailErrors) == 0 || emailErrors[0] != "The email domain is not allowed. Please use a permanent email address." {
		t.Errorf("Expected disposable email rejection message, got: %v", emailErrors)
	}

	// Scenario 2: Duplicate phone number rejection
	now := time.Now().UnixNano()
	existingPhone := fmt.Sprintf("+88017%08d", now%100000000)
	existingEmail := fmt.Sprintf("existing_%d@example.com", now)
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Existing User', ?, ?, 'hash', 0.00, 'customer', 'active', NOW(), NOW())
	`, existingEmail, existingPhone)
	if err != nil {
		t.Fatalf("Failed to insert existing user: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE email = ?", existingEmail)
	}()

	dupPhoneBody, _ := json.Marshal(map[string]string{
		"name":     "Dup Phone User",
		"email":    fmt.Sprintf("new_%d@example.com", now),
		"phone":    existingPhone,
		"password": "Password123!",
	})
	req2 := httptest.NewRequest("POST", "/register", bytes.NewReader(dupPhoneBody))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, testCfg)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	if resp2.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for duplicate phone, got %d", resp2.StatusCode)
	}
}

func TestAuthSecurity_LoginThrottlingAndLockout(t *testing.T) {
	db := getAuthTestDB(t)
	defer db.Close()
	ctx := context.Background()

	cfg := config.Config{
		JWTSecret:        []byte("01234567890123456789012345678901"),
		TokenTTL:         time.Hour,
		RememberTokenTTL: 24 * time.Hour,
	}
	server := New(cfg, db)

	app := fiber.New()
	app.Post("/login", server.login)

	now := time.Now().UnixNano()
	testEmail := fmt.Sprintf("throttle_user_%d@example.com", now)
	pwHash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), bcrypt.DefaultCost)
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Throttle User', ?, '+8801700000000', ?, 0.00, 'customer', 'active', NOW(), NOW())
	`, testEmail, string(pwHash))
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE email = ?", testEmail)
	}()

	// Perform 4 failed login attempts -> expect 422
	badBody, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": "WrongPassword!",
	})
	testCfg := fiber.TestConfig{Timeout: 15 * time.Second}

	for i := 1; i <= 4; i++ {
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(badBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, testCfg)
		if err != nil {
			t.Fatalf("Attempt %d failed with err: %v", i, err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Attempt %d: expected 422, got %d", i, resp.StatusCode)
		}
	}

	// 5th failed attempt -> locks account for 60 seconds
	req5 := httptest.NewRequest("POST", "/login", bytes.NewReader(badBody))
	req5.Header.Set("Content-Type", "application/json")
	resp5, err := app.Test(req5, testCfg)
	if err != nil {
		t.Fatalf("Attempt 5 failed with err: %v", err)
	}
	if resp5.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("Attempt 5: expected 422 on the hitting attempt, got %d", resp5.StatusCode)
	}

	// 6th attempt (even with CORRECT password) -> must be blocked with HTTP 429 Too Many Requests
	goodBody, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": "CorrectPassword123!",
	})
	req6 := httptest.NewRequest("POST", "/login", bytes.NewReader(goodBody))
	req6.Header.Set("Content-Type", "application/json")
	resp6, err := app.Test(req6, testCfg)
	if err != nil {
		t.Fatalf("Attempt 6 failed with err: %v", err)
	}
	if resp6.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Expected HTTP 429 Too Many Requests after 5 failures, got %d", resp6.StatusCode)
	}

	var res6 map[string]any
	_ = json.NewDecoder(resp6.Body).Decode(&res6)
	msg, _ := res6["message"].(string)
	if msg == "" || !bytes.Contains([]byte(msg), []byte("Too many login attempts")) {
		t.Errorf("Expected lockout message 'Too many login attempts...', got: %v", msg)
	}
}
