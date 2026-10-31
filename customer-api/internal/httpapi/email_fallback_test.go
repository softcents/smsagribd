package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	"smswhiteinfo/customer-api/internal/config"
)

func TestEmailFallback_ForgotPasswordAndOTP(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	ctx := context.Background()

	// Ensure SMTP is unconfigured (MAIL_HOST empty, matching MAIL_MAILER=log)
	os.Setenv("MAIL_HOST", "")
	os.Setenv("SMTP_HOST", "")
	defer os.Unsetenv("MAIL_HOST")
	defer os.Unsetenv("SMTP_HOST")

	cfg := config.Config{
		JWTSecret:        []byte("01234567890123456789012345678901"),
		TokenTTL:         time.Hour,
		RememberTokenTTL: 24 * time.Hour,
		FrontendOrigin:   "http://localhost:5173",
	}

	server := New(cfg, db)
	app := server.App()

	// Insert test customer
	now := time.Now().UnixNano()
	testEmail := fmt.Sprintf("fallback_%d@test.local", now)
	res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, password, balance, type, status, created_at, updated_at)
		VALUES ('Fallback User', ?, '$2y$12$e8Y5t1hR14fM5bV5i7Y8Z.9qW0e1r2t3y4u5i6o7p8a9s0d1f2g3h', 10.00, 'customer', 'active', NOW(), NOW())
	`, testEmail)
	if err != nil {
		t.Fatalf("Failed to insert test customer: %v", err)
	}
	userID, _ := res.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM password_reset_tokens WHERE email = ?", testEmail)
		_, _ = db.ExecContext(ctx, "DELETE FROM verification_otps WHERE identifier = ?", testEmail)
	}()

	// 1. Test POST /api/v1/customer/forgot-password with unconfigured SMTP
	payloadFP, _ := json.Marshal(map[string]string{"email": testEmail})
	reqFP := httptest.NewRequest("POST", "/api/v1/customer/forgot-password", bytes.NewReader(payloadFP))
	reqFP.Header.Set("Content-Type", "application/json")
	respFP, err := app.Test(reqFP, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Forgot password request failed: %v", err)
	}
	if respFP.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for forgot-password with unconfigured SMTP, got %d", respFP.StatusCode)
	}

	// Verify token row was created in password_reset_tokens
	var tokenCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM password_reset_tokens WHERE email = ?", testEmail).Scan(&tokenCount)
	if tokenCount != 1 {
		t.Errorf("Expected 1 password reset token row, got %d", tokenCount)
	}

	// 2. Test POST /api/v1/customer/otp/send with type="email" and unconfigured SMTP
	token, _ := server.tokens.Issue(userID, "Fallback User", testEmail, time.Now().Add(1*time.Hour))
	payloadOTP, _ := json.Marshal(map[string]string{
		"identifier": testEmail,
		"type":       "email",
	})
	reqOTP := httptest.NewRequest("POST", "/api/v1/customer/otp/send", bytes.NewReader(payloadOTP))
	reqOTP.Header.Set("Content-Type", "application/json")
	reqOTP.Header.Set("Authorization", "Bearer "+token)

	respOTP, err := app.Test(reqOTP, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("OTP send request failed: %v", err)
	}
	if respOTP.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for email OTP send with unconfigured SMTP, got %d", respOTP.StatusCode)
	}

	// Verify verification_otps row was created
	var otpCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM verification_otps WHERE identifier = ? AND type = 'email'", testEmail).Scan(&otpCount)
	if otpCount != 1 {
		t.Errorf("Expected 1 verification OTP row for email, got %d", otpCount)
	}
}
