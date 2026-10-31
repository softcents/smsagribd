package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

var allowedSettings = map[string]map[string]struct{}{
	"general": {
		"appName": {},
		"logoUrl": {},
	},
	"sms": {
		"sms_enabled": {}, "default_sms_type": {}, "default_encoding": {}, "auto_detect_unicode": {}, "force_unicode": {},
		"strip_incompatible_chars": {}, "allow_long_sms": {}, "max_sms_parts": {}, "duplicate_protection_enabled": {},
		"duplicate_window_seconds": {}, "validation_mode": {}, "default_country_code": {}, "default_sender_id": {},
		"log_level": {}, "log_payloads": {}, "log_retention_days": {}, "message_id_prefix": {}, "message_id_format": {},
	},
	"security": {
		"ip_whitelist_enforced": {}, "allow_all_when_whitelist_empty": {}, "auto_approve_api_requests": {},
		"web_portal_bypass_enabled": {}, "trusted_portal_ips": {}, "blocked_message": {},
	},
}

func (server *Server) settings(context fiber.Ctx) error {
	group, ok := validSettingGroup(context)
	if !ok {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Settings group not found."})
	}

	values, err := server.readSettings(group)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"data": values})
}

func (server *Server) updateSettings(context fiber.Ctx) error {
	group, ok := validSettingGroup(context)
	if !ok {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Settings group not found."})
	}

	values := map[string]json.RawMessage{}
	if err := json.Unmarshal(context.Body(), &values); err != nil || len(values) == 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied settings are invalid."})
	}

	transaction, err := server.db.Begin()
	if err != nil {
		return databaseError(context, err)
	}
	defer transaction.Rollback()

	for name, payload := range values {
		if _, allowed := allowedSettings[group][name]; !allowed || !json.Valid(payload) {
			return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied settings are invalid."})
		}

		result, err := transaction.Exec(`
			UPDATE settings SET payload = ?, updated_at = NOW()
			WHERE `+"`group`"+` = ? AND name = ? AND locked = 0
		`, payload, group, name)
		if err != nil {
			return databaseError(context, err)
		}
		updated, _ := result.RowsAffected()
		if updated == 0 {
			var locked bool
			err := transaction.QueryRow("SELECT locked FROM settings WHERE `group` = ? AND name = ?", group, name).Scan(&locked)
			if err == sql.ErrNoRows {
				return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "A requested setting does not exist."})
			}
			if err != nil {
				return databaseError(context, err)
			}
			if locked {
				return context.Status(http.StatusForbidden).JSON(fiber.Map{"message": "A requested setting is locked."})
			}
		}
	}

	if err := transaction.Commit(); err != nil {
		return databaseError(context, err)
	}

	updatedValues, err := server.readSettings(group)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Settings saved successfully.", "data": updatedValues})
}

func validSettingGroup(context fiber.Ctx) (string, bool) {
	group := context.Params("group")
	_, ok := allowedSettings[group]

	return group, ok
}

func (server *Server) readSettings(group string) (map[string]any, error) {
	rows, err := server.db.Query("SELECT name, payload FROM settings WHERE `group` = ?", group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]any)
	for rows.Next() {
		var name string
		var payload []byte
		if err := rows.Scan(&name, &payload); err != nil {
			return nil, err
		}
		var value any
		if err := json.Unmarshal(payload, &value); err != nil {
			return nil, err
		}
		values[name] = value
	}

	return values, rows.Err()
}
