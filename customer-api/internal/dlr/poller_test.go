package dlr

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return db
}

func TestDLR_FiveMinuteFallback_OnlySent(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	poller, err := NewDeliveryPoller(db, "base64:1234567890123456789012345678901212345678901=")
	if err != nil {
		t.Fatalf("Failed to create poller: %v", err)
	}

	// Create test customer
	now := time.Now().UnixNano()
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('DLR Test User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("dlr_user_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	// Insert 1 message in 'sent' (older than 5 min) -> SHOULD become delivered
	mSent, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, sent_at, created_at, updated_at)
		VALUES (?, '8801700000001', 'Sent msg', 'sent', 0.50, ?, ?, ?)
	`, userID, tenMinutesAgo, tenMinutesAgo, tenMinutesAgo)
	if err != nil {
		t.Fatalf("Failed to insert sent message: %v", err)
	}
	sentID, _ := mSent.LastInsertId()

	// Insert 1 message in 'pending' (older than 5 min) -> MUST NOT become delivered
	mPending, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, created_at, updated_at)
		VALUES (?, '8801700000002', 'Pending msg', 'pending', 0.50, ?, ?)
	`, userID, tenMinutesAgo, tenMinutesAgo)
	if err != nil {
		t.Fatalf("Failed to insert pending message: %v", err)
	}
	pendingID, _ := mPending.LastInsertId()

	// Insert 1 message in 'sending' (older than 5 min) -> MUST NOT become delivered
	mSending, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, created_at, updated_at)
		VALUES (?, '8801700000003', 'Sending msg', 'sending', 0.50, ?, ?)
	`, userID, tenMinutesAgo, tenMinutesAgo)
	if err != nil {
		t.Fatalf("Failed to insert sending message: %v", err)
	}
	sendingID, _ := mSending.LastInsertId()

	// Insert 1 message in 'sent' (only 1 min ago) -> MUST NOT become delivered
	oneMinAgo := time.Now().Add(-1 * time.Minute)
	mRecent, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, sent_at, created_at, updated_at)
		VALUES (?, '8801700000004', 'Recent sent msg', 'sent', 0.50, ?, ?, ?)
	`, userID, oneMinAgo, oneMinAgo, oneMinAgo)
	if err != nil {
		t.Fatalf("Failed to insert recent message: %v", err)
	}
	recentID, _ := mRecent.LastInsertId()

	// Run 5-minute fallback
	poller.runFiveMinuteFallback(ctx)

	// Verify sentID transitioned to 'delivered'
	var sentStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", sentID).Scan(&sentStatus)
	if sentStatus != "delivered" {
		t.Errorf("Expected sent message to transition to 'delivered', got '%s'", sentStatus)
	}

	// Verify pendingID remained 'pending'
	var pendingStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", pendingID).Scan(&pendingStatus)
	if pendingStatus != "pending" {
		t.Errorf("SECURITY CRITICAL: 5-minute fallback erroneously updated pending message to '%s'", pendingStatus)
	}

	// Verify sendingID remained 'sending'
	var sendingStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", sendingID).Scan(&sendingStatus)
	if sendingStatus != "sending" {
		t.Errorf("SECURITY CRITICAL: 5-minute fallback erroneously updated sending message to '%s'", sendingStatus)
	}

	// Verify recentID remained 'sent'
	var recentStatus string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", recentID).Scan(&recentStatus)
	if recentStatus != "sent" {
		t.Errorf("Expected recent sent message to stay 'sent', got '%s'", recentStatus)
	}
}

func TestDLR_BackoffSchedule(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	poller, err := NewDeliveryPoller(db, "base64:1234567890123456789012345678901212345678901=")
	if err != nil {
		t.Fatalf("Failed to create poller: %v", err)
	}

	now := time.Now().UnixNano()
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('DLR Backoff User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("dlr_backoff_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	mRes, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, poll_attempt, next_poll_at, created_at, updated_at)
		VALUES (?, '8801700000005', 'Test Backoff', 'sent', 0.50, 1, NOW(), NOW(), NOW())
	`, userID)
	if err != nil {
		t.Fatalf("Failed to insert message: %v", err)
	}
	msgID, _ := mRes.LastInsertId()

	// Test progression from attempt 1 through MaxPollAttempts (5)
	// Matches PollMessageDeliveryJob.php lines 206-214:
	// attempt 1 => 60s, attempt 2 => 120s, attempt 3 => 120s, default => 120s
	expectedDelays := map[int]int{
		2: 120, // +2 min
		3: 120, // +2 min
		4: 120, // +2 min
		5: 120, // +2 min
	}

	for currentAttempt := 1; currentAttempt < MaxPollAttempts; currentAttempt++ {
		before := time.Now()
		poller.scheduleNextPoll(ctx, msgID, currentAttempt)

		var nextAttempt int
		var nextPollAt time.Time
		err := db.QueryRowContext(ctx, "SELECT poll_attempt, next_poll_at FROM messages WHERE id = ?", msgID).Scan(&nextAttempt, &nextPollAt)
		if err != nil {
			t.Fatalf("Failed to scan message after attempt %d: %v", currentAttempt, err)
		}

		if nextAttempt != currentAttempt+1 {
			t.Errorf("Expected poll_attempt to be %d, got %d", currentAttempt+1, nextAttempt)
		}

		expectedDelay := expectedDelays[nextAttempt]
		actualDelay := int(nextPollAt.Sub(before).Seconds())
		// Allow small margin of 2 seconds for execution time
		if actualDelay < expectedDelay-2 || actualDelay > expectedDelay+2 {
			t.Errorf("Attempt %d -> %d expected delay ~%ds, got %ds", currentAttempt, nextAttempt, expectedDelay, actualDelay)
		}
	}

	// Test terminal state at MaxPollAttempts (5): should clear poll columns
	poller.scheduleNextPoll(ctx, msgID, MaxPollAttempts)
	var finalAttempt sql.NullInt64
	var finalNextPoll sql.NullTime
	_ = db.QueryRowContext(ctx, "SELECT poll_attempt, next_poll_at FROM messages WHERE id = ?", msgID).Scan(&finalAttempt, &finalNextPoll)
	if finalAttempt.Valid || finalNextPoll.Valid {
		t.Errorf("Expected poll columns to be cleared (NULL) after attempt %d, got attempt=%v, next_poll=%v", MaxPollAttempts, finalAttempt, finalNextPoll)
	}
}

func TestDLR_Delivered_And_Failed_RefundExactlyOnce(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	now := time.Now().UnixNano()
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('DLR Refund User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("dlr_ref_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	// 1. Delivered flow
	m1, _ := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, poll_attempt, next_poll_at, created_at, updated_at)
		VALUES (?, '8801700000006', 'Delivered test', 'sent', 0.50, 1, NOW(), NOW(), NOW())
	`, userID)
	msg1ID, _ := m1.LastInsertId()

	delivTime := time.Now()
	_, err = db.ExecContext(ctx, `
		UPDATE messages 
		SET status = 'delivered', delivered_at = ?, delivery_status = 'DELIVRD', poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW()
		WHERE id = ? AND status = 'sent'
	`, delivTime, msg1ID)
	if err != nil {
		t.Fatalf("Failed to mark delivered: %v", err)
	}

	var m1Status string
	var m1DelivAt sql.NullTime
	var m1Attempt sql.NullInt64
	_ = db.QueryRowContext(ctx, "SELECT status, delivered_at, poll_attempt FROM messages WHERE id = ?", msg1ID).Scan(&m1Status, &m1DelivAt, &m1Attempt)
	if m1Status != "delivered" || !m1DelivAt.Valid || m1Attempt.Valid {
		t.Errorf("Delivered assertion failed: status=%s, delivered_at=%v, poll_attempt=%v", m1Status, m1DelivAt, m1Attempt)
	}

	// 2. Failed flow with refund
	m2, _ := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, poll_attempt, next_poll_at, created_at, updated_at)
		VALUES (?, '8801700000007', 'Failed test', 'sent', 1.50, 1, NOW(), NOW(), NOW())
	`, userID)
	msg2ID, _ := m2.LastInsertId()

	// Initial balance: 10.00. First refund should succeed and increase balance to 11.50.
	fromStatus := "sent"
	reason := "Handset Busy / Undeliverable"

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	res1, err := tx1.ExecContext(ctx, `
		UPDATE messages 
		SET status = 'failed', failed_at = NOW(), failed_reason = ?, poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() 
		WHERE id = ? AND status = ?
	`, reason, msg2ID, fromStatus)
	if err != nil {
		_ = tx1.Rollback()
		t.Fatalf("Exec failed: %v", err)
	}
	aff1, _ := res1.RowsAffected()
	if aff1 != 1 {
		_ = tx1.Rollback()
		t.Fatalf("Expected 1 row affected on first refund attempt, got %d", aff1)
	}
	_, _ = tx1.ExecContext(ctx, "UPDATE users SET balance = balance + 1.50 WHERE id = ?", userID)
	_ = tx1.Commit()

	var balAfterFirst float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&balAfterFirst)
	if balAfterFirst != 11.50 {
		t.Fatalf("Expected balance 11.50 after first refund, got %.2f", balAfterFirst)
	}

	// 3. Duplicate DLR attempt on same message -> MUST affect 0 rows and NOT double-refund!
	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	res2, err := tx2.ExecContext(ctx, `
		UPDATE messages 
		SET status = 'failed', failed_at = NOW(), failed_reason = ?, poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() 
		WHERE id = ? AND status = ?
	`, reason, msg2ID, fromStatus)
	if err != nil {
		_ = tx2.Rollback()
		t.Fatalf("Exec 2 failed: %v", err)
	}
	aff2, _ := res2.RowsAffected()
	if aff2 != 0 {
		_ = tx2.Rollback()
		t.Fatalf("CRITICAL SECURITY FLAW: Duplicate DLR affected %d rows instead of 0!", aff2)
	}
	_ = tx2.Rollback()

	var balAfterDup float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&balAfterDup)
	if balAfterDup != 11.50 {
		t.Fatalf("CRITICAL: Balance changed on duplicate DLR callback! Expected 11.50, got %.2f", balAfterDup)
	}
}

