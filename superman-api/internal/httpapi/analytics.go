package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type dateRange struct {
	From time.Time
	To   time.Time
}

func parseAnalyticsDateRange(c fiber.Ctx) dateRange {
	period := strings.ToLower(strings.TrimSpace(c.Query("period", c.Query("date_range", "last_30_days"))))
	now := time.Now()

	switch period {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return dateRange{From: start, To: now}
	case "yesterday":
		y := now.AddDate(0, 0, -1)
		start := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, now.Location())
		end := time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, 999999999, now.Location())
		return dateRange{From: start, To: end}
	case "this_week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, now.Location())
		return dateRange{From: start, To: now}
	case "last_7_days", "7days":
		start := now.AddDate(0, 0, -7)
		return dateRange{From: start, To: now}
	case "this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return dateRange{From: start, To: now}
	case "last_30_days", "30days":
		start := now.AddDate(0, 0, -30)
		return dateRange{From: start, To: now}
	case "this_year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return dateRange{From: start, To: now}
	case "all", "all_time":
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location())
		return dateRange{From: start, To: now}
	case "custom":
		fromStr := c.Query("date_from", c.Query("from", ""))
		toStr := c.Query("date_to", c.Query("to", ""))
		start, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			start = now.AddDate(0, 0, -30)
		}
		end, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			end = now
		} else {
			end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, now.Location())
		}
		return dateRange{From: start, To: end}
	default:
		start := now.AddDate(0, 0, -30)
		return dateRange{From: start, To: now}
	}
}

func round2(val float64) float64 {
	return math.Round(val*100) / 100
}

func round1(val float64) float64 {
	return math.Round(val*10) / 10
}

type operatorItem struct {
	ID          int
	Name        string
	CountryName string
	CountryCode string
	Codes       []string
}

func (server *Server) loadActiveOperators() ([]operatorItem, error) {
	rows, err := server.db.Query(`
		SELECT o.id, o.name, o.codes, COALESCE(c.name, 'Bangladesh'), COALESCE(c.code, '880')
		FROM operators o
		LEFT JOIN countries c ON o.country_id = c.id
		WHERE o.is_active = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []operatorItem
	for rows.Next() {
		var op operatorItem
		var rawCodes sql.NullString
		if err := rows.Scan(&op.ID, &op.Name, &rawCodes, &op.CountryName, &op.CountryCode); err != nil {
			continue
		}
		if rawCodes.Valid && rawCodes.String != "" {
			var parsed []any
			if err := json.Unmarshal([]byte(rawCodes.String), &parsed); err == nil {
				for _, itm := range parsed {
					str := strings.TrimSpace(fmt.Sprintf("%v", itm))
					if str != "" {
						op.Codes = append(op.Codes, str)
						if !strings.HasPrefix(str, "0") {
							op.Codes = append(op.Codes, "0"+str)
						}
					}
				}
			}
		}
		list = append(list, op)
	}
	return list, nil
}

func matchOperator(phone string, operators []operatorItem) (string, string) {
	digits := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	for _, op := range operators {
		// Try country prefix match (e.g. 880 + 18 or 880 + 018)
		if strings.HasPrefix(digits, op.CountryCode) {
			rem := strings.TrimPrefix(digits, op.CountryCode)
			for _, code := range op.Codes {
				if strings.HasPrefix(rem, code) {
					return op.Name, op.CountryName
				}
			}
		}
		// Direct national prefix match (e.g. 018... or 18...)
		for _, code := range op.Codes {
			if strings.HasPrefix(digits, code) {
				return op.Name, op.CountryName
			}
		}
	}
	return "Unknown", "Unknown"
}

func (server *Server) analyticsCostProfit(c fiber.Ctx) error {
	dr := parseAnalyticsDateRange(c)

	// 1. Period Stats
	var periodCount int
	var periodRev, periodOpCost, periodDipCost float64
	err := server.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(cost), 0), COALESCE(SUM(operator_cost), 0), COALESCE(SUM(primary_dipping_cost), 0)
		FROM messages
		WHERE created_at >= ? AND created_at <= ? AND status != 'failed'
	`, dr.From, dr.To).Scan(&periodCount, &periodRev, &periodOpCost, &periodDipCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	periodCost := periodOpCost + periodDipCost
	periodProfit := periodRev - periodCost
	periodMargin := 0.0
	if periodRev > 0 {
		periodMargin = (periodProfit / periodRev) * 100
	}

	// 2. Overall Stats (All time)
	var overallCount int
	var overallRev, overallOpCost, overallDipCost float64
	_ = server.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(cost), 0), COALESCE(SUM(operator_cost), 0), COALESCE(SUM(primary_dipping_cost), 0)
		FROM messages
		WHERE status != 'failed'
	`).Scan(&overallCount, &overallRev, &overallOpCost, &overallDipCost)

	overallCost := overallOpCost + overallDipCost
	overallProfit := overallRev - overallCost
	overallMargin := 0.0
	if overallRev > 0 {
		overallMargin = (overallProfit / overallRev) * 100
	}

	// 3. Profit Trend (Daily)
	trendRows, err := server.db.Query(`
		SELECT DATE(created_at) as day,
		       COALESCE(SUM(cost), 0) as revenue,
		       COALESCE(SUM(operator_cost), 0) as op_cost,
		       COALESCE(SUM(primary_dipping_cost), 0) as dip_cost
		FROM messages
		WHERE created_at >= ? AND created_at <= ? AND status != 'failed'
		GROUP BY day
		ORDER BY day ASC
	`, dr.From, dr.To)

	dailyMap := make(map[string]map[string]float64)
	if err == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var day string
			var rev, opC, dipC float64
			if err := trendRows.Scan(&day, &rev, &opC, &dipC); err == nil {
				totC := opC + dipC
				dailyMap[day] = map[string]float64{
					"revenue": round2(rev),
					"cost":    round2(totC),
					"profit":  round2(rev - totC),
				}
			}
		}
	}

	var profitTrend []fiber.Map
	curr := time.Date(dr.From.Year(), dr.From.Month(), dr.From.Day(), 0, 0, 0, 0, dr.From.Location())
	endDay := time.Date(dr.To.Year(), dr.To.Month(), dr.To.Day(), 0, 0, 0, 0, dr.To.Location())
	for !curr.After(endDay) {
		dayStr := curr.Format("2006-01-02")
		val, exists := dailyMap[dayStr]
		if !exists {
			val = map[string]float64{"revenue": 0, "cost": 0, "profit": 0}
		}
		profitTrend = append(profitTrend, fiber.Map{
			"date":    dayStr,
			"label":   curr.Format("M 02"),
			"revenue": val["revenue"],
			"cost":    val["cost"],
			"profit":  val["profit"],
		})
		curr = curr.AddDate(0, 0, 1)
	}

	// 4. Gateway Cost & Profit Breakdown
	gwRows, err := server.db.Query(`
		SELECT m.gateway_id, COALESCE(g.name, 'Unknown'), COALESCE(g.driver, 'default'),
		       COUNT(m.id) as msg_count,
		       COALESCE(SUM(m.cost), 0) as rev,
		       COALESCE(SUM(m.operator_cost), 0) as op_cost,
		       COALESCE(SUM(m.primary_dipping_cost), 0) as dip_cost,
		       COALESCE(SUM(m.third_party_dipping_cost), 0) as third_dip_cost,
		       COALESCE(g.has_third_party_dipping, 0) as has_3rd,
		       COALESCE(g.dlr_support, 0) as dlr_sup,
		       SUM(CASE WHEN m.status = 'delivered' THEN 1 ELSE 0 END) as delivered_count
		FROM messages m
		LEFT JOIN gateways g ON m.gateway_id = g.id
		WHERE m.created_at >= ? AND m.created_at <= ? AND m.status != 'failed' AND m.gateway_id IS NOT NULL
		GROUP BY m.gateway_id, g.name, g.driver, g.has_third_party_dipping, g.dlr_support
		ORDER BY msg_count DESC
	`, dr.From, dr.To)

	var gatewayBreakdown []fiber.Map
	var gatewayPerformance []fiber.Map
	if err == nil {
		defer gwRows.Close()
		for gwRows.Next() {
			var gwID sql.NullInt64
			var name, driver string
			var count, has3rd, dlrSup, deliveredCount int
			var rev, opCost, dipCost, thirdDipCost float64
			if err := gwRows.Scan(&gwID, &name, &driver, &count, &rev, &opCost, &dipCost, &thirdDipCost, &has3rd, &dlrSup, &deliveredCount); err == nil {
				totCost := opCost + dipCost + thirdDipCost
				netProfit := rev - totCost
				margin := 0.0
				if rev > 0 {
					margin = (netProfit / rev) * 100
				}
				deliveryRate := 0.0
				if count > 0 {
					deliveryRate = (float64(deliveredCount) / float64(count)) * 100
				}

				var thirdPartyVal *float64
				if has3rd == 1 {
					v := round2(thirdDipCost)
					thirdPartyVal = &v
				}

				gatewayBreakdown = append(gatewayBreakdown, fiber.Map{
					"gateway_id":               gwID.Int64,
					"gateway":                  name,
					"driver":                   driver,
					"messages":                 count,
					"revenue":                  round2(rev),
					"operator_cost":            round2(opCost),
					"primary_dipping_cost":     round2(dipCost),
					"third_party_dipping_cost": thirdPartyVal,
					"gateway_cost":             round2(dipCost),
					"cost":                     round2(totCost),
					"profit":                   round2(netProfit),
					"margin":                   round1(margin),
					"status":                   "complete",
					"dlr_delivery_rate":        round1(deliveryRate),
					"dlr_support":              dlrSup == 1,
				})

				gatewayPerformance = append(gatewayPerformance, fiber.Map{
					"gateway": name,
					"count":   count,
				})
			}
		}
	}

	// 5. Operator Cost Breakdown & Operator Traffic
	operators, _ := server.loadActiveOperators()

	msgRows, err := server.db.Query(`
		SELECT `+"`to`"+`, cost, operator_cost, primary_dipping_cost
		FROM messages
		WHERE created_at >= ? AND created_at <= ? AND status != 'failed' AND `+"`to`"+` IS NOT NULL
	`, dr.From, dr.To)

	type opStat struct {
		OperatorName string
		CountryName  string
		Messages     int
		Revenue      float64
		OperatorCost float64
		GatewayCost  float64
	}

	opMap := make(map[string]*opStat)
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var phone sql.NullString
			var cost, opCost, dipCost float64
			if err := msgRows.Scan(&phone, &cost, &opCost, &dipCost); err == nil && phone.Valid {
				opName, cName := matchOperator(phone.String, operators)
				stat, ok := opMap[opName]
				if !ok {
					stat = &opStat{
						OperatorName: opName,
						CountryName:  cName,
					}
					opMap[opName] = stat
				}
				stat.Messages++
				stat.Revenue += cost
				stat.OperatorCost += opCost
				stat.GatewayCost += dipCost
			}
		}
	}

	var opList []*opStat
	for _, v := range opMap {
		opList = append(opList, v)
	}
	sort.Slice(opList, func(i, j int) bool {
		return opList[i].Messages > opList[j].Messages
	})

	var operatorBreakdown []fiber.Map
	var operatorTraffic []fiber.Map
	var topOperators []fiber.Map

	for i, op := range opList {
		totC := op.OperatorCost + op.GatewayCost
		profit := op.Revenue - totC

		operatorBreakdown = append(operatorBreakdown, fiber.Map{
			"operator":      op.OperatorName,
			"country":       op.CountryName,
			"messages":      op.Messages,
			"revenue":       round2(op.Revenue),
			"operator_cost": round2(op.OperatorCost),
			"gateway_cost":  round2(op.GatewayCost),
			"cost":          round2(totC),
			"profit":        round2(profit),
		})

		operatorTraffic = append(operatorTraffic, fiber.Map{
			"operator": op.OperatorName,
			"count":    op.Messages,
		})

		if i < 10 {
			topOperators = append(topOperators, fiber.Map{
				"operator": op.OperatorName,
				"messages": op.Messages,
				"revenue":  round2(op.Revenue),
			})
		}
	}

	return c.JSON(fiber.Map{
		"period_stats": fiber.Map{
			"revenue":  round2(periodRev),
			"cost":     round2(periodCost),
			"profit":   round2(periodProfit),
			"margin":   round1(periodMargin),
			"messages": periodCount,
		},
		"overall_stats": fiber.Map{
			"revenue":  round2(overallRev),
			"cost":     round2(overallCost),
			"profit":   round2(overallProfit),
			"margin":   round1(overallMargin),
			"messages": overallCount,
		},
		"profit_trend":        profitTrend,
		"gateway_breakdown":   gatewayBreakdown,
		"operator_breakdown":  operatorBreakdown,
		"gateway_performance": gatewayPerformance,
		"operator_traffic":    operatorTraffic,
		"top_operators":       topOperators,
	})
}

func (server *Server) analyticsMessaging(c fiber.Ctx) error {
	dr := parseAnalyticsDateRange(c)

	// 1. Core KPIs
	var total, successful, delivered, failed, unicode int
	var totalCost, avgParts float64
	err := server.db.QueryRow(`
		SELECT COUNT(*),
		       SUM(CASE WHEN status IN ('sent', 'delivered') THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status IN ('failed', 'dlr_failed') THEN 1 ELSE 0 END),
		       COALESCE(SUM(CASE WHEN status NOT IN ('failed', 'dlr_failed') THEN cost ELSE 0 END), 0),
		       COALESCE(AVG(parts), 1),
		       SUM(CASE WHEN is_unicode = 1 THEN 1 ELSE 0 END)
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
	`, dr.From, dr.To).Scan(&total, &successful, &delivered, &failed, &totalCost, &avgParts, &unicode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	successRate := 0.0
	deliveredRate := 0.0
	failedRate := 0.0
	unicodeRate := 0.0
	if total > 0 {
		successRate = (float64(successful) / float64(total)) * 100
		deliveredRate = (float64(delivered) / float64(total)) * 100
		failedRate = (float64(failed) / float64(total)) * 100
		unicodeRate = (float64(unicode) / float64(total)) * 100
	}

	// 2. Message Trends (Daily total count)
	tRows, err := server.db.Query(`
		SELECT DATE(created_at) as day, COUNT(*) as count
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY day
		ORDER BY day ASC
	`, dr.From, dr.To)

	dailyTrendMap := make(map[string]int)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var d string
			var c int
			if err := tRows.Scan(&d, &c); err == nil {
				dailyTrendMap[d] = c
			}
		}
	}

	var messageTrends []fiber.Map
	curr := time.Date(dr.From.Year(), dr.From.Month(), dr.From.Day(), 0, 0, 0, 0, dr.From.Location())
	endDay := time.Date(dr.To.Year(), dr.To.Month(), dr.To.Day(), 0, 0, 0, 0, dr.To.Location())
	for !curr.After(endDay) {
		dStr := curr.Format("2006-01-02")
		cnt := dailyTrendMap[dStr]
		messageTrends = append(messageTrends, fiber.Map{
			"date":  dStr,
			"label": curr.Format("M 02"),
			"count": cnt,
		})
		curr = curr.AddDate(0, 0, 1)
	}

	// 3. Delivery Rate Trend (Daily delivery rate %)
	drRows, err := server.db.Query(`
		SELECT DATE(created_at) as day,
		       COUNT(*) as total,
		       SUM(CASE WHEN status IN ('sent', 'delivered') THEN 1 ELSE 0 END) as successful
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY day
		ORDER BY day ASC
	`, dr.From, dr.To)

	dailyRateMap := make(map[string]float64)
	if err == nil {
		defer drRows.Close()
		for drRows.Next() {
			var d string
			var tot, succ int
			if err := drRows.Scan(&d, &tot, &succ); err == nil {
				rate := 0.0
				if tot > 0 {
					rate = (float64(succ) / float64(tot)) * 100
				}
				dailyRateMap[d] = round1(rate)
			}
		}
	}

	var deliveryRateTrends []fiber.Map
	curr = time.Date(dr.From.Year(), dr.From.Month(), dr.From.Day(), 0, 0, 0, 0, dr.From.Location())
	for !curr.After(endDay) {
		dStr := curr.Format("2006-01-02")
		rate, exists := dailyRateMap[dStr]
		if !exists {
			rate = 0.0
		}
		deliveryRateTrends = append(deliveryRateTrends, fiber.Map{
			"date":  dStr,
			"label": curr.Format("M 02"),
			"rate":  rate,
		})
		curr = curr.AddDate(0, 0, 1)
	}

	// 4. Source Distribution (web vs API)
	srcRows, err := server.db.Query(`
		SELECT COALESCE(source, 'web') as src, COUNT(*) as count
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY src
		ORDER BY count DESC
	`, dr.From, dr.To)

	var sourceDistribution []fiber.Map
	if err == nil {
		defer srcRows.Close()
		for srcRows.Next() {
			var src string
			var count int
			if err := srcRows.Scan(&src, &count); err == nil {
				pct := 0.0
				if total > 0 {
					pct = round1((float64(count) / float64(total)) * 100)
				}
				sourceDistribution = append(sourceDistribution, fiber.Map{
					"source":     strings.ToUpper(src),
					"count":      count,
					"percentage": pct,
				})
			}
		}
	}

	// 5. DLR Summary (Delivered, Failed, Sent, Pending, Sending)
	var dCount, fCount, sCount, pCount, sndCount int
	_ = server.db.QueryRow(`
		SELECT SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status IN ('failed', 'dlr_failed') THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'sending' THEN 1 ELSE 0 END)
		FROM messages
		WHERE created_at >= ? AND created_at <= ?
	`, dr.From, dr.To).Scan(&dCount, &fCount, &sCount, &pCount, &sndCount)

	dlrSummary := []fiber.Map{
		{"status": "Delivered", "count": dCount, "color": "#10b981"},
		{"status": "Failed", "count": fCount, "color": "#ef4444"},
		{"status": "Sent", "count": sCount, "color": "#0ea5e9"},
		{"status": "Pending", "count": pCount, "color": "#f59e0b"},
		{"status": "Sending", "count": sndCount, "color": "#8b5cf6"},
	}

	// 6. Failed Reason Summary
	frRows, err := server.db.Query(`
		SELECT failed_reason, COUNT(*) as count
		FROM messages
		WHERE created_at >= ? AND created_at <= ? AND status IN ('failed', 'dlr_failed') AND failed_reason IS NOT NULL AND failed_reason != ''
		GROUP BY failed_reason
		ORDER BY count DESC
		LIMIT 10
	`, dr.From, dr.To)

	var failedReasons []fiber.Map
	if err == nil {
		defer frRows.Close()
		for frRows.Next() {
			var reason string
			var count int
			if err := frRows.Scan(&reason, &count); err == nil {
				failedReasons = append(failedReasons, fiber.Map{
					"reason": reason,
					"count":  count,
				})
			}
		}
	}

	return c.JSON(fiber.Map{
		"stats": fiber.Map{
			"total":           total,
			"successful":      successful,
			"delivered":       delivered,
			"failed":          failed,
			"success_rate":    round1(successRate),
			"delivered_rate":  round1(deliveredRate),
			"failed_rate":     round1(failedRate),
			"total_cost":      round2(totalCost),
			"avg_parts":       round1(avgParts),
			"unicode":         unicode,
			"unicode_percent": round1(unicodeRate),
		},
		"message_trends":       messageTrends,
		"delivery_rate_trends": deliveryRateTrends,
		"source_distribution":  sourceDistribution,
		"dlr_summary":          dlrSummary,
		"failed_reasons":        failedReasons,
	})
}
