package payment

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

func TestConcurrentDepositCompletion(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	now := time.Now().UnixNano()
	// Create test customer
	cRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Deposit Test Customer', ?, ?, 'dummy_hash', 50.0000, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("cust_%d@example.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	custID, _ := cRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", custID)
	}()

	txnID := fmt.Sprintf("pp_txn_%d", time.Now().UnixNano())

	// Create pending deposit of 100.00
	dRes, err := db.ExecContext(ctx, `
		INSERT INTO deposits (user_id, amount, payment_method, transaction_id, status, created_at, updated_at)
		VALUES (?, 100.0000, 'piprapay', ?, 'pending', NOW(), NOW())
	`, custID, txnID)
	if err != nil {
		t.Fatalf("Failed to create pending deposit: %v", err)
	}
	depositID, _ := dRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE id = ?", depositID)
	}()

	// Run 20 concurrent CompleteDeposit calls
	var wg sync.WaitGroup
	var completedCount int64
	var ignoredCount int64

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, cErr := CompleteDeposit(ctx, db, "piprapay", txnID, true, map[string]any{"status": "completed"})
			if cErr == nil && ok {
				atomic.AddInt64(&completedCount, 1)
			} else {
				atomic.AddInt64(&ignoredCount, 1)
			}
		}()
	}

	wg.Wait()

	if completedCount != 1 {
		t.Fatalf("Expected exactly 1 CompleteDeposit to return true, got %d", completedCount)
	}
	if ignoredCount != 19 {
		t.Fatalf("Expected 19 calls to be ignored, got %d", ignoredCount)
	}

	// Verify balance was credited exactly once (50 + 100 = 150)
	var finalBalance float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", custID).Scan(&finalBalance)
	if finalBalance != 150.00 {
		t.Fatalf("Expected customer balance to be 150.00, got %.4f", finalBalance)
	}
}

func TestSpResellerBalanceHandling(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 1. Create SP reseller with 500.00 balance
	t1 := time.Now().UnixNano()
	spRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('SP Reseller', ?, ?, 'dummy_hash', 500.0000, 'sp', 'active', NOW(), NOW())
	`, fmt.Sprintf("sp_%d@example.com", t1), fmt.Sprintf("+88018%08d", t1%100000000))
	if err != nil {
		t.Fatalf("Failed to create SP: %v", err)
	}
	spID, _ := spRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", spID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", spID)
	}()

	// 2. Create Customer linked to this SP
	t2 := time.Now().UnixNano() + 1
	cRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, sp_id, status, created_at, updated_at)
		VALUES ('Customer of SP', ?, ?, 'dummy_hash', 10.0000, 'customer', ?, 'active', NOW(), NOW())
	`, fmt.Sprintf("cust_sp_%d@example.com", t2), fmt.Sprintf("+88019%08d", t2%100000000), spID)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	custID, _ := cRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE user_id = ?", custID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", custID)
	}()

	txnID := fmt.Sprintf("sp_deposit_%d", time.Now().UnixNano())

	// 3. Create deposit of 200.00
	dRes, err := db.ExecContext(ctx, `
		INSERT INTO deposits (user_id, amount, payment_method, transaction_id, status, created_at, updated_at)
		VALUES (?, 200.0000, 'piprapay', ?, 'pending', NOW(), NOW())
	`, custID, txnID)
	if err != nil {
		t.Fatalf("Failed to create deposit: %v", err)
	}
	depID, _ := dRes.LastInsertId()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM deposits WHERE id = ?", depID)
	}()

	// 4. Complete deposit
	ok, err := CompleteDeposit(ctx, db, "piprapay", txnID, true, map[string]any{"status": "completed"})
	if err != nil || !ok {
		t.Fatalf("CompleteDeposit failed: %v", err)
	}

	// 5. Verify SP balance debited: 500 - 200 = 300
	var spBalance float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", spID).Scan(&spBalance)
	if spBalance != 300.00 {
		t.Fatalf("Expected SP balance to be 300.00, got %.4f", spBalance)
	}

	// 6. Verify Customer balance credited: 10 + 200 = 210
	var custBalance float64
	_ = db.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ?", custID).Scan(&custBalance)
	if custBalance != 210.00 {
		t.Fatalf("Expected Customer balance to be 210.00, got %.4f", custBalance)
	}
}
