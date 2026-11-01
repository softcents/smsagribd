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

// spImpersonationClaims must match the claims sp-api expects (sp-api/internal/auth/token.go).
type spImpersonationClaims struct {
	SPID  int64  `json:"sp_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
	jwt.RegisteredClaims
}

// impersonateSP lets a signed-in superman admin open the reseller portal as a reseller.
func (server *Server) impersonateSP(context fiber.Ctx) error {
	spID, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil || spID <= 0 {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "Invalid reseller id."})
	}

	secret := os.Getenv("SP_API_JWT_SECRET")
	if secret == "" {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Reseller portal signing key is not configured."})
	}

	var name, email string
	err = server.db.QueryRow(`
		SELECT COALESCE(name, ''), email
		FROM users
		WHERE id = ? AND type = 'sp' AND deleted_at IS NULL
		LIMIT 1
	`, spID).Scan(&name, &email)
	if err == sql.ErrNoRows {
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Reseller not found."})
	}
	if err != nil {
		return databaseError(context, err)
	}

	now := time.Now()
	expiresAt := now.Add(2 * time.Hour)
	claims := spImpersonationClaims{
		SPID:  spID,
		Name:  name,
		Email: email,
		Type:  "sp",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "sp:" + strconv.FormatInt(spID, 10),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to create reseller session."})
	}

	return context.JSON(fiber.Map{"data": fiber.Map{
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	}})
}
