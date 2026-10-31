package scheduler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestSchedulerAdvisoryLockExclusivity(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lockName := "sms_whiteinfo:test_lock_exclusive"

	// 1. Instance A acquires lock on dedicated conn
	lockA, err := TryAcquireLock(ctx, db, lockName)
	if err != nil {
		t.Fatalf("Instance A failed to acquire lock: %v", err)
	}
	if lockA == nil {
		t.Fatalf("Instance A expected lock but got nil")
	}

	// 2. Instance B attempts to acquire the same lock on a different conn -> must skip (return nil, nil)
	lockB, err := TryAcquireLock(ctx, db, lockName)
	if err != nil {
		t.Fatalf("Instance B error: %v", err)
	}
	if lockB != nil {
		_ = lockB.Release(ctx)
		t.Fatalf("Instance B acquired lock while Instance A was holding it! Expected nil lock (skip).")
	}

	// 3. Instance A releases lock
	if err := lockA.Release(ctx); err != nil {
		t.Fatalf("Instance A failed to release lock: %v", err)
	}

	// 4. Instance B can now acquire the lock
	lockB2, err := TryAcquireLock(ctx, db, lockName)
	if err != nil {
		t.Fatalf("Instance B second attempt failed: %v", err)
	}
	if lockB2 == nil {
		t.Fatalf("Instance B expected to acquire released lock, but got nil")
	}
	_ = lockB2.Release(ctx)
}
