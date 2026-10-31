package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"smswhiteinfo/customer-api/internal/billing"
)

// CompleteDeposit performs the atomic verification, state transition, SP balance debit, and customer credit inside a single DB transaction.
func CompleteDeposit(ctx context.Context, db *sql.DB, gateway, transactionID string, isSuccess bool, rawMeta map[string]any) (bool, error) {
	if !isSuccess {
		return false, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var depositID, userID int64
	var amount float64
	var currentStatus string

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, amount, status 
		FROM deposits 
		WHERE payment_method = ? AND transaction_id = ? 
		FOR UPDATE
	`, gateway, transactionID).Scan(&depositID, &userID, &amount, &currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if currentStatus == "completed" {
		return false, nil // Already processed
	}

	metaJSON, _ := json.Marshal(rawMeta)

	// Update deposit
	_, err = tx.ExecContext(ctx, `
		UPDATE deposits 
		SET status = 'completed', paid_at = NOW(), meta = ?, updated_at = NOW() 
		WHERE id = ?
	`, string(metaJSON), depositID)
	if err != nil {
		return false, err
	}

	// Fetch user details
	var userType string
	var userName string
	var spID sql.NullInt64

	err = tx.QueryRowContext(ctx, "SELECT type, name, sp_id FROM users WHERE id = ?", userID).Scan(&userType, &userName, &spID)
	if err != nil {
		return false, err
	}

	// If customer belongs to an SP reseller, debit the SP reseller's balance
	if userType == "customer" && spID.Valid && spID.Int64 > 0 {
		desc := fmt.Sprintf("Customer top-up: %s (#%d)", userName, depositID)
		_, dErr := billing.DebitBalance(ctx, tx, spID.Int64, amount, "deposit", depositID, desc)
		if dErr != nil {
			return false, fmt.Errorf("failed to debit SP reseller: %w", dErr)
		}
	}

	// Credit customer balance
	creditDesc := fmt.Sprintf("Deposit of %.2f via %s", amount, gateway)
	_, cErr := billing.CreditBalance(ctx, tx, userID, amount, "deposit", depositID, creditDesc)
	if cErr != nil {
		return false, fmt.Errorf("failed to credit customer: %w", cErr)
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

// CompleteManualDeposit handles manual admin deposit approvals using the exact same atomic transaction logic.
func CompleteManualDeposit(ctx context.Context, db *sql.DB, depositID int64, adminID int64) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var userID int64
	var amount float64
	var currentStatus, paymentMethod string

	err = tx.QueryRowContext(ctx, `
		SELECT user_id, amount, status, payment_method 
		FROM deposits 
		WHERE id = ? 
		FOR UPDATE
	`, depositID).Scan(&userID, &amount, &currentStatus, &paymentMethod)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if currentStatus == "completed" {
		return false, nil
	}

	// Update deposit
	_, err = tx.ExecContext(ctx, `
		UPDATE deposits 
		SET status = 'completed', paid_at = NOW(), deposited_by_id = ?, updated_at = NOW() 
		WHERE id = ?
	`, adminID, depositID)
	if err != nil {
		return false, err
	}

	var userType, userName string
	var spID sql.NullInt64
	err = tx.QueryRowContext(ctx, "SELECT type, name, sp_id FROM users WHERE id = ?", userID).Scan(&userType, &userName, &spID)
	if err != nil {
		return false, err
	}

	if userType == "customer" && spID.Valid && spID.Int64 > 0 {
		desc := fmt.Sprintf("Customer top-up: %s (#%d)", userName, depositID)
		_, dErr := billing.DebitBalance(ctx, tx, spID.Int64, amount, "deposit", depositID, desc)
		if dErr != nil {
			return false, fmt.Errorf("failed to debit SP reseller: %w", dErr)
		}
	}

	creditDesc := fmt.Sprintf("Deposit of %.2f via %s (Manual Approval)", amount, paymentMethod)
	_, cErr := billing.CreditBalance(ctx, tx, userID, amount, "deposit", depositID, creditDesc)
	if cErr != nil {
		return false, fmt.Errorf("failed to credit customer: %w", cErr)
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}
