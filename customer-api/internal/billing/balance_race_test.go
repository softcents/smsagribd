package billing

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
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
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(20)
	return db
}

func TestConcurrentBalanceDebitNeverNegative(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 1. Create temporary test user with 10.00 balance
	res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, password, balance, type, status, created_at, updated_at)
		VALUES ('Race Test User', ?, 'dummy_hash', 10.0000, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("race_test_%d@example.com", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userID, _ := res.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	// 2. Launch 20 concurrent goroutines attempting to debit 1.00 each
	var wg sync.WaitGroup
	var successes int64
	var failures int64

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			tx, txErr := db.BeginTx(ctx, nil)
			if txErr != nil {
				atomic.AddInt64(&failures, 1)
				return
			}
			defer tx.Rollback()

			_, dErr := DebitBalance(ctx, tx, userID, 1.00, "message", int64(1000+idx), fmt.Sprintf("Race debit #%d", idx))
			if dErr != nil {
				atomic.AddInt64(&failures, 1)
				return
			}

			if cErr := tx.Commit(); cErr == nil {
				atomic.AddInt64(&successes, 1)
			} else {
				atomic.AddInt64(&failures, 1)
			}
		}(i)
	}

	wg.Wait()

	if successes != 10 {
		t.Fatalf("Expected exactly 10 successful debits, got %d (failures: %d)", successes, failures)
	}
	if failures != 10 {
		t.Fatalf("Expected exactly 10 failed debits, got %d", failures)
	}

	var finalBalance float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&finalBalance)
	if finalBalance != 0.00 {
		t.Fatalf("Expected final balance 0.00, got %.4f (balance went negative or was miscalculated!)", finalBalance)
	}
}

func TestDoubleClaimingRace(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Insert test pending message
	res, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, created_at, updated_at)
		VALUES (1, '8801700000000', 'Test atomic claim', 'pending', 0.50, NOW(), NOW())
	`)
	if err != nil {
		t.Fatalf("Failed to insert test message: %v", err)
	}
	msgID, _ := res.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", msgID)
	}()

	var wg sync.WaitGroup
	var claimSuccesses int64
	var claimSkips int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			claimRes, cErr := db.ExecContext(ctx, `
				UPDATE messages 
				SET status = 'sending', updated_at = NOW() 
				WHERE id = ? AND status = 'pending'
			`, msgID)
			if cErr != nil {
				return
			}
			aff, _ := claimRes.RowsAffected()
			if aff == 1 {
				atomic.AddInt64(&claimSuccesses, 1)
			} else {
				atomic.AddInt64(&claimSkips, 1)
			}
		}()
	}

	wg.Wait()

	if claimSuccesses != 1 {
		t.Fatalf("Expected exactly ONE worker to successfully claim message, got %d", claimSuccesses)
	}
	if claimSkips != 9 {
		t.Fatalf("Expected 9 workers to skip already claimed message, got %d", claimSkips)
	}
}

func TestDoubleRefundRace(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test user with 0.00 balance
	res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, password, balance, type, status, created_at, updated_at)
		VALUES ('Refund Test User', ?, 'dummy_hash', 0.0000, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("refund_test_%d@example.com", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userID, _ := res.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	// Create test message in 'sending' status with cost 2.50
	mRes, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, created_at, updated_at)
		VALUES (?, '8801700000000', 'Test refund', 'sending', 2.5000, NOW(), NOW())
	`, userID)
	if err != nil {
		t.Fatalf("Failed to insert message: %v", err)
	}
	msgID, _ := mRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", msgID)
	}()

	var wg sync.WaitGroup
	var refundSuccesses int64
	var refundSkips int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, _ := RefundMessageFailed(ctx, db, msgID, "sending", "Carrier rejection")
			if ok {
				atomic.AddInt64(&refundSuccesses, 1)
			} else {
				atomic.AddInt64(&refundSkips, 1)
			}
		}()
	}

	wg.Wait()

	if refundSuccesses != 1 {
		t.Fatalf("Expected exactly 1 refund to succeed, got %d", refundSuccesses)
	}
	if refundSkips != 9 {
		t.Fatalf("Expected 9 refund attempts to skip, got %d", refundSkips)
	}

	// Verify balance was credited exactly once
	var balanceAfter float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&balanceAfter)
	if balanceAfter != 2.50 {
		t.Fatalf("Expected balance 2.50 (refunded once), got %.4f", balanceAfter)
	}
}
