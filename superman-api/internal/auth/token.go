package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID int64  `json:"admin_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret []byte) TokenManager {
	return TokenManager{secret: secret}
}

func (manager TokenManager) Issue(adminID int64, name string, email string, expiresAt time.Time) (string, error) {
	claims := Claims{
		AdminID: adminID,
		Name:    name,
		Email:   email,
		Type:    "superman",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("superman:%d", adminID),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(manager.secret)
}

func (manager TokenManager) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		return manager.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.Type != "superman" {
		return Claims{}, fmt.Errorf("invalid superman token")
	}

	return *claims, nil
}
