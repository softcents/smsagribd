package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestScheduledBatchFlipRace(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	ctx := context.Background()

	// Insert scheduled batch due now
	res, err := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, body, source_type, source_config, status, approval, scheduled_at, created_at, updated_at)
		VALUES (1, 'Scheduled test', 'manual', '{"recipients":[]}', 'scheduled', 'approved', DATE_SUB(NOW(), INTERVAL 1 MINUTE), NOW(), NOW())
	`)
	if err != nil {
		t.Fatalf("Failed to insert scheduled batch: %v", err)
	}
	batchID, _ := res.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM batches WHERE id = ?", batchID)
	}()

	var wg sync.WaitGroup
	var flipSuccesses int64
	var flipSkips int64

	// Two concurrent scheduler routines attempt to flip the batch
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			flipRes, fErr := db.ExecContext(ctx, `
				UPDATE batches 
				SET status = 'processing', updated_at = NOW() 
				WHERE id = ? AND status = 'scheduled'
			`, batchID)
			if fErr != nil {
				return
			}
			aff, _ := flipRes.RowsAffected()
			if aff == 1 {
				atomic.AddInt64(&flipSuccesses, 1)
			} else {
				atomic.AddInt64(&flipSkips, 1)
			}
		}()
	}

	wg.Wait()

	if flipSuccesses != 1 {
		t.Fatalf("Expected exactly ONE loop to flip scheduled batch, got %d", flipSuccesses)
	}
	if flipSkips != 1 {
		t.Fatalf("Expected second loop to skip already flipped batch, got %d", flipSkips)
	}
}

func TestPurgeStale_WorkerEnabledGuard_And_RefundFlag(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	ctx := context.Background()
	sched := NewMasterScheduler(db, nil)

	// Create test user
	now := time.Now().UnixNano()
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Purge Test User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("purge_u_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	}()

	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	// Insert stale message in 'sending'
	mRes, err := db.ExecContext(ctx, `
		INSERT INTO messages (user_id, `+"`to`"+`, body, status, cost, created_at, updated_at)
		VALUES (?, '8801700000008', 'Stale message', 'sending', 2.00, ?, ?)
	`, userID, tenMinutesAgo, tenMinutesAgo)
	if err != nil {
		t.Fatalf("Failed to insert stale message: %v", err)
	}
	msgID, _ := mRes.LastInsertId()

	// 1. Guard check: WORKER_ENABLED=false -> purge-stale MUST SKIP!
	_ = os.Setenv("WORKER_ENABLED", "false")
	sched.runPurgeStale(ctx, 8)

	var s1 string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", msgID).Scan(&s1)
	if s1 != "sending" {
		t.Fatalf("Item 7 guard failed: purge-stale ran when WORKER_ENABLED was false! Status is %s", s1)
	}

	// 2. WORKER_ENABLED=true with PURGE_STALE_REFUND=true
	_ = os.Setenv("WORKER_ENABLED", "true")
	_ = os.Setenv("PURGE_STALE_REFUND", "true")
	defer func() {
		_ = os.Setenv("WORKER_ENABLED", "false")
		_ = os.Setenv("PURGE_STALE_REFUND", "false")
	}()

	sched.runPurgeStale(ctx, 8)

	var s2 string
	_ = db.QueryRowContext(ctx, "SELECT status FROM messages WHERE id = ?", msgID).Scan(&s2)
	if s2 != "failed" {
		t.Fatalf("Expected message to be marked 'failed', got '%s'", s2)
	}

	// Verify user received 2.00 refund (10.00 + 2.00 = 12.00)
	var bal float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", userID).Scan(&bal)
	if bal != 12.00 {
		t.Errorf("Expected balance 12.00 after purge-stale refund, got %.2f", bal)
	}
}

