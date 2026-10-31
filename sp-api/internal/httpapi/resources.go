package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// ─── Types ──────────────────────────────────────────────────────────────────

type listMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type spCustomer struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Balance   float64 `json:"balance"`
	KYCStatus string  `json:"kyc_status"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type distributionEntry struct {
	ID           int64   `json:"id"`
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"notes"`
	Date         string  `json:"date"`
}

type transactionEntry struct {
	ID           int64   `json:"id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	BalanceAfter float64 `json:"balance_after"`
	Date         string  `json:"date"`
}

type smsLog struct {
	ID           int64   `json:"id"`
	CustomerName string  `json:"customer"`
	Recipient    string  `json:"recipient"`
	Body         string  `json:"body"`
	Cost         float64 `json:"cost"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

type dashboardStats struct {
	Balance          float64 `json:"balance"`
	CustomerCount    int     `json:"customer_count"`
	TotalDistributed float64 `json:"total_distributed"`
}

type fundPayload struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (server *Server) dashboard(ctx fiber.Ctx) error {
	claims := sp(ctx)

	var stats dashboardStats

	// SP balance
	if err := server.db.QueryRow(
		"SELECT balance FROM users WHERE id = ? AND type = 'sp' AND deleted_at IS NULL",
		claims.SPID,
	).Scan(&stats.Balance); err != nil {
		return databaseError(ctx, err)
	}

	// Count customers belonging to this SP
	if err := server.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE sp_id = ? AND type = 'customer' AND deleted_at IS NULL",
		claims.SPID,
	).Scan(&stats.CustomerCount); err != nil {
		stats.CustomerCount = 0
	}

	// Total amount ever distributed TO customers by this SP
	row := server.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM sp_distributions
		WHERE sp_id = ?
	`, claims.SPID)
	_ = row.Scan(&stats.TotalDistributed)

	return ctx.JSON(fiber.Map{"data": stats})
}

// ─── Customers ───────────────────────────────────────────────────────────────

func (server *Server) customers(ctx fiber.Ctx) error {
	claims := sp(ctx)
	page, perPage := pagination(ctx)
	search := strings.TrimSpace(ctx.Query("search"))

	arguments := []any{claims.SPID, "customer"}
	where := "WHERE sp_id = ? AND type = ? AND deleted_at IS NULL"

	if search != "" {
		where += " AND (name LIKE ? OR email LIKE ? OR phone LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM users "+where, arguments...).Scan(&total); err != nil {
		return ctx.JSON(fiber.Map{
			"data": []spCustomer{},
			"meta": listMeta{Page: page, PerPage: perPage, Total: 0},
		})
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT id, name, email, COALESCE(phone, ''), balance, COALESCE(kyc_status, 'pending'), status, created_at
		FROM users `+where+`
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(ctx, err)
	}
	defer rows.Close()

	items := make([]spCustomer, 0)
	for rows.Next() {
		var item spCustomer
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Phone, &item.Balance,
			&item.KYCStatus, &item.Status, &createdAt); err != nil {
			return databaseError(ctx, err)
		}
		item.CreatedAt = createdAt.Format("2006-01-02 15:04")
		items = append(items, item)
	}

	return ctx.JSON(fiber.Map{
		"data": items,
		"meta": listMeta{Page: page, PerPage: perPage, Total: total},
	})
}

// fundCustomer deducts from SP balance and credits the customer's balance.
func (server *Server) fundCustomer(ctx fiber.Ctx) error {
	claims := sp(ctx)

	customerID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || customerID < 1 {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	payload := fundPayload{}
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil || payload.Amount <= 0 {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "A positive amount is required.",
		})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(ctx, err)
	}
	defer tx.Rollback()

	// Lock and read SP balance
	var spBalance float64
	var spName string
	if err := tx.QueryRow(
		"SELECT balance, name FROM users WHERE id = ? AND type = 'sp' AND deleted_at IS NULL FOR UPDATE",
		claims.SPID,
	).Scan(&spBalance, &spName); err != nil {
		return databaseError(ctx, err)
	}

	if payload.Amount > spBalance {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Insufficient reseller balance.",
		})
	}

	// Lock and read customer balance – must belong to this SP
	var customerBalance float64
	var customerName string
	err = tx.QueryRow(
		"SELECT balance, name FROM users WHERE id = ? AND sp_id = ? AND type = 'customer' AND deleted_at IS NULL FOR UPDATE",
		customerID, claims.SPID,
	).Scan(&customerBalance, &customerName)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found or not assigned to you."})
	}
	if err != nil {
		return databaseError(ctx, err)
	}

	newSPBalance := spBalance - payload.Amount
	newCustomerBalance := customerBalance + payload.Amount

	// Deduct from SP
	if _, err := tx.Exec(
		"UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?",
		newSPBalance, claims.SPID,
	); err != nil {
		return databaseError(ctx, err)
	}

	// Credit customer
	if _, err := tx.Exec(
		"UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?",
		newCustomerBalance, customerID,
	); err != nil {
		return databaseError(ctx, err)
	}

	note := payload.Description
	if note == "" {
		note = fmt.Sprintf("Fund transfer to %s", customerName)
	}

	// Record in sp_distributions
	res, err := tx.Exec(`
		INSERT INTO sp_distributions (sp_id, customer_id, amount, note, distributed_by, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`, claims.SPID, customerID, payload.Amount, note, claims.SPID)
	if err != nil {
		return databaseError(ctx, err)
	}
	distID, _ := res.LastInsertId()

	desc := fmt.Sprintf("S P distribution #%d: %s", distID, note)

	// Transaction ledger entry for customer (credit)
	if _, err := tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\SpDistribution', ?, ?, 'credit', ?, ?, ?, NOW(), NOW())
	`, customerID, distID, payload.Amount, customerBalance, newCustomerBalance, desc); err != nil {
		return databaseError(ctx, err)
	}

	// Transaction ledger entry for SP (debit)
	resDebit, err := tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\SpDistribution', ?, ?, 'debit', ?, ?, ?, NOW(), NOW())
	`, claims.SPID, distID, math.Abs(payload.Amount), spBalance, newSPBalance, desc)
	if err != nil {
		return databaseError(ctx, err)
	}
	debitTxID, _ := resDebit.LastInsertId()

	// Update sp_distributions with debit transaction_id
	if _, err := tx.Exec("UPDATE sp_distributions SET transaction_id = ? WHERE id = ?", debitTxID, distID); err != nil {
		return databaseError(ctx, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(ctx, err)
	}

	return ctx.JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully transferred ৳%.2f to %s.", payload.Amount, customerName),
		"data": fiber.Map{
			"sp_balance":       newSPBalance,
			"customer_balance": newCustomerBalance,
		},
	})
}

// ─── Wallet (distribution ledger) ────────────────────────────────────────────

func (server *Server) wallet(ctx fiber.Ctx) error {
	claims := sp(ctx)
	page, perPage := pagination(ctx)

	var total int
	_ = server.db.QueryRow(`
		SELECT COUNT(*) FROM sp_distributions
		WHERE sp_id = ?
	`, claims.SPID).Scan(&total)

	rows, err := server.db.Query(`
		SELECT d.id,
		       d.customer_id,
		       COALESCE(u.name, 'Unknown') AS customer_name,
		       d.amount,
		       COALESCE(d.note, '') AS note,
		       d.created_at
		FROM sp_distributions d
		INNER JOIN users u ON u.id = d.customer_id
		WHERE d.sp_id = ?
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`, claims.SPID, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(ctx, err)
	}
	defer rows.Close()

	items := make([]distributionEntry, 0)
	for rows.Next() {
		var item distributionEntry
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.CustomerID, &item.CustomerName, &item.Amount, &item.Description, &createdAt); err != nil {
			return databaseError(ctx, err)
		}
		item.Date = createdAt.Format("2006-01-02 15:04")
		items = append(items, item)
	}

	// Also fetch current SP balance for the wallet page header
	var balance float64
	_ = server.db.QueryRow(
		"SELECT balance FROM users WHERE id = ? AND type = 'sp' AND deleted_at IS NULL",
		claims.SPID,
	).Scan(&balance)

	return ctx.JSON(fiber.Map{
		"data": fiber.Map{
			"balance":       balance,
			"distributions": items,
		},
		"meta": listMeta{Page: page, PerPage: perPage, Total: total},
	})
}

// ─── Billing (full SP transaction ledger) ────────────────────────────────────

func (server *Server) billing(ctx fiber.Ctx) error {
	claims := sp(ctx)
	page, perPage := pagination(ctx)

	var total int
	_ = server.db.QueryRow(
		"SELECT COUNT(*) FROM transactions WHERE user_id = ?",
		claims.SPID,
	).Scan(&total)

	rows, err := server.db.Query(`
		SELECT id,
		       CASE direction WHEN 'credit' THEN 'Credit Allocation' ELSE 'Client Distribution' END AS type,
		       CASE direction WHEN 'credit' THEN amount ELSE -amount END AS signed_amount,
		       description,
		       balance_after,
		       created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, claims.SPID, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(ctx, err)
	}
	defer rows.Close()

	items := make([]transactionEntry, 0)
	for rows.Next() {
		var item transactionEntry
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Type, &item.Amount, &item.Description,
			&item.BalanceAfter, &createdAt); err != nil {
			return databaseError(ctx, err)
		}
		item.Date = createdAt.Format("2006-01-02 15:04")
		items = append(items, item)
	}

	return ctx.JSON(fiber.Map{
		"data": items,
		"meta": listMeta{Page: page, PerPage: perPage, Total: total},
	})
}

// ─── SMS Messages (traffic for this SP's customers) ──────────────────────────

func (server *Server) messages(ctx fiber.Ctx) error {
	claims := sp(ctx)
	page, perPage := pagination(ctx)

	// Count messages sent by any customer whose sp_id = this SP
	var total int
	_ = server.db.QueryRow(`
		SELECT COUNT(*) FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		WHERE u.sp_id = ? AND u.type = 'customer' AND u.deleted_at IS NULL
	`, claims.SPID).Scan(&total)

	rows, err := server.db.Query(`
		SELECT m.id,
		       u.name AS customer_name,
		       m.to,
		       COALESCE(m.body, '') AS body,
		       COALESCE(m.cost, 0) AS cost,
		       COALESCE(m.status, 'unknown') AS status,
		       m.created_at
		FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		WHERE u.sp_id = ? AND u.type = 'customer' AND u.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, claims.SPID, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(ctx, err)
	}
	defer rows.Close()

	items := make([]smsLog, 0)
	for rows.Next() {
		var item smsLog
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.CustomerName, &item.Recipient, &item.Body,
			&item.Cost, &item.Status, &createdAt); err != nil {
			return databaseError(ctx, err)
		}
		item.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		items = append(items, item)
	}

	return ctx.JSON(fiber.Map{
		"data": items,
		"meta": listMeta{Page: page, PerPage: perPage, Total: total},
	})
}
