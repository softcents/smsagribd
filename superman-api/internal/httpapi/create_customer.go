package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

type createCustomerPayload struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	BusinessName    string `json:"business_name"`
	Status          string `json:"status"`
	Password        string `json:"password"`
	APIAccess       bool   `json:"api_access"`
	SMSAutoApproved bool   `json:"sms_auto_approved"`
	TariffPlanID    *int64 `json:"tariff_plan_id"`
}

func customerFieldOrNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func invalidCustomer(context fiber.Ctx, message string) error {
	return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": message})
}

// createCustomer lets a signed-in superman admin create a customer account.
func (server *Server) createCustomer(context fiber.Ctx) error {
	payload := createCustomerPayload{}
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return invalidCustomer(context, "Invalid request payload.")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Phone = strings.TrimSpace(payload.Phone)
	payload.BusinessName = strings.TrimSpace(payload.BusinessName)
	payload.Status = strings.TrimSpace(payload.Status)
	if payload.Status == "" {
		payload.Status = "approved"
	}

	if payload.Name == "" {
		return invalidCustomer(context, "Full name is required.")
	}
	if payload.Email == "" || !strings.Contains(payload.Email, "@") {
		return invalidCustomer(context, "A valid email address is required.")
	}
	if len(payload.Password) < 8 {
		return invalidCustomer(context, "Password must be at least 8 characters.")
	}
	if len(payload.Password) > 72 {
		return invalidCustomer(context, "Password must be at most 72 characters.")
	}

	var existing int64
	err := server.db.QueryRow(`SELECT id FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1`, payload.Email).Scan(&existing)
	if err == nil {
		return invalidCustomer(context, "This email address is already registered.")
	}
	if err != sql.ErrNoRows {
		return databaseError(context, err)
	}
	if payload.Phone != "" {
		err = server.db.QueryRow(`SELECT id FROM users WHERE phone = ? AND deleted_at IS NULL LIMIT 1`, payload.Phone).Scan(&existing)
		if err == nil {
			return invalidCustomer(context, "This phone number is already registered.")
		}
		if err != sql.ErrNoRows {
			return databaseError(context, err)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to create customer."})
	}

	tx, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO users (name, email, phone, password, type, status, kyc_status, business_name,
			is_admin_created, created_by_admin, sms_auto_approved, api_access, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'customer', ?, 'pending', ?, ?, ?, ?, ?, NOW(), NOW())
	`, payload.Name, payload.Email, customerFieldOrNil(payload.Phone), string(hash), payload.Status,
		customerFieldOrNil(payload.BusinessName), true, true, payload.SMSAutoApproved, payload.APIAccess)
	if err != nil {
		return databaseError(context, err)
	}

	var id int64
	if err := tx.QueryRow(`SELECT id FROM users WHERE email = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1`, payload.Email).Scan(&id); err != nil {
		return databaseError(context, err)
	}

	if payload.TariffPlanID != nil && *payload.TariffPlanID > 0 {
		if _, err := tx.Exec(`INSERT INTO tariff_plan_assignments (user_id, tariff_plan_id, status, effective_from, created_at, updated_at) VALUES (?, ?, 'active', NOW(), NOW(), NOW())`, id, *payload.TariffPlanID); err != nil {
			return databaseError(context, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return databaseError(context, err)
	}

	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Customer created successfully.",
		"data":    fiber.Map{"id": id},
	})
}
