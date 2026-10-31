package scheduler

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestPurgeStaleMessagesAndRefund(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	ctx := context.Background()

	// Ensure WORKER_ENABLED=true so the safety guard allows purge-stale to run
	os.Setenv("WORKER_ENABLED", "true")
	os.Setenv("PURGE_STALE_REFUND", "true")
	defer os.Unsetenv("WORKER_ENABLED")
	defer os.Unsetenv("PURGE_STALE_REFUND")

	// 1. Create a test user with initial balance 10.00
	userRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, balance, created_at, updated_at) 
		VALUES ('Purge Test User', 'purge@test.local', 10.0000, NOW(), NOW())
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userID, _ := userRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	// 2. Insert two stale messages: one 'pending', one 'sending', both updated 10 minutes ago
	cost := 0.4500
	staleTime := time.Now().Add(-10 * time.Minute)

	msgRes1, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, sender_id, `+"`to`"+`, body, cost, status, created_at, updated_at)
		VALUES (?, NULL, '8801700000001', 'Stale pending test', ?, 'pending', ?, ?)
	`, userID, cost, staleTime, staleTime)
	if err != nil {
		t.Fatalf("Failed to insert stale pending message: %v", err)
	}
	msgID1, _ := msgRes1.LastInsertId()

	msgRes2, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, sender_id, `+"`to`"+`, body, cost, status, created_at, updated_at)
		VALUES (?, NULL, '8801700000002', 'Stale sending test', ?, 'sending', ?, ?)
	`, userID, cost, staleTime, staleTime)
	if err != nil {
		t.Fatalf("Failed to insert stale sending message: %v", err)
	}
	msgID2, _ := msgRes2.LastInsertId()

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE id IN (?, ?)", msgID1, msgID2)
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE transactionable_type = 'message' AND transactionable_id IN (?, ?)", msgID1, msgID2)
	}()

	// 3. Run purge stale with threshold 5 minutes
	scheduler := NewMasterScheduler(db, nil)
	scheduler.RunPurgeStaleForTest(ctx, 5)

	// 4. Verify both messages are marked 'failed' with failed_at set
	for _, mid := range []int64{msgID1, msgID2} {
		var status, reason string
		var failedAt *time.Time
		err := db.QueryRowContext(ctx, "SELECT status, failed_reason, failed_at FROM messages WHERE id = ?", mid).Scan(&status, &reason, &failedAt)
		if err != nil {
			t.Fatalf("Failed to query message %d: %v", mid, err)
		}
		if status != "failed" {
			t.Errorf("Expected message %d status='failed', got '%s'", mid, status)
		}
		if failedAt == nil {
			t.Errorf("Expected message %d to have non-nil failed_at", mid)
		}
	}

	// 5. Verify user balance was refunded exactly once per message: 10.00 + (2 * 0.45) = 10.90
	var finalBalance float64
	err = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&finalBalance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	expectedBalance := 10.90
	if finalBalance < expectedBalance-0.001 || finalBalance > expectedBalance+0.001 {
		t.Errorf("Expected user balance %.4f after purge refund, got %.4f", expectedBalance, finalBalance)
	}

	// 6. Verify exactly 2 refund transactions exist (one for each message)
	var txnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM transactions 
		WHERE transactionable_type = 'message' AND transactionable_id IN (?, ?) AND direction = 'credit'
	`, msgID1, msgID2).Scan(&txnCount)
	if err != nil {
		t.Fatalf("Failed to count refund transactions: %v", err)
	}
	if txnCount != 2 {
		t.Errorf("Expected exactly 2 refund transactions, found %d", txnCount)
	}

	// 7. Verify Idempotency: re-running purge stale does NOT create duplicate refunds
	scheduler.RunPurgeStaleForTest(ctx, 5)

	var txnCountAfterSecondRun int
	_ = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM transactions 
		WHERE transactionable_type = 'message' AND transactionable_id IN (?, ?) AND direction = 'credit'
	`, msgID1, msgID2).Scan(&txnCountAfterSecondRun)
	if txnCountAfterSecondRun != 2 {
		t.Errorf("Idempotency violation: second purge created duplicate transactions, got %d", txnCountAfterSecondRun)
	}
}
