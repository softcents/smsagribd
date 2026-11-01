package httpapi

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type listMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type customerSummary struct {
	TotalCustomers       int     `json:"total_customers"`
	TotalBalance         float64 `json:"total_balance"`
	AvgBalance           float64 `json:"avg_balance"`
	FundedCustomersCount int     `json:"funded_customers_count"`
	ZeroBalanceCount     int     `json:"zero_balance_count"`
	WithNidCount         int     `json:"with_nid_count"`
	VerifiedKycCount     int     `json:"verified_kyc_count"`
	UnderReviewKycCount  int     `json:"under_review_kyc_count"`
	OtherKycCount        int     `json:"other_kyc_count"`
	KycVerifiedRate      float64 `json:"kyc_verified_rate"`
	ApprovedStatusCount  int     `json:"approved_status_count"`
	PendingStatusCount   int     `json:"pending_status_count"`
	InactiveStatusCount  int     `json:"inactive_status_count"`
	ApprovedRate         float64 `json:"approved_rate"`
	FundedPercentage     float64 `json:"funded_percentage"`
}

type customer struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	BusinessName   *string `json:"business_name"`
	Balance        float64 `json:"balance"`
	KYCStatus      string  `json:"kyc_status"`
	Status         string  `json:"status"`
	NIDNumber      *string `json:"nid_number"`
	Tariff         *string `json:"tariff"`
	LastLoginAt    *string `json:"last_login_at"`
	CreatedByAdmin bool    `json:"created_by_admin"`
	CreatedAt      string  `json:"created_at"`
}

func (server *Server) customers(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	creationSource := strings.TrimSpace(context.Query("creation_source"))
	kycStatus := strings.TrimSpace(context.Query("kyc_status"))
	status := strings.TrimSpace(context.Query("status"))
	tariffPlan := strings.TrimSpace(context.Query("tariff_plan"))
	balanceRange := strings.TrimSpace(context.Query("balance_range"))
	regFrom := strings.TrimSpace(context.Query("registered_from"))
	regUntil := strings.TrimSpace(context.Query("registered_until"))

	arguments := []any{"customer"}
	where := "WHERE u.type = ? AND u.deleted_at IS NULL"

	if search != "" {
		where += " AND (u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ? OR u.business_name LIKE ? OR u.nid_number LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term, term, term)
	}

	if creationSource == "1" {
		where += " AND u.created_by_admin = true"
	} else if creationSource == "0" {
		where += " AND (u.created_by_admin = false OR u.created_by_admin IS NULL)"
	}

	if kycStatus != "" && kycStatus != "all" {
		where += " AND u.kyc_status = ?"
		arguments = append(arguments, kycStatus)
	}

	if status != "" && status != "all" {
		where += " AND u.status = ?"
		arguments = append(arguments, status)
	}

	if tariffPlan != "" && tariffPlan != "all" {
		where += " AND EXISTS (SELECT 1 FROM tariff_plan_assignments tpa WHERE tpa.user_id = u.id AND tpa.tariff_plan_id = ? AND tpa.status = 'active')"
		arguments = append(arguments, tariffPlan)
	}

	switch balanceRange {
	case "positive":
		where += " AND u.balance > 0"
	case "zero":
		where += " AND u.balance <= 0"
	case "low":
		where += " AND u.balance >= 0 AND u.balance < 100"
	case "high":
		where += " AND u.balance >= 1000"
	case "vip":
		where += " AND u.balance >= 10000"
	}

	if regFrom != "" {
		where += " AND DATE(u.created_at) >= ?"
		arguments = append(arguments, regFrom)
	}
	if regUntil != "" {
		where += " AND DATE(u.created_at) <= ?"
		arguments = append(arguments, regUntil)
	}

	// Calculate aggregates for current filtered subset
	summaryQuery := `
		SELECT
			COUNT(*),
			COALESCE(SUM(u.balance), 0),
			COALESCE(AVG(u.balance), 0),
			COALESCE(SUM(CASE WHEN u.balance > 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.balance <= 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.nid_number IS NOT NULL AND u.nid_number != '' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.kyc_status = 'verified' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.kyc_status = 'under_review' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.status = 'approved' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN u.status IN ('rejected', 'banned') THEN 1 ELSE 0 END), 0)
		FROM users u ` + where

	var summary customerSummary
	err := server.db.QueryRow(summaryQuery, arguments...).Scan(
		&summary.TotalCustomers,
		&summary.TotalBalance,
		&summary.AvgBalance,
		&summary.FundedCustomersCount,
		&summary.ZeroBalanceCount,
		&summary.WithNidCount,
		&summary.VerifiedKycCount,
		&summary.UnderReviewKycCount,
		&summary.ApprovedStatusCount,
		&summary.PendingStatusCount,
		&summary.InactiveStatusCount,
	)
	if err != nil {
		return databaseError(context, err)
	}

	summary.OtherKycCount = summary.TotalCustomers - (summary.VerifiedKycCount + summary.UnderReviewKycCount)
	if summary.OtherKycCount < 0 {
		summary.OtherKycCount = 0
	}
	if summary.TotalCustomers > 0 {
		summary.KycVerifiedRate = float64(int((float64(summary.VerifiedKycCount)/float64(summary.TotalCustomers))*1000+0.5)) / 10.0
		summary.ApprovedRate = float64(int((float64(summary.ApprovedStatusCount)/float64(summary.TotalCustomers))*1000+0.5)) / 10.0
		summary.FundedPercentage = float64(int((float64(summary.FundedCustomersCount)/float64(summary.TotalCustomers))*1000+0.5)) / 10.0
	}

	total := summary.TotalCustomers
	itemsArguments := append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT u.id, u.name, u.email, COALESCE(u.phone, ''), u.business_name, COALESCE(u.balance, 0), u.kyc_status, u.status,
			u.nid_number, u.last_login_at, COALESCE(u.created_by_admin, false),
			(SELECT tp.name FROM tariff_plan_assignments tpa JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) AS tariff_name,
			u.created_at
		FROM users u `+where+`
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`, itemsArguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]customer, 0)
	for rows.Next() {
		var item customer
		var nidNumber, tariffName sql.NullString
		var lastLoginAt sql.NullTime
		var createdByAdmin bool
		var createdAt time.Time

		if err := rows.Scan(
			&item.ID, &item.Name, &item.Email, &item.Phone, &item.BusinessName, &item.Balance, &item.KYCStatus, &item.Status,
			&nidNumber, &lastLoginAt, &createdByAdmin, &tariffName, &createdAt,
		); err != nil {
			return databaseError(context, err)
		}

		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.CreatedByAdmin = createdByAdmin
		if nidNumber.Valid && nidNumber.String != "" {
			val := nidNumber.String
			item.NIDNumber = &val
		}
		if tariffName.Valid && tariffName.String != "" {
			val := tariffName.String
			item.Tariff = &val
		}
		if lastLoginAt.Valid {
			val := lastLoginAt.Time.Format("Jan 02, 2006 03:04 PM")
			item.LastLoginAt = &val
		}
		items = append(items, item)
	}

	return context.JSON(fiber.Map{
		"data":    items,
		"meta":    listMeta{Page: page, PerPage: perPage, Total: total},
		"summary": summary,
	})
}

type transaction struct {
	ID            int64   `json:"id"`
	Customer      string  `json:"customer"`
	Reference     string  `json:"reference"`
	Direction     string  `json:"direction"`
	Amount        float64 `json:"amount"`
	BalanceBefore float64 `json:"balance_before"`
	BalanceAfter  float64 `json:"balance_after"`
	Description   *string `json:"description"`
	CreatedAt     string  `json:"created_at"`
}

func (server *Server) transactions(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	arguments := []any{}
	where := "WHERE 1 = 1"
	if search != "" {
		where += " AND (u.name LIKE ? OR t.transactionable_type LIKE ? OR t.description LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM transactions t INNER JOIN users u ON u.id = t.user_id "+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT t.id, u.name, t.transactionable_type, t.direction, t.amount, t.balance_before, t.balance_after, t.description, t.created_at
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		`+where+`
		ORDER BY t.id DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]transaction, 0)
	for rows.Next() {
		var item transaction
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.Reference, &item.Direction, &item.Amount, &item.BalanceBefore, &item.BalanceAfter, &item.Description, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type message struct {
	ID                 int64   `json:"id"`
	Customer           string  `json:"customer"`
	Recipient          string  `json:"recipient"`
	Body               string  `json:"body"`
	Status             string  `json:"status"`
	Cost               float64 `json:"cost"`
	OperatorCost       float64 `json:"operator_cost"`
	PrimaryDippingCost float64 `json:"primary_dipping_cost"`
	Source             string  `json:"source"`
	CreatedAt          string  `json:"created_at"`
}

func (server *Server) messages(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	status := strings.TrimSpace(context.Query("status"))
	arguments := []any{}
	where := "WHERE 1 = 1"
	if search != "" {
		where += " AND (u.name LIKE ? OR m.`to` LIKE ? OR m.body LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}
	if status != "" {
		where += " AND m.status = ?"
		arguments = append(arguments, status)
	}

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM messages m INNER JOIN users u ON u.id = m.user_id "+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT m.id, u.name, m.`+"`to`"+`, m.body, m.status, m.cost, m.operator_cost, m.primary_dipping_cost, m.source, m.created_at
		FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		`+where+`
		ORDER BY m.id DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]message, 0)
	for rows.Next() {
		var item message
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.Recipient, &item.Body, &item.Status, &item.Cost, &item.OperatorCost, &item.PrimaryDippingCost, &item.Source, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type batch struct {
	ID              int64   `json:"id"`
	Customer        string  `json:"customer"`
	SenderID        *string `json:"sender_id"`
	Source          string  `json:"source"`
	Status          string  `json:"status"`
	Approval        *string `json:"approval"`
	TotalRecipients int     `json:"total_recipients"`
	ValidCount      int     `json:"valid_count"`
	InvalidCount    int     `json:"invalid_count"`
	Remaining       int     `json:"remaining"`
	ProgressPercent int     `json:"progress_percent"`
	CreatedAt       string  `json:"created_at"`
}

func (server *Server) batches(context fiber.Ctx) error {
	page, perPage := pagination(context)
	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM batches").Scan(&total); err != nil {
		return databaseError(context, err)
	}

	rows, err := server.db.Query(`
		SELECT b.id, u.name, s.sender_id, b.source_type, b.status, b.approval, b.total_recipients, b.valid_count, b.invalid_count, b.created_at
		FROM batches b
		INNER JOIN users u ON u.id = b.user_id
		LEFT JOIN senders s ON s.id = b.sender_id
		ORDER BY b.id DESC
		LIMIT ? OFFSET ?
	`, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]batch, 0)
	for rows.Next() {
		var item batch
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.SenderID, &item.Source, &item.Status, &item.Approval, &item.TotalRecipients, &item.ValidCount, &item.InvalidCount, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.Remaining = max(item.TotalRecipients-item.ValidCount-item.InvalidCount, 0)
		if item.TotalRecipients > 0 {
			item.ProgressPercent = (item.ValidCount + item.InvalidCount) * 100 / item.TotalRecipients
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type kycRequest struct {
	ID           int64   `json:"id"`
	CustomerName string  `json:"customer_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	NIDNumber    *string `json:"nid_number"`
	BusinessName *string `json:"business_name"`
	TradeLicense *string `json:"trade_license"`
	Status       string  `json:"status"`
	SubmittedAt  *string `json:"submitted_at"`
}

func (server *Server) kycRequests(context fiber.Ctx) error {
	if err := server.authorizeKYCPermission(context, "kyc.review"); err != nil {
		return err
	}

	page, perPage := pagination(context)
	status := context.Query("status")
	arguments := []any{"customer"}
	where := "WHERE type = ? AND deleted_at IS NULL AND kyc_status IN ('pending', 'under_review', 'resubmit_required')"
	if status != "" {
		where += " AND kyc_status = ?"
		arguments = append(arguments, status)
	}

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM users "+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT id, name, email, phone, nid_number, business_name, trade_license_number, kyc_status, kyc_submitted_at
		FROM users `+where+`
		ORDER BY kyc_submitted_at ASC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]kycRequest, 0)
	for rows.Next() {
		var item kycRequest
		var submittedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CustomerName, &item.Email, &item.Phone, &item.NIDNumber, &item.BusinessName, &item.TradeLicense, &item.Status, &submittedAt); err != nil {
			return databaseError(context, err)
		}
		if submittedAt.Valid {
			item.SubmittedAt = apiTime(submittedAt.Time)
		}
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type deposit struct {
	ID            int64   `json:"id"`
	Customer      string  `json:"customer"`
	Amount        float64 `json:"amount"`
	Gateway       *string `json:"gateway"`
	TransactionID *string `json:"transaction_id"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

func (server *Server) deposits(context fiber.Ctx) error {
	page, perPage := pagination(context)
	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM deposits WHERE deleted_at IS NULL").Scan(&total); err != nil {
		return databaseError(context, err)
	}

	rows, err := server.db.Query(`
		SELECT d.id, u.name, d.amount, d.payment_method, d.transaction_id, d.status, d.created_at
		FROM deposits d
		INNER JOIN users u ON u.id = d.user_id
		WHERE d.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]deposit, 0)
	for rows.Next() {
		var item deposit
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.Amount, &item.Gateway, &item.TransactionID, &item.Status, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type ipWhitelistRequest struct {
	ID           int64   `json:"id"`
	CustomerID   int64   `json:"customer_id"`
	UserID       int64   `json:"user_id"`
	Customer     string  `json:"customer"`
	CustomerName string  `json:"customer_name"`
	IPAddress    string  `json:"ip_address"`
	Notes        *string `json:"notes"`
	Reason       *string `json:"reason"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

func (server *Server) ipWhitelistRequests(context fiber.Ctx) error {
	page, perPage := pagination(context)
	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM ip_whitelist_requests").Scan(&total); err != nil {
		return databaseError(context, err)
	}

	rows, err := server.db.Query(`
		SELECT w.id, w.user_id, u.name, w.ip_address, w.label, w.status, w.created_at
		FROM ip_whitelist_requests w
		INNER JOIN users u ON u.id = w.user_id
		ORDER BY w.created_at DESC
		LIMIT ? OFFSET ?
	`, perPage, (page-1)*perPage)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]ipWhitelistRequest, 0)
	for rows.Next() {
		var item ipWhitelistRequest
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.Customer, &item.IPAddress, &item.Notes, &item.Status, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CustomerID = item.UserID
		item.CustomerName = item.Customer
		item.Reason = item.Notes
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

type smsLog struct {
	ID        int64   `json:"id"`
	Customer  string  `json:"customer"`
	Recipient string  `json:"recipient"`
	Body      string  `json:"body"`
	Gateway   *string `json:"gateway"`
	Cost      float64 `json:"cost"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func (server *Server) smsLogs(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	arguments := []any{}
	where := "WHERE 1 = 1"
	if search != "" {
		where += " AND (m.`to` LIKE ? OR m.body LIKE ? OR u.name LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term)
	}

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM messages m INNER JOIN users u ON u.id = m.user_id "+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT m.id, u.name, m.`+"`to`"+`, m.body, g.name, m.cost, m.status, m.created_at
		FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		LEFT JOIN gateways g ON g.id = m.gateway_id
		`+where+`
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]smsLog, 0)
	for rows.Next() {
		var item smsLog
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.Recipient, &item.Body, &item.Gateway, &item.Cost, &item.Status, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

func listResponse(context fiber.Ctx, items any, meta listMeta) error {
	return context.Status(http.StatusOK).JSON(fiber.Map{"data": items, "meta": meta})
}
