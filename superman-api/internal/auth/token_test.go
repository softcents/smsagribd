package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssuesAndParsesSupermanTokens(t *testing.T) {
	manager := NewTokenManager([]byte("a-development-secret-that-is-long-enough"))

	token, err := manager.Issue(9, "Superman", "superman@example.com", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.AdminID != 9 || claims.Email != "superman@example.com" || claims.Type != "superman" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestTokenManagerRejectsTokensSignedWithAnotherSecret(t *testing.T) {
	issuer := NewTokenManager([]byte("a-development-secret-that-is-long-enough"))
	validator := NewTokenManager([]byte("a-different-development-secret-that-is-long-enough"))

	token, err := issuer.Issue(9, "Superman", "superman@example.com", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	if _, err := validator.Parse(token); err == nil {
		t.Fatal("expected token signed with another secret to be rejected")
	}
}
