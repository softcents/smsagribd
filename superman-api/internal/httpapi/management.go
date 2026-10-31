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

type depositPayload struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type statusPayload struct {
	Status string `json:"status"`
}

type kycResubmitPayload struct {
	Reason string `json:"reason"`
	Fields string `json:"fields"`
}

func (server *Server) depositCustomer(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	payload := depositPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	if payload.Amount == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Amount must be greater or less than zero."})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	var oldBalance float64
	var customerName string
	err = tx.QueryRow("SELECT balance, name FROM users WHERE id = ? AND type = 'customer' AND deleted_at IS NULL FOR UPDATE", id).Scan(&oldBalance, &customerName)
	if err != nil {
		return databaseError(context, err)
	}

	newBalance := oldBalance + payload.Amount
	if newBalance < 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Customer balance cannot be reduced below zero."})
	}

	_, err = tx.Exec("UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", newBalance, id)
	if err != nil {
		return databaseError(context, err)
	}

	admin := superman(context)
	txID := fmt.Sprintf("MANUAL-%d-%d", id, time.Now().Unix())

	res, err := tx.Exec(`
		INSERT INTO deposits (user_id, deposited_by_id, amount, currency, payment_method, transaction_id, status, paid_at, created_at, updated_at)
		VALUES (?, ?, ?, 'BDT', 'manual', ?, 'completed', NOW(), NOW(), NOW())
	`, id, admin.AdminID, payload.Amount, txID)
	if err != nil {
		return databaseError(context, err)
	}
	depositID, _ := res.LastInsertId()

	direction := "credit"
	desc := payload.Description
	if payload.Amount < 0 {
		direction = "debit"
		if desc == "" {
			desc = "Admin deducted balance manually"
		}
	} else if desc == "" {
		desc = "Admin added balance manually"
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\Deposit', ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, id, depositID, math.Abs(payload.Amount), direction, oldBalance, newBalance, desc)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully adjusted balance for %s by ৳%.2f", customerName, payload.Amount),
		"data": fiber.Map{
			"balance":        newBalance,
			"balance_before": oldBalance,
			"transaction_id": txID,
		},
	})
}

func (server *Server) updateCustomerStatus(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	payload := statusPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || payload.Status == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Status is required."})
	}

	_, err = server.db.Exec("UPDATE users SET status = ?, updated_at = NOW() WHERE id = ? AND type = 'customer' AND deleted_at IS NULL", payload.Status, id)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": "Customer status updated successfully.",
		"data":    fiber.Map{"status": payload.Status},
	})
}

func (server *Server) customerDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	var c customer
	var companyName, nidNumber, frontNid, backNid, livePhoto, businessAddress, tradeLicenseNum, tradeLicenseFile sql.NullString
	var kycSubmittedAt, kycVerifiedAt sql.NullTime
	var createdAt time.Time
	var apiAccess, smsAutoApproved int
	var tariffName sql.NullString
	var tariffPlanID sql.NullInt64

	err = server.db.QueryRow(`
		SELECT u.id, u.name, u.email, u.phone, u.business_name, u.company_name, u.balance, u.kyc_status, u.status,
			u.nid_number, u.front_nid, u.back_nid, u.live_photo, u.business_address, u.trade_license_number, u.trade_license_file,
			u.kyc_submitted_at, u.kyc_verified_at, COALESCE(u.api_access, 0), COALESCE(u.sms_auto_approved, 0), u.created_at,
			(SELECT tp.name FROM tariff_plan_assignments tpa JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) AS tariff_name,
			(SELECT tpa.tariff_plan_id FROM tariff_plan_assignments tpa WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) AS tariff_plan_id
		FROM users u
		WHERE u.id = ? AND u.type = 'customer' AND u.deleted_at IS NULL
	`, id).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.BusinessName, &companyName, &c.Balance, &c.KYCStatus, &c.Status,
		&nidNumber, &frontNid, &backNid, &livePhoto, &businessAddress, &tradeLicenseNum, &tradeLicenseFile,
		&kycSubmittedAt, &kycVerifiedAt, &apiAccess, &smsAutoApproved, &createdAt,
		&tariffName, &tariffPlanID,
	)
	if err != nil {
		return databaseError(context, err)
	}

	c.CreatedAt = createdAt.Format(time.RFC3339)
	if nidNumber.Valid && nidNumber.String != "" {
		val := nidNumber.String
		c.NIDNumber = &val
	}
	if tariffName.Valid && tariffName.String != "" {
		val := tariffName.String
		c.Tariff = &val
	}

	// Fetch assigned Senders
	type senderItem struct {
		ID          int64  `json:"id"`
		SenderID    string `json:"sender_id"`
		Type        string `json:"type"`
		GatewayName string `json:"gateway_name"`
	}
	senders := make([]senderItem, 0)
	sRows, err := server.db.Query(`
		SELECT s.id, s.sender_id, s.type, COALESCE(g.name, 'No Provider')
		FROM sender_user su
		JOIN senders s ON s.id = su.sender_id
		LEFT JOIN gateways g ON g.id = s.gateway_id
		WHERE su.user_id = ?
		ORDER BY s.sender_id ASC
	`, id)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var si senderItem
			if err := sRows.Scan(&si.ID, &si.SenderID, &si.Type, &si.GatewayName); err == nil {
				senders = append(senders, si)
			}
		}
	}

	// Fetch approved IP Whitelist
	type ipItem struct {
		ID        int64  `json:"id"`
		IPAddress string `json:"ip_address"`
		Status    string `json:"status"`
	}
	ipWhitelists := make([]ipItem, 0)
	ipRows, err := server.db.Query(`
		SELECT id, ip_address, status
		FROM ip_whitelist_requests
		WHERE user_id = ? AND status = 'approved'
		ORDER BY id DESC
	`, id)
	if err == nil {
		defer ipRows.Close()
		for ipRows.Next() {
			var ip ipItem
			if err := ipRows.Scan(&ip.ID, &ip.IPAddress, &ip.Status); err == nil {
				ipWhitelists = append(ipWhitelists, ip)
			}
		}
	}

	// Fetch recent deposits/transactions
	type txItem struct {
		ID            int64   `json:"id"`
		Amount        float64 `json:"amount"`
		Direction     string  `json:"direction"`
		BalanceBefore float64 `json:"balance_before"`
		BalanceAfter  float64 `json:"balance_after"`
		Description   *string `json:"description"`
		CreatedAt     string  `json:"created_at"`
	}
	recentTx := make([]txItem, 0)
	txRows, err := server.db.Query(`
		SELECT id, amount, direction, balance_before, balance_after, description, created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY id DESC
		LIMIT 10
	`, id)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var tx txItem
			var cat time.Time
			if err := txRows.Scan(&tx.ID, &tx.Amount, &tx.Direction, &tx.BalanceBefore, &tx.BalanceAfter, &tx.Description, &cat); err == nil {
				tx.CreatedAt = cat.Format(time.RFC3339)
				recentTx = append(recentTx, tx)
			}
		}
	}

	var tariffID *int64
	if tariffPlanID.Valid {
		val := tariffPlanID.Int64
		tariffID = &val
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"customer":             c,
			"company_name":         nullStringVal(companyName),
			"nid_number":           nullStringVal(nidNumber),
			"front_nid":            nullStringVal(frontNid),
			"back_nid":             nullStringVal(backNid),
			"live_photo":           nullStringVal(livePhoto),
			"business_address":     nullStringVal(businessAddress),
			"trade_license_number": nullStringVal(tradeLicenseNum),
			"trade_license_file":   nullStringVal(tradeLicenseFile),
			"kyc_submitted_at":     nullTimeVal(kycSubmittedAt),
			"kyc_verified_at":      nullTimeVal(kycVerifiedAt),
			"api_access":           apiAccess == 1,
			"sms_auto_approved":    smsAutoApproved == 1,
			"tariff_plan_id":       tariffID,
			"senders":              senders,
			"ip_whitelists":        ipWhitelists,
			"recent_transactions":  recentTx,
		},
	})
}

type updateUserPayload struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Status          string `json:"status"`
	BusinessName    string `json:"business_name"`
	APIAccess       *bool  `json:"api_access"`
	SMSAutoApproved *bool  `json:"sms_auto_approved"`
	TariffPlanID    *int64 `json:"tariff_plan_id"`
}

func (server *Server) updateCustomer(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	payload := updateUserPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET name = ?, email = ?, phone = ?, business_name = ?, status = ?, updated_at = NOW()`
	args := []any{payload.Name, payload.Email, payload.Phone, payload.BusinessName, payload.Status}

	if payload.APIAccess != nil {
		val := 0
		if *payload.APIAccess {
			val = 1
		}
		query += ", api_access = ?"
		args = append(args, val)
	}
	if payload.SMSAutoApproved != nil {
		val := 0
		if *payload.SMSAutoApproved {
			val = 1
		}
		query += ", sms_auto_approved = ?"
		args = append(args, val)
	}

	query += " WHERE id = ? AND type = 'customer' AND deleted_at IS NULL"
	args = append(args, id)

	if _, err := tx.Exec(query, args...); err != nil {
		return databaseError(context, err)
	}

	if payload.TariffPlanID != nil && *payload.TariffPlanID > 0 {
		_, _ = tx.Exec("UPDATE tariff_plan_assignments SET status = 'inactive', updated_at = NOW() WHERE user_id = ?", id)
		_, _ = tx.Exec("INSERT INTO tariff_plan_assignments (user_id, tariff_plan_id, status, effective_from, created_at, updated_at) VALUES (?, ?, 'active', NOW(), NOW(), NOW())", id, *payload.TariffPlanID)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Customer updated successfully."})
}

type spSummary struct {
	TotalSPs       int     `json:"total_sps"`
	TotalBalance   float64 `json:"total_balance"`
	TotalCustomers int     `json:"total_customers"`
	ActiveSPsCount int     `json:"active_sps_count"`
	ActiveRate     float64 `json:"active_rate"`
}

type spReseller struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	BusinessName  *string `json:"business_name"`
	Balance       float64 `json:"balance"`
	Status        string  `json:"status"`
	WholesalePlan *string `json:"wholesale_plan"`
	CustomerCount int     `json:"customer_count"`
	CreatedAt     string  `json:"created_at"`
}

func (server *Server) sps(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	status := strings.TrimSpace(context.Query("status"))
	arguments := []any{"sp"}
	where := "WHERE u.type = ? AND u.deleted_at IS NULL"

	if search != "" {
		where += " AND (u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ? OR u.business_name LIKE ?)"
		term := "%" + search + "%"
		arguments = append(arguments, term, term, term, term)
	}
	if status != "" && status != "all" {
		where += " AND u.status = ?"
		arguments = append(arguments, status)
	}

	summaryQuery := `
		SELECT
			COUNT(*),
			COALESCE(SUM(u.balance), 0),
			COALESCE((SELECT COUNT(*) FROM users c WHERE c.sp_id = u.id AND c.deleted_at IS NULL), 0),
			COALESCE(SUM(CASE WHEN u.status = 'approved' THEN 1 ELSE 0 END), 0)
		FROM users u ` + where

	var summary spSummary
	var totalCustSum int
	_ = server.db.QueryRow(summaryQuery, arguments...).Scan(&summary.TotalSPs, &summary.TotalBalance, &totalCustSum, &summary.ActiveSPsCount)
	summary.TotalCustomers = totalCustSum
	if summary.TotalSPs > 0 {
		summary.ActiveRate = float64(int((float64(summary.ActiveSPsCount)/float64(summary.TotalSPs))*1000+0.5)) / 10.0
	}

	total := summary.TotalSPs
	itemArgs := append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT u.id, u.name, u.email, u.phone, u.business_name, u.balance, u.status,
			(SELECT COUNT(*) FROM users c WHERE c.sp_id = u.id AND c.deleted_at IS NULL) as customer_count,
			(SELECT tp.name FROM tariff_plan_assignments tpa JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) as wholesale_plan,
			u.created_at
		FROM users u `+where+`
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`, itemArgs...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]spReseller, 0)
	for rows.Next() {
		var item spReseller
		var wholesalePlan sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Phone, &item.BusinessName, &item.Balance, &item.Status, &item.CustomerCount, &wholesalePlan, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		if wholesalePlan.Valid && wholesalePlan.String != "" {
			val := wholesalePlan.String
			item.WholesalePlan = &val
		}
		items = append(items, item)
	}

	return context.JSON(fiber.Map{
		"data":    items,
		"meta":    listMeta{Page: page, PerPage: perPage, Total: total},
		"summary": summary,
	})
}

func (server *Server) spDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Reseller not found."})
	}

	var sp spReseller
	var apiAccess, smsAutoApproved int
	var wholesalePlan sql.NullString
	var wholesalePlanID sql.NullInt64
	var createdAt time.Time

	err = server.db.QueryRow(`
		SELECT u.id, u.name, u.email, u.phone, u.business_name, u.balance, u.status,
			COALESCE(u.api_access, 0), COALESCE(u.sms_auto_approved, 0),
			(SELECT COUNT(*) FROM users c WHERE c.sp_id = u.id AND c.deleted_at IS NULL) as customer_count,
			(SELECT tp.name FROM tariff_plan_assignments tpa JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) as wholesale_plan,
			(SELECT tpa.tariff_plan_id FROM tariff_plan_assignments tpa WHERE tpa.user_id = u.id AND tpa.status = 'active' LIMIT 1) as wholesale_plan_id,
			u.created_at
		FROM users u
		WHERE u.id = ? AND u.type = 'sp' AND u.deleted_at IS NULL
	`, id).Scan(
		&sp.ID, &sp.Name, &sp.Email, &sp.Phone, &sp.BusinessName, &sp.Balance, &sp.Status,
		&apiAccess, &smsAutoApproved, &sp.CustomerCount, &wholesalePlan, &wholesalePlanID, &createdAt,
	)
	if err != nil {
		return databaseError(context, err)
	}

	sp.CreatedAt = createdAt.Format(time.RFC3339)
	if wholesalePlan.Valid && wholesalePlan.String != "" {
		val := wholesalePlan.String
		sp.WholesalePlan = &val
	}

	type spCustItem struct {
		ID        int64   `json:"id"`
		Name      string  `json:"name"`
		Email     string  `json:"email"`
		Phone     string  `json:"phone"`
		Balance   float64 `json:"balance"`
		Status    string  `json:"status"`
		CreatedAt string  `json:"created_at"`
	}
	customers := make([]spCustItem, 0)
	cRows, err := server.db.Query(`
		SELECT id, name, email, phone, balance, status, created_at
		FROM users
		WHERE sp_id = ? AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 20
	`, id)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var ci spCustItem
			var cat time.Time
			if err := cRows.Scan(&ci.ID, &ci.Name, &ci.Email, &ci.Phone, &ci.Balance, &ci.Status, &cat); err == nil {
				ci.CreatedAt = cat.Format(time.RFC3339)
				customers = append(customers, ci)
			}
		}
	}

	type txItem struct {
		ID            int64   `json:"id"`
		Amount        float64 `json:"amount"`
		Direction     string  `json:"direction"`
		BalanceBefore float64 `json:"balance_before"`
		BalanceAfter  float64 `json:"balance_after"`
		Description   *string `json:"description"`
		CreatedAt     string  `json:"created_at"`
	}
	recentTx := make([]txItem, 0)
	txRows, err := server.db.Query(`
		SELECT id, amount, direction, balance_before, balance_after, description, created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY id DESC
		LIMIT 10
	`, id)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var tx txItem
			var cat time.Time
			if err := txRows.Scan(&tx.ID, &tx.Amount, &tx.Direction, &tx.BalanceBefore, &tx.BalanceAfter, &tx.Description, &cat); err == nil {
				tx.CreatedAt = cat.Format(time.RFC3339)
				recentTx = append(recentTx, tx)
			}
		}
	}

	var wPlanID *int64
	if wholesalePlanID.Valid {
		val := wholesalePlanID.Int64
		wPlanID = &val
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"sp":                  sp,
			"api_access":          apiAccess == 1,
			"sms_auto_approved":   smsAutoApproved == 1,
			"wholesale_plan_id":   wPlanID,
			"customers":           customers,
			"recent_transactions": recentTx,
		},
	})
}

func (server *Server) updateSP(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Reseller not found."})
	}

	payload := updateUserPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET name = ?, email = ?, phone = ?, business_name = ?, status = ?, updated_at = NOW()`
	args := []any{payload.Name, payload.Email, payload.Phone, payload.BusinessName, payload.Status}

	if payload.APIAccess != nil {
		val := 0
		if *payload.APIAccess {
			val = 1
		}
		query += ", api_access = ?"
		args = append(args, val)
	}
	if payload.SMSAutoApproved != nil {
		val := 0
		if *payload.SMSAutoApproved {
			val = 1
		}
		query += ", sms_auto_approved = ?"
		args = append(args, val)
	}

	query += " WHERE id = ? AND type = 'sp' AND deleted_at IS NULL"
	args = append(args, id)

	if _, err := tx.Exec(query, args...); err != nil {
		return databaseError(context, err)
	}

	if payload.TariffPlanID != nil && *payload.TariffPlanID > 0 {
		_, _ = tx.Exec("UPDATE tariff_plan_assignments SET status = 'inactive', updated_at = NOW() WHERE user_id = ?", id)
		_, _ = tx.Exec("INSERT INTO tariff_plan_assignments (user_id, tariff_plan_id, status, effective_from, created_at, updated_at) VALUES (?, ?, 'active', NOW(), NOW(), NOW())", id, *payload.TariffPlanID)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Reseller updated successfully."})
}

func (server *Server) depositSP(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Reseller not found."})
	}

	payload := depositPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || payload.Amount == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A non-zero amount is required."})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	var oldBalance float64
	var spName string
	err = tx.QueryRow("SELECT balance, name FROM users WHERE id = ? AND type = 'sp' AND deleted_at IS NULL FOR UPDATE", id).Scan(&oldBalance, &spName)
	if err != nil {
		return databaseError(context, err)
	}

	newBalance := oldBalance + payload.Amount
	if newBalance < 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Reseller balance cannot be reduced below zero."})
	}

	_, err = tx.Exec("UPDATE users SET balance = ?, updated_at = NOW() WHERE id = ?", newBalance, id)
	if err != nil {
		return databaseError(context, err)
	}

	admin := superman(context)
	txID := fmt.Sprintf("SP-MANUAL-%d-%d", id, time.Now().Unix())

	res, err := tx.Exec(`
		INSERT INTO deposits (user_id, deposited_by_id, amount, currency, payment_method, transaction_id, status, paid_at, created_at, updated_at)
		VALUES (?, ?, ?, 'BDT', 'manual', ?, 'completed', NOW(), NOW(), NOW())
	`, id, admin.AdminID, payload.Amount, txID)
	if err != nil {
		return databaseError(context, err)
	}
	depositID, _ := res.LastInsertId()

	direction := "credit"
	if payload.Amount < 0 {
		direction = "debit"
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (user_id, transactionable_type, transactionable_id, amount, direction, balance_before, balance_after, description, created_at, updated_at)
		VALUES (?, 'App\\Models\\Deposit', ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, id, depositID, math.Abs(payload.Amount), direction, oldBalance, newBalance, payload.Description)
	if err != nil {
		return databaseError(context, err)
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully adjusted balance for %s by ৳%.2f", spName, payload.Amount),
		"data":    fiber.Map{"balance": newBalance},
	})
}

func (server *Server) updateSPStatus(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Reseller not found."})
	}

	payload := statusPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || payload.Status == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Status is required."})
	}

	_, err = server.db.Exec("UPDATE users SET status = ?, updated_at = NOW() WHERE id = ? AND type = 'sp' AND deleted_at IS NULL", payload.Status, id)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": "Reseller status updated successfully.",
		"data":    fiber.Map{"status": payload.Status},
	})
}

type tariffPlanItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Currency    string  `json:"currency"`
	Description *string `json:"description"`
	RatesCount  int     `json:"rates_count"`
	CreatedAt   string  `json:"created_at"`
}

func (server *Server) tariffPlans(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT tp.id, tp.name, tp.slug, tp.type, tp.status, tp.currency, tp.description,
			(SELECT COUNT(*) FROM tariff_rates tr WHERE tr.tariff_plan_id = tp.id) as rates_count,
			tp.created_at
		FROM tariff_plans tp
		ORDER BY tp.id ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]tariffPlanItem, 0)
	for rows.Next() {
		var item tariffPlanItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Type, &item.Status, &item.Currency, &item.Description, &item.RatesCount, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

type rateItem struct {
	ID              int64   `json:"id"`
	PlanName        string  `json:"plan_name"`
	OperatorName    string  `json:"operator_name"`
	MaskingPrice    float64 `json:"masking_price"`
	NonMaskingPrice float64 `json:"non_masking_price"`
	EffectiveFrom   string  `json:"effective_from"`
	Notes           *string `json:"notes"`
}

func (server *Server) rates(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT tr.id, tp.name, o.name, tr.masking_price, tr.non_masking_price, tr.effective_from, tr.notes
		FROM tariff_rates tr
		INNER JOIN tariff_plans tp ON tp.id = tr.tariff_plan_id
		INNER JOIN operators o ON o.id = tr.operator_id
		ORDER BY tp.name ASC, o.name ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]rateItem, 0)
	for rows.Next() {
		var item rateItem
		var effectiveFrom time.Time
		if err := rows.Scan(&item.ID, &item.PlanName, &item.OperatorName, &item.MaskingPrice, &item.NonMaskingPrice, &effectiveFrom, &item.Notes); err != nil {
			return databaseError(context, err)
		}
		item.EffectiveFrom = effectiveFrom.Format(time.RFC3339)
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) countries(context fiber.Ctx) error {
	rows, err := server.db.Query("SELECT id, name, code, iso_alpha2, is_active FROM countries ORDER BY name ASC")
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	type countryItem struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Code     string `json:"code"`
		ISO2     string `json:"iso2"`
		IsActive bool   `json:"is_active"`
	}

	items := make([]countryItem, 0)
	for rows.Next() {
		var item countryItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.ISO2, &item.IsActive); err != nil {
			return databaseError(context, err)
		}
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) operators(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT o.id, o.name, COALESCE(o.codes, ''), o.masking_price, o.non_masking_price, o.is_active, COALESCE(c.name, 'Bangladesh')
		FROM operators o
		LEFT JOIN countries c ON c.id = o.country_id
		ORDER BY o.name ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	type operatorItem struct {
		ID              int64   `json:"id"`
		Name            string  `json:"name"`
		Codes           string  `json:"codes"`
		MaskingPrice    float64 `json:"masking_price"`
		NonMaskingPrice float64 `json:"non_masking_price"`
		IsActive        bool    `json:"is_active"`
		CountryName     string  `json:"country_name"`
	}

	items := make([]operatorItem, 0)
	for rows.Next() {
		var item operatorItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Codes, &item.MaskingPrice, &item.NonMaskingPrice, &item.IsActive, &item.CountryName); err != nil {
			return databaseError(context, err)
		}
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) approvalsOverview(context fiber.Ctx) error {
	var pendingCustomers, pendingResellers, pendingSenders, pendingIPs, pendingKYC int

	err := server.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM users WHERE type = 'customer' AND status = 'pending' AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM users WHERE type = 'sp' AND status = 'pending' AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM senders WHERE approval_status = 'pending' AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM ip_whitelist_requests WHERE status = 'pending'),
			(SELECT COUNT(*) FROM users WHERE kyc_status IN ('pending', 'under_review') AND deleted_at IS NULL)
	`).Scan(&pendingCustomers, &pendingResellers, &pendingSenders, &pendingIPs, &pendingKYC)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"pending_customers": pendingCustomers,
			"pending_resellers": pendingResellers,
			"pending_senders":   pendingSenders,
			"pending_ips":       pendingIPs,
			"pending_kyc":       pendingKYC,
			"total_pending":     pendingCustomers + pendingResellers + pendingSenders + pendingIPs + pendingKYC,
		},
	})
}

func (server *Server) kycDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "KYC record not found."})
	}

	var req kycRequest
	var nidNum, bizName, tradeLic, bizAddr sql.NullString
	var frontNid, backNid, livePhoto, tradeLicFile, rejReason, resubFields sql.NullString
	var submittedAt, verifiedAt sql.NullTime

	err = server.db.QueryRow(`
		SELECT id, name, email, phone, nid_number, business_name, trade_license_number, business_address,
			front_nid, back_nid, live_photo, trade_license_file,
			kyc_status, kyc_rejected_reason, kyc_resubmit_fields, kyc_submitted_at, kyc_verified_at
		FROM users
		WHERE id = ? AND type = 'customer' AND deleted_at IS NULL
	`, id).Scan(
		&req.ID, &req.CustomerName, &req.Email, &req.Phone, &nidNum, &bizName, &tradeLic, &bizAddr,
		&frontNid, &backNid, &livePhoto, &tradeLicFile,
		&req.Status, &rejReason, &resubFields, &submittedAt, &verifiedAt,
	)
	if err != nil {
		return databaseError(context, err)
	}

	req.NIDNumber = nullStringVal(nidNum)
	req.BusinessName = nullStringVal(bizName)
	req.TradeLicense = nullStringVal(tradeLic)
	if submittedAt.Valid {
		req.SubmittedAt = apiTime(submittedAt.Time)
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"request":              req,
			"business_address":     nullStringVal(bizAddr),
			"front_nid":            nullStringVal(frontNid),
			"back_nid":             nullStringVal(backNid),
			"live_photo":           nullStringVal(livePhoto),
			"trade_license_file":   nullStringVal(tradeLicFile),
			"rejection_reason":     nullStringVal(rejReason),
			"resubmit_fields":      nullStringVal(resubFields),
			"kyc_verified_at":      nullTimeVal(verifiedAt),
		},
	})
}

func (server *Server) resubmitKYC(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}

	payload := kycResubmitPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || strings.TrimSpace(payload.Reason) == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A reason for resubmission is required."})
	}

	admin := superman(context)
	_, err = server.db.Exec(`
		UPDATE users
		SET kyc_status = 'resubmit_required', kyc_rejected_reason = ?, kyc_resubmit_fields = ?, updated_at = NOW()
		WHERE id = ? AND type = 'customer' AND deleted_at IS NULL
	`, strings.TrimSpace(payload.Reason), strings.TrimSpace(payload.Fields), id)
	if err != nil {
		return databaseError(context, err)
	}

	_, _ = server.db.Exec(`
		INSERT INTO kyc_audit_logs
			(user_id, admin_id, admin_name, action, field_name, old_value, new_value, note, ip_address, created_at, updated_at)
		VALUES (?, ?, ?, 'resubmit_required', 'kyc_status', 'under_review', 'resubmit_required', ?, ?, NOW(), NOW())
	`, id, admin.AdminID, admin.Name, payload.Reason, context.IP())

	return context.JSON(fiber.Map{"message": "KYC resubmission requested successfully."})
}

func nullStringVal(ns sql.NullString) *string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return &ns.String
}

func nullTimeVal(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	formatted := nt.Time.Format(time.RFC3339)
	return &formatted
}
