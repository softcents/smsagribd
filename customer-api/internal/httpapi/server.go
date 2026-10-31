package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"smswhiteinfo/customer-api/internal/auth"
	"smswhiteinfo/customer-api/internal/batch"
	"smswhiteinfo/customer-api/internal/config"
	"smswhiteinfo/customer-api/internal/dlr"
	"smswhiteinfo/customer-api/internal/email"
	"smswhiteinfo/customer-api/internal/kyc"
	"smswhiteinfo/customer-api/internal/scheduler"
	"smswhiteinfo/customer-api/internal/worker"
)

type Server struct {
	config         config.Config
	db             *sql.DB
	tokens         auth.TokenManager
	batchProcessor *batch.BatchProcessor
	workerPool     *worker.SendWorkerPool
	dlrPoller      *dlr.DeliveryPoller
	scheduler      *scheduler.MasterScheduler
	kycService     *kyc.KycService
	mailer         *email.Mailer
	loginLimiter   *LoginLimiter
}

func New(config config.Config, database *sql.DB) *Server {
	storageRoot := os.Getenv("STORAGE_PATH")
	if storageRoot == "" {
		storageRoot = os.Getenv("KYC_STORAGE_PATH")
	}
	if storageRoot == "" {
		storageRoot = "./storage/app/private"
	}

	appKey := os.Getenv("APP_KEY")

	batchProc := batch.NewBatchProcessor(database, storageRoot)
	kycSvc := kyc.NewKycService(database, storageRoot)
	mailer := email.NewMailerFromEnv()

	workerPool, _ := worker.NewSendWorkerPool(database, appKey, 4)
	dlrPoller, _ := dlr.NewDeliveryPoller(database, appKey)
	masterSched := scheduler.NewMasterScheduler(database, batchProc)

	return &Server{
		config:         config,
		db:             database,
		tokens:         auth.NewTokenManager(config.JWTSecret),
		batchProcessor: batchProc,
		workerPool:     workerPool,
		dlrPoller:      dlrPoller,
		scheduler:      masterSched,
		kycService:     kycSvc,
		mailer:         mailer,
		loginLimiter:   NewLoginLimiter(),
	}
}

func (server *Server) StartBackground(ctx context.Context) {
	if server.workerPool != nil {
		server.workerPool.Start(ctx)
	}
	if server.dlrPoller != nil {
		server.dlrPoller.Start(ctx)
	}
	if server.scheduler != nil {
		server.scheduler.Start(ctx)
	}
}

func (server *Server) StopBackground() {
	if server.workerPool != nil {
		server.workerPool.Stop()
	}
	if server.dlrPoller != nil {
		server.dlrPoller.Stop()
	}
	if server.scheduler != nil {
		server.scheduler.Stop()
	}
}

func (server *Server) App() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:            "SMS Whiteinfo Customer API",
		BodyLimit:          20 * 1024 * 1024, // 20MB for file uploads
		TrustProxy:         true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  []string{"127.0.0.1", "::1"},
			Loopback: true,
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})

	// Normalize proxy headers: use X-Real-IP only when the TCP peer is loopback
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

	origins := []string{server.config.FrontendOrigin, "http://localhost:5173", "http://127.0.0.1:5173"}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Requested-With", "Accept"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/health", func(context fiber.Ctx) error {
		if err := server.db.Ping(); err != nil {
			return context.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "unavailable",
				"database": "down",
			})
		}

		var inFlight int64
		var lastLoop time.Time
		if server.workerPool != nil {
			inFlight, lastLoop = server.workerPool.Status()
		}

		return context.JSON(fiber.Map{
			"status":    "ok",
			"database":  "ok",
			"in_flight": inFlight,
			"last_loop": lastLoop.Format(time.RFC3339),
		})
	})

	// Public Webhooks & Callbacks
	app.All("/webhooks/sms/:gateway", server.handleSmsWebhook)
	app.Post("/webhooks/payment/:gateway", server.handlePaymentWebhook)
	app.Get("/payments/:gateway/callback/:deposit", server.handlePaymentCallback)

	// Phone Photo capture
	app.Post("/phone-photo/session", server.kycService.CreatePhonePhotoSession)
	app.Get("/phone-photo/:token", server.kycService.ShowCapturePage)
	app.Post("/phone-photo/:token/upload", server.kycService.UploadPhoto)
	app.Get("/phone-photo/:token/status", server.kycService.CheckStatus)

	// Private KYC & Report streaming
	app.Get("/kyc-files/:user/:field", server.handleKycFileStream)
	app.Get("/downloads/batches/:batch/failed-report/:filename", server.handleFailedReportDownload)

	api := app.Group("/api/v1/customer")

	// Public routes
	api.Post("/login", server.login)
	api.Post("/register", server.register)
	api.Post("/forgot-password", server.forgotPassword)
	api.Post("/reset-password", server.resetPassword)
	api.Get("/register/plan/:slug", server.registerPlan)

	// Protected routes
	protected := api.Group("", server.authenticate)
	protected.Post("/logout", server.logout)
	protected.Get("/me", server.me)
	protected.Put("/profile", server.updateProfile)
	protected.Put("/password", server.updatePassword)

	// Dashboard
	protected.Get("/dashboard/stats", server.dashboardStats)

	// Messages
	protected.Get("/messages", server.messages)
	protected.Post("/messages/send", server.sendMessage)
	protected.Post("/messages/schedule", server.scheduleMessage)
	protected.Post("/messages/bulk-send", server.bulkSend)

	// Batches
	protected.Get("/batches", server.batches)
	protected.Get("/batches/:id", server.batchDetail)

	// Senders
	protected.Get("/senders", server.senders)
	protected.Post("/senders", server.createSender)

	// Contacts
	protected.Get("/contacts", server.contacts)
	protected.Post("/contacts", server.createContact)
	protected.Delete("/contacts/:id", server.deleteContact)

	// Groups
	protected.Get("/groups", server.groups)
	protected.Post("/groups", server.createGroup)
	protected.Delete("/groups/:id", server.deleteGroup)

	// Templates
	protected.Get("/templates", server.templates)
	protected.Post("/templates", server.createTemplate)
	protected.Delete("/templates/:id", server.deleteTemplate)

	// Transactions
	protected.Get("/transactions", server.transactions)

	// Deposits
	protected.Get("/deposits", server.deposits)
	protected.Post("/deposits", server.createDeposit)

	// KYC
	protected.Get("/kyc", server.kycShow)
	protected.Post("/kyc", server.kycSubmit)

	// IP Whitelists
	protected.Get("/ip-whitelists", server.ipWhitelists)
	protected.Post("/ip-whitelists", server.createIpWhitelist)
	protected.Delete("/ip-whitelists/:id", server.deleteIpWhitelist)

	// Pricing
	protected.Get("/pricing", server.pricing)

	// Developer / API tokens
	protected.Get("/tokens", server.apiTokens)
	protected.Post("/tokens", server.createApiToken)
	protected.Delete("/tokens/:id", server.deleteApiToken)

	// Analytics
	protected.Get("/analytics/messaging", server.analyticsMessaging)
	protected.Get("/analytics/billing", server.analyticsBilling)

	// Live queue
	protected.Get("/live-queue", server.liveQueue)

	// OTP
	protected.Post("/otp/send", server.otpSend)
	protected.Post("/otp/verify", server.otpVerify)

	// External developer API (replaces Laravel routes/api.php)
	server.RegisterExternalAPI(app)

	return app
}

func (server *Server) authenticate(context fiber.Ctx) error {
	var rawToken string
	authorization := context.Get(fiber.HeaderAuthorization)
	if strings.HasPrefix(authorization, "Bearer ") {
		rawToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if cookie := context.Cookies("customer_token"); cookie != "" {
		rawToken = cookie
	}

	if rawToken == "" {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}
	claims, err := server.tokens.Parse(rawToken)
	if err != nil {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	if server.isTokenRevoked(rawToken) {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	// Verify user still exists and is a customer
	var userType, userStatus string
	err = server.db.QueryRow(
		"SELECT type, status FROM users WHERE id = ? AND deleted_at IS NULL",
		claims.UserID,
	).Scan(&userType, &userStatus)
	if err != nil || userType != "customer" {
		return context.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	context.Locals("customer", claims)
	context.Locals("customer_token", rawToken)

	return context.Next()
}

func (server *Server) tokenCacheKey(token string) string {
	digest := sha256.Sum256([]byte(token))

	return server.config.CachePrefix + "customer_api_revoked_" + hex.EncodeToString(digest[:])
}

func (server *Server) isTokenRevoked(token string) bool {
	var exists bool
	err := server.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM cache WHERE `key` = ? AND expiration >= ?)",
		server.tokenCacheKey(token), time.Now().Unix(),
	).Scan(&exists)

	return err == nil && exists
}

func customer(context fiber.Ctx) auth.Claims {
	return context.Locals("customer").(auth.Claims)
}

func pagination(context fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(context.Query("page", "1"))
	perPage, _ := strconv.Atoi(context.Query("per_page", "15"))

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

func apiTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	return apiTime(*value)
}

func pageMeta(total, perPage, page int) fiber.Map {
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	return fiber.Map{
		"total":        total,
		"per_page":     perPage,
		"current_page": page,
		"last_page":    lastPage,
	}
}

func databaseError(context fiber.Ctx, err error) error {
	return context.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"message": "A database error occurred.",
		"detail":  err.Error(),
	})
}
