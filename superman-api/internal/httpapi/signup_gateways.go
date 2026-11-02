package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type signupGatewayPayload struct {
	Name           string `json:"name"`
	Priority       int    `json:"priority"`
	EndpointURL    string `json:"endpoint_url"`
	HTTPMethod     string `json:"http_method"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	AccessToken    string `json:"access_token"`
	SenderID       string `json:"sender_id"`
	PhoneFormat    string `json:"phone_format"`
	OTPTemplate    string `json:"otp_template"`
	IsActive       *bool  `json:"is_active"`
}

type signupGatewayItem struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Priority       int    `json:"priority"`
	EndpointURL    string `json:"endpoint_url"`
	HTTPMethod     string `json:"http_method"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	SenderID       string `json:"sender_id"`
	PhoneFormat    string `json:"phone_format"`
	OTPTemplate    string `json:"otp_template"`
	IsActive       bool   `json:"is_active"`
	HasToken       bool   `json:"has_token"`
}

const signupGatewayColumns = `id, name, priority, endpoint_url, http_method, timeout_seconds, sender_id, phone_format, otp_template, is_active, (access_token <> '')`

func signupGatewayInvalid(context fiber.Ctx, message string) error {
	return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": message})
}

func scanSignupGateway(scanner interface{ Scan(dest ...any) error }) (signupGatewayItem, error) {
	var item signupGatewayItem
	err := scanner.Scan(&item.ID, &item.Name, &item.Priority, &item.EndpointURL, &item.HTTPMethod, &item.TimeoutSeconds,
		&item.SenderID, &item.PhoneFormat, &item.OTPTemplate, &item.IsActive, &item.HasToken)
	return item, err
}

func (payload *signupGatewayPayload) normalize(requireToken bool) string {
	payload.Name = strings.TrimSpace(payload.Name)
	payload.EndpointURL = strings.TrimSpace(payload.EndpointURL)
	payload.HTTPMethod = strings.ToUpper(strings.TrimSpace(payload.HTTPMethod))
	payload.AccessToken = strings.TrimSpace(payload.AccessToken)
	payload.SenderID = strings.TrimSpace(payload.SenderID)
	payload.PhoneFormat = strings.TrimSpace(payload.PhoneFormat)
	payload.OTPTemplate = strings.TrimSpace(payload.OTPTemplate)
	if payload.HTTPMethod == "" {
		payload.HTTPMethod = "POST"
	}
	if payload.PhoneFormat == "" {
		payload.PhoneFormat = "local_bd"
	}
	if payload.Priority < 1 {
		payload.Priority = 1
	}
	if payload.TimeoutSeconds == 0 {
		payload.TimeoutSeconds = 15
	}
	switch {
	case payload.Name == "":
		return "Gateway name is required."
	case !strings.HasPrefix(payload.EndpointURL, "http://") && !strings.HasPrefix(payload.EndpointURL, "https://"):
		return "Endpoint URL must start with http:// or https://."
	case payload.HTTPMethod != "POST" && payload.HTTPMethod != "GET":
		return "HTTP method must be POST or GET."
	case payload.TimeoutSeconds < 1 || payload.TimeoutSeconds > 120:
		return "Timeout must be between 1 and 120 seconds."
	case payload.PhoneFormat != "local_bd" && payload.PhoneFormat != "intl_880" && payload.PhoneFormat != "intl_plus":
		return "Invalid phone format."
	case payload.SenderID == "":
		return "Sender ID is required."
	case !strings.Contains(payload.OTPTemplate, "{otp}"):
		return "OTP message template must contain {otp}."
	case requireToken && payload.AccessToken == "":
		return "Bearer access token is required."
	}
	return ""
}

func (server *Server) signupGateways(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if search != "" {
		where += " AND (name ILIKE ? OR endpoint_url ILIKE ? OR sender_id ILIKE ?)"
		term := "%" + search + "%"
		args = append(args, term, term, term)
	}
	var total int
	if err := server.db.QueryRow("SELECT COUNT(*) FROM signup_register_gateways "+where, args...).Scan(&total); err != nil {
		return databaseError(context, err)
	}
	listArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)
	rows, err := server.db.Query("SELECT "+signupGatewayColumns+" FROM signup_register_gateways "+where+" ORDER BY priority ASC, id ASC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()
	items := make([]signupGatewayItem, 0)
	for rows.Next() {
		item, err := scanSignupGateway(rows)
		if err != nil {
			return databaseError(context, err)
		}
		items = append(items, item)
	}
	return context.JSON(fiber.Map{"data": items, "meta": listMeta{Page: page, PerPage: perPage, Total: total}})
}

func (server *Server) signupGatewayDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	item, err := scanSignupGateway(server.db.QueryRow("SELECT "+signupGatewayColumns+" FROM signup_register_gateways WHERE id = ? AND deleted_at IS NULL", id))
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}
	return context.JSON(fiber.Map{"data": item})
}
