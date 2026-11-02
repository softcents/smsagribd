package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type operatorPayload struct {
	CountryID       int64    `json:"country_id"`
	Name            string   `json:"name"`
	Codes           string   `json:"codes"`
	MaskingPrice    *float64 `json:"masking_price"`
	NonMaskingPrice *float64 `json:"non_masking_price"`
	IsActive        *bool    `json:"is_active"`
}

type operatorDetailItem struct {
	ID              int64   `json:"id"`
	CountryID       int64   `json:"country_id"`
	Name            string  `json:"name"`
	Codes           string  `json:"codes"`
	MaskingPrice    float64 `json:"masking_price"`
	NonMaskingPrice float64 `json:"non_masking_price"`
	IsActive        bool    `json:"is_active"`
}

var operatorCodePattern = regexp.MustCompile(`^[0-9]{1,6}$`)

func operatorInvalid(context fiber.Ctx, message string) error {
	return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": message})
}

func displayOperatorCodes(raw string) string {
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return raw
	}
	return strings.Join(list, ", ")
}

func parseOperatorCodes(raw string) ([]string, string) {
	fields := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == ';' || r == '\n' || r == '\t' })
	seen := map[string]bool{}
	codes := []string{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if !operatorCodePattern.MatchString(field) {
			return nil, "Codes must be digits only, separated by commas (e.g. 17, 13)."
		}
		if !seen[field] {
			seen[field] = true
			codes = append(codes, field)
		}
	}
	if len(codes) == 0 {
		return nil, "At least one code is required."
	}
	return codes, ""
}

func (server *Server) validateOperator(payload *operatorPayload, excludeID int64) ([]string, string) {
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.CountryID < 1 {
		return nil, "Country is required."
	}
	if payload.Name == "" {
		return nil, "Name is required."
	}
	codes, message := parseOperatorCodes(payload.Codes)
	if message != "" {
		return nil, message
	}
	if (payload.MaskingPrice != nil && *payload.MaskingPrice < 0) || (payload.NonMaskingPrice != nil && *payload.NonMaskingPrice < 0) {
		return nil, "Prices cannot be negative."
	}

	var exists int64
	if err := server.db.QueryRow("SELECT id FROM countries WHERE id = CAST(? AS BIGINT)", payload.CountryID).Scan(&exists); err != nil {
		return nil, "Selected country does not exist."
	}

	rows, err := server.db.Query("SELECT name, COALESCE(codes, '') FROM operators WHERE country_id = CAST(? AS BIGINT) AND id <> CAST(? AS BIGINT)", payload.CountryID, excludeID)
	if err != nil {
		return nil, "Unable to check existing operator codes."
	}
	defer rows.Close()
	for rows.Next() {
		var otherName, raw string
		if err := rows.Scan(&otherName, &raw); err != nil {
			return nil, "Unable to check existing operator codes."
		}
		var others []string
		_ = json.Unmarshal([]byte(raw), &others)
		for _, code := range codes {
			for _, other := range others {
				if code == other {
					return nil, "Code " + code + " is already used by " + otherName + "."
				}
			}
		}
	}
	return codes, ""
}

func (server *Server) operatorDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Operator not found."})
	}
	var item operatorDetailItem
	var raw string
	err = server.db.QueryRow(`SELECT id, CAST(country_id AS BIGINT), name, COALESCE(codes, ''), masking_price, non_masking_price, is_active FROM operators WHERE id = ?`, id).
		Scan(&item.ID, &item.CountryID, &item.Name, &raw, &item.MaskingPrice, &item.NonMaskingPrice, &item.IsActive)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Operator not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}
	item.Codes = displayOperatorCodes(raw)
	return context.JSON(fiber.Map{"data": item})
}

func (server *Server) createOperator(context fiber.Ctx) error {
	var payload operatorPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return operatorInvalid(context, "Invalid request payload.")
	}
	codes, message := server.validateOperator(&payload, 0)
	if message != "" {
		return operatorInvalid(context, message)
	}
	codesJSON, _ := json.Marshal(codes)
	masking, nonMasking := 0.0, 0.0
	if payload.MaskingPrice != nil {
		masking = *payload.MaskingPrice
	}
	if payload.NonMaskingPrice != nil {
		nonMasking = *payload.NonMaskingPrice
	}
	active := true
	if payload.IsActive != nil {
		active = *payload.IsActive
	}
	var id int64
	err := server.db.QueryRow(`
		INSERT INTO operators (country_id, name, codes, masking_price, non_masking_price, is_active, created_at, updated_at)
		VALUES (CAST(? AS BIGINT), ?, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id
	`, payload.CountryID, payload.Name, string(codesJSON), masking, nonMasking, active).Scan(&id)
	if err != nil {
		return databaseError(context, err)
	}
	return context.Status(http.StatusCreated).JSON(fiber.Map{"message": "Operator created successfully.", "data": fiber.Map{"id": id}})
}

func (server *Server) updateOperator(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Operator not found."})
	}
	var payload operatorPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return operatorInvalid(context, "Invalid request payload.")
	}
	var existing int64
	if err := server.db.QueryRow("SELECT id FROM operators WHERE id = ?", id).Scan(&existing); err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Operator not found."})
	}
	codes, message := server.validateOperator(&payload, id)
	if message != "" {
		return operatorInvalid(context, message)
	}
	codesJSON, _ := json.Marshal(codes)
	masking, nonMasking := 0.0, 0.0
	if payload.MaskingPrice != nil {
		masking = *payload.MaskingPrice
	}
	if payload.NonMaskingPrice != nil {
		nonMasking = *payload.NonMaskingPrice
	}
	_, err = server.db.Exec(`
		UPDATE operators
		SET country_id = CAST(? AS BIGINT), name = ?, codes = ?, masking_price = ?, non_masking_price = ?,
		    is_active = COALESCE(?, is_active), updated_at = NOW()
		WHERE id = ?
	`, payload.CountryID, payload.Name, string(codesJSON), masking, nonMasking, payload.IsActive, id)
	if err != nil {
		return databaseError(context, err)
	}
	return context.JSON(fiber.Map{"message": "Operator updated successfully."})
}
