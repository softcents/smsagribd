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

type adminUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

func (server *Server) login(context fiber.Ctx) error {
	request := loginRequest{}
	if err := json.Unmarshal(context.Body(), &request); err != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "The supplied data is invalid."})
	}

	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	if request.Email == "" || request.Password == "" {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "The supplied data is invalid.",
			"errors":  fiber.Map{"email": []string{"Email and password are required."}},
		})
	}

	var user adminUser
	var passwordHash string
	err := server.db.QueryRow(`
		SELECT id, name, email, type, password
		FROM users
		WHERE email = ? AND type = 'superman' AND deleted_at IS NULL
		LIMIT 1
	`, request.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Type, &passwordHash)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)) != nil {
		return context.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "The supplied credentials are invalid.",
			"errors":  fiber.Map{"email": []string{"These credentials do not match our records."}},
		})
	}
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to sign in."})
	}

	_, _ = server.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = ?", user.ID)

	ttl := server.config.TokenTTL
	if request.Remember {
		ttl = server.config.RememberTokenTTL
	}

	token, err := server.tokens.Issue(user.ID, user.Name, user.Email, time.Now().Add(ttl))
	if err != nil {
		return context.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Unable to sign in."})
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}

func (server *Server) logout(context fiber.Ctx) error {
	claims := superman(context)
	token, _ := context.Locals("superman_token").(string)
	if token != "" && claims.ExpiresAt != nil {
		_, err := server.db.Exec(`
			INSERT INTO cache (`+"`key`"+`, value, expiration)
			VALUES (?, 'b:1;', ?)
			ON DUPLICATE KEY UPDATE value = VALUES(value), expiration = VALUES(expiration)
		`, server.tokenCacheKey(token), claims.ExpiresAt.Unix())
		if err != nil {
			return databaseError(context, err)
		}
	}

	return context.SendStatus(http.StatusNoContent)
}

func (server *Server) me(context fiber.Ctx) error {
	claims := superman(context)

	return context.JSON(fiber.Map{
		"data": adminUser{
			ID:    claims.AdminID,
			Name:  claims.Name,
			Email: claims.Email,
			Type:  claims.Type,
		},
	})
}
