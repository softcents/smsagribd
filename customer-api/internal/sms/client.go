package sms

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var testModeOverride atomic.Bool

// SetTestMode sets test mode for all outbound HTTP requests in sms package.
func SetTestMode(enabled bool) {
	testModeOverride.Store(enabled)
}

// IsTestMode returns true only if test mode was explicitly enabled in code (via SetTestMode).
// It cannot be activated by environment variables or CLI flags in production.
func IsTestMode() bool {
	return testModeOverride.Load()
}

// SafeHTTPClient creates an HTTP client that strictly blocks non-localhost
// traffic whenever running in test mode.
type SafeHTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewSafeHTTPClient(baseURL string, timeout time.Duration) *SafeHTTPClient {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &SafeHTTPClient{
		client:  &http.Client{Transport: tr, Timeout: timeout},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *SafeHTTPClient) checkSecurityGuard(targetURL string) error {
	if !IsTestMode() {
		return nil
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" && host != "::1" && host != "" {
		return fmt.Errorf("SECURITY GUARD BLOCKED OUTBOUND REQUEST: Attempted to contact '%s' during test mode. In test mode, only localhost mock servers are allowed", u.Host)
	}
	return nil
}

func (c *SafeHTTPClient) PostJSON(ctx context.Context, endpoint string, payload any, headers map[string]string) ([]byte, int, error) {
	fullURL := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		fullURL = c.baseURL + "/" + strings.TrimLeft(endpoint, "/")
	}

	if err := c.checkSecurityGuard(fullURL); err != nil {
		return nil, 0, err
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed reading response body: %w", err)
	}

	return respBytes, resp.StatusCode, nil
}

func (c *SafeHTTPClient) Get(ctx context.Context, endpoint string, queryParams map[string]string, headers map[string]string) ([]byte, int, error) {
	fullURL := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		fullURL = c.baseURL + "/" + strings.TrimLeft(endpoint, "/")
	}

	if len(queryParams) > 0 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, 0, err
		}
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	if err := c.checkSecurityGuard(fullURL); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed reading response body: %w", err)
	}

	return respBytes, resp.StatusCode, nil
}

var ErrRequestFailed = errors.New("upstream gateway returned an error status")
