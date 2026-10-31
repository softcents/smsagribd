package scheduler

import (
	"context"
	"database/sql"
	"fmt"
)

type AdvisoryLock struct {
	name string
	conn *sql.Conn
}

// TryAcquireLock acquires a MySQL advisory lock on a dedicated connection with timeout 0 (non-blocking).
// Returns nil, nil if the lock is currently held by another connection/process.
func TryAcquireLock(ctx context.Context, db *sql.DB, name string) (*AdvisoryLock, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dedicated connection from pool: %w", err)
	}

	var acquired sql.NullInt64
	err = conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 0)", name).Scan(&acquired)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error executing GET_LOCK: %w", err)
	}

	if !acquired.Valid || acquired.Int64 != 1 {
		conn.Close()
		return nil, nil // Lock not acquired (already held)
	}

	return &AdvisoryLock{
		name: name,
		conn: conn,
	}, nil
}

// Release releases the advisory lock on the exact same connection and closes the connection back to the pool.
func (l *AdvisoryLock) Release(ctx context.Context) error {
	if l == nil || l.conn == nil {
		return nil
	}
	defer l.conn.Close()

	var released sql.NullInt64
	_ = l.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", l.name).Scan(&released)
	l.conn = nil
	return nil
}
