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

func TestOTPLifecycle_Cooldown_Attempts_Expiry(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	cfg := config.Config{
		JWTSecret:      []byte("01234567890123456789012345678901"),
		FrontendOrigin: "http://localhost:5173",
		TokenTTL:       8 * time.Hour,
	}

	server := New(cfg, db)
	app := server.App()
	ctx := context.Background()

	now := time.Now().UnixNano()
	testPhone := fmt.Sprintf("+88017%08d", now%100000000)

	// Create authenticated customer for calling protected /api/v1/customer/otp endpoints
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('OTP User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("otp_%d@test.com", now), testPhone)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	token, _ := server.tokens.Issue(userID, "OTP User", fmt.Sprintf("otp_%d@test.com", now), time.Now().Add(8*time.Hour))

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM verification_otps WHERE identifier = ?", testPhone)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	authHeader := "Bearer " + token

	// 1. First send: Should succeed
	p1, _ := json.Marshal(map[string]any{"identifier": testPhone, "type": "sms"})
	req1 := httptest.NewRequest("POST", "/api/v1/customer/otp/send", bytes.NewReader(p1))
	req1.Header.Set("Authorization", authHeader)
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := app.Test(req1, fiber.TestConfig{Timeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("First send request failed: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 on first OTP send, got %d", resp1.StatusCode)
	}

	// 2. Immediate second send: Should trigger 60s cooldown (429)
	req2 := httptest.NewRequest("POST", "/api/v1/customer/otp/send", bytes.NewReader(p1))
	req2.Header.Set("Authorization", authHeader)
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(req2, fiber.TestConfig{Timeout: 10 * time.Second})
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429 on immediate second OTP send (cooldown), got %d", resp2.StatusCode)
	}

	// 3. Test verification with wrong code (should increment attempts)
	pWrong, _ := json.Marshal(map[string]any{"identifier": testPhone, "type": "sms", "code": "000000"})
	reqWrong := httptest.NewRequest("POST", "/api/v1/customer/otp/verify", bytes.NewReader(pWrong))
	reqWrong.Header.Set("Authorization", authHeader)
	reqWrong.Header.Set("Content-Type", "application/json")
	respWrong, _ := app.Test(reqWrong, fiber.TestConfig{Timeout: 10 * time.Second})
	if respWrong.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for invalid code, got %d", respWrong.StatusCode)
	}

	// 4. Test max attempts limit (5 attempts max -> 429)
	_, _ = db.ExecContext(ctx, "UPDATE verification_otps SET attempts = 5 WHERE identifier = ?", testPhone)
	reqMax := httptest.NewRequest("POST", "/api/v1/customer/otp/verify", bytes.NewReader(pWrong))
	reqMax.Header.Set("Authorization", authHeader)
	reqMax.Header.Set("Content-Type", "application/json")
	respMax, _ := app.Test(reqMax, fiber.TestConfig{Timeout: 10 * time.Second})
	if respMax.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429 when max attempts reached, got %d", respMax.StatusCode)
	}

	// 5. Test expiry (expired code -> 422)
	_, _ = db.ExecContext(ctx, "UPDATE verification_otps SET attempts = 0, expires_at = DATE_SUB(NOW(), INTERVAL 1 MINUTE) WHERE identifier = ?", testPhone)
	reqExp := httptest.NewRequest("POST", "/api/v1/customer/otp/verify", bytes.NewReader(pWrong))
	reqExp.Header.Set("Authorization", authHeader)
	reqExp.Header.Set("Content-Type", "application/json")
	respExp, _ := app.Test(reqExp, fiber.TestConfig{Timeout: 10 * time.Second})
	if respExp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for expired OTP, got %d", respExp.StatusCode)
	}

	// 6. Test successful verification
	correctCode := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(correctCode), 10)
	_, _ = db.ExecContext(ctx, `
		UPDATE verification_otps 
		SET otp_hash = ?, attempts = 0, expires_at = DATE_ADD(NOW(), INTERVAL 10 MINUTE)
		WHERE identifier = ?
	`, string(hash), testPhone)

	pCorrect, _ := json.Marshal(map[string]any{"identifier": testPhone, "type": "sms", "code": correctCode})
	reqCorrect := httptest.NewRequest("POST", "/api/v1/customer/otp/verify", bytes.NewReader(pCorrect))
	reqCorrect.Header.Set("Authorization", authHeader)
	reqCorrect.Header.Set("Content-Type", "application/json")
	respCorrect, _ := app.Test(reqCorrect, fiber.TestConfig{Timeout: 10 * time.Second})
	if respCorrect.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for correct OTP, got %d", respCorrect.StatusCode)
	}

	var verifiedAt sql.NullTime
	_ = db.QueryRowContext(ctx, "SELECT verified_at FROM verification_otps WHERE identifier = ?", testPhone).Scan(&verifiedAt)
	if !verifiedAt.Valid {
		t.Errorf("Expected verified_at to be populated in DB after successful verification")
	}
}
