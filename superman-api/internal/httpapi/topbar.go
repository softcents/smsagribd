package httpapi

import (
	"github.com/gofiber/fiber/v3"
)

type topbarStat struct {
	Title string `json:"title"`
	Value string `json:"value"`
	URL   string `json:"url"`
	Theme string `json:"theme"`
}

func (server *Server) topbarStats(context fiber.Ctx) error {
	var kycPending, smsSent, lowBalance, openTickets, ipWhitelist int
	var deposits, profit float64

	err := server.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM users WHERE type = 'customer' AND kyc_status = 'under_review' AND deleted_at IS NULL),
			(SELECT COALESCE(SUM(amount), 0) FROM deposits WHERE DATE(created_at) = CURDATE() AND status = 'completed' AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM messages WHERE DATE(created_at) = CURDATE() AND status IN ('sent', 'delivered')),
			(SELECT COALESCE(SUM(cost - (operator_cost + primary_dipping_cost)), 0) FROM messages WHERE DATE(created_at) = CURDATE()),
			(SELECT COUNT(*) FROM users WHERE type = 'customer' AND balance < 100 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM ip_whitelist_requests WHERE status = 'pending')
	`).Scan(&kycPending, &deposits, &smsSent, &profit, &lowBalance, &ipWhitelist)
	if err != nil {
		return databaseError(context, err)
	}

	if tableExists(server, "tickets") {
		_ = server.db.QueryRow("SELECT COUNT(*) FROM tickets WHERE status IN ('open', 'pending')").Scan(&openTickets)
	}

	return context.JSON(fiber.Map{
		"data": map[string]topbarStat{
			"kyc_pending": {
				Title: "KYC Pending", Value: integer(kycPending), URL: "/superman/kyc/requests", Theme: "purple",
			},
			"today_deposit": {
				Title: "Today Deposit", Value: currency(deposits), URL: "/superman/finance/deposits", Theme: "green",
			},
			"today_sms": {
				Title: "Today SMS Sent", Value: integer(smsSent), URL: "/superman/system-logs/sms-logs", Theme: "blue",
			},
			"today_profit": {
				Title: "Today Profit", Value: currency(profit), URL: "/superman/profit/global", Theme: "orange",
			},
			"low_balance": {
				Title: "Low Balance", Value: integer(lowBalance), URL: "/superman/users/customers?filter=low_balance", Theme: "red",
			},
			"open_tickets": {
				Title: "Open Tickets", Value: integer(openTickets), URL: "/superman/sms/inquiry", Theme: "cyan",
			},
			"ip_whitelist": {
				Title: "IP Whitelist", Value: integer(ipWhitelist), URL: "/superman/approval-center/ip-whitelists", Theme: "indigo",
			},
		},
	})
}

func tableExists(server *Server, name string) bool {
	var count int
	err := server.db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = ?
	`, name).Scan(&count)

	return err == nil && count == 1
}

func integer(value int) string {
	return formatNumber(float64(value), 0)
}

func currency(value float64) string {
	return "৳" + formatNumber(value, 2)
}
