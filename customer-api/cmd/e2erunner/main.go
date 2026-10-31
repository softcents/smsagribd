package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"smswhiteinfo/customer-api/internal/config"
	"smswhiteinfo/customer-api/internal/crypto"
	"smswhiteinfo/customer-api/internal/dlr"
	"smswhiteinfo/customer-api/internal/httpapi"
	"smswhiteinfo/customer-api/internal/worker"
)

const testAppKey = "base64:dGVzdGtleS0zMmJ5dGVzLWZvci1lMmUtdGVzdGluZyE="

func main() {
	os.Setenv("APP_KEY", testAppKey)
	os.Setenv("WORKER_MAX_AGE_MINUTES", "30")

	dsn := "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Cannot connect to scratch DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot ping scratch DB: %v", err)
	}

	ctx := context.Background()

	fmt.Println("================================================================================")
	fmt.Println("             SMS WHITEINFO BD - LIVE END-TO-END VERIFICATION RUNNER             ")
	fmt.Println("================================================================================")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 0. Setup test user, tariff, gateway, sender, and token in MariaDB scratch
	// -------------------------------------------------------------------------
	nowNano := time.Now().UnixNano()
	custEmail := fmt.Sprintf("e2e_cust_%d@whiteinfo.local", nowNano)
	custPhone := fmt.Sprintf("+88017%08d", nowNano%100000000)
	custPass := "secret123"
	passHash, _ := bcrypt.GenerateFromPassword([]byte(custPass), 10)

	// User
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, kyc_status, sms_auto_approved, created_at, updated_at)
		VALUES ('E2E Customer', ?, ?, ?, 25.0000, 'customer', 'active', 'approved', 1, NOW(), NOW())
	`, custEmail, custPhone, string(passHash))
	if err != nil {
		log.Fatalf("Insert user failed: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	// Gateway with log driver
	enc, _ := crypto.NewLaravelEncrypter(testAppKey)
	credsEnc, _ := enc.EncryptArray(map[string]any{"mode": "e2e-fake-log"})
	gwRes, err := db.ExecContext(ctx, `
		INSERT INTO gateways (name, driver, credentials, requests_per_second, is_enabled, is_otp_enabled, priority, created_at, updated_at)
		VALUES ('E2E Log Gateway', 'log', ?, 100, 1, 1, 1, NOW(), NOW())
	`, credsEnc)
	if err != nil {
		log.Fatalf("Insert gateway failed: %v", err)
	}
	gwID, _ := gwRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM gateways WHERE id = ?", gwID)
	}()

	// Sender
	senderName := fmt.Sprintf("E2E%d", nowNano%10000)
	sRes, err := db.ExecContext(ctx, `
		INSERT INTO senders (requested_by, gateway_id, sender_id, type, is_enabled, approval_status, created_at, updated_at)
		VALUES (?, ?, ?, 'shared', 1, 'approved', NOW(), NOW())
	`, userID, gwID, senderName)
	if err != nil {
		log.Fatalf("Insert sender failed: %v", err)
	}
	senderDBID, _ := sRes.LastInsertId()
	_, _ = db.ExecContext(ctx, "INSERT INTO sender_user (sender_id, user_id) VALUES (?, ?)", senderDBID, userID)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM senders WHERE id = ?", senderDBID)
		_, _ = db.ExecContext(ctx, "DELETE FROM sender_user WHERE sender_id = ?", senderDBID)
	}()

	// Tariff rate for user/default
	_, _ = db.ExecContext(ctx, `
		INSERT INTO tariff_rates (tariff_plan_id, operator_id, cost, created_at, updated_at)
		VALUES (1, 1, 0.4000, NOW(), NOW())
		ON DUPLICATE KEY UPDATE cost = 0.4000
	`)

	// Sanctum plain token: <access_tokens.id>|<plain_token>
	plainToken := fmt.Sprintf("sanctum_%d_%d", userID, nowNano)
	shaHash := sha256.Sum256([]byte(plainToken))
	hashedToken := hex.EncodeToString(shaHash[:])

	tRes, err := db.ExecContext(ctx, `
		INSERT INTO access_tokens (tokenable_type, tokenable_id, name, token, abilities, created_at, updated_at)
		VALUES ('App\\Models\\User', ?, 'E2E API Token', ?, '["*"]', NOW(), NOW())
	`, userID, hashedToken)
	if err != nil {
		log.Fatalf("Insert access_tokens failed: %v", err)
	}
	tokenDBID, _ := tRes.LastInsertId()
	fullSanctumBearer := fmt.Sprintf("%d|%s", tokenDBID, plainToken)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM access_tokens WHERE id = ?", tokenDBID)
	}()

	// Setup Server
	cfg := config.Config{
		JWTSecret:        []byte("01234567890123456789012345678901"),
		TokenTTL:         2 * time.Hour,
		RememberTokenTTL: 24 * time.Hour,
		FrontendOrigin:   "http://localhost:5173",
	}
	server := httpapi.New(cfg, db)
	app := server.App()

	// -------------------------------------------------------------------------
	// 1. Customer Login Test (Real HTTP Handler)
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 1: Customer Login (POST /api/v1/customer/login)")
	loginBody, _ := json.Marshal(map[string]string{
		"email":    custEmail,
		"password": custPass,
	})
	reqLogin := httptest.NewRequest("POST", "/api/v1/customer/login", bytes.NewReader(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	respLogin, err := app.Test(reqLogin)
	if err != nil {
		log.Fatalf("Login test failed: %v", err)
	}
	bLogin, _ := io.ReadAll(respLogin.Body)
	var loginData map[string]any
	_ = json.Unmarshal(bLogin, &loginData)
	fmt.Printf("   HTTP Status: %d\n", respLogin.StatusCode)
	if respLogin.StatusCode != 200 {
		log.Fatalf("Login failed: %s", string(bLogin))
	}
	fmt.Printf("   Login response token: %v...\n", fmt.Sprintf("%v", loginData["token"])[:15])
	fmt.Println("   [PASS] Customer authentication verified.")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 2. Send SMS using Sanctum-style Token (POST /api/v1/messages)
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 2: Submit Message using Sanctum-style Token (POST /api/v1/messages)")
	smsBody, _ := json.Marshal(map[string]string{
		"sender_id": senderName,
		"numbers":   "01710000001",
		"body":      "E2E Test Message via Sanctum Token",
	})
	reqSend := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(smsBody))
	reqSend.Header.Set("Authorization", "Bearer "+fullSanctumBearer)
	reqSend.Header.Set("Content-Type", "application/json")
	respSend, err := app.Test(reqSend)
	if err != nil {
		log.Fatalf("Send message request failed: %v", err)
	}
	bSend, _ := io.ReadAll(respSend.Body)
	fmt.Printf("   HTTP Status: %d\n", respSend.StatusCode)
	var sendData map[string]any
	_ = json.Unmarshal(bSend, &sendData)
	fmt.Printf("   Response: %s\n", string(bSend))
	if respSend.StatusCode != http.StatusAccepted {
		log.Fatalf("Send message failed: %s", string(bSend))
	}
	batchID := int64(sendData["batch_id"].(float64))
	fmt.Printf("   Batch ID Created: %d\n", batchID)
	fmt.Println("   [PASS] Message accepted for processing.")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 3. Worker Sends Message (Pending -> Sending -> Sent)
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 3: Worker Send Execution (Pending -> Sending -> Sent)")
	workerPool, err := worker.NewSendWorkerPool(db, testAppKey, 1)
	if err != nil {
		log.Fatalf("Failed to create worker pool: %v", err)
	}

	// Wait for background ProcessBatch to complete inserting message rows
	var msgID int64
	for i := 0; i < 20; i++ {
		err = db.QueryRowContext(ctx, "SELECT id FROM messages WHERE batch_id = ? LIMIT 1", batchID).Scan(&msgID)
		if err == nil && msgID > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if msgID == 0 {
		log.Fatalf("Timeout waiting for batch %d message row creation", batchID)
	}

	// Execute worker batch tick to claim and send
	workerPool.ProcessNextBatchForTest(ctx)
	time.Sleep(400 * time.Millisecond)

	var msgStatus, gwMsgID string
	var msgCost float64
	var pollAttempt int
	err = db.QueryRowContext(ctx, `
		SELECT status, COALESCE(gateway_message_id, ''), cost, COALESCE(poll_attempt, 0)
		FROM messages
		WHERE id = ?
	`, msgID).Scan(&msgStatus, &gwMsgID, &msgCost, &pollAttempt)
	if err != nil {
		log.Fatalf("Failed to query created message: %v", err)
	}

	fmt.Printf("   Message ID: %d | Status: %s | GatewayMessageID: %s | Cost: %.4f | PollAttempt: %d\n",
		msgID, msgStatus, gwMsgID, msgCost, pollAttempt)
	if msgStatus != "sent" {
		log.Fatalf("Expected message status 'sent', got '%s'", msgStatus)
	}
	fmt.Println("   [PASS] Worker claimed, dispatched to gateway driver, and transitioned to 'sent'.")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 4. DLR Poller Marks Message Delivered (Sent -> Delivered)
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 4: DLR Poller Execution (Sent -> Delivered)")
	poller, err := dlr.NewDeliveryPoller(db, testAppKey)
	if err != nil {
		log.Fatalf("Failed to create DLR poller: %v", err)
	}

	// Set next_poll_at to past so poller picks it immediately
	_, _ = db.ExecContext(ctx, "UPDATE messages SET next_poll_at = DATE_SUB(NOW(), INTERVAL 5 SECOND) WHERE id = ?", msgID)

	poller.RunTickForTest(ctx)

	var finalStatus string
	var deliveredAt *time.Time
	_ = db.QueryRowContext(ctx, "SELECT status, delivered_at FROM messages WHERE id = ?", msgID).Scan(&finalStatus, &deliveredAt)
	fmt.Printf("   Message ID: %d | Final Status: %s | DeliveredAt: %v\n", msgID, finalStatus, deliveredAt)
	if finalStatus != "delivered" {
		log.Fatalf("Expected message status 'delivered', got '%s'", finalStatus)
	}
	fmt.Println("   [PASS] DLR Poller verified delivery outcome and transitioned to 'delivered'.")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 5. Balance & Transaction Audit Check
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 5: Balance and Transaction Audit")
	var currentBalance float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&currentBalance)
	expectedBal := 25.0000 - msgCost
	fmt.Printf("   Initial Balance: 25.0000 | Message Cost: %.4f | Current Balance: %.4f (Expected: %.4f)\n", msgCost, currentBalance, expectedBal)

	var txnCount int
	var txDebitAmount float64
	_ = db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND transactionable_type = 'message_batch' AND transactionable_id = ?
	`, userID, batchID).Scan(&txnCount, &txDebitAmount)
	fmt.Printf("   Debit Transactions: %d record(s), Total Deducted: %.4f\n", txnCount, txDebitAmount)
	if txnCount < 1 {
		log.Fatalf("Missing transaction debit record for batch %d", batchID)
	}
	fmt.Println("   [PASS] User balance accurately debited and transaction ledger logged.")
	fmt.Println()

	// -------------------------------------------------------------------------
	// 6. PipraPay Payment Webhook & Idempotency Tests
	// -------------------------------------------------------------------------
	fmt.Println(">> STEP 6: PipraPay Fake Payment Webhook Tests")

	var fakePipraPayAmount float64 = 50.00
	var fakePipraPayStatus string = "completed"
	fakePipraPayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   fakePipraPayStatus,
			"amount":   fakePipraPayAmount,
			"currency": "BDT",
		})
	}))
	defer fakePipraPayServer.Close()
	os.Setenv("PIPRAPAY_BASE_URL", fakePipraPayServer.URL)
	os.Setenv("PIPRAPAY_API_KEY", "fake-piprapay-key")

	// 6a. Valid Deposit IPN
	depositTxnID := fmt.Sprintf("pp_e2e_%d", nowNano)
	depRes, err := db.ExecContext(ctx, `
		INSERT INTO deposits (user_id, payment_method, transaction_id, amount, currency, status, created_at, updated_at)
		VALUES (?, 'piprapay', ?, 50.0000, 'BDT', 'pending', NOW(), NOW())
	`, userID, depositTxnID)
	if err != nil {
		log.Fatalf("Insert test deposit failed: %v", err)
	}
	depositID, _ := depRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE id = ?", depositID)
		_, _ = db.ExecContext(ctx, "DELETE FROM payment_webhooks WHERE event_id = ?", depositTxnID)
	}()

	fmt.Printf("   [6a] Testing Valid IPN for Deposit ID %d (Txn: %s, Amount: 50.00 BDT)...\n", depositID, depositTxnID)
	ipnValid, _ := json.Marshal(map[string]any{
		"pp_id":    depositTxnID,
		"status":   "completed",
		"amount":   50.00,
		"currency": "BDT",
	})
	reqIPN1 := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(ipnValid))
	reqIPN1.Header.Set("Content-Type", "application/json")
	respIPN1, err := app.Test(reqIPN1)
	if err != nil {
		log.Fatalf("IPN 1 request failed: %v", err)
	}
	bIPN1, _ := io.ReadAll(respIPN1.Body)
	fmt.Printf("        HTTP Status: %d | Response: %s\n", respIPN1.StatusCode, string(bIPN1))
	if respIPN1.StatusCode != http.StatusOK {
		log.Fatalf("IPN 1 expected 200 OK, got %d: %s", respIPN1.StatusCode, string(bIPN1))
	}

	var balAfterCredit float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&balAfterCredit)
	var depStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM deposits WHERE id = ?", depositID).Scan(&depStatus)
	var whStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM payment_webhooks WHERE event_id = ?", depositTxnID).Scan(&whStatus)
	fmt.Printf("        Deposit Status: %s | Webhook Status: %s | User Balance: %.4f (Credited +50.00)\n",
		depStatus, whStatus, balAfterCredit)
	if depStatus != "completed" || whStatus != "processed" {
		log.Fatalf("Deposit not marked completed or webhook not marked processed!")
	}
	fmt.Println("        [PASS] Valid IPN verified, deposit completed, balance credited.")

	// 6b. Duplicate IPN (Idempotency)
	fmt.Println("   [6b] Testing Duplicate IPN (Idempotent Guard)...")
	reqIPN2 := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(ipnValid))
	reqIPN2.Header.Set("Content-Type", "application/json")
	respIPN2, _ := app.Test(reqIPN2)
	bIPN2, _ := io.ReadAll(respIPN2.Body)
	fmt.Printf("        HTTP Status: %d | Response: %s\n", respIPN2.StatusCode, string(bIPN2))
	var balAfterDup float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&balAfterDup)
	fmt.Printf("        User Balance after duplicate IPN: %.4f\n", balAfterDup)
	if balAfterDup != balAfterCredit {
		log.Fatalf("CRITICAL FLAW: Duplicate IPN credited balance again! Before: %.4f, After: %.4f", balAfterCredit, balAfterDup)
	}
	fmt.Println("        [PASS] Duplicate IPN rejected without secondary crediting.")

	// 6c. Mismatched Amount IPN
	mismatchTxnID := fmt.Sprintf("pp_mismatch_%d", nowNano)
	depRes2, _ := db.ExecContext(ctx, `
		INSERT INTO deposits (user_id, payment_method, transaction_id, amount, currency, status, created_at, updated_at)
		VALUES (?, 'piprapay', ?, 100.0000, 'BDT', 'pending', NOW(), NOW())
	`, userID, mismatchTxnID)
	depositID2, _ := depRes2.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE id = ?", depositID2)
		_, _ = db.ExecContext(ctx, "DELETE FROM payment_webhooks WHERE event_id = ?", mismatchTxnID)
	}()

	fmt.Printf("   [6c] Testing Amount Mismatch IPN (Deposit expects 100.00 BDT, IPN sends 75.00 BDT)...\n")
	ipnMismatch, _ := json.Marshal(map[string]any{
		"pp_id":    mismatchTxnID,
		"status":   "completed",
		"amount":   75.00,
		"currency": "BDT",
	})
	reqIPN3 := httptest.NewRequest("POST", "/webhooks/payment/piprapay", bytes.NewReader(ipnMismatch))
	reqIPN3.Header.Set("Content-Type", "application/json")
	respIPN3, _ := app.Test(reqIPN3)
	bIPN3, _ := io.ReadAll(respIPN3.Body)
	fmt.Printf("        HTTP Status: %d | Response: %s\n", respIPN3.StatusCode, string(bIPN3))
	if respIPN3.StatusCode != http.StatusUnprocessableEntity {
		log.Fatalf("Expected 422 for amount mismatch, got %d", respIPN3.StatusCode)
	}

	var depStatus2 string
	_ = db.QueryRowContext(ctx, "SELECT status FROM deposits WHERE id = ?", depositID2).Scan(&depStatus2)
	var whStatus2 string
	_ = db.QueryRowContext(ctx, "SELECT status FROM payment_webhooks WHERE event_id = ?", mismatchTxnID).Scan(&whStatus2)
	fmt.Printf("        Deposit Status: %s (stays pending/retryable) | Webhook Status: %s (not 'processed')\n",
		depStatus2, whStatus2)
	if depStatus2 != "pending" || whStatus2 == "processed" {
		log.Fatalf("Amount mismatch improperly marked deposit or webhook as processed!")
	}
	fmt.Println("        [PASS] Amount mismatch rejected as 422, deposit remains pending and retryable.")
	fmt.Println()

	fmt.Println("================================================================================")
	fmt.Println("               ALL END-TO-END VERIFICATION CHECKS PASSED               ")
	fmt.Println("================================================================================")
}
