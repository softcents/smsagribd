package httpapi

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"smswhiteinfo/superman-api/internal/auth"
	"smswhiteinfo/superman-api/internal/config"
)

type Server struct {
	config config.Config
	db     *sql.DB
	tokens auth.TokenManager
}

func New(config config.Config, database *sql.DB) *Server {
	return &Server{
		config: config,
		db:     database,
		tokens: auth.NewTokenManager(config.JWTSecret),
	}
}

func (server *Server) App() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:            "SMS Whiteinfo Superman API",
		TrustProxy:         true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  []string{"127.0.0.1", "::1"},
			Loopback: true,
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})

	app.Use(func(c fiber.Ctx) error {
		if remoteAddr := c.RemoteAddr(); remoteAddr != nil {
			peerAddr := remoteAddr.String()
			host, _, err := net.SplitHostPort(peerAddr)
			if err != nil {
				host = peerAddr
			}
			peerIP := net.ParseIP(host)
			if (peerIP != nil && peerIP.IsLoopback()) || c.IsFromLocal() {
				if realIP := c.Get("X-Real-IP"); realIP != "" {
					c.Request().Header.Set(fiber.HeaderXForwardedFor, realIP)
				}
			}
		}
		return c.Next()
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{server.config.FrontendOrigin},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/health", func(context fiber.Ctx) error {
		if err := server.db.Ping(); err != nil {
			return context.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}

		return context.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1/superman")
	api.Post("/auth/login", server.login)

	protected := api.Group("", server.authenticate)
	protected.Post("/auth/logout", server.logout)
	protected.Get("/auth/me", server.me)
	protected.Get("/dashboard", server.dashboard)
	protected.Get("/topbar-stats", server.topbarStats)
	protected.Get("/customers", server.customers)
	protected.Get("/transactions", server.transactions)
	protected.Get("/messages", server.messages)
	protected.Get("/batches", server.batches)
	protected.Get("/gateways", server.gateways)
	protected.Post("/gateways", server.createGateway)
	protected.Get("/gateways/:id", server.gatewayDetail)
	protected.Put("/gateways/:id", server.updateGateway)
	protected.Delete("/gateways/:id", server.deleteGateway)
	protected.Get("/gateways/:id/dashboard", server.gatewayDashboard)
	protected.Get("/gateways/:id/operator-rates", server.gatewayOperatorRates)
	protected.Post("/gateways/:id/operator-rates", server.saveOperatorRate)
	protected.Put("/gateways/:id/operator-rates/:rateId", server.saveOperatorRate)
	protected.Delete("/gateways/:id/operator-rates/:rateId", server.deleteOperatorRate)
	protected.Get("/gateways/:id/senders", server.gatewaySenders)
	protected.Get("/settings/:group", server.settings)
	protected.Put("/settings/:group", server.updateSettings)
	protected.Get("/kyc/dashboard", server.kycDashboard)
	protected.Get("/kyc/requests", server.kycRequests)
	protected.Post("/kyc/requests/:id/approve", server.approveKYC)
	protected.Post("/kyc/requests/:id/reject", server.rejectKYC)
	protected.Get("/kyc/audit", server.kycAudit)
	protected.Get("/kyc/permissions", server.kycPermissions)
	protected.Put("/kyc/permissions/:id", server.updateKYCPermissions)
	protected.Get("/deposits", server.deposits)
	protected.Get("/ip-whitelist-requests", server.ipWhitelistRequests)
	protected.Post("/ip-whitelist-requests/:id/approve", server.approveIPWhitelist)
	protected.Post("/ip-whitelist-requests/:id/reject", server.rejectIPWhitelist)
	protected.Get("/sms-logs", server.smsLogs)
	protected.Get("/queue-monitor", server.queueMonitor)
	protected.Post("/queue-actions/:action", server.queueAction)
	protected.Get("/signup-gateways", server.signupGateways)
	protected.Post("/signup-gateways", server.createSignupGateway)
	protected.Put("/signup-gateways/:id/toggle", server.toggleSignupGateway)
	protected.Get("/signup-gateways/:id", server.signupGatewayDetail)
	protected.Put("/signup-gateways/:id", server.updateSignupGateway)
	protected.Delete("/signup-gateways/:id", server.deleteSignupGateway)
	protected.Post("/customers", server.createCustomer)
	protected.Post("/customers/:id/impersonate", server.impersonateCustomer)
	protected.Post("/customers/:id/deposit", server.depositCustomer)
	protected.Put("/customers/:id/status", server.updateCustomerStatus)
	protected.Get("/customers/:id", server.customerDetail)
	protected.Put("/customers/:id", server.updateCustomer)
	protected.Post("/customers/:id/approve-kyc", server.approveKYC)
	protected.Post("/customers/:id/reject-kyc", server.rejectKYC)
	protected.Post("/customers/:id/resubmit-kyc", server.resubmitKYC)
	// S P (Resellers)
	protected.Get("/sps", server.sps)
	protected.Get("/sps/customers", server.spsCustomers)
	protected.Get("/sps/rates", server.spsRates)
	protected.Get("/sps/wallet", server.spsWallets)
	protected.Post("/sps/wallet/topup", server.spsTopup)
	protected.Get("/sps/white-label", server.spsWhiteLabel)
	protected.Post("/sps/white-label", server.spsWhiteLabel)
	protected.Get("/sps/custom-domain", server.spsCustomDomains)
	protected.Get("/sps/profit", server.spsProfit)
	protected.Get("/sps/senders", server.spsSenders)
	protected.Get("/sps/billing", server.spsBilling)
	protected.Get("/sps/kyc", server.spsKyc)
	protected.Get("/sps/statements", server.spsStatements)
	protected.Get("/sps/api", server.spsApi)
	protected.Get("/sps/audit", server.spsAudit)
	protected.Get("/sps/routing", server.spsRouting)
	protected.Get("/sps/webhooks", server.spsWebhooks)
	protected.Get("/sps/reports", server.spsReports)
	protected.Get("/sps/employees", server.spsEmployees)
	protected.Get("/sps/:id", server.spDetail)
	protected.Put("/sps/:id", server.updateSP)
	protected.Post("/sps/:id/deposit", server.depositSP)
	protected.Post("/sps/:id/impersonate", server.impersonateSP)
	protected.Put("/sps/:id/status", server.updateSPStatus)
	protected.Get("/senders", server.senders)
	protected.Post("/senders", server.createSender)
	protected.Delete("/senders", server.deleteSenders)
	protected.Get("/senders/:id", server.senderDetail)
	protected.Put("/senders/:id", server.updateSender)
	protected.Delete("/senders/:id", server.deleteSender)
	protected.Post("/senders/:id/approve", server.approveSender)
	protected.Post("/senders/:id/reject", server.rejectSender)
	protected.Post("/senders/:id/assignment", server.assignSenderUsers)
	protected.Patch("/senders/:id/enabled", server.updateSenderEnabled)
	protected.Get("/sender-cli/options", server.senderCliOptions)
	protected.Get("/sender-cli/summary", server.senderAuditSummary)
	protected.Get("/sender-cli/ledger", server.senderAuditLedger)
	protected.Get("/sender-cli/dispatches", server.senderDispatches)
	protected.Get("/sender-cli/resolution", server.senderResolution)
	protected.Get("/tariff-plans", server.tariffPlans)
	protected.Get("/rates", server.rates)
	protected.Get("/countries", server.countries)
	protected.Get("/operators", server.operators)
protected.Post("/operators", server.createOperator)
	protected.Get("/operators/:id", server.operatorDetail)
	protected.Put("/operators/:id", server.updateOperator)
		protected.Get("/approvals/overview", server.approvalsOverview)
	protected.Get("/kyc/requests/:id", server.kycDetail)
	protected.Post("/kyc/requests/:id/resubmit", server.resubmitKYC)

	// Billing & Financial endpoints
	protected.Get("/billing/dashboard-stats", server.billingDashboardStats)
	protected.Post("/billing/adjust-balance", server.adjustBalance)
	protected.Get("/billing/tariff-plans", server.billingTariffPlans)
	protected.Post("/billing/tariff-plans", server.createTariffPlan)
	protected.Get("/billing/tariff-plans/:id", server.billingTariffPlanDetail)
	protected.Put("/billing/tariff-plans/:id", server.updateTariffPlan)
	protected.Delete("/billing/tariff-plans/:id", server.deleteTariffPlan)
	protected.Get("/billing/assignments", server.billingAssignments)
	protected.Get("/billing/assignments/:user_id", server.billingCustomerAssignmentHistory)
	protected.Post("/billing/assignments", server.assignTariffPlan)
	protected.Get("/billing/rate-history", server.billingRateHistory)
	protected.Post("/deposits/:id/approve", server.approveDeposit)
	protected.Post("/deposits/:id/reject", server.rejectDeposit)
	protected.Get("/billing/invoices", server.billingInvoices)
	protected.Get("/billing/payments", server.billingPayments)
	protected.Get("/billing/usage-cdr", server.billingUsageCdr)
	protected.Get("/billing/reports", server.billingReports)

	// Analytics endpoints
	protected.Get("/analytics/cost-profit", server.analyticsCostProfit)
	protected.Get("/analytics/messaging", server.analyticsMessaging)

	// Operational endpoints (Providers, Reports, Health, Routing, Audit)
	protected.Get("/providers/health", server.providerHealth)
	protected.Post("/providers/health/check-all", server.providerHealthCheckAll)
	protected.Get("/providers/api-logs", server.providerApiLogs)
	protected.Get("/providers/dlr-mapping", server.dlrMappings)
	protected.Get("/providers/error-mapping", server.errorMappings)
	protected.Get("/reports/executive", server.reportsExecutive)
	protected.Get("/templates", server.templates)
	protected.Get("/contacts", server.contacts)
	protected.Get("/monitoring/system-health", server.monitoringSystemHealth)
	protected.Get("/audit-logs/global", server.auditLogsGlobal)
	protected.Get("/routing/dashboard", server.routingDashboard)

	// System Settings
	protected.Get("/system-settings/providers", server.systemSettingsProviders)
	protected.Get("/system-settings/routing", server.systemSettingsRouting)
	protected.Get("/system-settings/queue", server.systemSettingsQueue)
	protected.Get("/system-settings/error-codes", server.systemSettingsErrorCodes)
	protected.Get("/system-settings/dlr", server.systemSettingsDlr)
	protected.Post("/system-settings/dlr", server.systemSettingsDlr)
	protected.Get("/system-settings/maintenance", server.systemSettingsMaintenance)
	protected.Post("/system-settings/maintenance", server.systemSettingsMaintenance)
	protected.Get("/system-settings/backups", server.systemSettingsBackups)
	protected.Post("/system-settings/backups", server.systemSettingsBackups)

	// Security & Fraud Control
	protected.Get("/security/dashboard", server.securityDashboard)
	protected.Get("/security/ip-whitelist", server.securityIpWhitelist)
	protected.Get("/security/fraud-rules", server.securityFraudRules)
	protected.Post("/security/fraud-rules", server.securityFraudRules)
	protected.Get("/security/logs", server.securityLogs)

	// Monitoring / NOC
	protected.Get("/monitoring/live", server.monitoringLive)
	protected.Get("/monitoring/providers", server.monitoringProviders)
	protected.Get("/monitoring/queue-health", server.monitoringQueueHealth)
	protected.Get("/monitoring/api-health", server.monitoringApiHealth)
	protected.Get("/monitoring/tps", server.monitoringTps)

	// Routing
	protected.Get("/routing/rules", server.routingRules)
	protected.Get("/routing/routes", server.routingRoutes)
	protected.Get("/routing/health", server.routingHealth)
	protected.Get("/routing/failover", server.routingFailover)
	protected.Post("/routing/testing", server.routingSimulate)
	protected.Get("/routing/load-balancing", server.routingLoadBalancing)
	protected.Post("/routing/load-balancing", server.routingLoadBalancing)

	// Audit Logs
	protected.Get("/audit-logs/ip-whitelist", server.auditLogsIpWhitelist)
	protected.Get("/audit-logs/employees", server.auditLogsEmployees)
	protected.Get("/audit-logs/security", server.auditLogsSecurity)

	return app
}

func (server *Server) authenticate(context fiber.Ctx) error {
	authorization := context.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(authorization, "Bearer ") {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	claims, err := server.tokens.Parse(strings.TrimPrefix(authorization, "Bearer "))
	if err != nil {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}
	rawToken := strings.TrimPrefix(authorization, "Bearer ")
	if server.isTokenRevoked(rawToken) {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	var userType string
	err = server.db.QueryRow("SELECT type FROM users WHERE id = ? AND deleted_at IS NULL", claims.AdminID).Scan(&userType)
	if err != nil || userType != "superman" {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	context.Locals("superman", claims)
	context.Locals("superman_token", rawToken)

	return context.Next()
}

func (server *Server) tokenCacheKey(token string) string {
	digest := sha256.Sum256([]byte(token))

	return server.config.CachePrefix + "superman_api_revoked_" + hex.EncodeToString(digest[:])
}

func (server *Server) isTokenRevoked(token string) bool {
	var exists bool
	err := server.db.QueryRow("SELECT EXISTS(SELECT 1 FROM cache WHERE `key` = ? AND expiration >= ?)", server.tokenCacheKey(token), time.Now().Unix()).Scan(&exists)

	return err == nil && exists
}

func superman(context fiber.Ctx) auth.Claims {
	return context.Locals("superman").(auth.Claims)
}

func pagination(context fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(context.Query("page", "1"))
	perPage, _ := strconv.Atoi(context.Query("per_page", "25"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 1
	}
	if perPage > 100 {
		perPage = 100
	}

	return page, perPage
}

func apiTime(value time.Time) *string {
	if value.IsZero() {
		return nil
	}

	formatted := value.Format(time.RFC3339)

	return &formatted
}
