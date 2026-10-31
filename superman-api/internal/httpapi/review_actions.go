package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type kycRejectionRequest struct {
	Reason       string `json:"reason"`
	AdminRemarks string `json:"admin_remarks"`
}

type ipRejectionRequest struct {
	AdminNote string `json:"admin_note"`
}

func (server *Server) authorizeKYCPermission(context fiber.Ctx, permission string) error {
	admin := superman(context)
	var encodedPermissions []byte
	err := server.db.QueryRow("SELECT permissions FROM kyc_role_permissions WHERE user_id = ?", admin.AdminID).Scan(&encodedPermissions)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return databaseError(context, err)
	}

	var permissions []string
	if err := json.Unmarshal(encodedPermissions, &permissions); err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to verify Superman permissions."})
	}
	for _, grantedPermission := range permissions {
		if grantedPermission == permission {
			return nil
		}
	}

	return context.Status(http.StatusForbidden).JSON(fiber.Map{"message": "You do not have permission to perform this KYC action."})
}

func (server *Server) approveKYC(context fiber.Ctx) error {
	if err := server.authorizeKYCPermission(context, "kyc.approve"); err != nil {
		return err
	}

	return server.updateKYCDecision(context, "verified", "approved", "")
}

func (server *Server) rejectKYC(context fiber.Ctx) error {
	if err := server.authorizeKYCPermission(context, "kyc.reject"); err != nil {
		return err
	}

	request := kycRejectionRequest{}
	if err := json.Unmarshal(context.Body(), &request); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied data is invalid."})
	}
	request.Reason = strings.TrimSpace(request.Reason)
	request.AdminRemarks = strings.TrimSpace(request.AdminRemarks)
	if request.Reason == "" || len(request.Reason) > 1000 || len(request.AdminRemarks) > 1000 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A rejection reason of up to 1000 characters is required."})
	}

	note := "Reason: " + request.Reason
	if request.AdminRemarks != "" {
		note += " | Remarks: " + request.AdminRemarks
	}

	return server.updateKYCDecision(context, "rejected", "rejected", note)
}

func (server *Server) updateKYCDecision(context fiber.Ctx, newKYCStatus string, newUserStatus string, note string) error {
	userID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || userID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Record not found."})
	}

	transaction, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer transaction.Rollback()

	var oldStatus string
	if err := transaction.QueryRow("SELECT kyc_status FROM users WHERE id = ? AND type = 'customer' AND deleted_at IS NULL FOR UPDATE", userID).Scan(&oldStatus); err != nil {
		return databaseError(context, err)
	}
	if oldStatus == newKYCStatus {
		return context.Status(http.StatusConflict).JSON(fiber.Map{"message": "This KYC request has already been processed."})
	}

	admin := superman(context)
	if newKYCStatus == "verified" {
		_, err = transaction.Exec(`
			UPDATE users
			SET kyc_status = 'verified', status = ?, nid_verified = 1, live_photo_verified = 1,
				kyc_verified_at = NOW(), kyc_rejected_reason = NULL, kyc_resubmit_fields = NULL, updated_at = NOW()
			WHERE id = ?
		`, newUserStatus, userID)
	} else {
		reason := strings.TrimPrefix(note, "Reason: ")
		if separator := strings.Index(reason, " | Remarks: "); separator >= 0 {
			reason = reason[:separator]
		}
		_, err = transaction.Exec(`
			UPDATE users
			SET kyc_status = 'rejected', status = ?, kyc_rejected_reason = ?, updated_at = NOW()
			WHERE id = ?
		`, newUserStatus, reason, userID)
	}
	if err != nil {
		return databaseError(context, err)
	}

	action := "approved"
	if newKYCStatus == "rejected" {
		action = "rejected"
	}
	_, err = transaction.Exec(`
		INSERT INTO kyc_audit_logs
			(user_id, admin_id, admin_name, action, field_name, old_value, new_value, note, ip_address, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'kyc_status', ?, ?, ?, ?, NOW(), NOW())
	`, userID, admin.AdminID, admin.Name, action, oldStatus, newKYCStatus, nullableString(note), context.IP())
	if err != nil {
		return databaseError(context, err)
	}

	if err := transaction.Commit(); err != nil {
		return databaseError(context, err)
	}

	message := "KYC approved successfully."
	if newKYCStatus == "rejected" {
		message = "KYC rejected successfully."
	}

	return context.JSON(fiber.Map{"message": message})
}

func (server *Server) approveIPWhitelist(context fiber.Ctx) error {
	return server.updateIPWhitelistDecision(context, "approved", "")
}

func (server *Server) rejectIPWhitelist(context fiber.Ctx) error {
	request := ipRejectionRequest{}
	if err := json.Unmarshal(context.Body(), &request); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied data is invalid."})
	}
	request.AdminNote = strings.TrimSpace(request.AdminNote)
	if request.AdminNote == "" || len(request.AdminNote) > 1000 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A rejection reason of up to 1000 characters is required."})
	}

	return server.updateIPWhitelistDecision(context, "rejected", request.AdminNote)
}

func (server *Server) updateIPWhitelistDecision(context fiber.Ctx, newStatus string, adminNote string) error {
	requestID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || requestID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Record not found."})
	}

	transaction, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer transaction.Rollback()

	var userID int64
	var ipAddress string
	var oldStatus string
	if err := transaction.QueryRow(`
		SELECT user_id, ip_address, status
		FROM ip_whitelist_requests
		WHERE id = ?
		FOR UPDATE
	`, requestID).Scan(&userID, &ipAddress, &oldStatus); err != nil {
		return databaseError(context, err)
	}
	if oldStatus != "pending" {
		return context.Status(http.StatusConflict).JSON(fiber.Map{"message": "This IP whitelist request has already been processed."})
	}

	admin := superman(context)
	_, err = transaction.Exec(`
		UPDATE ip_whitelist_requests
		SET status = ?, admin_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`, newStatus, nullableString(adminNote), admin.AdminID, requestID)
	if err != nil {
		return databaseError(context, err)
	}

	oldValues, _ := json.Marshal(map[string]string{"status": oldStatus})
	newValues, _ := json.Marshal(map[string]string{"status": newStatus})
	note := "Approved by admin"
	if newStatus == "rejected" {
		note = "Reason: " + adminNote
	}
	_, err = transaction.Exec(`
		INSERT INTO ip_whitelist_audit_logs
			(ip_whitelist_request_id, user_id, actor_id, actor_name, actor_type, action, ip_address, old_values, new_values, note, source_ip, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'admin', ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, requestID, userID, admin.AdminID, admin.Name, newStatus, ipAddress, oldValues, newValues, note, context.IP())
	if err != nil {
		return databaseError(context, err)
	}

	if err := transaction.Commit(); err != nil {
		return databaseError(context, err)
	}

	message := "IP whitelist request approved successfully."
	if newStatus == "rejected" {
		message = "IP whitelist request rejected successfully."
	}

	return context.JSON(fiber.Map{"message": message})
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}
