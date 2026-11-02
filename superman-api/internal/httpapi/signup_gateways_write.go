package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func (server *Server) createSignupGateway(context fiber.Ctx) error {
	var payload signupGatewayPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return signupGatewayInvalid(context, "Invalid request payload.")
	}
	if message := payload.normalize(true); message != "" {
		return signupGatewayInvalid(context, message)
	}
	active := true
	if payload.IsActive != nil {
		active = *payload.IsActive
	}
	enc, err := server.getEncrypter()
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Encryption configuration error."})
	}
	encrypted, err := enc.EncryptArray(map[string]any{"token": payload.AccessToken})
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to encrypt token."})
	}
	var id int64
	err = server.db.QueryRow(`
		INSERT INTO signup_register_gateways
			(name, priority, endpoint_url, http_method, timeout_seconds, access_token, sender_id, phone_format, otp_template, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id
	`, payload.Name, payload.Priority, payload.EndpointURL, payload.HTTPMethod, payload.TimeoutSeconds, encrypted,
		payload.SenderID, payload.PhoneFormat, payload.OTPTemplate, active).Scan(&id)
	if err != nil {
		return databaseError(context, err)
	}
	return context.Status(http.StatusCreated).JSON(fiber.Map{"message": "Gateway created successfully.", "data": fiber.Map{"id": id}})
}

func (server *Server) updateSignupGateway(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	var payload signupGatewayPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return signupGatewayInvalid(context, "Invalid request payload.")
	}
	if message := payload.normalize(false); message != "" {
		return signupGatewayInvalid(context, message)
	}
	var existing int64
	if err := server.db.QueryRow("SELECT id FROM signup_register_gateways WHERE id = ? AND deleted_at IS NULL", id).Scan(&existing); err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	sets := []string{"name = ?", "priority = ?", "endpoint_url = ?", "http_method = ?", "timeout_seconds = ?", "sender_id = ?", "phone_format = ?", "otp_template = ?", "updated_at = NOW()"}
	args := []any{payload.Name, payload.Priority, payload.EndpointURL, payload.HTTPMethod, payload.TimeoutSeconds, payload.SenderID, payload.PhoneFormat, payload.OTPTemplate}
	if payload.IsActive != nil {
		sets = append(sets, "is_active = ?")
		args = append(args, *payload.IsActive)
	}
	if payload.AccessToken != "" {
		enc, err := server.getEncrypter()
		if err != nil {
			return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Encryption configuration error."})
		}
		encrypted, err := enc.EncryptArray(map[string]any{"token": payload.AccessToken})
		if err != nil {
			return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to encrypt token."})
		}
		sets = append(sets, "access_token = ?")
		args = append(args, encrypted)
	}
	args = append(args, id)
	if _, err := server.db.Exec(fmt.Sprintf("UPDATE signup_register_gateways SET %s WHERE id = ?", strings.Join(sets, ", ")), args...); err != nil {
		return databaseError(context, err)
	}
	return context.JSON(fiber.Map{"message": "Gateway updated successfully."})
}

func (server *Server) toggleSignupGateway(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	var active bool
	err = server.db.QueryRow("UPDATE signup_register_gateways SET is_active = NOT is_active, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL RETURNING is_active", id).Scan(&active)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}
	return context.JSON(fiber.Map{"message": "Gateway status updated.", "data": fiber.Map{"is_active": active}})
}

func (server *Server) deleteSignupGateway(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	result, err := server.db.Exec("UPDATE signup_register_gateways SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return databaseError(context, err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	return context.JSON(fiber.Map{"message": "Gateway deleted."})
}
