package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"smswhiteinfo/customer-api/internal/payment"
)

func (server *Server) handleSmsWebhook(context fiber.Ctx) error {
	gateway := context.Params("gateway")

	// Process incoming DLR status if payload contains message ID and status
	msgID := context.Query("Message_ID")
	status := context.Query("Status")

	if msgID == "" {
		var bodyMap map[string]any
		if err := json.Unmarshal(context.Body(), &bodyMap); err == nil {
			if v, ok := bodyMap["Message_ID"].(string); ok {
				msgID = v
			} else if v, ok := bodyMap["message_id"].(string); ok {
				msgID = v
			} else if v, ok := bodyMap["msg_id"].(string); ok {
				msgID = v
			}

			if s, ok := bodyMap["Status"].(string); ok {
				status = s
			} else if s, ok := bodyMap["status"].(string); ok {
				status = s
			}
		}
	}

	if msgID != "" && status != "" {
		sUpper := strings.ToUpper(status)
		if strings.Contains(sUpper, "DELIV") || sUpper == "SUCCESS" {
			_, _ = server.db.ExecContext(context.Context(), `
				UPDATE messages 
				SET status = 'delivered', delivered_at = NOW(), delivery_status = ?, poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() 
				WHERE gateway_message_id = ? AND status = 'sent'
			`, status, msgID)
		} else if strings.Contains(sUpper, "FAIL") || strings.Contains(sUpper, "REJECT") {
			var mid int64
			err := server.db.QueryRowContext(context.Context(), "SELECT id FROM messages WHERE gateway_message_id = ? AND status = 'sent'", msgID).Scan(&mid)
			if err == nil && mid > 0 {
				_, _ = payment.CompleteDeposit(context.Context(), server.db, gateway, "", false, nil)
			}
		}
	}

	return context.JSON(fiber.Map{
		"message": "SMS webhook received.",
	})
}

// handlePaymentWebhook handles PipraPay (and future) payment gateway webhooks.
//
// Idempotency design (B.1 compliant):
//  1. Parse payload, extract event_id.
//  2. Check payment_webhooks for a SUCCESSFULLY PROCESSED record (processed_at IS NOT NULL).
//     If found → return 200 "already processed" without re-crediting.
//  3. Look up the deposit row by (gateway, event_id).
//  4. Call the provider verify-payment API.
//  5. Compare amounts (decimal, tolerance 0.01) and currency (case-insensitive).
//  6. On any failure in steps 3-5: INSERT/UPSERT payment_webhooks with processed_at=NULL
//     (retryable) and log the rejection reason. Return 4xx/5xx so the caller retries.
//  7. Call CompleteDeposit atomically.
//  8. On success: UPDATE/INSERT payment_webhooks with processed_at=NOW().
//
// A webhook row with processed_at=NULL is retryable on every call.
// A webhook row with processed_at IS NOT NULL is idempotent-complete.
func (server *Server) handlePaymentWebhook(context fiber.Ctx) error {
	gateway := strings.ToLower(context.Params("gateway"))

	if gateway != "piprapay" {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("Unknown payment gateway: %s", gateway),
		})
	}

	var payload map[string]any
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid JSON payload."})
	}

	// Extract event_id / pp_id
	eventID := ""
	for _, key := range []string{"pp_id", "transaction_id", "payment_id", "event_id", "id"} {
		if val, ok := payload[key].(string); ok && val != "" {
			eventID = val
			break
		}
	}

	if eventID == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Missing payment transaction identifier."})
	}

	statusStr := fmt.Sprintf("%v", payload["status"])
	payloadJSON, _ := json.Marshal(payload)
	headersJSON, _ := json.Marshal(context.GetReqHeaders())

	// ─── Step 2: Idempotency check — has this event been fully processed? ───────
	// We only consider it done if status = 'processed'.
	var existingID int64
	var existingStatus string
	err := server.db.QueryRowContext(context.Context(), `
		SELECT id, status 
		FROM payment_webhooks 
		WHERE gateway = ? AND event_id = ?
		LIMIT 1
	`, gateway, eventID).Scan(&existingID, &existingStatus)

	if err == nil && existingStatus == "processed" {
		// Already fully processed — idempotent success response.
		return context.JSON(fiber.Map{"message": "Webhook already processed."})
	}

	// ─── Step 3: Look up the deposit ─────────────────────────────────────────────
	var depositAmount float64
	var depositCurrency, depositStatus string
	err = server.db.QueryRowContext(context.Context(), `
		SELECT amount, currency, status 
		FROM deposits 
		WHERE payment_method = ? AND transaction_id = ?
	`, gateway, eventID).Scan(&depositAmount, &depositCurrency, &depositStatus)

	if err != nil {
		reason := fmt.Sprintf("Deposit not found for gateway=%s event_id=%s", gateway, eventID)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Deposit not found."})
	}

	if depositStatus == "completed" {
		// Deposit already completed via a different path — mark webhook processed.
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, true)
		return context.JSON(fiber.Map{"message": "Deposit already completed."})
	}

	// ─── Step 4: Verify with provider ─────────────────────────────────────────────
	apiKey := os.Getenv("PIPRAPAY_API_KEY")
	baseURL := os.Getenv("PIPRAPAY_BASE_URL")
	if apiKey == "" {
		apiKey = "dummy-piprapay-key"
	}
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9999"
	}

	driver, err := payment.NewPipraPayDriver(apiKey, baseURL)
	if err != nil {
		reason := fmt.Sprintf("Payment provider initialization failed: %v", err)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Payment provider initialization failed."})
	}

	verified, err := driver.Verify(context.Context(), eventID)
	if err != nil {
		reason := fmt.Sprintf("Provider verify error for event_id=%s: %v", eventID, err)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusBadGateway).JSON(fiber.Map{
			"message": fmt.Sprintf("Payment provider verification failed: %v", err),
		})
	}

	if !verified.Successful || strings.ToLower(verified.GatewayStatus) != "completed" {
		reason := fmt.Sprintf("Provider status not completed for event_id=%s: status=%s successful=%v", eventID, verified.GatewayStatus, verified.Successful)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Payment provider reported incomplete or failed status.",
		})
	}

	// ─── Step 5: Amount and currency validation ─────────────────────────────────
	if math.Abs(verified.Amount-depositAmount) > 0.01 {
		reason := fmt.Sprintf("Amount mismatch for event_id=%s: verified=%.4f deposit=%.4f", eventID, verified.Amount, depositAmount)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": fmt.Sprintf("Verified amount %.2f does not match deposit amount %.2f", verified.Amount, depositAmount),
		})
	}

	if verified.Currency != "" && strings.ToUpper(verified.Currency) != strings.ToUpper(depositCurrency) {
		reason := fmt.Sprintf("Currency mismatch for event_id=%s: verified=%s deposit=%s", eventID, verified.Currency, depositCurrency)
		log.Printf("[PAYMENT WEBHOOK REJECT] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": fmt.Sprintf("Verified currency %s does not match deposit currency %s", verified.Currency, depositCurrency),
		})
	}

	// ─── Step 6: Complete the deposit (atomic DB transaction) ──────────────────
	completed, err := payment.CompleteDeposit(context.Context(), server.db, gateway, eventID, true, verified.Raw)
	if err != nil {
		reason := fmt.Sprintf("CompleteDeposit error for event_id=%s: %v", eventID, err)
		log.Printf("[PAYMENT WEBHOOK ERROR] %s", reason)
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, false)
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": fmt.Sprintf("Error completing deposit: %v", err)})
	}

	if !completed {
		// Deposit was already completed by a concurrent call — idempotent.
		server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, true)
		return context.JSON(fiber.Map{"message": "Webhook received, deposit already completed."})
	}

	// ─── Step 7: Mark webhook as processed only AFTER successful credit ─────────
	server.recordWebhookAttempt(context, gateway, eventID, statusStr, string(payloadJSON), string(headersJSON), existingID, true)

	return context.JSON(fiber.Map{"message": "Webhook received."})
}

// recordWebhookAttempt upserts a payment_webhooks record.
// If processed=true, sets processed_at=NOW(). If false, leaves processed_at=NULL (retryable).
func (server *Server) recordWebhookAttempt(context fiber.Ctx, gateway, eventID, status, payloadJSON, headersJSON string, existingID int64, processed bool) {
	recordStatus := status
	if processed {
		recordStatus = "processed"
	} else if recordStatus == "processed" || recordStatus == "" {
		recordStatus = "failed"
	}

	if existingID > 0 {
		_, _ = server.db.ExecContext(context.Context(), `
			UPDATE payment_webhooks 
			SET status = ?, payload = ?, headers = ?, updated_at = NOW() 
			WHERE id = ?
		`, recordStatus, payloadJSON, headersJSON, existingID)
	} else {
		_, _ = server.db.ExecContext(context.Context(), `
			INSERT INTO payment_webhooks (gateway, event_id, event_type, status, payload, headers, created_at, updated_at)
			VALUES (?, ?, 'payment.updated', ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE status = VALUES(status), payload = VALUES(payload), headers = VALUES(headers), updated_at = NOW()
		`, gateway, eventID, recordStatus, payloadJSON, headersJSON)
	}
}

func (server *Server) handlePaymentCallback(context fiber.Ctx) error {
	gateway := context.Params("gateway")
	depositID := context.Params("deposit")

	var status string
	err := server.db.QueryRowContext(context.Context(), "SELECT status FROM deposits WHERE id = ? AND payment_method = ?", depositID, gateway).Scan(&status)
	if err != nil {
		return context.Status(http.StatusNotFound).SendString("Deposit not found.")
	}

	redirectURL := fmt.Sprintf("%s/customer/deposits?payment_status=%s", server.config.FrontendOrigin, status)
	return context.Redirect().To(redirectURL)
}

func (server *Server) handleKycFileStream(context fiber.Ctx) error {
	// Guard: check superman/admin authorization
	authHeader := context.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := server.tokens.Parse(token)
	if err != nil {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	var userType string
	err = server.db.QueryRowContext(context.Context(), "SELECT type FROM users WHERE id = ?", claims.UserID).Scan(&userType)
	if err != nil || (userType != "superman" && userType != "admin") {
		return context.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Admin authorization required."})
	}

	return server.kycService.StreamKycFile(context)
}

func (server *Server) handleFailedReportDownload(context fiber.Ctx) error {
	authHeader := context.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := server.tokens.Parse(token)
	if err != nil {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	var userType string
	_ = server.db.QueryRowContext(context.Context(), "SELECT type FROM users WHERE id = ?", claims.UserID).Scan(&userType)
	isAdmin := userType == "superman" || userType == "admin"

	return server.kycService.DownloadFailedReport(context, claims.UserID, isAdmin)
}
