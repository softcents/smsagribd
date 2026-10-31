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

	"smswhiteinfo/sp-api/internal/auth"
	"smswhiteinfo/sp-api/internal/config"
)

// Server is the SP API HTTP server.
type Server struct {
	config config.Config
	db     *sql.DB
	tokens auth.TokenManager
}

func New(cfg config.Config, database *sql.DB) *Server {
	return &Server{
		config: cfg,
		db:     database,
		tokens: auth.NewTokenManager(cfg.JWTSecret),
	}
}

func (server *Server) App() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:            "SMS Whiteinfo SP Reseller API",
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

	app.Get("/health", func(ctx fiber.Ctx) error {
		if err := server.db.Ping(); err != nil {
			return ctx.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}

		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1/sp")

	// Public
	api.Post("/auth/login", server.login)

	// Protected – require valid SP Bearer token
	protected := api.Group("", server.authenticate)
	protected.Post("/auth/logout", server.logout)
	protected.Get("/auth/me", server.me)

	// Dashboard stats
	protected.Get("/dashboard", server.dashboard)

	// Customers managed by this SP
	protected.Get("/customers", server.customers)
	protected.Post("/customers/:id/fund", server.fundCustomer)

	// Wallet / distribution ledger
	protected.Get("/wallet", server.wallet)

	// Billing – SP transaction ledger
	protected.Get("/billing", server.billing)

	// SMS traffic for customers under this SP
	protected.Get("/messages", server.messages)

	return app
}

// authenticate validates the Bearer JWT and ensures the user is of type "sp".
func (server *Server) authenticate(ctx fiber.Ctx) error {
	authorization := ctx.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(authorization, "Bearer ") {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	rawToken := strings.TrimPrefix(authorization, "Bearer ")
	claims, err := server.tokens.Parse(rawToken)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	if server.isTokenRevoked(rawToken) {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	// Verify the user still exists and is still of type "sp"
	var userType string
	err = server.db.QueryRow(
		"SELECT type FROM users WHERE id = ? AND deleted_at IS NULL",
		claims.SPID,
	).Scan(&userType)
	if err != nil || userType != "sp" {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
	}

	ctx.Locals("sp", claims)
	ctx.Locals("sp_token", rawToken)

	return ctx.Next()
}

func (server *Server) tokenCacheKey(token string) string {
	digest := sha256.Sum256([]byte(token))

	return server.config.CachePrefix + "sp_api_revoked_" + hex.EncodeToString(digest[:])
}

func (server *Server) isTokenRevoked(token string) bool {
	var exists bool
	err := server.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM cache WHERE `key` = ? AND expiration >= ?)",
		server.tokenCacheKey(token),
		time.Now().Unix(),
	).Scan(&exists)

	return err == nil && exists
}

// sp extracts the validated SP claims from the request context.
func sp(ctx fiber.Ctx) auth.Claims {
	return ctx.Locals("sp").(auth.Claims)
}

func pagination(ctx fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "25"))

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
