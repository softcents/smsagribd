package email

import (
	"os"
	"testing"
)

func TestMailer_UnconfiguredFallback(t *testing.T) {
	// Ensure SMTP is unconfigured (matching Laravel MAIL_MAILER=log)
	_ = os.Setenv("MAIL_HOST", "")
	_ = os.Setenv("SMTP_HOST", "")

	mailer := NewMailerFromEnv()
	if mailer.Host != "" {
		t.Fatalf("Expected empty Host for unconfigured mailer, got: %s", mailer.Host)
	}

	// 1. Test SendMail does not return error
	err := mailer.SendMail("test@example.com", "Test Subject", "<p>Test Body</p>")
	if err != nil {
		t.Errorf("Expected nil error for unconfigured mailer, got: %v", err)
	}

	// 2. Test Send does not return error
	err2 := mailer.Send("test@example.com", "Password Reset", "Reset link: http://example.com")
	if err2 != nil {
		t.Errorf("Expected nil error for Send method, got: %v", err2)
	}
}
