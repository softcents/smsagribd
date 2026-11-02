package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/crypto"
)

type signupOTPGateway struct {
	endpoint       string
	method         string
	timeoutSeconds int
	tokenEnc       string
	senderID       string
	phoneFormat    string
	template       string
}

// dispatchSignupOTP sends the register OTP through active signup gateways in priority order.
// It returns false when none is configured so the caller can use its old route.
func (server *Server) dispatchSignupOTP(phone string, code string) bool {
	rows, err := server.db.Query(`
		SELECT endpoint_url, http_method, timeout_seconds, access_token, sender_id, phone_format, otp_template
		FROM signup_register_gateways
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY priority ASC, id ASC
	`)
	if err != nil {
		return false
	}
	defer rows.Close()
	list := []signupOTPGateway{}
	for rows.Next() {
		var g signupOTPGateway
		if err := rows.Scan(&g.endpoint, &g.method, &g.timeoutSeconds, &g.tokenEnc, &g.senderID, &g.phoneFormat, &g.template); err != nil {
			return false
		}
		list = append(list, g)
	}
	if len(list) == 0 {
		return false
	}
	go func() {
		enc, err := crypto.NewLaravelEncrypter(os.Getenv("APP_KEY"))
		if err != nil {
			return
		}
		for _, g := range list {
			creds, err := enc.DecryptArray(g.tokenEnc)
			if err != nil {
				continue
			}
			token := strings.TrimSpace(fmt.Sprint(creds["token"]))
			if token == "" {
				continue
			}
			text := strings.NewReplacer("{otp}", code, "{minutes}", "10").Replace(g.template)
			if signupGatewaySend(g, token, signupNumber(phone, g.phoneFormat), text) {
				return
			}
		}
	}()
	return true
}

func signupNumber(phone string, format string) string {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	local := digits
	if strings.HasPrefix(digits, "880") && len(digits) >= 12 {
		local = "0" + digits[3:]
	}
	switch format {
	case "intl_880":
		if strings.HasPrefix(local, "0") {
			return "880" + local[1:]
		}
		return digits
	case "intl_plus":
		if strings.HasPrefix(local, "0") {
			return "+880" + local[1:]
		}
		return "+" + digits
	default:
		return local
	}
}

func signupGatewaySend(g signupOTPGateway, token string, number string, body string) bool {
	timeout := time.Duration(g.timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var req *http.Request
	var err error
	if strings.EqualFold(g.method, "GET") {
		u, perr := url.Parse(g.endpoint)
		if perr != nil {
			return false
		}
		q := u.Query()
		q.Set("sender_id", g.senderID)
		q.Set("numbers", number)
		q.Set("body", body)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	} else {
		payload, _ := json.Marshal(map[string]string{"sender_id": g.senderID, "numbers": number, "body": body})
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewReader(payload))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
