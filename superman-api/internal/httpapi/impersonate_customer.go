package httpapi

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// customerImpersonationClaims must match customer-api/internal/auth/token.go.
type customerImpersonationClaims struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func (server *Server) impersonateCustomer(context fiber.Ctx) error {
	userID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || userID <= 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid customer id."})
	}

	secret := os.Getenv("CUSTOMER_API_JWT_SECRET")
	if secret == "" {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Customer portal signing key is not configured."})
	}

	var name, email string
	err = server.db.QueryRow(`
		SELECT COALESCE(name, ''), email FROM users
		WHERE id = ? AND type = 'customer' AND deleted_at IS NULL LIMIT 1
	`, userID).Scan(&name, &email)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Customer not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	now := time.Now()
	expiresAt := now.Add(2 * time.Hour)
	claims := customerImpersonationClaims{
		UserID: userID,
		Name:   name,
		Email:  email,
		Type:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "customer:" + strconv.FormatInt(userID, 10),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to create customer session."})
	}

	return context.JSON(fiber.Map{"data": fiber.Map{"token": token, "expires_at": expiresAt.Format(time.RFC3339)}})
}
