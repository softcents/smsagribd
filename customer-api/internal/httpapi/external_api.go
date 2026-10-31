package httpapi

import (
	stdctx "context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"smswhiteinfo/customer-api/internal/billing"
)

// ─────────────────────────────────────────────────────────────────────────────
// External Developer API — token-authenticated (/api/v1/*)
// Replaces Laravel routes/api.php external developer endpoints.
// Authentication: Bearer <id|token> (Sanctum-compatible personal access token)
// ─────────────────────────────────────────────────────────────────────────────

// devUserData carries the resolved customer identity for developer API requests.
type devUserData struct {
	ID              int64
	Name            string
	Email           string
	Phone           *string
	Balance         float64
	Status          string
	SMSAutoApproved bool
	TokenID         int64
}

// RegisterExternalAPI mounts the external developer API routes on the fiber app.
// Called from server.App() after the customer portal routes.
// RegisterExternalAPI mounts the external developer API routes on the fiber app.
func (server *Server) RegisterExternalAPI(app *fiber.App) {
	ext := app.Group("/api/v1", server.devAuthenticate)

	// Account
	ext.Get("/me", server.extMe)
	ext.Get("/balance", server.extBalance)
	ext.Get("/senders", server.extSenders)

	// Messaging
	ext.Get("/messages", server.extSendMessage)
	ext.Post("/messages", server.extSendMessage)
	ext.Get("/batches/:batch", server.extShowBatch)

	// IP Whitelisting
	ext.Get("/ip-whitelists", server.extIPWhitelists)
	ext.Post("/ip-whitelists", server.extCreateIPWhitelist)
	ext.Delete("/ip-whitelists/:id", server.extDeleteIPWhitelist)
}

// devAuthenticate validates a Sanctum-style personal access token.
// Token format: "<id>|<plain-token>"
func (server *Server) devAuthenticate(context fiber.Ctx) error {
	authHeader := context.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthenticated.",
		})
	}
	rawToken := strings.TrimPrefix(authHeader, "Bearer ")

	tokenParts := strings.SplitN(rawToken, "|", 2)
	if len(tokenParts) != 2 {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthenticated.",
		})
	}
	tokenID := tokenParts[0]
	plainToken := tokenParts[1]

	// Hash the plain token to compare with DB
	digest := sha256.Sum256([]byte(plainToken))
	hashedToken := hex.EncodeToString(digest[:])

	var user devUserData
	err := server.db.QueryRow(`
		SELECT u.id, u.name, u.email, u.phone, u.balance, u.status,
		       COALESCE(u.sms_auto_approved, 1), t.id
		FROM access_tokens t
		JOIN users u ON u.id = t.tokenable_id
		WHERE t.id = ?
		  AND t.token = ?
		  AND t.tokenable_type = 'App\\Models\\User'
		  AND u.type = 'customer'
		  AND u.deleted_at IS NULL
		LIMIT 1
	`, tokenID, hashedToken).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.Balance, &user.Status, &user.SMSAutoApproved, &user.TokenID,
	)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Server error."})
	}

	// Check IP whitelist if customer account requires it
	clientIP := context.IP()
	if !server.isIPAllowed(user.ID, clientIP) {
		return context.Status(http.StatusForbidden).JSON(fiber.Map{
			"message": "Access from IP " + clientIP + " is not permitted. Add this IP to your whitelist.",
		})
	}

	// Update last_used_at on the token
	_, _ = server.db.Exec(
		"UPDATE access_tokens SET last_used_at = ? WHERE id = ?",
		time.Now(), user.TokenID,
	)

	context.Locals("dev_user", user)
	return context.Next()
}

// getDevUser retrieves the authenticated devUserData from the context.
func getDevUser(context fiber.Ctx) devUserData {
	return context.Locals("dev_user").(devUserData)
}

// isIPAllowed returns true if the customer has no approved whitelisted IPs
// (open access) or if the request IP matches an approved whitelist entry.
func (server *Server) isIPAllowed(userID int64, clientIP string) bool {
	var count int
	_ = server.db.QueryRow(
		"SELECT COUNT(*) FROM ip_whitelist_requests WHERE user_id = ? AND status = 'approved'",
		userID,
	).Scan(&count)

	if count == 0 {
		// No whitelist configured — open access
		return true
	}

	rows, err := server.db.Query(
		"SELECT ip_address FROM ip_whitelist_requests WHERE user_id = ? AND status = 'approved'",
		userID,
	)
	if err != nil {
		return false
	}
	defer rows.Close()

	requestIP := net.ParseIP(clientIP)
	for rows.Next() {
		var entry string
		if rows.Scan(&entry) != nil {
			continue
		}
		if strings.Contains(entry, "/") {
			_, ipNet, err := net.ParseCIDR(entry)
			if err == nil && requestIP != nil && ipNet.Contains(requestIP) {
				return true
			}
		} else if entry == clientIP {
			return true
		}
	}

	return false
}

// (devUser accessor moved above isIPAllowed)

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/me
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extMe(context fiber.Ctx) error {
	u := getDevUser(context)
	return context.JSON(fiber.Map{
		"id":       u.ID,
		"name":     u.Name,
		"email":    u.Email,
		"phone":    u.Phone,
		"balance":  u.Balance,
		"approval": u.Status,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/balance
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extBalance(context fiber.Ctx) error {
	u := getDevUser(context)
	// Re-fetch fresh balance
	var balance float64
	_ = server.db.QueryRow("SELECT balance FROM users WHERE id = ?", u.ID).Scan(&balance)
	return context.JSON(fiber.Map{
		"balance":  balance,
		"currency": "BDT",
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/senders
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extSenders(context fiber.Ctx) error {
	u := getDevUser(context)

	rows, err := server.db.Query(`
		SELECT DISTINCT s.sender_id, s.type
		FROM senders s
		LEFT JOIN sender_user su ON su.sender_id = s.id
		WHERE (su.user_id = ? OR s.requested_by = ?)
		  AND s.deleted_at IS NULL
		  AND s.is_enabled = 1
		ORDER BY s.sender_id ASC
	`, u.ID, u.ID)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := []fiber.Map{}
	for rows.Next() {
		var senderID string
		var senderType *string
		if rows.Scan(&senderID, &senderType) != nil {
			continue
		}
		category := "unknown"
		if senderType != nil {
			if strings.Contains(strings.ToLower(*senderType), "masking") {
				category = "masking"
			} else {
				category = "non_masking"
			}
		}
		items = append(items, fiber.Map{
			"sender_id": senderID,
			"type":      category,
		})
	}

	return context.JSON(fiber.Map{"data": items})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/messages  POST /api/v1/messages
// Sends an SMS batch. Accepts GET and POST (URL-params or JSON body).
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extSendMessage(context fiber.Ctx) error {
	u := getDevUser(context)

	// Accept both GET query params and POST body
	senderLabel := context.Query("sender_id")
	numbersRaw := context.Query("numbers")
	body := context.Query("body")

	if context.Method() == "POST" {
		var req struct {
			SenderID string `json:"sender_id"`
			Numbers  any    `json:"numbers"`
			Body     string `json:"body"`
		}
		if err := json.Unmarshal(context.Body(), &req); err == nil {
			if req.SenderID != "" {
				senderLabel = req.SenderID
			}
			if req.Numbers != nil {
				switch v := req.Numbers.(type) {
				case string:
					if numbersRaw == "" {
						numbersRaw = v
					}
				case []any:
					strs := make([]string, 0, len(v))
					for _, n := range v {
						if s, ok := n.(string); ok {
							strs = append(strs, s)
						}
					}
					numbersRaw = strings.Join(strs, ",")
				}
			}
			if req.Body != "" {
				body = req.Body
			}
		}
	}

	// Validate inputs
	if senderLabel == "" || body == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "sender_id, numbers, and body are required.",
		})
	}

	numbers := normalizePhones(numbersRaw)
	if len(numbers) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "At least one phone number is required.",
		})
	}

	// Verify sender belongs to user
	var senderDBID int64
	var senderType string
	err := server.db.QueryRow(`
		SELECT s.id, s.type FROM senders s
		LEFT JOIN sender_user su ON su.sender_id = s.id
		WHERE s.sender_id = ?
		  AND (su.user_id = ? OR s.requested_by = ?)
		  AND s.deleted_at IS NULL
		  AND s.is_enabled = 1
		LIMIT 1
	`, senderLabel, u.ID, u.ID).Scan(&senderDBID, &senderType)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": fmt.Sprintf("Sender %q is not available for this account.", senderLabel),
		})
	}
	if err != nil {
		return databaseError(context, err)
	}

	bodyStr := strings.TrimSpace(body)

	// Rate resolver & affordability check
	resolver, err := billing.ForUser(context.Context(), server.db, u.ID)
	if err != nil {
		return databaseError(context, err)
	}

	isMasking := strings.Contains(strings.ToLower(senderType), "masking")
	parts, _ := billing.CalculateParts(bodyStr)
	rates := resolver.RatesForUser(isMasking)
	maxRate := 0.0
	for _, r := range rates {
		if r > maxRate {
			maxRate = r
		}
	}
	estimatedCost := billing.RoundFloat(maxRate*float64(parts)*float64(len(numbers)), 4)

	var balance float64
	_ = server.db.QueryRow("SELECT balance FROM users WHERE id = ?", u.ID).Scan(&balance)
	if balance < estimatedCost {
		return context.Status(http.StatusPaymentRequired).JSON(fiber.Map{
			"message": "Insufficient balance for this batch.",
			"balance": balance,
		})
	}

	status := "processing"
	approval := (*string)(nil)
	if !u.SMSAutoApproved {
		status = "pending"
		pendingStr := "pending"
		approval = &pendingStr
	}

	recipients := make([]any, len(numbers))
	for i, n := range numbers {
		recipients[i] = map[string]any{"phone": n, "values": map[string]string{}}
	}
	srcConfigBytes, _ := json.Marshal(map[string]any{
		"recipients": recipients,
		"ip":         context.IP(),
	})

	now := time.Now()
	result, err := server.db.Exec(`
		INSERT INTO batches (user_id, sender_id, body, source_type, source_ip, source_config, status, approval, total_recipients, valid_count, created_at, updated_at)
		VALUES (?, ?, ?, 'api', ?, ?, ?, ?, ?, 0, ?, ?)
	`, u.ID, senderDBID, bodyStr, context.IP(), string(srcConfigBytes), status, approval, len(numbers), now, now)
	if err != nil {
		return databaseError(context, err)
	}

	batchID, _ := result.LastInsertId()

	if status == "processing" {
		go func(bid int64) {
			_ = server.batchProcessor.ProcessBatch(stdctx.Background(), bid)
		}(batchID)
	}

	return context.Status(http.StatusAccepted).JSON(fiber.Map{
		"batch_id":   batchID,
		"status":     status,
		"approval":   approval,
		"recipients": len(numbers),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/batches/:batch
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extShowBatch(context fiber.Ctx) error {
	u := getDevUser(context)
	batchParam := context.Params("batch")

	var id int
	var senderID, status string
	var approval *string
	var createdAt time.Time

	err := server.db.QueryRow(`
		SELECT b.id, COALESCE(s.sender_id, ''), b.status, b.approval, b.created_at
		FROM batches b
		LEFT JOIN senders s ON s.id = b.sender_id
		WHERE b.id = ? AND b.user_id = ?
	`, batchParam, u.ID).Scan(&id, &senderID, &status, &approval, &createdAt)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Batch not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"id":         id,
		"sender_id":  senderID,
		"status":     status,
		"approval":   approval,
		"created_at": createdAt.Format(time.RFC3339),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/ip-whitelists
// POST /api/v1/ip-whitelists
// DELETE /api/v1/ip-whitelists/:id
// ─────────────────────────────────────────────────────────────────────────────

func (server *Server) extIPWhitelists(context fiber.Ctx) error {
	u := getDevUser(context)

	rows, err := server.db.Query(`
		SELECT id, ip_address, COALESCE(label, ''), status, admin_note, created_at, reviewed_at
		FROM ip_whitelist_requests
		WHERE user_id = ?
		ORDER BY id DESC
	`, u.ID)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := []fiber.Map{}
	for rows.Next() {
		var id int
		var ip, label, status string
		var adminNote *string
		var createdAt time.Time
		var reviewedAt *time.Time

		if rows.Scan(&id, &ip, &label, &status, &adminNote, &createdAt, &reviewedAt) != nil {
			continue
		}
		items = append(items, fiber.Map{
			"id":          id,
			"ip_address":  ip,
			"label":       label,
			"status":      status,
			"admin_note":  adminNote,
			"created_at":  createdAt.Format(time.RFC3339),
			"reviewed_at": apiTimePtr(reviewedAt),
		})
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) extCreateIPWhitelist(context fiber.Ctx) error {
	u := getDevUser(context)

	var req struct {
		IPAddress string `json:"ip_address"`
		Label     string `json:"label"`
	}
	if err := json.Unmarshal(context.Body(), &req); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request body."})
	}

	ip := strings.TrimSpace(req.IPAddress)
	if ip == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The ip_address field is required."})
	}

	if errMsg := validatePublicIPStr(ip); errMsg != "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": errMsg})
	}

	// Check duplicate
	var existingID int
	var existingStatus string
	err := server.db.QueryRow(
		"SELECT id, status FROM ip_whitelist_requests WHERE user_id = ? AND ip_address = ? LIMIT 1",
		u.ID, ip,
	).Scan(&existingID, &existingStatus)
	if err == nil {
		return context.Status(http.StatusConflict).JSON(fiber.Map{
			"message": fmt.Sprintf("IP address %s is already registered on your account with status [%s].", ip, existingStatus),
			"data": fiber.Map{
				"id":         existingID,
				"ip_address": ip,
				"status":     existingStatus,
			},
		})
	}

	label := strings.TrimSpace(req.Label)
	now := time.Now()

	result, err := server.db.Exec(`
		INSERT INTO ip_whitelist_requests (user_id, ip_address, label, status, created_at, updated_at)
		VALUES (?, ?, ?, 'pending', ?, ?)
	`, u.ID, ip, nullString(label), now, now)
	if err != nil {
		return databaseError(context, err)
	}

	id, _ := result.LastInsertId()

	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "IP whitelist request submitted successfully. Awaiting administrator review.",
		"data": fiber.Map{
			"id":         id,
			"ip_address": ip,
			"label":      label,
			"status":     "pending",
			"created_at": now.Format(time.RFC3339),
		},
	})
}

func (server *Server) extDeleteIPWhitelist(context fiber.Ctx) error {
	u := getDevUser(context)
	id := context.Params("id")

	var ipAddress string
	err := server.db.QueryRow(
		"SELECT ip_address FROM ip_whitelist_requests WHERE id = ? AND user_id = ?",
		id, u.ID,
	).Scan(&ipAddress)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "IP whitelist entry not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	_, _ = server.db.Exec("DELETE FROM ip_whitelist_requests WHERE id = ? AND user_id = ?", id, u.ID)

	return context.JSON(fiber.Map{
		"message": fmt.Sprintf("IP whitelist entry for %s removed successfully.", ipAddress),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// IP validation helpers
// ─────────────────────────────────────────────────────────────────────────────

// validatePublicIPStr returns an error message string if the IP is invalid or private.
func validatePublicIPStr(val string) string {
	if strings.Contains(val, "/") {
		// CIDR
		ip, _, err := net.ParseCIDR(val)
		if err != nil {
			return "Invalid CIDR format."
		}
		if isPrivateIP(ip) {
			return "Private or reserved IP ranges cannot be whitelisted. Only public internet-facing IP blocks are permitted."
		}
		return ""
	}

	parsed := net.ParseIP(val)
	if parsed == nil {
		return "The ip_address must be a valid IPv4 or IPv6 address."
	}
	if isPrivateIP(parsed) {
		return "Private or reserved IP addresses cannot be whitelisted. Only public internet-facing IPs are permitted."
	}
	return ""
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		"0.0.0.0/8",
		"100.64.0.0/10",
		"169.254.0.0/16",
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}


