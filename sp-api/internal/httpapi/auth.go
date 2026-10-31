package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type spUser struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Balance float64 `json:"balance"`
	Type    string  `json:"type"`
	Status  string  `json:"status"`
}

func (server *Server) login(ctx fiber.Ctx) error {
	request := loginRequest{}
	if err := json.Unmarshal(ctx.Body(), &request); err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied data is invalid."})
	}

	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	if request.Email == "" || request.Password == "" {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "The supplied data is invalid.",
			"errors":  fiber.Map{"email": []string{"Email and password are required."}},
		})
	}

	var user spUser
	var passwordHash string
	err := server.db.QueryRow(`
		SELECT id, name, email, COALESCE(phone, ''), balance, type, status, password
		FROM users
		WHERE email = ? AND type = 'sp' AND deleted_at IS NULL
		LIMIT 1
	`, request.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Balance, &user.Type, &user.Status, &passwordHash)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)) != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "The supplied credentials are invalid.",
			"errors":  fiber.Map{"email": []string{"These credentials do not match our records."}},
		})
	}
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to sign in."})
	}

	_, _ = server.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = ?", user.ID)

	ttl := server.config.TokenTTL
	if request.Remember {
		ttl = server.config.RememberTokenTTL
	}

	token, err := server.tokens.Issue(user.ID, user.Name, user.Email, time.Now().Add(ttl))
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to sign in."})
	}

	return ctx.JSON(fiber.Map{
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}

func (server *Server) logout(ctx fiber.Ctx) error {
	claims := sp(ctx)
	token, _ := ctx.Locals("sp_token").(string)
	if token != "" && claims.ExpiresAt != nil {
		_, err := server.db.Exec(`
			INSERT INTO cache (`+"`key`"+`, value, expiration)
			VALUES (?, 'b:1;', ?)
			ON DUPLICATE KEY UPDATE value = VALUES(value), expiration = VALUES(expiration)
		`, server.tokenCacheKey(token), claims.ExpiresAt.Unix())
		if err != nil {
			return databaseError(ctx, err)
		}
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (server *Server) me(ctx fiber.Ctx) error {
	claims := sp(ctx)

	var user spUser
	err := server.db.QueryRow(`
		SELECT id, name, email, COALESCE(phone, ''), balance, type, status
		FROM users
		WHERE id = ? AND type = 'sp' AND deleted_at IS NULL
	`, claims.SPID).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Balance, &user.Type, &user.Status)
	if err != nil {
		return databaseError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"data": user})
}
