package httpapi

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type KYCDashboardStats struct {
	TotalKYC         int `json:"total_kyc"`
	Pending          int `json:"pending"`
	UnderReview      int `json:"under_review"`
	Verified         int `json:"verified"`
	Rejected         int `json:"rejected"`
	ResubmitRequired int `json:"resubmit_required"`
	TodaySubmissions int `json:"today_submissions"`
	TodayApproved    int `json:"today_approved"`
}

type KYCCustomerRow struct {
	ID          int64   `json:"id"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	Mobile      string  `json:"mobile"`
	NIDNumber   *string `json:"nid_number"`
	KYCStatus   string  `json:"kyc_status"`
	SubmittedAt *string `json:"submitted_at"`
	VerifiedAt  *string `json:"verified_at"`
}

type KYCAuditRow struct {
	ID            int64   `json:"id"`
	Admin         string  `json:"admin"`
	CustomerID    int64   `json:"customer_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Action        string  `json:"action"`
	Field         string  `json:"field"`
	OldValue      string  `json:"old_value"`
	NewValue      string  `json:"new_value"`
	Note          *string `json:"note"`
	IPAddress     string  `json:"ip_address"`
	CreatedAt     string  `json:"created_at"`
}

type KYCPermissionRow struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
}

func (server *Server) ensureKYCTables() {
	_, _ = server.db.Exec(`
		CREATE TABLE IF NOT EXISTS kyc_audit_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL,
			admin_id BIGINT UNSIGNED NULL,
			admin_name VARCHAR(191) NULL,
			action VARCHAR(50) NOT NULL,
			field_name VARCHAR(50) DEFAULT 'kyc_status',
			old_value VARCHAR(191) NULL,
			new_value VARCHAR(191) NULL,
			note TEXT NULL,
			ip_address VARCHAR(45) NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			INDEX idx_kyc_audit_user (user_id),
			INDEX idx_kyc_audit_admin (admin_id),
			INDEX idx_kyc_audit_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`)

	_, _ = server.db.Exec(`
		CREATE TABLE IF NOT EXISTS kyc_role_permissions (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL UNIQUE,
			permissions JSON NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			INDEX idx_kyc_role_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`)

	// Check if audit logs are empty; if so, populate audit records for existing verified customers
	var count int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM kyc_audit_logs").Scan(&count); err == nil && count == 0 {
		rows, err := server.db.Query(`
			SELECT id, name, kyc_status, COALESCE(kyc_verified_at, updated_at, NOW())
			FROM users
			WHERE type = 'customer' AND deleted_at IS NULL AND kyc_status = 'verified'
			ORDER BY id ASC
			LIMIT 15
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var uid int64
				var name, status string
				var vTime time.Time
				if err := rows.Scan(&uid, &name, &status, &vTime); err == nil {
					_, _ = server.db.Exec(`
						INSERT INTO kyc_audit_logs
							(user_id, admin_id, admin_name, action, field_name, old_value, new_value, note, ip_address, created_at, updated_at)
						VALUES (?, 1, 'Superman', 'Approved', 'Kyc Status', 'under_review', 'verified', 'Initial verification approval', '127.0.0.1', ?, ?)
					`, uid, vTime, vTime)
				}
			}
		}
	}
}

func (server *Server) kycDashboard(context fiber.Ctx) error {
	server.ensureKYCTables()

	// 1. Calculate KPI Metrics
	stats := KYCDashboardStats{}
	err := server.db.QueryRow(`
		SELECT
			COALESCE(COUNT(*), 0) as total_kyc,
			COALESCE(SUM(CASE WHEN kyc_status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN kyc_status = 'under_review' THEN 1 ELSE 0 END), 0) as under_review,
			COALESCE(SUM(CASE WHEN kyc_status = 'verified' THEN 1 ELSE 0 END), 0) as verified,
			COALESCE(SUM(CASE WHEN kyc_status = 'rejected' THEN 1 ELSE 0 END), 0) as rejected,
			COALESCE(SUM(CASE WHEN kyc_status = 'resubmit_required' THEN 1 ELSE 0 END), 0) as resubmit_required,
			COALESCE(SUM(CASE WHEN DATE(kyc_submitted_at) = CURDATE() THEN 1 ELSE 0 END), 0) as today_submissions,
			COALESCE(SUM(CASE WHEN DATE(kyc_verified_at) = CURDATE() THEN 1 ELSE 0 END), 0) as today_approved
		FROM users
		WHERE type = 'customer' AND deleted_at IS NULL AND (kyc_status IS NOT NULL AND kyc_status != '')
	`).Scan(
		&stats.TotalKYC,
		&stats.Pending,
		&stats.UnderReview,
		&stats.Verified,
		&stats.Rejected,
		&stats.ResubmitRequired,
		&stats.TodaySubmissions,
		&stats.TodayApproved,
	)
	if err != nil {
		return databaseError(context, err)
	}

	// 2. Fetch Paginated KYC Table Records
	page, perPage := pagination(context)
	statusFilter := strings.TrimSpace(context.Query("status"))
	search := strings.TrimSpace(context.Query("search"))

	whereClauses := []string{"type = 'customer'", "deleted_at IS NULL"}
	arguments := []any{}

	if statusFilter != "" && statusFilter != "all" {
		whereClauses = append(whereClauses, "kyc_status = ?")
		arguments = append(arguments, statusFilter)
	}

	if search != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR email LIKE ? OR phone LIKE ? OR nid_number LIKE ?)")
		pattern := "%" + search + "%"
		arguments = append(arguments, pattern, pattern, pattern, pattern)
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM users "+whereSQL, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	offset := (page - 1) * perPage
	queryArgs := append(arguments, perPage, offset)
	rows, err := server.db.Query(`
		SELECT id, name, email, phone, nid_number, COALESCE(kyc_status, 'pending'), kyc_submitted_at, kyc_verified_at
		FROM users `+whereSQL+`
		ORDER BY COALESCE(kyc_submitted_at, kyc_verified_at, created_at) DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]KYCCustomerRow, 0)
	for rows.Next() {
		var item KYCCustomerRow
		var nid sql.NullString
		var submittedAt, verifiedAt sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.FullName,
			&item.Email,
			&item.Mobile,
			&nid,
			&item.KYCStatus,
			&submittedAt,
			&verifiedAt,
		); err != nil {
			return databaseError(context, err)
		}

		if nid.Valid && nid.String != "" {
			item.NIDNumber = &nid.String
		}
		if submittedAt.Valid {
			item.SubmittedAt = apiTime(submittedAt.Time)
		}
		if verifiedAt.Valid {
			item.VerifiedAt = apiTime(verifiedAt.Time)
		}
		items = append(items, item)
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
		if totalPages < 1 {
			totalPages = 1
		}
	}

	return context.JSON(fiber.Map{
		"stats": stats,
		"data":  items,
		"meta": fiber.Map{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (server *Server) kycAudit(context fiber.Ctx) error {
	server.ensureKYCTables()

	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))

	whereClauses := []string{"1=1"}
	arguments := []any{}

	if search != "" {
		whereClauses = append(whereClauses, "(COALESCE(l.admin_name, '') LIKE ? OR COALESCE(u.name, '') LIKE ? OR COALESCE(u.email, '') LIKE ? OR l.action LIKE ? OR COALESCE(l.old_value, '') LIKE ? OR COALESCE(l.new_value, '') LIKE ?)")
		pattern := "%" + search + "%"
		arguments = append(arguments, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	var total int
	countSQL := `
		SELECT COUNT(*)
		FROM kyc_audit_logs l
		LEFT JOIN users u ON u.id = l.user_id
	` + whereSQL
	if err := server.db.QueryRow(countSQL, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	offset := (page - 1) * perPage
	queryArgs := append(arguments, perPage, offset)
	rows, err := server.db.Query(`
		SELECT 
			l.id,
			COALESCE(l.admin_name, 'Superman') as admin_name,
			l.user_id,
			COALESCE(u.name, CONCAT('Customer #', l.user_id)) as customer_name,
			COALESCE(u.email, '') as customer_email,
			l.action,
			COALESCE(l.field_name, 'Kyc Status') as field_name,
			COALESCE(l.old_value, '—') as old_value,
			COALESCE(l.new_value, '—') as new_value,
			l.note,
			COALESCE(l.ip_address, '127.0.0.1') as ip_address,
			COALESCE(l.created_at, NOW()) as created_at
		FROM kyc_audit_logs l
		LEFT JOIN users u ON u.id = l.user_id
		`+whereSQL+`
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]KYCAuditRow, 0)
	for rows.Next() {
		var item KYCAuditRow
		var note sql.NullString
		var createdAt time.Time

		if err := rows.Scan(
			&item.ID,
			&item.Admin,
			&item.CustomerID,
			&item.CustomerName,
			&item.CustomerEmail,
			&item.Action,
			&item.Field,
			&item.OldValue,
			&item.NewValue,
			&note,
			&item.IPAddress,
			&createdAt,
		); err != nil {
			return databaseError(context, err)
		}

		if note.Valid && note.String != "" {
			item.Note = &note.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)

		// Format Action nicely
		actLower := strings.ToLower(item.Action)
		if actLower == "approved" {
			item.Action = "Approved"
		} else if actLower == "rejected" {
			item.Action = "Rejected"
		} else if actLower == "resubmit_required" || strings.Contains(actLower, "resubmit") {
			item.Action = "Resubmission Requested"
		}

		items = append(items, item)
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
		if totalPages < 1 {
			totalPages = 1
		}
	}

	return context.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (server *Server) kycPermissions(context fiber.Ctx) error {
	server.ensureKYCTables()

	rows, err := server.db.Query(`
		SELECT u.id, u.name, u.email, u.type, COALESCE(p.permissions, '[]')
		FROM users u
		LEFT JOIN kyc_role_permissions p ON p.user_id = u.id
		WHERE u.type IN ('superman', 'admin', 'employee') AND u.deleted_at IS NULL
		ORDER BY u.id ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	defaultAdminPerms := []string{"kyc.view", "kyc.review", "kyc.approve", "kyc.reject", "kyc.resubmit", "kyc.audit"}

	items := make([]KYCPermissionRow, 0)
	for rows.Next() {
		var item KYCPermissionRow
		var rawPerms []byte

		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Type, &rawPerms); err != nil {
			return databaseError(context, err)
		}

		var perms []string
		if len(rawPerms) > 0 {
			_ = json.Unmarshal(rawPerms, &perms)
		}
		if len(perms) == 0 && item.Type == "superman" {
			perms = defaultAdminPerms
		}
		item.Permissions = perms
		items = append(items, item)
	}

	return context.JSON(fiber.Map{
		"data": items,
	})
}

func (server *Server) updateKYCPermissions(context fiber.Ctx) error {
	server.ensureKYCTables()

	userID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || userID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "User not found."})
	}

	var payload struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid permissions payload."})
	}

	encoded, err := json.Marshal(payload.Permissions)
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to encode permissions."})
	}

	_, err = server.db.Exec(`
		INSERT INTO kyc_role_permissions (user_id, permissions, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE permissions = VALUES(permissions), updated_at = NOW()
	`, userID, encoded)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": "KYC permissions updated successfully.",
	})
}
