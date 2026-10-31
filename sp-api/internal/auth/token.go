package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the SP reseller JWT payload.
type Claims struct {
	SPID  int64  `json:"sp_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
	jwt.RegisteredClaims
}

// TokenManager issues and parses SP JWT tokens.
type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret []byte) TokenManager {
	return TokenManager{secret: secret}
}

func (manager TokenManager) Issue(spID int64, name string, email string, expiresAt time.Time) (string, error) {
	claims := Claims{
		SPID:  spID,
		Name:  name,
		Email: email,
		Type:  "sp",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("sp:%d", spID),
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
	if !ok || !parsed.Valid || claims.Type != "sp" {
		return Claims{}, fmt.Errorf("invalid sp token")
	}

	return *claims, nil
}
