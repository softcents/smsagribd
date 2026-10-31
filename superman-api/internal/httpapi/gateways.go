package httpapi

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"smswhiteinfo/superman-api/internal/crypto"
)

type gatewayPayload struct {
	Name                 string         `json:"name"`
	Driver               string         `json:"driver"`
	Credentials          map[string]any `json:"credentials"`
	Priority             int            `json:"priority"`
	IsEnabled            *bool          `json:"is_enabled"`
	IsOtpEnabled         *bool          `json:"is_otp_enabled"`
	SupportsBulk         *bool          `json:"supports_bulk"`
	RequestsPerSecond    int            `json:"requests_per_second"`
	TimeoutSeconds       int            `json:"timeout_seconds"`
	DlrSupport           *bool          `json:"dlr_support"`
	CostMode             string         `json:"cost_mode"`
	DippingRequired      *bool          `json:"dipping_required"`
	HasThirdPartyDipping *bool          `json:"has_third_party_dipping"`
}

func (server *Server) getEncrypter() (*crypto.LaravelEncrypter, error) {
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		return nil, fmt.Errorf("APP_KEY not set")
	}
	return crypto.NewLaravelEncrypter(appKey)
}

func extractRouteCode(credMap map[string]any) *string {
	if val, ok := credMap["enabled_carriers"]; ok {
		switch v := val.(type) {
		case string:
			trimmed := strings.TrimSpace(v)
			if trimmed != "" {
				return &trimmed
			}
		case []any:
			var parts []string
			for _, item := range v {
				s := strings.TrimSpace(fmt.Sprintf("%v", item))
				if s != "" && s != "<nil>" {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				joined := strings.Join(parts, ", ")
				return &joined
			}
		case []string:
			if len(v) > 0 {
				joined := strings.Join(v, ", ")
				return &joined
			}
		}
	}
	if val, ok := credMap["route_code"]; ok {
		s := strings.TrimSpace(fmt.Sprintf("%v", val))
		if s != "" && s != "<nil>" {
			return &s
		}
	}
	return nil
}

type GatewayListItem struct {
	ID                int64   `json:"id"`
	Name              string  `json:"name"`
	Driver            string  `json:"driver"`
	Priority          int     `json:"priority"`
	IsEnabled         bool    `json:"is_enabled"`
	Enabled           bool    `json:"enabled"`
	IsOtpEnabled      bool    `json:"is_otp_enabled"`
	OtpEnabled        bool    `json:"otp_enabled"`
	RequestsPerSecond int     `json:"requests_per_second"`
	HealthStatus      string  `json:"health_status"`
	CostMode          string  `json:"cost_mode"`
	DlrSupport        bool    `json:"dlr_support"`
	RouteCode         *string `json:"route_code"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

func (server *Server) gateways(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT id, name, driver, priority, is_otp_enabled, is_enabled, requests_per_second,
		       health_status, cost_mode, dlr_support, credentials,
		       DATE_FORMAT(created_at, '%b %d, %Y %H:%i:%s'),
		       DATE_FORMAT(updated_at, '%b %d, %Y %H:%i:%s')
		FROM gateways
		WHERE deleted_at IS NULL
		ORDER BY priority ASC, id ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	enc, _ := server.getEncrypter()

	items := make([]GatewayListItem, 0)
	for rows.Next() {
		var item GatewayListItem
		var rawCreds sql.NullString
		var created, updated sql.NullString

		if err := rows.Scan(
			&item.ID, &item.Name, &item.Driver, &item.Priority, &item.IsOtpEnabled, &item.IsEnabled,
			&item.RequestsPerSecond, &item.HealthStatus, &item.CostMode, &item.DlrSupport,
			&rawCreds, &created, &updated,
		); err != nil {
			return databaseError(context, err)
		}

		item.Enabled = item.IsEnabled
		item.OtpEnabled = item.IsOtpEnabled
		if created.Valid {
			item.CreatedAt = created.String
		}
		if updated.Valid {
			item.UpdatedAt = updated.String
		}

		if rawCreds.Valid && rawCreds.String != "" && enc != nil {
			if credMap, err := enc.DecryptArray(rawCreds.String); err == nil {
				item.RouteCode = extractRouteCode(credMap)
			}
		}

		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}

func (server *Server) createGateway(context fiber.Ctx) error {
	var payload gatewayPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	payload.Name = strings.TrimSpace(payload.Name)
	payload.Driver = strings.ToLower(strings.TrimSpace(payload.Driver))

	if payload.Name == "" || payload.Driver == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Name and Driver are required."})
	}

	enc, err := server.getEncrypter()
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Encryption configuration error: " + err.Error()})
	}

	var encryptedCreds string
	if len(payload.Credentials) > 0 {
		encCreds, err := enc.EncryptArray(payload.Credentials)
		if err != nil {
			return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to encrypt credentials: " + err.Error()})
		}
		encryptedCreds = encCreds
	}

	priority := payload.Priority
	if priority < 1 {
		priority = 1
	}

	isEnabled := true
	if payload.IsEnabled != nil {
		isEnabled = *payload.IsEnabled
	}

	isOtpEnabled := true
	if payload.IsOtpEnabled != nil {
		isOtpEnabled = *payload.IsOtpEnabled
	}

	supportsBulk := true
	if payload.SupportsBulk != nil {
		supportsBulk = *payload.SupportsBulk
	}

	rps := payload.RequestsPerSecond
	if rps < 1 {
		rps = 60
	}

	timeoutSec := payload.TimeoutSeconds
	if timeoutSec < 1 {
		timeoutSec = 10
	}

	dlrSupport := true
	if payload.DlrSupport != nil {
		dlrSupport = *payload.DlrSupport
	}

	costMode := payload.CostMode
	if costMode == "" {
		costMode = "manual"
	}

	dippingRequired := false
	if payload.DippingRequired != nil {
		dippingRequired = *payload.DippingRequired
	}

	hasThirdPartyDipping := false
	if payload.HasThirdPartyDipping != nil {
		hasThirdPartyDipping = *payload.HasThirdPartyDipping
	}

	res, err := server.db.Exec(`
		INSERT INTO gateways (
			name, driver, credentials, priority, is_enabled, is_otp_enabled, supports_bulk,
			requests_per_second, timeout_seconds, dlr_support, cost_mode,
			dipping_required, has_third_party_dipping, health_status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'healthy', NOW(), NOW())
	`, payload.Name, payload.Driver, encryptedCreds, priority, isEnabled, isOtpEnabled, supportsBulk,
		rps, timeoutSec, dlrSupport, costMode, dippingRequired, hasThirdPartyDipping)
	if err != nil {
		return databaseError(context, err)
	}

	id, _ := res.LastInsertId()
	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Gateway created successfully.",
		"id":      id,
	})
}

func (server *Server) updateGateway(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	var payload gatewayPayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	var existingID int64
	err = server.db.QueryRow("SELECT id FROM gateways WHERE id = ? AND deleted_at IS NULL", id).Scan(&existingID)
	if err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	enc, err := server.getEncrypter()
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Encryption configuration error: " + err.Error()})
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []any{}

	if strings.TrimSpace(payload.Name) != "" {
		setClauses = append(setClauses, "name = ?")
		args = append(args, strings.TrimSpace(payload.Name))
	}
	if strings.TrimSpace(payload.Driver) != "" {
		setClauses = append(setClauses, "driver = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(payload.Driver)))
	}
	if payload.Priority > 0 {
		setClauses = append(setClauses, "priority = ?")
		args = append(args, payload.Priority)
	}
	if payload.IsEnabled != nil {
		setClauses = append(setClauses, "is_enabled = ?")
		args = append(args, *payload.IsEnabled)
	}
	if payload.IsOtpEnabled != nil {
		setClauses = append(setClauses, "is_otp_enabled = ?")
		args = append(args, *payload.IsOtpEnabled)
	}
	if payload.SupportsBulk != nil {
		setClauses = append(setClauses, "supports_bulk = ?")
		args = append(args, *payload.SupportsBulk)
	}
	if payload.RequestsPerSecond > 0 {
		setClauses = append(setClauses, "requests_per_second = ?")
		args = append(args, payload.RequestsPerSecond)
	}
	if payload.TimeoutSeconds > 0 {
		setClauses = append(setClauses, "timeout_seconds = ?")
		args = append(args, payload.TimeoutSeconds)
	}
	if payload.DlrSupport != nil {
		setClauses = append(setClauses, "dlr_support = ?")
		args = append(args, *payload.DlrSupport)
	}
	if payload.CostMode != "" {
		setClauses = append(setClauses, "cost_mode = ?")
		args = append(args, payload.CostMode)
	}
	if payload.DippingRequired != nil {
		setClauses = append(setClauses, "dipping_required = ?")
		args = append(args, *payload.DippingRequired)
	}
	if payload.HasThirdPartyDipping != nil {
		setClauses = append(setClauses, "has_third_party_dipping = ?")
		args = append(args, *payload.HasThirdPartyDipping)
	}

	if len(payload.Credentials) > 0 {
		encCreds, err := enc.EncryptArray(payload.Credentials)
		if err != nil {
			return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to encrypt credentials: " + err.Error()})
		}
		setClauses = append(setClauses, "credentials = ?")
		args = append(args, encCreds)
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE gateways SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	_, err = server.db.Exec(query, args...)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{"message": "Gateway updated successfully."})
}

func (server *Server) deleteGateway(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	res, err := server.db.Exec("UPDATE gateways SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return databaseError(context, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	return context.JSON(fiber.Map{"message": "Gateway deleted successfully."})
}

type GatewayBillingSummary struct {
	Revenue            float64  `json:"revenue"`
	OperatorCost       float64  `json:"operator_cost"`
	DippingCost        float64  `json:"dipping_cost"`
	TotalCost          float64  `json:"total_cost"`
	NetProfit          float64  `json:"net_profit"`
	ProfitMargin       float64  `json:"profit_margin"`
	DlrDeliveryPercent *float64 `json:"dlr_delivery_percent"`
	TotalMessages      int64    `json:"total_messages"`
}

type GatewayDetailResponse struct {
	ID                   int64                  `json:"id"`
	Name                 string                 `json:"name"`
	Driver               string                 `json:"driver"`
	RouteCode            *string                `json:"route_code"`
	Priority             int                    `json:"priority"`
	IsEnabled            bool                   `json:"is_enabled"`
	Enabled              bool                   `json:"enabled"`
	IsOtpEnabled         bool                   `json:"is_otp_enabled"`
	OtpEnabled           bool                   `json:"otp_enabled"`
	SupportsBulk         bool                   `json:"supports_bulk"`
	RequestsPerSecond    int                    `json:"requests_per_second"`
	TimeoutSeconds       int                    `json:"timeout_seconds"`
	DlrSupport           bool                   `json:"dlr_support"`
	CostMode             string                 `json:"cost_mode"`
	DippingRequired      bool                   `json:"dipping_required"`
	HasThirdPartyDipping bool                   `json:"has_third_party_dipping"`
	HealthStatus         string                 `json:"health_status"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
	Credentials          map[string]any         `json:"credentials,omitempty"`
	BillingSummary       GatewayBillingSummary  `json:"billing_summary"`
}

func (server *Server) gatewayDetail(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	var g GatewayDetailResponse
	var rawCreds sql.NullString
	var created, updated sql.NullString

	err = server.db.QueryRow(`
		SELECT id, name, driver, priority, is_enabled, is_otp_enabled, supports_bulk,
		       requests_per_second, timeout_seconds, dlr_support, cost_mode,
		       dipping_required, has_third_party_dipping, health_status, credentials,
		       DATE_FORMAT(created_at, '%b %d, %Y %H:%i:%s'),
		       DATE_FORMAT(updated_at, '%b %d, %Y %H:%i:%s')
		FROM gateways
		WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(
		&g.ID, &g.Name, &g.Driver, &g.Priority, &g.IsEnabled, &g.IsOtpEnabled, &g.SupportsBulk,
		&g.RequestsPerSecond, &g.TimeoutSeconds, &g.DlrSupport, &g.CostMode,
		&g.DippingRequired, &g.HasThirdPartyDipping, &g.HealthStatus, &rawCreds,
		&created, &updated,
	)
	if err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	g.Enabled = g.IsEnabled
	g.OtpEnabled = g.IsOtpEnabled
	if created.Valid {
		g.CreatedAt = created.String
	}
	if updated.Valid {
		g.UpdatedAt = updated.String
	}

	if rawCreds.Valid && rawCreds.String != "" {
		enc, err := server.getEncrypter()
		if err == nil {
			if credMap, err := enc.DecryptArray(rawCreds.String); err == nil {
				g.RouteCode = extractRouteCode(credMap)
				masked := make(map[string]any)
				for k, v := range credMap {
					strVal := fmt.Sprintf("%v", v)
					if len(strVal) > 6 {
						masked[k] = strVal[:3] + "..." + strVal[len(strVal)-2:]
					} else {
						masked[k] = "***"
					}
				}
				g.Credentials = masked
			}
		}
	}

	// Calculate Billing & Profit Summary for Current Month
	// (Matches Laravel's GatewayProfitCalculationService and CostResolutionService)
	var summary GatewayBillingSummary
	var delivered, dlrPending, dlrFailed int64
	var thirdPartyCost float64

	_ = server.db.QueryRow(`
		SELECT 
			COUNT(*) as total_messages,
			COALESCE(SUM(cost), 0) as revenue,
			COALESCE(SUM(operator_cost), 0) as operator_cost,
			COALESCE(SUM(primary_dipping_cost), 0) as dipping_cost,
			COALESCE(SUM(third_party_dipping_cost), 0) as third_party_cost,
			COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0) as delivered,
			COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0) as dlr_pending,
			COALESCE(SUM(CASE WHEN status = 'dlr_failed' THEN 1 ELSE 0 END), 0) as dlr_failed
		FROM messages
		WHERE gateway_id = ?
		  AND status != 'failed'
		  AND created_at >= DATE_FORMAT(NOW(), '%Y-%m-01 00:00:00')
	`, id).Scan(
		&summary.TotalMessages,
		&summary.Revenue,
		&summary.OperatorCost,
		&summary.DippingCost,
		&thirdPartyCost,
		&delivered,
		&dlrPending,
		&dlrFailed,
	)

	if summary.TotalMessages == 0 {
		_ = server.db.QueryRow(`
			SELECT 
				COUNT(*) as total_messages,
				COALESCE(SUM(cost), 0) as revenue,
				COALESCE(SUM(operator_cost), 0) as operator_cost,
				COALESCE(SUM(primary_dipping_cost), 0) as dipping_cost,
				COALESCE(SUM(third_party_dipping_cost), 0) as third_party_cost,
				COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0) as delivered,
				COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0) as dlr_pending,
				COALESCE(SUM(CASE WHEN status = 'dlr_failed' THEN 1 ELSE 0 END), 0) as dlr_failed
			FROM messages
			WHERE gateway_id = ?
			  AND status != 'failed'
			  AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		`, id).Scan(
			&summary.TotalMessages,
			&summary.Revenue,
			&summary.OperatorCost,
			&summary.DippingCost,
			&thirdPartyCost,
			&delivered,
			&dlrPending,
			&dlrFailed,
		)
	}

		summary.Revenue = math.Round(summary.Revenue*100) / 100
		summary.OperatorCost = math.Round(summary.OperatorCost*100) / 100
		summary.DippingCost = math.Round(summary.DippingCost*100) / 100
		summary.TotalCost = math.Round((summary.OperatorCost+summary.DippingCost+thirdPartyCost)*100) / 100
		summary.NetProfit = math.Round((summary.Revenue-summary.TotalCost)*100) / 100

		if summary.Revenue > 0 {
			summary.ProfitMargin = math.Round((summary.NetProfit/summary.Revenue)*1000) / 10
		} else {
			summary.ProfitMargin = 0.0
		}

		if g.DlrSupport {
			denom := delivered + dlrPending + dlrFailed
			if denom > 0 {
				pct := math.Round((float64(delivered)/float64(denom))*1000) / 10
				summary.DlrDeliveryPercent = &pct
			} else {
				zero := 0.0
				summary.DlrDeliveryPercent = &zero
			}
		}

	g.BillingSummary = summary

	return context.JSON(fiber.Map{"data": g})
}

func (server *Server) gatewayDashboard(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	var name, driver string
	var rawCreds sql.NullString
	err = server.db.QueryRow("SELECT name, driver, credentials FROM gateways WHERE id = ? AND deleted_at IS NULL", id).Scan(&name, &driver, &rawCreds)
	if err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	cards := make([]fiber.Map, 0)

	if strings.ToLower(driver) == "infozillion" && rawCreds.Valid && rawCreds.String != "" {
		enc, err := server.getEncrypter()
		if err == nil {
			if credMap, err := enc.DecryptArray(rawCreds.String); err == nil {
				apiKey, _ := credMap["api_key"].(string)
				baseURL, _ := credMap["base_url"].(string)
				username, _ := credMap["username"].(string)
				password, _ := credMap["password"].(string)

				baseURL = strings.TrimRight(baseURL, "/")

				client := &http.Client{
					Timeout: 3 * time.Second,
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				}

				// 1. Wallet Balance
				walletVal := "N/A"
				walletColor := "danger"
				if apiKey != "" && baseURL != "" {
					reqBody, _ := json.Marshal(map[string]string{"apiKey": apiKey})
					req, err := http.NewRequest("POST", baseURL+"/a2p-wallet/api/v1/check-current-balance", bytes.NewBuffer(reqBody))
					if err == nil {
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Accept", "application/json")
						if resp, err := client.Do(req); err == nil {
							var resData struct {
								ServerResponseCode    string `json:"serverResponseCode"`
								ServerResponseMessage string `json:"serverResponseMessage"`
								AvailableBalance      any    `json:"availableBalance"`
							}
							if json.NewDecoder(resp.Body).Decode(&resData) == nil && resData.AvailableBalance != nil {
								if num, err := strconv.ParseFloat(fmt.Sprintf("%v", resData.AvailableBalance), 64); err == nil {
									walletVal = fmt.Sprintf("%.2f...", num)
									walletColor = "success"
								} else {
									walletVal = fmt.Sprintf("%v", resData.AvailableBalance)
									walletColor = "success"
								}
							}
							resp.Body.Close()
						}
					}
				}

				cards = append(cards, fiber.Map{
					"label":       "Wallet Balance",
					"value":       walletVal,
					"description": "Available balance at CP end",
					"color":       walletColor,
					"icon":        "heroicon-o-banknotes",
				})

				// 2. Carrier Credit Balance(s)
				carrierCode := "RB"
				if rc := extractRouteCode(credMap); rc != nil && *rc != "" {
					firstCarrier := strings.Split(*rc, ",")[0]
					carrierCode = strings.TrimSpace(firstCarrier)
				}

				carrierName := "Robi"
				if carrierCode == "GP" {
					carrierName = "Grameenphone"
				} else if carrierCode == "BL" {
					carrierName = "Banglalink"
				} else if carrierCode == "TT" {
					carrierName = "Teletalk"
				} else if carrierCode == "AT" {
					carrierName = "Airtel"
				}

				carrierVal := "N/A"
				carrierColor := "danger"
				if apiKey != "" && baseURL != "" && username != "" && password != "" {
					reqBody, _ := json.Marshal(map[string]string{
						"username": username,
						"password": password,
						"mno":      carrierCode,
						"apiKey":   apiKey,
					})
					req, err := http.NewRequest("POST", baseURL+"/a2p-proxy-api/api/v1/check-credit-balance", bytes.NewBuffer(reqBody))
					if err == nil {
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Accept", "application/json")
						if resp, err := client.Do(req); err == nil {
							var resData struct {
								ServerResponseCode    string `json:"serverResponseCode"`
								ServerResponseMessage string `json:"serverResponseMessage"`
								AvailableBalance      any    `json:"availableBalance"`
							}
							if json.NewDecoder(resp.Body).Decode(&resData) == nil && resData.AvailableBalance != nil {
								if num, err := strconv.ParseFloat(fmt.Sprintf("%v", resData.AvailableBalance), 64); err == nil {
									carrierVal = fmt.Sprintf("%.1f...", num)
									carrierColor = "success"
								} else {
									carrierVal = fmt.Sprintf("%v", resData.AvailableBalance)
									carrierColor = "success"
								}
							}
							resp.Body.Close()
						}
					}
				}

				cards = append(cards, fiber.Map{
					"label":       carrierName + " Balance",
					"value":       carrierVal,
					"description": carrierCode + " carrier credit",
					"color":       carrierColor,
					"icon":        "heroicon-o-device-phone-mobile",
				})
			}
		}
	}

	if len(cards) == 0 {
		cards = append(cards, fiber.Map{
			"label":       "Wallet Balance",
			"value":       "—",
			"description": "Available balance at CP end",
			"color":       "success",
			"icon":        "heroicon-o-banknotes",
		})
		cards = append(cards, fiber.Map{
			"label":       "Carrier Balance",
			"value":       "—",
			"description": "Carrier credit status",
			"color":       "success",
			"icon":        "heroicon-o-device-phone-mobile",
		})
	}

	return context.JSON(fiber.Map{
		"cards": cards,
	})
}

type OperatorRateItem struct {
	ID              int64    `json:"id"`
	OperatorID      int64    `json:"operator_id"`
	OperatorName    string   `json:"operator_name"`
	RouteCode       *string  `json:"route_code"`
	CountryName     string   `json:"country_name"`
	MaskingPrice    *float64 `json:"masking_price"`
	NonMaskingPrice *float64 `json:"non_masking_price"`
	DippingCharge   *float64 `json:"dipping_charge"`
	OperatorCost    float64  `json:"operator_cost"`
	DippingCost     float64  `json:"dipping_cost"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

func (server *Server) gatewayOperatorRates(context fiber.Ctx) error {
	gatewayID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || gatewayID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	// First verify gateway exists and get route_code from credentials
	var rawCreds sql.NullString
	err = server.db.QueryRow("SELECT credentials FROM gateways WHERE id = ? AND deleted_at IS NULL", gatewayID).Scan(&rawCreds)
	if err != nil {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	var gatewayRouteCode *string
	if rawCreds.Valid && rawCreds.String != "" {
		if enc, err := server.getEncrypter(); err == nil {
			if credMap, err := enc.DecryptArray(rawCreds.String); err == nil {
				gatewayRouteCode = extractRouteCode(credMap)
			}
		}
	}

	search := strings.TrimSpace(context.Query("search"))
	page := 1
	if p, err := strconv.Atoi(context.Query("page")); err == nil && p > 0 {
		page = p
	}
	perPage := 10
	if pp, err := strconv.Atoi(context.Query("per_page")); err == nil && pp > 0 && pp <= 100 {
		perPage = pp
	}

	whereClause := "WHERE gor.gateway_id = ?"
	args := []any{gatewayID}

	if search != "" {
		whereClause += " AND (o.name LIKE ? OR c.name LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM gateway_operator_rates gor
		JOIN operators o ON o.id = gor.operator_id
		LEFT JOIN countries c ON c.id = o.country_id
		%s
	`, whereClause)
	if err := server.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return databaseError(context, err)
	}

	query := fmt.Sprintf(`
		SELECT gor.id, gor.operator_id, o.name, COALESCE(c.name, 'Bangladesh'),
		       gor.masking_price, gor.non_masking_price, gor.dipping_charge,
		       COALESCE(o.codes, ''), COALESCE(c.code, 880),
		       DATE_FORMAT(gor.created_at, '%%b %%d, %%Y %%H:%%i:%%s'),
		       DATE_FORMAT(gor.updated_at, '%%b %%d, %%Y %%H:%%i:%%s')
		FROM gateway_operator_rates gor
		JOIN operators o ON o.id = gor.operator_id
		LEFT JOIN countries c ON c.id = o.country_id
		%s
		ORDER BY gor.id ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	limitArgs := append(args, perPage, (page-1)*perPage)
	rows, err := server.db.Query(query, limitArgs...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]OperatorRateItem, 0)
	for rows.Next() {
		var item OperatorRateItem
		var rawCodes string
		var countryCode int
		var created, updated sql.NullString

		if err := rows.Scan(
			&item.ID, &item.OperatorID, &item.OperatorName, &item.CountryName,
			&item.MaskingPrice, &item.NonMaskingPrice, &item.DippingCharge,
			&rawCodes, &countryCode, &created, &updated,
		); err != nil {
			return databaseError(context, err)
		}

		item.RouteCode = gatewayRouteCode
		if created.Valid {
			item.CreatedAt = created.String
		}
		if updated.Valid {
			item.UpdatedAt = updated.String
		}

		// Calculate operator_spent and dipping_spent for this operator & gateway from messages
		// Matches App\Models\GatewayOperatorRate::getOperatorSpentAttribute and getGatewaySpentAttribute
		var codes []string
		if rawCodes != "" {
			var parsedCodes []any
			if json.Unmarshal([]byte(rawCodes), &parsedCodes) == nil {
				for _, c := range parsedCodes {
					codes = append(codes, fmt.Sprintf("%v", c))
				}
			}
		}

		if len(codes) > 0 {
			var prefixConditions []string
			var prefixArgs []any
			prefixArgs = append(prefixArgs, gatewayID)

			for _, code := range codes {
				prefixConditions = append(prefixConditions, "m.to LIKE ?")
				prefixArgs = append(prefixArgs, fmt.Sprintf("+%d%s%%", countryCode, code))
			}

			spendQuery := fmt.Sprintf(`
				SELECT COALESCE(SUM(m.operator_cost), 0), COALESCE(SUM(m.primary_dipping_cost), 0)
				FROM messages m
				WHERE m.gateway_id = ? AND m.status != 'failed' AND (%s)
			`, strings.Join(prefixConditions, " OR "))

			_ = server.db.QueryRow(spendQuery, prefixArgs...).Scan(&item.OperatorCost, &item.DippingCost)
			item.OperatorCost = math.Round(item.OperatorCost*10000) / 10000
			item.DippingCost = math.Round(item.DippingCost*10000) / 10000
		}

		items = append(items, item)
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	return context.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{
			"current_page": page,
			"per_page":     perPage,
			"total":        total,
			"last_page":    lastPage,
		},
	})
}

type OperatorRatePayload struct {
	OperatorID      int64    `json:"operator_id"`
	MaskingPrice    *float64 `json:"masking_price"`
	NonMaskingPrice *float64 `json:"non_masking_price"`
	DippingCharge   *float64 `json:"dipping_charge"`
}

func (server *Server) saveOperatorRate(context fiber.Ctx) error {
	gatewayID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || gatewayID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	var payload OperatorRatePayload
	if err := json.Unmarshal(context.Body(), &payload); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid request payload."})
	}

	rateIDParam := context.Params("rateId")
	if rateIDParam != "" {
		rateID, err := strconv.ParseInt(rateIDParam, 10, 64)
		if err != nil || rateID < 1 {
			return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Rate not found."})
		}

		_, err = server.db.Exec(`
			UPDATE gateway_operator_rates
			SET masking_price = ?, non_masking_price = ?, dipping_charge = ?, updated_at = NOW()
			WHERE id = ? AND gateway_id = ?
		`, payload.MaskingPrice, payload.NonMaskingPrice, payload.DippingCharge, rateID, gatewayID)
		if err != nil {
			return databaseError(context, err)
		}

		return context.JSON(fiber.Map{"message": "Operator rate updated successfully."})
	}

	// Create new rate
	if payload.OperatorID < 1 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Operator is required."})
	}

	res, err := server.db.Exec(`
		INSERT INTO gateway_operator_rates (
			gateway_id, operator_id, masking_price, non_masking_price, dipping_charge, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			masking_price = VALUES(masking_price),
			non_masking_price = VALUES(non_masking_price),
			dipping_charge = VALUES(dipping_charge),
			updated_at = NOW()
	`, gatewayID, payload.OperatorID, payload.MaskingPrice, payload.NonMaskingPrice, payload.DippingCharge)
	if err != nil {
		return databaseError(context, err)
	}

	newID, _ := res.LastInsertId()
	return context.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Operator rate saved successfully.",
		"id":      newID,
	})
}

func (server *Server) deleteOperatorRate(context fiber.Ctx) error {
	gatewayID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || gatewayID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	rateID, err := strconv.ParseInt(context.Params("rateId"), 10, 64)
	if err != nil || rateID < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Rate not found."})
	}

	res, err := server.db.Exec("DELETE FROM gateway_operator_rates WHERE id = ? AND gateway_id = ?", rateID, gatewayID)
	if err != nil {
		return databaseError(context, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Rate not found."})
	}

	return context.JSON(fiber.Map{"message": "Operator rate deleted successfully."})
}

func (server *Server) gatewaySenders(context fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Gateway not found."})
	}

	rows, err := server.db.Query(`
		SELECT id, sender_id, type, status, is_default, created_at
		FROM senders
		WHERE gateway_id = ? AND deleted_at IS NULL
		ORDER BY sender_id ASC
	`, id)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	type senderItem struct {
		ID        int64  `json:"id"`
		SenderID  string `json:"sender_id"`
		Type      string `json:"type"`
		Status    string `json:"status"`
		IsDefault bool   `json:"is_default"`
		CreatedAt string `json:"created_at"`
	}

	items := make([]senderItem, 0)
	for rows.Next() {
		var s senderItem
		if err := rows.Scan(&s.ID, &s.SenderID, &s.Type, &s.Status, &s.IsDefault, &s.CreatedAt); err == nil {
			items = append(items, s)
		}
	}

	return context.JSON(fiber.Map{"data": items})
}
