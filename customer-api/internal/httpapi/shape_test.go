package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"smswhiteinfo/customer-api/internal/config"
)

func setupTestApp(t *testing.T) (*Server, string, int64, func()) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	cfg := config.Config{
		JWTSecret:        []byte("01234567890123456789012345678901"),
		FrontendOrigin:   "http://localhost:5173",
		TokenTTL:         8 * time.Hour,
		RememberTokenTTL: 720 * time.Hour,
	}

	server := New(cfg, db)

	// Create test customer
	now := time.Now().UnixNano()
	cRes, err := db.Exec(`
		INSERT INTO users (name, email, phone, password, balance, type, status, sms_auto_approved, created_at, updated_at)
		VALUES ('Shape Test User', ?, ?, 'dummy_hash', 100.00, 'customer', 'active', 1, NOW(), NOW())
	`, fmt.Sprintf("shape_%d@example.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	userID, _ := cRes.LastInsertId()

	// Create test token in access_tokens
	rawSecret := "test_token_secret_123"
	hash := sha256.Sum256([]byte(rawSecret))
	hashedToken := hex.EncodeToString(hash[:])

	tRes, err := db.Exec(`
		INSERT INTO access_tokens (tokenable_type, tokenable_id, name, token, abilities, created_at, updated_at)
		VALUES ('App\\Models\\User', ?, 'test_api_token', ?, '["*"]', NOW(), NOW())
	`, userID, hashedToken)
	if err != nil {
		t.Fatalf("Failed to insert test token: %v", err)
	}
	tokenID, _ := tRes.LastInsertId()
	bearer := fmt.Sprintf("%d|%s", tokenID, rawSecret)

	// Create sender for this user
	sRes, err := db.Exec(`
		INSERT INTO senders (sender_id, type, is_enabled, created_at, updated_at)
		VALUES ('TEST_CLI', 'non_masking', 1, NOW(), NOW())
	`)
	var senderDBID int64
	if err == nil {
		senderDBID, _ = sRes.LastInsertId()
		_, _ = db.Exec("INSERT INTO sender_user (sender_id, user_id) VALUES (?, ?)", senderDBID, userID)
	}

	teardown := func() {
		_, _ = db.Exec("DELETE FROM sender_user WHERE user_id = ?", userID)
		if senderDBID > 0 {
			_, _ = db.Exec("DELETE FROM senders WHERE id = ?", senderDBID)
		}
		_, _ = db.Exec("DELETE FROM access_tokens WHERE id = ?", tokenID)
		_, _ = db.Exec("DELETE FROM transactions WHERE user_id = ?", userID)
		_, _ = db.Exec("DELETE FROM batches WHERE user_id = ?", userID)
		_, _ = db.Exec("DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = ?", userID)
		db.Close()
	}

	return server, bearer, userID, teardown
}

func TestExternalApiShapes(t *testing.T) {
	server, bearer, userID, teardown := setupTestApp(t)
	defer teardown()

	app := server.App()

	// 1. GET /api/v1/me
	reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+bearer)
	respMe, err := app.Test(reqMe)
	if err != nil {
		t.Fatalf("GET /api/v1/me failed: %v", err)
	}
	if respMe.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/me expected 200, got %d", respMe.StatusCode)
	}
	var meData map[string]any
	bodyMe, _ := io.ReadAll(respMe.Body)
	_ = json.Unmarshal(bodyMe, &meData)

	for _, expectedKey := range []string{"id", "name", "email", "phone", "balance", "approval"} {
		if _, ok := meData[expectedKey]; !ok {
			t.Errorf("GET /api/v1/me response shape missing key: %s (got: %s)", expectedKey, string(bodyMe))
		}
	}

	// 2. GET /api/v1/balance
	reqBal := httptest.NewRequest("GET", "/api/v1/balance", nil)
	reqBal.Header.Set("Authorization", "Bearer "+bearer)
	respBal, err := app.Test(reqBal)
	if err != nil || respBal.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/balance failed")
	}
	var balData map[string]any
	bodyBal, _ := io.ReadAll(respBal.Body)
	_ = json.Unmarshal(bodyBal, &balData)
	if _, ok := balData["balance"]; !ok {
		t.Errorf("GET /api/v1/balance missing 'balance'")
	}
	if balData["currency"] != "BDT" {
		t.Errorf("GET /api/v1/balance expected currency 'BDT', got %v", balData["currency"])
	}

	// 3. GET /api/v1/senders
	reqSenders := httptest.NewRequest("GET", "/api/v1/senders", nil)
	reqSenders.Header.Set("Authorization", "Bearer "+bearer)
	respSenders, err := app.Test(reqSenders)
	if err != nil || respSenders.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/senders failed")
	}
	var sendersData map[string]any
	bodySenders, _ := io.ReadAll(respSenders.Body)
	_ = json.Unmarshal(bodySenders, &sendersData)
	if _, ok := sendersData["data"]; !ok {
		t.Errorf("GET /api/v1/senders missing 'data' array")
	}

	// 4. POST /api/v1/messages (Send batch)
	sendPayload := map[string]any{
		"sender_id": "TEST_CLI",
		"numbers":   "01700000001,01700000002",
		"body":      "Hello from API test",
	}
	sendBytes, _ := json.Marshal(sendPayload)
	reqSend := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(sendBytes))
	reqSend.Header.Set("Authorization", "Bearer "+bearer)
	reqSend.Header.Set("Content-Type", "application/json")
	respSend, err := app.Test(reqSend)
	if err != nil {
		t.Fatalf("POST /api/v1/messages failed: %v", err)
	}
	if respSend.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(respSend.Body)
		t.Fatalf("POST /api/v1/messages expected 202 Accepted, got %d: %s", respSend.StatusCode, string(b))
	}
	var sendResp map[string]any
	bodySend, _ := io.ReadAll(respSend.Body)
	_ = json.Unmarshal(bodySend, &sendResp)

	for _, expectedKey := range []string{"batch_id", "status", "approval", "recipients"} {
		if _, ok := sendResp[expectedKey]; !ok {
			t.Errorf("POST /api/v1/messages missing key: %s (got %s)", expectedKey, string(bodySend))
		}
	}

	// 5. GET /api/v1/batches/:batch
	batchID := fmt.Sprintf("%v", sendResp["batch_id"])
	reqBatch := httptest.NewRequest("GET", "/api/v1/batches/"+batchID, nil)
	reqBatch.Header.Set("Authorization", "Bearer "+bearer)
	respBatch, err := app.Test(reqBatch)
	if err != nil || respBatch.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/batches/:id failed")
	}
	var batchData map[string]any
	bodyBatch, _ := io.ReadAll(respBatch.Body)
	_ = json.Unmarshal(bodyBatch, &batchData)
	for _, expectedKey := range []string{"id", "sender_id", "status", "approval", "created_at"} {
		if _, ok := batchData[expectedKey]; !ok {
			t.Errorf("GET /api/v1/batches/:id missing key: %s", expectedKey)
		}
	}

	// 6. POST /api/v1/messages with insufficient balance
	// Set balance to 0
	_, _ = server.db.ExecContext(context.Background(), "UPDATE users SET balance = 0 WHERE id = ?", userID)
	reqNoBal := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(sendBytes))
	reqNoBal.Header.Set("Authorization", "Bearer "+bearer)
	reqNoBal.Header.Set("Content-Type", "application/json")
	respNoBal, _ := app.Test(reqNoBal)
	if respNoBal.StatusCode != http.StatusPaymentRequired {
		t.Errorf("Expected 402 Payment Required for insufficient balance, got %d", respNoBal.StatusCode)
	}
	var noBalData map[string]any
	bNoBal, _ := io.ReadAll(respNoBal.Body)
	_ = json.Unmarshal(bNoBal, &noBalData)
	if noBalData["message"] != "Insufficient balance for this batch." {
		t.Errorf("Expected 'Insufficient balance for this batch.', got %v", noBalData["message"])
	}

	// 7. POST /api/v1/messages with unavailable sender
	badSenderPayload := map[string]any{
		"sender_id": "UNKNOWN_SENDER_123",
		"numbers":   "01700000001",
		"body":      "Test",
	}
	bBad, _ := json.Marshal(badSenderPayload)
	reqBad := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(bBad))
	reqBad.Header.Set("Authorization", "Bearer "+bearer)
	reqBad.Header.Set("Content-Type", "application/json")
	respBad, _ := app.Test(reqBad)
	if respBad.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for unavailable sender, got %d", respBad.StatusCode)
	}

	// 8. POST /api/v1/messages with missing/empty numbers (422)
	emptyNumPayload := map[string]any{
		"sender_id": "TEST_CLI",
		"numbers":   "",
		"body":      "Test empty numbers",
	}
	bInv, _ := json.Marshal(emptyNumPayload)
	reqInv := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(bInv))
	reqInv.Header.Set("Authorization", "Bearer "+bearer)
	reqInv.Header.Set("Content-Type", "application/json")
	respInv, _ := app.Test(reqInv)
	if respInv.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for empty phone numbers, got %d", respInv.StatusCode)
	}
	var emptyNumResp map[string]any
	bEmpty, _ := io.ReadAll(respInv.Body)
	_ = json.Unmarshal(bEmpty, &emptyNumResp)
	if emptyNumResp["message"] != "At least one phone number is required." {
		t.Errorf("Expected 'At least one phone number is required.', got %v", emptyNumResp["message"])
	}

	// 9. GET /api/v1/messages with URL query parameters (Laravel legacy parity)
	// Reset balance so message send succeeds
	_, _ = server.db.ExecContext(context.Background(), "UPDATE users SET balance = 50.00 WHERE id = ?", userID)
	reqGetSend := httptest.NewRequest("GET", "/api/v1/messages?sender_id=TEST_CLI&numbers=01700000003&body=GetParamTest", nil)
	reqGetSend.Header.Set("Authorization", "Bearer "+bearer)
	respGetSend, err := app.Test(reqGetSend)
	if err != nil {
		t.Fatalf("GET /api/v1/messages failed: %v", err)
	}
	if respGetSend.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202 Accepted for GET /api/v1/messages, got %d", respGetSend.StatusCode)
	}
}
