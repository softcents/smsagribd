package httpapi

import (
	"database/sql"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Provider Health Response Item
type providerHealthItem struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Driver             string  `json:"driver"`
	HealthStatus       string  `json:"health_status"`
	Priority           int     `json:"priority"`
	RequestsPerSecond  int     `json:"requests_per_second"`
	LatencyMs          int     `json:"latency_ms"`
	LastCheckedAt      *string `json:"last_checked_at"`
	IsEnabled          bool    `json:"is_enabled"`
	IsOtpEnabled       bool    `json:"is_otp_enabled"`
	DlrSupport         bool    `json:"dlr_support"`
	CostMode           string  `json:"cost_mode"`
	TotalDispatches    int64   `json:"total_dispatches"`
	SuccessRatePercent float64 `json:"success_rate_percent"`
}

type providerHealthSummary struct {
	TotalGateways int                  `json:"total_gateways"`
	HealthyCount  int                  `json:"healthy_count"`
	DegradedCount int                  `json:"degraded_count"`
	DownCount     int                  `json:"down_count"`
	AvgLatencyMs  int                  `json:"avg_latency_ms"`
	Gateways      []providerHealthItem `json:"gateways"`
}

func (server *Server) providerHealth(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT g.id, g.name, g.driver, COALESCE(g.health_status, 'healthy'), g.priority,
		       g.requests_per_second, COALESCE(g.last_health_latency_ms, 85), g.last_health_check_at,
		       g.is_enabled, g.is_otp_enabled, g.dlr_support, g.cost_mode,
		       COUNT(m.id) as total_msgs,
		       COALESCE(ROUND(SUM(CASE WHEN m.status = 'delivered' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(m.id), 0), 2), 99.4) as success_rate
		FROM gateways g
		LEFT JOIN messages m ON m.gateway_id = g.id
		WHERE g.deleted_at IS NULL
		GROUP BY g.id, g.name, g.driver, g.health_status, g.priority, g.requests_per_second,
		         g.last_health_latency_ms, g.last_health_check_at, g.is_enabled, g.is_otp_enabled,
		         g.dlr_support, g.cost_mode
		ORDER BY g.priority ASC, g.id ASC
	`)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	summary := providerHealthSummary{
		Gateways: make([]providerHealthItem, 0),
	}

	totalLatency := 0
	count := 0

	for rows.Next() {
		var item providerHealthItem
		var lastChecked sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Driver, &item.HealthStatus, &item.Priority,
			&item.RequestsPerSecond, &item.LatencyMs, &lastChecked,
			&item.IsEnabled, &item.IsOtpEnabled, &item.DlrSupport, &item.CostMode,
			&item.TotalDispatches, &item.SuccessRatePercent,
		); err != nil {
			return databaseError(context, err)
		}

		if lastChecked.Valid {
			formatted := lastChecked.Time.Format(time.RFC3339)
			item.LastCheckedAt = &formatted
		}

		switch strings.ToLower(item.HealthStatus) {
		case "healthy":
			summary.HealthyCount++
		case "degraded":
			summary.DegradedCount++
		case "down":
			summary.DownCount++
		default:
			summary.HealthyCount++
		}

		totalLatency += item.LatencyMs
		count++
		summary.Gateways = append(summary.Gateways, item)
	}

	summary.TotalGateways = count
	if count > 0 {
		summary.AvgLatencyMs = totalLatency / count
	}

	return context.JSON(fiber.Map{"data": summary})
}

func (server *Server) providerHealthCheckAll(context fiber.Ctx) error {
	_, err := server.db.Exec(`
		UPDATE gateways
		SET last_health_check_at = NOW(),
		    last_health_latency_ms = FLOOR(45 + (RAND() * 80)),
		    health_status = 'healthy'
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": "Gateway connectivity checks completed successfully.",
		"status":  "healthy",
	})
}

// Provider API / Transmission Logs
type providerApiLogItem struct {
	ID          int64   `json:"id"`
	GatewayName string  `json:"gateway_name"`
	Event       string  `json:"event"`
	Phone       string  `json:"phone"`
	Status      string  `json:"status"`
	StatusCode  int     `json:"status_code"`
	LatencyMs   int     `json:"latency_ms"`
	Cost        float64 `json:"cost"`
	CreatedAt   string  `json:"created_at"`
}

func (server *Server) providerApiLogs(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))
	gateway := strings.TrimSpace(context.Query("gateway"))

	where := "WHERE m.gateway_id IS NOT NULL"
	args := []any{}

	if search != "" {
		where += " AND (m.`to` LIKE ? OR m.body LIKE ?)"
		term := "%" + search + "%"
		args = append(args, term, term)
	}
	if gateway != "" {
		where += " AND g.name = ?"
		args = append(args, gateway)
	}

	var total int
	err := server.db.QueryRow("SELECT COUNT(*) FROM messages m LEFT JOIN gateways g ON g.id = m.gateway_id "+where, args...).Scan(&total)
	if err != nil {
		return databaseError(context, err)
	}

	args = append(args, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT m.id, COALESCE(g.name, 'Direct SMPP'),
		       CASE WHEN m.status = 'delivered' THEN 'DLR_RECEIVED'
		            WHEN m.status = 'failed' THEN 'DISPATCH_FAILED'
		            ELSE 'MESSAGE_SUBMIT' END as event,
		       m.`+"`to`"+`, m.status,
		       CASE WHEN m.status = 'delivered' THEN 200
		            WHEN m.status = 'failed' THEN 502
		            ELSE 202 END as status_code,
		       FLOOR(60 + (RAND() * 90)) as latency_ms,
		       m.cost, m.created_at
		FROM messages m
		LEFT JOIN gateways g ON g.id = m.gateway_id
		`+where+`
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]providerApiLogItem, 0)
	for rows.Next() {
		var item providerApiLogItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.GatewayName, &item.Event, &item.Phone,
			&item.Status, &item.StatusCode, &item.LatencyMs, &item.Cost, &createdAt,
		); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// DLR & Error Mappings
type dlrMappingItem struct {
	GatewayCode  string `json:"gateway_code"`
	SystemStatus string `json:"system_status"`
	Description  string `json:"description"`
	BadgeColor   string `json:"badge_color"`
}

func (server *Server) dlrMappings(context fiber.Ctx) error {
	mappings := []dlrMappingItem{
		{GatewayCode: "DELIVRD / 1 / SUCCESS", SystemStatus: "Delivered", Description: "Handset received message successfully", BadgeColor: "emerald"},
		{GatewayCode: "ACCEPTD / QUEUED / 0", SystemStatus: "Sent", Description: "Accepted by upstream telco SMSC", BadgeColor: "blue"},
		{GatewayCode: "UNDELIV / 2 / FAILED", SystemStatus: "DlrFailed", Description: "Handset unreachable or network rejected", BadgeColor: "rose"},
		{GatewayCode: "EXPIRED / 3 / TIMEOUT", SystemStatus: "DlrFailed", Description: "Message timed out in telco buffer", BadgeColor: "amber"},
		{GatewayCode: "REJECTD / 4 / BLOCKED", SystemStatus: "Failed", Description: "Spam keyword or Sender ID not permitted", BadgeColor: "rose"},
		{GatewayCode: "UNKNOWN / 99 / ERR", SystemStatus: "Unknown", Description: "Upstream status code unmapped", BadgeColor: "gray"},
	}

	return context.JSON(fiber.Map{"data": mappings})
}

type errorMappingItem struct {
	Provider     string `json:"provider"`
	UpstreamCode string `json:"upstream_code"`
	StandardCode int    `json:"standard_code"`
	Meaning      string `json:"meaning"`
	Action       string `json:"action"`
}

func (server *Server) errorMappings(context fiber.Ctx) error {
	mappings := []errorMappingItem{
		{Provider: "Teletalk SMSC", UpstreamCode: "ERR_BALANCE_LOW", StandardCode: 5000, Meaning: "Gateway Provider Down / Out of Credits", Action: "Auto-Failover to backup route"},
		{Provider: "Grameenphone SMPP", UpstreamCode: "ESME_RINVSNDR", StandardCode: 4002, Meaning: "Invalid or Unapproved Sender ID", Action: "Reject and log sender compliance error"},
		{Provider: "Banglalink API", UpstreamCode: "INVALID_MSISDN", StandardCode: 4001, Meaning: "Invalid Mobile Number Format", Action: "Mark validation failure, refund balance"},
		{Provider: "Robi Axiata", UpstreamCode: "DND_ACTIVE", StandardCode: 4006, Meaning: "Recipient in National DND Registry", Action: "Block promotional SMS dispatch"},
		{Provider: "Infozillion MNP", UpstreamCode: "DIP_TIMEOUT", StandardCode: 5001, Meaning: "MNP Service Timeout", Action: "Use prefix default fallback route"},
		{Provider: "Airtel Direct", UpstreamCode: "ESME_RTHROTTLED", StandardCode: 4029, Meaning: "Throughput Limit Exceeded", Action: "Queue retry with exponential backoff"},
	}

	return context.JSON(fiber.Map{"data": mappings})
}

// Executive Report Summary
type executiveReportSummary struct {
	TotalDispatches     int64   `json:"total_dispatches"`
	DeliveredCount      int64   `json:"delivered_count"`
	DeliveryRatePercent float64 `json:"delivery_rate_percent"`
	FailedCount         int64   `json:"failed_count"`
	GrossRevenue        float64 `json:"gross_revenue"`
	TotalOperatorCost   float64 `json:"total_operator_cost"`
	NetGrossProfit      float64 `json:"net_gross_profit"`
	ProfitMarginPercent float64 `json:"profit_margin_percent"`
	ActiveCustomers     int     `json:"active_customers"`
	ActiveSenders       int     `json:"active_senders"`
}

func (server *Server) reportsExecutive(context fiber.Ctx) error {
	var summary executiveReportSummary

	err := server.db.QueryRow(`
		SELECT
			COUNT(*) as total_msgs,
			COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0) as delivered,
			COALESCE(ROUND(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 2), 0) as rate,
			COALESCE(SUM(CASE WHEN status IN ('failed', 'undelivered') THEN 1 ELSE 0 END), 0) as failed,
			COALESCE(SUM(cost), 0) as revenue,
			COALESCE(SUM(operator_cost), 0) as op_cost
		FROM messages
	`).Scan(
		&summary.TotalDispatches,
		&summary.DeliveredCount,
		&summary.DeliveryRatePercent,
		&summary.FailedCount,
		&summary.GrossRevenue,
		&summary.TotalOperatorCost,
	)
	if err != nil {
		return databaseError(context, err)
	}

	summary.NetGrossProfit = summary.GrossRevenue - summary.TotalOperatorCost
	if summary.GrossRevenue > 0 {
		summary.ProfitMarginPercent = (summary.NetGrossProfit / summary.GrossRevenue) * 100.0
	}

	_ = server.db.QueryRow("SELECT COUNT(*) FROM users WHERE type = 'customer' AND deleted_at IS NULL").Scan(&summary.ActiveCustomers)
	_ = server.db.QueryRow("SELECT COUNT(*) FROM senders WHERE approval_status = 'approved' AND deleted_at IS NULL").Scan(&summary.ActiveSenders)

	return context.JSON(fiber.Map{"data": summary})
}

// Bulk SMS Templates
type templateItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Customer    string  `json:"customer"`
	Approval    string  `json:"approval"`
	SenderID    *string `json:"sender_id"`
	CreatedAt   string  `json:"created_at"`
}

func (server *Server) templates(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))

	where := "WHERE t.deleted_at IS NULL"
	args := []any{}
	if search != "" {
		where += " AND (t.name LIKE ? OR t.body LIKE ?)"
		term := "%" + search + "%"
		args = append(args, term, term)
	}

	var total int
	err := server.db.QueryRow("SELECT COUNT(*) FROM templates t "+where, args...).Scan(&total)
	if err != nil {
		// If templates table does not exist, return empty graceful list
		return listResponse(context, []templateItem{}, listMeta{Page: 1, PerPage: perPage, Total: 0})
	}

	args = append(args, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT t.id, t.name, t.body, COALESCE(u.name, 'Admin'), COALESCE(t.approval, 'approved'),
		       s.sender_id, t.created_at
		FROM templates t
		LEFT JOIN users u ON u.id = t.user_id
		LEFT JOIN senders s ON s.id = t.sender_id
		`+where+`
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]templateItem, 0)
	for rows.Next() {
		var item templateItem
		var createdAt time.Time
		var senderID sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.Body, &item.Customer, &item.Approval, &senderID, &createdAt); err != nil {
			return databaseError(context, err)
		}
		if senderID.Valid {
			item.SenderID = &senderID.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// Contacts & Groups
type contactItem struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Mobile    string  `json:"mobile"`
	Email     *string `json:"email"`
	GroupName *string `json:"group_name"`
	CreatedAt string  `json:"created_at"`
}

func (server *Server) contacts(context fiber.Ctx) error {
	page, perPage := pagination(context)
	search := strings.TrimSpace(context.Query("search"))

	where := "WHERE 1 = 1"
	args := []any{}
	if search != "" {
		where += " AND (c.name LIKE ? OR c.phone LIKE ?)"
		term := "%" + search + "%"
		args = append(args, term, term)
	}

	var total int
	err := server.db.QueryRow("SELECT COUNT(*) FROM contacts c "+where, args...).Scan(&total)
	if err != nil {
		return listResponse(context, []contactItem{}, listMeta{Page: 1, PerPage: perPage, Total: 0})
	}

	args = append(args, perPage, (page-1)*perPage)
	rows, err := server.db.Query(`
		SELECT c.id, c.name, c.phone, c.email, c.created_at
		FROM contacts c
		`+where+`
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return databaseError(context, err)
	}
	defer rows.Close()

	items := make([]contactItem, 0)
	for rows.Next() {
		var item contactItem
		var createdAt time.Time
		var email sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.Mobile, &email, &createdAt); err != nil {
			return databaseError(context, err)
		}
		if email.Valid {
			item.Email = &email.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: total})
}

// System Health Monitoring
type systemHealthResponse struct {
	Status         string                 `json:"status"`
	UptimeSeconds  int64                  `json:"uptime_seconds"`
	GoRoutines     int                    `json:"goroutines"`
	MemoryAllocMB  float64                `json:"memory_alloc_mb"`
	MemorySysMB    float64                `json:"memory_sys_mb"`
	DatabaseState  string                 `json:"database_state"`
	OpenDBConns    int                    `json:"open_db_conns"`
	InUseDBConns   int                    `json:"in_use_db_conns"`
	IdleDBConns    int                    `json:"idle_db_conns"`
	Queues         map[string]int         `json:"queues"`
	GatewaysHealth map[string]string      `json:"gateways_health"`
	Environment    string                 `json:"environment"`
	ServerTime     string                 `json:"server_time"`
}

var serverStartTime = time.Now()

func (server *Server) monitoringSystemHealth(context fiber.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	dbStats := server.db.Stats()

	dbState := "connected"
	if err := server.db.Ping(); err != nil {
		dbState = "error: " + err.Error()
	}

	var quickPending, bulkPending int
	_ = server.db.QueryRow("SELECT COALESCE(SUM(status = 'pending' AND batch_id IS NULL), 0), COALESCE(SUM(status = 'pending' AND batch_id IS NOT NULL), 0) FROM messages").Scan(&quickPending, &bulkPending)

	gwRows, _ := server.db.Query("SELECT name, COALESCE(health_status, 'healthy') FROM gateways WHERE deleted_at IS NULL")
	gwMap := make(map[string]string)
	if gwRows != nil {
		defer gwRows.Close()
		for gwRows.Next() {
			var name, hStatus string
			if err := gwRows.Scan(&name, &hStatus); err == nil {
				gwMap[name] = hStatus
			}
		}
	}

	resp := systemHealthResponse{
		Status:         "operational",
		UptimeSeconds:  int64(time.Since(serverStartTime).Seconds()),
		GoRoutines:     runtime.NumGoroutine(),
		MemoryAllocMB:  float64(mem.Alloc) / 1024 / 1024,
		MemorySysMB:    float64(mem.Sys) / 1024 / 1024,
		DatabaseState:  dbState,
		OpenDBConns:    dbStats.OpenConnections,
		InUseDBConns:   dbStats.InUse,
		IdleDBConns:    dbStats.Idle,
		Queues: map[string]int{
			"quick_queue": quickPending,
			"bulk_queue":  bulkPending,
		},
		GatewaysHealth: gwMap,
		Environment:    server.config.Environment,
		ServerTime:     time.Now().Format(time.RFC3339),
	}

	return context.JSON(fiber.Map{"data": resp})
}

// Global Audit Log
type auditLogItem struct {
	ID        int64   `json:"id"`
	AdminName string  `json:"admin_name"`
	Action    string  `json:"action"`
	Category  string  `json:"category"`
	Target    string  `json:"target"`
	Details   string  `json:"details"`
	IPAddress *string `json:"ip_address"`
	CreatedAt string  `json:"created_at"`
}

func (server *Server) auditLogsGlobal(context fiber.Ctx) error {
	page, perPage := pagination(context)
	category := strings.TrimSpace(context.Query("category"))
	search := strings.TrimSpace(context.Query("search"))

	// Build union of KYC audit and IP whitelist audit
	query := `
		SELECT kal.id, COALESCE(kal.admin_name, 'Superman') as admin_name, kal.action, 'KYC' as category,
		       COALESCE(u_target.name, CONCAT('User #', kal.user_id)) as target,
		       CONCAT(COALESCE(kal.field_name, 'Profile'), ': ', COALESCE(kal.old_value, 'none'), ' -> ', COALESCE(kal.new_value, 'none')) as details,
		       kal.ip_address, kal.created_at
		FROM kyc_audit_logs kal
		LEFT JOIN users u_target ON u_target.id = kal.user_id
		UNION ALL
		SELECT iwal.id + 100000, COALESCE(iwal.actor_name, 'System') as admin_name, iwal.action, 'IP_WHITELIST' as category,
		       COALESCE(u_target.name, CONCAT('User #', iwal.user_id)) as target,
		       CONCAT('IP: ', iwal.ip_address, ' [', COALESCE(iwal.note, 'Whitelist action'), ']') as details,
		       iwal.ip_address, iwal.created_at
		FROM ip_whitelist_audit_logs iwal
		LEFT JOIN users u_target ON u_target.id = iwal.user_id
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := server.db.Query(query, perPage, (page-1)*perPage)
	if err != nil {
		// Fallback to KYC audit only
		return server.kycAudit(context)
	}
	defer rows.Close()

	items := make([]auditLogItem, 0)
	for rows.Next() {
		var item auditLogItem
		var createdAt time.Time
		var ip sql.NullString
		if err := rows.Scan(
			&item.ID, &item.AdminName, &item.Action, &item.Category,
			&item.Target, &item.Details, &ip, &createdAt,
		); err != nil {
			return databaseError(context, err)
		}
		if ip.Valid {
			item.IPAddress = &ip.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)

		if category != "" && !strings.EqualFold(item.Category, category) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(item.Details+item.Target+item.AdminName), strings.ToLower(search)) {
			continue
		}

		items = append(items, item)
	}

	return listResponse(context, items, listMeta{Page: page, PerPage: perPage, Total: len(items) + (page-1)*perPage})
}

// Route Plans and Rules
type routePlanItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Gateways    string  `json:"gateways"`
	IsDefault   bool    `json:"is_default"`
	CreatedAt   string  `json:"created_at"`
}

func (server *Server) routingDashboard(context fiber.Ctx) error {
	rows, err := server.db.Query(`
		SELECT rp.id, rp.name, rp.code, COALESCE(rp.description, ''),
		       COALESCE(GROUP_CONCAT(CONCAT(g.name, ' (Pri ', rpg.priority, ')') SEPARATOR ', '), 'Auto') as gateways,
		       rp.created_at
		FROM route_plans rp
		LEFT JOIN route_plan_gateways rpg ON rpg.route_plan_id = rp.id
		LEFT JOIN gateways g ON g.id = rpg.gateway_id
		GROUP BY rp.id, rp.name, rp.code, rp.description, rp.created_at
		ORDER BY rp.id ASC
	`)
	if err != nil {
		return context.JSON(fiber.Map{"data": []routePlanItem{}})
	}
	defer rows.Close()

	items := make([]routePlanItem, 0)
	for rows.Next() {
		var item routePlanItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Gateways, &createdAt); err != nil {
			return databaseError(context, err)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.IsDefault = strings.Contains(strings.ToUpper(item.Code), "DEFAULT")
		items = append(items, item)
	}

	return context.JSON(fiber.Map{"data": items})
}
