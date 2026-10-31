package validation

import (
	"testing"
)

func TestIsDisposable(t *testing.T) {
	tests := []struct {
		email      string
		disposable bool
	}{
		{"user@gmail.com", false},
		{"customer@whiteinfobd.com", false},
		{"admin@yahoo.com", false},
		{"info@outlook.com", false},
		{"test@mailinator.com", true},
		{"someone@guerrillamail.com", true},
		{"attacker@sharklasers.com", true},
		{"temp@sub.mailinator.com", true},
		{"bad@10minutemail.com", true},
		{"disposable@trashmail.com", true},
		{"invalid-email", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsDisposable(tt.email)
		if got != tt.disposable {
			t.Errorf("IsDisposable(%q) = %v, want %v", tt.email, got, tt.disposable)
		}
	}
}
