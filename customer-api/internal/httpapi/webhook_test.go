package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"smswhiteinfo/customer-api/internal/config"
)

func TestPaymentAndSmsWebhooks(t *testing.T) {
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
	}

	server := New(cfg, db)
	app := server.App()
	ctx := context.Background()

	// 1. Unknown payment gateway rejection (Item 4c)
	reqUnknown := httptest.NewRequest("POST", "/webhooks/payment/stripe", bytes.NewBufferString(`{}`))
	reqUnknown.Header.Set("Content-Type", "application/json")
	respUnknown, err := app.Test(reqUnknown)
	if err != nil {
		t.Fatalf("Unknown gateway request failed: %v", err)
	}
	if respUnknown.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for unknown gateway, got %d", respUnknown.StatusCode)
	}

	// 2. Setup mock PipraPay verification server
	var mockVerifyAmount float64 = 100.00
	var mockVerifyCurrency string = "BDT"
	var mockVerifyStatus string = "completed"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/verify-payment" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":   mockVerifyStatus,
				"pp_id":    "test_pp_12345",
				"amount":   mockVerifyAmount,
				"currency": mockVerifyCurrency,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	_ = os.Setenv("PIPRAPAY_BASE_URL", mockServer.URL)
	_ = os.Setenv("PIPRAPAY_API_KEY", "test_key_abc")

	// Create test customer
	now := time.Now().UnixNano()
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Webhook Test Customer', ?, ?, 'hash', 50.0000, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("wh_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	custID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM payment_webhooks WHERE gateway = 'piprapay'")
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", custID)
	}()

	ppTxnID := fmt.Sprintf("pp_wh_%d", now)

	// Create deposit row exactly as Laravel stored it:
	// status='pending', amount=100.0000, currency='BDT', payment_method='piprapay'
	dRes, err := db.ExecContext(ctx, `
		INSERT INTO deposits (user_id, amount, currency, payment_method, transaction_id, status, created_at, updated_at)
		VALUES (?, 100.0000, 'BDT', 'piprapay', ?, 'pending', NOW(), NOW())
	`, custID, ppTxnID)
	if err != nil {
		t.Fatalf("Failed to create deposit: %v", err)
	}
	depositID, _ := dRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE id = ?", depositID)
	}()

	// 2a. Test Amount Mismatch rejection: provider returns 50.00, deposit is 100.00
	mockVerifyAmount = 50.00
	mockVerifyCurrency = "bdt" // lowercase to test case-insensitivity on provider side
	payloadMismatch, _ := json.Marshal(map[string]any{
		"pp_id":    ppTxnID,
		"status":   "completed",
		"amount":   50.00,
		"currency": "bdt",
	})
	reqMismatch := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(payloadMismatch))
	reqMismatch.Header.Set("Content-Type", "application/json")
	respMismatch, _ := app.Test(reqMismatch)
	if respMismatch.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for amount mismatch, got %d", respMismatch.StatusCode)
	}

	// Verify that payment_webhooks record was inserted with status != 'processed' (retryable)
	var whStatus string
	err = db.QueryRowContext(ctx, "SELECT status FROM payment_webhooks WHERE event_id = ?", ppTxnID).Scan(&whStatus)
	if err != nil {
		t.Fatalf("Failed to query payment_webhooks: %v", err)
	}
	if whStatus == "processed" {
		t.Errorf("CRITICAL BUG: payment_webhooks was marked processed on amount mismatch rejection!")
	}

	// 2b. Test that the webhook is RETRYABLE without deleting any database records
	// Provider now returns correct matching amount 100.00 and currency "bdt" (case-insensitive test)
	mockVerifyAmount = 100.00
	mockVerifyCurrency = "bdt"
	mockVerifyStatus = "completed"

	payloadValid, _ := json.Marshal(map[string]any{
		"pp_id":    ppTxnID,
		"status":   "completed",
		"amount":   100.00,
		"currency": "BDT",
	})
	reqValid := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(payloadValid))
	reqValid.Header.Set("Content-Type", "application/json")
	respValid, err := app.Test(reqValid)
	if err != nil {
		t.Fatalf("Valid webhook request failed: %v", err)
	}
	if respValid.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respValid.Body)
		t.Fatalf("Expected 200 for valid payment webhook retry, got %d: %s", respValid.StatusCode, string(body))
	}

	// Verify deposit was credited: 50.00 + 100.00 = 150.00
	var finalBal float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", custID).Scan(&finalBal)
	if finalBal != 150.00 {
		t.Errorf("Expected balance 150.00 after successful webhook credit, got %.2f", finalBal)
	}

	// Verify payment_webhooks now has status = 'processed'
	err = db.QueryRowContext(ctx, "SELECT status FROM payment_webhooks WHERE event_id = ?", ppTxnID).Scan(&whStatus)
	if err != nil || whStatus != "processed" {
		t.Errorf("Expected payment_webhooks to have status='processed' after success, got %s", whStatus)
	}

	// 2c. Test Idempotency on repeated webhook (Item 4d)
	reqDup := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(payloadValid))
	reqDup.Header.Set("Content-Type", "application/json")
	respDup, _ := app.Test(reqDup)
	if respDup.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 on duplicate webhook, got %d", respDup.StatusCode)
	}

	var balAfterDup float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", custID).Scan(&balAfterDup)
	if balAfterDup != 150.00 {
		t.Errorf("CRITICAL SECURITY FLAW: Duplicate webhook credited balance again! Expected 150.00, got %.2f", balAfterDup)
	}

	// 3. SMS DLR Webhook test
	gwMsgID := fmt.Sprintf("gw_dlr_%d", now)
	mRes, _ := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, gateway_message_id, created_at, updated_at)
		VALUES (?, '8801700000009', 'DLR Webhook msg', 'sent', 0.50, ?, NOW(), NOW())
	`, custID, gwMsgID)
	smsID, _ := mRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", smsID)
	}()

	reqSmsDLR := httptest.NewRequest("GET", fmt.Sprintf("/webhooks/sms/infozillion?Message_ID=%s&Status=DELIVRD", gwMsgID), nil)
	respSmsDLR, err := app.Test(reqSmsDLR)
	if err != nil {
		t.Fatalf("SMS DLR request failed: %v", err)
	}
	if respSmsDLR.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for SMS DLR webhook, got %d", respSmsDLR.StatusCode)
	}

	var smsStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", smsID).Scan(&smsStatus)
	if smsStatus != "delivered" {
		t.Errorf("Expected SMS status to be updated to 'delivered', got '%s'", smsStatus)
	}
}
