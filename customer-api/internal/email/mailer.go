package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type Mailer struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

func getEnvWithFallback(primary, secondary string) string {
	if val := os.Getenv(primary); val != "" {
		return val
	}
	return os.Getenv(secondary)
}

func NewMailerFromEnv() *Mailer {
	rawPort := getEnvWithFallback("MAIL_PORT", "SMTP_PORT")
	port, _ := strconv.Atoi(rawPort)
	if port <= 0 {
		port = 587
	}

	fromAddr := getEnvWithFallback("MAIL_FROM_ADDRESS", "SMTP_FROM_ADDRESS")
	if fromAddr == "" {
		fromAddr = "noreply@whiteinfobd.com"
	}

	fromName := getEnvWithFallback("MAIL_FROM_NAME", "SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "SMS WhiteInfo"
	}

	return &Mailer{
		Host:        getEnvWithFallback("MAIL_HOST", "SMTP_HOST"),
		Port:        port,
		Username:    getEnvWithFallback("MAIL_USERNAME", "SMTP_USERNAME"),
		Password:    getEnvWithFallback("MAIL_PASSWORD", "SMTP_PASSWORD"),
		Encryption:  strings.ToLower(getEnvWithFallback("MAIL_ENCRYPTION", "SMTP_ENCRYPTION")),
		FromAddress: fromAddr,
		FromName:    fromName,
	}
}

func (m *Mailer) SendMail(to, subject, bodyHTML string) error {
	if m.Host == "" {
		// Log mailer fallback matching Laravel MAIL_MAILER=log (Item 5)
		log.Printf("[MAIL LOG] (SMTP unconfigured, log mailer fallback) To: %s | Subject: %s | Body: %s", to, subject, bodyHTML)
		return nil
	}

	fromHeader := fmt.Sprintf("%s <%s>", m.FromName, m.FromAddress)
	headers := make(map[string]string)
	headers["From"] = fromHeader
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(bodyHTML)

	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))

	// SSL / TLS direct connection (port 465 or ssl)
	if m.Encryption == "ssl" || m.Port == 465 {
		tlsConfig := &tls.Config{
			ServerName: m.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to establish SSL connection: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, m.Host)
		if err != nil {
			return err
		}
		defer c.Quit()

		if m.Username != "" {
			auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
			if err := c.Auth(auth); err != nil {
				return err
			}
		}

		if err := c.Mail(m.FromAddress); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}

		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg.String())); err != nil {
			return err
		}
		return w.Close()
	}

	// Plain or STARTTLS (port 25 / 587)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if m.Encryption == "tls" || m.Port == 587 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: m.Host}
			if err := c.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}

	if m.Username != "" {
		auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	if err := c.Mail(m.FromAddress); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg.String())); err != nil {
		return err
	}
	return w.Close()
}

// Send sends an email with plain text or HTML body, falling back to log when SMTP is unconfigured.
func (m *Mailer) Send(to, subject, body string) error {
	return m.SendMail(to, subject, body)
}
