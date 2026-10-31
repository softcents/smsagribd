package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// --- Dashboard & Stats Types ---

type BillingDashboardStatsResponse struct {
	FinancialStats      BillingFinancialStats `json:"financial_stats"`
	ProfitStats         BillingProfitStats    `json:"profit_stats"`
	DepositTrend        []ChartTrendPoint     `json:"deposit_trend"`
	SpendingTrend       []SpendingTrendPoint  `json:"spending_trend"`
	MethodsDistribution []MethodStatPoint     `json:"methods_distribution"`
	TransactionVolume   []TxVolumePoint       `json:"transaction_volume"`
	RecentDeposits      []DepositItem         `json:"recent_deposits"`
	RecentTransactions  []TransactionItem     `json:"recent_transactions"`
}

type BillingFinancialStats struct {
	CompletedDeposits    float64 `json:"completed_deposits"`
	CustomerSpend        float64 `json:"customer_spend"`
	GatewayCost          float64 `json:"gateway_cost"`
	PendingDeposits      float64 `json:"pending_deposits"`
	PendingDepositsCount int     `json:"pending_deposits_count"`
	TransactionsCount    int     `json:"transactions_count"`
	SystemBalance        float64 `json:"system_balance"`
}

type BillingProfitStats struct {
	PeriodRevenue float64 `json:"period_revenue"`
	PeriodCost    float64 `json:"period_cost"`
	PeriodProfit  float64 `json:"period_profit"`
	PeriodMargin  float64 `json:"period_margin"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalCost     float64 `json:"total_cost"`
	TotalProfit   float64 `json:"total_profit"`
	TotalMargin   float64 `json:"total_margin"`
}

type ChartTrendPoint struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

type SpendingTrendPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Cost    float64 `json:"cost"`
}

type MethodStatPoint struct {
	Method string  `json:"method"`
	Count  int     `json:"count"`
	Total  float64 `json:"total"`
}

type TxVolumePoint struct {
	Date    string `json:"date"`
	Credits int    `json:"credits"`
	Debits  int    `json:"debits"`
}

type DepositItem struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	TransactionID *string `json:"transaction_id"`
	Status        string  `json:"status"`
	DepositedBy   *string `json:"deposited_by"`
	PaidAt        *string `json:"paid_at"`
	CreatedAt     string  `json:"created_at"`
}

type TransactionItem struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	Reference     string  `json:"reference"`
	Direction     string  `json:"direction"`
	Amount        float64 `json:"amount"`
	BalanceBefore float64 `json:"balance_before"`
	BalanceAfter  float64 `json:"balance_after"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
}

// --- Handler: Billing Dashboard Stats ---

func (server *Server) billingDashboardStats(context fiber.Ctx) error {
	period := strings.TrimSpace(context.Query("period", "30days"))
	fromDate := strings.TrimSpace(context.Query("from"))
	toDate := strings.TrimSpace(context.Query("to"))

	now := time.Now()
	var start time.Time
	var end time.Time = now

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "yesterday":
		y := now.AddDate(0, 0, -1)
		start = time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "7days":
		start = now.AddDate(0, 0, -7)
	case "30days":
		start = now.AddDate(0, 0, -30)
	case "this_month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "all":
		start = now.AddDate(-10, 0, 0)
	case "custom":
		if fromDate != "" {
			if t, err := time.Parse("2006-01-02", fromDate); err == nil {
				start = t
			}
		}
		if toDate != "" {
			if t, err := time.Parse("2006-01-02", toDate); err == nil {
				end = t.Add(24*time.Hour - time.Second)
			}
		}
		if start.IsZero() {
			start = now.AddDate(0, 0, -30)
		}
	default:
		start = now.AddDate(0, 0, -30)
	}

	startStr := start.Format("2006-01-02 15:04:05")
	endStr := end.Format("2006-01-02 15:04:05")

	var finStats BillingFinancialStats
	// Completed deposits in period
	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM deposits
		WHERE status = 'completed' AND deleted_at IS NULL AND created_at >= ? AND created_at <= ?
	`, startStr, endStr).Scan(&finStats.CompletedDeposits)

	// Customer spend in period
	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(cost), 0)
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
	`, startStr, endStr).Scan(&finStats.CustomerSpend)

	// Gateway cost in period
	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(operator_cost + primary_dipping_cost), 0)
		FROM messages
		WHERE status != 'failed' AND created_at >= ? AND created_at <= ?
	`, startStr, endStr).Scan(&finStats.GatewayCost)

	// Pending deposits
	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM deposits
		WHERE status = 'pending' AND deleted_at IS NULL AND created_at >= ? AND created_at <= ?
	`, startStr, endStr).Scan(&finStats.PendingDeposits, &finStats.PendingDepositsCount)

	// Transactions count
	_ = server.db.QueryRow(`
		SELECT COUNT(*)
		FROM transactions
		WHERE created_at >= ? AND created_at <= ?
	`, startStr, endStr).Scan(&finStats.TransactionsCount)

	// System balance (total customer balance)
	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(balance), 0)
		FROM users
		WHERE type = 'customer' AND deleted_at IS NULL
	`).Scan(&finStats.SystemBalance)

	// Profit Stats
	var profitStats BillingProfitStats
	profitStats.PeriodRevenue = finStats.CustomerSpend
	profitStats.PeriodCost = finStats.GatewayCost
	profitStats.PeriodProfit = profitStats.PeriodRevenue - profitStats.PeriodCost
	if profitStats.PeriodRevenue > 0 {
		profitStats.PeriodMargin = (profitStats.PeriodProfit / profitStats.PeriodRevenue) * 100
	}

	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(cost), 0)
		FROM messages
	`).Scan(&profitStats.TotalRevenue)

	_ = server.db.QueryRow(`
		SELECT COALESCE(SUM(operator_cost + primary_dipping_cost), 0)
		FROM messages
		WHERE status != 'failed'
	`).Scan(&profitStats.TotalCost)

	profitStats.TotalProfit = profitStats.TotalRevenue - profitStats.TotalCost
	if profitStats.TotalRevenue > 0 {
		profitStats.TotalMargin = (profitStats.TotalProfit / profitStats.TotalRevenue) * 100
	}

	// Deposit Trend (daily completed deposits)
	depositTrend := make([]ChartTrendPoint, 0)
	dRows, err := server.db.Query(`
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') as dt, COALESCE(SUM(amount), 0)
		FROM deposits
		WHERE status = 'completed' AND deleted_at IS NULL AND created_at >= ? AND created_at <= ?
		GROUP BY dt
		ORDER BY dt ASC
	`, startStr, endStr)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var p ChartTrendPoint
			if err := dRows.Scan(&p.Date, &p.Total); err == nil {
				depositTrend = append(depositTrend, p)
			}
		}
	}

	// Spending Trend
	spendingTrend := make([]SpendingTrendPoint, 0)
	sRows, err := server.db.Query(`
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') as dt, COALESCE(SUM(cost), 0) as rev, COALESCE(SUM(operator_cost + primary_dipping_cost), 0) as cst
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY dt
		ORDER BY dt ASC
	`, startStr, endStr)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var p SpendingTrendPoint
			if err := sRows.Scan(&p.Date, &p.Revenue, &p.Cost); err == nil {
				spendingTrend = append(spendingTrend, p)
			}
		}
	}

	// Deposits by Method
	methodsDist := make([]MethodStatPoint, 0)
	mRows, err := server.db.Query(`
		SELECT COALESCE(payment_method, 'manual') as method, COUNT(*), COALESCE(SUM(amount), 0)
		FROM deposits
		WHERE status = 'completed' AND deleted_at IS NULL AND created_at >= ? AND created_at <= ?
		GROUP BY method
	`, startStr, endStr)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var p MethodStatPoint
			if err := mRows.Scan(&p.Method, &p.Count, &p.Total); err == nil {
				methodsDist = append(methodsDist, p)
			}
		}
	}

	// Transaction volume (credits vs debits)
	txVolume := make([]TxVolumePoint, 0)
	txRows, err := server.db.Query(`
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') as dt,
			COALESCE(SUM(CASE WHEN direction = 'credit' THEN 1 ELSE 0 END), 0) as credits,
			COALESCE(SUM(CASE WHEN direction = 'debit' THEN 1 ELSE 0 END), 0) as debits
		FROM transactions
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY dt
		ORDER BY dt ASC
	`, startStr, endStr)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var p TxVolumePoint
			if err := txRows.Scan(&p.Date, &p.Credits, &p.Debits); err == nil {
				txVolume = append(txVolume, p)
			}
		}
	}

	// Recent Deposits (latest 5)
	recentDeposits := make([]DepositItem, 0)
	rdRows, err := server.db.Query(`
		SELECT d.id, d.user_id, u.name, COALESCE(u.phone, ''), d.amount, d.currency,
			COALESCE(d.payment_method, 'manual'), d.transaction_id, d.status,
			admin.name, d.paid_at, d.created_at
		FROM deposits d
		INNER JOIN users u ON u.id = d.user_id
		LEFT JOIN users admin ON admin.id = d.deposited_by_id
		WHERE d.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT 5
	`)
	if err == nil {
		defer rdRows.Close()
		for rdRows.Next() {
			var item DepositItem
			var paidAt sql.NullTime
			var createdAt time.Time
			var depBy sql.NullString
			if err := rdRows.Scan(
				&item.ID, &item.UserID, &item.CustomerName, &item.CustomerPhone,
				&item.Amount, &item.Currency, &item.PaymentMethod, &item.TransactionID,
				&item.Status, &depBy, &paidAt, &createdAt,
			); err == nil {
				if depBy.Valid {
					item.DepositedBy = &depBy.String
				}
				if paidAt.Valid {
					f := paidAt.Time.Format(time.RFC3339)
					item.PaidAt = &f
				}
				item.CreatedAt = createdAt.Format(time.RFC3339)
				recentDeposits = append(recentDeposits, item)
			}
		}
	}

	// Recent Transactions (latest 5)
	recentTransactions := make([]TransactionItem, 0)
	rtRows, err := server.db.Query(`
		SELECT t.id, t.user_id, u.name, COALESCE(u.phone, ''),
			COALESCE(t.transactionable_type, 'Manual'), t.direction, t.amount,
			t.balance_before, t.balance_after, t.description, t.created_at
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		ORDER BY t.created_at DESC
		LIMIT 5
	`)
	if err == nil {
		defer rtRows.Close()
		for rtRows.Next() {
			var item TransactionItem
			var createdAt time.Time
			if err := rtRows.Scan(
				&item.ID, &item.UserID, &item.CustomerName, &item.CustomerPhone,
				&item.Reference, &item.Direction, &item.Amount, &item.BalanceBefore,
				&item.BalanceAfter, &item.Description, &createdAt,
			); err == nil {
				item.CreatedAt = createdAt.Format(time.RFC3339)
				recentTransactions = append(recentTransactions, item)
			}
		}
	}

	return context.JSON(fiber.Map{
		"data": BillingDashboardStatsResponse{
			FinancialStats:      finStats,
			ProfitStats:         profitStats,
			DepositTrend:        depositTrend,
			SpendingTrend:       spendingTrend,
			MethodsDistribution: methodsDist,
			TransactionVolume:   txVolume,
			RecentDeposits:      recentDeposits,
			RecentTransactions:  recentTransactions,
		},
	})
}

// --- Handler: Adjust Balance (Manual Credit / Debit) ---

type AdjustBalanceRequest struct {
	UserID      int64   `json:"user_id"`
	Direction   string  `json:"direction"` // "credit" or "debit"
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func (server *Server) adjustBalance(context fiber.Ctx) error {
	var payload AdjustBalanceRequest
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	if payload.UserID <= 0 {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Valid Customer ID is required."})
	}
	if payload.Amount <= 0 {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Amount must be greater than zero."})
	}
	payload.Direction = strings.ToLower(strings.TrimSpace(payload.Direction))
	if payload.Direction != "credit" && payload.Direction != "debit" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Adjustment type must be either 'credit' or 'debit'."})
	}
	payload.Description = strings.TrimSpace(payload.Description)
	if payload.Description == "" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Reason / reference description is required."})
	}

	admin := superman(context)

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	// Lock customer row
	var currentBalance float64
	var customerName string
	err = tx.QueryRow(`
		SELECT name, balance
		FROM users
		WHERE id = ? AND deleted_at IS NULL
		FOR UPDATE
	`, payload.UserID).Scan(&customerName, &currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
		}
		return databaseError(context, err)
	}

	if payload.Direction == "debit" && currentBalance < payload.Amount {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot deduct more than current balance (৳%.2f).", currentBalance),
		})
	}

	balanceBefore := currentBalance
	var balanceAfter float64
	if payload.Direction == "credit" {
		balanceAfter = balanceBefore + payload.Amount
	} else {
		balanceAfter = balanceBefore - payload.Amount
	}

	// Update user balance
	_, err = tx.Exec("UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", balanceAfter, payload.UserID)
	if err != nil {
		return databaseError(context, err)
	}

	// Create deposit record
	txID := fmt.Sprintf("MANUAL-%d-%d", time.Now().Unix(), payload.UserID)
	res, err := tx.Exec(`
		INSERT INTO deposits (user_id, deposited_by_id, amount, currency, payment_method, transaction_id, status, paid_at, created_at, updated_at)
		VALUES (?, ?, ?, 'BDT', 'manual', ?, 'completed', NOW(), NOW(), NOW())
	`, payload.UserID, admin.AdminID, payload.Amount, txID)
	if err != nil {
		return databaseError(context, err)
	}
	depositID, _ := res.LastInsertId()

	// Create transaction record
	desc := fmt.Sprintf("%s (by %s)", payload.Description, admin.Name)
	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\Deposit', ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, payload.UserID, depositID, payload.Amount, payload.Direction, balanceBefore, balanceAfter, desc)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	actionVerb := "Added ৳"
	if payload.Direction == "debit" {
		actionVerb = "Deducted ৳"
	}

	return context.JSON(fiber.Map{
		"message": fmt.Sprintf("%s%.2f successfully for %s. New balance: ৳%.2f", actionVerb, payload.Amount, customerName, balanceAfter),
		"data": fiber.Map{
			"user_id":        payload.UserID,
			"balance_before": balanceBefore,
			"balance_after":  balanceAfter,
			"amount":         payload.Amount,
			"direction":      payload.Direction,
		},
	})
}

// --- Tariff Plans Full CRUD ---

type TariffPlanFullItem struct {
	ID                 int64                 `json:"id"`
	Name               string                `json:"name"`
	Slug               string                `json:"slug"`
	Type               string                `json:"type"`
	Status             string                `json:"status"`
	Currency           string                `json:"currency"`
	Description        *string               `json:"description"`
	RatesCount         int                   `json:"rates_count"`
	AssignedUsersCount int                   `json:"assigned_users_count"`
	Rates              []TariffRateWithOper  `json:"rates,omitempty"`
	CreatedAt          string                `json:"created_at"`
	UpdatedAt          string                `json:"updated_at"`
}

type TariffRateWithOper struct {
	ID              int64   `json:"id"`
	TariffPlanID    int64   `json:"tariff_plan_id"`
	OperatorID      int64   `json:"operator_id"`
	OperatorName    string  `json:"operator_name"`
	MaskingPrice    float64 `json:"masking_price"`
	NonMaskingPrice float64 `json:"non_masking_price"`
	EffectiveFrom   string  `json:"effective_from"`
	Notes           *string `json:"notes"`
}

func (server *Server) billingTariffPlans(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT tp.id, tp.name, COALESCE(tp.slug, ''), tp.type, tp.status, tp.currency, tp.description,
			(SELECT COUNT(*) FROM tariff_rates tr WHERE tr.tariff_plan_id = tp.id) as rates_count,
			(SELECT COUNT(*) FROM tariff_plan_assignments tpa WHERE tpa.tariff_plan_id = tp.id AND tpa.status = 'active') as users_count,
			tp.created_at, tp.updated_at
		FROM tariff_plans tp
		ORDER BY tp.id ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]TariffPlanFullItem, 0)
	for rows.Next() {
		var item TariffPlanFullItem
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Slug, &item.Type, &item.Status, &item.Currency,
			&item.Description, &item.RatesCount, &item.AssignedUsersCount,
			&createdAt, &updatedAt,
		); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) billingTariffPlanDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid plan ID."})
	}

	var item TariffPlanFullItem
	var createdAt, updatedAt time.Time
	err = server.db.QueryRow(`
		SELECT tp.id, tp.name, COALESCE(tp.slug, ''), tp.type, tp.status, tp.currency, tp.description,
			(SELECT COUNT(*) FROM tariff_rates tr WHERE tr.tariff_plan_id = tp.id) as rates_count,
			(SELECT COUNT(*) FROM tariff_plan_assignments tpa WHERE tpa.tariff_plan_id = tp.id AND tpa.status = 'active') as users_count,
			tp.created_at, tp.updated_at
		FROM tariff_plans tp
		WHERE tp.id = ?
	`, id).Scan(
		&item.ID, &item.Name, &item.Slug, &item.Type, &item.Status, &item.Currency,
		&item.Description, &item.RatesCount, &item.AssignedUsersCount,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Tariff plan not found."})
		}
		return databaseError(context, err)
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)

	// Fetch rates
	rateRows, err := server.db.Query(`
		SELECT tr.id, tr.tariff_plan_id, tr.operator_id, o.name, tr.masking_price, tr.non_masking_price, tr.effective_from, tr.notes
		FROM tariff_rates tr
		INNER JOIN operators o ON o.id = tr.operator_id
		WHERE tr.tariff_plan_id = ?
		ORDER BY o.name ASC
	`, id)
	if err == nil {
		defer rateRows.Close()
		rates := make([]TariffRateWithOper, 0)
		for rateRows.Next() {
			var r TariffRateWithOper
			var effFrom time.Time
			if err := rateRows.Scan(&r.ID, &r.TariffPlanID, &r.OperatorID, &r.OperatorName, &r.MaskingPrice, &r.NonMaskingPrice, &effFrom, &r.Notes); err == nil {
				r.EffectiveFrom = effFrom.Format("2006-01-02")
				rates = append(rates, r)
			}
		}
		item.Rates = rates
	}

	return context.JSON(fiber.Map{"data": item})
}

type SaveTariffPlanRequest struct {
	Name        string                    `json:"name"`
	Slug        string                    `json:"slug"`
	Type        string                    `json:"type"`
	Status      string                    `json:"status"`
	Currency    string                    `json:"currency"`
	Description *string                   `json:"description"`
	Rates       []SaveTariffRateInput     `json:"rates"`
}

type SaveTariffRateInput struct {
	OperatorID      int64   `json:"operator_id"`
	MaskingPrice    float64 `json:"masking_price"`
	NonMaskingPrice float64 `json:"non_masking_price"`
	Notes           *string `json:"notes"`
}

func (server *Server) createTariffPlan(context fiber.Ctx) error {
	var payload SaveTariffPlanRequest
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Plan name is required."})
	}
	if payload.Type == "" {
		payload.Type = "standard"
	}
	if payload.Status == "" {
		payload.Status = "active"
	}
	if payload.Currency == "" {
		payload.Currency = "BDT"
	}
	if payload.Slug == "" {
		payload.Slug = strings.ToLower(strings.ReplaceAll(payload.Name, " ", "-"))
	}

	admin := superman(context)

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO tariff_plans (name, slug, type, status, currency, description, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, payload.Name, payload.Slug, payload.Type, payload.Status, payload.Currency, payload.Description, admin.AdminID)
	if err != nil {
		return databaseError(context, err)
	}
	planID, _ := res.LastInsertId()

	for _, rate := range payload.Rates {
		rRes, err := tx.Exec(`
			INSERT INTO tariff_rates (tariff_plan_id, operator_id, masking_price, non_masking_price, effective_from, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, CURDATE(), ?, NOW(), NOW())
		`, planID, rate.OperatorID, rate.MaskingPrice, rate.NonMaskingPrice, rate.Notes)
		if err != nil {
			return databaseError(context, err)
		}
		rateID, _ := rRes.LastInsertId()

		// Log into tariff_rate_history
		_, _ = tx.Exec(`
			INSERT INTO tariff_rate_history (tariff_plan_id, tariff_rate_id, operator_id, old_masking_price, new_masking_price, old_non_masking_price, new_non_masking_price, event, changed_by, changed_at, created_at, updated_at)
			VALUES (?, ?, ?, NULL, ?, NULL, ?, 'created', ?, NOW(), NOW(), NOW())
		`, planID, rateID, rate.OperatorID, rate.MaskingPrice, rate.NonMaskingPrice, admin.AdminID)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": "Tariff plan created successfully.",
		"id":      planID,
	})
}

func (server *Server) updateTariffPlan(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid plan ID."})
	}

	var payload SaveTariffPlanRequest
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Plan name is required."})
	}

	admin := superman(context)

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE tariff_plans
		SET name = ?, type = ?, status = ?, currency = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`, payload.Name, payload.Type, payload.Status, payload.Currency, payload.Description, id)
	if err != nil {
		return databaseError(context, err)
	}

	// Update or insert operator rates
	for _, rate := range payload.Rates {
		var existingRateID int64
		var oldMasking, oldNonMasking float64
		err := tx.QueryRow(`
			SELECT id, masking_price, non_masking_price
			FROM tariff_rates
			WHERE tariff_plan_id = ? AND operator_id = ?
		`, id, rate.OperatorID).Scan(&existingRateID, &oldMasking, &oldNonMasking)

		if err == sql.ErrNoRows {
			// Insert new rate
			rRes, err := tx.Exec(`
				INSERT INTO tariff_rates (tariff_plan_id, operator_id, masking_price, non_masking_price, effective_from, notes, created_at, updated_at)
				VALUES (?, ?, ?, ?, CURDATE(), ?, NOW(), NOW())
			`, id, rate.OperatorID, rate.MaskingPrice, rate.NonMaskingPrice, rate.Notes)
			if err != nil {
				return databaseError(context, err)
			}
			newRateID, _ := rRes.LastInsertId()

			_, _ = tx.Exec(`
				INSERT INTO tariff_rate_history (tariff_plan_id, tariff_rate_id, operator_id, old_masking_price, new_masking_price, old_non_masking_price, new_non_masking_price, event, changed_by, changed_at, created_at, updated_at)
				VALUES (?, ?, ?, NULL, ?, NULL, ?, 'created', ?, NOW(), NOW(), NOW())
			`, id, newRateID, rate.OperatorID, rate.MaskingPrice, rate.NonMaskingPrice, admin.AdminID)
		} else if err == nil {
			// Update if changed
			if oldMasking != rate.MaskingPrice || oldNonMasking != rate.NonMaskingPrice {
				_, err := tx.Exec(`
					UPDATE tariff_rates
					SET masking_price = ?, non_masking_price = ?, notes = ?, updated_at = NOW()
					WHERE id = ?
				`, rate.MaskingPrice, rate.NonMaskingPrice, rate.Notes, existingRateID)
				if err != nil {
					return databaseError(context, err)
				}

				_, _ = tx.Exec(`
					INSERT INTO tariff_rate_history (tariff_plan_id, tariff_rate_id, operator_id, old_masking_price, new_masking_price, old_non_masking_price, new_non_masking_price, event, changed_by, changed_at, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, 'updated', ?, NOW(), NOW(), NOW())
				`, id, existingRateID, rate.OperatorID, oldMasking, rate.MaskingPrice, oldNonMasking, rate.NonMaskingPrice, admin.AdminID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Tariff plan updated successfully."})
}

func (server *Server) deleteTariffPlan(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid plan ID."})
	}

	var activeAssignments int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM tariff_plan_assignments WHERE tariff_plan_id = ? AND status = 'active'", id).Scan(&activeAssignments)
	if activeAssignments > 0 {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot delete plan. It has %d active customer assignment(s). Reassign them first.", activeAssignments),
		})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	_, _ = tx.Exec("DELETE FROM tariff_rate_history WHERE tariff_plan_id = ?", id)
	_, _ = tx.Exec("DELETE FROM tariff_rates WHERE tariff_plan_id = ?", id)
	_, _ = tx.Exec("DELETE FROM tariff_plan_assignments WHERE tariff_plan_id = ?", id)
	_, err = tx.Exec("DELETE FROM tariff_plans WHERE id = ?", id)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Tariff plan deleted successfully."})
}

// --- User-wise Assignment Handlers ---

type CustomerAssignmentItem struct {
	UserID        int64   `json:"user_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	CustomerPhone string  `json:"customer_phone"`
	PlanID        *int64  `json:"plan_id"`
	PlanName      *string `json:"plan_name"`
	PlanType      *string `json:"plan_type"`
	EffectiveFrom *string `json:"effective_from"`
	AssignedBy    *string `json:"assigned_by"`
	Status        *string `json:"status"`
}

type AssignmentHistoryItem struct {
	ID            int64   `json:"id"`
	PlanName      string  `json:"plan_name"`
	PlanType      string  `json:"plan_type"`
	Status        string  `json:"status"`
	EffectiveFrom string  `json:"effective_from"`
	AssignedBy    *string `json:"assigned_by"`
	CreatedAt     string  `json:"created_at"`
	Notes         *string `json:"notes"`
}

func (server *Server) billingAssignments(context fiber.Ctx) error {
	search := strings.TrimSpace(context.Query("search"))
	arguments := []any{}
	where := "WHERE u.type = 'customer' AND u.deleted_at IS NULL"
	if search != "" {
		where += " AND (u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}

	query := `
		SELECT u.id, u.name, u.email, COALESCE(u.phone, ''),
			tp.id, tp.name, tp.type, tpa.effective_from, admin.name, tpa.status
		FROM users u
		LEFT JOIN tariff_plan_assignments tpa ON tpa.user_id = u.id AND tpa.status = 'active'
		LEFT JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id
		LEFT JOIN users admin ON admin.id = tpa.assigned_by
		` + where + `
		ORDER BY u.name ASC
	`

	rows, err := server.db.Query(query, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]CustomerAssignmentItem, 0)
	for rows.Next() {
		var item CustomerAssignmentItem
		var planID sql.NullInt64
		var planName, planType, adminName, status sql.NullString
		var effFrom sql.NullTime

		if err := rows.Scan(
			&item.UserID, &item.CustomerName, &item.CustomerEmail, &item.CustomerPhone,
			&planID, &planName, &planType, &effFrom, &adminName, &status,
		); err == nil {
			if planID.Valid {
				item.PlanID = &planID.Int64
			}
			if planName.Valid {
				item.PlanName = &planName.String
			}
			if planType.Valid {
				item.PlanType = &planType.String
			}
			if effFrom.Valid {
				f := effFrom.Time.Format("2006-01-02")
				item.EffectiveFrom = &f
			}
			if adminName.Valid {
				item.AssignedBy = &adminName.String
			}
			if status.Valid {
				item.Status = &status.String
			}
			items = append(items, item)
		}
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) billingCustomerAssignmentHistory(context fiber.Ctx) error {
	userID, err := strconv.ParseInt(context.Params("user_id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID."})
	}

	rows, err := server.db.Query(`
		SELECT tpa.id, tp.name, tp.type, tpa.status, tpa.effective_from, admin.name, tpa.created_at, tpa.notes
		FROM tariff_plan_assignments tpa
		INNER JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id
		LEFT JOIN users admin ON admin.id = tpa.assigned_by
		WHERE tpa.user_id = ?
		ORDER BY tpa.created_at DESC
	`, userID)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	history := make([]AssignmentHistoryItem, 0)
	for rows.Next() {
		var h AssignmentHistoryItem
		var effFrom, createdAt time.Time
		if err := rows.Scan(&h.ID, &h.PlanName, &h.PlanType, &h.Status, &effFrom, &h.AssignedBy, &createdAt, &h.Notes); err == nil {
			h.EffectiveFrom = effFrom.Format("2006-01-02")
			h.CreatedAt = createdAt.Format(time.RFC3339)
			history = append(history, h)
		}
	}

	return context.JSON(fiber.Map{"data": history})
}

type AssignPlanRequest struct {
	UserID        int64   `json:"user_id"`
	TariffPlanID  int64   `json:"tariff_plan_id"`
	EffectiveFrom string  `json:"effective_from"`
	Notes         *string `json:"notes"`
}

func (server *Server) assignTariffPlan(context fiber.Ctx) error {
	var payload AssignPlanRequest
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	if payload.UserID <= 0 || payload.TariffPlanID <= 0 {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Customer and Tariff Plan are required."})
	}
	if payload.EffectiveFrom == "" {
		payload.EffectiveFrom = time.Now().Format("2006-01-02")
	}

	admin := superman(context)

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	// Verify plan is active
	var planStatus, planName string
	err = tx.QueryRow("SELECT status, name FROM tariff_plans WHERE id = ?", payload.TariffPlanID).Scan(&planStatus, &planName)
	if err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Tariff plan not found."})
	}
	if planStatus != "active" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Cannot assign a non-Active tariff plan."})
	}

	// Supersede existing active assignments
	_, err = tx.Exec("UPDATE tariff_plan_assignments SET status = 'superseded', updated_at = NOW() WHERE user_id = ? AND status = 'active'", payload.UserID)
	if err != nil {
		return databaseError(context, err)
	}

	// Insert new assignment
	_, err = tx.Exec(`
		INSERT INTO tariff_plan_assignments (user_id, tariff_plan_id, assigned_by, effective_from, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', ?, NOW(), NOW())
	`, payload.UserID, payload.TariffPlanID, admin.AdminID, payload.EffectiveFrom, payload.Notes)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("Assigned '%s' successfully.", planName)})
}

// --- Handler: Rate History ---

type RateHistoryRecord struct {
	ID                 int64    `json:"id"`
	PlanName           string   `json:"plan_name"`
	OperatorName       string   `json:"operator_name"`
	Event              string   `json:"event"`
	OldMaskingPrice    *float64 `json:"old_masking_price"`
	NewMaskingPrice    *float64 `json:"new_masking_price"`
	OldNonMaskingPrice *float64 `json:"old_non_masking_price"`
	NewNonMaskingPrice *float64 `json:"new_non_masking_price"`
	ChangedBy          *string  `json:"changed_by"`
	ChangedAt          string   `json:"changed_at"`
}

func (server *Server) billingRateHistory(context fiber.Ctx) error {
	page, perPage := pagination(context)
	planID := strings.TrimSpace(context.Query("plan_id"))
	event := strings.TrimSpace(context.Query("event"))

	arguments := []any{}
	where := "WHERE 1 = 1"
	if planID != "" && planID != "all" {
		where += " AND trh.tariff_plan_id = ?"
		arguments = append(arguments, planID)
	}
	if event != "" && event != "all" {
		where += " AND trh.event = ?"
		arguments = append(arguments, event)
	}

	var total int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM tariff_rate_history trh "+where, arguments...).Scan(&total)

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT trh.id, tp.name, o.name, trh.event,
			trh.old_masking_price, trh.new_masking_price,
			trh.old_non_masking_price, trh.new_non_masking_price,
			admin.name, trh.changed_at
		FROM tariff_rate_history trh
		INNER JOIN tariff_plans tp ON tp.id = trh.tariff_plan_id
		INNER JOIN operators o ON o.id = trh.operator_id
		LEFT JOIN users admin ON admin.id = trh.changed_by
		`+where+`
		ORDER BY trh.changed_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]RateHistoryRecord, 0)
	for rows.Next() {
		var item RateHistoryRecord
		var changedAt time.Time
		if err := rows.Scan(
			&item.ID, &item.PlanName, &item.OperatorName, &item.Event,
			&item.OldMaskingPrice, &item.NewMaskingPrice,
			&item.OldNonMaskingPrice, &item.NewNonMaskingPrice,
			&item.ChangedBy, &changedAt,
		); err == nil {
			item.ChangedAt = changedAt.Format("02 Jan 2006, 03:04 PM")
			items = append(items, item)
		}
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// --- Handler: Approve & Reject Deposit ---

func (server *Server) approveDeposit(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid deposit ID."})
	}

	admin := superman(context)

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	var depositUserID int64
	var amount float64
	var currentStatus string
	var paymentMethod string
	err = tx.QueryRow(`
		SELECT user_id, amount, status, COALESCE(payment_method, 'manual')
		FROM deposits
		WHERE id = ? AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&depositUserID, &amount, &currentStatus, &paymentMethod)
	if err != nil {
		if err == sql.ErrNoRows {
			return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Deposit not found."})
		}
		return databaseError(context, err)
	}

	if currentStatus == "completed" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Deposit has already been approved."})
	}

	// Update deposit
	_, err = tx.Exec(`
		UPDATE deposits
		SET status = 'completed', deposited_by_id = ?, paid_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`, admin.AdminID, id)
	if err != nil {
		return databaseError(context, err)
	}

	// Lock customer balance
	var balanceBefore float64
	var customerName string
	err = tx.QueryRow("SELECT name, balance FROM users WHERE id = ? FOR UPDATE", depositUserID).Scan(&customerName, &balanceBefore)
	if err != nil {
		return databaseError(context, err)
	}

	balanceAfter := balanceBefore + amount
	_, err = tx.Exec("UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", balanceAfter, depositUserID)
	if err != nil {
		return databaseError(context, err)
	}

	// Insert transaction record
	desc := fmt.Sprintf("Deposit of ৳%.2f via %s approved by %s (#%d)", amount, paymentMethod, admin.Name, id)
	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\Deposit', ?, ?, 'credit', ?, ?, ?, NOW(), NOW())
	`, depositUserID, id, amount, balanceBefore, balanceAfter, desc)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": fmt.Sprintf("Deposit #%d approved. ৳%.2f credited to %s.", id, amount, customerName),
	})
}

func (server *Server) rejectDeposit(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid deposit ID."})
	}

	res, err := server.db.Exec(`
		UPDATE deposits
		SET status = 'failed', updated_at = NOW()
		WHERE id = ? AND status = 'pending' AND deleted_at IS NULL
	`, id)
	if err != nil {
		return databaseError(context, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Deposit not found or not in pending status."})
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("Deposit #%d rejected.", id)})
}

// --- Handler: Invoices ---

type InvoiceRecord struct {
	ID            string  `json:"id"`
	DepositID     int64   `json:"deposit_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	CustomerPhone string  `json:"customer_phone"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	IssuedDate    string  `json:"issued_date"`
	PaidAt        *string `json:"paid_at"`
}

func (server *Server) billingInvoices(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	status := strings.TrimSpace(context.Query("status"))

	arguments := []any{}
	where := "WHERE d.deleted_at IS NULL"
	if search != "" {
		where += " AND (u.name LIKE ? OR u.email LIKE ? OR d.transaction_id LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}
	if status != "" && status != "all" {
		where += " AND d.status = ?"
		arguments = append(arguments, status)
	}

	var total int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM deposits d INNER JOIN users u ON u.id = d.user_id "+where, arguments...).Scan(&total)

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT d.id, u.name, u.email, COALESCE(u.phone, ''), d.amount, d.currency,
			COALESCE(d.payment_method, 'manual'), d.status, d.created_at, d.paid_at
		FROM deposits d
		INNER JOIN users u ON u.id = d.user_id
		`+where+`
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]InvoiceRecord, 0)
	for rows.Next() {
		var item InvoiceRecord
		var createdAt time.Time
		var paidAt sql.NullTime
		if err := rows.Scan(
			&item.DepositID, &item.CustomerName, &item.CustomerEmail, &item.CustomerPhone,
			&item.Amount, &item.Currency, &item.PaymentMethod, &item.Status,
			&createdAt, &paidAt,
		); err == nil {
			item.ID = fmt.Sprintf("INV-%04d-%04d", createdAt.Year(), item.DepositID)
			item.IssuedDate = createdAt.Format("02 Jan 2006")
			if paidAt.Valid {
				f := paidAt.Time.Format("02 Jan 2006, 03:04 PM")
				item.PaidAt = &f
			}
			items = append(items, item)
		}
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// --- Handler: Payments ---

type PaymentItem struct {
	ID            int64   `json:"id"`
	TransactionID *string `json:"transaction_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	PaidAt        *string `json:"paid_at"`
	CreatedAt     string  `json:"created_at"`
}

func (server *Server) billingPayments(context fiber.Ctx) error {
	page, perPage := pagination(context)
	method := strings.TrimSpace(context.Query("method"))
	status := strings.TrimSpace(context.Query("status"))
	search := strings.TrimSpace(context.Query("search"))

	arguments := []any{}
	where := "WHERE d.deleted_at IS NULL"
	if method != "" && method != "all" {
		where += " AND d.payment_method = ?"
		arguments = append(arguments, method)
	}
	if status != "" && status != "all" {
		where += " AND d.status = ?"
		arguments = append(arguments, status)
	}
	if search != "" {
		where += " AND (u.name LIKE ? OR u.phone LIKE ? OR d.transaction_id LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}

	var total int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM deposits d INNER JOIN users u ON u.id = d.user_id "+where, arguments...).Scan(&total)

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT d.id, d.transaction_id, u.name, COALESCE(u.phone, ''), d.amount, d.currency,
			COALESCE(d.payment_method, 'manual'), d.status, d.paid_at, d.created_at
		FROM deposits d
		INNER JOIN users u ON u.id = d.user_id
		`+where+`
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]PaymentItem, 0)
	for rows.Next() {
		var item PaymentItem
		var paidAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.TransactionID, &item.CustomerName, &item.CustomerPhone,
			&item.Amount, &item.Currency, &item.Method, &item.Status,
			&paidAt, &createdAt,
		); err == nil {
			if paidAt.Valid {
				f := paidAt.Time.Format("02 Jan 2006, 03:04 PM")
				item.PaidAt = &f
			}
			item.CreatedAt = createdAt.Format("02 Jan 2006, 03:04 PM")
			items = append(items, item)
		}
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// --- Handler: Usage / CDR ---

type UsageCdrItem struct {
	ID            int64   `json:"id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	SenderName    string  `json:"sender_name"`
	Recipient     string  `json:"recipient"`
	OperatorName  string  `json:"operator_name"`
	Cost          float64 `json:"cost"`
	GatewayCost   float64 `json:"gateway_cost"`
	Margin        float64 `json:"margin"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

func (server *Server) billingUsageCdr(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	status := strings.TrimSpace(context.Query("status"))

	arguments := []any{}
	where := "WHERE 1 = 1"
	if search != "" {
		where += " AND (u.name LIKE ? OR m.`to` LIKE ? OR m.sender_name LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}
	if status != "" && status != "all" {
		where += " AND m.status = ?"
		arguments = append(arguments, status)
	}

	var total int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM messages m INNER JOIN users u ON u.id = m.user_id "+where, arguments...).Scan(&total)

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT m.id, u.name, COALESCE(u.phone, ''), COALESCE(m.sender_name, 'N/A'),
			m.`+"`to`"+`, COALESCE(m.operator_name, 'Unknown'), m.cost,
			COALESCE(m.operator_cost + m.primary_dipping_cost, 0) as gw_cost,
			m.status, m.created_at
		FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		`+where+`
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]UsageCdrItem, 0)
	for rows.Next() {
		var item UsageCdrItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.CustomerName, &item.CustomerPhone, &item.SenderName,
			&item.Recipient, &item.OperatorName, &item.Cost, &item.GatewayCost,
			&item.Status, &createdAt,
		); err == nil {
			item.Margin = item.Cost - item.GatewayCost
			item.CreatedAt = createdAt.Format("02 Jan 2006, 03:04 PM")
			items = append(items, item)
		}
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// --- Handler: Billing & Financial Audit Reports ---

type BillingReportSummary struct {
	TotalCredits float64 `json:"total_credits"`
	TotalDebits  float64 `json:"total_debits"`
	NetFlow      float64 `json:"net_flow"`
	TotalTxCount int     `json:"total_tx_count"`
}

func (server *Server) billingReports(context fiber.Ctx) error {
	page, perPage := pagination(context)
	direction := strings.TrimSpace(context.Query("direction"))
	search := strings.TrimSpace(context.Query("search"))
	from := strings.TrimSpace(context.Query("from"))
	until := strings.TrimSpace(context.Query("until"))

	arguments := []any{}
	where := "WHERE 1 = 1"
	if direction != "" && direction != "all" {
		where += " AND t.direction = ?"
		arguments = append(arguments, direction)
	}
	if search != "" {
		where += " AND (u.name LIKE ? OR u.phone LIKE ? OR t.description LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}
	if from != "" {
		where += " AND DATE(t.created_at) >= ?"
		arguments = append(arguments, from)
	}
	if until != "" {
		where += " AND DATE(t.created_at) <= ?"
		arguments = append(arguments, until)
	}

	var summary BillingReportSummary
	summaryQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END), 0),
			COUNT(*)
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id ` + where
	_ = server.db.QueryRow(summaryQuery, arguments...).Scan(&summary.TotalCredits, &summary.TotalDebits, &summary.TotalTxCount)
	summary.NetFlow = summary.TotalCredits - summary.TotalDebits

	total := summary.TotalTxCount
	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT t.id, t.user_id, u.name, COALESCE(u.phone, ''),
			COALESCE(t.transactionable_type, 'General'), t.direction, t.amount,
			t.balance_before, t.balance_after, t.description, t.created_at
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		`+where+`
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]TransactionItem, 0)
	for rows.Next() {
		var item TransactionItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.CustomerName, &item.CustomerPhone,
			&item.Reference, &item.Direction, &item.Amount, &item.BalanceBefore,
			&item.BalanceAfter, &item.Description, &createdAt,
		); err == nil {
			item.CreatedAt = createdAt.Format("02 Jan 2006, 03:04 PM")
			items = append(items, item)
		}
	}

	return context.JSON(fiber.Map{
		"data":    items,
		"meta":    listMeta{Page: page, PerPage: perPage, Total: total},
		"summary": summary,
	})
}
