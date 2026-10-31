package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type senderUserRef struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email"`
}

type senderListItem struct {
	ID              int64           `json:"id"`
	SenderID        string          `json:"sender_id"`
	Type            string          `json:"type"`
	MaskingCategory string          `json:"masking_category"`
	IsEnabled       bool            `json:"is_enabled"`
	ApprovalStatus  string          `json:"approval_status"`
	RejectionReason *string         `json:"rejection_reason"`
	GatewayID       *int64          `json:"gateway_id"`
	GatewayName     *string         `json:"gateway_name"`
	RequestedByID   *int64          `json:"requested_by_id"`
	RequestedByName *string         `json:"requested_by_name"`
	ReviewedAt      *string         `json:"reviewed_at"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       *string         `json:"updated_at"`
	Users           []senderUserRef `json:"users"`
	UsersCount      int             `json:"users_count"`
	MessagesCount   int             `json:"messages_count"`
}

type senderCreatePayload struct {
	SenderID  string   `json:"sender_id"`
	SenderIDs []string `json:"sender_ids"`
	GatewayID int64    `json:"gateway_id"`
	Type      string   `json:"type"`
	IsEnabled *bool    `json:"is_enabled"`
	UserIDs   []int64  `json:"user_ids"`
}

type senderUpdatePayload struct {
	SenderID       string   `json:"sender_id"`
	GatewayID      *int64   `json:"gateway_id"`
	Type           string   `json:"type"`
	IsEnabled      *bool    `json:"is_enabled"`
	ApprovalStatus string   `json:"approval_status"`
	UserIDs        *[]int64 `json:"user_ids"`
}

type senderAssignmentPayload struct {
	UserIDs []int64 `json:"user_ids"`
}

type senderEnabledPayload struct {
	IsEnabled bool `json:"is_enabled"`
}

type senderApprovePayload struct {
	GatewayID int64 `json:"gateway_id"`
}

type senderRejectPayload struct {
	Reason string `json:"reason"`
}

type senderBulkPayload struct {
	IDs []int64 `json:"ids"`
}

// senderTypeClause maps the masking/non-masking vocabulary used across the panel
// onto the three values the senders table actually stores.
func senderTypeClause(value string) (string, []any) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "masking", "alpha", "alphanumeric+shortcode":
		return "s.type IN ('alphanumeric', 'shortcode')", nil
	case "non_masking", "non-masking":
		return "s.type = 'numeric'", nil
	case "numeric", "alphanumeric", "shortcode":
		return "s.type = ?", []any{strings.ToLower(strings.TrimSpace(value))}
	default:
		return "", nil
	}
}

func normalizeSenderType(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	switch strings.ReplaceAll(trimmed, "_", "-") {
	case "non-masking", "number":
		return "numeric"
	case "short-code":
		return "shortcode"
	case "masking", "alpha", "alpha-numeric":
		return "alphanumeric"
	case "":
		return "alphanumeric"
	}

	return trimmed
}

func senderMaskingCategory(senderType string) string {
	if senderType == "numeric" {
		return "non-masking"
	}

	return "masking"
}

func placeholders(values []int64) (string, []any) {
	if len(values) == 0 {
		return "", nil
	}
	bindings := make([]any, 0, len(values))
	for _, value := range values {
		bindings = append(bindings, value)
	}

	return "(" + strings.TrimSuffix(strings.Repeat("?, ", len(values)), ", ") + ")", bindings
}

func (server *Server) senderUsers(senderIDs []int64) (map[int64][]senderUserRef, error) {
	refs := make(map[int64][]senderUserRef, len(senderIDs))
	if len(senderIDs) == 0 {
		return refs, nil
	}

	group, bindings := placeholders(senderIDs)
	if group == "" {
		return refs, nil
	}

	rows, err := server.db.Query(`
		SELECT su.sender_id, u.id, u.name, u.email
		FROM sender_user su
		INNER JOIN users u ON u.id = su.user_id
		WHERE su.sender_id IN `+group+`
		ORDER BY u.name ASC
	`, bindings...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var senderID int64
		var ref senderUserRef
		var email sql.NullString
		if err := rows.Scan(&senderID, &ref.ID, &ref.Name, &email); err != nil {
			return nil, err
		}
		if email.Valid {
			ref.Email = &email.String
		}
		refs[senderID] = append(refs[senderID], ref)
	}

	return refs, rows.Err()
}

func (server *Server) senders(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	status := strings.TrimSpace(context.Query("status"))
	assignment := strings.ToLower(strings.TrimSpace(context.Query("assignment")))
	gatewayID, _ := strconv.ParseInt(context.Query("gateway_id"), 10, 64)

	arguments := []any{}
	where := "WHERE s.deleted_at IS NULL"

	if clause, bindings := senderTypeClause(context.Query("type")); clause != "" {
		where += " AND " + clause
		arguments = append(arguments, bindings...)
	}
	if gatewayID > 0 {
		where += " AND s.gateway_id = ?"
		arguments = append(arguments, gatewayID)
	}
	if status != "" && status != "all" {
		where += " AND s.approval_status = ?"
		arguments = append(arguments, status)
	}
	switch assignment {
	case "used":
		where += " AND EXISTS (SELECT 1 FROM sender_user su WHERE su.sender_id = s.id)"
	case "unused":
		where += " AND NOT EXISTS (SELECT 1 FROM sender_user su WHERE su.sender_id = s.id)"
	}
	if search != "" {
		term := "%" + search + "%"
		where += ` AND (s.sender_id LIKE ?
			OR g.name LIKE ?
			OR u.name LIKE ?
			OR EXISTS (SELECT 1 FROM sender_user su2 INNER JOIN users cu ON cu.id = su2.user_id
				WHERE su2.sender_id = s.id AND (cu.name LIKE ? OR cu.business_name LIKE ? OR cu.email LIKE ?)))`
		arguments = append(arguments, term, term, term, term, term, term)
	}

	var total int
	if err := server.db.QueryRow(`
		SELECT COUNT(*)
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		LEFT JOIN users u ON u.id = s.requested_by
		`+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT s.id, s.sender_id, s.type, s.is_enabled, s.approval_status, s.rejection_reason,
			s.gateway_id, g.name, s.requested_by, u.name, s.reviewed_at, s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM sender_user su WHERE su.sender_id = s.id),
			(SELECT COUNT(*) FROM messages m WHERE m.sender_id = s.id)
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		LEFT JOIN users u ON u.id = s.requested_by
		`+where+`
		ORDER BY s.id DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]senderListItem, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var item senderListItem
		var senderType string
		var gatewayIDValue sql.NullInt64
		var gatewayName sql.NullString
		var requestedByID sql.NullInt64
		var requestedByName sql.NullString
		var rejectionReason sql.NullString
		var reviewedAt, updatedAt sql.NullTime
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.SenderID, &senderType, &item.IsEnabled, &item.ApprovalStatus, &rejectionReason,
			&gatewayIDValue, &gatewayName, &requestedByID, &requestedByName, &reviewedAt, &createdAt, &updatedAt,
			&item.UsersCount, &item.MessagesCount,
		); err != nil {
			return databaseError(context, err)
		}
		item.Type = senderType
		item.MaskingCategory = senderMaskingCategory(senderType)
		if rejectionReason.Valid {
			item.RejectionReason = &rejectionReason.String
		}
		if gatewayIDValue.Valid {
			item.GatewayID = &gatewayIDValue.Int64
		}
		if gatewayName.Valid {
			item.GatewayName = &gatewayName.String
		}
		if requestedByID.Valid {
			item.RequestedByID = &requestedByID.Int64
		}
		if requestedByName.Valid {
			item.RequestedByName = &requestedByName.String
		}
		if reviewedAt.Valid {
			item.ReviewedAt = apiTime(reviewedAt.Time)
		}
		if updatedAt.Valid {
			item.UpdatedAt = apiTime(updatedAt.Time)
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, item)
		ids = append(ids, item.ID)
	}
	if err := rows.Err(); err != nil {
		return databaseError(context, err)
	}

	refs, err := server.senderUsers(ids)
	if err != nil {
		return databaseError(context, err)
	}
	for index := range items {
		if users, ok := refs[items[index].ID]; ok {
			items[index].Users = users
		} else {
			items[index].Users = []senderUserRef{}
		}
	}

	totalPages := (total + perPage - 1) / perPage

	return context.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{"page": page, "per_page": perPage, "total": total, "total_pages": totalPages},
	})
}

func (server *Server) senderDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	var item senderListItem
	var senderType string
	var gatewayIDValue sql.NullInt64
	var gatewayName sql.NullString
	var requestedByID sql.NullInt64
	var requestedByName sql.NullString
	var rejectionReason sql.NullString
	var reviewedAt, updatedAt, createdAt sql.NullTime
	err = server.db.QueryRow(`
		SELECT s.id, s.sender_id, s.type, s.is_enabled, s.approval_status, s.rejection_reason,
			s.gateway_id, g.name, s.requested_by, u.name, s.reviewed_at, s.created_at, s.updated_at
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		LEFT JOIN users u ON u.id = s.requested_by
		WHERE s.id = ? AND s.deleted_at IS NULL
	`, id).Scan(
		&item.ID, &item.SenderID, &senderType, &item.IsEnabled, &item.ApprovalStatus, &rejectionReason,
		&gatewayIDValue, &gatewayName, &requestedByID, &requestedByName, &reviewedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return databaseError(context, err)
	}
	item.Type = senderType
	item.MaskingCategory = senderMaskingCategory(senderType)
	if rejectionReason.Valid {
		item.RejectionReason = &rejectionReason.String
	}
	if gatewayIDValue.Valid {
		item.GatewayID = &gatewayIDValue.Int64
	}
	if gatewayName.Valid {
		item.GatewayName = &gatewayName.String
	}
	if requestedByID.Valid {
		item.RequestedByID = &requestedByID.Int64
	}
	if requestedByName.Valid {
		item.RequestedByName = &requestedByName.String
	}
	if reviewedAt.Valid {
		item.ReviewedAt = apiTime(reviewedAt.Time)
	}
	if updatedAt.Valid {
		item.UpdatedAt = apiTime(updatedAt.Time)
	}
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	refs, err := server.senderUsers([]int64{item.ID})
	if err != nil {
		return databaseError(context, err)
	}
	item.Users = refs[item.ID]
	if item.Users == nil {
		item.Users = []senderUserRef{}
	}
	item.UsersCount = len(item.Users)
	_ = server.db.QueryRow("SELECT COUNT(*) FROM messages WHERE sender_id = ?", item.ID).Scan(&item.MessagesCount)

	return context.JSON(fiber.Map{"data": item})
}

func (server *Server) createSender(context fiber.Ctx) error {
	payload := senderCreatePayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid sender payload."})
	}

	senderType := normalizeSenderType(payload.Type)
	isEnabled := true
	if payload.IsEnabled != nil {
		isEnabled = *payload.IsEnabled
	}

	requested := []string{}
	for _, candidate := range append([]string{payload.SenderID}, payload.SenderIDs...) {
		trimmed := strings.TrimSpace(candidate)
		if trimmed != "" {
			requested = append(requested, trimmed)
		}
	}
	if len(requested) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A sender ID is required."})
	}
	if payload.GatewayID <= 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A gateway is required."})
	}

	var gatewayExists int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM gateways WHERE id = ? AND deleted_at IS NULL", payload.GatewayID).Scan(&gatewayExists); err != nil {
		return databaseError(context, err)
	}
	if gatewayExists == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The selected gateway no longer exists."})
	}

	created := 0
	updated := 0
	lastID := int64(0)
	for _, senderID := range requested {
		var existingID int64
		err := server.db.QueryRow(
			"SELECT id FROM senders WHERE sender_id = ? AND gateway_id = ? LIMIT 1",
			senderID, payload.GatewayID,
		).Scan(&existingID)
		switch {
		case err == nil:
			if _, err := server.db.Exec(`
				UPDATE senders SET type = ?, is_enabled = ?, deleted_at = NULL, updated_at = NOW() WHERE id = ?
			`, senderType, isEnabled, existingID); err != nil {
				return databaseError(context, err)
			}
			lastID = existingID
			updated++
		default:
			result, err := server.db.Exec(`
				INSERT INTO senders (gateway_id, sender_id, type, is_enabled, approval_status, created_at, updated_at)
				VALUES (?, ?, ?, ?, 'approved', NOW(), NOW())
			`, payload.GatewayID, senderID, senderType, isEnabled)
			if err != nil {
				return databaseError(context, err)
			}
			newID, _ := result.LastInsertId()
			lastID = newID
			created++
		}

		if len(payload.UserIDs) > 0 && lastID > 0 {
			if err := server.attachSenderUsers(lastID, payload.UserIDs); err != nil {
				return databaseError(context, err)
			}
		}
	}

	message := fmt.Sprintf("%d sender ID(s) created.", created)
	if updated > 0 {
		message = fmt.Sprintf("%d sender ID(s) created, %d updated.", created, updated)
	}

	return context.Status(http.StatusCreated).JSON(fiber.Map{"message": message, "data": fiber.Map{
		"created": created,
		"updated": updated,
		"id":      lastID,
	}})
}

func (server *Server) updateSender(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	var existing int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM senders WHERE id = ? AND deleted_at IS NULL", id).Scan(&existing); err != nil {
		return databaseError(context, err)
	}
	if existing == 0 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	payload := senderUpdatePayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid sender payload."})
	}

	sets := []string{}
	arguments := []any{}
	if senderID := strings.TrimSpace(payload.SenderID); senderID != "" {
		sets = append(sets, "sender_id = ?")
		arguments = append(arguments, senderID)
	}
	if payload.GatewayID != nil && *payload.GatewayID > 0 {
		sets = append(sets, "gateway_id = ?")
		arguments = append(arguments, *payload.GatewayID)
	}
	if strings.TrimSpace(payload.Type) != "" {
		sets = append(sets, "type = ?")
		arguments = append(arguments, normalizeSenderType(payload.Type))
	}
	if payload.IsEnabled != nil {
		sets = append(sets, "is_enabled = ?")
		arguments = append(arguments, *payload.IsEnabled)
	}
	if status := strings.TrimSpace(payload.ApprovalStatus); status != "" {
		if !isSenderApprovalStatus(status) {
			return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Unknown approval status."})
		}
		sets = append(sets, "approval_status = ?")
		arguments = append(arguments, status)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = NOW()")
		arguments = append(arguments, id)
		if _, err := server.db.Exec(
			"UPDATE senders SET "+strings.Join(sets, ", ")+" WHERE id = ?", arguments...,
		); err != nil {
			return databaseError(context, err)
		}
	}

	if payload.UserIDs != nil {
		if _, err := server.db.Exec("DELETE FROM sender_user WHERE sender_id = ?", id); err != nil {
			return databaseError(context, err)
		}
		if err := server.attachSenderUsers(id, *payload.UserIDs); err != nil {
			return databaseError(context, err)
		}
	}

	return context.JSON(fiber.Map{"message": "Sender updated successfully."})
}

func isSenderApprovalStatus(value string) bool {
	switch value {
	case "pending", "approved", "rejected":
		return true
	}

	return false
}

// attachSenderUsers mirrors Eloquent's syncWithoutDetaching: existing grants stay,
// missing ones are appended, and non-customer accounts are ignored.
func (server *Server) attachSenderUsers(senderID int64, userIDs []int64) error {
	customers := make([]int64, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID > 0 {
			customers = append(customers, userID)
		}
	}
	if len(customers) == 0 {
		return nil
	}

	group, bindings := placeholders(customers)
	rows, err := server.db.Query(
		"SELECT id FROM users WHERE type = 'customer' AND deleted_at IS NULL AND id IN "+group, bindings...)
	if err != nil {
		return err
	}
	defer rows.Close()

	valid := make([]int64, 0, len(customers))
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		valid = append(valid, userID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, userID := range valid {
		if _, err := server.db.Exec(
			"INSERT IGNORE INTO sender_user (sender_id, user_id) VALUES (?, ?)", senderID, userID); err != nil {
			return err
		}
	}

	return nil
}

func (server *Server) assignSenderUsers(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	payload := senderAssignmentPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || len(payload.UserIDs) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Select at least one customer."})
	}

	var senderID string
	if err := server.db.QueryRow(
		"SELECT sender_id FROM senders WHERE id = ? AND deleted_at IS NULL", id).Scan(&senderID); err != nil {
		return databaseError(context, err)
	}

	if err := server.attachSenderUsers(id, payload.UserIDs); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("Sender ID %s assigned successfully.", senderID)})
}

func (server *Server) updateSenderEnabled(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	payload := senderEnabledPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid payload."})
	}

	var approvalStatus string
	if err := server.db.QueryRow(
		"SELECT approval_status FROM senders WHERE id = ? AND deleted_at IS NULL", id).Scan(&approvalStatus); err != nil {
		return databaseError(context, err)
	}
	if approvalStatus != "approved" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Only approved senders can be enabled.",
		})
	}

	if _, err := server.db.Exec(
		"UPDATE senders SET is_enabled = ?, updated_at = NOW() WHERE id = ?", payload.IsEnabled, id); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Sender state updated.", "data": fiber.Map{"is_enabled": payload.IsEnabled}})
}

func (server *Server) deleteSender(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	if _, err := server.db.Exec(
		"UPDATE senders SET deleted_at = NOW(), updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", id); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Sender deleted successfully."})
}

func (server *Server) deleteSenders(context fiber.Ctx) error {
	payload := senderBulkPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || len(payload.IDs) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Select at least one sender."})
	}

	group, bindings := placeholders(payload.IDs)
	if _, err := server.db.Exec(
		"UPDATE senders SET deleted_at = NOW(), updated_at = NOW() WHERE deleted_at IS NULL AND id IN "+group, bindings...); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("%d sender(s) deleted.", len(payload.IDs))})
}

func (server *Server) approveSender(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	payload := senderApprovePayload{}
	_ = json.Unmarshal(context.Body(), &payload)

	var approvalStatus string
	var senderID string
	var storedGatewayID sql.NullInt64
	err = server.db.QueryRow(
		"SELECT sender_id, approval_status, gateway_id FROM senders WHERE id = ? AND deleted_at IS NULL",
		id).Scan(&senderID, &approvalStatus, &storedGatewayID)
	if err != nil {
		return databaseError(context, err)
	}
	if approvalStatus != "pending" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Only pending sender requests can be approved.",
		})
	}

	gatewayID := payload.GatewayID
	if gatewayID <= 0 && storedGatewayID.Valid {
		gatewayID = storedGatewayID.Int64
	}
	if gatewayID <= 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Assign a gateway before approving this sender request.",
		})
	}

	var gatewayExists int
	if err := server.db.QueryRow(
		"SELECT COUNT(*) FROM gateways WHERE id = ? AND deleted_at IS NULL", gatewayID).Scan(&gatewayExists); err != nil {
		return databaseError(context, err)
	}
	if gatewayExists == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The selected gateway no longer exists."})
	}

	admin := superman(context)
	if _, err := server.db.Exec(`
		UPDATE senders
		SET gateway_id = ?, approval_status = 'approved', is_enabled = 1, reviewed_by = ?, reviewed_at = NOW(),
			rejection_reason = NULL, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, gatewayID, admin.AdminID, id); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("Sender %s is approved and enabled.", senderID)})
}

func (server *Server) rejectSender(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Sender not found."})
	}

	payload := senderRejectPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil || strings.TrimSpace(payload.Reason) == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "A rejection reason is required."})
	}

	var approvalStatus string
	var senderID string
	if err := server.db.QueryRow(
		"SELECT sender_id, approval_status FROM senders WHERE id = ? AND deleted_at IS NULL", id,
	).Scan(&senderID, &approvalStatus); err != nil {
		return databaseError(context, err)
	}
	if approvalStatus != "pending" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Only pending sender requests can be rejected.",
		})
	}

	admin := superman(context)
	if _, err := server.db.Exec(`
		UPDATE senders
		SET approval_status = 'rejected', is_enabled = 0, reviewed_by = ?, reviewed_at = NOW(),
			rejection_reason = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, admin.AdminID, strings.TrimSpace(payload.Reason), id); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": fmt.Sprintf("Sender %s was rejected.", senderID)})
}

// senderCliOptions feeds the filter selects and the assignment modal.
func (server *Server) senderCliOptions(context fiber.Ctx) error {
	gatewayRows, err := server.db.Query(`
		SELECT id, name FROM gateways WHERE deleted_at IS NULL ORDER BY name ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer gatewayRows.Close()

	gateways := make([]fiber.Map, 0)
	for gatewayRows.Next() {
		var id int64
		var name string
		if err := gatewayRows.Scan(&id, &name); err != nil {
			return databaseError(context, err)
		}
		gateways = append(gateways, fiber.Map{"id": id, "name": name})
	}

	customerRows, err := server.db.Query(`
		SELECT id, name, email FROM users
		WHERE type = 'customer' AND deleted_at IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer customerRows.Close()

	customers := make([]senderUserRef, 0)
	for customerRows.Next() {
		var ref senderUserRef
		var email sql.NullString
		if err := customerRows.Scan(&ref.ID, &ref.Name, &email); err != nil {
			return databaseError(context, err)
		}
		if email.Valid {
			ref.Email = &email.String
		}
		customers = append(customers, ref)
	}

	typeClause, typeBindings := senderTypeClause(context.Query("type"))
	senderQuery := `SELECT id, sender_id, type FROM senders WHERE deleted_at IS NULL`
	arguments := []any{}
	if typeClause != "" {
		senderQuery += " AND " + typeClause
		arguments = append(arguments, typeBindings...)
	}
	senderQuery += " ORDER BY sender_id ASC"

	senderRows, err := server.db.Query(senderQuery, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer senderRows.Close()

	senders := make([]fiber.Map, 0)
	for senderRows.Next() {
		var id int64
		var senderID, senderType string
		if err := senderRows.Scan(&id, &senderID, &senderType); err != nil {
			return databaseError(context, err)
		}
		senders = append(senders, fiber.Map{
			"id": id, "sender_id": senderID, "type": senderType,
			"masking_category": senderMaskingCategory(senderType),
		})
	}

	return context.JSON(fiber.Map{"data": fiber.Map{
		"gateways":  gateways,
		"customers": customers,
		"senders":   senders,
	}})
}

type senderAuditLedgerItem struct {
	ID              int64           `json:"id"`
	SenderID        string          `json:"sender_id"`
	Type            string          `json:"type"`
	MaskingCategory string          `json:"masking_category"`
	GatewayName     *string         `json:"gateway_name"`
	IsEnabled       bool            `json:"is_enabled"`
	ApprovalStatus  string          `json:"approval_status"`
	TotalMessages   int             `json:"total_messages"`
	Delivered       int             `json:"delivered_messages"`
	Failed          int             `json:"failed_messages"`
	SuccessRate     float64         `json:"success_rate"`
	TotalSpend      float64         `json:"total_spend"`
	Users           []senderUserRef `json:"users"`
}

func (server *Server) senderAuditSummary(context fiber.Ctx) error {
	var totalSenders int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM senders WHERE deleted_at IS NULL").Scan(&totalSenders); err != nil {
		return databaseError(context, err)
	}

	var totalMessages, deliveredMessages int
	var totalRevenue sql.NullFloat64
	err := server.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(status = 'delivered'), 0), COALESCE(SUM(CASE WHEN status != 'failed' THEN cost ELSE 0 END), 0)
		FROM messages WHERE sender_id IS NOT NULL
	`).Scan(&totalMessages, &deliveredMessages, &totalRevenue)
	if err != nil {
		return databaseError(context, err)
	}

	successRate := 100.0
	if totalMessages > 0 {
		successRate = float64(deliveredMessages) * 100 / float64(totalMessages)
	}

	return context.JSON(fiber.Map{"data": fiber.Map{
		"total_senders":      totalSenders,
		"total_messages":     totalMessages,
		"delivered_messages": deliveredMessages,
		"total_revenue":      totalRevenue.Float64,
		"success_rate":       math.Round(successRate*10) / 10,
	}})
}

func (server *Server) senderAuditLedger(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	gatewayID, _ := strconv.ParseInt(context.Query("gateway_id"), 10, 64)

	arguments := []any{}
	where := "WHERE s.deleted_at IS NULL"
	if clause, bindings := senderTypeClause(context.Query("type")); clause != "" {
		where += " AND " + clause
		arguments = append(arguments, bindings...)
	}
	if gatewayID > 0 {
		where += " AND s.gateway_id = ?"
		arguments = append(arguments, gatewayID)
	}
	if search != "" {
		term := "%" + search + "%"
		where += ` AND (s.sender_id LIKE ?
			OR g.name LIKE ?
			OR EXISTS (SELECT 1 FROM sender_user su2 INNER JOIN users cu ON cu.id = su2.user_id
				WHERE su2.sender_id = s.id AND (cu.name LIKE ? OR cu.business_name LIKE ?)))`
		arguments = append(arguments, term, term, term, term)
	}

	var total int
	if err := server.db.QueryRow(`
		SELECT COUNT(*)
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		`+where, arguments...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	arguments = append(arguments, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT s.id, s.sender_id, s.type, s.is_enabled, s.approval_status, g.name,
			COUNT(m.id),
			COALESCE(SUM(m.status = 'delivered'), 0),
			COALESCE(SUM(m.status = 'failed'), 0),
			COALESCE(SUM(m.cost), 0)
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		LEFT JOIN messages m ON m.sender_id = s.id
		`+where+`
		GROUP BY s.id, s.sender_id, s.type, s.is_enabled, s.approval_status, g.name
		ORDER BY COUNT(m.id) DESC, s.id DESC
		LIMIT ? OFFSET ?
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]senderAuditLedgerItem, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var item senderAuditLedgerItem
		var gatewayName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.SenderID, &item.Type, &item.IsEnabled, &item.ApprovalStatus, &gatewayName,
			&item.TotalMessages, &item.Delivered, &item.Failed, &item.TotalSpend,
		); err != nil {
			return databaseError(context, err)
		}
		if gatewayName.Valid {
			item.GatewayName = &gatewayName.String
		}
		item.MaskingCategory = senderMaskingCategory(item.Type)
		if item.TotalMessages > 0 {
			item.SuccessRate = float64(item.Delivered) * 100 / float64(item.TotalMessages)
		} else {
			item.SuccessRate = 100
		}
		items = append(items, item)
		ids = append(ids, item.ID)
	}
	if err := rows.Err(); err != nil {
		return databaseError(context, err)
	}

	refs, err := server.senderUsers(ids)
	if err != nil {
		return databaseError(context, err)
	}
	for index := range items {
		items[index].Users = refs[items[index].ID]
		if items[index].Users == nil {
			items[index].Users = []senderUserRef{}
		}
	}

	totalPages := (total + perPage - 1) / perPage

	return context.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{"page": page, "per_page": perPage, "total": total, "total_pages": totalPages},
	})
}

func (server *Server) senderDispatches(context fiber.Ctx) error {
	search := strings.TrimSpace(context.Query("search"))

	arguments := []any{}
	where := "WHERE m.sender_id IS NOT NULL"
	if search != "" {
		term := "%" + search + "%"
		where += " AND (m.to LIKE ? OR s.sender_id LIKE ?)"
		arguments = append(arguments, term, term)
	}

	rows, err := server.db.Query(`
		SELECT m.id, m.created_at, m.to, m.cost, m.status,
			s.sender_id, g.name, u.name
		FROM messages m
		LEFT JOIN senders s ON s.id = m.sender_id
		LEFT JOIN gateways g ON g.id = m.gateway_id
		LEFT JOIN users u ON u.id = m.user_id
		`+where+`
		ORDER BY m.id DESC
		LIMIT 50
	`, arguments...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]fiber.Map, 0)
	for rows.Next() {
		var id int64
		var recipient, status string
		var cost sql.NullFloat64
		var senderID, gatewayName, userName sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &createdAt, &recipient, &cost, &status, &senderID, &gatewayName, &userName); err != nil {
			return databaseError(context, err)
		}
		item := fiber.Map{
			"id": id, "recipient": recipient, "status": status,
			"cost": cost.Float64,
		}
		if createdAt.Valid {
			item["created_at"] = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		if senderID.Valid {
			item["sender_id"] = senderID.String
		}
		if gatewayName.Valid {
			item["gateway_name"] = gatewayName.String
		}
		if userName.Valid {
			item["customer_name"] = userName.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"data": items})
}

type senderResolutionResult struct {
	SenderID           string  `json:"sender_id"`
	IsRegistered       bool    `json:"is_registered"`
	SenderType         string  `json:"sender_type"`
	MaskingCategory    string  `json:"masking_category"`
	IsMasking          bool    `json:"is_masking"`
	ApprovalStatus     string  `json:"approval_status"`
	GatewayName        string  `json:"gateway_name"`
	GatewayDriver      string  `json:"gateway_driver"`
	RecipientPhone     string  `json:"recipient_phone"`
	OperatorName       string  `json:"operator_name"`
	OperatorBaseCost   float64 `json:"operator_base_cost"`
	GatewayChargedRate float64 `json:"gateway_charged_rate"`
	RouteHealth        string  `json:"route_health"`
	Status             string  `json:"status"`
}

type operatorRecord struct {
	ID              int64
	Name            string
	Codes           []string
	MaskingPrice    sql.NullFloat64
	NonMaskingPrice sql.NullFloat64
}

func (server *Server) senderResolution(context fiber.Ctx) error {
	senderInput := strings.TrimSpace(context.Query("sender"))
	phoneInput := strings.TrimSpace(context.Query("phone"))
	if senderInput == "" || phoneInput == "" {
		return context.JSON(fiber.Map{"data": nil})
	}

	result := senderResolutionResult{
		SenderID:       senderInput,
		RecipientPhone: phoneInput,
	}
	result.IsMasking = true
	result.SenderType = "Masking (Brand Alphanumeric)"
	result.OperatorBaseCost = 0.25
	result.GatewayChargedRate = 0.45
	result.RouteHealth = "Optimal (Latency ~80ms)"
	result.Status = "Route Active"

	var gatewayID sql.NullInt64
	var gatewayName, gatewayDriver sql.NullString
	var senderType, approvalStatus string
	err := server.db.QueryRow(`
		SELECT s.sender_id, s.type, s.approval_status, g.id, g.name, g.driver
		FROM senders s
		LEFT JOIN gateways g ON g.id = s.gateway_id
		WHERE s.sender_id = ? AND s.deleted_at IS NULL
		ORDER BY s.id DESC
		LIMIT 1
	`, senderInput).Scan(&result.SenderID, &senderType, &approvalStatus, &gatewayID, &gatewayName, &gatewayDriver)
	if err != nil && err != sql.ErrNoRows {
		return databaseError(context, err)
	}

	if err == nil {
		result.IsRegistered = true
		result.ApprovalStatus = approvalStatus
		result.IsMasking = senderType != "numeric"
		result.MaskingCategory = senderMaskingCategory(senderType)
		switch senderType {
		case "numeric":
			result.SenderType = "Non-Masking (Numeric CLI)"
		case "shortcode":
			result.SenderType = "Masking (Shortcode)"
		default:
			result.SenderType = "Masking (Brand Alphanumeric)"
		}
	} else {
		switch {
		case isAllDigits(senderInput) && len(senderInput) >= 10:
			result.IsMasking = false
			result.SenderType = "Non-Masking (Numeric CLI)"
			result.MaskingCategory = "non-masking"
		case isAllDigits(senderInput):
			result.SenderType = "Masking (Shortcode)"
			result.MaskingCategory = "masking"
		}
		result.ApprovalStatus = "unregistered"
	}

	if !gatewayID.Valid {
		var fallbackID sql.NullInt64
		var fallbackName, fallbackDriver sql.NullString
		_ = server.db.QueryRow(`
			SELECT id, name, driver FROM gateways WHERE deleted_at IS NULL AND is_enabled = 1 ORDER BY id ASC LIMIT 1
		`).Scan(&fallbackID, &fallbackName, &fallbackDriver)
		gatewayID, gatewayName, gatewayDriver = fallbackID, fallbackName, fallbackDriver
	}
	if gatewayName.Valid {
		result.GatewayName = gatewayName.String
	} else {
		result.GatewayName = "Default Gateway"
	}
	if gatewayDriver.Valid {
		result.GatewayDriver = gatewayDriver.String
	} else {
		result.GatewayDriver = "smpp"
	}

	nationalNumber := strings.ReplaceAll(phoneInput, " ", "")
	nationalNumber = strings.TrimPrefix(nationalNumber, "+")
	digits := onlyDigits(nationalNumber)
	if strings.HasPrefix(digits, "880") {
		digits = digits[3:]
	}

	operators, err := server.activeOperators()
	if err != nil {
		return databaseError(context, err)
	}

	matchedOperator := (*operatorRecord)(nil)
	for index := range operators {
		for _, code := range operators[index].Codes {
			normalized := onlyDigits(code)
			if normalized == "" {
				continue
			}
			if strings.HasPrefix(digits, normalized) || strings.HasPrefix(digits, "0"+normalized) {
				matchedOperator = &operators[index]
				break
			}
		}
		if matchedOperator != nil {
			break
		}
	}
	if matchedOperator == nil {
		matchedOperator = bangladeshOperatorByPrefix(operators, digits)
	}

	result.OperatorName = "Unknown Operator"
	if matchedOperator != nil {
		result.OperatorName = matchedOperator.Name
		if result.IsMasking && matchedOperator.MaskingPrice.Valid {
			result.OperatorBaseCost = matchedOperator.MaskingPrice.Float64
		} else if !result.IsMasking && matchedOperator.NonMaskingPrice.Valid {
			result.OperatorBaseCost = matchedOperator.NonMaskingPrice.Float64
		}

		if gatewayID.Valid {
			var rate sql.NullFloat64
			rateColumn := "masking_price"
			if !result.IsMasking {
				rateColumn = "non_masking_price"
			}
			err := server.db.QueryRow(fmt.Sprintf(`
				SELECT %s FROM gateway_operator_rates WHERE gateway_id = ? AND operator_id = ? LIMIT 1
			`, rateColumn), gatewayID.Int64, matchedOperator.ID).Scan(&rate)
			if err == nil && rate.Valid {
				result.GatewayChargedRate = rate.Float64
			} else if result.IsMasking {
				result.GatewayChargedRate = 0.45
			} else {
				result.GatewayChargedRate = 0.23
			}
		}
	}

	return context.JSON(fiber.Map{"data": result})
}

func (server *Server) activeOperators() ([]operatorRecord, error) {
	rows, err := server.db.Query(`
		SELECT id, name, codes, masking_price, non_masking_price
		FROM operators WHERE is_active = 1 ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operators := make([]operatorRecord, 0)
	for rows.Next() {
		var record operatorRecord
		var rawCodes sql.NullString
		if err := rows.Scan(&record.ID, &record.Name, &rawCodes, &record.MaskingPrice, &record.NonMaskingPrice); err != nil {
			return nil, err
		}
		if rawCodes.Valid {
			codes := []json.RawMessage{}
			if err := json.Unmarshal([]byte(rawCodes.String), &codes); err == nil {
				for _, code := range codes {
					text := strings.Trim(string(code), `"`)
					record.Codes = append(record.Codes, text)
				}
			}
		}
		operators = append(operators, record)
	}

	return operators, rows.Err()
}

func bangladeshOperatorByPrefix(operators []operatorRecord, digits string) *operatorRecord {
	preferred := "Unknown Operator"
	switch {
	case hasAnyPrefix(digits, "017", "013", "17", "13"):
		preferred = "Grameenphone"
	case hasAnyPrefix(digits, "018", "18"):
		preferred = "Robi"
	case hasAnyPrefix(digits, "019", "014", "19", "14"):
		preferred = "Banglalink"
	case hasAnyPrefix(digits, "015", "15"):
		preferred = "Teletalk"
	case hasAnyPrefix(digits, "016", "16"):
		preferred = "Airtel"
	default:
		return nil
	}

	for index := range operators {
		if strings.Contains(strings.ToLower(operators[index].Name), strings.ToLower(preferred)) {
			return &operators[index]
		}
	}

	return nil
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

func isAllDigits(value string) bool {
	if value == "" {
		return false
	}

	return onlyDigits(value) == value
}

func onlyDigits(value string) string {
	builder := strings.Builder{}
	for _, character := range value {
		if character >= '0' && character <= '9' {
			builder.WriteRune(character)
		}
	}

	return builder.String()
}
