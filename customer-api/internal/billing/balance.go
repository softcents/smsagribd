package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DebitBalance locks the user row, verifies affordability, debits the balance, and records a transaction.
func DebitBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, transType string, transID int64, description string) (float64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("debit amount must be greater than zero")
	}

	var balanceBefore float64
	err := tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ? FOR UPDATE", userID).Scan(&balanceBefore)
	if err != nil {
		return 0, fmt.Errorf("failed to lock user for debit: %w", err)
	}

	if balanceBefore < amount {
		return balanceBefore, ErrInsufficientBalance
	}

	balanceAfter := RoundFloat(balanceBefore-amount, 4)

	_, err = tx.ExecContext(ctx, "UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", balanceAfter, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to update user balance: %w", err)
	}

	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'debit', ?, ?, ?, ?, ?)
	`, userID, transType, transID, amount, balanceBefore, balanceAfter, description, now, now)
	if err != nil {
		return 0, fmt.Errorf("failed to insert debit transaction: %w", err)
	}

	return balanceAfter, nil
}

// CreditBalance locks the user row, increments the balance, and records a transaction.
func CreditBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, transType string, transID int64, description string) (float64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("credit amount must be greater than zero")
	}

	var balanceBefore float64
	err := tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = ? FOR UPDATE", userID).Scan(&balanceBefore)
	if err != nil {
		return 0, fmt.Errorf("failed to lock user for credit: %w", err)
	}

	balanceAfter := RoundFloat(balanceBefore+amount, 4)

	_, err = tx.ExecContext(ctx, "UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", balanceAfter, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to update user balance: %w", err)
	}

	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'credit', ?, ?, ?, ?, ?)
	`, userID, transType, transID, amount, balanceBefore, balanceAfter, description, now, now)
	if err != nil {
		return 0, fmt.Errorf("failed to insert credit transaction: %w", err)
	}

	return balanceAfter, nil
}

// RefundMessageFailed atomically marks a message as failed and refunds its cost if and only if the status change affected 1 row.
func RefundMessageFailed(ctx context.Context, db *sql.DB, messageID int64, expectedStatus string, reason string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// 1. Fetch message details
	var userID int64
	var cost float64
	var currentStatus string
	err = tx.QueryRowContext(ctx, "SELECT user_id, cost, status FROM messages WHERE id = ?", messageID).Scan(&userID, &cost, &currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// 2. Perform atomic state transition
	var res sql.Result
	if expectedStatus != "" {
		res, err = tx.ExecContext(ctx, `
			UPDATE messages 
			SET status = 'failed', failed_at = NOW(), failed_reason = ?, updated_at = NOW() 
			WHERE id = ? AND status = ?
		`, reason, messageID, expectedStatus)
	} else {
		res, err = tx.ExecContext(ctx, `
			UPDATE messages 
			SET status = 'failed', failed_at = NOW(), failed_reason = ?, updated_at = NOW() 
			WHERE id = ? AND status IN ('pending', 'sending', 'sent')
		`, reason, messageID)
	}
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		// Another worker already processed or refunded this message
		return false, nil
	}

	// 3. Exactly-once refund
	if cost > 0 {
		_, err = CreditBalance(ctx, tx, userID, cost, "message", messageID, fmt.Sprintf("Refund: SMS failed (%s)", reason))
		if err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}
