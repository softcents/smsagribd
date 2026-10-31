package httpapi

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestFiber_HeaderSpoofingPrevention(t *testing.T) {
	// ── Setup 1: App where 127.0.0.1 is trusted proxy (e.g. Behind local Nginx) ──
	trustedApp := fiber.New(fiber.Config{
		AppName:    "Trusted Proxy App",
		TrustProxy: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  []string{"127.0.0.1", "::1"},
			Loopback: true,
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})

	trustedApp.Use(func(c fiber.Ctx) error {
		if c.Get(fiber.HeaderXForwardedFor) == "" {
			if realIP := c.Get("X-Real-IP"); realIP != "" {
				c.Request().Header.Set(fiber.HeaderXForwardedFor, realIP)
			}
		}
		return c.Next()
	})

	trustedApp.Get("/client-ip", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ip": c.IP()})
	})

	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln1.Close()
	go trustedApp.Listener(ln1)
	time.Sleep(50 * time.Millisecond)

	trustedAddr := "http://" + ln1.Addr().String()

	// Scenario 1: Trusted proxy forwards client IP in X-Forwarded-For
	req1, _ := http.NewRequest(http.MethodGet, trustedAddr+"/client-ip", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.42")
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("Request 1 failed: %v", err)
	}
	defer resp1.Body.Close()
	b1, _ := io.ReadAll(resp1.Body)
	var res1 map[string]string
	_ = json.Unmarshal(b1, &res1)
	if res1["ip"] != "203.0.113.42" {
		t.Errorf("Expected client IP 203.0.113.42 from trusted proxy, got: %v", res1["ip"])
	}

	// Scenario 2: Trusted proxy forwards client IP in X-Real-IP
	req2, _ := http.NewRequest(http.MethodGet, trustedAddr+"/client-ip", nil)
	req2.Header.Set("X-Real-IP", "198.51.100.77")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("Request 2 failed: %v", err)
	}
	defer resp2.Body.Close()
	b2, _ := io.ReadAll(resp2.Body)
	var res2 map[string]string
	_ = json.Unmarshal(b2, &res2)
	if res2["ip"] != "198.51.100.77" {
		t.Errorf("Expected client IP 198.51.100.77 from X-Real-IP via trusted proxy, got: %v", res2["ip"])
	}

	// Scenario 2b (Requirement B.2):
	// peer 127.0.0.1, "X-Forwarded-For: 1.2.3.4, 5.6.7.8", "X-Real-IP: 5.6.7.8" must yield 5.6.7.8
	// (Never use leftmost X-Forwarded-For 1.2.3.4; loopback peer trusts X-Real-IP 5.6.7.8)
	req2b, _ := http.NewRequest(http.MethodGet, trustedAddr+"/client-ip", nil)
	req2b.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	req2b.Header.Set("X-Real-IP", "5.6.7.8")
	resp2b, err := http.DefaultClient.Do(req2b)
	if err != nil {
		t.Fatalf("Request 2b failed: %v", err)
	}
	defer resp2b.Body.Close()
	b2b, _ := io.ReadAll(resp2b.Body)
	var res2b map[string]string
	_ = json.Unmarshal(b2b, &res2b)
	if res2b["ip"] != "5.6.7.8" {
		t.Errorf("Requirement B.2 failure: expected client IP 5.6.7.8, got %v (leftmost spoofing occurred!)", res2b["ip"])
	}

	// ── Setup 2: App where connecting client IP (127.0.0.1) is UNTRUSTED ──
	// (Simulates untrusted direct client sending spoofed proxy headers)
	untrustedApp := fiber.New(fiber.Config{
		AppName:    "Untrusted Client App",
		TrustProxy: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies:  []string{"10.254.254.254"}, // Only trust 10.254.254.254, NOT 127.0.0.1
			Loopback: false,
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})

	untrustedApp.Use(func(c fiber.Ctx) error {
		if c.Get(fiber.HeaderXForwardedFor) == "" {
			if realIP := c.Get("X-Real-IP"); realIP != "" {
				c.Request().Header.Set(fiber.HeaderXForwardedFor, realIP)
			}
		}
		return c.Next()
	})

	untrustedApp.Get("/client-ip", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ip": c.IP()})
	})

	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln2.Close()
	go untrustedApp.Listener(ln2)
	time.Sleep(50 * time.Millisecond)

	untrustedAddr := "http://" + ln2.Addr().String()

	// Scenario 3: Untrusted client attempts to spoof X-Forwarded-For
	req3, _ := http.NewRequest(http.MethodGet, untrustedAddr+"/client-ip", nil)
	req3.Header.Set("X-Forwarded-For", "8.8.8.8")
	req3.Header.Set("X-Real-IP", "8.8.8.8")
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("Request 3 failed: %v", err)
	}
	defer resp3.Body.Close()
	b3, _ := io.ReadAll(resp3.Body)
	var res3 map[string]string
	_ = json.Unmarshal(b3, &res3)

	// Since 127.0.0.1 is NOT in untrustedApp's Proxies, Fiber MUST ignore 8.8.8.8 and return 127.0.0.1
	if res3["ip"] == "8.8.8.8" {
		t.Fatalf("CRITICAL SECURITY FLAW: Untrusted client successfully spoofed IP to 8.8.8.8!")
	}
	if res3["ip"] != "127.0.0.1" {
		t.Errorf("Expected socket IP 127.0.0.1 for untrusted connection, got: %v", res3["ip"])
	}
}
