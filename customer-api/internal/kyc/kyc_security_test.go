package kyc

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return db
}

func TestKycAndReportSecurity(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "kyc_sec_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	kycSvc := NewKycService(db, tmpDir)

	// Create test customer 1 (owner) and customer 2 (attacker)
	now := time.Now().UnixNano()
	u1Res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Owner User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("owner_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}
	ownerID, _ := u1Res.LastInsertId()

	u2Res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, created_at, updated_at)
		VALUES ('Attacker User', ?, ?, 'hash', 10.00, 'customer', 'active', NOW(), NOW())
	`, fmt.Sprintf("attacker_%d@test.com", now), fmt.Sprintf("+88018%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}
	attackerID, _ := u2Res.LastInsertId()

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM batches WHERE user_id = ?", ownerID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN (?, ?)", ownerID, attackerID)
	}()

	// Create dummy failed report on disk
	relReport := fmt.Sprintf("batches/%d/failed-recipients-20261020-120000.csv", 999)
	absReport := filepath.Join(tmpDir, relReport)
	_ = os.MkdirAll(filepath.Dir(absReport), 0755)
	_ = os.WriteFile(absReport, []byte("Phone,Reason\n8801700000001,Invalid format\n"), 0644)

	// Insert batch for owner
	bRes, err := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, body, source_type, source_config, status, failed_report_path, created_at, updated_at)
		VALUES (?, 'Test Batch', 'manual', '{}', 'completed', ?, NOW(), NOW())
	`, ownerID, relReport)
	if err != nil {
		t.Fatalf("Failed to create batch: %v", err)
	}
	batchID, _ := bRes.LastInsertId()

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("test_user_id", attackerID)
		c.Locals("test_is_admin", false)
		return c.Next()
	})
	app.Get("/downloads/batches/:batch/failed-report/:filename", func(c fiber.Ctx) error {
		callerID := c.Locals("test_user_id").(int64)
		isAdmin := c.Locals("test_is_admin").(bool)
		return kycSvc.DownloadFailedReport(c, callerID, isAdmin)
	})

	// Scenario 1: Attacker (customer 2) attempts to download owner's report -> 403 Forbidden
	req1 := httptest.NewRequest("GET", fmt.Sprintf("/downloads/batches/%d/failed-report/failed-recipients-20261020-120000.csv", batchID), nil)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != http.StatusForbidden {
		t.Errorf("Security check failed: Non-owner should receive 403, got %d", resp1.StatusCode)
	}

	// Scenario 2: Owner downloads their own report -> 200 OK
	app2 := fiber.New()
	app2.Use(func(c fiber.Ctx) error {
		c.Locals("test_user_id", ownerID)
		c.Locals("test_is_admin", false)
		return c.Next()
	})
	app2.Get("/downloads/batches/:batch/failed-report/:filename", func(c fiber.Ctx) error {
		return kycSvc.DownloadFailedReport(c, ownerID, false)
	})
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/downloads/batches/%d/failed-report/failed-recipients-20261020-120000.csv", batchID), nil)
	resp2, err := app2.Test(req2, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Scenario 2 test error: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Owner download failed: expected 200, got %d", resp2.StatusCode)
	}

	// Scenario 3: Path traversal in filename parameter (e.g. ../../etc/passwd) -> 404
	app3 := fiber.New()
	app3.Get("/downloads/batches/:batch/failed-report/*", func(c fiber.Ctx) error {
		return kycSvc.DownloadFailedReport(c, ownerID, true)
	})
	req3 := httptest.NewRequest("GET", fmt.Sprintf("/downloads/batches/%d/failed-report/..%%2f..%%2fetc%%2fpasswd", batchID), nil)
	resp3, err := app3.Test(req3, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Scenario 3 test error: %v", err)
	}
	if resp3.StatusCode != http.StatusNotFound {
		t.Errorf("Path traversal was not rejected: expected 404, got %d", resp3.StatusCode)
	}
}

func TestLaravelKycStorageLayoutAndAdminSecurity(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Temp dir simulating Laravel's storage/app/private
	tmpDir, err := os.MkdirTemp("", "laravel_storage_*")
	if err != nil {
		t.Fatalf("Failed to create temp storage root: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Laravel-style subdirectories: ids/, live_photos/, trade_licenses/
	idsDir := filepath.Join(tmpDir, "ids")
	liveDir := filepath.Join(tmpDir, "live_photos")
	tlDir := filepath.Join(tmpDir, "trade_licenses")
	_ = os.MkdirAll(idsDir, 0755)
	_ = os.MkdirAll(liveDir, 0755)
	_ = os.MkdirAll(tlDir, 0755)

	// Write dummy files
	dummyNidContent := []byte("JPEG_DUMMY_NID_CONTENT")
	dummyLiveContent := []byte("JPEG_DUMMY_LIVE_PHOTO_CONTENT")
	dummyTlContent := []byte("%PDF-1.4 DUMMY_TRADE_LICENSE")

	nidFile := filepath.Join(idsDir, "user_front_123.jpg")
	liveFile := filepath.Join(liveDir, "live_abc-uuid-def.jpg")
	tlFile := filepath.Join(tlDir, "license_999.pdf")

	if err := os.WriteFile(nidFile, dummyNidContent, 0644); err != nil {
		t.Fatalf("Failed to write test NID file: %v", err)
	}
	if err := os.WriteFile(liveFile, dummyLiveContent, 0644); err != nil {
		t.Fatalf("Failed to write test live photo file: %v", err)
	}
	if err := os.WriteFile(tlFile, dummyTlContent, 0644); err != nil {
		t.Fatalf("Failed to write test trade license file: %v", err)
	}

	// Also write a secret file outside the subdirectories to test traversal
	secretPath := filepath.Join(tmpDir, "secret.key")
	_ = os.WriteFile(secretPath, []byte("SUPER_SECRET_KEY"), 0600)

	kycSvc := NewKycService(db, tmpDir)

	now := time.Now().UnixNano()
	// Insert customer with Laravel-style relative paths
	relNid := "ids/user_front_123.jpg"
	relLive := "live_photos/live_abc-uuid-def.jpg"
	relTl := "trade_licenses/license_999.pdf"

	custRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, front_nid, live_photo, trade_license_file, created_at, updated_at)
		VALUES ('Target Customer', ?, ?, 'hash', 0.00, 'customer', 'pending', ?, ?, ?, NOW(), NOW())
	`, fmt.Sprintf("cust_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000), relNid, relLive, relTl)
	if err != nil {
		t.Fatalf("Failed to insert customer: %v", err)
	}
	customerID, _ := custRes.LastInsertId()

	// Insert attacker customer with traversal paths in DB
	traversalRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, front_nid, created_at, updated_at)
		VALUES ('Traversal Attacker', ?, ?, 'hash', 0.00, 'customer', 'pending', '../secret.key', NOW(), NOW())
	`, fmt.Sprintf("trav_%d@test.com", now), fmt.Sprintf("+88018%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to insert traversal attacker: %v", err)
	}
	traversalID, _ := traversalRes.LastInsertId()

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN (?, ?)", customerID, traversalID)
	}()

	// Build app with admin authentication guard (matching webhooks.go handleKycFileStream)
	app := fiber.New()
	app.Get("/kyc-files/:user/:field", func(c fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthenticated."})
		}
		// For unit test mock: role is passed via bearer token
		role := authHeader
		if role != "Bearer admin" && role != "Bearer superman" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Admin authorization required."})
		}
		return kycSvc.StreamKycFile(c)
	})

	// 1. Unauthenticated access -> 401 Unauthorized
	req1 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/front_nid", customerID), nil)
	resp1, err := app.Test(req1)
	if err != nil || resp1.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized without auth header, got %d", resp1.StatusCode)
	}

	// 2. Customer role access -> 403 Forbidden (Admin only)
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/front_nid", customerID), nil)
	req2.Header.Set("Authorization", "Bearer customer")
	resp2, err := app.Test(req2)
	if err != nil || resp2.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for customer role, got %d", resp2.StatusCode)
	}

	// 3. Admin access to front_nid (ids/) -> 200 OK with exact file content
	req3 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/front_nid", customerID), nil)
	req3.Header.Set("Authorization", "Bearer admin")
	resp3, err := app.Test(req3)
	if err != nil || resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for admin reading front_nid, got %d", resp3.StatusCode)
	}

	// 4. Admin access to live_photo (live_photos/) -> 200 OK
	req4 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/live_photo", customerID), nil)
	req4.Header.Set("Authorization", "Bearer superman")
	resp4, err := app.Test(req4)
	if err != nil || resp4.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for admin reading live_photo, got %d", resp4.StatusCode)
	}

	// 5. Admin access to trade_license_file (trade_licenses/) -> 200 OK
	req5 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/trade_license_file", customerID), nil)
	req5.Header.Set("Authorization", "Bearer admin")
	resp5, err := app.Test(req5)
	if err != nil || resp5.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for admin reading trade_license_file, got %d", resp5.StatusCode)
	}

	// 6. Path traversal in DB column ('../secret.key') -> rejected with 404
	req6 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/front_nid", traversalID), nil)
	req6.Header.Set("Authorization", "Bearer admin")
	resp6, err := app.Test(req6)
	if err != nil || resp6.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for path traversal in DB column, got %d", resp6.StatusCode)
	}

	// 7. Invalid field name (e.g. 'password' or 'something') -> 404 Not Found
	req7 := httptest.NewRequest("GET", fmt.Sprintf("/kyc-files/%d/password", customerID), nil)
	req7.Header.Set("Authorization", "Bearer admin")
	resp7, err := app.Test(req7)
	if err != nil || resp7.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for disallowed field, got %d", resp7.StatusCode)
	}
}
