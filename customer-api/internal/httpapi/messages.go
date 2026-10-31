package httpapi

import (
	stdctx "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"smswhiteinfo/customer-api/internal/billing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) dashboardStats(context fiber.Ctx) error {
	claims := customer(context)

	var total, successful, failed int
	var cost float64
	var contacts int
	today := time.Now().Format("2006-01-02")

	err := server.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status IN ('sent','delivered') THEN 1 ELSE 0 END),0) as successful,
			COALESCE(SUM(CASE WHEN status IN ('failed','dlr_failed') THEN 1 ELSE 0 END),0) as failed,
			COALESCE(SUM(CASE WHEN status NOT IN ('failed','dlr_failed') THEN cost ELSE 0 END),0) as cost
		FROM messages
		WHERE user_id = ? AND DATE(created_at) = ?
	`, claims.UserID, today).Scan(&total, &successful, &failed, &cost)
	if err != nil && err != sql.ErrNoRows {
		return databaseError(context, err)
	}

	_ = server.db.QueryRow(
		"SELECT COUNT(*) FROM contacts WHERE user_id = ? AND deleted_at IS NULL", claims.UserID,
	).Scan(&contacts)

	deliveryRate := 0.0
	if total > 0 {
		deliveryRate = float64(successful) / float64(total) * 100
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"messages_today":   total,
			"successful_today": successful,
			"failed_today":     failed,
			"cost_today":       roundFloat(cost, 2),
			"delivery_rate":    roundFloat(deliveryRate, 1),
			"contacts_total":   contacts,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Messages
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) messages(context fiber.Ctx) error {
	claims := customer(context)
	page, perPage := pagination(context)
	offset := (page - 1) * perPage

	conditions := []string{"m.user_id = ?"}
	args := []any{claims.UserID}

	if search := context.Query("search"); search != "" {
		conditions = append(conditions, "(m.to LIKE ? OR m.body LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	if status := context.Query("status"); status != "" {
		conditions = append(conditions, "m.status = ?")
		args = append(args, status)
	}
	if dateFrom := context.Query("date_from"); dateFrom != "" {
		conditions = append(conditions, "DATE(m.created_at) >= ?")
		args = append(args, dateFrom)
	}
	if dateTo := context.Query("date_to"); dateTo != "" {
		conditions = append(conditions, "DATE(m.created_at) <= ?")
		args = append(args, dateTo)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := server.db.QueryRow("SELECT COUNT(*) FROM messages m "+where, countArgs...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	sort := "m.id"
	dir := "DESC"
	if s := context.Query("sort"); s != "" {
		allowed := map[string]bool{"id": true, "to": true, "status": true, "cost": true, "created_at": true, "sent_at": true}
		if allowed[s] {
			sort = "m." + s
		}
	}
	if context.Query("dir") == "asc" {
		dir = "ASC"
	}

	args = append(args, perPage, offset)
	rows, err := server.db.Query(fmt.Sprintf(`
		SELECT m.id, COALESCE(s.sender_id, '—'), m.to, m.body, m.status, m.cost, m.parts,
		       m.created_at, m.sent_at, m.delivered_at, m.failed_reason, COALESCE(m.operator_name, '')
		FROM messages m
		LEFT JOIN senders s ON s.id = m.sender_id
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, where, sort, dir), args...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := []fiber.Map{}
	for rows.Next() {
		var id int
		var senderID, to, body, status, operatorName string
		var cost float64
		var parts int
		var createdAt time.Time
		var sentAt, deliveredAt *time.Time
		var failedReason *string

		if err := rows.Scan(&id, &senderID, &to, &body, &status, &cost, &parts,
			&createdAt, &sentAt, &deliveredAt, &failedReason, &operatorName); err != nil {
			continue
		}

		op := operatorName
		if op == "" {
			op = operatorFromNumber(to)
		}

		items = append(items, fiber.Map{
			"id":            id,
			"msg_id":        fmt.Sprintf("MSG-%06d", id),
			"sender_id":     senderID,
			"to":            to,
			"body":          body,
			"status":        status,
			"cost":          fmt.Sprintf("%.4f", cost),
			"parts":         parts,
			"created_at":    createdAt.Format(time.RFC3339),
			"submitted_at":  createdAt.Format(time.RFC3339),
			"sent_at":       apiTimePtr(sentAt),
			"delivered_at":  apiTimePtr(deliveredAt),
			"error_message": failedReason,
			"operator":      op,
		})
	}

	return context.JSON(fiber.Map{
		"data": items,
		"meta": pageMeta(total, perPage, page),
	})
}

func (server *Server) sendMessage(context fiber.Ctx) error {
	claims := customer(context)

	var req struct {
		SenderID int64  `json:"sender_id"`
		Phones   any    `json:"phones"`
		Body     string `json:"body"`
	}
	if err := json.Unmarshal(context.Body(), &req); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request body."})
	}

	if req.SenderID == 0 || req.Body == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "sender_id and body are required."})
	}

	// Verify sender belongs to user or is approved
	var senderLabel, senderType string
	err := server.db.QueryRow(`
		SELECT s.sender_id, s.type FROM senders s
		LEFT JOIN sender_user su ON su.sender_id = s.id
		WHERE (su.user_id = ? OR s.requested_by = ?) AND s.id = ? AND s.deleted_at IS NULL
		LIMIT 1
	`, claims.UserID, claims.UserID, req.SenderID).Scan(&senderLabel, &senderType)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Selected Sender ID is not available."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	phones := normalizePhones(req.Phones)
	if len(phones) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "At least one valid phone number is required."})
	}

	body := strings.TrimSpace(req.Body)

	// Check rate resolver & balance estimate
	resolver, err := billing.ForUser(context.Context(), server.db, claims.UserID)
	if err != nil {
		return databaseError(context, err)
	}

	isMasking := strings.Contains(strings.ToLower(senderType), "masking")
	parts, _ := billing.CalculateParts(body)
	rates := resolver.RatesForUser(isMasking)
	maxRate := 0.0
	for _, r := range rates {
		if r > maxRate {
			maxRate = r
		}
	}
	estimatedCost := billing.RoundFloat(maxRate*float64(parts)*float64(len(phones)), 4)

	var balance float64
	_ = server.db.QueryRow("SELECT balance FROM users WHERE id = ?", claims.UserID).Scan(&balance)
	if balance < estimatedCost {
		return context.Status(http.StatusPaymentRequired).JSON(fiber.Map{
			"message": "Insufficient balance for this message.",
			"balance": balance,
		})
	}

	recipients := make([]any, len(phones))
	for i, p := range phones {
		recipients[i] = map[string]any{"phone": p, "values": map[string]string{}}
	}
	srcConfigBytes, _ := json.Marshal(map[string]any{"recipients": recipients})

	// Create batch with status='processing'
	result, err := server.db.Exec(`
		INSERT INTO batches (user_id, sender_id, body, source_type, source_ip, source_config, status, approval, total_recipients, valid_count, created_at, updated_at)
		VALUES (?, ?, ?, 'manual', ?, ?, 'processing', 'approved', ?, 0, NOW(), NOW())
	`, claims.UserID, req.SenderID, body, context.IP(), string(srcConfigBytes), len(phones))
	if err != nil {
		return databaseError(context, err)
	}

	batchID, _ := result.LastInsertId()

	// Dispatch batch processor asynchronously
	go func(bid int64) {
		_ = server.batchProcessor.ProcessBatch(stdctx.Background(), bid)
	}(batchID)

	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":         batchID,
			"status":     "processing",
			"recipients": len(phones),
		},
	})
}

func (server *Server) scheduleMessage(context fiber.Ctx) error {
	claims := customer(context)

	var req struct {
		SenderID        int64  `json:"sender_id"`
		Phones          any    `json:"phones"`
		Body            string `json:"body"`
		ScheduledDate   string `json:"scheduled_date"`
		ScheduledHour   string `json:"scheduled_hour"`
		ScheduledMinute string `json:"scheduled_minute"`
		ScheduledPeriod string `json:"scheduled_period"`
	}
	if err := json.Unmarshal(context.Body(), &req); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request body."})
	}

	phones := normalizePhones(req.Phones)
	if len(phones) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "At least one valid phone number is required."})
	}

	scheduledAt, err := parseScheduledTime(req.ScheduledDate, req.ScheduledHour, req.ScheduledMinute, req.ScheduledPeriod)
	if err != nil || scheduledAt.Before(time.Now().Add(-time.Minute)) {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The selected schedule time has already passed."})
	}

	body := strings.TrimSpace(req.Body)
	now := time.Now()

	result, err := server.db.Exec(`
		INSERT INTO batches (user_id, sender_id, body, source_type, source_ip, source_config, status, total_recipients, valid_count, scheduled_at, created_at, updated_at)
		VALUES (?, ?, ?, 'manual', ?, ?, 'scheduled', ?, ?, ?, ?, ?)
	`, claims.UserID, req.SenderID, body, context.IP(), buildSourceConfig(phones), len(phones), len(phones), scheduledAt, now, now)
	if err != nil {
		return databaseError(context, err)
	}

	batchID, _ := result.LastInsertId()

	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":           batchID,
			"status":       "scheduled",
			"scheduled_at": scheduledAt.Format(time.RFC3339),
		},
	})
}

func (server *Server) bulkSend(context fiber.Ctx) error {
	claims := customer(context)

	var req struct {
		SenderID        int64   `json:"sender_id"`
		Body            string  `json:"body"`
		Source          string  `json:"source"`
		Phones          any     `json:"phones"`
		ScheduledDate   string  `json:"scheduled_date"`
		ScheduledHour   string  `json:"scheduled_hour"`
		ScheduledMinute string  `json:"scheduled_minute"`
		ScheduledPeriod string  `json:"scheduled_period"`
		Groups          []int64 `json:"groups"`
	}
	if err := json.Unmarshal(context.Body(), &req); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request body."})
	}

	var phones []string

	switch req.Source {
	case "manual":
		phones = normalizePhones(req.Phones)
	case "group":
		if len(req.Groups) > 0 {
			placeholders := strings.Repeat("?,", len(req.Groups))
			placeholders = placeholders[:len(placeholders)-1]
			args := []any{claims.UserID}
			for _, gid := range req.Groups {
				args = append(args, gid)
			}
			rows, err := server.db.Query(fmt.Sprintf(`
				SELECT c.phone FROM contacts c
				JOIN contact_group cg ON cg.contact_id = c.id
				WHERE c.user_id = ? AND cg.group_id IN (%s) AND c.deleted_at IS NULL
			`, placeholders), args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var p string
					if rows.Scan(&p) == nil {
						phones = append(phones, p)
					}
				}
			}
		}
	default:
		phones = normalizePhones(req.Phones)
	}

	if len(phones) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "No valid recipients found for this bulk campaign."})
	}

	var scheduledAt *time.Time
	if req.ScheduledDate != "" && req.ScheduledHour != "" {
		t, err := parseScheduledTime(req.ScheduledDate, req.ScheduledHour, req.ScheduledMinute, req.ScheduledPeriod)
		if err == nil {
			scheduledAt = &t
		}
	}

	body := strings.TrimSpace(req.Body)
	now := time.Now()

	if scheduledAt != nil {
		result, err := server.db.Exec(`
			INSERT INTO batches (user_id, sender_id, body, source_type, source_ip, source_config, status, total_recipients, valid_count, scheduled_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'scheduled', ?, 0, ?, ?, ?)
		`, claims.UserID, req.SenderID, body, req.Source, context.IP(), buildSourceConfig(phones), len(phones), scheduledAt, now, now)
		if err != nil {
			return databaseError(context, err)
		}

		batchID, _ := result.LastInsertId()
		return context.Status(http.StatusCreated).JSON(fiber.Map{
			"data": fiber.Map{
				"id":         batchID,
				"status":     "scheduled",
				"recipients": len(phones),
			},
		})
	}

	// Verify affordability
	var senderType string
	_ = server.db.QueryRow("SELECT type FROM senders WHERE id = ?", req.SenderID).Scan(&senderType)

	resolver, err := billing.ForUser(context.Context(), server.db, claims.UserID)
	if err != nil {
		return databaseError(context, err)
	}

	isMasking := strings.Contains(strings.ToLower(senderType), "masking")
	parts, _ := billing.CalculateParts(body)
	rates := resolver.RatesForUser(isMasking)
	maxRate := 0.0
	for _, r := range rates {
		if r > maxRate {
			maxRate = r
		}
	}
	estimatedCost := billing.RoundFloat(maxRate*float64(parts)*float64(len(phones)), 4)

	var balance float64
	_ = server.db.QueryRow("SELECT balance FROM users WHERE id = ?", claims.UserID).Scan(&balance)
	if balance < estimatedCost {
		return context.Status(http.StatusPaymentRequired).JSON(fiber.Map{
			"message": "Insufficient balance for this batch.",
			"balance": balance,
		})
	}

	result, err := server.db.Exec(`
		INSERT INTO batches (user_id, sender_id, body, source_type, source_ip, source_config, status, total_recipients, valid_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'processing', ?, 0, ?, ?)
	`, claims.UserID, req.SenderID, body, req.Source, context.IP(), buildSourceConfig(phones), len(phones), now, now)
	if err != nil {
		return databaseError(context, err)
	}

	batchID, _ := result.LastInsertId()

	go func(bid int64) {
		_ = server.batchProcessor.ProcessBatch(stdctx.Background(), bid)
	}(batchID)

	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":         batchID,
			"status":     "processing",
			"recipients": len(phones),
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Batches
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) batches(context fiber.Ctx) error {
	claims := customer(context)
	page, perPage := pagination(context)
	offset := (page - 1) * perPage

	conditions := []string{"b.user_id = ?"}
	args := []any{claims.UserID}

	if search := context.Query("search"); search != "" {
		conditions = append(conditions, "(b.id LIKE ? OR b.body LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	_ = server.db.QueryRow("SELECT COUNT(*) FROM batches b "+where, countArgs...).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := server.db.Query(`
		SELECT b.id, COALESCE(s.sender_id, '—'), b.status, b.approval, b.source_type,
		       b.total_recipients,
		       (SELECT COUNT(*) FROM messages m WHERE m.batch_id = b.id AND m.status IN ('sent', 'delivered')) as sent_count,
		       (SELECT COUNT(*) FROM messages m WHERE m.batch_id = b.id AND m.status IN ('failed', 'dlr_failed')) as failed_count,
		       (SELECT COALESCE(SUM(m.cost), 0) FROM messages m WHERE m.batch_id = b.id) as cost,
		       b.scheduled_at, b.created_at
		FROM batches b
		LEFT JOIN senders s ON s.id = b.sender_id
		`+where+`
		ORDER BY b.id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := []fiber.Map{}
	for rows.Next() {
		var id, totalRec, sentCnt, failedCnt int
		var senderID, status, sourceType string
		var approval *string
		var cost float64
		var scheduledAt *time.Time
		var createdAt time.Time

		if err := rows.Scan(&id, &senderID, &status, &approval, &sourceType,
			&totalRec, &sentCnt, &failedCnt, &cost, &scheduledAt, &createdAt); err != nil {
			continue
		}
		items = append(items, fiber.Map{
			"id":               id,
			"sender_id":        senderID,
			"status":           status,
			"approval":         approval,
			"source_type":      sourceType,
			"total_recipients": totalRec,
			"sent_count":       sentCnt,
			"failed_count":     failedCnt,
			"cost":             fmt.Sprintf("%.2f", cost),
			"scheduled_at":     apiTimePtr(scheduledAt),
			"created_at":       createdAt.Format(time.RFC3339),
		})
	}

	return context.JSON(fiber.Map{
		"data": items,
		"meta": pageMeta(total, perPage, page),
	})
}

func (server *Server) batchDetail(context fiber.Ctx) error {
	claims := customer(context)
	id := context.Params("id")

	var batchID, totalRec, sentCnt, deliveredCnt, failedCnt int
	var senderID, status, sourceType, body string
	var approval *string
	var cost float64
	var scheduledAt *time.Time
	var createdAt time.Time

	err := server.db.QueryRow(`
		SELECT b.id, COALESCE(s.sender_id, '—'), b.status, b.approval, b.source_type, b.body,
		       b.total_recipients,
		       (SELECT COUNT(*) FROM messages m WHERE m.batch_id = b.id AND m.status IN ('sent', 'delivered')) as sent_count,
		       (SELECT COUNT(*) FROM messages m WHERE m.batch_id = b.id AND m.status = 'delivered') as delivered_count,
		       (SELECT COUNT(*) FROM messages m WHERE m.batch_id = b.id AND m.status IN ('failed', 'dlr_failed')) as failed_count,
		       (SELECT COALESCE(SUM(m.cost), 0) FROM messages m WHERE m.batch_id = b.id) as cost,
		       b.scheduled_at, b.created_at
		FROM batches b
		LEFT JOIN senders s ON s.id = b.sender_id
		WHERE b.id = ? AND b.user_id = ?
	`, id, claims.UserID).Scan(
		&batchID, &senderID, &status, &approval, &sourceType, &body,
		&totalRec, &sentCnt, &deliveredCnt, &failedCnt, &cost,
		&scheduledAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Batch not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	// Get messages for this batch
	msgRows, err := server.db.Query(`
		SELECT id, `+"`to`"+`, body, status, cost, failed_reason, sent_at, delivered_at
		FROM messages
		WHERE batch_id = ?
		ORDER BY id DESC
		LIMIT 100
	`, batchID)
	if err != nil {
		return databaseError(context, err)
	}
	defer msgRows.Close()

	messages := []fiber.Map{}
	for msgRows.Next() {
		var mid int
		var to, msgBody, msgStatus string
		var msgCost float64
		var failedReason *string
		var sentAt, deliveredAt *time.Time

		if err := msgRows.Scan(&mid, &to, &msgBody, &msgStatus, &msgCost, &failedReason, &sentAt, &deliveredAt); err != nil {
			continue
		}
		messages = append(messages, fiber.Map{
			"id":            mid,
			"to":            to,
			"body":          msgBody,
			"status":        msgStatus,
			"cost":          msgCost,
			"error_message": failedReason,
			"sent_at":       apiTimePtr(sentAt),
			"delivered_at":  apiTimePtr(deliveredAt),
		})
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"batch": fiber.Map{
				"id":               batchID,
				"sender_id":        senderID,
				"status":           status,
				"approval":         approval,
				"source_type":      sourceType,
				"body":             body,
				"total_recipients": totalRec,
				"sent_count":       sentCnt,
				"delivered_count":  deliveredCnt,
				"failed_count":     failedCnt,
				"cost":             fmt.Sprintf("%.2f", cost),
				"scheduled_at":     apiTimePtr(scheduledAt),
				"created_at":       createdAt.Format(time.RFC3339),
			},
			"messages": messages,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func operatorFromNumber(phone string) string {
	cleaned := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(phone), "+"), "88")
	if len(cleaned) >= 3 {
		switch cleaned[:3] {
		case "017", "013":
			return "Grameenphone"
		case "018":
			return "Robi"
		case "019", "014":
			return "Banglalink"
		case "015":
			return "Teletalk"
		case "016":
			return "Airtel"
		}
	}
	return "Other"
}

func normalizePhones(raw any) []string {
	var items []string
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					items = append(items, s)
				}
			}
		}
	case string:
		for _, p := range strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\r' || r == ' '
		}) {
			p = strings.TrimSpace(p)
			if p != "" {
				items = append(items, p)
			}
		}
	}

	return items
}

func buildSourceConfig(phones []string) string {
	parts := make([]string, len(phones))
	for i, p := range phones {
		parts[i] = fmt.Sprintf(`{"phone":%q,"values":{}}`, p)
	}

	return fmt.Sprintf(`{"recipients":[%s]}`, strings.Join(parts, ","))
}

func parseScheduledTime(date, hour, minute, period string) (time.Time, error) {
	if hour == "" {
		hour = "12"
	}
	if minute == "" {
		minute = "00"
	}
	if period == "" {
		period = "AM"
	}
	period = strings.ToUpper(period)

	raw := fmt.Sprintf("%s %02s:%02s %s", date, hour, minute, period)
	loc, _ := time.LoadLocation("Asia/Dhaka")
	if loc == nil {
		loc = time.FixedZone("Asia/Dhaka", 6*60*60)
	}

	return time.ParseInLocation("2006-01-02 03:04 PM", raw, loc)
}

func roundFloat(val float64, prec int) float64 {
	shift := 1.0
	for i := 0; i < prec; i++ {
		shift *= 10
	}

	return float64(int64(val*shift+0.5)) / shift
}
