package httpapi

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
)

type dashboardKpis struct {
	TotalQueue   int     `json:"total_queue"`
	PendingQueue int     `json:"pending_queue"`
	Processing   int     `json:"processing"`
	RecentSent   int     `json:"recent_sent"`
	RecentFailed int     `json:"recent_failed"`
	DLRSuccess   float64 `json:"dlr_success"`
	CurrentTPS   float64 `json:"current_tps"`
	DBLatencyMS  float64 `json:"db_latency_ms"`
	QuickQueue   int     `json:"quick_queue"`
	BulkQueue    int     `json:"bulk_queue"`
}

type dashboardGateway struct {
	Name        string  `json:"name"`
	Driver      string  `json:"driver"`
	Status      string  `json:"status"`
	TPS         float64 `json:"tps"`
	SuccessRate float64 `json:"success_rate"`
}

type dashboardMessage struct {
	ID           int64   `json:"id"`
	CustomerName string  `json:"customer_name"`
	Source       string  `json:"source"`
	Recipient    string  `json:"recipient"`
	Body         string  `json:"body"`
	Cost         float64 `json:"cost"`
	Status       string  `json:"status"`
	UpdatedAt    string  `json:"updated_at"`
}

func (server *Server) dashboard(context fiber.Ctx) error {
	startedAt := time.Now()

	var pending, processing, quickQueue, bulkQueue int
	err := server.db.QueryRow(`
		SELECT
			COALESCE(SUM(status = 'pending'), 0),
			COALESCE(SUM(status = 'sending'), 0),
			COALESCE(SUM(status IN ('pending', 'sending') AND batch_id IS NULL), 0),
			COALESCE(SUM(status IN ('pending', 'sending') AND batch_id IS NOT NULL), 0)
		FROM messages
	`).Scan(&pending, &processing, &quickQueue, &bulkQueue)
	if err != nil {
		return databaseError(context, err)
	}

	var recentSent, recentFailed, sentLastMinute int
	err = server.db.QueryRow(`
		SELECT
			COALESCE(SUM(status IN ('sent', 'delivered')), 0),
			COALESCE(SUM(status IN ('failed', 'dlr_failed')), 0),
			COALESCE(SUM(updated_at >= DATE_SUB(NOW(), INTERVAL 1 MINUTE) AND status IN ('sent', 'delivered')), 0)
		FROM messages
		WHERE updated_at >= DATE_SUB(NOW(), INTERVAL 15 MINUTE)
	`).Scan(&recentSent, &recentFailed, &sentLastMinute)
	if err != nil {
		return databaseError(context, err)
	}

	deliveryTotal := recentSent + recentFailed
	deliveryRate := 100.0
	if deliveryTotal > 0 {
		deliveryRate = float64(recentSent) / float64(deliveryTotal) * 100
	}

	gateways, err := server.dashboardGateways()
	if err != nil {
		return databaseError(context, err)
	}

	messages, err := server.dashboardMessages()
	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"kpis": dashboardKpis{
				TotalQueue:   pending + processing,
				PendingQueue: pending,
				Processing:   processing,
				RecentSent:   recentSent,
				RecentFailed: recentFailed,
				DLRSuccess:   deliveryRate,
				CurrentTPS:   float64(sentLastMinute) / 60,
				DBLatencyMS:  float64(time.Since(startedAt).Microseconds()) / 1000,
				QuickQueue:   quickQueue,
				BulkQueue:    bulkQueue,
			},
			"gateways":      gateways,
			"live_messages": messages,
			"system_time":   time.Now().Format(time.RFC3339),
		},
	})
}

func (server *Server) dashboardGateways() ([]dashboardGateway, error) {
	rows, err := server.db.Query(`
		SELECT
			g.name,
			COALESCE(g.driver, 'default'),
			g.is_enabled,
			COALESCE(SUM(m.updated_at >= DATE_SUB(NOW(), INTERVAL 15 MINUTE) AND m.status IN ('sent', 'delivered')), 0),
			COALESCE(SUM(m.updated_at >= DATE_SUB(NOW(), INTERVAL 15 MINUTE) AND m.status IN ('failed', 'dlr_failed')), 0),
			COALESCE(SUM(m.updated_at >= DATE_SUB(NOW(), INTERVAL 1 MINUTE) AND m.status IN ('sent', 'delivered')), 0)
		FROM gateways g
		LEFT JOIN messages m ON m.gateway_id = g.id
		GROUP BY g.id, g.name, g.driver, g.is_enabled
		ORDER BY g.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gateways := make([]dashboardGateway, 0)
	for rows.Next() {
		var gateway dashboardGateway
		var enabled bool
		var sent, failed, sentLastMinute int
		if err := rows.Scan(&gateway.Name, &gateway.Driver, &enabled, &sent, &failed, &sentLastMinute); err != nil {
			return nil, err
		}

		gateway.Status = "Online"
		if !enabled {
			gateway.Status = "Offline"
		} else if failed > 10 {
			gateway.Status = "Degraded"
		}
		gateway.TPS = float64(sentLastMinute) / 60
		gateway.SuccessRate = 100
		if sent+failed > 0 {
			gateway.SuccessRate = float64(sent) / float64(sent+failed) * 100
		}
		gateways = append(gateways, gateway)
	}

	return gateways, rows.Err()
}

func (server *Server) dashboardMessages() ([]dashboardMessage, error) {
	rows, err := server.db.Query(`
		SELECT m.id, u.name, m.source, m.` + "`to`" + `, m.body, m.cost, m.status, m.updated_at
		FROM messages m
		INNER JOIN users u ON u.id = m.user_id
		WHERE u.deleted_at IS NULL
		ORDER BY m.updated_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]dashboardMessage, 0)
	for rows.Next() {
		var message dashboardMessage
		var updatedAt time.Time
		if err := rows.Scan(
			&message.ID,
			&message.CustomerName,
			&message.Source,
			&message.Recipient,
			&message.Body,
			&message.Cost,
			&message.Status,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		message.UpdatedAt = updatedAt.Format(time.RFC3339)
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func databaseError(context fiber.Ctx, err error) error {
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Record not found."})
	}

	return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to load Superman data."})
}
